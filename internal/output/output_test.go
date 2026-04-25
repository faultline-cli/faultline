package output

import (
	"encoding/json"
	"strings"
	"testing"

	analysiscompare "faultline/internal/compare"
	"faultline/internal/model"
	"faultline/internal/renderer"
	tracereport "faultline/internal/trace"
	"faultline/internal/workflow"
)

func makeAnalysis(id, title, category string, confidence float64, evidence []string) *model.Analysis {
	return &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:        id,
					Title:     title,
					Category:  category,
					Summary:   "Summary for " + id,
					Diagnosis: "Diagnosis for " + id,
					Fix:       "1. Fix step 1\n2. Fix step 2",
				},
				Detector:   "log",
				Confidence: confidence,
				Score:      confidence,
				Evidence:   evidence,
				Explanation: model.ResultExplanation{
					TriggeredBy: []string{"primary trigger"},
				},
				Breakdown: model.ScoreBreakdown{
					BaseSignalScore: confidence,
					FinalScore:      confidence,
				},
			},
		},
	}
}

// ── JSON ─────────────────────────────────────────────────────────────────────

func TestFormatAnalysisJSONMatched(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 1.0, []string{"authentication required"})
	a.RepoContext = &model.RepoContext{
		RepoRoot:           "/repo",
		RecentFiles:        []string{"Dockerfile"},
		HotspotDirectories: []string{"deploy"},
	}
	a.Results[0].Hypothesis = &model.HypothesisAssessment{
		BaseScore:  1.0,
		FinalScore: 1.4,
		Why:        []string{"registry rejected credentials"},
	}
	a.Differential = &model.DifferentialDiagnosis{
		Version: "hypothesis.v1",
		Likely: &model.DifferentialCandidate{
			FailureID:      "docker-auth",
			Title:          "Docker auth",
			ConfidenceText: "High",
			Why:            []string{"registry rejected credentials"},
		},
	}
	data, err := FormatAnalysisJSON(a, 1)
	if err != nil {
		t.Fatalf("format json: %v", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(data)), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["matched"] != true {
		t.Errorf("expected matched=true, got %v", out["matched"])
	}
	results, ok := out["results"].([]interface{})
	if !ok || len(results) != 1 {
		t.Fatalf("expected results array with 1 element, got %v", out["results"])
	}
	r := results[0].(map[string]interface{})
	if r["failure_id"] != "docker-auth" {
		t.Errorf("expected failure_id docker-auth, got %v", r["failure_id"])
	}
	if r["detector"] != "log" {
		t.Errorf("expected detector log, got %v", r["detector"])
	}
	if _, ok := r["hypothesis"].(map[string]interface{}); !ok {
		t.Fatalf("expected hypothesis object, got %v", r["hypothesis"])
	}
	repoCtx, ok := out["repo_context"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected repo_context object, got %v", out["repo_context"])
	}
	if repoCtx["repo_root"] != "/repo" {
		t.Fatalf("expected repo_root in repo_context, got %v", repoCtx["repo_root"])
	}
	if _, ok := out["differential"].(map[string]interface{}); !ok {
		t.Fatalf("expected differential object, got %v", out["differential"])
	}
}

