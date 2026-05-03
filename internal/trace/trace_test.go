package trace

import (
	"testing"

	"faultline/internal/model"
)

// makeLines builds []model.Line from raw strings, normalising via lowercase.
func makeLines(raw ...string) []model.Line {
	lines := make([]model.Line, len(raw))
	for i, s := range raw {
		lines[i] = model.Line{
			Original:   s,
			Normalized: makeNormalized(s),
			Number:     i + 1,
		}
	}
	return lines
}

// makeNormalized applies the same simple lower-case rule the engine uses for
// log lines. (We avoid importing engine to keep the test dependency-light.)
func makeNormalized(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	// Very simple whitespace collapse is done by the engine too, but for
	// these tests single-token patterns suffice without it.
	return string(result)
}

// ── resolvePlaybook ───────────────────────────────────────────────────────────

func TestResolvePlaybookEmptyIDUsesTopResult(t *testing.T) {
	pb := model.Playbook{ID: "docker-auth", Title: "Docker auth"}
	analysis := &model.Analysis{
		Results: []model.Result{
			{Playbook: pb, Score: 0.9, Confidence: 0.85},
		},
	}
	got, result, rank, err := resolvePlaybook(analysis, nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "docker-auth" {
		t.Errorf("expected playbook ID 'docker-auth', got %q", got.ID)
	}
	if result == nil {
		t.Fatalf("expected non-nil result")
	}
	if result.Score != 0.9 {
		t.Errorf("expected score 0.9, got %f", result.Score)
	}
	if rank != 1 {
		t.Errorf("expected rank 1, got %d", rank)
	}
}

func TestResolvePlaybookExplicitIDFoundInResults(t *testing.T) {
	pb1 := model.Playbook{ID: "alpha", Title: "Alpha"}
	pb2 := model.Playbook{ID: "beta", Title: "Beta"}
	analysis := &model.Analysis{
		Results: []model.Result{
			{Playbook: pb1, Score: 0.9},
			{Playbook: pb2, Score: 0.7},
		},
	}
	got, result, rank, err := resolvePlaybook(analysis, nil, "beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "beta" {
		t.Errorf("expected 'beta', got %q", got.ID)
	}
	if result == nil || result.Score != 0.7 {
		t.Errorf("expected result with score 0.7, got %v", result)
	}
	if rank != 2 {
		t.Errorf("expected rank 2, got %d", rank)
	}
}

func TestResolvePlaybookExplicitIDFoundInList(t *testing.T) {
	// ID exists in playbooks list but not in analysis results (unmatched playbook).
	pb := model.Playbook{ID: "orphan", Title: "Orphan"}
	analysis := &model.Analysis{
		Results: []model.Result{
			{Playbook: model.Playbook{ID: "other"}, Score: 0.5},
		},
	}
	got, result, rank, err := resolvePlaybook(analysis, []model.Playbook{pb}, "orphan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "orphan" {
		t.Errorf("expected 'orphan', got %q", got.ID)
	}
	if result != nil {
		t.Errorf("expected nil result for unmatched playbook, got %v", result)
	}
	if rank != 0 {
		t.Errorf("expected rank 0 for unmatched playbook, got %d", rank)
	}
}

func TestResolvePlaybookIDNotFoundReturnsError(t *testing.T) {
	_, _, _, err := resolvePlaybook(nil, nil, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown playbook ID, got nil")
	}
}

func TestResolvePlaybookEmptyIDNilAnalysisReturnsError(t *testing.T) {
	_, _, _, err := resolvePlaybook(nil, nil, "")
	if err == nil {
		t.Fatal("expected error for empty ID with nil analysis, got nil")
	}
}

func TestResolvePlaybookEmptyIDNoResultsReturnsError(t *testing.T) {
	analysis := &model.Analysis{Results: nil}
	_, _, _, err := resolvePlaybook(analysis, nil, "")
	if err == nil {
		t.Fatal("expected error when analysis has no results and ID is empty, got nil")
	}
}

// ── buildRules ────────────────────────────────────────────────────────────────

func TestBuildRulesMatchAnyMatched(t *testing.T) {
	pb := model.Playbook{
		ID: "test",
		Match: model.MatchSpec{
			Any: []string{"exec: foo: no such file"},
		},
	}
	lines := makeLines("exec: foo: no such file or directory")
	rules := buildRules(pb, lines)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	r := rules[0]
	if r.Group != "match.any" {
		t.Errorf("expected group 'match.any', got %q", r.Group)
	}
	if r.Status != StatusMatched {
		t.Errorf("expected StatusMatched, got %q", r.Status)
	}
	if !r.Matched {
		t.Error("expected Matched=true")
	}
	if len(r.LineMatches) == 0 {
		t.Error("expected at least one line match")
	}
}

func TestBuildRulesMatchAnyMissing(t *testing.T) {
	pb := model.Playbook{
		ID: "test",
		Match: model.MatchSpec{
			Any: []string{"pattern-that-will-not-match"},
		},
	}
	lines := makeLines("totally unrelated log line")
	rules := buildRules(pb, lines)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Status != StatusMissing {
		t.Errorf("expected StatusMissing, got %q", rules[0].Status)
	}
	if rules[0].Matched {
		t.Error("expected Matched=false")
	}
}

func TestBuildRulesMatchNoneBlocked(t *testing.T) {
	pb := model.Playbook{
		ID: "test",
		Match: model.MatchSpec{
			None: []string{"cache hit"},
		},
	}
	lines := makeLines("cache hit restored properly")
	rules := buildRules(pb, lines)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Group != "match.none" {
		t.Errorf("expected group 'match.none', got %q", rules[0].Group)
	}
	if rules[0].Status != StatusBlocked {
		t.Errorf("expected StatusBlocked, got %q", rules[0].Status)
	}
	if !rules[0].Matched {
		t.Error("expected Matched=true for a triggered none-rule")
	}
}

func TestBuildRulesMatchNoneClear(t *testing.T) {
	pb := model.Playbook{
		ID: "test",
		Match: model.MatchSpec{
			None: []string{"cache hit"},
		},
	}
	lines := makeLines("unrelated log line here")
	rules := buildRules(pb, lines)
	if rules[0].Status != StatusClear {
		t.Errorf("expected StatusClear, got %q", rules[0].Status)
	}
}

func TestBuildRulesEmptyPlaybook(t *testing.T) {
	pb := model.Playbook{ID: "empty"}
	rules := buildRules(pb, nil)
	if len(rules) != 0 {
		t.Errorf("expected 0 rules for empty playbook, got %d", len(rules))
	}
}

func TestBuildRulesRelevanceFields(t *testing.T) {
	pb := model.Playbook{
		ID: "test",
		Match: model.MatchSpec{
			Any: []string{"trigger-pattern"},
			All: []string{"required-pattern"},
		},
	}
	rules := buildRules(pb, nil)
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}

	anyRule := rules[0]
	if anyRule.Relevance != "trigger" {
		t.Errorf("expected match.any relevance 'trigger', got %q", anyRule.Relevance)
	}

	allRule := rules[1]
	if allRule.Relevance != "required" {
		t.Errorf("expected match.all relevance 'required', got %q", allRule.Relevance)
	}
}

