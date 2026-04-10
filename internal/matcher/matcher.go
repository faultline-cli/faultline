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
func Rank(playbooks []model.Playbook, lines []model.Line, ctx model.Context) []model.Result {
	results := make([]model.Result, 0, len(playbooks))
	for _, pb := range playbooks {
		r := matchPlaybook(pb, lines, ctx)
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
//   - Each matched any-pattern  → +1.0
//   - Each matched all-pattern  → +1.5
//   - All all-patterns matched  → +2.0 bonus
//   - Stage hint matches ctx    → +0.75
//   - Playbook base_score       → added unconditionally when patterns match
//
// Confidence reflects pattern coverage only (stage bonus excluded) and is
// rounded to two decimal places.
func matchPlaybook(pb model.Playbook, lines []model.Line, ctx model.Context) model.Result {
	evidence := make([]string, 0)
	seenEvidence := make(map[string]struct{})

	addEvidence := func(line string) {
		if _, ok := seenEvidence[line]; !ok {
			evidence = append(evidence, line)
			seenEvidence[line] = struct{}{}
		}
	}

	// Score any-patterns (OR semantics).
	anyHits := 0
	for _, pat := range pb.Match.Any {
		norm := normalize(pat)
		if norm == "" {
			continue
		}
		for _, line := range lines {
			if strings.Contains(line.Normalized, norm) {
				anyHits++
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

	if anyHits == 0 && allHits == 0 {
		return model.Result{} // no match
	}

	score := pb.BaseScore +
		float64(anyHits)*1.0 +
		float64(allHits)*1.5
	if allComplete {
		score += 2.0
	}

	// Stage bonus (does not contribute to confidence calculation).
	if ctx.Stage != "" {
		for _, hint := range pb.StageHints {
			if strings.EqualFold(hint, ctx.Stage) {
				score += 0.75
				break
			}
		}
	}

	// Confidence: normalised pattern coverage, capped at 1.0.
	// Computed without the situational stage bonus.
	maxScore := pb.BaseScore +
		float64(len(pb.Match.Any))*1.0 +
		float64(len(pb.Match.All))*1.5
	if len(pb.Match.All) > 0 {
		maxScore += 2.0
	}
	if maxScore < 1.0 {
		maxScore = 1.0
	}
	patternScore := pb.BaseScore +
		float64(anyHits)*1.0 +
		float64(allHits)*1.5
	if allComplete {
		patternScore += 2.0
	}
	confidence := math.Round(math.Min(1.0, patternScore/maxScore)*100) / 100

	return model.Result{
		Playbook:   pb,
		Score:      math.Round(score*100) / 100,
		Confidence: confidence,
		Evidence:   evidence,
	}
}

// normalize lower-cases s and collapses internal whitespace.
func normalize(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(s)), " "))
}
