package renderer

import (
	"strings"
	"testing"

	"faultline/internal/model"
)

// ── shortHistorySignature ─────────────────────────────────────────────────────

func TestShortHistorySignatureEmpty(t *testing.T) {
	if got := shortHistorySignature(""); got != "" {
		t.Fatalf("expected empty for empty input, got %q", got)
	}
}

func TestShortHistorySignatureWhitespace(t *testing.T) {
	if got := shortHistorySignature("   "); got != "" {
		t.Fatalf("expected empty for whitespace input, got %q", got)
	}
}

func TestShortHistorySignatureShort(t *testing.T) {
	got := shortHistorySignature("abc123")
	if got != "abc123" {
		t.Fatalf("expected %q unchanged, got %q", "abc123", got)
	}
}

func TestShortHistorySignatureTruncatesAt12(t *testing.T) {
	long := "abcdef1234567890"
	got := shortHistorySignature(long)
	if got != "abcdef123456" {
		t.Fatalf("expected first 12 chars, got %q", got)
	}
}

func TestShortHistorySignatureExactly12(t *testing.T) {
	val := "abcdef123456"
	got := shortHistorySignature(val)
	if got != val {
		t.Fatalf("expected unchanged for exactly-12 input, got %q", got)
	}
}

// ── historySummaryLines ───────────────────────────────────────────────────────

func TestHistorySummaryLinesEmpty(t *testing.T) {
	lines := historySummaryLines(model.Result{})
	if len(lines) != 0 {
		t.Fatalf("expected empty lines for empty result, got %v", lines)
	}
}

func TestHistorySummaryLinesFirstOccurrence(t *testing.T) {
	result := model.Result{
		OccurrenceCount: 1,
		FirstSeenAt:     "2026-01-01T00:00:00Z",
	}
	lines := historySummaryLines(result)
	found := false
	for _, l := range lines {
		if strings.Contains(l, "first recorded occurrence") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'first recorded occurrence' in lines: %v", lines)
	}
}

func TestHistorySummaryLinesManyOccurrencesWithSpan(t *testing.T) {
	result := model.Result{
		OccurrenceCount: 5,
		FirstSeenAt:     "2026-01-01T00:00:00Z",
		LastSeenAt:      "2026-01-05T00:00:00Z",
	}
	lines := historySummaryLines(result)
	found := false
	for _, l := range lines {
		if strings.Contains(l, "seen 5 times") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'seen 5 times' in lines: %v", lines)
	}
}

func TestHistorySummaryLinesManyOccurrencesNoSpan(t *testing.T) {
	result := model.Result{
		OccurrenceCount: 3,
		FirstSeenAt:     "not-a-date",
		LastSeenAt:      "also-not-a-date",
	}
	lines := historySummaryLines(result)
	found := false
	for _, l := range lines {
		if strings.Contains(l, "seen 3 times in local history") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'seen 3 times in local history' in lines: %v", lines)
	}
}

func TestHistorySummaryLinesWithSignatureHash(t *testing.T) {
	result := model.Result{
		SignatureHash:   "abc123456789xyz",
		OccurrenceCount: 0,
	}
	lines := historySummaryLines(result)
	found := false
	for _, l := range lines {
		if strings.Contains(l, "history available for signature") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected signature summary line, got %v", lines)
	}
}

func TestHistorySummaryLinesFirstLastSeenAt(t *testing.T) {
	result := model.Result{
		OccurrenceCount: 1,
		FirstSeenAt:     "2026-03-01T10:00:00Z",
		LastSeenAt:      "2026-03-10T12:00:00Z",
	}
	lines := historySummaryLines(result)
	seenFirst, seenLast := false, false
	for _, l := range lines {
		if strings.HasPrefix(l, "first seen:") {
			seenFirst = true
		}
		if strings.HasPrefix(l, "last seen:") {
			seenLast = true
		}
	}
	if !seenFirst {
		t.Errorf("expected 'first seen' line, got %v", lines)
	}
	if !seenLast {
		t.Errorf("expected 'last seen' line, got %v", lines)
	}
}

