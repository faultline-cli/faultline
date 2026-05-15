package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"faultline/internal/artifact"
	"faultline/internal/metrics"
	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/policy"
	"faultline/internal/store"
)

type historySnapshot struct {
	seenCount    int
	signature    store.Signature
	signatureHit store.SignatureHistory
}

func prepareAnalysisWithStore(a *model.Analysis, rawInput string, sourceKind, surface string, opts AnalyzeOptions, persist bool) (*model.Analysis, error) {
	if a == nil {
		return nil, nil
	}
	prepared := a
	prepared.InputHash = ""
	if sourceKind == "log" && rawInput != "" {
		prepared.InputHash = store.InputHashForLog(rawInput)
	}

	storePath := opts.Store
	cfg, err := store.ResolveConfig(storePath)
	if err != nil {
		return prepared, err
	}
	st, info, err := store.OpenBestEffort(cfg)
	if err != nil {
		return prepared, err
	}
	defer st.Close()
	if info.Degraded {
		fmt.Fprintf(os.Stderr, "WARN: faultline store degraded to no-op: %s\n", info.Warning)
	}

	ctx := context.Background()
	now := optionNow(opts)
	historyEnabled := info.Mode != store.ModeOff && !info.Degraded
	historyOutput := shouldIncludeHistoryOutput(opts)
	snapshots := captureHistorySnapshots(ctx, st, prepared)
	previousFailures, _ := st.RecentTopFailures(ctx, 500)

	withoutHistory := applySignatureSnapshots(prepared, snapshots)
	withoutCurrent := applyHistorySnapshots(prepared, snapshots, now, false)
	withCurrent := applyHistorySnapshots(prepared, snapshots, now, persist && historyEnabled)
	withoutHistory = artifact.Sync(withoutHistory)
	withoutCurrent = artifact.Sync(withoutCurrent)
	withCurrent = artifact.Sync(withCurrent)

	if historyOutput {
		withoutCurrent.Metrics = buildMetricsFromHistory(withoutCurrent, previousFailures, opts.MetricsHistoryFile, false)
		if withoutCurrent.Metrics != nil && len(withoutCurrent.Results) > 0 {
			withoutCurrent.Policy = policy.Compute(withoutCurrent.Metrics, withoutCurrent.Results[0].Playbook.Severity)
		}
	}

	if persist && historyEnabled && len(withCurrent.Results) > 0 {
		withCurrent.Metrics = buildMetricsFromHistory(withCurrent, previousFailures, opts.MetricsHistoryFile, true)
		if withCurrent.Metrics != nil {
			withCurrent.Policy = policy.Compute(withCurrent.Metrics, withCurrent.Results[0].Playbook.Severity)
		}
		if hash, err := output.HashAnalysisOutput(withCurrent); err == nil {
			withCurrent.OutputHash = hash
		}
		handle, beginErr := st.BeginRun(ctx, store.BeginRunParams{
			Surface:    surface,
			SourceKind: sourceKind,
			Source:     withCurrent.Source,
			InputHash:  withCurrent.InputHash,
			StartedAt:  now,
		})
		if beginErr == nil {
			if completeErr := st.CompleteRun(ctx, handle, store.CompleteRunParams{
				CompletedAt: now,
				Analysis:    withCurrent,
			}); completeErr == nil {
				if historyOutput {
					return withCurrent, nil
				}
				if hash, err := output.HashAnalysisOutput(withoutHistory); err == nil {
					withoutHistory.OutputHash = hash
				}
				return withoutHistory, nil
			}
		}
	}

	if historyOutput {
		if hash, err := output.HashAnalysisOutput(withoutCurrent); err == nil {
			withoutCurrent.OutputHash = hash
		}
		return withoutCurrent, nil
	}
	if hash, err := output.HashAnalysisOutput(withoutHistory); err == nil {
		withoutHistory.OutputHash = hash
	}
	return withoutHistory, nil
}

func captureHistorySnapshots(ctx context.Context, st store.Store, a *model.Analysis) []historySnapshot {
	snapshots := make([]historySnapshot, len(a.Results))
	for i, result := range a.Results {
		sig := store.SignatureForResult(result)
		snapshots[i].signature = sig
		seenCount, _ := st.CountSeenFailure(ctx, result.Playbook.ID)
		snapshots[i].seenCount = seenCount
		history, _ := st.LookupSignatureHistory(ctx, sig.Hash)
		snapshots[i].signatureHit = history
	}
	return snapshots
}

func applyHistorySnapshots(base *model.Analysis, snapshots []historySnapshot, now time.Time, includeCurrent bool) *model.Analysis {
	clone := cloneAnalysis(base)
	for i := range clone.Results {
		result := clone.Results[i]
		snapshot := snapshots[i]
		result.SeenCount = snapshot.seenCount
		result.SignatureHash = snapshot.signature.Hash
		result.SeenBefore = snapshot.signatureHit.OccurrenceCount > 0
		result.FirstSeenAt = snapshot.signatureHit.FirstSeenAt
		result.LastSeenAt = snapshot.signatureHit.LastSeenAt
		result.OccurrenceCount = snapshot.signatureHit.OccurrenceCount
		if includeCurrent && i == 0 {
			result.OccurrenceCount = snapshot.signatureHit.OccurrenceCount + 1
			if result.FirstSeenAt == "" {
				result.FirstSeenAt = now.Format(time.RFC3339)
			}
			result.LastSeenAt = now.Format(time.RFC3339)
		}
		clone.Results[i] = result
	}
	return clone
}

func applySignatureSnapshots(base *model.Analysis, snapshots []historySnapshot) *model.Analysis {
	clone := cloneAnalysis(base)
	for i := range clone.Results {
		result := clone.Results[i]
		if i < len(snapshots) {
			result.SignatureHash = snapshots[i].signature.Hash
		}
		result.SeenCount = 0
		result.SeenBefore = false
		result.OccurrenceCount = 0
		result.FirstSeenAt = ""
		result.LastSeenAt = ""
		clone.Results[i] = result
	}
	return clone
}

func shouldIncludeHistoryOutput(opts AnalyzeOptions) bool {
	if opts.History {
		return true
	}
	if strings.TrimSpace(opts.MetricsHistoryFile) != "" {
		return true
	}
	value := strings.TrimSpace(opts.Store)
	if value == "" || strings.EqualFold(value, string(store.ModeAuto)) || strings.EqualFold(value, string(store.ModeOff)) {
		return false
	}
	return true
}

func buildMetricsFromHistory(a *model.Analysis, previousFailures []string, explicitHistoryPath string, includeCurrent bool) *model.Metrics {
	if a == nil || len(a.Results) == 0 {
		return nil
	}
	failures := append([]string(nil), previousFailures...)
	if includeCurrent {
		failures = append([]string{a.Results[0].Playbook.ID}, failures...)
	}
	localEntries := make([]metrics.LocalEntry, 0, len(failures))
	for _, failureID := range failures {
		localEntries = append(localEntries, metrics.LocalEntry{FailureID: failureID})
	}
	m := metrics.FromLocalHistory(a.Results[0].Playbook.ID, localEntries)
	if explicitHistoryPath != "" {
		explicit, err := metrics.LoadHistoryFile(explicitHistoryPath)
		if err == nil {
			m = metrics.WithExplicitHistory(m, explicit)
		}
	}
	return m
}

func cloneAnalysis(a *model.Analysis) *model.Analysis {
	if a == nil {
		return nil
	}
	clone := *a
	clone.Results = append([]model.Result(nil), a.Results...)
	return &clone
}
