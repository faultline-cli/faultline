package output

import (
	"encoding/json"
	"strings"
	"testing"

	"faultline/internal/model"
	"faultline/internal/renderer"
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
	text := FormatAnalysisText(a, 1, ModeQuick, renderer.Options{Plain: true, Width: 88})
	if !strings.Contains(text, "docker-auth") {
		t.Errorf("expected playbook ID in quick output, got %q", text)
	}
	if !strings.Contains(text, "Summary") {
		t.Errorf("expected Summary section in quick output, got %q", text)
	}
}

func TestFormatAnalysisMarkdownDetailed(t *testing.T) {
	a := makeAnalysis("docker-auth", "Docker auth", "auth", 0.91, []string{"authentication required"})
	a.Context = model.Context{Stage: "deploy"}
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
	for _, want := range []string{"# Docker auth", "- ID: `docker-auth`", "## Summary", "## Evidence", "## Differential Diagnosis", "## Confidence Breakdown", "## Triggered By", "## Score Breakdown", "## Suggested Fix", "## Repo Context"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in markdown output, got:\n%s", want, text)
		}
	}
}

func TestFormatAnalysisTextQuickTopN(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{Playbook: model.Playbook{ID: "a", Title: "A", Summary: "A"}, Confidence: 1.0, Score: 2},
			{Playbook: model.Playbook{ID: "b", Title: "B", Summary: "B"}, Confidence: 0.5, Score: 1},
			{Playbook: model.Playbook{ID: "c", Title: "C", Summary: "C"}, Confidence: 0.3, Score: 0.5},
		},
	}
	text := FormatAnalysisText(a, 2, ModeQuick, renderer.Options{Plain: true, Width: 88})
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

// ── ParseFormat / Valid ───────────────────────────────────────────────────────

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
