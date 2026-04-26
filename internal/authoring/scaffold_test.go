package authoring_test

import (
	"os"
	"strings"
	"testing"

	"faultline/internal/authoring"

	"gopkg.in/yaml.v3"
)

type scaffoldYAML struct {
	Category string `yaml:"category"`
	Match    struct {
		Any []string `yaml:"any"`
	} `yaml:"match"`
}

func parseScaffoldYAML(t *testing.T, text string) scaffoldYAML {
	t.Helper()

	var got scaffoldYAML
	if err := yaml.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("unmarshal scaffold yaml: %v\n%s", err, text)
	}
	return got
}

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

func TestExtractCandidatePatternsDefaultMaxAndAlphabeticalTieBreak(t *testing.T) {
	log := strings.Join([]string{
		"cannot resolve workspace lockfile",
		"cannot open workspace lockfile",
		"cannot parse workspace lockfile",
		"cannot read workspace lockfile",
		"cannot stat workspace lockfile",
		"cannot create workspace lockfile",
	}, "\n")

	got := authoring.ExtractCandidatePatterns(log, 0)
	want := []string{
		"cannot create workspace lockfile",
		"cannot open workspace lockfile",
		"cannot parse workspace lockfile",
		"cannot read workspace lockfile",
		"cannot resolve workspace lockfile",
	}

	if len(got) != len(want) {
		t.Fatalf("expected default max of %d candidates, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected alphabetical tie-break at index %d: want %q got %q", i, want[i], got[i])
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

func TestScaffoldPlaybookNormalizesCategoryAndPreservesAllCandidates(t *testing.T) {
	log := strings.Join([]string{
		"fatal error: pull access denied for private/image",
		"Error response from daemon: unauthorized: authentication required",
		"permission denied while fetching image metadata",
	}, "\n")

	result, err := authoring.ScaffoldPlaybook(log, authoring.ScaffoldOptions{
		Category: " auth ",
		MaxMatch: 1,
	})
	if err != nil {
		t.Fatalf("ScaffoldPlaybook: %v", err)
	}

	parsed := parseScaffoldYAML(t, result.YAML)
	if parsed.Category != "auth" {
		t.Fatalf("expected trimmed category auth, got %q", parsed.Category)
	}
	if len(parsed.Match.Any) != 1 {
		t.Fatalf("expected one emitted match pattern, got %d", len(parsed.Match.Any))
	}
	if len(result.Candidates) < 2 {
		t.Fatalf("expected full candidate list to be preserved, got %v", result.Candidates)
	}
	if parsed.Match.Any[0] != result.Candidates[0] {
		t.Fatalf("expected emitted pattern %q to match top candidate %q", parsed.Match.Any[0], result.Candidates[0])
	}
}

func TestScaffoldPlaybookDefaultMaxMatchUsesFivePatterns(t *testing.T) {
	log := strings.Join([]string{
		"cannot create workspace lockfile",
		"cannot open workspace lockfile",
		"cannot parse workspace lockfile",
		"cannot read workspace lockfile",
		"cannot resolve workspace lockfile",
		"cannot stat workspace lockfile",
	}, "\n")

	result, err := authoring.ScaffoldPlaybook(log, authoring.ScaffoldOptions{Category: "build"})
	if err != nil {
		t.Fatalf("ScaffoldPlaybook: %v", err)
	}

	parsed := parseScaffoldYAML(t, result.YAML)
	if len(parsed.Match.Any) != 5 {
		t.Fatalf("expected default match cap of 5, got %d: %v", len(parsed.Match.Any), parsed.Match.Any)
	}
}

func TestScaffoldPlaybookYAMLPreservesQuotedPatterns(t *testing.T) {
	log := "Error: invalid image \"repo/app:latest\" # tag mismatch\n"

	result, err := authoring.ScaffoldPlaybook(log, authoring.ScaffoldOptions{Category: "build", MaxMatch: 1})
	if err != nil {
		t.Fatalf("ScaffoldPlaybook: %v", err)
	}

	parsed := parseScaffoldYAML(t, result.YAML)
	if len(parsed.Match.Any) != 1 {
		t.Fatalf("expected one match pattern, got %d", len(parsed.Match.Any))
	}
	if parsed.Match.Any[0] != `Error: invalid image "repo/app:latest" # tag mismatch` {
		t.Fatalf("expected yaml round-trip to preserve pattern, got %q", parsed.Match.Any[0])
	}
}

func TestScaffoldPlaybookWritesExactYAMLToDisk(t *testing.T) {
	packDir := t.TempDir()

	result, err := authoring.ScaffoldPlaybook(
		"fatal error: connection reset by peer\n",
		authoring.ScaffoldOptions{Category: "network", PackDir: packDir},
	)
	if err != nil {
		t.Fatalf("ScaffoldPlaybook: %v", err)
	}

	content, err := os.ReadFile(result.OutputPath)
	if err != nil {
		t.Fatalf("read scaffold output: %v", err)
	}
	if string(content) != result.YAML {
		t.Fatalf("expected written scaffold to match returned YAML\nwritten:\n%s\nreturned:\n%s", string(content), result.YAML)
	}
}
