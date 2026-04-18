package output

import (
	"strings"
	"testing"

	"faultline/internal/model"
	"faultline/internal/renderer"
)

// Tests for differentialSummaryLines via FormatAnalysisMarkdown (ModeDetailed).

func TestFormatAnalysisMarkdownDifferentialSummaryLikelyCause(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, []string{"authentication required"})
	a.Differential = &model.DifferentialDiagnosis{
		Version: "hypothesis.v1",
		Likely: &model.DifferentialCandidate{
			FailureID:      "docker-auth",
			Title:          "Docker auth",
			ConfidenceText: "High",
			Why:            []string{"registry rejected credentials"},
			DisproofChecks: []string{"no successful push found"},
		},
	}
	out := FormatAnalysisMarkdown(a, 1, ModeDetailed)
	if !strings.Contains(out, "likely cause: docker-auth") {
		t.Errorf("expected likely cause in differential summary, got:\n%s", out)
	}
	if !strings.Contains(out, "confidence: High") {
		t.Errorf("expected confidence in differential summary, got:\n%s", out)
	}
	if !strings.Contains(out, "evidence: registry rejected credentials") {
		t.Errorf("expected evidence in differential summary, got:\n%s", out)
	}
	if !strings.Contains(out, "disproof check") {
		t.Errorf("expected disproof check in differential summary, got:\n%s", out)
	}
}

func TestFormatAnalysisMarkdownDifferentialSummaryAlternatives(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, []string{"authentication required"})
	a.Differential = &model.DifferentialDiagnosis{
		Version: "hypothesis.v1",
		Likely: &model.DifferentialCandidate{
			FailureID: "docker-auth",
			Title:     "Docker auth",
		},
		Alternatives: []model.DifferentialCandidate{
			{
				FailureID:     "image-pull",
				Title:         "Image pull failure",
				WhyLessLikely: []string{"missing pull access denied evidence"},
			},
		},
	}
	out := FormatAnalysisMarkdown(a, 1, ModeDetailed)
	if !strings.Contains(out, "alternative: image-pull") {
		t.Errorf("expected alternative in differential summary, got:\n%s", out)
	}
	if !strings.Contains(out, "why less likely: missing pull access denied evidence") {
		t.Errorf("expected why less likely reason, got:\n%s", out)
	}
}

func TestFormatAnalysisMarkdownDifferentialSummaryRuledOut(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, []string{"authentication required"})
	a.Differential = &model.DifferentialDiagnosis{
		Version: "hypothesis.v1",
		Likely: &model.DifferentialCandidate{
			FailureID: "docker-auth",
			Title:     "Docker auth",
		},
		RuledOut: []model.DifferentialCandidate{
			{
				FailureID:  "cache-corruption",
				Title:      "Cache corruption",
				RuledOutBy: []string{"checksum mismatch was absent"},
			},
		},
	}
	out := FormatAnalysisMarkdown(a, 1, ModeDetailed)
	if !strings.Contains(out, "ruled out: cache-corruption") {
		t.Errorf("expected ruled out in differential summary, got:\n%s", out)
	}
	if !strings.Contains(out, "reason: checksum mismatch was absent") {
		t.Errorf("expected rule-out reason, got:\n%s", out)
	}
}

func TestFormatAnalysisMarkdownDifferentialSummaryNilDiffFallsBack(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   model.Playbook{ID: "a", Title: "A", Summary: "A", Fix: "1. Fix A"},
				Confidence: 1.0,
				Score:      2.0,
			},
			{
				Playbook:   model.Playbook{ID: "b", Title: "B", Summary: "B"},
				Confidence: 0.5,
				Score:      1.0,
			},
		},
	}
	// No a.Differential - should use fallback differential lines
	out := FormatAnalysisMarkdown(a, 1, ModeDetailed)
	if !strings.Contains(out, "top candidate: a") {
		t.Errorf("expected fallback top candidate in differential, got:\n%s", out)
	}
}

// Tests for FormatAnalysisMarkdown detailed sections exercising further paths.

