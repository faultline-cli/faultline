package hypothesis

import (
	"strings"
	"testing"

	"faultline/internal/model"
)

func TestValidSignalKnownSignals(t *testing.T) {
	// Verify that signals from the registry are recognized as valid
	for signal := range signalRegistry {
		if !ValidSignal(signal) {
			t.Errorf("ValidSignal(%q) = false, expected true for known registry signal", signal)
		}
	}
}

func TestValidSignalPrefixedForms(t *testing.T) {
	cases := []struct {
		signal string
		wantOK bool
	}{
		{"log.contains:error", true},
		{"log.contains:  ", false}, // blank value after prefix
		{"log.absent:panic", true},
		{"log.absent:", false},
		{"delta.signal:dep.changed", true},
		{"delta.signal:  ", false},
		{"delta.absent:some.signal", true},
		{"delta.absent:", false},
		{"context.stage:build", true},
		{"context.stage:", false},
		{"context.stage.absent:test", true},
		{"context.stage.absent:  ", false},
		{"unknown.prefix:value", false},
		{"", false},
		{"   ", false},
	}
	for _, tc := range cases {
		got := ValidSignal(tc.signal)
		if got != tc.wantOK {
			t.Errorf("ValidSignal(%q) = %v, want %v", tc.signal, got, tc.wantOK)
		}
	}
}

func TestDescribeSignalRegistrySignalHasDescription(t *testing.T) {
	for signal, def := range signalRegistry {
		got := DescribeSignal(signal)
		if got == "" {
			t.Errorf("DescribeSignal(%q): expected non-empty description, got empty (registered description: %q)", signal, def.description)
		}
	}
}

func TestDescribeSignalPrefixedForms(t *testing.T) {
	cases := []struct {
		signal      string
		wantContain string
	}{
		{"log.contains:npm error", "npm error"},
		{"log.absent:unreachable", "unreachable"},
		{"delta.signal:dep.changed", "dep.changed"},
		{"delta.absent:cache.hit", "cache.hit"},
		{"context.stage:build", "build"},
		{"context.stage.absent:test", "test"},
	}
	for _, tc := range cases {
		got := DescribeSignal(tc.signal)
		if !strings.Contains(got, tc.wantContain) {
			t.Errorf("DescribeSignal(%q) = %q, expected to contain %q", tc.signal, got, tc.wantContain)
		}
	}
}

func TestDescribeSignalUnknownReturnsSignalItself(t *testing.T) {
	signal := "some.unknown.signal"
	got := DescribeSignal(signal)
	if got != signal {
		t.Errorf("DescribeSignal(%q) = %q, expected the signal itself", signal, got)
	}
}

func TestTrimEvidenceFiltersEmptyAndCapsAtTwo(t *testing.T) {
	cases := []struct {
		input []string
		want  []string
	}{
		{[]string{"  ", "a", "b", "c"}, []string{"a", "b"}},
		{[]string{"", "  "}, []string{}},
		{[]string{"only"}, []string{"only"}},
	}
	for _, tc := range cases {
		got := trimEvidence(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("trimEvidence(%v) = %v, want %v", tc.input, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("trimEvidence(%v)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

func TestCompareAgainstLikelyReturnsNilWhenAlternativeScoreIsZero(t *testing.T) {
	likely := model.DifferentialCandidate{HypothesisScore: 3.0}
	alt := model.DifferentialCandidate{HypothesisScore: 0}
	if got := compareAgainstLikely(likely, alt); got != nil {
		t.Fatalf("expected nil when alternative score is 0, got %v", got)
	}
}

func TestCompareAgainstLikelyReturnsNilWhenNotLeading(t *testing.T) {
	likely := model.DifferentialCandidate{HypothesisScore: 2.0}
	alt := model.DifferentialCandidate{HypothesisScore: 3.0}
	if got := compareAgainstLikely(likely, alt); got != nil {
		t.Fatalf("expected nil when likely score <= alternative score, got %v", got)
	}
}

func TestCompareAgainstLikelyExplainsPointGapWithoutWhy(t *testing.T) {
	likely := model.DifferentialCandidate{HypothesisScore: 4.0, Why: nil}
	alt := model.DifferentialCandidate{HypothesisScore: 2.0}
	got := compareAgainstLikely(likely, alt)
	if len(got) != 1 {
		t.Fatalf("expected 1 explanation, got %v", got)
	}
	if !strings.Contains(got[0], "trailed") {
		t.Errorf("expected 'trailed' in explanation, got %q", got[0])
	}
}

func TestCompareAgainstLikelyExplainsUsingFirstWhyReason(t *testing.T) {
	likely := model.DifferentialCandidate{
		HypothesisScore: 4.0,
		Why:             []string{"npm lockfile mismatch detected"},
	}
	alt := model.DifferentialCandidate{HypothesisScore: 2.0}
	got := compareAgainstLikely(likely, alt)
	if len(got) != 1 {
		t.Fatalf("expected 1 explanation, got %v", got)
	}
	if !strings.Contains(got[0], "npm lockfile mismatch") {
		t.Errorf("expected discriminator reason in explanation, got %q", got[0])
	}
}

func TestContextEvidenceReturnsStageLabel(t *testing.T) {
	got := contextEvidence("build")
	if len(got) != 1 || got[0] != "stage: build" {
		t.Errorf("contextEvidence(build) = %v, want [stage: build]", got)
	}
	if contextEvidence("") != nil {
		t.Error("contextEvidence('') expected nil")
	}
	if contextEvidence("  ") != nil {
		t.Error("contextEvidence('  ') expected nil for blank stage")
	}
}

func TestGenericSignalDescriptionCoversAllPrefixes(t *testing.T) {
	cases := []struct {
		signal  string
		contain string
	}{
		{"log.contains:timeout", "timeout"},
		{"log.absent:cache hit", "cache hit"},
		{"delta.signal:env.changed", "env.changed"},
		{"delta.absent:hotfix", "hotfix"},
		{"context.stage:deploy", "deploy"},
		{"context.stage.absent:test", "test"},
		{"bare.signal", "bare.signal"}, // unknown: returned as-is
	}
	for _, tc := range cases {
		got := genericSignalDescription(tc.signal)
		if !strings.Contains(got, tc.contain) {
			t.Errorf("genericSignalDescription(%q) = %q, expected to contain %q", tc.signal, got, tc.contain)
		}
	}
}
