package scoring

import (
	"path/filepath"
	"sort"
	"strings"

	"faultline/internal/model"
)

type deltaBucket struct {
	kind    string
	score   float64
	reasons []string
}

func buildDelta(repoState *RepoState) *model.Delta {
	if repoState == nil {
		return nil
	}
	files := dedupeStrings(append(append([]string(nil), repoState.ChangedFiles...), repoState.RecentFiles...))
	tests := dedupeStrings(repoState.TestsNewlyFailing)
	errors := dedupeStrings(repoState.ErrorsAdded)
	if len(files) == 0 && len(repoState.DriftSignals) == 0 && len(repoState.HotfixSignals) == 0 && len(tests) == 0 && len(errors) == 0 && len(repoState.EnvDiff) == 0 {
		return nil
	}

	buckets := map[string]*deltaBucket{
		"dependency":        {kind: "dependency"},
		"runtime_toolchain": {kind: "runtime_toolchain"},
		"ci_config":         {kind: "ci_config"},
		"environment":       {kind: "environment"},
		"deploy_infra":      {kind: "deploy_infra"},
		"source_code":       {kind: "source_code"},
		"test_data":         {kind: "test_data"},
	}

	for _, file := range files {
		class, reason, score := classifyDeltaFile(file)
		if class == "" || score == 0 {
			continue
		}
		bucket := buckets[class]
		bucket.score += score
		appendReason(bucket, reason)
	}

	for _, signal := range repoState.HotfixSignals {
		appendReason(buckets["environment"], "hotfix signal present")
		buckets["environment"].score += 0.25
		if strings.Contains(strings.ToLower(signal), "revert") {
			appendReason(buckets["environment"], "revert signal present")
			buckets["environment"].score += 0.15
		}
	}
	for _, signal := range repoState.DriftSignals {
		appendReason(buckets["environment"], strings.ToLower(signal))
		buckets["environment"].score += 0.2
	}
	if len(tests) > 0 {
		appendReason(buckets["test_data"], "new failing tests detected")
		buckets["test_data"].score += 0.9
	}

	var causes []model.DeltaCause
	for _, bucket := range buckets {
		if bucket.score <= 0 {
			continue
		}
		causes = append(causes, model.DeltaCause{
			Kind:    bucket.kind,
			Score:   round(bucket.score),
			Reasons: append([]string(nil), bucket.reasons...),
		})
	}
	sort.Slice(causes, func(i, j int) bool {
		if causes[i].Score != causes[j].Score {
			return causes[i].Score > causes[j].Score
		}
		return causes[i].Kind < causes[j].Kind
	})
	if len(causes) == 0 {
		if len(files) == 0 && len(tests) == 0 && len(errors) == 0 && len(repoState.EnvDiff) == 0 {
			return nil
		}
	}
	return &model.Delta{
		Version:           loadedVersionOrDefault(),
		Provider:          strings.TrimSpace(repoState.Provider),
		FilesChanged:      dedupeStrings(repoState.ChangedFiles),
		TestsNewlyFailing: tests,
		ErrorsAdded:       errors,
		EnvDiff:           cloneEnvDiff(repoState.EnvDiff),
		Signals:           buildDeltaSignals(repoState),
		Causes:            causes,
	}
}

func loadedVersionOrDefault() string {
	weights, err := defaultWeights()
	if err != nil || strings.TrimSpace(weights.Version) == "" {
		return "bayes_v1"
	}
	return weights.Version
}

func classifyDeltaFile(file string) (kind, reason string, score float64) {
	base := strings.ToLower(filepath.Base(file))
	file = strings.ToLower(filepath.ToSlash(file))
	switch {
	case isDependencyFile(base, file):
		return "dependency", "dependency file changed", 1.0
	case isRuntimeFile(base, file):
		return "runtime_toolchain", "runtime or toolchain file changed", 1.0
	case isCIFile(base, file):
		return "ci_config", "workflow or CI config changed", 1.0
	case isEnvironmentFile(base, file):
		return "environment", "environment or config file changed", 0.9
	case isDeployFile(base, file):
		return "deploy_infra", "deploy or infra file changed", 1.0
	case isTestFile(base, file):
		return "test_data", "test-only file changed", 0.8
	default:
		if isSourceFile(base) {
			return "source_code", "source file changed", 0.6
		}
	}
	return "", "", 0
}

func appendReason(bucket *deltaBucket, reason string) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return
	}
	for _, existing := range bucket.reasons {
		if existing == reason {
			return
		}
	}
	bucket.reasons = append(bucket.reasons, reason)
	sort.Strings(bucket.reasons)
}

func cloneEnvDiff(in map[string]model.DeltaEnvChange) map[string]model.DeltaEnvChange {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]model.DeltaEnvChange, len(in))
	for key, value := range in {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildDeltaSignals(repoState *RepoState) []model.DeltaSignal {
	if repoState == nil {
		return nil
	}
	var signals []model.DeltaSignal
	files := dedupeStrings(repoState.ChangedFiles)
	if len(files) > 0 {
		signals = append(signals, model.DeltaSignal{
			ID:     "delta.scope.changed",
			Detail: strings.Join(files, ", "),
			Weight: 1.0,
		})
	}
	if hasClass(files, "dependency") {
		signals = append(signals, model.DeltaSignal{
			ID:     "delta.dependency.changed",
			Detail: "dependency-related files changed",
			Weight: 1.2,
		})
	}
	if len(repoState.TestsNewlyFailing) > 0 {
		signals = append(signals, model.DeltaSignal{
			ID:     "delta.test.failure.introduced",
			Detail: strings.Join(dedupeStrings(repoState.TestsNewlyFailing), ", "),
			Weight: 1.2,
		})
	}
	if len(repoState.ErrorsAdded) > 0 {
		signals = append(signals, model.DeltaSignal{
			ID:     "delta.error.new",
			Detail: repoState.ErrorsAdded[0],
			Weight: 1.0,
		})
	}
	sort.Slice(signals, func(i, j int) bool {
		if signals[i].ID != signals[j].ID {
			return signals[i].ID < signals[j].ID
		}
		if signals[i].Weight != signals[j].Weight {
			return signals[i].Weight > signals[j].Weight
		}
		return signals[i].Detail < signals[j].Detail
	})
	return signals
}

func hasClass(files []string, class string) bool {
	for _, file := range files {
		k, _, _ := classifyDeltaFile(file)
		if k == class {
			return true
		}
	}
	return false
}