func TestFormatAnalysisMarkdownDetailedScoreBreakdown(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, []string{"authentication required"})
	a.Results[0].Breakdown = model.ScoreBreakdown{
		BaseSignalScore:     0.9,
		FinalScore:          1.1,
		CompoundSignalBonus: 0.2,
	}
	out := FormatAnalysisMarkdown(a, 1, ModeDetailed)
	if !strings.Contains(out, "Score Breakdown") {
		t.Errorf("expected Score Breakdown section, got:\n%s", out)
	}
}

func TestFormatAnalysisMarkdownDetailedExplanation(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, []string{"authentication required"})
	a.Results[0].Explanation = model.ResultExplanation{
		TriggeredBy: []string{"registry push denied"},
		AmplifiedBy: []string{"recent auth config change"},
		MitigatedBy: []string{"fallback registry configured"},
	}
	out := FormatAnalysisMarkdown(a, 1, ModeDetailed)
	for _, want := range []string{"Triggered By", "Amplified By", "Mitigated By"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in detailed markdown, got:\n%s", want, out)
		}
	}
}

func TestFormatAnalysisMarkdownNoMatchReturnsNoMatchHeader(t *testing.T) {
	out := FormatAnalysisMarkdown(nil, 1, ModeDetailed)
	if !strings.Contains(out, "# No Match") {
		t.Errorf("expected # No Match header, got:\n%s", out)
	}
}

// Tests for FormatAnalysisMarkdown with DeltaDiagnosis/RepoContext.

func TestFormatAnalysisMarkdownDetailedRepoContext(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, []string{"auth required"})
	a.RepoContext = &model.RepoContext{
		RepoRoot:           "/repo",
		RecentFiles:        []string{"Dockerfile"},
		HotspotDirectories: []string{"deploy"},
	}
	out := FormatAnalysisMarkdown(a, 1, ModeDetailed)
	if !strings.Contains(out, "Repo Context") {
		t.Errorf("expected Repo Context section, got:\n%s", out)
	}
}

// Tests for FormatFixMarkdown.

func TestFormatFixMarkdownContainsFix(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, []string{"auth required"})
	out := FormatFixMarkdown(a)
	if !strings.Contains(out, "## Fix") {
		t.Errorf("expected Fix section in fix markdown, got:\n%s", out)
	}
}

func TestFormatFixMarkdownNilIsNoMatch(t *testing.T) {
	out := FormatFixMarkdown(nil)
	if !strings.Contains(out, "No Match") {
		t.Errorf("expected No Match for nil analysis, got:\n%s", out)
	}
}

// Tests for FormatAnalysisText that exercise renderer paths.

func TestFormatAnalysisTextDetailedDifferential(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   model.Playbook{ID: "a", Title: "A", Summary: "A summary", Fix: "1. Fix A", Category: "build"},
				Confidence: 0.9,
				Score:      2.0,
				Evidence:   []string{"error A"},
				Explanation: model.ResultExplanation{
					TriggeredBy: []string{"error A"},
				},
				Ranking: &model.Ranking{
					BaselineScore: 2.0,
					FinalScore:    2.0,
					Contributions: []model.RankingContribution{
						{Feature: "detector_score", Contribution: 1.6, Reason: "anchor"},
					},
				},
				Breakdown: model.ScoreBreakdown{
					BaseSignalScore: 2.0,
					FinalScore:      2.0,
				},
			},
			{
				Playbook: model.Playbook{ID: "b", Title: "B", Summary: "B summary"},
				Score:    1.0,
				Evidence: []string{"error B"},
			},
		},
		Differential: &model.DifferentialDiagnosis{
			Version: "hypothesis.v1",
			Likely: &model.DifferentialCandidate{
				FailureID:      "a",
				Title:          "A",
				ConfidenceText: "High",
				Why:            []string{"error A evidence"},
			},
			Alternatives: []model.DifferentialCandidate{
				{FailureID: "b", Title: "B", WhyLessLikely: []string{"weaker match"}},
			},
		},
	}
	out := FormatAnalysisText(a, 1, ModeDetailed, renderer.Options{Plain: true, Width: 88})
	// In detailed mode, we get the playbook title header directly
	if !strings.Contains(out, "Differential Diagnosis") {
		t.Errorf("expected Differential Diagnosis in detailed output, got:\n%s", out)
	}
	if !strings.Contains(out, "Likely cause: a") {
		t.Errorf("expected likely cause in detailed differential, got:\n%s", out)
	}
}
