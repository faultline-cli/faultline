package fixtures

import (
	"strings"
	"testing"
)

func TestEvaluateFixturesRejectsDuplicateFingerprints(t *testing.T) {
	layout := Layout{Root: t.TempDir()}
	loaded := []Fixture{
		{
			ID:            "real-1",
			Title:         "One",
			FixtureClass:  ClassReal,
			NormalizedLog: "pull access denied",
			Fingerprint:   "dup-fingerprint",
			Expectation:   Expectation{ExpectedPlaybook: "docker-auth"},
			Source:        SourceMetadata{Adapter: "github-issue", Provider: "github", URL: "https://example.com/1"},
			Review:        ReviewMetadata{Status: "promoted", PromotedAt: "2026-04-18T00:00:00Z"},
		},
		{
			ID:            "real-2",
			Title:         "Two",
			FixtureClass:  ClassReal,
			NormalizedLog: "authentication required",
			Fingerprint:   "dup-fingerprint",
			Expectation:   Expectation{ExpectedPlaybook: "docker-auth"},
			Source:        SourceMetadata{Adapter: "github-issue", Provider: "github", URL: "https://example.com/2"},
			Review:        ReviewMetadata{Status: "promoted", PromotedAt: "2026-04-18T00:00:00Z"},
		},
	}

	_, err := EvaluateFixtures(layout, ClassReal, loaded, EvaluateOptions{NoHistory: true})
	if err == nil || !strings.Contains(err.Error(), "duplicate fingerprint") {
		t.Fatalf("expected duplicate fingerprint error, got %v", err)
	}
}

func TestEvaluateFixturesRejectsMissingRealFixtureMetadata(t *testing.T) {
	layout := Layout{Root: t.TempDir()}
	loaded := []Fixture{{
		ID:            "real-1",
		FixtureClass:  ClassReal,
		NormalizedLog: "pull access denied",
		Fingerprint:   "unique-fingerprint",
		Expectation:   Expectation{ExpectedPlaybook: "docker-auth"},
		Source:        SourceMetadata{Provider: "github"},
		Review:        ReviewMetadata{Status: "candidate"},
	}}

	_, err := EvaluateFixtures(layout, ClassReal, loaded, EvaluateOptions{NoHistory: true})
	if err == nil {
		t.Fatal("expected metadata validation error")
	}
	for _, want := range []string{
		"missing title",
		"missing source.adapter",
		"missing source.url",
		"review.status must be promoted",
		"missing review.promoted_at",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected %q in error, got %v", want, err)
		}
	}
}

func TestCheckBaselineRejectsWeakMatchAndFingerprintRegression(t *testing.T) {
	report := Report{
		Class:              ClassReal,
		FixtureCount:       2,
		Top1Count:          2,
		Top3Count:          2,
		RecurringPatterns:  map[string]int{"docker-auth": 2},
		Providers:          map[string]int{"github": 2},
		Adapters:           map[string]int{"github-issue": 2},
		WeakMatchCount:     1,
		UnmatchedCount:     0,
		FalsePositiveCount: 0,
	}
	baseline := Baseline{
		Class:             ClassReal,
		FixtureCount:      2,
		Top1Rate:          1,
		Top3Rate:          1,
		UnmatchedRate:     0,
		FalsePositiveRate: 0,
		WeakMatchRate:     0,
		Fingerprint:       "baseline-fingerprint",
		Thresholds: Thresholds{
			MinTop1:          1,
			MinTop3:          1,
			MaxUnmatched:     0,
			MaxFalsePositive: 0,
			MaxWeakMatch:     0.25,
		},
	}

	err := CheckBaseline(&report, baseline)
	if err == nil {
		t.Fatal("expected baseline regression error")
	}
	if !strings.Contains(err.Error(), "weak-match rate regressed") {
		t.Fatalf("expected weak-match regression, got %v", err)
	}
	if !strings.Contains(err.Error(), "corpus fingerprint changed") {
		t.Fatalf("expected fingerprint regression, got %v", err)
	}
}
