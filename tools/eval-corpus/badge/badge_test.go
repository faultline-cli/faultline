package badge_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"faultline/tools/eval-corpus/badge"
	"faultline/tools/eval-corpus/model"
)

func makeResults(matchedIDs, unmatchedSnippets []string) []model.EvalResult {
	results := make([]model.EvalResult, 0, len(matchedIDs)+len(unmatchedSnippets))
	for i, id := range matchedIDs {
		results = append(results, model.EvalResult{
			FixtureID: strings.Repeat("m", i+1),
			Matched:   true,
			FailureID: id,
		})
	}
	for i, snip := range unmatchedSnippets {
		tag := strings.ReplaceAll(snip, " ", "")[:8]
		results = append(results, model.EvalResult{
			FixtureID:        strings.Repeat("u", i+1),
			Matched:          false,
			FirstLineTag:     tag,
			FirstLineSnippet: snip,
		})
	}
	return results
}

func TestComputeMatchRate(t *testing.T) {
	results := makeResults(
		[]string{"auth", "auth", "missing-exec"},
		[]string{"network timeout", "connection refused"},
	)
	s := badge.Compute(results, badge.Options{})
	if s.CorpusSize != 5 {
		t.Errorf("CorpusSize = %d, want 5", s.CorpusSize)
	}
	if s.Matched != 3 {
		t.Errorf("Matched = %d, want 3", s.Matched)
	}
	want := 0.6
	if s.MatchRate != want {
		t.Errorf("MatchRate = %.2f, want %.2f", s.MatchRate, want)
	}
}

func TestComputeTopCovered(t *testing.T) {
	results := makeResults(
		[]string{"auth", "auth", "auth", "missing-exec", "missing-exec"},
		nil,
	)
	s := badge.Compute(results, badge.Options{TopN: 3})
	if len(s.TopCovered) == 0 {
		t.Fatal("TopCovered is empty")
	}
	if s.TopCovered[0] != "auth" {
		t.Errorf("TopCovered[0] = %q, want %q", s.TopCovered[0], "auth")
	}
}

func TestComputeTopGaps(t *testing.T) {
	results := makeResults(nil, []string{
		"network timeout",
		"network timeout",
		"connection refused",
	})
	s := badge.Compute(results, badge.Options{TopN: 5})
	if len(s.TopGaps) == 0 {
		t.Fatal("TopGaps is empty")
	}
	// "network timeout" has count 2, should be first
	if s.TopGaps[0] != "network timeout" {
		t.Errorf("TopGaps[0] = %q, want network timeout", s.TopGaps[0])
	}
}

func TestComputeDeterministicField(t *testing.T) {
	results := makeResults([]string{"auth"}, nil)
	s := badge.Compute(results, badge.Options{Deterministic: "pass"})
	if s.Deterministic != "pass" {
		t.Errorf("Deterministic = %q, want %q", s.Deterministic, "pass")
	}
}

func TestComputeUnknownWhenNotSet(t *testing.T) {
	s := badge.Compute(nil, badge.Options{})
	if s.Deterministic != "unknown" {
		t.Errorf("Deterministic = %q, want %q", s.Deterministic, "unknown")
	}
}

func TestWriteJSONRoundtrip(t *testing.T) {
	results := makeResults([]string{"auth"}, []string{"network timeout"})
	s := badge.Compute(results, badge.Options{CorpusVersion: "v1", TopN: 5})

	var buf bytes.Buffer
	if err := badge.WriteJSON(&buf, s); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var decoded badge.Summary
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.CorpusVersion != "v1" {
		t.Errorf("CorpusVersion = %q, want v1", decoded.CorpusVersion)
	}
	if decoded.CorpusSize != s.CorpusSize {
		t.Errorf("CorpusSize = %d, want %d", decoded.CorpusSize, s.CorpusSize)
	}
}

func TestWriteMarkdownContainsKeyFields(t *testing.T) {
	results := makeResults([]string{"auth", "auth"}, []string{"connection refused"})
	s := badge.Compute(results, badge.Options{CorpusVersion: "ci-v1", Deterministic: "pass"})

	var buf bytes.Buffer
	badge.WriteMarkdown(&buf, s)
	md := buf.String()

	for _, want := range []string{"Faultline", "ci-v1", "pass", "Coverage"} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q", want)
		}
	}
}
