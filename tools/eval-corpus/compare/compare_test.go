package compare_test

import (
	"testing"

	"faultline/tools/eval-corpus/compare"
	"faultline/tools/eval-corpus/model"
)

func makeResult(id string, matched bool, failureID string) model.EvalResult {
	return model.EvalResult{
		FixtureID: id,
		Matched:   matched,
		FailureID: failureID,
	}
}

func TestCompareMatchRates(t *testing.T) {
	baseline := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", true, "auth"),
		makeResult("c", false, ""),
		makeResult("d", false, ""),
	}
	current := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", true, "auth"),
		makeResult("c", true, "missing-exec"),
		makeResult("d", false, ""),
	}
	res := compare.Compare(baseline, current, compare.Options{})
	if res.BaselineMatchRate != 0.5 {
		t.Errorf("BaselineMatchRate = %.2f, want 0.50", res.BaselineMatchRate)
	}
	if res.CurrentMatchRate != 0.75 {
		t.Errorf("CurrentMatchRate = %.2f, want 0.75", res.CurrentMatchRate)
	}
	if res.Delta != 0.25 {
		t.Errorf("Delta = %.2f, want 0.25", res.Delta)
	}
}

func TestComparePassWhenAboveMinRate(t *testing.T) {
	results := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", true, "auth"),
		makeResult("c", false, ""),
	}
	res := compare.Compare(results, results, compare.Options{MinMatchRate: 0.50})
	if !res.Pass {
		t.Errorf("expected pass when match rate (0.67) >= min (0.50)")
	}
}

func TestCompareFailsWhenBelowMinRate(t *testing.T) {
	baseline := []model.EvalResult{makeResult("a", true, "x"), makeResult("b", false, "")}
	current := []model.EvalResult{makeResult("a", false, ""), makeResult("b", false, "")}
	res := compare.Compare(baseline, current, compare.Options{MinMatchRate: 0.72})
	if res.Pass {
		t.Error("expected fail when match rate (0.00) < min (0.72)")
	}
	if len(res.FailReasons) == 0 {
		t.Error("expected at least one fail reason")
	}
}

func TestCompareFailsOnCoverageDrop(t *testing.T) {
	baseline := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", true, "auth"),
	}
	current := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", false, ""),
	}
	res := compare.Compare(baseline, current, compare.Options{FailOnCoverageDrop: true})
	if res.Pass {
		t.Error("expected fail when coverage dropped")
	}
}

func TestCompareNoDropDoesNotFail(t *testing.T) {
	baseline := []model.EvalResult{makeResult("a", false, "")}
	current := []model.EvalResult{makeResult("a", true, "auth")}
	res := compare.Compare(baseline, current, compare.Options{FailOnCoverageDrop: true})
	if !res.Pass {
		t.Errorf("expected pass when coverage improved: %v", res.FailReasons)
	}
}

func TestComparePlaybookDeltas(t *testing.T) {
	baseline := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", true, "auth"),
		makeResult("c", true, "missing-exec"),
	}
	current := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", false, ""),
		makeResult("c", true, "missing-exec"),
	}
	res := compare.Compare(baseline, current, compare.Options{})
	if len(res.PlaybooksLostMatches) != 1 || res.PlaybooksLostMatches[0].FailureID != "auth" {
		t.Errorf("expected auth in lost matches, got %v", res.PlaybooksLostMatches)
	}
	if len(res.PlaybooksGainedMatches) != 0 {
		t.Errorf("expected no gained matches, got %v", res.PlaybooksGainedMatches)
	}
}

func TestCompareNewUnmatchedTags(t *testing.T) {
	baseline := []model.EvalResult{
		{FixtureID: "a", Matched: false, FirstLineTag: "tag1"},
	}
	current := []model.EvalResult{
		{FixtureID: "a", Matched: false, FirstLineTag: "tag1"},
		{FixtureID: "b", Matched: false, FirstLineTag: "tag2"}, // new tag
	}
	res := compare.Compare(baseline, current, compare.Options{})
	if len(res.NewUnmatchedTags) != 1 || res.NewUnmatchedTags[0] != "tag2" {
		t.Errorf("NewUnmatchedTags = %v, want [tag2]", res.NewUnmatchedTags)
	}
}

func TestCompareEmptyInputsPass(t *testing.T) {
	res := compare.Compare(nil, nil, compare.Options{MinMatchRate: 0.50})
	// Zero results means 0% match rate which is < 0.50 — gate should fail.
	// But there are no results at all, so the gate still applies.
	if res.Pass {
		t.Error("expected fail: 0 results → 0% match rate < min 0.50")
	}
}

func TestCompareDeterministicOutput(t *testing.T) {
	// Running the same compare twice should produce identical results.
	baseline := []model.EvalResult{makeResult("a", true, "auth"), makeResult("b", false, "")}
	current := []model.EvalResult{makeResult("a", false, ""), makeResult("b", false, "")}
	r1 := compare.Compare(baseline, current, compare.Options{})
	r2 := compare.Compare(baseline, current, compare.Options{})
	if r1.Pass != r2.Pass {
		t.Error("Pass differs between runs")
	}
	if r1.Delta != r2.Delta {
		t.Error("Delta differs between runs")
	}
}
