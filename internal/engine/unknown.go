package engine

import (
	"fmt"
	"sort"
	"strings"

	"faultline/internal/authoring"
	"faultline/internal/model"
)

func buildUnknownDiagnosis(lines []model.Line, ctx model.Context) ([]model.CandidateCluster, []string, *model.SuggestedPlaybookSeed) {
	raw := joinOriginalLines(lines)
	candidates := authoring.ExtractCandidatePatterns(raw, 8)
	if len(candidates) == 0 {
		return nil, nil, nil
	}

	dominant := append([]string(nil), candidates...)
	if len(dominant) > 5 {
		dominant = dominant[:5]
	}

	grouped := map[string][]string{}
	for _, candidate := range dominant {
		category := inferUnknownCategory(candidate, ctx)
		grouped[category] = append(grouped[category], candidate)
	}

	keys := make([]string, 0, len(grouped))
	for key := range grouped {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	clusters := make([]model.CandidateCluster, 0, len(keys))
	for index, key := range keys {
		signals := grouped[key]
		summary := fmt.Sprintf("Unmatched %s signals remain clustered around one root cause.", key)
		confidence := 0.45 - float64(index)*0.08
		if confidence < 0.2 {
			confidence = 0.2
		}
		clusters = append(clusters, model.CandidateCluster{
			Key:            key,
			Summary:        summary,
			LikelyCategory: key,
			Confidence:     confidence,
			Signals:        append([]string(nil), signals...),
			Evidence:       append([]string(nil), signals...),
		})
	}

	primaryCategory := inferUnknownCategory(dominant[0], ctx)
	seed := &model.SuggestedPlaybookSeed{
		Category: primaryCategory,
		Title:    fmt.Sprintf("Observed %s failure signature", strings.ReplaceAll(primaryCategory, "-", " ")),
		MatchAny: append([]string(nil), dominant...),
		Workflow: model.WorkflowSpec{
			LikelyFiles: likelyFilesForUnknownCategory(primaryCategory),
			LocalRepro: []string{
				"faultline analyze <logfile> --json",
			},
			Verify: []string{
				"make review",
				"make test",
			},
		},
	}
	return clusters, dominant, seed
}

func inferUnknownCategory(signal string, ctx model.Context) string {
	value := strings.ToLower(strings.TrimSpace(signal))
	switch {
	case containsAny(value, "docker", "image pull", "kubectl", "terraform", "deployment", "helm"):
		return "deploy"
	case containsAny(value, "timeout", "dial tcp", "lookup", "certificate", "x509", "refused", "host"):
		return "network"
	case containsAny(value, "command not found", "no such file or directory", "version", "runtime", "node", "python"):
		return "runtime"
	case containsAny(value, "fixture", "assert", "expected", "snapshot", "panic", "test"):
		return "test"
	case containsAny(value, "permission", "auth", "token", "login", "unauthorized", "forbidden"):
		return "auth"
	case containsAny(value, "workflow", "yaml", "runner", "artifact", "pipeline", "job"):
		return "ci"
	case ctx.Stage == "deploy":
		return "deploy"
	case ctx.Stage == "test":
		return "test"
	default:
		return "build"
	}
}

func likelyFilesForUnknownCategory(category string) []string {
	switch category {
	case "auth":
		return []string{".github/workflows/*.yml", ".gitlab-ci.yml", ".npmrc", ".docker/config.json"}
	case "ci":
		return []string{".github/workflows/*.yml", ".gitlab-ci.yml", ".circleci/config.yml", "Makefile"}
	case "deploy":
		return []string{"Dockerfile", ".github/workflows/*.yml", "deploy/", "infra/"}
	case "network":
		return []string{".github/workflows/*.yml", ".gitlab-ci.yml", "Dockerfile", "infra/"}
	case "runtime":
		return []string{"Dockerfile", ".tool-versions", ".nvmrc", "package.json", "pyproject.toml"}
	case "test":
		return []string{"tests/", "testdata/", "fixtures/", "package.json", "pytest.ini"}
	default:
		return []string{"Makefile", "Dockerfile", ".github/workflows/*.yml", ".gitlab-ci.yml"}
	}
}

func containsAny(value string, items ...string) bool {
	for _, item := range items {
		if strings.Contains(value, item) {
			return true
		}
	}
	return false
}
