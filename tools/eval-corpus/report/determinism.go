package report

import (
	"fmt"
	"io"
	"sort"

	"faultline/tools/eval-corpus/model"
)

// DeterminismResult records the outcome of comparing two result files.
type DeterminismResult struct {
	// BaselineCount is the number of records in the first (reference) run.
	BaselineCount int `json:"baseline_count"`
	// CompareCount is the number of records in the second run.
	CompareCount int `json:"compare_count"`
	// Differences holds individual discrepancies found between the two files.
	Differences []Difference `json:"differences"`
	// Deterministic is true when no functional differences were found.
	Deterministic bool `json:"deterministic"`
}

// Difference describes a single mismatch between two EvalResult records.
type Difference struct {
	FixtureID string `json:"fixture_id"`
	Field     string `json:"field"`
	Baseline  any    `json:"baseline"`
	Compare   any    `json:"compare"`
}

// CheckDeterminism compares two result sets by FixtureID.
// Timing fields (DurationMS) are excluded from the comparison.
// Both slices are sorted by FixtureID before comparison.
func CheckDeterminism(baseline, compare []model.EvalResult) DeterminismResult {
	sort.Slice(baseline, func(i, j int) bool {
		return baseline[i].FixtureID < baseline[j].FixtureID
	})
	sort.Slice(compare, func(i, j int) bool {
		return compare[i].FixtureID < compare[j].FixtureID
	})

	res := DeterminismResult{
		BaselineCount: len(baseline),
		CompareCount:  len(compare),
	}

	// Index compare by FixtureID for lookup.
	compareIdx := make(map[string]model.EvalResult, len(compare))
	for _, r := range compare {
		compareIdx[r.FixtureID] = r
	}

	for _, b := range baseline {
		c, ok := compareIdx[b.FixtureID]
		if !ok {
			res.Differences = append(res.Differences, Difference{
				FixtureID: b.FixtureID,
				Field:     "fixture_id",
				Baseline:  "present",
				Compare:   "missing",
			})
			continue
		}
		res.Differences = append(res.Differences, diffResults(b, c)...)
	}

	// Report fixtures present in compare but not in baseline.
	baselineIdx := make(map[string]struct{}, len(baseline))
	for _, r := range baseline {
		baselineIdx[r.FixtureID] = struct{}{}
	}
	for _, r := range compare {
		if _, ok := baselineIdx[r.FixtureID]; !ok {
			res.Differences = append(res.Differences, Difference{
				FixtureID: r.FixtureID,
				Field:     "fixture_id",
				Baseline:  "missing",
				Compare:   "present",
			})
		}
	}

	res.Deterministic = len(res.Differences) == 0
	return res
}

// diffResults returns Difference entries for all functional fields that differ
// between two EvalResult records with the same FixtureID. DurationMS is excluded.
func diffResults(b, c model.EvalResult) []Difference {
	var diffs []Difference
	add := func(field string, bv, cv any) {
		diffs = append(diffs, Difference{FixtureID: b.FixtureID, Field: field, Baseline: bv, Compare: cv})
	}

	if b.Matched != c.Matched {
		add("matched", b.Matched, c.Matched)
	}
	if b.FailureID != c.FailureID {
		add("failure_id", b.FailureID, c.FailureID)
	}
	if b.Confidence != c.Confidence {
		add("confidence", b.Confidence, c.Confidence)
	}
	if b.Error != c.Error {
		add("error", b.Error, c.Error)
	}
	if !equalStrSlice(b.Evidence, c.Evidence) {
		add("evidence", b.Evidence, c.Evidence)
	}
	return diffs
}

func equalStrSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// PrintDeterminismText writes a determinism check summary to w.
func PrintDeterminismText(w io.Writer, dr DeterminismResult) {
	fmt.Fprintln(w, "Determinism Check")
	fmt.Fprintln(w, "=================")
	fmt.Fprintf(w, "Baseline : %d records\n", dr.BaselineCount)
	fmt.Fprintf(w, "Compare  : %d records\n", dr.CompareCount)

	if dr.Deterministic {
		fmt.Fprintln(w, "Result   : DETERMINISTIC (no functional differences)")
		return
	}

	fmt.Fprintf(w, "Result   : NON-DETERMINISTIC (%d differences)\n", len(dr.Differences))
	limit := 20
	for i, d := range dr.Differences {
		if i >= limit {
			fmt.Fprintf(w, "  ... and %d more\n", len(dr.Differences)-limit)
			break
		}
		fmt.Fprintf(w, "  [%s] %s: %v → %v\n", d.FixtureID[:min(12, len(d.FixtureID))], d.Field, d.Baseline, d.Compare)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
