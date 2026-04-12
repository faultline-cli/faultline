package output

import (
	"encoding/json"
	"strings"
	"testing"

	"faultline/internal/model"
	"faultline/internal/workflow"
)

func makeAnalysis(id, title, category string, confidence float64, evidence []string) *model.Analysis {
	return &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:       id,
					Title:    title,
					Category: category,
					Explain:  "Explanation for " + id,
					Fix:      []string{"Fix step 1", "Fix step 2"},
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
	repoCtx, ok := out["repo_context"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected repo_context object, got %v", out["repo_context"])
	}
	if repoCtx["repo_root"] != "/repo" {
		t.Fatalf("expected repo_root in repo_context, got %v", repoCtx["repo_root"])
	}
}

func TestFormatAnalysisJSONNoMatch(t *testing.T) {
	data, err := FormatAnalysisJSON(nil, 1)
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

// ── Quick text ────────────────────────────────────────────────────────────────

func TestFormatAnalysisTextQuickSingleMatch(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 1.0, []string{"authentication required"})
	text := FormatAnalysisText(a, 1, ModeQuick)
	if !strings.Contains(text, "docker-auth") {
		t.Errorf("expected playbook ID in quick output, got %q", text)
	}
	if !strings.Contains(text, "FIX") {
		t.Errorf("expected FIX section in quick output, got %q", text)
	}
}

func TestFormatAnalysisTextQuickTopN(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{Playbook: model.Playbook{ID: "a", Title: "A", Fix: []string{"fix a"}}, Confidence: 1.0, Score: 2},
			{Playbook: model.Playbook{ID: "b", Title: "B", Fix: []string{"fix b"}}, Confidence: 0.5, Score: 1},
			{Playbook: model.Playbook{ID: "c", Title: "C", Fix: []string{"fix c"}}, Confidence: 0.3, Score: 0.5},
		},
	}
	text := FormatAnalysisText(a, 2, ModeQuick)
	if !strings.Contains(text, "#1") {
		t.Errorf("expected rank header #1, got %q", text)
	}
	if !strings.Contains(text, "#2") {
		t.Errorf("expected rank header #2, got %q", text)
	}
	if strings.Contains(text, "#3") {
		t.Errorf("should not include #3 (top=2), got %q", text)
	}
}

// ── Detailed text ─────────────────────────────────────────────────────────────

func TestFormatAnalysisTextDetailed(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker registry auth failure", "auth", 1.0,
		[]string{"pull access denied"})
	a.Context = model.Context{Stage: "deploy", CommandHint: "docker push"}
	a.RepoContext = &model.RepoContext{
		RepoRoot:           "/repo",
		RecentFiles:        []string{"Dockerfile"},
		RelatedCommits:     []model.RepoCommit{{Hash: "abc1234", Date: "2026-04-10", Subject: "hotfix: adjust docker login"}},
		HotspotDirectories: []string{"deploy"},
		CoChangeHints:      []string{"Dockerfile <-> .github/workflows/deploy.yml"},
		HotfixSignals:      []string{"hotfix: adjust docker login"},
		DriftSignals:       []string{"Repeated edits in deploy"},
	}
	text := FormatAnalysisText(a, 1, ModeDetailed)

	checks := []string{"Diagnosis:", "Category:", "Stage:", "Command:", "Cause", "Fix", "Evidence", "Triggered by", "Score Breakdown", "Repo Context", "Related commit:", "Hotfix signal:"}
	for _, want := range checks {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in detailed output, got:\n%s", want, text)
		}
	}
}

// ── No match ─────────────────────────────────────────────────────────────────

func TestFormatAnalysisTextNilAnalysis(t *testing.T) {
	text := FormatAnalysisText(nil, 1, ModeQuick)
	if !strings.Contains(text, "No known failure") {
		t.Errorf("expected no-match message, got %q", text)
	}
}

// ── CI annotations ────────────────────────────────────────────────────────────

