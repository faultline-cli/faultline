package report_test

import (
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
