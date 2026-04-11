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
		t.Fatalf("expected alpha to rank first via confidence tie-break, got %s", results[0].Playbook.ID)
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
	if results[0].Confidence != 1.0 {
		t.Fatalf("expected confidence 1.0, got %v", results[0].Confidence)
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
	if results[0].Confidence != 1.0 {
		t.Fatalf("expected confidence 1.0, got %v", results[0].Confidence)
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
