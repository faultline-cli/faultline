package runner_test

import (
	"context"
	"testing"

	"faultline/tools/eval-corpus/model"
	"faultline/tools/eval-corpus/runner"
)

func TestRunMatchesDockerAuthLog(t *testing.T) {
	// This log content matches the docker-auth playbook bundled with Faultline.
	fixtures := []model.Fixture{
		{
			ID:     "docker-auth-test",
			Raw:    "Error response from daemon: unauthorized: authentication required",
			Source: "test",
		},
	}

	results := evalFixtures(t, fixtures)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	r := results[0]
	if !r.Matched {
		t.Errorf("expected match for docker-auth log, got unmatched; error=%q", r.Error)
	}
	if r.FailureID != "docker-auth" {
		t.Errorf("FailureID = %q, want %q", r.FailureID, "docker-auth")
	}
}

func TestRunUnmatchedLog(t *testing.T) {
	fixtures := []model.Fixture{
		{
			ID:     "unmatched-test",
			Raw:    "everything is fine, nothing to see here",
			Source: "test",
		},
	}

	results := evalFixtures(t, fixtures)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	r := results[0]
	if r.Matched {
		t.Errorf("expected no match for generic log, got %q", r.FailureID)
	}
	if r.Error != "" {
		t.Errorf("unexpected error: %q", r.Error)
	}
}

func TestRunEmptyLog(t *testing.T) {
	fixtures := []model.Fixture{
		{ID: "empty", Raw: "   ", Source: "test"},
	}

	results := evalFixtures(t, fixtures)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Error == "" {
		t.Error("expected error for empty log fixture")
	}
}

func TestRunExtractsFirstLine(t *testing.T) {
	fixtures := []model.Fixture{
		{
			ID:     "multiline",
			Raw:    "first line\nsecond line\nthird line",
			Source: "test",
		},
	}

	results := evalFixtures(t, fixtures)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	r := results[0]
	if r.FirstLineSnippet != "first line" {
		t.Errorf("FirstLineSnippet = %q, want %q", r.FirstLineSnippet, "first line")
	}
	if r.FirstLineTag == "" {
		t.Error("FirstLineTag should be non-empty")
	}
}

func TestRunDeterministic(t *testing.T) {
	fixtures := []model.Fixture{
		{
			ID:     "det-a",
			Raw:    "Error response from daemon: unauthorized: authentication required",
			Source: "test",
		},
		{
			ID:     "det-b",
			Raw:    "exec /__e/node20/bin/node: no such file or directory",
			Source: "test",
		},
		{
			ID:     "det-c",
			Raw:    "nothing matches here",
			Source: "test",
		},
	}

	run1 := evalFixtures(t, fixtures)
	run2 := evalFixtures(t, fixtures)

	if len(run1) != len(run2) {
		t.Fatalf("result counts differ: %d vs %d", len(run1), len(run2))
	}
	for i := range run1 {
		if run1[i].Matched != run2[i].Matched {
			t.Errorf("[%d] Matched differs: %v vs %v", i, run1[i].Matched, run2[i].Matched)
		}
		if run1[i].FailureID != run2[i].FailureID {
			t.Errorf("[%d] FailureID differs: %q vs %q", i, run1[i].FailureID, run2[i].FailureID)
		}
		if run1[i].Error != run2[i].Error {
			t.Errorf("[%d] Error differs: %q vs %q", i, run1[i].Error, run2[i].Error)
		}
	}
}

// evalFixtures is a helper that feeds fixtures through the runner and returns
// collected results in fixture-id order.
func evalFixtures(t *testing.T, fixtures []model.Fixture) []model.EvalResult {
	t.Helper()

	in := make(chan model.Fixture, len(fixtures))
	for _, f := range fixtures {
		in <- f
	}
	close(in)

	out := make(chan model.EvalResult, len(fixtures))

	ctx := context.Background()
	if err := runner.Run(ctx, runner.Options{Workers: 1}, in, out); err != nil {
		t.Fatalf("runner.Run: %v", err)
	}
	close(out)

	var results []model.EvalResult
	for r := range out {
		results = append(results, r)
	}
	return results
}
