// Package matcher implements deterministic pattern matching and scoring of
// playbooks against normalised log lines.
package matcher

import (
	"math"
	"sort"
	"strings"

	"faultline/internal/model"
)

// Rank matches every playbook against lines and returns results sorted by
// score descending, then confidence descending, then playbook ID ascending.
// Only playbooks with at least one matching pattern are included.
//
// match.any patterns are weighted by inverse document frequency: a pattern
// shared by N playbooks contributes 1/N to the score, so generic terms that
// appear across many playbooks are automatically less decisive than patterns
// unique to a single playbook. match.all and match.none semantics are unchanged.
func Rank(playbooks []model.Playbook, lines []model.Line, ctx model.Context) []model.Result {
	weights := computeAnyWeights(playbooks)
	results := make([]model.Result, 0, len(playbooks))
	for _, pb := range playbooks {
		r := matchPlaybook(pb, lines, ctx, weights)
		if r.Score == 0 {
			continue
		}
		results = append(results, r)
	}

	sort.Slice(results, func(i, j int) bool {
		a, b := results[i], results[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		return a.Playbook.ID < b.Playbook.ID
	})
	calibrateConfidence(results)

	sort.Slice(results, func(i, j int) bool {
		a, b := results[i], results[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.Confidence != b.Confidence {
			return a.Confidence > b.Confidence
		}
		return a.Playbook.ID < b.Playbook.ID
	})

	return results
}

// matchPlaybook scores a single playbook against lines using the following
// rules:
//
//   - Each matched any-pattern  → +weight (1/N where N = playbooks sharing that pattern)
//   - Each matched all-pattern  → +1.5 (flat; AND semantics already discriminate)
//   - All all-patterns matched  → +2.0 bonus
//   - Stage hint matches ctx    → +0.75
//   - Playbook base_score       → added unconditionally when patterns match
//
// Confidence is calibrated from the matched score and competitive separation
// after the full ranked result set is known.
func matchPlaybook(pb model.Playbook, lines []model.Line, ctx model.Context, weights map[string]float64) model.Result {
	evidence := make([]string, 0)
	seenEvidence := make(map[string]struct{})

	addEvidence := func(line string) {
		if _, ok := seenEvidence[line]; !ok {
			evidence = append(evidence, line)
			seenEvidence[line] = struct{}{}
		}
	}

	// Score any-patterns (OR semantics, IDF-weighted).
	anyScore := 0.0
	for _, pat := range pb.Match.Any {
		norm := normalize(pat)
		if norm == "" {
			continue
		}
		w := weights[norm]
		if w == 0 {
			w = 1.0
		}
		for _, line := range lines {
			if strings.Contains(line.Normalized, norm) {
				anyScore += w
				addEvidence(line.Original)
				break
			}
		}
	}

	// Score all-patterns (AND semantics; partial matches still accumulate).
	allHits := 0
	allComplete := len(pb.Match.All) > 0
	for _, pat := range pb.Match.All {
		norm := normalize(pat)
		if norm == "" {
			continue
		}
		matched := false
		for _, line := range lines {
			if strings.Contains(line.Normalized, norm) {
				allHits++
				addEvidence(line.Original)
				matched = true
				break
			}
		}
		if !matched {
			allComplete = false
		}
	}

	for _, pat := range pb.Match.None {
		norm := normalize(pat)
		if norm == "" {
			continue
		}
		for _, line := range lines {
			if strings.Contains(line.Normalized, norm) {
				return model.Result{}
			}
		}
	}

	if anyScore == 0 && allHits == 0 {
		return model.Result{} // no match
	}

	score := pb.BaseScore + anyScore + float64(allHits)*1.5
	if allComplete {
		score += compoundBonus(allComplete)
	}

	// Stage bonus (does not contribute to confidence calculation).
	score += stageBonus(pb, ctx)

	return model.Result{
		Playbook:   pb,
		Detector:   "log",
		Score:      math.Round(score*100) / 100,
		Confidence: 0,
		Evidence:   evidence,
		EvidenceBy: model.EvidenceBundle{
			Triggers: buildLogEvidence(evidence),
		},
		Explanation: model.ResultExplanation{
			TriggeredBy: evidence,
		},
		Breakdown: model.ScoreBreakdown{
			BaseSignalScore:     math.Round((pb.BaseScore+anyScore+float64(allHits)*1.5)*100) / 100,
			CompoundSignalBonus: math.Round(compoundBonus(allComplete)*100) / 100,
			HotPathMultiplier:   math.Round(stageBonus(pb, ctx)*100) / 100,
			FinalScore:          math.Round(score*100) / 100,
		},
	}
}

func calibrateConfidence(results []model.Result) {
	if len(results) == 0 {
		return
	}

	topScore := results[0].Score
	topScoreCount := 0
	secondScore := 0.0
	for _, result := range results {
		if result.Score == topScore {
			topScoreCount++
			continue
		}
		secondScore = result.Score
		break
	}

	for i := range results {
		competitorScore := topScore
		if results[i].Score == topScore && topScoreCount == 1 {
			competitorScore = secondScore
		}
		results[i].Confidence = confidenceFromScores(results[i].Score, competitorScore)
	}
}

func confidenceFromScores(score, competitorScore float64) float64 {
	if score <= 0 {
		return 0
	}

	coverage := score / (score + 1)
	separation := 0.0
	if competitorScore < score {
		separation = 1 - (competitorScore / score)
	}

	return math.Round(((coverage+separation)/2)*100) / 100
}

func buildLogEvidence(lines []string) []model.Evidence {
	out := make([]model.Evidence, 0, len(lines))
	for _, line := range lines {
		out = append(out, model.Evidence{
			Kind:   model.EvidenceTrigger,
			Label:  "Matched log evidence",
			Detail: line,
			Source: "log",
		})
	}
	return out
}

func compoundBonus(allComplete bool) float64 {
	if allComplete {
		return 2.0
	}
	return 0
}

func stageBonus(pb model.Playbook, ctx model.Context) float64 {
	if ctx.Stage == "" {
		return 0
	}
	for _, hint := range pb.StageHints {
		if strings.EqualFold(hint, ctx.Stage) {
			return 0.75
		}
	}
	return 0
}

// computeAnyWeights returns a map from normalized pattern string to its IDF
// weight: 1.0 / count, where count is the number of distinct playbooks that
// list that pattern in match.any. Patterns unique to one playbook keep
// weight 1.0; patterns shared by N playbooks each contribute 1/N.
func computeAnyWeights(playbooks []model.Playbook) map[string]float64 {
	counts := make(map[string]int, len(playbooks)*8)
	for _, pb := range playbooks {
		seen := make(map[string]struct{}, len(pb.Match.Any))
		for _, p := range pb.Match.Any {
			n := normalize(p)
			if n == "" {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			counts[n]++
		}
	}
	weights := make(map[string]float64, len(counts))
	for p, c := range counts {
		weights[p] = 1.0 / float64(c)
	}
	return weights
}

// normalize lower-cases s and collapses internal whitespace.
func normalize(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(s)), " "))
}
