// Package compare implements the CI regression gate logic for faultline-eval.
// It compares two result sets (baseline vs current) and produces a structured
// comparison report with a pass/fail decision.
package compare

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"faultline/tools/eval-corpus/model"
	"faultline/tools/eval-corpus/report"
)

// Options controls the comparison thresholds and gate behaviour.
type Options struct {
	// MinMatchRate is the minimum acceptable current match rate (0–1).
	// A value of 0 disables this gate.
	MinMatchRate float64
	// FailOnNewNondeterminism causes a gate failure when any fixture that was
	// deterministic in the baseline shows non-determinism in the current run.
	FailOnNewNondeterminism bool
	// FailOnCoverageDrop causes a gate failure when the current match rate is
	// lower than the baseline match rate.
	FailOnCoverageDrop bool
}

// Result is the full structured output of a comparison.
type Result struct {
	// GeneratedAt is an RFC-3339 timestamp for the comparison run.
	GeneratedAt string `json:"generated_at"`

	BaselineMatchRate float64 `json:"baseline_match_rate"`
	CurrentMatchRate  float64 `json:"current_match_rate"`
	Delta             float64 `json:"delta"`

	BaselineTotal int `json:"baseline_total"`
	CurrentTotal  int `json:"current_total"`

	// NewUnmatchedClusters lists first-line tags present in current unmatched
	// results but absent from baseline unmatched results.
	NewUnmatchedTags []string `json:"new_unmatched_tags,omitempty"`

	// PlaybooksLostMatches lists failure IDs that had hits in baseline but
	// fewer (or zero) hits in current.
	PlaybooksLostMatches []PlaybookDelta `json:"playbooks_lost_matches,omitempty"`
	// PlaybooksGainedMatches lists failure IDs that have more hits in current
	// than in baseline.
	PlaybooksGainedMatches []PlaybookDelta `json:"playbooks_gained_matches,omitempty"`

	// NondeterministicFixtureIDs are fixtures whose results differed between
	// the two runs (only populated when a second determinism run is available
	// via the CheckDeterminism path in report).
	NondeterministicFixtureIDs []string `json:"nondeterministic_fixture_ids,omitempty"`

	// Pass is true when all enabled gate conditions are satisfied.
	Pass bool `json:"pass"`
	// FailReasons enumerates each failed gate condition.
	FailReasons []string `json:"fail_reasons,omitempty"`
}

// PlaybookDelta records the change in match count for a single failure ID.
type PlaybookDelta struct {
	FailureID     string `json:"failure_id"`
	BaselineCount int    `json:"baseline_count"`
	CurrentCount  int    `json:"current_count"`
	Delta         int    `json:"delta"`
}

// Compare computes a comparison Result between baseline and current result sets.
func Compare(baseline, current []model.EvalResult, opts Options) Result {
	baseRate, baseTotal := matchRate(baseline)
	currRate, currTotal := matchRate(current)

	delta := currRate - baseRate

	newTags := newUnmatchedTags(baseline, current)
	lost, gained := playbookDeltas(baseline, current)

	res := Result{
		GeneratedAt:            time.Now().UTC().Format(time.RFC3339),
		BaselineMatchRate:      baseRate,
		CurrentMatchRate:       currRate,
		Delta:                  delta,
		BaselineTotal:          baseTotal,
		CurrentTotal:           currTotal,
		NewUnmatchedTags:       newTags,
		PlaybooksLostMatches:   lost,
		PlaybooksGainedMatches: gained,
		Pass:                   true,
	}

	// Evaluate gate conditions.
	if opts.MinMatchRate > 0 && currRate < opts.MinMatchRate {
		res.Pass = false
		res.FailReasons = append(res.FailReasons,
			fmt.Sprintf("match rate %.4f is below minimum %.4f", currRate, opts.MinMatchRate))
	}
	if opts.FailOnCoverageDrop && delta < 0 {
		res.Pass = false
		res.FailReasons = append(res.FailReasons,
			fmt.Sprintf("coverage dropped by %.4f (%.4f → %.4f)", -delta, baseRate, currRate))
	}

	return res
}