func TestFormatAnalysisJSONNoMatch(t *testing.T) {
	a := &model.Analysis{
		Status:            model.ArtifactStatusUnknown,
		Fingerprint:       "abc12345",
		CandidateClusters: []model.CandidateCluster{{Key: "build", Summary: "cluster"}},
		DominantSignals:   []string{"unknown signal"},
		SuggestedPlaybookSeed: &model.SuggestedPlaybookSeed{
			Category: "build",
			Title:    "Observed build failure signature",
			MatchAny: []string{"unknown signal"},
		},
		Artifact: &model.FailureArtifact{
			SchemaVersion: "failure_artifact.v1",
			Status:        model.ArtifactStatusUnknown,
			Fingerprint:   "abc12345",
			DominantSignals: []string{
				"unknown signal",
			},
		},
	}
	data, err := FormatAnalysisJSON(a, 1)
	if err != nil {
		t.Fatalf("format json no match: %v", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(data)), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["matched"] != false {
		t.Errorf("expected matched=false, got %v", out["matched"])
	}
	if out["status"] != "unknown" {
		t.Errorf("expected status=unknown, got %v", out["status"])
	}
	if _, ok := out["artifact"].(map[string]interface{}); !ok {
		t.Fatalf("expected artifact object for no-match, got %v", out["artifact"])
	}
	if _, ok := out["candidate_clusters"].([]interface{}); !ok {
		t.Fatalf("expected candidate_clusters for no-match, got %v", out["candidate_clusters"])
	}
	if out["message"] == "" {
		t.Error("expected a message field for no-match")
	}
}

func TestFormatAnalysisJSONEmptyResults(t *testing.T) {
	a := &model.Analysis{Results: []model.Result{}}
	data, err := FormatAnalysisJSON(a, 1)
	if err != nil {
		t.Fatalf("format json: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(data)), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["matched"] != false {
		t.Errorf("expected matched=false for empty results, got %v", out["matched"])
	}
}

func TestParseAnalysisJSONRoundTrip(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.67, []string{"authentication required"})
	a.Source = "stdin"
	a.Fingerprint = "abc123"
	a.InputHash = "input123"
	a.OutputHash = "output123"
	a.Context = model.Context{Stage: "deploy", CommandHint: "docker push"}
	a.RepoContext = &model.RepoContext{
		RepoRoot:           "/repo",
		RecentFiles:        []string{"Dockerfile"},
		HotspotDirectories: []string{"deploy"},
	}
	a.PackProvenances = []model.PackProvenance{
		{
			Name:          "premium",
			Version:       "1.2.3",
			SourceURL:     "https://example.com/premium.git",
			PinnedRef:     "abc1234",
			PlaybookCount: 7,
		},
	}
	tss := 0.75
	fpc := 0.60
	phi := 0.58
	a.Metrics = &model.Metrics{
		TSS:             &tss,
		FPC:             &fpc,
		PHI:             &phi,
		HistoryCount:    8,
		DriftComponents: []string{"recurring auth failures"},
	}
	a.Policy = &model.Policy{
		Recommendation: "quarantine",
		Reason:         "trace stability is degrading",
		Basis:          []string{"tss", "phi"},
	}
	a.Results[0].Hooks = &model.HookReport{
		Mode:            model.HookModeSafe,
		BaseConfidence:  0.67,
		ConfidenceDelta: 0.05,
		FinalConfidence: 0.72,
		Results: []model.HookResult{{
			ID:       "go-mod-present",
			Category: model.HookCategoryVerify,
			Kind:     model.HookKindFileExists,
			Status:   model.HookStatusExecuted,
		}},
	}
	a.Results[0].SignatureHash = "sig123"
	a.Results[0].SeenBefore = true
	a.Results[0].OccurrenceCount = 3
	a.Results[0].FirstSeenAt = "2026-04-20T10:00:00Z"
	a.Results[0].LastSeenAt = "2026-04-22T12:00:00Z"
	a.Results[0].HookHistorySummary = &model.HookHistorySummary{
		TotalCount:    2,
		ExecutedCount: 2,
		PassedCount:   1,
		FailedCount:   1,
		LastSeenAt:    "2026-04-22T12:00:00Z",
	}
	a.Artifact = &model.FailureArtifact{
		SchemaVersion: "failure_artifact.v1",
		Status:        model.ArtifactStatusMatched,
		Fingerprint:   "artifact123",
		MatchedPlaybook: &model.ArtifactPlaybook{
			ID: "docker-auth",
		},
		Evidence:   []string{"authentication required"},
		Confidence: 0.72,
		Environment: model.ArtifactEnvironment{
			Source: "stdin",
			Context: model.Context{
				Stage: "deploy",
			},
		},
		HistoryContext: &model.ArtifactHistoryContext{
			SignatureHash:   "sig123",
			OccurrenceCount: 3,
		},
		FixSteps: []string{"Fix the registry credentials."},
		Remediation: &model.RemediationPlan{
			Commands: []model.RemediationCommand{{
				ID:      "verify-1",
				Phase:   "verify",
				Command: []string{"docker", "login"},
			}},
		},
	}

	data, err := FormatAnalysisJSON(a, 1)
	if err != nil {
		t.Fatalf("FormatAnalysisJSON: %v", err)
	}
	parsed, err := ParseAnalysisJSON([]byte(data))
	if err != nil {
		t.Fatalf("ParseAnalysisJSON: %v", err)
	}
	if parsed.Source != a.Source || parsed.Fingerprint != a.Fingerprint || parsed.InputHash != a.InputHash || parsed.OutputHash != a.OutputHash {
		t.Fatalf("round-trip metadata mismatch: got %#v", parsed)
	}
	if len(parsed.Results) != 1 || parsed.Results[0].Playbook.ID != "docker-auth" {
		t.Fatalf("round-trip result mismatch: got %#v", parsed.Results)
	}
	if parsed.RepoContext == nil || parsed.RepoContext.RepoRoot != "/repo" {
		t.Fatalf("expected repo context to survive round trip, got %#v", parsed.RepoContext)
	}
	if len(parsed.PackProvenances) != 1 {
		t.Fatalf("expected pack provenance to survive round trip, got %#v", parsed.PackProvenances)
	}
	if parsed.PackProvenances[0].SourceURL != "https://example.com/premium.git" {
		t.Fatalf("expected pack source_url to survive round trip, got %#v", parsed.PackProvenances[0])
	}
	if parsed.Metrics == nil || parsed.Metrics.TSS == nil || *parsed.Metrics.TSS != tss {
		t.Fatalf("expected metrics to survive round trip, got %#v", parsed.Metrics)
	}
	if parsed.Policy == nil || parsed.Policy.Recommendation != "quarantine" {
		t.Fatalf("expected policy to survive round trip, got %#v", parsed.Policy)
	}
	if parsed.Results[0].Hooks == nil || parsed.Results[0].Hooks.FinalConfidence != 0.72 {
		t.Fatalf("expected hooks to survive round trip, got %#v", parsed.Results[0].Hooks)
	}
	if parsed.Results[0].SignatureHash != "sig123" || parsed.Results[0].OccurrenceCount != 3 || parsed.Results[0].HookHistorySummary == nil {
		t.Fatalf("expected history fields to survive round trip, got %#v", parsed.Results[0])
	}
	if parsed.Artifact == nil || parsed.Artifact.Fingerprint != "artifact123" {
		t.Fatalf("expected artifact to survive round trip, got %#v", parsed.Artifact)
	}
	if parsed.Artifact.Remediation == nil || len(parsed.Artifact.Remediation.Commands) != 1 {
		t.Fatalf("expected remediation to survive round trip, got %#v", parsed.Artifact)
	}
}

func TestFormatHookSummariesText(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.84, []string{"authentication required"})
	a.Results[0].Hooks = &model.HookReport{
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

	text := FormatHookSummariesText(a)
	if !strings.Contains(text, "docker-auth: mode: safe") {
		t.Fatalf("expected hook mode in text summary, got %q", text)
	}
	if !strings.Contains(text, "verify/registry-config: executed (passed)") {
		t.Fatalf("expected hook result in text summary, got %q", text)
	}
}

// ── Quick text ────────────────────────────────────────────────────────────────

func TestFormatAnalysisTextQuickSingleMatch(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 1.0, []string{"authentication required"})
	text := FormatAnalysisText(a, 1, ModeQuick, renderer.Options{Plain: true, Width: 88})
	for _, want := range []string{"Most Likely Diagnosis", "docker-auth", "Matched Evidence", "Recommended Action"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in quick output, got %q", want, text)
		}
	}
	if !strings.Contains(text, "Summary") {
		t.Errorf("expected Summary section in quick output, got %q", text)
	}
}

func TestFormatAnalysisTextQuickShowsConfidenceLabel(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.84, []string{"authentication required"})
	text := FormatAnalysisText(a, 1, ModeQuick, renderer.Options{Plain: true, Width: 88})
	if !strings.Contains(text, "Confidence: high (84%)") {
		t.Errorf("expected confidence label in quick output, got %q", text)
	}
}

