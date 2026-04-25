package output

import (
	"encoding/json"
	"strings"
	"testing"

	"faultline/internal/model"
	"faultline/internal/signature"
	tracereport "faultline/internal/trace"
)

func makeTraceReport(matched bool) tracereport.Report {
	return tracereport.Report{
		Playbook: model.Playbook{
			ID:       "docker-auth",
			Title:    "Docker auth failure",
			Detector: "log",
		},
		Matched: matched,
		Source:  "stdin",
		Rules: []tracereport.Rule{
			{
				Group:   "any",
				Index:   0,
				Pattern: "authentication required",
				Status:  tracereport.StatusMatched,
				Matched: true,
				LineMatches: []tracereport.LineMatch{
					{Number: 3, Text: "authentication required for registry"},
				},
			},
			{
				Group:   "any",
				Index:   1,
				Pattern: "access denied",
				Status:  tracereport.StatusMissing,
				Matched: false,
			},
		},
		Why: []string{"registry rejected credentials"},
	}
}

// ── FormatTraceMarkdown ───────────────────────────────────────────────────────

func TestFormatTraceMarkdownMatchedContainsPlaybookID(t *testing.T) {
	report := makeTraceReport(true)
	out := FormatTraceMarkdown(report, false, false, false)
	if !strings.Contains(out, "docker-auth") {
		t.Errorf("expected playbook ID in markdown, got:\n%s", out)
	}
}

func TestFormatTraceMarkdownMatchedShowsOutcome(t *testing.T) {
	report := makeTraceReport(true)
	out := FormatTraceMarkdown(report, false, false, false)
	if !strings.Contains(out, "- Outcome: matched") {
		t.Errorf("expected '- Outcome: matched' in markdown, got:\n%s", out)
	}
	if strings.Contains(out, "not matched") {
		t.Errorf("did not expect 'not matched' in markdown, got:\n%s", out)
	}
}

func TestFormatTraceMarkdownNotMatchedShowsOutcome(t *testing.T) {
	report := makeTraceReport(false)
	out := FormatTraceMarkdown(report, false, false, false)
	if !strings.Contains(out, "not matched") {
		t.Errorf("expected 'not matched' outcome in markdown, got:\n%s", out)
	}
}

func TestFormatTraceMarkdownShowsRankedStatus(t *testing.T) {
	report := makeTraceReport(true)
	report.Rank = 2
	out := FormatTraceMarkdown(report, false, false, false)
	if !strings.Contains(out, "ranked #2") {
		t.Errorf("expected ranked #2 in markdown, got:\n%s", out)
	}
}

func TestFormatTraceMarkdownContainsRuleEvaluationSection(t *testing.T) {
	report := makeTraceReport(true)
	out := FormatTraceMarkdown(report, false, false, false)
	if !strings.Contains(out, "Rule Evaluation") {
		t.Errorf("expected Rule Evaluation section in markdown, got:\n%s", out)
	}
}

func TestFormatTraceMarkdownShowsWhySection(t *testing.T) {
	report := makeTraceReport(true)
	out := FormatTraceMarkdown(report, false, false, false)
	if !strings.Contains(out, "Why This Result") {
		t.Errorf("expected Why This Result section in markdown, got:\n%s", out)
	}
}

func TestFormatTraceMarkdownShowsScoreSection(t *testing.T) {
	report := makeTraceReport(true)
	report.Score = 1.5
	report.Confidence = 0.85
	report.Scoring = &model.ScoreBreakdown{
		BaseSignalScore: 1.2,
		FinalScore:      1.5,
	}
	out := FormatTraceMarkdown(report, false, true, false)
	if !strings.Contains(out, "Score") {
		t.Errorf("expected Score section in markdown, got:\n%s", out)
	}
}

func TestFormatTraceMarkdownShowsEvidenceSection(t *testing.T) {
	report := makeTraceReport(true)
	out := FormatTraceMarkdown(report, true, false, false)
	if !strings.Contains(out, "Raw Evidence") {
		t.Errorf("expected Raw Evidence section in markdown, got:\n%s", out)
	}
}

