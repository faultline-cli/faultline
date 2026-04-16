package scoring

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"faultline/internal/model"
)

func Score(inputs Inputs) ([]model.Result, *model.Delta, error) {
	if len(inputs.Results) == 0 {
		return nil, nil, nil
	}
	weights, err := defaultWeights()
	if err != nil {
		return nil, nil, err
	}

	baseline := append([]model.Result(nil), inputs.Results...)
	delta := buildDelta(inputs.RepoState)
	version := weights.Version
	if version == "" {
		version = "bayes_v1"
	}

	results := append([]model.Result(nil), inputs.Results...)
	for i := range results {
		prior := blendedPrior(weights, results[i].Playbook, baseline)
		features := featureSet(weights, inputs, results[i], baseline, i, delta)
		ranking := rankingFromFeatures(version, results[i].Score, prior, features)
		results[i].Ranking = ranking
		results[i].Score = ranking.FinalScore
	}

	sort.Slice(results, func(i, j int) bool {
		a, b := results[i], results[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.Ranking != nil && b.Ranking != nil && a.Ranking.BaselineScore != b.Ranking.BaselineScore {
			return a.Ranking.BaselineScore > b.Ranking.BaselineScore
		}
		if a.Confidence != b.Confidence {
			return a.Confidence > b.Confidence
		}
		return a.Playbook.ID < b.Playbook.ID
	})
	recalibrateConfidence(results)
	return results, delta, nil
}

func recalibrateConfidence(results []model.Result) {
	if len(results) == 0 {
		return
	}
	topScore := results[0].Score
	for i := range results {
		next := 0.0
		if i+1 < len(results) {
			next = results[i+1].Score
		}
		coverage := clamp01(results[i].Score / (results[i].Score + 1))
		separation := clamp01((results[i].Score - next) / math.Max(results[i].Score, 1))
		base := results[i].Confidence
		results[i].Confidence = round((coverage + separation + base) / 3)
		if results[i].Score == topScore && next == 0 {
			results[i].Confidence = round(math.Max(results[i].Confidence, base))
		}
	}
}

func blendedPrior(weights weightsFile, pb model.Playbook, baseline []model.Result) float64 {
	if len(baseline) == 0 {
		return 0
	}
	totalCount := 0
	categoryCounts := map[string]int{}
	for id, count := range weights.PlaybookCounts {
		totalCount += count
		category := categoryForPlaybookID(id, baseline)
		if category != "" {
			categoryCounts[category] += count
		}
	}
	if totalCount == 0 {
		return 0
	}
	smoothing := weights.PriorSmoothing
	uniform := 1.0 / math.Max(float64(len(weights.PlaybookCounts)), 1)
	playbookProb := (float64(weights.PlaybookCounts[pb.ID]) + smoothing) / (float64(totalCount) + smoothing*float64(len(weights.PlaybookCounts)))
	categoryProb := uniform
	if categoryCount := categoryCounts[pb.Category]; categoryCount > 0 {
		categoryProb = (float64(categoryCount) + smoothing) / (float64(totalCount) + smoothing*float64(len(categoryCounts)))
	}
	playbookPrior := clamp(math.Log(playbookProb/uniform), -0.35, 0.35)
	categoryPrior := clamp(math.Log(categoryProb/uniform), -0.20, 0.20)
	return round((playbookPrior * 0.7) + (categoryPrior * 0.3))
}

func categoryForPlaybookID(id string, baseline []model.Result) string {
	for _, result := range baseline {
		if result.Playbook.ID == id {
			return result.Playbook.Category
		}
	}
	return ""
}

func round(value float64) float64 {
	return math.Round(value*100) / 100
}

func ratio(num, denom int) float64 {
	if denom <= 0 {
		return 0
	}
	return clamp01(float64(num) / float64(denom))
}

func clamp01(value float64) float64 {
	return clamp(value, 0, 1)
}

func clamp(value, lo, hi float64) float64 {
	if value < lo {
		return lo
	}
	if value > hi {
		return hi
	}
	return value
}

func normalizeText(value string) string {
	return strings.Join(tokenSet(value), " ")
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func DebugString(result model.Result) string {
	if result.Ranking == nil {
		return ""
	}
	return fmt.Sprintf("%s %.2f", result.Playbook.ID, result.Ranking.FinalScore)
}