func TestFormatAnalysisTextQuickIncludesHistorySummary(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.84, []string{"authentication required"})
	a.Results[0].SignatureHash = "0123456789abcdef0123456789abcdef"
	a.Results[0].OccurrenceCount = 3
	a.Results[0].FirstSeenAt = "2026-04-20T10:00:00Z"
	a.Results[0].LastSeenAt = "2026-04-23T12:00:00Z"
	a.Results[0].HookHistorySummary = &model.HookHistorySummary{
		TotalCount:    3,
		ExecutedCount: 3,
		PassedCount:   2,
		FailedCount:   1,
		LastSeenAt:    "2026-04-23T12:00:00Z",
	}

	text := FormatAnalysisText(a, 1, ModeQuick, renderer.Options{Plain: true, Width: 88})
	for _, want := range []string{"History", "history available for signature 0123456789ab", "seen 3 times over 3d in local history", "hook verification history: 3 run(s)"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in quick output, got %q", want, text)
		}
	}
}

func TestFormatAnalysisTextQuickTopN(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{Playbook: model.Playbook{ID: "a", Title: "A", Summary: "A", Fix: "1. Fix A"}, Confidence: 1.0, Score: 2},
			{Playbook: model.Playbook{ID: "b", Title: "B", Summary: "B", Fix: "1. Fix B"}, Confidence: 0.5, Score: 1},
			{Playbook: model.Playbook{ID: "c", Title: "C", Summary: "C", Fix: "1. Fix C"}, Confidence: 0.3, Score: 0.5},
		},
	}
	text := FormatAnalysisText(a, 2, ModeQuick, renderer.Options{Plain: true, Width: 88})
	if !strings.Contains(text, "Other Likely Matches") {
		t.Errorf("expected alternatives section, got %q", text)
	}
	if !strings.Contains(text, "#2 b (50%)") {
		t.Errorf("expected #2 candidate in quick output, got %q", text)
	}
	if strings.Contains(text, "#3") {
		t.Errorf("should not include #3 (top=2), got %q", text)
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func TestFormatAnalysisMarkdownDetailed(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.91, []string{"authentication required"})
	a.Context = model.Context{Stage: "deploy"}
	a.Results[0].OccurrenceCount = 2
	a.Results[0].FirstSeenAt = "2026-04-20T10:00:00Z"
	a.Results[0].LastSeenAt = "2026-04-23T12:00:00Z"
	a.Results[0].Explanation = model.ResultExplanation{
		TriggeredBy: []string{"registry rejected credentials"},
	}
	a.Results[0].Ranking = &model.Ranking{
		BaselineScore: 2.0,
		FinalScore:    2.4,
		Prior:         0.1,
		Contributions: []model.RankingContribution{
			{Feature: "detector_score", Contribution: 1.6, Reason: "baseline detector score remains the anchor"},
			{Feature: "tool_or_stack_match", Contribution: 0.2, Reason: "tool or stack tokens align with the evidence"},
		},
	}
	a.Results[0].Breakdown = model.ScoreBreakdown{
		BaseSignalScore:     0.91,
		FinalScore:          1.11,
		CompoundSignalBonus: 0.20,
	}
	a.Results = append(a.Results, model.Result{
		Playbook: model.Playbook{
			ID:      "image-pull-backoff",
			Title:   "Image pull backoff",
			Summary: "Alternative summary",
		},
		Score:    1.95,
		Evidence: []string{"authentication required"},
		Ranking: &model.Ranking{
			BaselineScore: 2.0,
			FinalScore:    1.95,
			Contributions: []model.RankingContribution{
				{Feature: "detector_score", Contribution: 1.6, Reason: "baseline detector score remains the anchor"},
			},
		},
	})
	a.RepoContext = &model.RepoContext{RepoRoot: "/repo", RecentFiles: []string{"Dockerfile"}}

	text := FormatAnalysisMarkdown(a, 1, ModeDetailed)
	for _, want := range []string{"# Docker auth", "- ID: `docker-auth`", "## Summary", "## History", "## Evidence", "## Differential Diagnosis", "## Confidence Breakdown", "## Triggered By", "## Score Breakdown", "## Suggested Fix", "## Repo Context"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in markdown output, got:\n%s", want, text)
		}
	}
}

func TestFormatTraceTextIncludesRulesAndCompeting(t *testing.T) {
	report := tracereport.Report{
		Source: "stdin",
		Playbook: model.Playbook{
			ID:    "missing-executable",
			Title: "Required executable or runtime binary missing",
		},
		Matched:         true,
		Rank:            1,
		Score:           2.10,
		Confidence:      0.84,
		OccurrenceCount: 2,
		FirstSeenAt:     "2026-04-20T10:00:00Z",
		LastSeenAt:      "2026-04-23T12:00:00Z",
		Detector:        "log",
		Rules: []tracereport.Rule{
			{
				Group:       "match.any",
				Index:       0,
				Pattern:     "no such file or directory",
				Status:      tracereport.StatusMatched,
				LineMatches: []tracereport.LineMatch{{Number: 12, Text: "exec /__e/node20/bin/node: no such file or directory"}},
			},
		},
		Why: []string{"1 trigger rule matched explicit log evidence"},
		Competing: []tracereport.Candidate{
			{Status: "alternative", FailureID: "runtime-mismatch", Title: "Runtime mismatch", Reasons: []string{"version conflict wording was absent"}},
		},
	}

	text := FormatTraceText(report, true, true, true)
	for _, want := range []string{"TRACE  missing-executable", "Rule Evaluation", "MATCHED", "History", "line 12", "Why This Result", "Competing Matches"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in trace text, got:\n%s", want, text)
		}
	}
}

