package report_test

import (
	"bytes"
	"strings"
	"testing"

	"faultline/tools/eval-corpus/model"
	"faultline/tools/eval-corpus/report"
)

func TestCheckDeterminismIdentical(t *testing.T) {
	results := []model.EvalResult{
		{FixtureID: "aaa", Matched: true, FailureID: "docker-auth", Confidence: 0.9},
		{FixtureID: "bbb", Matched: false},
	}
	// deep copy for second run
	results2 := make([]model.EvalResult, len(results))
	copy(results2, results)

	dr := report.CheckDeterminism(results, results2)
	if !dr.Deterministic {
		t.Errorf("expected deterministic, got %d differences", len(dr.Differences))
		for _, d := range dr.Differences {
			t.Logf("  diff: %+v", d)
		}
	}
}

func TestCheckDeterminismMatchedFlip(t *testing.T) {
	baseline := []model.EvalResult{
		{FixtureID: "aaa", Matched: true, FailureID: "docker-auth"},
	}
	compare := []model.EvalResult{
		{FixtureID: "aaa", Matched: false, FailureID: ""},
	}

	dr := report.CheckDeterminism(baseline, compare)
	if dr.Deterministic {
		t.Error("expected non-deterministic result")
	}
	found := false
	for _, d := range dr.Differences {
		if d.Field == "matched" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'matched' field difference to be reported")
	}
}

func TestCheckDeterminismMissingFixture(t *testing.T) {
	baseline := []model.EvalResult{
		{FixtureID: "aaa", Matched: true},
		{FixtureID: "bbb", Matched: false},
	}
	compare := []model.EvalResult{
		{FixtureID: "aaa", Matched: true},
		// bbb is missing
	}

	dr := report.CheckDeterminism(baseline, compare)
	if dr.Deterministic {
		t.Error("expected non-deterministic result due to missing fixture")
	}
}

func TestCheckDeterminismIgnoresDurationMS(t *testing.T) {
	baseline := []model.EvalResult{
		{FixtureID: "aaa", Matched: true, FailureID: "docker-auth", DurationMS: 10},
	}
	compare := []model.EvalResult{
		{FixtureID: "aaa", Matched: true, FailureID: "docker-auth", DurationMS: 999},
	}

	dr := report.CheckDeterminism(baseline, compare)
	if !dr.Deterministic {
		t.Errorf("DurationMS should be excluded from determinism check, got diffs: %v", dr.Differences)
	}
}

// --- PrintDeterminismText ---

func TestPrintDeterminismTextDeterministic(t *testing.T) {
	dr := report.DeterminismResult{
		Deterministic: true,
		BaselineCount: 10,
		CompareCount:  10,
	}
	var buf bytes.Buffer
	report.PrintDeterminismText(&buf, dr)
	out := buf.String()
	if !strings.Contains(out, "DETERMINISTIC") {
		t.Errorf("expected DETERMINISTIC in output:\n%s", out)
	}
	if !strings.Contains(out, "Determinism Check") {
		t.Errorf("expected header in output:\n%s", out)
	}
}

func TestPrintDeterminismTextNonDeterministic(t *testing.T) {
	dr := report.DeterminismResult{
		Deterministic: false,
		BaselineCount: 5,
		CompareCount:  5,
		Differences: []report.Difference{
			{FixtureID: "aaa111222333", Field: "matched", Baseline: true, Compare: false},
			{FixtureID: "bbb444555666", Field: "failure_id", Baseline: "docker-auth", Compare: ""},
		},
	}
	var buf bytes.Buffer
	report.PrintDeterminismText(&buf, dr)
	out := buf.String()
	if !strings.Contains(out, "NON-DETERMINISTIC") {
		t.Errorf("expected NON-DETERMINISTIC in output:\n%s", out)
	}
	if !strings.Contains(out, "matched") {
		t.Errorf("expected 'matched' field in output:\n%s", out)
	}
}

func TestPrintDeterminismTextTruncatesLongDiffList(t *testing.T) {
	diffs := make([]report.Difference, 25)
	for i := range diffs {
		diffs[i] = report.Difference{
			FixtureID: strings.Repeat("a", 12+i%3), // valid fixture IDs >=12 chars
			Field:     "matched",
			Baseline:  true,
			Compare:   false,
		}
	}
	dr := report.DeterminismResult{
		Deterministic: false,
		BaselineCount: 25,
		CompareCount:  25,
		Differences:   diffs,
	}
	var buf bytes.Buffer
	report.PrintDeterminismText(&buf, dr)
	out := buf.String()
	if !strings.Contains(out, "and") || !strings.Contains(out, "more") {
		t.Errorf("expected truncation message for >20 diffs:\n%s", out)
	}
}
