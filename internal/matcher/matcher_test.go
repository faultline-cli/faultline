package matcher

import (
	"testing"

	"faultline/internal/model"
)

func TestRankAnyPatternScoring(t *testing.T) {
	playbooks := []model.Playbook{
		{
			ID:       "alpha",
			Title:    "Alpha",
			Category: "ci",
			Match:    model.MatchSpec{Any: []string{"foo", "bar"}},
		},
		{
			ID:       "beta",
			Title:    "Beta",
			Category: "ci",
			Match:    model.MatchSpec{Any: []string{"foo", "bar", "baz"}},
		},
	}
	lines := []model.Line{
		{Original: "Foo exploded", Normalized: "foo exploded"},
		{Original: "Bar exploded", Normalized: "bar exploded"},
	}

	results := Rank(playbooks, lines, model.Context{})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Playbook.ID != "alpha" {
		t.Fatalf("expected alpha to rank first, got %s", results[0].Playbook.ID)
	}
	// foo and bar are each shared by both playbooks → IDF weight 0.5 each.
	// alpha hits both → score = 0 + 0.5 + 0.5 = 1.0.
	if results[0].Score != 1.0 {
		t.Fatalf("expected score 1.0, got %v", results[0].Score)
	}
	if len(results[0].Evidence) != 2 {
		t.Fatalf("expected 2 evidence lines, got %d", len(results[0].Evidence))
	}
}