func TestFormatTraceJSONIncludesRules(t *testing.T) {
	report := tracereport.Report{
		Source: "stdin",
		Playbook: model.Playbook{
			ID:    "docker-auth",
			Title: "Docker auth",
		},
		Matched:         true,
		OccurrenceCount: 2,
		FirstSeenAt:     "2026-04-20T10:00:00Z",
		LastSeenAt:      "2026-04-23T12:00:00Z",
		Rules: []tracereport.Rule{
			{Group: "match.any", Index: 0, Pattern: "authentication required", Status: tracereport.StatusMatched},
		},
	}

	data, err := FormatTraceJSON(report, false, false, false)
	if err != nil {
		t.Fatalf("FormatTraceJSON: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(data)), &payload); err != nil {
		t.Fatalf("unmarshal trace JSON: %v", err)
	}
	if payload["playbook_id"] != "docker-auth" {
		t.Fatalf("expected playbook_id docker-auth, got %v", payload["playbook_id"])
	}
	if payload["history"] == nil {
		t.Fatalf("expected history object in trace JSON, got %v", payload["history"])
	}
	rules, ok := payload["rules"].([]interface{})
	if !ok || len(rules) != 1 {
		t.Fatalf("expected one rule in trace JSON, got %v", payload["rules"])
	}
}

func TestFormatCompareMarkdownIncludesDiagnosisAndEvidence(t *testing.T) {
	report := analysiscompare.Report{
		LeftSource:       "previous.json",
		RightSource:      "current.json",
		Changed:          true,
		DiagnosisChanged: true,
		Previous:         &analysiscompare.Candidate{FailureID: "docker-auth", Title: "Docker auth", Confidence: 0.67},
		Current:          &analysiscompare.Candidate{FailureID: "permission-denied", Title: "Permission denied", Confidence: 0.33},
		Summary:          []string{"top diagnosis changed from docker-auth to permission-denied"},
		Evidence:         analysiscompare.StringDelta{Added: []string{"permission denied"}, Removed: []string{"authentication required"}},
	}

	text := FormatCompareMarkdown(report)
	for _, want := range []string{"# Faultline Compare", "## Diagnosis", "`docker-auth`", "`permission-denied`", "## Evidence Changes"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in compare markdown, got:\n%s", want, text)
		}
	}
}

func TestFormatCompareJSON(t *testing.T) {
	report := analysiscompare.Report{
		Changed:          true,
		DiagnosisChanged: true,
		Summary:          []string{"top diagnosis changed"},
	}
	data, err := FormatCompareJSON(report)
	if err != nil {
		t.Fatalf("FormatCompareJSON: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(data)), &payload); err != nil {
		t.Fatalf("unmarshal compare JSON: %v", err)
	}
	if payload["changed"] != true {
		t.Fatalf("expected changed=true, got %v", payload["changed"])
	}
}

func TestFormatAnalysisEvidenceText(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.67, []string{"authentication required"})
	a.Source = "stdin"
	text := FormatAnalysisEvidenceText(a)
	for _, want := range []string{"EVIDENCE  docker-auth", "Source: stdin", "Matched evidence:", "authentication required"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in evidence text, got:\n%s", want, text)
		}
	}
}

func TestFormatAnalysisEvidenceMarkdown(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.67, []string{"authentication required"})
	a.Source = "stdin"
	text := FormatAnalysisEvidenceMarkdown(a)
	for _, want := range []string{"# Faultline Evidence", "- ID: `docker-auth`", "## Matched Evidence", "```text"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in evidence markdown, got:\n%s", want, text)
		}
	}
}

// ── Detailed text ─────────────────────────────────────────────────────────────

func TestFormatAnalysisTextDetailed(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker registry auth failure", "auth", 1.0,
		[]string{"pull access denied"})
	a.Context = model.Context{Stage: "deploy", CommandHint: "docker push"}
	a.Results[0].Playbook.Summary = "Service failed readiness checks."
	a.Results[0].Ranking = &model.Ranking{
		BaselineScore: 2.0,
		FinalScore:    2.4,
		Prior:         0.1,
		Contributions: []model.RankingContribution{
			{Feature: "detector_score", Contribution: 1.6, Reason: "baseline detector score remains the anchor"},
			{Feature: "tool_or_stack_match", Contribution: 0.2, Reason: "tool or stack tokens align with the evidence"},
		},
	}
	a.Results = append(a.Results, model.Result{
		Playbook: model.Playbook{ID: "image-pull-backoff", Title: "Image pull backoff"},
		Score:    1.8,
		Evidence: []string{"pull access denied"},
		Ranking: &model.Ranking{
			BaselineScore: 2.0,
			FinalScore:    1.8,
			Contributions: []model.RankingContribution{
				{Feature: "detector_score", Contribution: 1.6, Reason: "baseline detector score remains the anchor"},
			},
		},
	})
	a.RepoContext = &model.RepoContext{
		RepoRoot:           "/repo",
		RecentFiles:        []string{"Dockerfile"},
		RelatedCommits:     []model.RepoCommit{{Hash: "abc1234", Date: "2026-04-10", Subject: "hotfix: adjust docker login"}},
		HotspotDirectories: []string{"deploy"},
		CoChangeHints:      []string{"Dockerfile <-> .github/workflows/deploy.yml"},
		HotfixSignals:      []string{"hotfix: adjust docker login"},
		DriftSignals:       []string{"Repeated edits in deploy"},
	}
	text := FormatAnalysisText(a, 1, ModeDetailed, renderer.Options{Plain: true, Width: 88})

	checks := []string{"Summary", "Category:", "Stage:", "Evidence", "Differential Diagnosis", "Confidence Breakdown", "Triggered by", "Suggested Fix", "Repo Context", "Related commit:", "Hotfix signal:"}
	for _, want := range checks {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in detailed output, got:\n%s", want, text)
		}
	}
}

// ── No match ─────────────────────────────────────────────────────────────────

func TestFormatAnalysisTextNilAnalysis(t *testing.T) {
	text := FormatAnalysisText(nil, 1, ModeQuick, renderer.Options{Plain: true, Width: 88})
	if !strings.Contains(text, "No known playbook matched this input.") {
		t.Errorf("expected no-match message, got %q", text)
	}
}

func TestFormatAnalysisMarkdownNilAnalysis(t *testing.T) {
	text := FormatAnalysisMarkdown(nil, 1, ModeQuick)
	if !strings.Contains(text, "# No Match") {
		t.Fatalf("expected markdown no-match heading, got %q", text)
	}
}

// ── CI annotations ────────────────────────────────────────────────────────────

func TestFormatCIAnnotations(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker Auth", "auth", 1.0, nil)
	a.Results[0].Playbook.Fix = "1. docker login"
	out := FormatCIAnnotations(a, 1)
	if !strings.Contains(out, "::warning") {
		t.Errorf("expected ::warning annotation, got %q", out)
	}
	if !strings.Contains(out, "docker-auth") {
		t.Errorf("expected playbook ID in annotation, got %q", out)
	}
}

// ── Playbook list & details ──────────────────────────────────────────────────