func TestFormatTraceMarkdownShowsCompetingSection(t *testing.T) {
	report := makeTraceReport(true)
	report.Competing = []tracereport.Candidate{
		{Status: "rejected", FailureID: "image-pull", Title: "Image pull failure", Reasons: []string{"missing evidence"}},
	}
	out := FormatTraceMarkdown(report, false, false, true)
	if !strings.Contains(out, "Competing Matches") {
		t.Errorf("expected Competing Matches section in markdown, got:\n%s", out)
	}
	if !strings.Contains(out, "image-pull") {
		t.Errorf("expected competitor ID in markdown, got:\n%s", out)
	}
}

func TestFormatTraceMarkdownShowsHooksSection(t *testing.T) {
	report := makeTraceReport(true)
	report.Hooks = &model.HookReport{
		Mode:            model.HookModeSafe,
		BaseConfidence:  0.84,
		ConfidenceDelta: 0.05,
		FinalConfidence: 0.89,
		Results: []model.HookResult{{
			ID:       "registry-config",
			Category: model.HookCategoryVerify,
			Status:   model.HookStatusExecuted,
			Passed:   boolPtr(true),
		}},
	}

	out := FormatTraceMarkdown(report, false, false, false)
	if !strings.Contains(out, "## Hooks") {
		t.Fatalf("expected hooks section in markdown, got:\n%s", out)
	}
	if !strings.Contains(out, "verify/registry-config: executed") {
		t.Fatalf("expected hook summary in markdown, got:\n%s", out)
	}
}

func TestFormatTraceMarkdownShowsSignatureSection(t *testing.T) {
	report := makeTraceReport(true)
	sig := signature.ForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Detector: "log",
		Evidence: []string{"authentication required"},
	})
	report.Signature = &sig

	out := FormatTraceMarkdown(report, false, false, false)
	if !strings.Contains(out, "## Signature") {
		t.Fatalf("expected signature section in markdown, got:\n%s", out)
	}
	if !strings.Contains(out, sig.Hash) {
		t.Fatalf("expected signature hash in markdown, got:\n%s", out)
	}
}

func TestFormatTraceTextShowsHistoryWindowAndSignature(t *testing.T) {
	report := makeTraceReport(true)
	sig := signature.ForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Detector: "log",
		Evidence: []string{"authentication required"},
	})
	report.Signature = &sig
	report.OccurrenceCount = 3
	report.FirstSeenAt = "2026-04-20T10:00:00Z"
	report.LastSeenAt = "2026-04-23T12:00:00Z"
	report.HookHistory = &model.HookHistorySummary{
		TotalCount:    3,
		ExecutedCount: 3,
		PassedCount:   2,
		FailedCount:   1,
		LastSeenAt:    "2026-04-23T12:00:00Z",
	}

	out := FormatTraceText(report, false, false, false)
	for _, want := range []string{
		"History",
		"History available for signature " + sig.Hash[:12],
		"Seen 3 times over 3d in local history",
		"Hook verification history: 3 run(s)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in trace output, got:\n%s", want, out)
		}
	}
}

func TestFormatTraceMarkdownScoreInHeader(t *testing.T) {
	report := makeTraceReport(true)
	report.Score = 2.5
	report.Confidence = 0.9
	out := FormatTraceMarkdown(report, false, false, false)
	if !strings.Contains(out, "- Score:") {
		t.Errorf("expected score information in markdown when score > 0, got:\n%s", out)
	}
}

// ── FormatTraceJSON ───────────────────────────────────────────────────────────

