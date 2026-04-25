package model_test

import (
	"testing"

	"faultline/tools/eval-corpus/model"
)

func TestScoreResultUnlabelledWhenNoExpectedID(t *testing.T) {
	r := model.EvalResult{FixtureID: "f1", Matched: true, FailureID: "auth"}
	got := model.ScoreResult(r)
	if got != model.OutcomeUnlabelled {
		t.Errorf("got %q, want %q", got, model.OutcomeUnlabelled)
	}
}

func TestScoreResultUnlabelledZeroValue(t *testing.T) {
	var r model.EvalResult
	got := model.ScoreResult(r)
	if got != model.OutcomeUnlabelled {
		t.Errorf("got %q, want %q", got, model.OutcomeUnlabelled)
	}
}

func TestScoreResultTP(t *testing.T) {
	r := model.EvalResult{
		FixtureID:         "f1",
		Matched:           true,
		FailureID:         "auth",
		ExpectedFailureID: "auth",
	}
	got := model.ScoreResult(r)
	if got != model.OutcomeTP {
		t.Errorf("got %q, want %q", got, model.OutcomeTP)
	}
}

func TestScoreResultFPWhenMatchedWithWrongID(t *testing.T) {
	r := model.EvalResult{
		FixtureID:         "f1",
		Matched:           true,
		FailureID:         "missing-exec",
		ExpectedFailureID: "auth",
	}
	got := model.ScoreResult(r)
	if got != model.OutcomeFP {
		t.Errorf("got %q, want %q", got, model.OutcomeFP)
	}
}

func TestScoreResultFNWhenNotMatchedButExpected(t *testing.T) {
	r := model.EvalResult{
		FixtureID:         "f1",
		Matched:           false,
		ExpectedFailureID: "auth",
	}
	got := model.ScoreResult(r)
	if got != model.OutcomeFN {
		t.Errorf("got %q, want %q", got, model.OutcomeFN)
	}
}

func TestScoreResultUnlabelledWhenNotMatchedAndNoExpected(t *testing.T) {
	r := model.EvalResult{FixtureID: "f1", Matched: false}
	got := model.ScoreResult(r)
	if got != model.OutcomeUnlabelled {
		t.Errorf("got %q, want %q", got, model.OutcomeUnlabelled)
	}
}