// ── renderHistorySummary via RenderAnalyze ────────────────────────────────────

func TestRenderAnalyzeQuickHistorySummaryAppearsInOutput(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{{
			Playbook:        samplePlaybook(),
			Confidence:      0.80,
			Score:           2.0,
			Evidence:        []string{"build error"},
			OccurrenceCount: 3,
			FirstSeenAt:     "2026-01-01T00:00:00Z",
			LastSeenAt:      "2026-01-05T00:00:00Z",
			SignatureHash:   "abcdef0123456789",
		}},
	}
	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, false)
	if !strings.Contains(out, "seen 3 times") {
		t.Errorf("expected occurrence count in quick render, got:\n%s", out)
	}
}

func TestRenderAnalyzeDetailedHistorySummaryAppearsInOutput(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{{
			Playbook:        samplePlaybook(),
			Confidence:      0.80,
			Score:           2.0,
			Evidence:        []string{"build error"},
			Breakdown:       model.ScoreBreakdown{BaseSignalScore: 2.0, FinalScore: 2.0},
			OccurrenceCount: 2,
			FirstSeenAt:     "2026-02-01T00:00:00Z",
			LastSeenAt:      "2026-02-03T00:00:00Z",
			SignatureHash:   "abcdef0123456789",
		}},
	}
	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, true)
	if !strings.Contains(out, "seen 2 times") {
		t.Errorf("expected occurrence count in detailed render, got:\n%s", out)
	}
}

// ── higherRankedReason ────────────────────────────────────────────────────────

func TestHigherRankedReasonReturnsFirstNonMetadataDelta(t *testing.T) {
	top := model.Result{
		Ranking: &model.Ranking{
			Contributions: []model.RankingContribution{
				{Feature: "detector_score", Contribution: 1.6, Reason: "anchor"},
				{Feature: "tool_or_stack_match", Contribution: 0.8, Reason: "tool tokens align"},
			},
		},
	}
	runnerUp := model.Result{
		Ranking: &model.Ranking{
			Contributions: []model.RankingContribution{
				{Feature: "detector_score", Contribution: 1.6, Reason: "anchor"},
				{Feature: "tool_or_stack_match", Contribution: 0.2, Reason: "weak tool match"},
			},
		},
	}
	got := higherRankedReason(top, runnerUp)
	// delta for tool_or_stack_match = 0.8 - 0.2 = 0.6 > 0 → returns reason
	if !strings.Contains(got, "tool tokens align") {
		t.Errorf("expected reason %q in output, got %q", "tool tokens align", got)
	}
}

func TestHigherRankedReasonNilTopRankingReturnsEmpty(t *testing.T) {
	top := model.Result{Ranking: nil}
	runnerUp := model.Result{Ranking: nil}
	got := higherRankedReason(top, runnerUp)
	if got != "" {
		t.Errorf("expected empty for nil rankings, got %q", got)
	}
}

func TestHigherRankedReasonNilRunnerUpRanking(t *testing.T) {
	top := model.Result{
		Ranking: &model.Ranking{
			Contributions: []model.RankingContribution{
				{Feature: "evidence_density", Contribution: 1.2, Reason: "dense evidence"},
			},
		},
	}
	runnerUp := model.Result{Ranking: nil}
	got := higherRankedReason(top, runnerUp)
	// delta = 1.2 - 0 = 1.2 > 0 → returns reason
	if !strings.Contains(got, "dense evidence") {
		t.Errorf("expected reason from nil runner-up, got %q", got)
	}
}

func TestHigherRankedReasonMetadataOnlyReturnsEmpty(t *testing.T) {
	top := model.Result{
		Ranking: &model.Ranking{
			Contributions: []model.RankingContribution{
				{Feature: "historical_fixture_support", Contribution: 0.5, Reason: "historical support"},
				{Feature: "candidate_separation", Contribution: 0.3, Reason: "separation"},
			},
		},
	}
	runnerUp := model.Result{Ranking: nil}
	// All metadata features → no reason returned
	got := higherRankedReason(top, runnerUp)
	if got != "" {
		t.Errorf("expected empty for metadata-only contributions, got %q", got)
	}
}

