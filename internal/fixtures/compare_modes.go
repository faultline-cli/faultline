package fixtures

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ModeCompareResult holds the per-fixture rank and status changes between two
// evaluation runs (typically baseline vs Bayes-enabled).
type ModeCompareResult struct {
	FixtureID    string
	BaselineRank int
	BayesRank    int
	// Change is "improved", "regressed", or "unchanged".
	Change string
	// BaselineTop1 is true when the expected playbook was rank-1 in the baseline run.
	BaselineTop1 bool
	// BayesTop1 is true when the expected playbook was rank-1 in the Bayes run.
	BayesTop1 bool
}

// ModeComparison summarises the aggregate and per-fixture differences between a
// baseline and a Bayes-enabled evaluation over the same fixture corpus.
type ModeComparison struct {
	Class string

	BaselineTop1Rate float64
	BayesTop1Rate    float64
	BaselineTop3Rate float64
	BayesTop3Rate    float64

	Improved     int
	Regressed    int
	Unchanged    int
	FixtureCount int

	Regressions  []ModeCompareResult
	Improvements []ModeCompareResult
	Details      []ModeCompareResult
}

// Top1Delta returns the signed change in Top-1 rate (Bayes minus baseline).
func (c ModeComparison) Top1Delta() float64 {
	return round2(c.BayesTop1Rate - c.BaselineTop1Rate)
}

// Top3Delta returns the signed change in Top-3 rate (Bayes minus baseline).
func (c ModeComparison) Top3Delta() float64 {
	return round2(c.BayesTop3Rate - c.BaselineTop3Rate)
}

// HasRegressions returns true when Bayes mode regresses any fixture's rank.
func (c ModeComparison) HasRegressions() bool {
	return len(c.Regressions) > 0
}

// CompareReports compares a baseline (non-Bayes) report against a Bayes report
// over the same corpus and returns a structured comparison. Both reports must
// cover the same fixture class.
func CompareReports(baseline, bayes Report) (ModeComparison, error) {
	if baseline.Class != bayes.Class {
		return ModeComparison{}, fmt.Errorf("report class mismatch: baseline=%s bayes=%s", baseline.Class, bayes.Class)
	}

	baselineByID := make(map[string]EvaluatedFixture, len(baseline.Fixtures))
	for _, f := range baseline.Fixtures {
		baselineByID[f.Fixture.ID] = f
	}

	cmp := ModeComparison{
		Class:            string(baseline.Class),
		BaselineTop1Rate: baseline.Top1Rate(),
		BayesTop1Rate:    bayes.Top1Rate(),
		BaselineTop3Rate: baseline.Top3Rate(),
		BayesTop3Rate:    bayes.Top3Rate(),
		FixtureCount:     baseline.FixtureCount,
	}

	// Iterate over Bayes fixtures in deterministic order.
	seen := make([]ModeCompareResult, 0, len(bayes.Fixtures))
	for _, bf := range bayes.Fixtures {
		if strings.TrimSpace(bf.Fixture.Expectation.ExpectedPlaybook) == "" {
			continue
		}
		bl, ok := baselineByID[bf.Fixture.ID]
		if !ok {
			continue
		}
		result := ModeCompareResult{
			FixtureID:    bf.Fixture.ID,
			BaselineRank: bl.ExpectedRank,
			BayesRank:    bf.ExpectedRank,
			BaselineTop1: bl.ExpectedRank == 1,
			BayesTop1:    bf.ExpectedRank == 1,
		}
		switch {
		case bf.ExpectedRank > 0 && (bl.ExpectedRank == 0 || bf.ExpectedRank < bl.ExpectedRank):
			result.Change = "improved"
			cmp.Improved++
			cmp.Improvements = append(cmp.Improvements, result)
		case bl.ExpectedRank > 0 && (bf.ExpectedRank == 0 || bf.ExpectedRank > bl.ExpectedRank):
			result.Change = "regressed"
			cmp.Regressed++
			cmp.Regressions = append(cmp.Regressions, result)
		default:
			result.Change = "unchanged"
			cmp.Unchanged++
		}
		seen = append(seen, result)
	}

	sort.Slice(seen, func(i, j int) bool { return seen[i].FixtureID < seen[j].FixtureID })
	cmp.Details = seen
	sort.Slice(cmp.Regressions, func(i, j int) bool { return cmp.Regressions[i].FixtureID < cmp.Regressions[j].FixtureID })
	sort.Slice(cmp.Improvements, func(i, j int) bool { return cmp.Improvements[i].FixtureID < cmp.Improvements[j].FixtureID })
	return cmp, nil
}

// FormatModeComparison returns a human-readable or JSON representation of a
// ModeComparison result.
func FormatModeComparison(cmp ModeComparison, jsonOut bool) (string, error) {
	if jsonOut {
		data, err := json.MarshalIndent(cmp, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	}

	delta1 := cmp.Top1Delta()
	delta3 := cmp.Top3Delta()

	sign := func(v float64) string {
		if v > 0 {
			return fmt.Sprintf("+%.3f", v)
		}
		return fmt.Sprintf("%.3f", v)
	}

	lines := []string{
		fmt.Sprintf("class: %s", cmp.Class),
		fmt.Sprintf("fixtures: %d", cmp.FixtureCount),
		fmt.Sprintf("baseline top_1: %.3f  bayes top_1: %.3f  delta: %s", cmp.BaselineTop1Rate, cmp.BayesTop1Rate, sign(delta1)),
		fmt.Sprintf("baseline top_3: %.3f  bayes top_3: %.3f  delta: %s", cmp.BaselineTop3Rate, cmp.BayesTop3Rate, sign(delta3)),
		fmt.Sprintf("improved: %d  regressed: %d  unchanged: %d", cmp.Improved, cmp.Regressed, cmp.Unchanged),
	}
	if len(cmp.Regressions) > 0 {
		lines = append(lines, "regressions:")
		for _, r := range cmp.Regressions {
			baseRank := fmt.Sprintf("%d", r.BaselineRank)
			if r.BaselineRank == 0 {
				baseRank = "unmatched"
			}
			bayesRank := fmt.Sprintf("%d", r.BayesRank)
			if r.BayesRank == 0 {
				bayesRank = "unmatched"
			}
			lines = append(lines, fmt.Sprintf("  - %s: rank %s -> %s", r.FixtureID, baseRank, bayesRank))
		}
	}
	if len(cmp.Improvements) > 0 {
		lines = append(lines, "improvements:")
		for _, r := range cmp.Improvements {
			baseRank := fmt.Sprintf("%d", r.BaselineRank)
			if r.BaselineRank == 0 {
				baseRank = "unmatched"
			}
			bayesRank := fmt.Sprintf("%d", r.BayesRank)
			if r.BayesRank == 0 {
				bayesRank = "unmatched"
			}
			lines = append(lines, fmt.Sprintf("  + %s: rank %s -> %s", r.FixtureID, baseRank, bayesRank))
		}
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func round2(v float64) float64 {
	return float64(int(v*1000+0.5)) / 1000
}
