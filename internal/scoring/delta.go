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
	if len(files) == 0 && len(repoState.DriftSignals) == 0 && len(repoState.HotfixSignals) == 0 {
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
		return nil
	}
	return &model.Delta{
		Version: loadedVersionOrDefault(),
		Causes:  causes,
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
