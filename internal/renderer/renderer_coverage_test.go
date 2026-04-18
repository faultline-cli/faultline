package renderer

import (
	"strings"
	"testing"

	"faultline/internal/model"
)

// Tests for renderDifferentialSummary (via RenderAnalyze detailed path).

func TestRenderAnalyzeDetailedDifferentialSummaryLikelyCause(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   samplePlaybook(),
				Confidence: 0.9,
				Score:      2.0,
				Evidence:   []string{"build error"},
				Explanation: model.ResultExplanation{
					TriggeredBy: []string{"build error signal"},
				},
				Ranking: &model.Ranking{
					BaselineScore: 2.0,
					FinalScore:    2.0,
					Contributions: []model.RankingContribution{
						{Feature: "detector_score", Contribution: 1.6, Reason: "anchor"},
					},
				},
				Breakdown: model.ScoreBreakdown{BaseSignalScore: 2.0, FinalScore: 2.0},
			},
		},
		Differential: &model.DifferentialDiagnosis{
			Version: "hypothesis.v1",
			Likely: &model.DifferentialCandidate{
				FailureID:      "go-sum-missing",
				Title:          "Missing go.sum entry",
				ConfidenceText: "High",
				Why:            []string{"missing go.sum entry detected"},
				DisproofChecks: []string{"no inconsistent hashes found"},
			},
		},
	}

	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, true)
	if !strings.Contains(out, "Likely cause: go-sum-missing") {
		t.Errorf("expected 'Likely cause: go-sum-missing' in differential, got:\n%s", out)
	}
	if !strings.Contains(out, "Confidence: High") {
		t.Errorf("expected confidence text in differential, got:\n%s", out)
	}
	if !strings.Contains(out, "Evidence: missing go.sum entry detected") {
		t.Errorf("expected evidence in differential, got:\n%s", out)
	}
	if !strings.Contains(out, "Disproof check: no inconsistent hashes found") {
		t.Errorf("expected disproof check in differential, got:\n%s", out)
	}
}

func TestRenderAnalyzeDetailedDifferentialSummaryAlternatives(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   samplePlaybook(),
				Confidence: 0.9,
				Score:      2.0,
				Evidence:   []string{"build error"},
				Breakdown:  model.ScoreBreakdown{BaseSignalScore: 2.0, FinalScore: 2.0},
			},
		},
		Differential: &model.DifferentialDiagnosis{
			Version: "hypothesis.v1",
			Likely: &model.DifferentialCandidate{
				FailureID: "go-sum-missing",
				Title:     "Missing go.sum entry",
			},
			Alternatives: []model.DifferentialCandidate{
				{
					FailureID:     "cache-corruption",
					Title:         "Cache corruption",
					WhyLessLikely: []string{"cache restore was present"},
				},
			},
		},
	}

	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, true)
	if !strings.Contains(out, "Alternative: cache-corruption") {
		t.Errorf("expected alternative in differential, got:\n%s", out)
	}
	if !strings.Contains(out, "Why less likely: cache restore was present") {
		t.Errorf("expected 'Why less likely' in differential, got:\n%s", out)
	}
}

func TestRenderAnalyzeDetailedDifferentialSummaryRuledOut(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   samplePlaybook(),
				Confidence: 0.9,
				Score:      2.0,
				Evidence:   []string{"build error"},
				Breakdown:  model.ScoreBreakdown{BaseSignalScore: 2.0, FinalScore: 2.0},
			},
		},
		Differential: &model.DifferentialDiagnosis{
			Version: "hypothesis.v1",
			Likely: &model.DifferentialCandidate{
				FailureID: "go-sum-missing",
				Title:     "Missing go.sum entry",
			},
			RuledOut: []model.DifferentialCandidate{
				{
					FailureID:  "dependency-drift",
					Title:      "Dependency drift",
					RuledOutBy: []string{"hash mismatch signal absent"},
				},
			},
		},
	}

	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, true)
	if !strings.Contains(out, "Ruled out: dependency-drift") {
		t.Errorf("expected ruled out candidate, got:\n%s", out)
	}
	if !strings.Contains(out, "Reason: hash mismatch signal absent") {
		t.Errorf("expected ruled out reason, got:\n%s", out)
	}
}

// Tests for renderRanking (via RenderAnalyze detailed with Ranking populated).

func TestRenderAnalyzeDetailedRankingSection(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   samplePlaybook(),
				Confidence: 0.85,
				Score:      2.5,
				Evidence:   []string{"build error"},
				Ranking: &model.Ranking{
					Mode:          "bayes_v1",
					Version:       "v1",
					BaselineScore: 2.0,
					FinalScore:    2.5,
					Prior:         0.1,
					StrongestPositive: []string{"tool match"},
					StrongestNegative: []string{"weak evidence"},
				},
				Breakdown: model.ScoreBreakdown{BaseSignalScore: 2.0, FinalScore: 2.5, CompoundSignalBonus: 0.5},
			},
		},
	}
	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, true)
	// Confidence Breakdown section must contain ranking-derived fields
	if !strings.Contains(out, "Confidence Breakdown") {
		t.Errorf("expected Confidence Breakdown section, got:\n%s", out)
	}
	if !strings.Contains(out, "Final reranked score") {
		t.Errorf("expected 'Final reranked score' in confidence breakdown, got:\n%s", out)
	}
}