func TestFormatPlaybookList(t *testing.T) {
	pbs := []model.Playbook{
		{ID: "docker-auth", Category: "auth", Severity: "high", Title: "Docker Auth"},
		{ID: "aws-credentials", Category: "auth", Severity: "high", Title: "AWS Credentials", Metadata: model.PlaybookMeta{PackName: "team-pack"}},
		{ID: "oom-killed", Category: "runtime", Severity: "critical", Title: "OOM Killed"},
	}
	text := FormatPlaybookList(pbs, "", renderer.Options{Plain: true, Width: 100})
	if !strings.Contains(text, "docker-auth") || !strings.Contains(text, "oom-killed") || !strings.Contains(text, "team-pack") {
		t.Errorf("expected both playbooks in list, got %q", text)
	}
}

func TestFormatPlaybookListCategoryFilter(t *testing.T) {
	pbs := []model.Playbook{
		{ID: "docker-auth", Category: "auth", Title: "Docker Auth"},
		{ID: "oom-killed", Category: "runtime", Title: "OOM Killed"},
	}
	text := FormatPlaybookList(pbs, "auth", renderer.Options{Plain: true, Width: 100})
	if !strings.Contains(text, "docker-auth") {
		t.Errorf("expected docker-auth in filtered list, got %q", text)
	}
	if strings.Contains(text, "oom-killed") {
		t.Errorf("oom-killed should be filtered out, got %q", text)
	}
}

func TestFormatPlaybookDetails(t *testing.T) {
	pb := model.Playbook{
		ID:           "docker-auth",
		Title:        "Docker Registry Auth",
		Category:     "auth",
		Severity:     "high",
		Metadata:     model.PlaybookMeta{PackName: "team-pack"},
		Summary:      "The CI job could not authenticate.",
		Diagnosis:    "The CI job could not authenticate.",
		WhyItMatters: "Token expired.",
		Fix:          "1. Run docker login",
		Validation:   "- Retry the image pull",
		Match:        model.MatchSpec{Any: []string{"pull access denied"}},
	}
	text := FormatPlaybookDetails(pb, renderer.Options{Plain: true, Width: 88})
	for _, want := range []string{"docker-auth", "Docker Registry Auth", "auth", "high", "team-pack", "Token expired", "Run docker login"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in details, got:\n%s", want, text)
		}
	}
}

func TestFormatPlaybookDetailsMarkdown(t *testing.T) {
	pb := model.Playbook{
		ID:           "docker-auth",
		Title:        "Docker Registry Auth",
		Category:     "auth",
		Severity:     "high",
		Summary:      "The CI job could not authenticate.",
		Diagnosis:    "The registry credentials were rejected.",
		Fix:          "1. Run docker login",
		Validation:   "- Retry the image pull",
		WhyItMatters: "Builds cannot fetch images.",
		Match:        model.MatchSpec{Any: []string{"pull access denied"}},
	}
	text := FormatPlaybookDetailsMarkdown(pb)
	for _, want := range []string{"# Docker Registry Auth", "- ID: `docker-auth`", "## Diagnosis", "## Fix", "## Match Rules", "### match.any"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in markdown details, got:\n%s", want, text)
		}
	}
}

func TestFormatAnalysisTextShowsPack(t *testing.T) {
	a := makeAnalysis("aws-credentials", "AWS credentials missing or invalid", "auth", 1.0, nil)
	a.Results[0].Playbook.Metadata.PackName = "team-pack"
	out := FormatAnalysisText(a, 1, ModeQuick, renderer.Options{Plain: true, Width: 88})
	if !strings.Contains(out, "Pack: team-pack") {
		t.Fatalf("expected pack line in quick output, got %q", out)
	}
}

func TestFormatAnalysisJSONIncludesPack(t *testing.T) {
	a := makeAnalysis("aws-credentials", "AWS credentials missing or invalid", "auth", 1.0, nil)
	a.Results[0].Playbook.Metadata.PackName = "team-pack"
	text, err := FormatAnalysisJSON(a, 1)
	if err != nil {
		t.Fatalf("format analysis json: %v", err)
	}
	if !strings.Contains(text, "\"pack\":\"team-pack\"") {
		t.Fatalf("expected pack in json output, got %q", text)
	}
}

func TestFormatFixMarkdown(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker Auth", "auth", 1.0, nil)
	text := FormatFixMarkdown(a)
	for _, want := range []string{"# Docker Auth", "## Fix", "1. Fix step 1"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in markdown fix output, got %q", want, text)
		}
	}
}

func TestFormatWorkflowText(t *testing.T) {
	plan := workflow.Plan{
		SchemaVersion: "workflow.v1",
		Mode:          workflow.ModeAgent,
		FailureID:     "docker-build-context",
		Title:         "Docker build context or Dockerfile path issue",
		Source:        "build.log",
		Context: model.Context{
			Stage:       "build",
			CommandHint: "docker build -f Dockerfile .",
		},
		Evidence:   []string{"failed to read Dockerfile"},
		Files:      []string{"Dockerfile", ".dockerignore"},
		LocalRepro: []string{"docker build -f Dockerfile ."},
		Verify:     []string{"docker build -f Dockerfile ."},
		MetricsHints: []string{
			"TSS 0.40 (5 runs)",
		},
		PolicyHints: []string{
			"policy: quarantine",
		},
		Steps:       []string{"Verify the exact `docker build` command."},
		AgentPrompt: "You are helping resolve a deterministic CI failure.",
	}

	text := FormatWorkflowText(plan)
	for _, want := range []string{"WORKFLOW", "docker-build-context", "Local repro:", "Verify:", "Metrics:", "Policy:", "Next steps:", "Agent prompt:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in workflow text, got:\n%s", want, text)
		}
	}
}

// ── ParseFormat / Valid ───────────────────────────────────────────────────────

// ── View ──────────────────────────────────────────────────────────────────────

