package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"faultline/internal/artifact"
	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/store"
)

type historySnapshot struct {
	seenCount    int
	signature    store.Signature
	signatureHit store.SignatureHistory
}

func prepareAnalysisWithStore(ctx context.Context, a *model.Analysis, rawInput string, sourceKind, surface string, opts AnalyzeOptions, persist bool) (*model.Analysis, error) {
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

	now := optionNow(opts)
	historyEnabled := info.Mode != store.ModeOff && !info.Degraded
	historyOutput := shouldIncludeHistoryOutput(opts)
	snapshots := captureHistorySnapshots(ctx, st, prepared)

	withoutHistory := applySignatureSnapshots(prepared, snapshots)
	withoutCurrent := applyHistorySnapshots(prepared, snapshots, now, false)
	withCurrent := applyHistorySnapshots(prepared, snapshots, now, persist && historyEnabled)
	withoutHistory = artifact.Sync(withoutHistory)
	withoutCurrent = artifact.Sync(withoutCurrent)
	withCurrent = artifact.Sync(withCurrent)
	if persist && historyEnabled && len(withCurrent.Results) > 0 {
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
		if seenCount, err := st.CountSeenFailure(ctx, result.Playbook.ID); err == nil {
			snapshots[i].seenCount = seenCount
		} else {
			fmt.Fprintf(os.Stderr, "WARN: store.CountSeenFailure: %s\n", err)
		}
		if history, err := st.LookupSignatureHistory(ctx, sig.Hash); err == nil {
			snapshots[i].signatureHit = history
		} else {
			fmt.Fprintf(os.Stderr, "WARN: store.LookupSignatureHistory: %s\n", err)
		}
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
	value := strings.TrimSpace(opts.Store)
	if value == "" || strings.EqualFold(value, string(store.ModeAuto)) || strings.EqualFold(value, string(store.ModeOff)) {
		return false
	}
	return true
}

func cloneAnalysis(a *model.Analysis) *model.Analysis {
	if a == nil {
		return nil
	}
	clone := *a
	clone.Results = append([]model.Result(nil), a.Results...)
	return &clone
}