// AttachNondeterminism populates NondeterministicFixtureIDs from a determinism
// check result and evaluates the fail-on-new-nondeterminism gate condition.
func AttachNondeterminism(res *Result, det report.DeterminismResult, opts Options) {
	for _, d := range det.Differences {
		if d.Field != "fixture_id" { // skip missing/extra fixture markers
			res.NondeterministicFixtureIDs = appendUniq(res.NondeterministicFixtureIDs, d.FixtureID)
		}
	}
	if opts.FailOnNewNondeterminism && len(res.NondeterministicFixtureIDs) > 0 {
		res.Pass = false
		res.FailReasons = append(res.FailReasons,
			fmt.Sprintf("%d non-deterministic fixture(s) detected", len(res.NondeterministicFixtureIDs)))
	}
}

// PrintTextReport writes a human-readable comparison report to w.
func PrintTextReport(w io.Writer, res Result) {
	fmt.Fprintln(w, "Comparison Report")
	fmt.Fprintln(w, "=================")
	fmt.Fprintf(w, "Generated     : %s\n", res.GeneratedAt)
	fmt.Fprintf(w, "Baseline total: %d  (match rate: %.2f%%)\n", res.BaselineTotal, res.BaselineMatchRate*100)
	fmt.Fprintf(w, "Current total : %d  (match rate: %.2f%%)\n", res.CurrentTotal, res.CurrentMatchRate*100)
	delta := res.Delta * 100
	sign := "+"
	if delta < 0 {
		sign = ""
	}
	fmt.Fprintf(w, "Delta         : %s%.4f%%\n", sign, delta)
	fmt.Fprintln(w)

	if len(res.PlaybooksLostMatches) > 0 {
		fmt.Fprintln(w, "Playbooks that lost matches")
		fmt.Fprintln(w, "--------------------------")
		for _, d := range res.PlaybooksLostMatches {
			fmt.Fprintf(w, "  %-40s  %d → %d  (Δ%d)\n", d.FailureID, d.BaselineCount, d.CurrentCount, d.Delta)
		}
		fmt.Fprintln(w)
	}

	if len(res.PlaybooksGainedMatches) > 0 {
		fmt.Fprintln(w, "Playbooks that gained matches")
		fmt.Fprintln(w, "----------------------------")
		for _, d := range res.PlaybooksGainedMatches {
			fmt.Fprintf(w, "  %-40s  %d → %d  (+%d)\n", d.FailureID, d.BaselineCount, d.CurrentCount, d.Delta)
		}
		fmt.Fprintln(w)
	}

	if len(res.NewUnmatchedTags) > 0 {
		fmt.Fprintln(w, "New unmatched clusters")
		fmt.Fprintln(w, "----------------------")
		for _, t := range res.NewUnmatchedTags {
			fmt.Fprintf(w, "  %s\n", t)
		}
		fmt.Fprintln(w)
	}

	if len(res.NondeterministicFixtureIDs) > 0 {
		fmt.Fprintln(w, "Non-deterministic fixtures")
		fmt.Fprintln(w, "--------------------------")
		for _, id := range res.NondeterministicFixtureIDs {
			fmt.Fprintf(w, "  %s\n", id)
		}
		fmt.Fprintln(w)
	}

	if res.Pass {
		fmt.Fprintln(w, "Gate: PASS")
	} else {
		fmt.Fprintln(w, "Gate: FAIL")
		for _, r := range res.FailReasons {
			fmt.Fprintf(w, "  - %s\n", r)
		}
	}
}

