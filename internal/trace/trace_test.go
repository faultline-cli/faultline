package trace

import (
	"strings"
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

// ── buildCompeting ────────────────────────────────────────────────────────────

func TestBuildCompetingNilAnalysis(t *testing.T) {
	got := buildCompeting(nil, "some-id")
	if got != nil {
		t.Errorf("expected nil for nil analysis, got %v", got)
	}
}

func TestBuildCompetingNoDifferential(t *testing.T) {
	got := buildCompeting(&model.Analysis{}, "some-id")
	if got != nil {
		t.Errorf("expected nil when no Differential, got %v", got)
	}
}

func TestBuildCompetingLikelyDifferentID(t *testing.T) {
	a := &model.Analysis{
		Differential: &model.DifferentialDiagnosis{
			Likely: &model.DifferentialCandidate{
				FailureID:      "other-failure",
				Title:          "Other Failure",
				Confidence:     0.9,
				ConfidenceText: "high",
				Why:            []string{"strong signal"},
			},
		},
	}
	got := buildCompeting(a, "current-id")
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(got))
	}
	c := got[0]
	if c.Status != "higher_ranked" {
		t.Errorf("Status = %q, want %q", c.Status, "higher_ranked")
	}
	if c.FailureID != "other-failure" {
		t.Errorf("FailureID = %q, want %q", c.FailureID, "other-failure")
	}
	if len(c.Reasons) != 1 || c.Reasons[0] != "strong signal" {
		t.Errorf("Reasons = %v, want [strong signal]", c.Reasons)
	}
}

func TestBuildCompetingLikelySameIDSkipped(t *testing.T) {
	a := &model.Analysis{
		Differential: &model.DifferentialDiagnosis{
			Likely: &model.DifferentialCandidate{
				FailureID: "current-id",
				Title:     "Current",
			},
		},
	}
	got := buildCompeting(a, "current-id")
	if len(got) != 0 {
		t.Errorf("expected 0 candidates when Likely is current ID, got %d", len(got))
	}
}

func TestBuildCompetingAlternativesSomeSkipped(t *testing.T) {
	a := &model.Analysis{
		Differential: &model.DifferentialDiagnosis{
			Alternatives: []model.DifferentialCandidate{
				{FailureID: "current-id", Title: "skip me"},
				{FailureID: "alt-1", Title: "Alt 1", WhyLessLikely: []string{"weaker signal"}},
			},
		},
	}
	got := buildCompeting(a, "current-id")
	if len(got) != 1 {
		t.Fatalf("expected 1 alternative, got %d", len(got))
	}
	if got[0].Status != "alternative" {
		t.Errorf("Status = %q, want %q", got[0].Status, "alternative")
	}
	if got[0].FailureID != "alt-1" {
		t.Errorf("FailureID = %q, want %q", got[0].FailureID, "alt-1")
	}
	if len(got[0].Reasons) != 1 || got[0].Reasons[0] != "weaker signal" {
		t.Errorf("Reasons = %v, want [weaker signal]", got[0].Reasons)
	}
}

func TestBuildCompetingRuledOut(t *testing.T) {
	a := &model.Analysis{
		Differential: &model.DifferentialDiagnosis{
			RuledOut: []model.DifferentialCandidate{
				{FailureID: "ruled-out-1", Title: "Ruled Out", RuledOutBy: []string{"evidence X"}},
			},
		},
	}
	got := buildCompeting(a, "current-id")
	if len(got) != 1 {
		t.Fatalf("expected 1 ruled-out candidate, got %d", len(got))
	}
	if got[0].Status != "ruled_out" {
		t.Errorf("Status = %q, want %q", got[0].Status, "ruled_out")
	}
	if len(got[0].Reasons) != 1 || got[0].Reasons[0] != "evidence X" {
		t.Errorf("Reasons = %v, want [evidence X]", got[0].Reasons)
	}
}

// ── matchLinesForPatterns ─────────────────────────────────────────────────────