func TestHigherRankedReasonZeroDeltaReturnsEmpty(t *testing.T) {
	top := model.Result{
		Ranking: &model.Ranking{
			Contributions: []model.RankingContribution{
				{Feature: "evidence_density", Contribution: 1.0, Reason: "equal"},
			},
		},
	}
	runnerUp := model.Result{
		Ranking: &model.Ranking{
			Contributions: []model.RankingContribution{
				{Feature: "evidence_density", Contribution: 1.0, Reason: "equal"},
			},
		},
	}
	got := higherRankedReason(top, runnerUp)
	if got != "" {
		t.Errorf("expected empty when delta is zero, got %q", got)
	}
}

func TestHigherRankedReasonBlankReasonFallsBackToFeature(t *testing.T) {
	top := model.Result{
		Ranking: &model.Ranking{
			Contributions: []model.RankingContribution{
				{Feature: "evidence_density", Contribution: 1.5, Reason: ""},
			},
		},
	}
	runnerUp := model.Result{Ranking: nil}
	got := higherRankedReason(top, runnerUp)
	if got != "evidence_density" {
		t.Errorf("expected feature name as fallback reason, got %q", got)
	}
}

// ── alternateReason ───────────────────────────────────────────────────────────

func TestAlternateReasonSharedEvidence(t *testing.T) {
	top := model.Result{Evidence: []string{"auth failure"}}
	runnerUp := model.Result{Evidence: []string{"auth failure"}}
	got := alternateReason(top, runnerUp)
	if !strings.Contains(got, "same failing evidence line") {
		t.Errorf("expected shared evidence reason, got %q", got)
	}
}

func TestAlternateReasonNoSharedButHasEvidence(t *testing.T) {
	top := model.Result{Evidence: []string{"auth failure"}}
	runnerUp := model.Result{Evidence: []string{"different error"}}
	got := alternateReason(top, runnerUp)
	if !strings.Contains(got, "explicit evidence") {
		t.Errorf("expected explicit evidence reason, got %q", got)
	}
}

func TestAlternateReasonNoEvidenceReturnsEmpty(t *testing.T) {
	top := model.Result{Evidence: []string{"auth failure"}}
	runnerUp := model.Result{Evidence: []string{}}
	got := alternateReason(top, runnerUp)
	if got != "" {
		t.Errorf("expected empty reason when runner-up has no evidence, got %q", got)
	}
}

// ── renderScoreBreakdown with discounts ───────────────────────────────────────

func TestRenderScoreBreakdownZeroFinalScoreReturnsEmpty(t *testing.T) {
	r := New(Options{Plain: true, Width: 88})
	got := r.renderScoreBreakdown(model.ScoreBreakdown{FinalScore: 0})
	if got != "" {
		t.Fatalf("expected empty for zero FinalScore, got %q", got)
	}
}

func TestRenderScoreBreakdownNoModifiersReturnsEmpty(t *testing.T) {
	r := New(Options{Plain: true, Width: 88})
	got := r.renderScoreBreakdown(model.ScoreBreakdown{
		BaseSignalScore: 2.0,
		FinalScore:      2.0,
	})
	// FinalScore non-zero but all bonuses/discounts are zero → no breakdown needed.
	if got != "" {
		t.Fatalf("expected empty when no modifiers, got %q", got)
	}
}