// PrintMarkdownReport writes a Markdown comparison report to w.
func PrintMarkdownReport(w io.Writer, res Result) {
	status := "✅ PASS"
	if !res.Pass {
		status = "❌ FAIL"
	}
	fmt.Fprintf(w, "# CI Comparison Report — %s\n\n", status)
	fmt.Fprintf(w, "_Generated: %s_\n\n", res.GeneratedAt)
	fmt.Fprintln(w, "## Coverage")
	fmt.Fprintf(w, "| | Baseline | Current | Delta |\n|---|---|---|---|\n")
	fmt.Fprintf(w, "| Total | %d | %d | %+d |\n", res.BaselineTotal, res.CurrentTotal, res.CurrentTotal-res.BaselineTotal)
	fmt.Fprintf(w, "| Match rate | %.2f%% | %.2f%% | %+.4f%% |\n\n",
		res.BaselineMatchRate*100, res.CurrentMatchRate*100, res.Delta*100)

	if len(res.PlaybooksLostMatches) > 0 {
		fmt.Fprintf(w, "## Playbooks That Lost Matches\n\n")
		fmt.Fprintln(w, "| Failure ID | Baseline | Current | Δ |")
		fmt.Fprintln(w, "|------------|:--------:|:-------:|:-:|")
		for _, d := range res.PlaybooksLostMatches {
			fmt.Fprintf(w, "| `%s` | %d | %d | %d |\n", d.FailureID, d.BaselineCount, d.CurrentCount, d.Delta)
		}
		fmt.Fprintln(w)
	}

	if len(res.PlaybooksGainedMatches) > 0 {
		fmt.Fprintf(w, "## Playbooks That Gained Matches\n\n")
		fmt.Fprintln(w, "| Failure ID | Baseline | Current | Δ |")
		fmt.Fprintln(w, "|------------|:--------:|:-------:|:-:|")
		for _, d := range res.PlaybooksGainedMatches {
			fmt.Fprintf(w, "| `%s` | %d | %d | +%d |\n", d.FailureID, d.BaselineCount, d.CurrentCount, d.Delta)
		}
		fmt.Fprintln(w)
	}

	if !res.Pass {
		fmt.Fprintf(w, "## Gate Failures\n\n")
		for _, r := range res.FailReasons {
			fmt.Fprintf(w, "- %s\n", r)
		}
		fmt.Fprintln(w)
	}
}

// WriteJSON serialises res to w as indented JSON.
func WriteJSON(w io.Writer, res Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}

// --- helpers ---

func matchRate(results []model.EvalResult) (rate float64, total int) {
	total = len(results)
	if total == 0 {
		return 0, 0
	}
	matched := 0
	for _, r := range results {
		if r.Matched {
			matched++
		}
	}
	return float64(matched) / float64(total), total
}

func newUnmatchedTags(baseline, current []model.EvalResult) []string {
	baseTags := map[string]struct{}{}
	for _, r := range baseline {
		if !r.Matched && r.FirstLineTag != "" {
			baseTags[r.FirstLineTag] = struct{}{}
		}
	}
	seen := map[string]struct{}{}
	var out []string
	for _, r := range current {
		if !r.Matched && r.FirstLineTag != "" {
			if _, inBase := baseTags[r.FirstLineTag]; !inBase {
				if _, done := seen[r.FirstLineTag]; !done {
					seen[r.FirstLineTag] = struct{}{}
					out = append(out, r.FirstLineTag)
				}
			}
		}
	}
	sort.Strings(out)
	return out
}

func playbookDeltas(baseline, current []model.EvalResult) (lost, gained []PlaybookDelta) {
	baseCount := map[string]int{}
	for _, r := range baseline {
		if r.Matched && r.FailureID != "" {
			baseCount[r.FailureID]++
		}
	}
	currCount := map[string]int{}
	for _, r := range current {
		if r.Matched && r.FailureID != "" {
			currCount[r.FailureID]++
		}
	}

	// Collect all failure IDs from both runs.
	allIDs := map[string]struct{}{}
	for id := range baseCount {
		allIDs[id] = struct{}{}
	}
	for id := range currCount {
		allIDs[id] = struct{}{}
	}

	ids := make([]string, 0, len(allIDs))
	for id := range allIDs {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		b := baseCount[id]
		c := currCount[id]
		d := c - b
		if d < 0 {
			lost = append(lost, PlaybookDelta{FailureID: id, BaselineCount: b, CurrentCount: c, Delta: d})
		} else if d > 0 {
			gained = append(gained, PlaybookDelta{FailureID: id, BaselineCount: b, CurrentCount: c, Delta: d})
		}
	}
	return lost, gained
}

func appendUniq(s []string, v string) []string {
	for _, x := range s {
		if x == v {
			return s
		}
	}
	return append(s, v)
}
