package fixtures

import (
	"strings"
	"testing"
)

func TestFormatIngestResultTextAndJSON(t *testing.T) {
	result := IngestResult{
		Written: []Fixture{{
			ID: "fixture-1",
			Source: SourceMetadata{
				URL: "https://example.com/1",
			},
		}},
		Skipped: []string{"https://example.com/2: duplicate"},
	}

	text, err := FormatIngestResult(result, false)
	if err != nil {
		t.Fatalf("FormatIngestResult(text): %v", err)
	}
	for _, want := range []string{
		"Written: 1",
		"- fixture-1 (https://example.com/1)",
		"Skipped: 1",
		"- https://example.com/2: duplicate",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in text output, got %q", want, text)
		}
	}

	jsonText, err := FormatIngestResult(result, true)
	if err != nil {
		t.Fatalf("FormatIngestResult(json): %v", err)
	}
	if !strings.Contains(jsonText, "\"fixture-1\"") || !strings.Contains(jsonText, "\"skipped\"") {
		t.Fatalf("unexpected JSON output: %q", jsonText)
	}
}

func TestFormatReviewReportTextAndEmpty(t *testing.T) {
	report := ReviewReport{
		Items: []ReviewItem{{
			Fixture: Fixture{
				ID: "staging-1",
				Source: SourceMetadata{
					URL: "https://example.com/issue/1",
				},
			},
			Status:         "candidate",
			PredictedTopID: "docker-auth",
			PredictedTop3:  []string{"docker-auth", "github-auth"},
			DuplicateOf:    "real-1",
			NearDuplicates: []string{"real-2", "real-3"},
		}},
	}

	text, err := FormatReviewReport(report, false)
	if err != nil {
		t.Fatalf("FormatReviewReport(text): %v", err)
	}
	for _, want := range []string{
		"staging-1 [candidate] top=docker-auth",
		"duplicate_of: real-1",
		"near_duplicates: real-2, real-3",
		"top3: docker-auth, github-auth",
		"source: https://example.com/issue/1",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in text output, got %q", want, text)
		}
	}

	empty, err := FormatReviewReport(ReviewReport{}, false)
	if err != nil {
		t.Fatalf("FormatReviewReport(empty): %v", err)
	}
	if empty != "No staging fixtures found.\n" {
		t.Fatalf("empty output = %q", empty)
	}
}

func TestFormatStatsReportTextAndJSON(t *testing.T) {
	report := Report{
		Class:              ClassReal,
		FixtureCount:       4,
		Top1Count:          3,
		Top3Count:          4,
		UnmatchedCount:     1,
		FalsePositiveCount: 1,
		WeakMatchCount:     2,
		UnmatchedFixtureIDs: []string{
			"fixture-unmatched",
		},
		WeakMatchFixtureIDs: []string{
			"fixture-weak",
		},
		RecurringPatterns: map[string]int{
			"b-pattern": 1,
			"a-pattern": 2,
		},
		ThresholdViolations: []string{"z-last", "a-first"},
	}

	text, err := FormatStatsReport(report, false)
	if err != nil {
		t.Fatalf("FormatStatsReport(text): %v", err)
	}
	for _, want := range []string{
		"class: real",
		"fixtures: 4",
		"top_1: 0.750",
		"top_3: 1.000",
		"unmatched: 0.250",
		"false_positive: 0.250",
		"weak_match: 0.500",
		"unmatched_ids: fixture-unmatched",
		"weak_ids: fixture-weak",
		"patterns: a-pattern=2, b-pattern=1",
		"violations: a-first | z-last",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in text output, got %q", want, text)
		}
	}

	jsonText, err := FormatStatsReport(report, true)
	if err != nil {
		t.Fatalf("FormatStatsReport(json): %v", err)
	}
	if !strings.Contains(jsonText, "\"fixture_count\": 4") || !strings.Contains(jsonText, "\"class\": \"real\"") {
		t.Fatalf("unexpected JSON output: %q", jsonText)
	}
}
