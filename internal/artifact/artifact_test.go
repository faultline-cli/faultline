package artifact

import (
	"testing"

	"faultline/internal/model"
)

// ── statusForAnalysis ─────────────────────────────────────────────────────────

func TestStatusForAnalysisNil(t *testing.T) {
	got := statusForAnalysis(nil)
	if got != model.ArtifactStatusUnknown {
		t.Errorf("expected ArtifactStatusUnknown for nil analysis, got %q", got)
	}
}

func TestStatusForAnalysisNoResults(t *testing.T) {
	a := &model.Analysis{Results: nil}
	got := statusForAnalysis(a)
	if got != model.ArtifactStatusUnknown {
		t.Errorf("expected ArtifactStatusUnknown for empty results, got %q", got)
	}
}

func TestStatusForAnalysisWithResults(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{Playbook: model.Playbook{ID: "docker-auth"}, Score: 0.9},
		},
	}
	got := statusForAnalysis(a)
	if got != model.ArtifactStatusMatched {
		t.Errorf("expected ArtifactStatusMatched, got %q", got)
	}
}

// ── refinedConfidence ─────────────────────────────────────────────────────────

func TestRefinedConfidenceBaseOnly(t *testing.T) {
	a := &model.Analysis{}
	result := model.Result{Confidence: 0.75}
	got := refinedConfidence(a, result)
	if got != 0.75 {
		t.Errorf("expected 0.75, got %f", got)
	}
}

func TestRefinedConfidenceWithRepoContext(t *testing.T) {
	a := &model.Analysis{
		RepoContext: &model.RepoContext{
			RelatedCommits: []model.RepoCommit{{Hash: "abc123", Subject: "fix thing"}},
		},
	}
	result := model.Result{Confidence: 0.75}
	got := refinedConfidence(a, result)
	want := 0.75 + 0.04
	if got != want {
		t.Errorf("expected %f with repo context boost, got %f", want, got)
	}
}

func TestRefinedConfidenceWithDelta(t *testing.T) {
	a := &model.Analysis{
		Delta: &model.Delta{
			Causes: []model.DeltaCause{{Kind: "dep-bump", Score: 1.0}},
		},
	}
	result := model.Result{Confidence: 0.75}
	got := refinedConfidence(a, result)
	want := 0.75 + 0.03
	if got != want {
		t.Errorf("expected %f with delta boost, got %f", want, got)
	}
}

func TestRefinedConfidenceWithBothBoosts(t *testing.T) {
	a := &model.Analysis{
		RepoContext: &model.RepoContext{
			RelatedCommits: []model.RepoCommit{{Hash: "abc"}},
		},
		Delta: &model.Delta{
			Signals: []model.DeltaSignal{{ID: "dep-change"}},
		},
	}
	result := model.Result{Confidence: 0.75}
	got := refinedConfidence(a, result)
	// Both boosts applied: +0.04 (repo) + 0.03 (delta) = 0.82.
	// Use epsilon comparison for floating-point safety.
	const epsilon = 1e-9
	want := 0.75 + 0.04 + 0.03
	if got < want-epsilon || got > want+epsilon {
		t.Errorf("expected ~%f with both boosts, got %f", want, got)
	}
}

func TestRefinedConfidenceCapsAt1(t *testing.T) {
	a := &model.Analysis{
		RepoContext: &model.RepoContext{
			RelatedCommits: []model.RepoCommit{{Hash: "abc"}},
		},
		Delta: &model.Delta{
			Causes: []model.DeltaCause{{Kind: "x", Score: 1.0}},
		},
	}
	result := model.Result{Confidence: 0.99}
	got := refinedConfidence(a, result)
	if got > 1.0 {
		t.Errorf("expected confidence capped at 1.0, got %f", got)
	}
	if got != 1.0 {
		t.Errorf("expected 1.0 after cap, got %f", got)
	}
}

func TestRefinedConfidenceNilAnalysis(t *testing.T) {
	result := model.Result{Confidence: 0.5}
	got := refinedConfidence(nil, result)
	if got != 0.5 {
		t.Errorf("expected 0.5 for nil analysis, got %f", got)
	}
}

// ── Build ─────────────────────────────────────────────────────────────────────

func TestBuildNilAnalysis(t *testing.T) {
	got := Build(nil)
	if got != nil {
		t.Errorf("expected nil for nil analysis, got %v", got)
	}
}