func TestParseViewKnownValues(t *testing.T) {
	tests := []struct {
		input string
		want  View
	}{
		{"", ViewDefault},
		{"summary", ViewSummary},
		{"Summary", ViewSummary},
		{"SUMMARY", ViewSummary},
		{"  summary  ", ViewSummary},
		{"evidence", ViewEvidence},
		{"fix", ViewFix},
		{"raw", ViewRaw},
		{"trace", ViewTrace},
	}
	for _, tt := range tests {
		got, ok := ParseView(tt.input)
		if !ok {
			t.Errorf("ParseView(%q) ok=false, want true", tt.input)
		}
		if got != tt.want {
			t.Errorf("ParseView(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseViewUnknown(t *testing.T) {
	for _, bad := range []string{"detail", "html", "full", "short"} {
		_, ok := ParseView(bad)
		if ok {
			t.Errorf("ParseView(%q) ok=true, want false", bad)
		}
	}
}

func TestViewValid(t *testing.T) {
	for _, v := range []View{ViewDefault, ViewSummary, ViewEvidence, ViewFix, ViewRaw, ViewTrace} {
		if !v.Valid() {
			t.Errorf("View(%q).Valid() = false, want true", v)
		}
	}
}

func TestViewInvalidNotValid(t *testing.T) {
	for _, bad := range []View{"detail", "html", "plain"} {
		if bad.Valid() {
			t.Errorf("View(%q).Valid() = true, want false", bad)
		}
	}
}

// ── FormatCompareText ─────────────────────────────────────────────────────────

func TestFormatCompareTextIncludesDiagnosis(t *testing.T) {
	report := analysiscompare.Report{
		LeftSource:       "prev.json",
		RightSource:      "curr.json",
		Changed:          true,
		DiagnosisChanged: true,
		Previous:         &analysiscompare.Candidate{FailureID: "docker-auth", Title: "Docker auth", Confidence: 0.9},
		Current:          &analysiscompare.Candidate{FailureID: "permission-denied", Title: "Permission denied", Confidence: 0.6},
		Summary:          []string{"top diagnosis changed from docker-auth to permission-denied"},
		Evidence:         analysiscompare.StringDelta{Added: []string{"permission denied"}, Removed: []string{"authentication required"}},
	}
	text := FormatCompareText(report)
	for _, want := range []string{"COMPARE", "Previous: prev.json", "Current: curr.json", "docker-auth", "permission-denied", "permission denied"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in compare text, got:\n%s", want, text)
		}
	}
}

func TestFormatCompareTextNoDiagnosis(t *testing.T) {
	report := analysiscompare.Report{
		Summary: []string{"neither artifact contains a matched diagnosis"},
	}
	text := FormatCompareText(report)
	if !strings.Contains(text, "COMPARE") {
		t.Errorf("expected COMPARE header in text, got:\n%s", text)
	}
}

func TestFormatCompareTextHasNewline(t *testing.T) {
	report := analysiscompare.Report{
		Summary: []string{"same diagnosis"},
	}
	text := FormatCompareText(report)
	if !strings.HasSuffix(text, "\n") {
		t.Error("expected trailing newline in FormatCompareText")
	}
}

// ── FormatAnalysisEvidenceText nil/empty ──────────────────────────────────────

func TestFormatAnalysisEvidenceTextNilAnalysis(t *testing.T) {
	text := FormatAnalysisEvidenceText(nil)
	if !strings.Contains(text, "No known playbook matched") {
		t.Errorf("expected no-match message for nil analysis, got: %q", text)
	}
}

func TestFormatAnalysisEvidenceTextNoEvidence(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, nil)
	text := FormatAnalysisEvidenceText(a)
	if !strings.Contains(text, "No extracted evidence") {
		t.Errorf("expected no-evidence message, got: %q", text)
	}
}

func TestFormatAnalysisEvidenceMarkdownNilAnalysis(t *testing.T) {
	text := FormatAnalysisEvidenceMarkdown(nil)
	if !strings.Contains(text, "# No Match") {
		t.Errorf("expected no-match heading for nil analysis, got: %q", text)
	}
}

func TestFormatAnalysisEvidenceMarkdownNoEvidence(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, nil)
	text := FormatAnalysisEvidenceMarkdown(a)
	if !strings.Contains(text, "No extracted evidence") {
		t.Errorf("expected no-evidence message, got: %q", text)
	}
}