func TestMatchLinesForPatternsEmpty(t *testing.T) {
	lines := makeLines("exec node: no such file or directory")
	got, hits := matchLinesForPatterns(nil, lines)
	if len(got) != 0 {
		t.Errorf("expected no matches for empty patterns, got %v", got)
	}
	if hits != 0 {
		t.Errorf("expected 0 hits for empty patterns, got %d", hits)
	}
}

func TestMatchLinesForPatternsSingleMatch(t *testing.T) {
	lines := makeLines("npm install failed", "some other line")
	got, hits := matchLinesForPatterns([]string{"npm install failed"}, lines)
	if len(got) != 1 {
		t.Fatalf("expected 1 match, got %d", len(got))
	}
	if hits != 1 {
		t.Errorf("expected 1 hit, got %d", hits)
	}
	if got[0].Number != 1 {
		t.Errorf("expected line number 1, got %d", got[0].Number)
	}
}

func TestMatchLinesForPatternsDedupedByLineNumber(t *testing.T) {
	// Two patterns both match line 1 — output deduped, hits counted twice.
	lines := makeLines("authentication required: docker pull")
	patterns := []string{"authentication required", "docker pull"}
	got, hits := matchLinesForPatterns(patterns, lines)
	if len(got) != 1 {
		t.Errorf("expected 1 deduped match, got %d", len(got))
	}
	if hits != 2 {
		t.Errorf("expected 2 hits (one per pattern), got %d", hits)
	}
}

func TestMatchLinesForPatternsTwoPatternsDistinctLines(t *testing.T) {
	lines := makeLines("error: connection refused", "warning: timeout exceeded")
	patterns := []string{"connection refused", "timeout exceeded"}
	got, hits := matchLinesForPatterns(patterns, lines)
	if len(got) != 2 {
		t.Errorf("expected 2 matches, got %d", len(got))
	}
	if hits != 2 {
		t.Errorf("expected 2 hits, got %d", hits)
	}
}

// ── partialRulePattern ────────────────────────────────────────────────────────

func TestPartialRulePattern(t *testing.T) {
	cases := []struct {
		name  string
		group model.PartialMatchGroup
		want  string
	}{
		{
			name: "with label",
			group: model.PartialMatchGroup{
				Label:    "auth-signals",
				Minimum:  2,
				Patterns: []string{"login failed", "unauthorized", "401"},
			},
			want: "auth-signals (2-of-3): login failed | unauthorized | 401",
		},
		{
			name: "with ID but no label",
			group: model.PartialMatchGroup{
				ID:       "group-id",
				Minimum:  1,
				Patterns: []string{"pat1", "pat2"},
			},
			want: "group-id (1-of-2): pat1 | pat2",
		},
		{
			name: "no label no ID",
			group: model.PartialMatchGroup{
				Minimum:  2,
				Patterns: []string{"a", "b", "c"},
			},
			want: "2-of-3: a | b | c",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := partialRulePattern(tc.group)
			if got != tc.want {
				t.Errorf("partialRulePattern() = %q, want %q", got, tc.want)
			}
		})
	}
}

// ── partialRuleNote ───────────────────────────────────────────────────────────

func TestPartialRuleNote(t *testing.T) {
	cases := []struct {
		name    string
		group   model.PartialMatchGroup
		hits    int
		wantSub string
	}{
		{
			name:    "threshold reached",
			group:   model.PartialMatchGroup{Minimum: 2, Patterns: []string{"a", "b", "c"}},
			hits:    2,
			wantSub: "reached threshold with 2/3 matched patterns",
		},
		{
			name:    "threshold exceeded",
			group:   model.PartialMatchGroup{Minimum: 2, Patterns: []string{"a", "b", "c"}},
			hits:    3,
			wantSub: "reached threshold with 3/3 matched patterns",
		},
		{
			name:    "below threshold",
			group:   model.PartialMatchGroup{Minimum: 3, Patterns: []string{"a", "b", "c", "d"}},
			hits:    1,
			wantSub: "matched 1/4 patterns and stayed below the 3-pattern threshold",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := partialRuleNote(tc.group, tc.hits)
			if !strings.Contains(got, tc.wantSub) {
				t.Errorf("partialRuleNote() = %q, want substring %q", got, tc.wantSub)
			}
		})
	}
}
