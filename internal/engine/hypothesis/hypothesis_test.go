package hypothesis

import (
	"strings"
	"testing"

	"faultline/internal/model"
)

func TestBuildReordersCandidatesUsingHypothesisEvidence(t *testing.T) {
	results := []model.Result{
		{
			Playbook: model.Playbook{
				ID:    "cache-corruption",
				Title: "Cache corruption",
				Hypothesis: model.HypothesisSpec{
					Supports: []model.HypothesisSignal{
						{Signal: "dependency.cache.corrupt", Weight: 0.8},
					},
					Contradicts: []model.HypothesisSignal{
						{Signal: "cache.restore.absent", Weight: -0.4},
					},
				},
			},
			Score: 1.5,
		},
		{
			Playbook: model.Playbook{
				ID:    "dependency-drift",
				Title: "Dependency drift",
				Hypothesis: model.HypothesisSpec{
					Supports: []model.HypothesisSignal{
						{Signal: "dependency.resolution.conflict", Weight: 0.7},
					},
				},
			},
			Score: 1.5,
		},
	}
	lines := []model.Line{
		{Original: "checksum mismatch", Normalized: "checksum mismatch"},
		{Original: "failed to resolve dependencies", Normalized: "failed to resolve dependencies"},
	}

	got, diff := Build(Inputs{Results: results, Lines: lines})
	if got[0].Playbook.ID != "dependency-drift" {
		t.Fatalf("expected dependency-drift to be first, got %#v", got)
	}
	if diff == nil || len(diff.Alternatives) == 0 {
		t.Fatalf("expected differential summary, got %#v", diff)
	}
	if diff.Alternatives[0].FailureID != "cache-corruption" {
		t.Fatalf("expected cache-corruption as alternative, got %#v", diff.Alternatives[0])
	}
	if joined := strings.ToLower(strings.Join(diff.Alternatives[0].WhyLessLikely, " ")); !strings.Contains(joined, "cache restore") {
		t.Fatalf("expected cache restore explanation, got %#v", diff.Alternatives[0].WhyLessLikely)
	}
}

func TestBuildMarksExcludedCandidatesAsRuledOut(t *testing.T) {
	results := []model.Result{
		{
			Playbook: model.Playbook{
				ID:    "dependency-drift",
				Title: "Dependency drift",
				Hypothesis: model.HypothesisSpec{
					Supports: []model.HypothesisSignal{
						{Signal: "dependency.resolution.conflict", Weight: 0.7},
					},
					Excludes: []model.HypothesisSignal{
						{Signal: "dependency.hash.mismatch"},
					},
				},
			},
			Score: 1.5,
		},
		{
			Playbook: model.Playbook{
				ID:    "pip-hash-mismatch",
				Title: "Hash mismatch",
				Hypothesis: model.HypothesisSpec{
					Supports: []model.HypothesisSignal{
						{Signal: "dependency.hash.mismatch", Weight: 0.9},
					},
				},
			},
			Score: 1.4,
		},
	}
	lines := []model.Line{
		{
			Original:   "THESE PACKAGES DO NOT MATCH THE HASHES FROM THE REQUIREMENTS FILE",
			Normalized: "these packages do not match the hashes from the requirements file",
		},
		{Original: "failed to resolve dependencies", Normalized: "failed to resolve dependencies"},
	}

	got, diff := Build(Inputs{Results: results, Lines: lines})
	if got[0].Playbook.ID != "pip-hash-mismatch" {
		t.Fatalf("expected pip-hash-mismatch to be first, got %#v", got)
	}
	if diff == nil || len(diff.RuledOut) == 0 || diff.RuledOut[0].FailureID != "dependency-drift" {
		t.Fatalf("expected dependency-drift to be ruled out, got %#v", diff)
	}
}
