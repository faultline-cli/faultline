package report_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"faultline/tools/eval-corpus/model"
	"faultline/tools/eval-corpus/report"
)

func TestComputeCoverage(t *testing.T) {
	results := []model.EvalResult{
		{FixtureID: "a", Matched: true, FailureID: "docker-auth"},
		{FixtureID: "b", Matched: true, FailureID: "docker-auth"},
		{FixtureID: "c", Matched: false, FirstLineTag: "t1", FirstLineSnippet: "error x"},
		{FixtureID: "d", Matched: false, FirstLineTag: "t1", FirstLineSnippet: "error x"},
		{FixtureID: "e", Error: "empty fixture"},
	}

	rpt := report.Compute(results, 10)

	if rpt.Coverage.Total != 5 {
		t.Errorf("Total = %d, want 5", rpt.Coverage.Total)
	}
	if rpt.Coverage.Matched != 2 {
		t.Errorf("Matched = %d, want 2", rpt.Coverage.Matched)
	}
	if rpt.Coverage.Unmatched != 3 {
		t.Errorf("Unmatched = %d, want 3", rpt.Coverage.Unmatched)
	}
	if rpt.Coverage.Errors != 1 {
		t.Errorf("Errors = %d, want 1", rpt.Coverage.Errors)
	}

	wantRate := 2.0 / 5.0
	if rpt.Coverage.MatchRate != wantRate {
		t.Errorf("MatchRate = %v, want %v", rpt.Coverage.MatchRate, wantRate)
	}
}

func TestComputeDistribution(t *testing.T) {
	results := []model.EvalResult{
		{FixtureID: "a", Matched: true, FailureID: "docker-auth"},
		{FixtureID: "b", Matched: true, FailureID: "docker-auth"},
		{FixtureID: "c", Matched: true, FailureID: "missing-exec"},
	}

	rpt := report.Compute(results, 10)

	if len(rpt.Distribution) == 0 {
		t.Fatal("Distribution should not be empty")
	}
	top := rpt.Distribution[0]
	if top.FailureID != "docker-auth" {
		t.Errorf("top FailureID = %q, want %q", top.FailureID, "docker-auth")
	}
	if top.Count != 2 {
		t.Errorf("top Count = %d, want 2", top.Count)
	}
}

func TestComputeClusters(t *testing.T) {
	results := []model.EvalResult{
		{FixtureID: "a", Matched: false, FirstLineTag: "tag1", FirstLineSnippet: "error in module load"},
		{FixtureID: "b", Matched: false, FirstLineTag: "tag1", FirstLineSnippet: "error in module load"},
		{FixtureID: "c", Matched: false, FirstLineTag: "tag2", FirstLineSnippet: "npm ERR!"},
	}

	rpt := report.Compute(results, 10)

	if len(rpt.Clusters) == 0 {
		t.Fatal("Clusters should not be empty")
	}
	top := rpt.Clusters[0]
	if top.Tag != "tag1" {
		t.Errorf("top cluster Tag = %q, want %q", top.Tag, "tag1")
	}
	if top.Count != 2 {
		t.Errorf("top cluster Count = %d, want 2", top.Count)
	}
}

func TestPrintText(t *testing.T) {
	results := []model.EvalResult{
		{FixtureID: "a", Matched: true, FailureID: "docker-auth"},
		{FixtureID: "b", Matched: false, FirstLineTag: "t1", FirstLineSnippet: "error x"},
	}
	rpt := report.Compute(results, 10)

	var buf bytes.Buffer
	report.PrintText(&buf, rpt)
	out := buf.String()

	if !strings.Contains(out, "Coverage Report") {
		t.Error("output missing Coverage Report header")
	}
	if !strings.Contains(out, "docker-auth") {
		t.Error("output missing failure ID")
	}
}

func TestDecodeResultsRoundTrip(t *testing.T) {
	original := []model.EvalResult{
		{FixtureID: "abc", Matched: true, FailureID: "docker-auth", Confidence: 0.9},
		{FixtureID: "def", Matched: false},
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, r := range original {
		if err := enc.Encode(r); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}

	got, err := report.DecodeResults(&buf)
	if err != nil {
		t.Fatalf("DecodeResults: %v", err)
	}
	if len(got) != len(original) {
		t.Fatalf("got %d results, want %d", len(got), len(original))
	}
	for i, r := range got {
		if r.FixtureID != original[i].FixtureID {
			t.Errorf("[%d] FixtureID = %q, want %q", i, r.FixtureID, original[i].FixtureID)
		}
	}
}