// ── Build ─────────────────────────────────────────────────────────────────────

func TestBuildWithNilAnalysisAndExplicitPlaybook(t *testing.T) {
	pb := model.Playbook{
		ID:      "my-playbook",
		Title:   "My Playbook",
		Summary: "test summary",
		Match: model.MatchSpec{
			Any: []string{"exec: foo:"},
		},
	}
	lines := makeLines("exec: foo: no such file or directory")

	report, err := Build(nil, lines, []model.Playbook{pb}, "my-playbook", false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if report.Playbook.ID != "my-playbook" {
		t.Errorf("expected playbook ID 'my-playbook', got %q", report.Playbook.ID)
	}
	if report.Matched {
		t.Error("expected Matched=false when analysis is nil")
	}
	if len(report.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(report.Rules))
	}
}

func TestBuildWithMatchedAnalysis(t *testing.T) {
	pb := model.Playbook{
		ID:      "docker-auth",
		Title:   "Docker auth failure",
		Summary: "Docker credentials not configured",
		Match: model.MatchSpec{
			Any: []string{"authentication required"},
		},
	}
	result := model.Result{
		Playbook:   pb,
		Score:      0.9,
		Confidence: 0.85,
		Detector:   "log",
		Evidence:   []string{"authentication required"},
	}
	analysis := &model.Analysis{
		Source:      "ci.log",
		Fingerprint: "fp-abc",
		Results:     []model.Result{result},
	}
	lines := makeLines("authentication required for registry.example.com")

	report, err := Build(analysis, lines, []model.Playbook{pb}, "", false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !report.Matched {
		t.Error("expected Matched=true")
	}
	if report.Score != 0.9 {
		t.Errorf("expected score 0.9, got %f", report.Score)
	}
	if report.Source != "ci.log" {
		t.Errorf("expected source 'ci.log', got %q", report.Source)
	}
	if report.Rank != 1 {
		t.Errorf("expected rank 1, got %d", report.Rank)
	}
}

func TestBuildUnknownPlaybookIDReturnsError(t *testing.T) {
	_, err := Build(nil, nil, nil, "nonexistent-id", false)
	if err == nil {
		t.Fatal("expected error for unknown playbook ID, got nil")
	}
}

func TestBuildEmptyIDWithNilAnalysisReturnsError(t *testing.T) {
	_, err := Build(nil, nil, nil, "", false)
	if err == nil {
		t.Fatal("expected error for empty ID with nil analysis, got nil")
	}
}