func TestFormatTraceJSONMatchedIsValid(t *testing.T) {
	report := makeTraceReport(true)
	out, err := FormatTraceJSON(report, false, false, false)
	if err != nil {
		t.Fatalf("FormatTraceJSON: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if m["matched"] != true {
		t.Errorf("expected matched=true, got %v", m["matched"])
	}
	if m["playbook_id"] != "docker-auth" {
		t.Errorf("expected playbook_id=docker-auth, got %v", m["playbook_id"])
	}
}

func TestFormatTraceJSONNotMatchedIsValid(t *testing.T) {
	report := makeTraceReport(false)
	out, err := FormatTraceJSON(report, false, false, false)
	if err != nil {
		t.Fatalf("FormatTraceJSON: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["matched"] != false {
		t.Errorf("expected matched=false, got %v", m["matched"])
	}
}

func TestFormatTraceJSONIncludesScoringWhenRequested(t *testing.T) {
	report := makeTraceReport(true)
	report.Scoring = &model.ScoreBreakdown{BaseSignalScore: 1.0, FinalScore: 1.2}
	out, err := FormatTraceJSON(report, false, true, false)
	if err != nil {
		t.Fatalf("FormatTraceJSON: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["scoring"] == nil {
		t.Error("expected scoring field when showScoring=true")
	}
}

func TestFormatTraceJSONOmitsScoringByDefault(t *testing.T) {
	report := makeTraceReport(true)
	report.Scoring = &model.ScoreBreakdown{BaseSignalScore: 1.0, FinalScore: 1.2}
	out, err := FormatTraceJSON(report, false, false, false)
	if err != nil {
		t.Fatalf("FormatTraceJSON: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["scoring"] != nil {
		t.Error("expected scoring field to be omitted when showScoring=false")
	}
}

func TestFormatTraceJSONIncludesCompetingWhenRequested(t *testing.T) {
	report := makeTraceReport(true)
	report.Competing = []tracereport.Candidate{
		{Status: "rejected", FailureID: "image-pull", Title: "Image pull"},
	}
	out, err := FormatTraceJSON(report, false, false, true)
	if err != nil {
		t.Fatalf("FormatTraceJSON: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	competing, ok := m["competing"].([]interface{})
	if !ok || len(competing) == 0 {
		t.Error("expected competing field with entries when showRejected=true")
	}
}

func TestFormatTraceJSONIncludesHooks(t *testing.T) {
	report := makeTraceReport(true)
	report.Hooks = &model.HookReport{
		Mode:            model.HookModeSafe,
		BaseConfidence:  0.84,
		ConfidenceDelta: 0.05,
		FinalConfidence: 0.89,
	}
	out, err := FormatTraceJSON(report, false, false, false)
	if err != nil {
		t.Fatalf("FormatTraceJSON: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["hooks"] == nil {
		t.Fatal("expected hooks field in trace json")
	}
}

func TestFormatTraceJSONIncludesSignatureWhenAvailable(t *testing.T) {
	report := makeTraceReport(true)
	sig := signature.ForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Detector: "log",
		Evidence: []string{"authentication required"},
	})
	report.Signature = &sig

	out, err := FormatTraceJSON(report, false, false, false)
	if err != nil {
		t.Fatalf("FormatTraceJSON: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	signatureValue, ok := m["signature"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected signature object, got %#v", m["signature"])
	}
	if signatureValue["hash"] != sig.Hash {
		t.Fatalf("expected signature hash %s, got %#v", sig.Hash, signatureValue["hash"])
	}
}

// ── renderTraceRulesMarkdown (via FormatTraceMarkdown) ────────────────────────

func TestRenderTraceRulesMarkdownShowsMatchStatus(t *testing.T) {
	report := makeTraceReport(true)
	out := FormatTraceMarkdown(report, false, false, false)
	if !strings.Contains(out, "MATCHED") {
		t.Errorf("expected MATCHED rule status in markdown rules, got:\n%s", out)
	}
	if !strings.Contains(out, "MISSING") {
		t.Errorf("expected MISSING rule status in markdown rules, got:\n%s", out)
	}
}

func TestRenderTraceRulesMarkdownShowsRuleWithNote(t *testing.T) {
	report := makeTraceReport(true)
	report.Rules[0].Note = "important note"
	out := FormatTraceMarkdown(report, false, false, false)
	if !strings.Contains(out, "important note") {
		t.Errorf("expected note in markdown rules, got:\n%s", out)
	}
}

func TestRenderTraceRulesMarkdownEmptyRules(t *testing.T) {
	report := makeTraceReport(false)
	report.Rules = nil
	out := FormatTraceMarkdown(report, false, false, false)
	if !strings.Contains(out, "# Faultline Trace") {
		t.Errorf("expected trace header even with no rules, got:\n%s", out)
	}
}

// ── renderTraceScoringMarkdown ────────────────────────────────────────────────

func TestRenderTraceScoringMarkdownWithRanking(t *testing.T) {
	report := makeTraceReport(true)
	report.Score = 1.5
	report.Confidence = 0.8
	report.Scoring = &model.ScoreBreakdown{
		BaseSignalScore: 1.0,
		FinalScore:      1.5,
	}
	report.Ranking = &model.Ranking{
		Mode:       "bayes_v1",
		FinalScore: 1.5,
	}
	out := FormatTraceMarkdown(report, false, true, false)
	if !strings.Contains(out, "Score") {
		t.Errorf("expected Score section with ranking, got:\n%s", out)
	}
}

// ── renderTraceEvidenceMarkdown ───────────────────────────────────────────────

func TestRenderTraceEvidenceMarkdownShowsEvidenceLines(t *testing.T) {
	report := makeTraceReport(true)
	out := FormatTraceMarkdown(report, true, false, false)
	if !strings.Contains(out, "authentication required") {
		t.Errorf("expected evidence text in markdown, got:\n%s", out)
	}
}

func TestRenderTraceEvidenceMarkdownNoEvidenceOmitsSection(t *testing.T) {
	report := makeTraceReport(true)
	// Remove line matches so there's no evidence
	for i := range report.Rules {
		report.Rules[i].LineMatches = nil
	}
	out := FormatTraceMarkdown(report, true, false, false)
	if strings.Contains(out, "Raw Evidence") {
		t.Errorf("expected Raw Evidence section to be omitted when no evidence, got:\n%s", out)
	}
}

func TestRenderTraceEvidenceMarkdownDeduplicatesLines(t *testing.T) {
	report := makeTraceReport(true)
	report.Rules = []tracereport.Rule{
		{
			Group:   "any",
			Index:   0,
			Pattern: "error",
			Status:  tracereport.StatusMatched,
			Matched: true,
			LineMatches: []tracereport.LineMatch{
				{Number: 1, Text: "same error"},
			},
		},
		{
			Group:   "any",
			Index:   1,
			Pattern: "error duplicate",
			Status:  tracereport.StatusMatched,
			Matched: true,
			LineMatches: []tracereport.LineMatch{
				{Number: 1, Text: "same error"},
			},
		},
	}
	out := FormatTraceMarkdown(report, true, false, false)
	// Locate the Raw Evidence block and check "same error" appears exactly once there.
	evidenceStart := strings.Index(out, "```text")
	evidenceEnd := strings.LastIndex(out, "```")
	if evidenceStart < 0 || evidenceEnd <= evidenceStart {
		t.Fatalf("expected ``` code block for raw evidence, got:\n%s", out)
	}
	evidenceBlock := out[evidenceStart:evidenceEnd]
	count := strings.Count(evidenceBlock, "same error")
	if count != 1 {
		t.Errorf("expected deduplicated evidence (1 occurrence in raw block), got %d in:\n%s", count, evidenceBlock)
	}
}

// ── renderTraceCompetingMarkdown ──────────────────────────────────────────────

func TestRenderTraceCompetingMarkdownIncludesReasons(t *testing.T) {
	report := makeTraceReport(true)
	report.Competing = []tracereport.Candidate{
		{
			Status:    "rejected",
			FailureID: "image-pull",
			Title:     "Image pull failure",
			Reasons:   []string{"missing pull access denied evidence"},
		},
	}
	out := FormatTraceMarkdown(report, false, false, true)
	if !strings.Contains(out, "missing pull access denied evidence") {
		t.Errorf("expected competitor reason in markdown, got:\n%s", out)
	}
}

func TestRenderTraceCompetingMarkdownNoCompetitorsOmitsSection(t *testing.T) {
	report := makeTraceReport(true)
	report.Competing = nil
	out := FormatTraceMarkdown(report, false, false, true)
	if strings.Contains(out, "Competing Matches") {
		t.Errorf("expected Competing Matches to be absent when no competitors, got:\n%s", out)
	}
}

// ── FormatTraceText (for completeness of edge cases) ─────────────────────────

func TestFormatTraceTextMatchedIsNonEmpty(t *testing.T) {
	report := makeTraceReport(true)
	out := FormatTraceText(report, false, false, false)
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty trace text output")
	}
	if !strings.Contains(out, "docker-auth") {
		t.Errorf("expected playbook ID in text output, got:\n%s", out)
	}
}

func TestFormatTraceTextWithScoreAndConfidence(t *testing.T) {
	report := makeTraceReport(true)
	report.Score = 2.1
	report.Confidence = 0.75
	out := FormatTraceText(report, false, false, false)
	if !strings.Contains(out, "Score:") {
		t.Errorf("expected Score line in text output, got:\n%s", out)
	}
}
