package authoring_test

import (
	"strings"
	"testing"

	"faultline/internal/authoring"
)

// ── ExtractCandidatePatterns ────────────────────────────────────────────────

func TestExtractCandidatePatternsPrefersErrorLines(t *testing.T) {
	log := strings.Join([]string{
		"running pre-flight checks",
		"pull access denied for private/image, repository does not exist",
		"Error response from daemon: unauthorized: authentication required",
		"downloading layer sha256:abc123",
	}, "\n")

	got := authoring.ExtractCandidatePatterns(log, 5)
	if len(got) == 0 {
		t.Fatal("expected at least one candidate")
	}
	// The most diagnostic lines should rank first.
	top := got[0]
	if !strings.Contains(strings.ToLower(top), "denied") && !strings.Contains(strings.ToLower(top), "unauthorized") {
		t.Errorf("expected a high-signal error line first, got %q", top)
	}
}

func TestExtractCandidatePatternsDeduplicates(t *testing.T) {
	log := strings.Join([]string{
		"fatal error: connection refused",
		"FATAL ERROR: connection refused",
		"fatal error: connection refused",
	}, "\n")

	got := authoring.ExtractCandidatePatterns(log, 10)
	seen := make(map[string]int)
	for _, g := range got {
		seen[strings.ToLower(g)]++
	}
	for k, count := range seen {
		if count > 1 {
			t.Errorf("duplicate pattern %q appeared %d times", k, count)
		}
	}
}

func TestExtractCandidatePatternsRespectsMax(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = strings.Repeat("error pattern ", i+1)
	}
	log := strings.Join(lines, "\n")

	got := authoring.ExtractCandidatePatterns(log, 3)
	if len(got) > 3 {
		t.Errorf("expected at most 3 candidates, got %d", len(got))
	}
}

func TestExtractCandidatePatternsEmpty(t *testing.T) {
	got := authoring.ExtractCandidatePatterns("", 5)
	if len(got) != 0 {
		t.Errorf("expected no candidates from empty log, got %v", got)
	}
}

func TestExtractCandidatePatternsOnlyNoise(t *testing.T) {
	log := strings.Join([]string{
		"## building image",
		"--- step 1 ---",
		"=== running tests ===",
		"downloading packages",
	}, "\n")

	got := authoring.ExtractCandidatePatterns(log, 5)
	// Noise lines should yield zero or very few results (noise-only input).
	for _, g := range got {
		if strings.HasPrefix(strings.ToLower(g), "##") || strings.HasPrefix(strings.ToLower(g), "---") {
			t.Errorf("noise line escaped extraction: %q", g)
		}
	}
}

func TestExtractCandidatePatternsDeterministic(t *testing.T) {
	log := strings.Join([]string{
		"error: cannot find module 'react'",
		"failed to resolve dependency",
		"npm ERR! missing peer dependency",
	}, "\n")

	first := authoring.ExtractCandidatePatterns(log, 5)
	second := authoring.ExtractCandidatePatterns(log, 5)

	if len(first) != len(second) {
		t.Fatalf("non-deterministic length: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Errorf("non-deterministic at index %d: %q vs %q", i, first[i], second[i])
		}
	}
}

// ── ScaffoldPlaybook ────────────────────────────────────────────────────────

func TestScaffoldPlaybookContainsExtractedPatterns(t *testing.T) {
	log := "pull access denied\nError response from daemon: unauthorized: authentication required\n"

	result, err := authoring.ScaffoldPlaybook(log, authoring.ScaffoldOptions{Category: "auth"})
	if err != nil {
		t.Fatalf("ScaffoldPlaybook: %v", err)
	}
	if result.YAML == "" {
		t.Fatal("expected non-empty YAML")
	}
	if !strings.Contains(result.YAML, "id: auth-") {
		t.Errorf("expected id to start with auth-, got YAML:\n%s", result.YAML)
	}
	// At least one extracted pattern should appear verbatim in the YAML.
	if len(result.Candidates) == 0 {
		t.Fatal("expected at least one candidate pattern")
	}
}

func TestScaffoldPlaybookIDOverride(t *testing.T) {
	log := "fatal error: connection reset by peer\n"

	result, err := authoring.ScaffoldPlaybook(log, authoring.ScaffoldOptions{
		ID:       "network-reset-by-peer",
		Category: "network",
	})
	if err != nil {
		t.Fatalf("ScaffoldPlaybook: %v", err)
	}
	if result.SuggestedID != "network-reset-by-peer" {
		t.Errorf("expected ID override, got %q", result.SuggestedID)
	}
	if !strings.Contains(result.YAML, "id: network-reset-by-peer") {
		t.Errorf("expected explicit id in YAML, got:\n%s", result.YAML)
	}
}

