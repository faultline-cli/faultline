package scoring

import (
	"math"
	"sort"

	"faultline/internal/model"
)

func rankingFromFeatures(version string, baselineScore, prior float64, features []feature) *model.Ranking {
	contributions := make([]model.RankingContribution, 0, len(features))
	finalScore := baselineScore + prior
	for _, item := range features {
		contribution := round(item.Value * item.Weight)
		if contribution == 0 {
			continue
		}
		finalScore += contribution
		direction := "positive"
		if contribution < 0 {
			direction = "negative"
		}
		contributions = append(contributions, model.RankingContribution{
			Feature:      item.Name,
			Value:        round(item.Value),
			Weight:       round(item.Weight),
			Contribution: contribution,
			Direction:    direction,
			Reason:       item.Reason,
			EvidenceRefs: append([]string(nil), item.EvidenceRefs...),
		})
	}
	sort.Slice(contributions, func(i, j int) bool {
		ai := math.Abs(contributions[i].Contribution)
		aj := math.Abs(contributions[j].Contribution)
		if ai != aj {
			return ai > aj
		}
		return contributions[i].Feature < contributions[j].Feature
	})
	return &model.Ranking{
		Mode:              ModeBayes,
		Version:           version,
		BaselineScore:     round(baselineScore),
		Prior:             round(prior),
		FinalScore:        round(finalScore),
		Contributions:     contributions,
		StrongestPositive: strongestReasons(contributions, "positive"),
		StrongestNegative: strongestReasons(contributions, "negative"),
	}
}

func strongestReasons(contributions []model.RankingContribution, direction string) []string {
	var out []string
	for _, item := range contributions {
		if item.Direction != direction {
			continue
		}
		out = append(out, item.Reason)
		if len(out) >= 3 {
			break
		}
	}
	return dedupeStrings(out)
}