func TestRenderScoreBreakdownAllModifiers(t *testing.T) {
	r := New(Options{Plain: true, Width: 88})
	got := r.renderScoreBreakdown(model.ScoreBreakdown{
		BaseSignalScore:            2.0,
		FinalScore:                 4.5,
		CompoundSignalBonus:        0.5,
		BlastRadiusMultiplier:      0.3,
		HotPathMultiplier:          0.2,
		ChangeIntroducedBonus:      0.1,
		MitigatingEvidenceDiscount: 0.4,
		ExplicitExceptionDiscount:  0.1,
		SafeContextDiscount:        0.2,
	})
	for _, want := range []string{
		"base: 2.00",
		"final: 4.50",
		"compound:",
		"blast radius:",
		"hot path:",
		"change bonus:",
		"mitigations:",
		"suppressions:",
		"safe context:",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in score breakdown, got:\n%s", want, got)
		}
	}
}

// ── renderRepoContext ─────────────────────────────────────────────────────────

func TestRenderRepoContextNilReturnsEmpty(t *testing.T) {
	r := New(Options{Plain: true, Width: 88})
	if got := r.renderRepoContext(nil); got != "" {
		t.Fatalf("expected empty for nil repo context, got %q", got)
	}
}

func TestRenderRepoContextAllFields(t *testing.T) {
	r := New(Options{Plain: true, Width: 88})
	repo := &model.RepoContext{
		RepoRoot:           "/home/runner/work/repo",
		RecentFiles:        []string{"Dockerfile"},
		HotspotDirectories: []string{"deploy/"},
		CoChangeHints:      []string{"Makefile"},
		HotfixSignals:      []string{"hotfix-flag"},
		DriftSignals:       []string{"stale-lock"},
		ConfigDriftSignals: []string{"env-mismatch"},
		CIChangeSignals:    []string{"workflow-edit"},
		LargeCommitSignals: []string{"large-diff"},
		RelatedCommits: []model.RepoCommit{
			{Date: "2026-01-01", Hash: "abc1234", Subject: "fix: auth"},
		},
	}
	got := r.renderRepoContext(repo)
	for _, want := range []string{
		"Repo root: /home/runner/work/repo",
		"Recent file: Dockerfile",
		"Hotspot area: deploy/",
		"Co-change: Makefile",
		"Hotfix signal: hotfix-flag",
		"Drift hint: stale-lock",
		"Config drift: env-mismatch",
		"CI change: workflow-edit",
		"Large commit: large-diff",
		"Related commit: 2026-01-01 abc1234 fix: auth",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in repo context, got:\n%s", want, got)
		}
	}
}

// ── renderDifferential without Differential field ────────────────────────────

func TestRenderDifferentialFallbackWithTwoResults(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   samplePlaybook(),
				Confidence: 0.9,
				Score:      3.0,
				Evidence:   []string{"missing go.sum entry"},
				Ranking: &model.Ranking{
					Contributions: []model.RankingContribution{
						{Feature: "evidence_density", Contribution: 1.5, Reason: "dense evidence"},
					},
				},
			},
			{
				Playbook:   model.Playbook{ID: "runner-up", Title: "Runner-up"},
				Confidence: 0.5,
				Score:      1.5,
				Evidence:   []string{"missing go.sum entry"},
				Ranking: &model.Ranking{
					Contributions: []model.RankingContribution{
						{Feature: "evidence_density", Contribution: 0.5, Reason: "sparse evidence"},
					},
				},
			},
		},
		Differential: nil, // no structured differential → fallback path
	}
	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 2, true)
	if !strings.Contains(out, "runner-up") {
		t.Errorf("expected runner-up in differential fallback, got:\n%s", out)
	}
}

func TestRenderDifferentialTiedScore(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   samplePlaybook(),
				Confidence: 0.9,
				Score:      2.0,
				Evidence:   []string{"error A"},
			},
			{
				Playbook:   model.Playbook{ID: "tied-runner-up", Title: "Tied"},
				Confidence: 0.9,
				Score:      2.0,
				Evidence:   []string{"error B"},
			},
		},
	}
	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 2, true)
	if !strings.Contains(out, "tied on score") {
		t.Errorf("expected 'tied on score' in output for equal scores, got:\n%s", out)
	}
}