func TestRankAllPatternBonus(t *testing.T) {
	pb := model.Playbook{
		ID:    "test-all",
		Title: "Test All",
		Match: model.MatchSpec{All: []string{"error", "timeout"}},
	}
	lines := []model.Line{
		{Original: "error occurred", Normalized: "error occurred"},
		{Original: "connection timeout", Normalized: "connection timeout"},
	}
	results := Rank([]model.Playbook{pb}, lines, model.Context{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Score != 5.0 {
		t.Fatalf("expected score 5.0, got %v", results[0].Score)
	}
	if results[0].Confidence != 0.92 {
		t.Fatalf("expected confidence 0.92, got %v", results[0].Confidence)
	}
}

func TestRankStageHintBonus(t *testing.T) {
	pb := model.Playbook{
		ID:         "deploy-err",
		Title:      "Deploy Error",
		Match:      model.MatchSpec{Any: []string{"failed"}},
		StageHints: []string{"deploy"},
	}
	lines := []model.Line{{Original: "deploy failed", Normalized: "deploy failed"}}

	results := Rank([]model.Playbook{pb}, lines, model.Context{Stage: "deploy"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Score != 1.75 {
		t.Fatalf("expected score 1.75, got %v", results[0].Score)
	}
	if results[0].Confidence != 0.82 {
		t.Fatalf("expected confidence 0.82, got %v", results[0].Confidence)
	}
}

func TestRankBaseScoreAdded(t *testing.T) {
	pb := model.Playbook{
		ID:        "base",
		Title:     "Base",
		BaseScore: 2.0,
		Match:     model.MatchSpec{Any: []string{"error"}},
	}
	lines := []model.Line{{Original: "error", Normalized: "error"}}

	results := Rank([]model.Playbook{pb}, lines, model.Context{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Score != 3.0 {
		t.Fatalf("expected score 3.0, got %v", results[0].Score)
	}
}

func TestRankNoMatchReturnsEmpty(t *testing.T) {
	pb := model.Playbook{
		ID:    "docker-auth",
		Title: "Docker Auth",
		Match: model.MatchSpec{Any: []string{"authentication required"}},
	}
	lines := []model.Line{{Original: "all green", Normalized: "all green"}}

	results := Rank([]model.Playbook{pb}, lines, model.Context{})
	if len(results) != 0 {
		t.Fatalf("expected no results, got %d", len(results))
	}
}

func TestRankPartialAllPatterns(t *testing.T) {
	pb := model.Playbook{
		ID:    "partial-all",
		Title: "Partial All",
		Match: model.MatchSpec{All: []string{"error", "timeout", "missing"}},
	}
	lines := []model.Line{
		{Original: "error here", Normalized: "error here"},
		{Original: "connection timeout", Normalized: "connection timeout"},
	}

	results := Rank([]model.Playbook{pb}, lines, model.Context{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result (partial match), got %d", len(results))
	}
	if results[0].Score != 3.0 {
		t.Fatalf("expected score 3.0, got %v", results[0].Score)
	}
}

func TestRankPartialGroupThreshold(t *testing.T) {
	pb := model.Playbook{
		ID:    "partial-group",
		Title: "Partial Group",
		Match: model.MatchSpec{
			Partial: []model.PartialMatchGroup{
				{
					ID:      "node-env",
					Minimum: 2,
					Patterns: []string{
						".nvmrc",
						"engines.node",
						"node version",
					},
				},
			},
		},
	}
	lines := []model.Line{
		{Original: "project has .nvmrc checked in", Normalized: "project has .nvmrc checked in"},
		{Original: "package.json engines.node requires 20", Normalized: "package.json engines.node requires 20"},
	}

	results := Rank([]model.Playbook{pb}, lines, model.Context{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Score != 2.5 {
		t.Fatalf("expected score 2.5, got %v", results[0].Score)
	}
	if len(results[0].Evidence) != 2 {
		t.Fatalf("expected 2 evidence lines, got %d", len(results[0].Evidence))
	}
}

func TestRankNonePatternExcludesPlaybook(t *testing.T) {
	playbooks := []model.Playbook{
		{
			ID:    "generic-timeout",
			Title: "Generic timeout",
			Match: model.MatchSpec{
				Any:  []string{"timed out"},
				None: []string{"no such host"},
			},
		},
		{
			ID:    "dns-resolution",
			Title: "DNS resolution failure",
			Match: model.MatchSpec{
				Any: []string{"no such host"},
			},
		},
	}
	lines := []model.Line{
		{Original: "dial tcp: lookup registry.example.com: no such host", Normalized: "dial tcp: lookup registry.example.com: no such host"},
		{Original: "request timed out", Normalized: "request timed out"},
	}

	results := Rank(playbooks, lines, model.Context{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result after exclusion, got %d", len(results))
	}
	if results[0].Playbook.ID != "dns-resolution" {
		t.Fatalf("expected dns-resolution after exclusion, got %s", results[0].Playbook.ID)
	}
}

func TestRankOverlapImagePullBeatsGenericDockerAuth(t *testing.T) {
	playbooks := []model.Playbook{
		{
			ID:        "docker-auth",
			Title:     "Docker auth",
			BaseScore: 1.0,
			Match: model.MatchSpec{
				Any:  []string{"authentication required", "pull access denied"},
				None: []string{"failed to pull image", "imagepullbackoff", "errimagepull", "back-off pulling image"},
			},
		},
		{
			ID:        "image-pull-backoff",
			Title:     "Image pull",
			BaseScore: 1.0,
			Match: model.MatchSpec{
				Any: []string{"failed to pull image", "back-off pulling image", "pull access denied"},
			},
		},
	}
	lines := []model.Line{
		{Original: "Failed to pull image \"ghcr.io/acme/app:missing\": pull access denied", Normalized: "failed to pull image \"ghcr.io/acme/app:missing\": pull access denied"},
		{Original: "Back-off pulling image", Normalized: "back-off pulling image"},
	}

	results := Rank(playbooks, lines, model.Context{Stage: "deploy"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result after exclusion, got %d", len(results))
	}
	if results[0].Playbook.ID != "image-pull-backoff" {
		t.Fatalf("expected image-pull-backoff to rank first, got %s", results[0].Playbook.ID)
	}
}

// --- RE2 regex pattern tests ---

func TestRankRegexPatternMatches(t *testing.T) {
	pb := model.Playbook{
		ID:    "exit-nonzero",
		Title: "Non-zero exit",
		Match: model.MatchSpec{Any: []string{"re:exited? with (code )?[1-9][0-9]*"}},
	}
	lines := []model.Line{
		{Original: "Process exited with code 137", Normalized: "process exited with code 137"},
		{Original: "Process exited with 1", Normalized: "process exited with 1"},
		{Original: "Exit code: 0", Normalized: "exit code: 0"},
	}
	results := Rank([]model.Playbook{pb}, lines, model.Context{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// firstOriginal returns only the first matching line.
	if len(results[0].Evidence) != 1 {
		t.Fatalf("expected 1 evidence line, got %d", len(results[0].Evidence))
	}
	if results[0].Evidence[0] != "Process exited with code 137" {
		t.Fatalf("unexpected evidence: %q", results[0].Evidence[0])
	}
}

func TestRankRegexPatternNoMatch(t *testing.T) {
	pb := model.Playbook{
		ID:    "exit-nonzero",
		Title: "Non-zero exit",
		Match: model.MatchSpec{Any: []string{"re:exit code [1-9][0-9]*"}},
	}
	lines := []model.Line{
		{Original: "Exit code: 0", Normalized: "exit code: 0"},
	}
	results := Rank([]model.Playbook{pb}, lines, model.Context{})
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestRankRegexInvalidPatternSkipped(t *testing.T) {
	// An invalid regex must be silently skipped, not panic.
	pb := model.Playbook{
		ID:    "bad-regex",
		Title: "Bad Regex",
		Match: model.MatchSpec{Any: []string{"re:[invalid"}},
	}
	lines := []model.Line{
		{Original: "[invalid match", Normalized: "[invalid match"},
	}
	results := Rank([]model.Playbook{pb}, lines, model.Context{})
	if len(results) != 0 {
		t.Fatalf("expected 0 results for invalid regex, got %d", len(results))
	}
}

func TestRankRegexNonePatternSuppresses(t *testing.T) {
	pb := model.Playbook{
		ID:    "no-ok-lines",
		Title: "No OK Lines",
		Match: model.MatchSpec{
			Any:  []string{"failed"},
			None: []string{"re:ok|success|passed"},
		},
	}

	// With a suppressing line present the playbook must not match.
	linesWithOk := []model.Line{
		{Original: "failed to deploy", Normalized: "failed to deploy"},
		{Original: "ok - all passed", Normalized: "ok - all passed"},
	}
	if got := Rank([]model.Playbook{pb}, linesWithOk, model.Context{}); len(got) != 0 {
		t.Fatalf("expected 0 results (none regex suppresses), got %d", len(got))
	}

	// Without the suppressing line the playbook must match.
	linesWithoutOk := []model.Line{
		{Original: "failed to deploy", Normalized: "failed to deploy"},
	}
	if got := Rank([]model.Playbook{pb}, linesWithoutOk, model.Context{}); len(got) != 1 {
		t.Fatalf("expected 1 result (no suppress), got %d", len(got))
	}
}

func TestRankRegexAllPattern(t *testing.T) {
	// match.all with a regex pattern must earn the compound bonus when every
	// pattern hits.
	pb := model.Playbook{
		ID:    "oom-killed",
		Title: "OOM Killed",
		Match: model.MatchSpec{
			All: []string{"re:killed|oom", "out of memory"},
		},
	}
	lines := []model.Line{
		{Original: "Process killed by OOM killer", Normalized: "process killed by oom killer"},
		{Original: "out of memory: kill process", Normalized: "out of memory: kill process"},
	}
	results := Rank([]model.Playbook{pb}, lines, model.Context{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// 2 all-hits × 1.5 + compound bonus 2.0 = 5.0
	if results[0].Score != 5.0 {
		t.Fatalf("expected score 5.0, got %v", results[0].Score)
	}
}

func TestPatternKey(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"error", "error"},
		{"  Error  ", "error"},
		{"connection refused", "connection refused"},
		{"re:exit code [0-9]+", "re:exit code [0-9]+"},
		{"re:  Exit Code [0-9]+  ", "re:exit code [0-9]+"},
		{"re:", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := patternKey(tc.input); got != tc.want {
			t.Errorf("patternKey(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// --- Word-boundary (containsPhrase) tests ---

func TestContainsPhraseWordBoundary(t *testing.T) {
	cases := []struct {
		text   string
		phrase string
		want   bool
	}{
		// Short word inside a longer token must NOT match (simple patterns check end boundary).
		{"errorcode: 0", "error", false},
		{"foobarbaz", "bar", false},
		// Single word at end/start of string IS a match.
		{"exit error", "error", true},
		{"error occurred", "error", true},
		// Simple word must not match when fused into a longer word (suffix).
		{"running concurrently", "concurrent", false},
		// Multi-word phrase checks start boundary only; end suffixes are intentional.
		{"connection refused here", "connection refused", true},
		{"preconnection refused", "connection refused", false},
		// Multi-word phrase must match plural forms (no end-boundary check for non-simple).
		{"import environment variables from a file", "environment variable", true},
		{"previously-included files matching", "included file", true},
		{"name: upload artifacts", "upload artifact", true},
		{"failed with exit code 128", "exit code 1", true},
		// Mixed pattern (has non-word chars) checks start boundary only.
		{"error[e0502]: cannot borrow", "error[e", true},
		// Pattern starting with non-word char: start boundary not checked.
		{"project has .nvmrc", ".nvmrc", true},
		{"project has.nvmrc checked", ".nvmrc", true}, // '.' starts pattern → no start-boundary check
		// Pattern ending with non-word char: end boundary not checked.
		{"error: something", "error:", true},
		{"error:something", "error:", true},
		// Pattern with non-word char at both edges: no boundary check at all.
		{".foo.", ".foo.", true},
	}
	for _, tc := range cases {
		if got := containsPhrase(tc.text, tc.phrase); got != tc.want {
			t.Errorf("containsPhrase(%q, %q) = %v, want %v", tc.text, tc.phrase, got, tc.want)
		}
	}
}

func TestRankWordBoundaryNoFalsePositive(t *testing.T) {
	pb := model.Playbook{
		ID:    "error-exact",
		Title: "Error exact",
		Match: model.MatchSpec{Any: []string{"error"}},
	}
	// "errorcode" must NOT trigger the playbook; "exit error" must.
	lines := []model.Line{
		{Original: "errorcode: 0", Normalized: "errorcode: 0"},
		{Original: "Exit error", Normalized: "exit error"},
	}
	results := Rank([]model.Playbook{pb}, lines, model.Context{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Evidence[0] != "Exit error" {
		t.Fatalf("expected evidence from 'Exit error', got %q", results[0].Evidence[0])
	}
}

// --- within_lines proximity tests ---

func TestRankWithinLinesCompoundBonusAwarded(t *testing.T) {
	pb := model.Playbook{
		ID:    "adjacent",
		Title: "Adjacent patterns",
		Match: model.MatchSpec{
			All:         []string{"connection refused", "retrying"},
			WithinLines: 5,
		},
	}
	lines := []model.Line{
		{Original: "connection refused", Normalized: "connection refused", Number: 1},
		{Original: "retrying request", Normalized: "retrying request", Number: 2},
	}
	results := Rank([]model.Playbook{pb}, lines, model.Context{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// 2 all-hits × 1.5 + compound bonus 2.0 = 5.0
	if results[0].Score != 5.0 {
		t.Fatalf("expected score 5.0 (compound bonus awarded), got %v", results[0].Score)
	}
}

func TestRankWithinLinesPreventsCompoundBonus(t *testing.T) {
	pb := model.Playbook{
		ID:    "far-apart",
		Title: "Far apart patterns",
		Match: model.MatchSpec{
			All:         []string{"connection refused", "retrying"},
			WithinLines: 3,
		},
	}
	// Patterns are 10 line-positions apart — outside the 3-line window.
	lines := make([]model.Line, 12)
	for i := range lines {
		lines[i] = model.Line{Original: "filler", Normalized: "filler", Number: i + 1}
	}
	lines[0] = model.Line{Original: "connection refused", Normalized: "connection refused", Number: 1}
	lines[10] = model.Line{Original: "retrying request", Normalized: "retrying request", Number: 11}

	results := Rank([]model.Playbook{pb}, lines, model.Context{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// 2 all-hits × 1.5 = 3.0; compound bonus NOT awarded (patterns too far apart)
	if results[0].Score != 3.0 {
		t.Fatalf("expected score 3.0 (no compound bonus), got %v", results[0].Score)
	}
}

func TestRankWithinLinesZeroDisablesProximityCheck(t *testing.T) {
	// WithinLines: 0 (default) must not apply any proximity gate.
	pb := model.Playbook{
		ID:    "no-proximity",
		Title: "No proximity",
		Match: model.MatchSpec{
			All: []string{"alpha", "omega"},
			// WithinLines not set (zero)
		},
	}
	lines := make([]model.Line, 102)
	for i := range lines {
		lines[i] = model.Line{Original: "noise", Normalized: "noise", Number: i + 1}
	}
	lines[0] = model.Line{Original: "alpha hit", Normalized: "alpha hit", Number: 1}
	lines[100] = model.Line{Original: "omega hit", Normalized: "omega hit", Number: 101}

	results := Rank([]model.Playbook{pb}, lines, model.Context{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// Compound bonus must be awarded regardless of distance.
	if results[0].Score != 5.0 {
		t.Fatalf("expected score 5.0 (compound bonus, no proximity gate), got %v", results[0].Score)
	}
}