func TestFormatCIAnnotations(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker Auth", "auth", 1.0, nil)
	a.Results[0].Playbook.Fix = []string{"docker login"}
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
		{ID: "aws-credentials", Category: "auth", Severity: "high", Title: "AWS Credentials", Metadata: model.PlaybookMeta{PackName: "faultline-premium-pack"}},
		{ID: "oom-killed", Category: "runtime", Severity: "critical", Title: "OOM Killed"},
	}
	text := FormatPlaybookList(pbs, "")
	if !strings.Contains(text, "docker-auth") || !strings.Contains(text, "oom-killed") || !strings.Contains(text, "faultline-premium-pack") {
		t.Errorf("expected both playbooks in list, got %q", text)
	}
}

func TestFormatPlaybookListCategoryFilter(t *testing.T) {
	pbs := []model.Playbook{
		{ID: "docker-auth", Category: "auth", Title: "Docker Auth"},
		{ID: "oom-killed", Category: "runtime", Title: "OOM Killed"},
	}
	text := FormatPlaybookList(pbs, "auth")
	if !strings.Contains(text, "docker-auth") {
		t.Errorf("expected docker-auth in filtered list, got %q", text)
	}
	if strings.Contains(text, "oom-killed") {
		t.Errorf("oom-killed should be filtered out, got %q", text)
	}
}

func TestFormatPlaybookDetails(t *testing.T) {
	pb := model.Playbook{
		ID:       "docker-auth",
		Title:    "Docker Registry Auth",
		Category: "auth",
		Severity: "high",
		Metadata: model.PlaybookMeta{PackName: "faultline-premium-pack"},
		Explain:  "The CI job could not authenticate.",
		Why:      "Token expired.",
		Fix:      []string{"Run docker login"},
		Prevent:  []string{"Rotate tokens"},
		Match:    model.MatchSpec{Any: []string{"pull access denied"}},
	}
	text := FormatPlaybookDetails(pb)
	for _, want := range []string{"docker-auth", "Docker Registry Auth", "auth", "high", "faultline-premium-pack", "Token expired", "Run docker login"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in details, got:\n%s", want, text)
		}
	}
}

func TestFormatAnalysisTextShowsPremiumPack(t *testing.T) {
	a := makeAnalysis("aws-credentials", "AWS credentials missing or invalid", "auth", 1.0, nil)
	a.Results[0].Playbook.Metadata.PackName = "faultline-premium-pack"
	out := FormatAnalysisText(a, 1, ModeQuick)
	if !strings.Contains(out, "Pack: faultline-premium-pack") {
		t.Fatalf("expected pack line in quick output, got %q", out)
	}
}

func TestFormatAnalysisJSONIncludesPack(t *testing.T) {
	a := makeAnalysis("aws-credentials", "AWS credentials missing or invalid", "auth", 1.0, nil)
	a.Results[0].Playbook.Metadata.PackName = "faultline-premium-pack"
	text, err := FormatAnalysisJSON(a, 1)
	if err != nil {
		t.Fatalf("format analysis json: %v", err)
	}
	if !strings.Contains(text, "\"pack\":\"faultline-premium-pack\"") {
		t.Fatalf("expected pack in json output, got %q", text)
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
		Evidence:    []string{"failed to read Dockerfile"},
		Files:       []string{"Dockerfile", ".dockerignore"},
		LocalRepro:  []string{"docker build -f Dockerfile ."},
		Verify:      []string{"docker build -f Dockerfile ."},
		Steps:       []string{"Verify the exact `docker build` command."},
		AgentPrompt: "You are helping resolve a deterministic CI failure.",
	}

	text := FormatWorkflowText(plan)
	for _, want := range []string{"WORKFLOW", "docker-build-context", "Local repro:", "Verify:", "Next steps:", "Agent prompt:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in workflow text, got:\n%s", want, text)
		}
	}
}

func TestFormatWorkflowJSON(t *testing.T) {
	plan := workflow.Plan{
		SchemaVersion: "workflow.v1",
		Mode:          workflow.ModeLocal,
		FailureID:     "snapshot-mismatch",
		Title:         "Snapshot or golden-file mismatch",
		Verify:        []string{"go test ./..."},
		Steps:         []string{"Inspect the diff."},
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
}
