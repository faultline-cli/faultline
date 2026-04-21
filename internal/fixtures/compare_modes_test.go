package fixtures

import (
	"strings"
	"testing"
)

func TestCompareReportsSameResultsAllUnchanged(t *testing.T) {
	// Both runs produce identical ranks — everything should be unchanged.
	baseline := Report{
		Class:        ClassReal,
		FixtureCount: 2,
		Top1Count:    2,
		Top3Count:    2,
		Fixtures: []EvaluatedFixture{
			{Fixture: Fixture{ID: "fx-a", Expectation: Expectation{ExpectedPlaybook: "p1"}}, ExpectedRank: 1},
			{Fixture: Fixture{ID: "fx-b", Expectation: Expectation{ExpectedPlaybook: "p2"}}, ExpectedRank: 2},
		},
	}
	bayes := Report{
		Class:        ClassReal,
		FixtureCount: 2,
		Top1Count:    2,
		Top3Count:    2,
		Fixtures: []EvaluatedFixture{
			{Fixture: Fixture{ID: "fx-a", Expectation: Expectation{ExpectedPlaybook: "p1"}}, ExpectedRank: 1},
			{Fixture: Fixture{ID: "fx-b", Expectation: Expectation{ExpectedPlaybook: "p2"}}, ExpectedRank: 2},
		},
	}

	cmp, err := CompareReports(baseline, bayes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmp.Improved != 0 || cmp.Regressed != 0 || cmp.Unchanged != 2 {
		t.Fatalf("expected 0 improved, 0 regressed, 2 unchanged — got %+v", cmp)
	}
	if cmp.HasRegressions() {
		t.Fatal("expected no regressions")
	}
}

func TestCompareReportsDetectsRegression(t *testing.T) {
	// Bayes drops fx-a from rank 1 to rank 3.
	baseline := Report{
		Class:        ClassReal,
		FixtureCount: 1,
		Top1Count:    1,
		Top3Count:    1,
		Fixtures: []EvaluatedFixture{
			{Fixture: Fixture{ID: "fx-a", Expectation: Expectation{ExpectedPlaybook: "p1"}}, ExpectedRank: 1},
		},
	}
	bayes := Report{
		Class:        ClassReal,
		FixtureCount: 1,
		Top1Count:    0,
		Top3Count:    1,
		Fixtures: []EvaluatedFixture{
			{Fixture: Fixture{ID: "fx-a", Expectation: Expectation{ExpectedPlaybook: "p1"}}, ExpectedRank: 3},
		},
	}

	cmp, err := CompareReports(baseline, bayes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmp.Regressed != 1 || !cmp.HasRegressions() {
		t.Fatalf("expected 1 regression, got %+v", cmp)
	}
	if cmp.Regressions[0].FixtureID != "fx-a" || cmp.Regressions[0].BaselineRank != 1 || cmp.Regressions[0].BayesRank != 3 {
		t.Fatalf("unexpected regression detail: %+v", cmp.Regressions[0])
	}
}

func TestCompareReportsDetectsImprovement(t *testing.T) {
	// Bayes promotes fx-b from rank 3 to rank 1.
	baseline := Report{
		Class:        ClassReal,
		FixtureCount: 1,
		Top1Count:    0,
		Top3Count:    1,
		Fixtures: []EvaluatedFixture{
			{Fixture: Fixture{ID: "fx-b", Expectation: Expectation{ExpectedPlaybook: "p2"}}, ExpectedRank: 3},
		},
	}
	bayes := Report{
		Class:        ClassReal,
		FixtureCount: 1,
		Top1Count:    1,
		Top3Count:    1,
		Fixtures: []EvaluatedFixture{
			{Fixture: Fixture{ID: "fx-b", Expectation: Expectation{ExpectedPlaybook: "p2"}}, ExpectedRank: 1},
		},
	}

	cmp, err := CompareReports(baseline, bayes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmp.Improved != 1 || cmp.Regressed != 0 {
		t.Fatalf("expected 1 improvement, got %+v", cmp)
	}
	if cmp.Improvements[0].FixtureID != "fx-b" {
		t.Fatalf("unexpected improvement detail: %+v", cmp.Improvements[0])
	}
}

func TestCompareReportsClassMismatchErrors(t *testing.T) {
	baseline := Report{Class: ClassReal}
	bayes := Report{Class: ClassMinimal}
	_, err := CompareReports(baseline, bayes)
	if err == nil || !strings.Contains(err.Error(), "class mismatch") {
		t.Fatalf("expected class mismatch error, got %v", err)
	}
}

func TestFormatModeComparisonText(t *testing.T) {
	cmp := ModeComparison{
		Class:            "real",
		FixtureCount:     3,
		BaselineTop1Rate: 0.667,
		BayesTop1Rate:    1.0,
		BaselineTop3Rate: 1.0,
		BayesTop3Rate:    1.0,
		Improved:         1,
		Regressed:        0,
		Unchanged:        2,
		Improvements: []ModeCompareResult{
			{FixtureID: "fx-a", BaselineRank: 2, BayesRank: 1, Change: "improved"},
		},
	}
	text, err := FormatModeComparison(cmp, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{
		"class: real",
		"fixtures: 3",
		"improved: 1  regressed: 0  unchanged: 2",
		"improvements:",
		"+ fx-a: rank 2 -> 1",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in output:\n%s", want, text)
		}
	}
}

func TestFormatModeComparisonTextRegression(t *testing.T) {
	cmp := ModeComparison{
		Class:        "real",
		FixtureCount: 1,
		Regressed:    1,
		Regressions: []ModeCompareResult{
			{FixtureID: "fx-z", BaselineRank: 1, BayesRank: 0, Change: "regressed"},
		},
	}
	text, err := FormatModeComparison(cmp, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text, "regressions:") || !strings.Contains(text, "fx-z") || !strings.Contains(text, "unmatched") {
		t.Fatalf("expected regression detail in output:\n%s", text)
	}
}

func TestFormatModeComparisonJSON(t *testing.T) {
	cmp := ModeComparison{
		Class:        "real",
		FixtureCount: 2,
		Improved:     1,
	}
	jsonText, err := FormatModeComparison(cmp, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{`"Class"`, `"real"`, `"FixtureCount"`, `"Improved"`} {
		if !strings.Contains(jsonText, want) {
			t.Fatalf("expected %q in JSON output:\n%s", want, jsonText)
		}
	}
}

func TestTop1DeltaSign(t *testing.T) {
	cmp := ModeComparison{BaselineTop1Rate: 0.9, BayesTop1Rate: 0.95}
	delta := cmp.Top1Delta()
	if delta <= 0 {
		t.Fatalf("expected positive delta, got %f", delta)
	}

	cmpNeg := ModeComparison{BaselineTop1Rate: 0.95, BayesTop1Rate: 0.9}
	deltaNeg := cmpNeg.Top1Delta()
	if deltaNeg >= 0 {
		t.Fatalf("expected negative delta, got %f", deltaNeg)
	}
}