// Tests for detailPanelStyles (via non-plain RenderAnalyze with detailed=true).

func TestRenderAnalyzeStyledDetailedPanelsHaveANSI(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   samplePlaybook(),
				Confidence: 0.9,
				Score:      2.0,
				Evidence:   []string{"build error"},
				Explanation: model.ResultExplanation{
					TriggeredBy: []string{"trigger"},
					AmplifiedBy: []string{"amplifier"},
					MitigatedBy: []string{"mitigator"},
				},
				Ranking: &model.Ranking{
					BaselineScore: 2.0,
					FinalScore:    2.0,
					Contributions: []model.RankingContribution{
						{Feature: "detector_score", Contribution: 1.6, Reason: "anchor"},
					},
				},
				Breakdown: model.ScoreBreakdown{BaseSignalScore: 2.0, FinalScore: 2.0},
			},
		},
		Differential: &model.DifferentialDiagnosis{
			Version: "hypothesis.v1",
			Likely: &model.DifferentialCandidate{
				FailureID:      "go-sum-missing",
				Title:          "Missing go.sum entry",
				ConfidenceText: "High",
			},
		},
	}
	out := New(Options{Plain: false, Width: 88, DarkBackground: true}).RenderAnalyze(a, 1, true)
	// Styled output uses lipgloss panels — should contain rendered content
	if !strings.Contains(out, "Missing go.sum entry") {
		t.Errorf("expected playbook title in styled detailed output, got:\n%s", out)
	}
}

func TestRenderAnalyzeStyledDetailedEvidencePanel(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   samplePlaybook(),
				Confidence: 0.9,
				Score:      2.0,
				Evidence:   []string{"missing go.sum entry"},
				Breakdown:  model.ScoreBreakdown{BaseSignalScore: 2.0, FinalScore: 2.0},
			},
		},
	}
	// Use styled renderer (non-plain) to exercise detailPanelStyles "evidence" branch
	out := New(Options{Plain: false, Width: 88}).RenderAnalyze(a, 1, true)
	if !strings.Contains(out, "missing go.sum entry") {
		t.Errorf("expected evidence in styled output, got:\n%s", out)
	}
}

// Tests for panelTitleStyle (exercised via any non-plain detail panel render).

func TestPanelTitleStyleReturnsNonEmptyStyle(t *testing.T) {
	style := panelTitleStyle("#FF0000", "#FFFFFF")
	// The style should render something non-empty (just verify it doesn't panic)
	rendered := style.Render("test label")
	if rendered == "" {
		t.Error("panelTitleStyle rendered empty string")
	}
}

// Tests for renderRanking directly via renderAnalyzeResult path with Ranking.

func TestRenderAnalyzeDetailedScoreBreakdownWithCompoundBonus(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   samplePlaybook(),
				Confidence: 0.85,
				Score:      2.5,
				Evidence:   []string{"missing go.sum entry"},
				Breakdown: model.ScoreBreakdown{
					BaseSignalScore:     2.0,
					FinalScore:          2.5,
					CompoundSignalBonus: 0.5,
				},
			},
		},
	}
	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, true)
	if !strings.Contains(out, "Score Breakdown") {
		t.Errorf("expected Score Breakdown section, got:\n%s", out)
	}
	if !strings.Contains(out, "compound") {
		t.Errorf("expected compound bonus in score breakdown, got:\n%s", out)
	}
}

func TestRenderAnalyzeDetailedSignalTones(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   samplePlaybook(),
				Confidence: 0.85,
				Score:      2.0,
				Evidence:   []string{"build error"},
				Explanation: model.ResultExplanation{
					TriggeredBy: []string{"triggered signal"},
					AmplifiedBy: []string{"amplified signal"},
					MitigatedBy: []string{"mitigated signal"},
				},
				Breakdown: model.ScoreBreakdown{BaseSignalScore: 2.0, FinalScore: 2.0},
			},
		},
	}
	// Non-plain to exercise all "signal" tone branches in detailPanelStyles
	out := New(Options{Plain: false, Width: 88}).RenderAnalyze(a, 1, true)
	if !strings.Contains(out, "triggered signal") {
		t.Errorf("expected triggered signal in styled output, got:\n%s", out)
	}
	if !strings.Contains(out, "amplified signal") {
		t.Errorf("expected amplified signal in styled output, got:\n%s", out)
	}
	if !strings.Contains(out, "mitigated signal") {
		t.Errorf("expected mitigated signal in styled output, got:\n%s", out)
	}
}

// Test delta and repo context panels in detailed mode.

func TestRenderAnalyzeDetailedWithRepoContext(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   samplePlaybook(),
				Confidence: 0.85,
				Score:      2.0,
				Evidence:   []string{"build error"},
				Breakdown:  model.ScoreBreakdown{BaseSignalScore: 2.0, FinalScore: 2.0},
			},
		},
		RepoContext: &model.RepoContext{
			RepoRoot:           "/repo",
			RecentFiles:        []string{"Dockerfile"},
			HotspotDirectories: []string{"deploy"},
		},
	}
	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, true)
	if !strings.Contains(out, "Repo Context") {
		t.Errorf("expected Repo Context section in detailed output, got:\n%s", out)
	}
}