func TestBuildNoResults(t *testing.T) {
	a := &model.Analysis{
		Source:      "test.log",
		Fingerprint: "fp-xyz",
	}
	got := Build(a)
	if got == nil {
		t.Fatal("expected non-nil artifact")
	}
	if got.Status != model.ArtifactStatusUnknown {
		t.Errorf("expected ArtifactStatusUnknown, got %q", got.Status)
	}
	if got.SchemaVersion != SchemaVersion {
		t.Errorf("expected schema version %q, got %q", SchemaVersion, got.SchemaVersion)
	}
	if got.MatchedPlaybook != nil {
		t.Errorf("expected nil MatchedPlaybook for no-result analysis, got %v", got.MatchedPlaybook)
	}
}

func TestBuildWithResult(t *testing.T) {
	a := &model.Analysis{
		Source:      "ci.log",
		Fingerprint: "fp-123",
		Results: []model.Result{
			{
				Playbook:   model.Playbook{ID: "docker-auth", Title: "Docker auth", Category: "auth"},
				Score:      0.9,
				Confidence: 0.85,
				Detector:   "log",
				Evidence:   []string{"authentication required"},
			},
		},
	}
	got := Build(a)
	if got == nil {
		t.Fatal("expected non-nil artifact")
	}
	if got.Status != model.ArtifactStatusMatched {
		t.Errorf("expected ArtifactStatusMatched, got %q", got.Status)
	}
	if got.MatchedPlaybook == nil {
		t.Fatal("expected non-nil MatchedPlaybook")
	}
	if got.MatchedPlaybook.ID != "docker-auth" {
		t.Errorf("expected ID 'docker-auth', got %q", got.MatchedPlaybook.ID)
	}
	if got.Fingerprint != "fp-123" {
		t.Errorf("expected fingerprint 'fp-123', got %q", got.Fingerprint)
	}
	if len(got.Evidence) == 0 {
		t.Error("expected non-empty Evidence")
	}
	if got.Confidence != 0.85 {
		t.Errorf("expected confidence 0.85, got %f", got.Confidence)
	}
}

func TestBuildFingerprint(t *testing.T) {
	a := &model.Analysis{
		Fingerprint: "  padded-fp  ",
	}
	got := Build(a)
	if got.Fingerprint != "padded-fp" {
		t.Errorf("expected trimmed fingerprint 'padded-fp', got %q", got.Fingerprint)
	}
}

// ── Sync ──────────────────────────────────────────────────────────────────────

func TestSyncNilAnalysis(t *testing.T) {
	got := Sync(nil)
	if got != nil {
		t.Errorf("expected nil for nil analysis, got %v", got)
	}
}

func TestSyncSetsStatusAndArtifact(t *testing.T) {
	a := &model.Analysis{
		Source:      "ci.log",
		Fingerprint: "fp-sync",
		Results: []model.Result{
			{
				Playbook:   model.Playbook{ID: "timeout", Title: "Timeout"},
				Score:      0.8,
				Confidence: 0.75,
			},
		},
	}
	got := Sync(a)
	if got == nil {
		t.Fatal("expected non-nil result from Sync")
	}
	if got.Status != model.ArtifactStatusMatched {
		t.Errorf("expected ArtifactStatusMatched after Sync, got %q", got.Status)
	}
	if got.Artifact == nil {
		t.Error("expected non-nil Artifact after Sync")
	}
	// original must not be mutated
	if a.Status != "" {
		t.Errorf("Sync must not mutate original analysis, but Status was set to %q", a.Status)
	}
}

func TestSyncDoesNotMutateOriginal(t *testing.T) {
	a := &model.Analysis{
		Source:         "ci.log",
		DominantSignals: []string{"sig1"},
	}
	synced := Sync(a)
	// Mutate the synced clone's DominantSignals to verify it's a copy.
	synced.DominantSignals = append(synced.DominantSignals, "sig2")
	if len(a.DominantSignals) != 1 {
		t.Errorf("Sync mutated original DominantSignals slice: %v", a.DominantSignals)
	}
}

func TestSyncWithEmptyResults(t *testing.T) {
	a := &model.Analysis{
		Source:      "ci.log",
		Fingerprint: "fp-empty",
	}
	got := Sync(a)
	if got.Status != model.ArtifactStatusUnknown {
		t.Errorf("expected ArtifactStatusUnknown for no results, got %q", got.Status)
	}
}
