package fixtures

import (
	"strings"
	"testing"
)

func TestParseClassValid(t *testing.T) {
	cases := []struct {
		input string
		want  Class
	}{
		{"minimal", ClassMinimal},
		{"real", ClassReal},
		{"staging", ClassStaging},
		{"all", ClassAll},
		{"", ClassAll},
		{"  ALL  ", ClassAll},
		{"MINIMAL", ClassMinimal},
	}
	for _, tc := range cases {
		got, err := ParseClass(tc.input)
		if err != nil {
			t.Errorf("ParseClass(%q): unexpected error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseClass(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestParseClassInvalid(t *testing.T) {
	for _, input := range []string{"beta", "unknown", "fixture", "test"} {
		if _, err := ParseClass(input); err == nil {
			t.Errorf("ParseClass(%q): expected error, got nil", input)
		}
	}
}

func TestFingerprintForLogIsDeterministic(t *testing.T) {
	text := "npm ERR! ERESOLVE unable to resolve dependency tree"
	a := FingerprintForLog(text)
	b := FingerprintForLog(text)
	if a != b {
		t.Fatalf("FingerprintForLog not deterministic: %q != %q", a, b)
	}
	if len(a) != 16 {
		t.Fatalf("expected 16-char fingerprint, got %d chars: %q", len(a), a)
	}
}

func TestFingerprintForLogDiffersForDifferentInputs(t *testing.T) {
	a := FingerprintForLog("error A")
	b := FingerprintForLog("error B")
	if a == b {
		t.Fatalf("expected different fingerprints for different inputs, got %q", a)
	}
}

func TestAdapterByNameKnownAdapters(t *testing.T) {
	names := []string{"github-issue", "gitlab-issue", "stackexchange-question", "discourse-topic", "reddit-post"}
	for _, name := range names {
		adapter, err := adapterByName(name)
		if err != nil {
			t.Errorf("adapterByName(%q): unexpected error: %v", name, err)
			continue
		}
		if adapter.Name() != name {
			t.Errorf("adapterByName(%q).Name() = %q, want %q", name, adapter.Name(), name)
		}
	}
}

func TestAdapterByNameUnknownReturnsError(t *testing.T) {
	if _, err := adapterByName("unknown"); err == nil {
		t.Error("expected error for unknown adapter, got nil")
	}
}

func TestMergeFixturesSortsById(t *testing.T) {
	group1 := []Fixture{{ID: "z-fixture"}, {ID: "m-fixture"}}
	group2 := []Fixture{{ID: "a-fixture"}}
	merged := mergeFixtures([][]Fixture{group1, group2})
	if len(merged) != 3 {
		t.Fatalf("expected 3 fixtures, got %d", len(merged))
	}
	if merged[0].ID != "a-fixture" {
		t.Errorf("expected a-fixture first, got %q", merged[0].ID)
	}
	if merged[2].ID != "z-fixture" {
		t.Errorf("expected z-fixture last, got %q", merged[2].ID)
	}
}

func TestMergeFixturesEmptyInput(t *testing.T) {
	merged := mergeFixtures(nil)
	if merged != nil && len(merged) != 0 {
		t.Fatalf("expected empty result, got %#v", merged)
	}
}

func TestLineSignatureBuildsSetFromLines(t *testing.T) {
	text := "Error: module not found\nFailed to install\n\nError: module not found"
	sig := lineSignature(text)
	if _, ok := sig["error: module not found"]; !ok {
		t.Error("expected lowercase 'error: module not found' in signature")
	}
	if _, ok := sig["failed to install"]; !ok {
		t.Error("expected 'failed to install' in signature")
	}
	// Duplicate lines collapse to one entry
	if len(sig) != 2 {
		t.Errorf("expected 2 unique lines, got %d", len(sig))
	}
}

func TestLineSignatureSkipsEmptyLines(t *testing.T) {
	sig := lineSignature("\n\n  \n")
	if len(sig) != 0 {
		t.Errorf("expected empty signature for blank text, got %#v", sig)
	}
}

func TestJaccardSimilarityIdenticalSets(t *testing.T) {
	a := map[string]struct{}{"x": {}, "y": {}}
	if got := jaccardSimilarity(a, a); got != 1.0 {
		t.Errorf("expected 1.0 for identical sets, got %.2f", got)
	}
}

func TestJaccardSimilarityDisjointSets(t *testing.T) {
	a := map[string]struct{}{"x": {}}
	b := map[string]struct{}{"y": {}}
	if got := jaccardSimilarity(a, b); got != 0.0 {
		t.Errorf("expected 0.0 for disjoint sets, got %.2f", got)
	}
}

func TestJaccardSimilarityPartialOverlap(t *testing.T) {
	a := map[string]struct{}{"a": {}, "b": {}}
	b := map[string]struct{}{"b": {}, "c": {}}
	got := jaccardSimilarity(a, b)
	// intersection=1, union=3 => 0.333
	if got < 0.3 || got > 0.4 {
		t.Errorf("expected ~0.33 for partial overlap, got %.4f", got)
	}
}

func TestJaccardSimilarityEmptyInputs(t *testing.T) {
	empty := map[string]struct{}{}
	nonEmpty := map[string]struct{}{"x": {}}
	if got := jaccardSimilarity(empty, nonEmpty); got != 0.0 {
		t.Errorf("expected 0 when left is empty, got %.2f", got)
	}
	if got := jaccardSimilarity(nonEmpty, empty); got != 0.0 {
		t.Errorf("expected 0 when right is empty, got %.2f", got)
	}
}

func TestAdapterSupportsURLPatterns(t *testing.T) {
	cases := []struct {
		adapter  string
		url      string
		wantTrue bool
	}{
		{"github-issue", "https://github.com/owner/repo/issues/1", true},
		{"github-issue", "https://gitlab.com/owner/repo/-/issues/1", false},
		{"gitlab-issue", "https://gitlab.com/owner/repo/-/issues/1", true},
		{"gitlab-issue", "https://github.com/owner/repo/issues/1", false},
		{"stackexchange-question", "https://stackoverflow.com/questions/123", true},
		{"discourse-topic", "https://discuss.example.org/t/topic/123", true},
		{"reddit-post", "https://www.reddit.com/r/golang/comments/abc/", true},
	}
	for _, tc := range cases {
		adapter, err := adapterByName(tc.adapter)
		if err != nil {
			t.Fatalf("adapterByName(%q): %v", tc.adapter, err)
		}
		got := adapter.Supports(tc.url)
		if got != tc.wantTrue {
			t.Errorf("%s.Supports(%q) = %v, want %v", tc.adapter, tc.url, got, tc.wantTrue)
		}
	}
}

func TestFirstNonEmptyFixtures(t *testing.T) {
	if got := firstNonEmpty("", "  ", "hello"); got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
	if got := firstNonEmpty("", "  "); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
	if got := firstNonEmpty("first", "second"); got != "first" {
		t.Errorf("expected 'first', got %q", got)
	}
}

func TestResolveLayoutSetsExpectedPaths(t *testing.T) {
	layout, err := ResolveLayout("/tmp")
	if err != nil {
		t.Fatalf("ResolveLayout: %v", err)
	}
	if !strings.HasSuffix(layout.Fixtures, "/fixtures") {
		t.Errorf("expected Fixtures to end with /fixtures, got %q", layout.Fixtures)
	}
	if !strings.HasSuffix(layout.MinimalDir, "/minimal") {
		t.Errorf("expected MinimalDir to end with /minimal, got %q", layout.MinimalDir)
	}
}

func TestResolveLayoutEmptyUsesCurrentDir(t *testing.T) {
	layout, err := ResolveLayout("")
	if err != nil {
		t.Fatalf("ResolveLayout with empty root: %v", err)
	}
	if layout.Root == "" {
		t.Error("expected non-empty Root")
	}
}