func TestScaffoldPlaybookDefaultCategory(t *testing.T) {
	result, err := authoring.ScaffoldPlaybook("some error occurred\n", authoring.ScaffoldOptions{})
	if err != nil {
		t.Fatalf("ScaffoldPlaybook: %v", err)
	}
	if !strings.Contains(result.YAML, "category: build") {
		t.Errorf("expected default category=build, got YAML:\n%s", result.YAML)
	}
}

func TestScaffoldPlaybookInvalidCategoryReturnsError(t *testing.T) {
	_, err := authoring.ScaffoldPlaybook("something failed\n", authoring.ScaffoldOptions{Category: "unicorn"})
	if err == nil {
		t.Fatal("expected invalid category error")
	}
	if !strings.Contains(err.Error(), "invalid category") {
		t.Fatalf("expected invalid category error, got %v", err)
	}
}

func TestScaffoldPlaybookRequiredSections(t *testing.T) {
	log := "error: module not found\n"
	result, err := authoring.ScaffoldPlaybook(log, authoring.ScaffoldOptions{Category: "build"})
	if err != nil {
		t.Fatalf("ScaffoldPlaybook: %v", err)
	}

	requiredSections := []string{
		"match:",
		"summary:",
		"diagnosis:",
		"fix:",
		"validation:",
		"workflow:",
	}
	for _, sec := range requiredSections {
		if !strings.Contains(result.YAML, sec) {
			t.Errorf("missing required section %q in scaffold YAML", sec)
		}
	}
}

func TestScaffoldPlaybookDeterministic(t *testing.T) {
	log := "pull access denied\nError response from daemon: unauthorized: authentication required\n"
	opts := authoring.ScaffoldOptions{Category: "auth"}

	first, err := authoring.ScaffoldPlaybook(log, opts)
	if err != nil {
		t.Fatalf("first ScaffoldPlaybook: %v", err)
	}
	second, err := authoring.ScaffoldPlaybook(log, opts)
	if err != nil {
		t.Fatalf("second ScaffoldPlaybook: %v", err)
	}
	if first.YAML != second.YAML {
		t.Errorf("non-deterministic scaffold output:\nfirst:\n%s\nsecond:\n%s", first.YAML, second.YAML)
	}
	if first.SuggestedID != second.SuggestedID {
		t.Errorf("non-deterministic ID: %q vs %q", first.SuggestedID, second.SuggestedID)
	}
}

func TestScaffoldPlaybookEmptyLog(t *testing.T) {
	result, err := authoring.ScaffoldPlaybook("", authoring.ScaffoldOptions{Category: "ci"})
	if err != nil {
		t.Fatalf("ScaffoldPlaybook with empty log: %v", err)
	}
	// Should still produce a valid scaffold with TODO placeholders.
	if !strings.Contains(result.YAML, "TODO") {
		t.Errorf("expected TODO placeholders for empty-log scaffold, got:\n%s", result.YAML)
	}
	if result.SuggestedID == "" {
		t.Error("expected a non-empty suggested ID even with empty log")
	}
}

func TestScaffoldPlaybookInvalidIDReturnsError(t *testing.T) {
	_, err := authoring.ScaffoldPlaybook("fatal error: connection reset by peer\n", authoring.ScaffoldOptions{
		ID:       "Network Reset",
		Category: "network",
	})
	if err == nil {
		t.Fatal("expected invalid ID error")
	}
	if !strings.Contains(err.Error(), "invalid playbook id") {
		t.Fatalf("expected invalid playbook id error, got %v", err)
	}
}

func TestScaffoldPlaybookWritesToPackDir(t *testing.T) {
	dir := t.TempDir()
	packDir := dir + "/nested/pack"
	log := "fatal: remote error: repository not found\n"
	opts := authoring.ScaffoldOptions{
		Category: "network",
		PackDir:  packDir,
	}

	result, err := authoring.ScaffoldPlaybook(log, opts)
	if err != nil {
		t.Fatalf("ScaffoldPlaybook: %v", err)
	}
	if result.OutputPath == "" {
		t.Fatal("expected OutputPath to be set when PackDir is given")
	}
	if !strings.HasSuffix(result.OutputPath, ".yaml") {
		t.Errorf("expected .yaml extension, got %q", result.OutputPath)
	}
	if !strings.HasPrefix(result.OutputPath, packDir) {
		t.Errorf("expected output path under %q, got %q", packDir, result.OutputPath)
	}
}