func TestParseFormatKnownValues(t *testing.T) {
	tests := []struct {
		input string
		want  Format
	}{
		{"terminal", FormatTerminal},
		{"Terminal", FormatTerminal},
		{"TERMINAL", FormatTerminal},
		{"  terminal  ", FormatTerminal},
		{"markdown", FormatMarkdown},
		{"Markdown", FormatMarkdown},
		{"json", FormatJSON},
		{"JSON", FormatJSON},
	}
	for _, tt := range tests {
		got, ok := ParseFormat(tt.input)
		if !ok {
			t.Errorf("ParseFormat(%q) ok=false, want true", tt.input)
		}
		if got != tt.want {
			t.Errorf("ParseFormat(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseFormatUnknown(t *testing.T) {
	for _, bad := range []string{"", "text", "html", "xml", "plain"} {
		_, ok := ParseFormat(bad)
		if ok {
			t.Errorf("ParseFormat(%q) ok=true, want false", bad)
		}
	}
}

func TestFormatValid(t *testing.T) {
	if !FormatTerminal.Valid() {
		t.Error("FormatTerminal.Valid() = false, want true")
	}
	if !FormatMarkdown.Valid() {
		t.Error("FormatMarkdown.Valid() = false, want true")
	}
	if !FormatJSON.Valid() {
		t.Error("FormatJSON.Valid() = false, want true")
	}
}

func TestFormatInvalidNotValid(t *testing.T) {
	for _, bad := range []Format{"", "html", "plain", "text"} {
		if bad.Valid() {
			t.Errorf("Format(%q).Valid() = true, want false", bad)
		}
	}
}

// ── FormatFix ─────────────────────────────────────────────────────────────────

func TestFormatFix(t *testing.T) {
	a := makeAnalysis("git-auth", "Git auth failure", "auth", 1.0, []string{"terminal prompts disabled"})
	a.Results[0].Playbook.Fix = "1. Export GH_TOKEN\n2. Retry the push"
	out := FormatFix(a, renderer.Options{Plain: true, Width: 88})
	if out == "" {
		t.Fatal("expected non-empty fix output")
	}
	if !strings.Contains(out, "git-auth") {
		t.Errorf("expected playbook ID in fix output, got %q", out)
	}
}

func TestFormatFixNilAnalysis(t *testing.T) {
	out := FormatFix(nil, renderer.Options{Plain: true, Width: 88})
	if !strings.Contains(out, "No known playbook matched") {
		t.Errorf("expected no-match message for nil analysis, got %q", out)
	}
}

// ── FormatPlaybookDetailsJSON ─────────────────────────────────────────────────

func TestFormatPlaybookDetailsJSON(t *testing.T) {
	pb := model.Playbook{
		ID:       "docker-auth",
		Title:    "Docker Registry Auth",
		Category: "auth",
		Severity: "high",
		Fix:      "1. docker login",
		Match:    model.MatchSpec{Any: []string{"pull access denied"}},
	}
	data, err := FormatPlaybookDetailsJSON(pb)
	if err != nil {
		t.Fatalf("FormatPlaybookDetailsJSON: %v", err)
	}
	if !strings.Contains(data, `"docker-auth"`) {
		t.Errorf("expected playbook ID in JSON, got %q", data)
	}
	if !strings.Contains(data, `"Docker Registry Auth"`) {
		t.Errorf("expected playbook title in JSON, got %q", data)
	}
	if !strings.Contains(data, "\n") {
		t.Error("expected JSON to end with newline")
	}
}

// ── topN helper ──────────────────────────────────────────────────────────────

func TestTopNZeroReturnsAll(t *testing.T) {
	results := []model.Result{
		{Playbook: model.Playbook{ID: "a"}},
		{Playbook: model.Playbook{ID: "b"}},
	}
	got := topN(results, 0)
	if len(got) != 2 {
		t.Errorf("topN(results, 0) = %d items, want 2", len(got))
	}
}

func TestTopNNegativeReturnsAll(t *testing.T) {
	results := []model.Result{
		{Playbook: model.Playbook{ID: "a"}},
		{Playbook: model.Playbook{ID: "b"}},
		{Playbook: model.Playbook{ID: "c"}},
	}
	got := topN(results, -1)
	if len(got) != 3 {
		t.Errorf("topN(results, -1) = %d items, want 3", len(got))
	}
}

func TestTopNExceedsLengthReturnsAll(t *testing.T) {
	results := []model.Result{{Playbook: model.Playbook{ID: "a"}}}
	got := topN(results, 10)
	if len(got) != 1 {
		t.Errorf("topN(results, 10) = %d items, want 1", len(got))
	}
}

func TestFormatWorkflowJSON(t *testing.T) {
	plan := workflow.Plan{
		SchemaVersion: "workflow.v1",
		Mode:          workflow.ModeLocal,
		Status:        model.ArtifactStatusMatched,
		FailureID:     "snapshot-mismatch",
		Title:         "Snapshot or golden-file mismatch",
		MetricsHints:  []string{"TSS 0.40 (5 runs)"},
		PolicyHints:   []string{"policy: observe"},
		Verify:        []string{"go test ./..."},
		Steps:         []string{"Inspect the diff."},
		Artifact: &model.FailureArtifact{
			SchemaVersion: "failure_artifact.v1",
			Status:        model.ArtifactStatusMatched,
		},
		Remediation: &model.RemediationPlan{
			Commands: []model.RemediationCommand{{
				ID:      "verify-1",
				Phase:   "verify",
				Command: []string{"go", "test", "./..."},
			}},
		},
	}
	data, err := FormatWorkflowJSON(plan)
	if err != nil {
		t.Fatalf("format workflow json: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(data)), &out); err != nil {
		t.Fatalf("unmarshal workflow json: %v", err)
	}
	if out["failure_id"] != "snapshot-mismatch" {
		t.Fatalf("expected failure_id, got %v", out["failure_id"])
	}
	if out["schema_version"] != "workflow.v1" {
		t.Fatalf("expected schema_version, got %v", out["schema_version"])
	}
	if out["status"] != "matched" {
		t.Fatalf("expected status in workflow JSON, got %v", out["status"])
	}
	if _, ok := out["metrics_hints"].([]interface{}); !ok {
		t.Fatalf("expected metrics_hints in workflow JSON, got %v", out["metrics_hints"])
	}
	if _, ok := out["policy_hints"].([]interface{}); !ok {
		t.Fatalf("expected policy_hints in workflow JSON, got %v", out["policy_hints"])
	}
	if _, ok := out["artifact"].(map[string]interface{}); !ok {
		t.Fatalf("expected artifact in workflow JSON, got %v", out["artifact"])
	}
	if _, ok := out["remediation"].(map[string]interface{}); !ok {
		t.Fatalf("expected remediation in workflow JSON, got %v", out["remediation"])
	}
}

// ── HashAnalysisOutput ────────────────────────────────────────────────────────

func TestHashAnalysisOutputIsDeterministic(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, []string{"pull access denied"})
	h1, err := HashAnalysisOutput(a)
	if err != nil {
		t.Fatalf("HashAnalysisOutput: %v", err)
	}
	h2, err := HashAnalysisOutput(a)
	if err != nil {
		t.Fatalf("HashAnalysisOutput second call: %v", err)
	}
	if h1 == "" {
		t.Fatal("expected non-empty hash")
	}
	if h1 != h2 {
		t.Fatalf("hash is not deterministic: %q != %q", h1, h2)
	}
}

func TestHashAnalysisOutputDistinctForDifferentAnalyses(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, []string{"pull access denied"})
	b := makeAnalysis("git-auth", "Git auth", "auth", 0.8, []string{"fatal: could not read Username"})
	ha, err := HashAnalysisOutput(a)
	if err != nil {
		t.Fatalf("HashAnalysisOutput a: %v", err)
	}
	hb, err := HashAnalysisOutput(b)
	if err != nil {
		t.Fatalf("HashAnalysisOutput b: %v", err)
	}
	if ha == hb {
		t.Fatalf("expected distinct hashes for distinct analyses, got %q", ha)
	}
}

func TestHashAnalysisOutputExcludesOutputHash(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, []string{"pull access denied"})
	a.OutputHash = "output-hash-abc"
	h1, err := HashAnalysisOutput(a)
	if err != nil {
		t.Fatalf("HashAnalysisOutput: %v", err)
	}
	a.OutputHash = "output-hash-xyz"
	h2, err := HashAnalysisOutput(a)
	if err != nil {
		t.Fatalf("HashAnalysisOutput: %v", err)
	}
	// Hash should be stable regardless of OutputHash field value
	if h1 != h2 {
		t.Fatalf("expected same hash regardless of OutputHash field: %q != %q", h1, h2)
	}
}

// ── FormatHookSummariesMarkdown ───────────────────────────────────────────────

func TestFormatHookSummariesMarkdownNilAnalysis(t *testing.T) {
	got := FormatHookSummariesMarkdown(nil)
	if got != "" {
		t.Fatalf("expected empty for nil analysis, got %q", got)
	}
}

func TestFormatHookSummariesMarkdownNoHooks(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, []string{"pull access denied"})
	got := FormatHookSummariesMarkdown(a)
	if got != "" {
		t.Fatalf("expected empty when no hooks, got %q", got)
	}
}

func TestFormatHookSummariesMarkdownWithHooks(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, []string{"pull access denied"})
	passed := true
	a.Results[0].Hooks = &model.HookReport{
		Mode:            "observe",
		BaseConfidence:  0.5,
		FinalConfidence: 0.8,
		ConfidenceDelta: 0.3,
		Results: []model.HookResult{{
			Category: "remediation",
			ID:       "auto-fix",
			Status:   "passed",
			Passed:   &passed,
		}},
	}
	got := FormatHookSummariesMarkdown(a)
	if got == "" {
		t.Fatal("expected non-empty markdown for analysis with hooks")
	}
	if !strings.Contains(got, "## Hooks") {
		t.Fatalf("expected '## Hooks' header, got %q", got)
	}
	if !strings.Contains(got, "docker-auth:") {
		t.Fatalf("expected failure_id prefix, got %q", got)
	}
}

func TestFormatHookSummariesMarkdownEndsWithNewline(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.9, []string{"pull access denied"})
	passed := true
	a.Results[0].Hooks = &model.HookReport{
		Mode:            "observe",
		BaseConfidence:  0.5,
		FinalConfidence: 0.8,
		ConfidenceDelta: 0.3,
		Results: []model.HookResult{{
			Category: "remediation",
			ID:       "auto-fix",
			Status:   "passed",
			Passed:   &passed,
		}},
	}
	got := FormatHookSummariesMarkdown(a)
	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("expected trailing newline, got %q", got)
	}
}

// ── hookHistoryMarkdownLine ───────────────────────────────────────────────────

func TestHookHistoryMarkdownLineMinimal(t *testing.T) {
	summary := &model.HookHistorySummary{TotalCount: 3}
	got := hookHistoryMarkdownLine(summary)
	if !strings.Contains(got, "hook history: 3 run(s)") {
		t.Fatalf("expected total count in line, got %q", got)
	}
}

func TestHookHistoryMarkdownLineAllCounters(t *testing.T) {
	summary := &model.HookHistorySummary{
		TotalCount:    10,
		ExecutedCount: 8,
		PassedCount:   6,
		FailedCount:   1,
		BlockedCount:  1,
		SkippedCount:  2,
		LastSeenAt:    "2026-01-01T00:00:00Z",
	}
	got := hookHistoryMarkdownLine(summary)
	for _, want := range []string{
		"hook history: 10 run(s)",
		"8 executed",
		"6 passed",
		"1 failed",
		"1 blocked",
		"2 skipped",
		"last 2026-01-01T00:00:00Z",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in line, got %q", want, got)
		}
	}
}

func TestHookHistoryMarkdownLineZeroCountsOmitted(t *testing.T) {
	summary := &model.HookHistorySummary{TotalCount: 5}
	got := hookHistoryMarkdownLine(summary)
	for _, absent := range []string{"executed", "passed", "failed", "blocked", "skipped", "last"} {
		if strings.Contains(got, absent) {
			t.Errorf("expected %q to be omitted for zero count, got %q", absent, got)
		}
	}
}

// ── deltaLines ────────────────────────────────────────────────────────────────

func TestDeltaLinesNilDelta(t *testing.T) {
	got := deltaLines(nil)
	if got != nil {
		t.Fatalf("expected nil for nil delta, got %v", got)
	}
}

func TestDeltaLinesWithFields(t *testing.T) {
	delta := &model.Delta{
		Provider:          "github-actions",
		FilesChanged:      []string{"main.go"},
		TestsNewlyFailing: []string{"TestAuth"},
		ErrorsAdded:       []string{"authentication required"},
		Causes: []model.DeltaCause{{
			Kind:    "test_regression",
			Score:   0.9,
			Reasons: []string{"test started failing on this commit"},
		}},
	}
	got := deltaLines(delta)
	for _, want := range []string{
		"Provider: github-actions",
		"Changed file: main.go",
		"New failing test: TestAuth",
		"New error: authentication required",
		"test_regression: 0.90",
		"test_regression reason: test started failing on this commit",
	} {
		found := false
		for _, line := range got {
			if line == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected line %q in delta output, got: %v", want, got)
		}
	}
}

// ── partialMatchLines ─────────────────────────────────────────────────────────

func TestPartialMatchLinesEmpty(t *testing.T) {
	got := partialMatchLines(nil)
	if got != nil {
		t.Fatalf("expected nil for empty groups, got %v", got)
	}
}

func TestPartialMatchLinesWithLabel(t *testing.T) {
	groups := []model.PartialMatchGroup{
		{ID: "auth-check", Label: "Auth Check", Minimum: 1, Patterns: []string{"login failed", "auth error"}},
	}
	got := partialMatchLines(groups)
	if len(got) != 1 {
		t.Fatalf("expected 1 line, got %v", got)
	}
	if !strings.Contains(got[0], "Auth Check") {
		t.Errorf("expected label in line, got %q", got[0])
	}
}

func TestPartialMatchLinesWithoutLabel(t *testing.T) {
	groups := []model.PartialMatchGroup{
		{ID: "", Label: "", Minimum: 2, Patterns: []string{"login failed", "auth error", "access denied"}},
	}
	got := partialMatchLines(groups)
	if len(got) != 1 {
		t.Fatalf("expected 1 line, got %v", got)
	}
	if !strings.Contains(got[0], "2-of-3") {
		t.Errorf("expected minimum count in line, got %q", got[0])
	}
}
