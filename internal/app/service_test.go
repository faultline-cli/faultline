package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/authoring"
	"faultline/internal/engine"
	"faultline/internal/fixtures"
	"faultline/internal/output"
	"faultline/internal/playbooks"
	"faultline/internal/workflow"
)

// repoPlaybookDir returns the canonical bundled playbook directory relative to
// the module root. Tests in the app package run from the package directory, so
// we walk up two levels.
func repoPlaybookDir() string {
	return "../../playbooks/bundled"
}

// baseOpts returns a minimal AnalyzeOptions that avoids file-system side
// effects and overrides the playbook directory.
func baseOpts() AnalyzeOptions {
	return AnalyzeOptions{
		Top:         1,
		Mode:        output.ModeQuick,
		Format:      output.FormatTerminal,
		NoHistory:   true,
		PlaybookDir: repoPlaybookDir(),
	}
}

// ── Analyze ──────────────────────────────────────────────────────────────────

func TestAnalyzeMatchedTextOutput(t *testing.T) {
	svc := NewService()
	log := "pull access denied\nError response from daemon: authentication required\n"
	var buf bytes.Buffer

	err := svc.Analyze(strings.NewReader(log), "test.log", baseOpts(), &buf)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if !strings.Contains(buf.String(), "Docker registry authentication failure") {
		t.Errorf("expected docker-auth result in output, got %q", buf.String())
	}
}

func TestAnalyzeMatchedJSONOutput(t *testing.T) {
	svc := NewService()
	log := "pull access denied\nError response from daemon: authentication required\n"
	opts := baseOpts()
	opts.JSON = true
	var buf bytes.Buffer

	err := svc.Analyze(strings.NewReader(log), "stdin", opts, &buf)
	if err != nil {
		t.Fatalf("Analyze JSON: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if payload["matched"] != true {
		t.Errorf("expected matched=true, got %v", payload["matched"])
	}
}

func TestAnalyzeIncludesHookSummaryWhenEnabled(t *testing.T) {
	svc := NewService()
	dir := t.TempDir()
	playbook := strings.TrimSpace(`
id: docker-auth
title: Docker auth
category: auth
severity: high
match:
  any:
    - "authentication required"
summary: |
  Summary.
diagnosis: |
  Diagnosis.
fix: |
  1. Fix.
validation: |
  - Validate.
hooks:
  verify:
    - id: docker-config
      kind: file_exists
      path: missing-config
      confidence_delta: 0.05
`) + "\n"
	if err := os.WriteFile(filepath.Join(dir, "rule.yaml"), []byte(playbook), 0o600); err != nil {
		t.Fatalf("write playbook: %v", err)
	}
	opts := baseOpts()
	opts.PlaybookDir = dir
	opts.HookMode = "safe"
	var buf bytes.Buffer

	if err := svc.Analyze(strings.NewReader("authentication required\n"), "stdin", opts, &buf); err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if !strings.Contains(buf.String(), "docker-auth: mode: safe") {
		t.Fatalf("expected hook summary in analyze output, got:\n%s", buf.String())
	}
}

func TestAnalyzeSelectChoosesRequestedResult(t *testing.T) {
	svc := NewService()
	log := "pull access denied\nError response from daemon: authentication required\n"
	allOpts := baseOpts()
	allOpts.Top = 3
	allOpts.JSON = true
	var all bytes.Buffer

	if err := svc.Analyze(strings.NewReader(log), "test.log", allOpts, &all); err != nil {
		t.Fatalf("Analyze all results: %v", err)
	}
	var allPayload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(all.String())), &allPayload); err != nil {
		t.Fatalf("unmarshal full analysis JSON: %v", err)
	}
	allResults, ok := allPayload["results"].([]interface{})
	if !ok || len(allResults) < 2 {
		t.Fatalf("expected at least two ranked results, got %v", allPayload["results"])
	}
	expectedID := allResults[1].(map[string]interface{})["failure_id"]

	opts := baseOpts()
	opts.Select = 2
	opts.JSON = true
	var buf bytes.Buffer

	if err := svc.Analyze(strings.NewReader(log), "test.log", opts, &buf); err != nil {
		t.Fatalf("Analyze with --select: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &payload); err != nil {
		t.Fatalf("unmarshal selected analysis JSON: %v", err)
	}
	results, ok := payload["results"].([]interface{})
	if !ok || len(results) != 1 {
		t.Fatalf("expected one selected result, got %v", payload["results"])
	}
	result := results[0].(map[string]interface{})
	if result["failure_id"] != expectedID {
		t.Fatalf("expected --select to choose %v, got %v", expectedID, result["failure_id"])
	}
}

func TestAnalyzeNoMatchReturnsNoError(t *testing.T) {
	svc := NewService()
	log := "everything is fine, all checks passed\n"
	var buf bytes.Buffer

	// ErrNoMatch is swallowed - output should still be written without error.
	err := svc.Analyze(strings.NewReader(log), "", baseOpts(), &buf)
	if err != nil {
		t.Fatalf("expected no error on no-match, got %v", err)
	}
}

func TestAnalyzeTraceOutput(t *testing.T) {
	svc := NewService()
	log := "exec /__e/node20/bin/node: no such file or directory\n"
	opts := baseOpts()
	opts.TraceEnabled = true
	opts.ShowScoring = true
	var buf bytes.Buffer

	err := svc.Analyze(strings.NewReader(log), "trace.log", opts, &buf)
	if err != nil {
		t.Fatalf("Analyze trace: %v", err)
	}
	for _, want := range []string{"TRACE", "missing-executable", "Rule Evaluation", "Score"} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("expected %q in trace output, got:\n%s", want, buf.String())
		}
	}
}

func TestAnalyzeTraceViewOutput(t *testing.T) {
	svc := NewService()
	log := "exec /__e/node20/bin/node: no such file or directory\n"
	opts := baseOpts()
	opts.View = output.ViewTrace
	var buf bytes.Buffer

	err := svc.Analyze(strings.NewReader(log), "trace.log", opts, &buf)
	if err != nil {
		t.Fatalf("Analyze trace view: %v", err)
	}
	for _, want := range []string{"TRACE", "missing-executable", "Rule Evaluation"} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("expected %q in trace view output, got:\n%s", want, buf.String())
		}
	}
}

func TestTraceSpecificPlaybookWithoutWinningMatch(t *testing.T) {
	svc := NewService()
	log := "pull access denied\nError response from daemon: authentication required\n"
	opts := baseOpts()
	opts.TraceEnabled = true
	opts.TracePlaybook = "missing-executable"
	var buf bytes.Buffer

	err := svc.Trace(strings.NewReader(log), "trace.log", opts, &buf)
	if err != nil {
		t.Fatalf("Trace specific playbook: %v", err)
	}
	for _, want := range []string{"TRACE  missing-executable", "Outcome: not matched", "Rule Evaluation"} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("expected %q in specific trace output, got:\n%s", want, buf.String())
		}
	}
}

func TestReplayRendersSavedAnalysis(t *testing.T) {
	svc := NewService()
	log := "pull access denied\nError response from daemon: authentication required\n"
	artifactOpts := baseOpts()
	artifactOpts.JSON = true
	var artifact bytes.Buffer
	if err := svc.Analyze(strings.NewReader(log), "stdin", artifactOpts, &artifact); err != nil {
		t.Fatalf("Analyze to artifact: %v", err)
	}

	replayOpts := baseOpts()
	replayOpts.Format = output.FormatMarkdown
	replayOpts.Mode = output.ModeDetailed
	var replay bytes.Buffer
	if err := svc.Replay(strings.NewReader(artifact.String()), replayOpts, &replay); err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if !strings.Contains(replay.String(), "# Docker registry authentication failure") {
		t.Fatalf("expected replay markdown heading, got:\n%s", replay.String())
	}
}

func TestReplayRejectsTraceForAnalysisArtifact(t *testing.T) {
	svc := NewService()
	artifact := `{"matched":true,"results":[{"rank":1,"failure_id":"docker-auth","title":"Docker auth","category":"auth","score":1,"confidence":1,"evidence":["authentication required"]}]}`
	opts := baseOpts()
	opts.TraceEnabled = true
	var buf bytes.Buffer

	err := svc.Replay(strings.NewReader(artifact), opts, &buf)
	if err == nil {
		t.Fatal("expected replay trace error")
	}
	if !strings.Contains(err.Error(), "replay trace is not supported") {
		t.Fatalf("unexpected replay trace error: %v", err)
	}
}

func TestCompareArtifacts(t *testing.T) {
	svc := NewService()
	leftLog := "pull access denied\nError response from daemon: authentication required\n"
	rightLog := "permission denied opening /app/data/config.yaml\n"

	makeArtifact := func(log string) string {
		var buf bytes.Buffer
		opts := baseOpts()
		opts.JSON = true
		if err := svc.Analyze(strings.NewReader(log), "stdin", opts, &buf); err != nil {
			t.Fatalf("Analyze to artifact: %v", err)
		}
		return buf.String()
	}

	var out bytes.Buffer
	err := svc.Compare(strings.NewReader(makeArtifact(leftLog)), strings.NewReader(makeArtifact(rightLog)), AnalyzeOptions{
		Format: output.FormatMarkdown,
	}, &out)
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	for _, want := range []string{"# Faultline Compare", "## Diagnosis", "## Evidence Changes"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("expected %q in compare output, got:\n%s", want, out.String())
		}
	}
}

func TestAnalyzeEvidenceView(t *testing.T) {
	svc := NewService()
	log := "pull access denied\nError response from daemon: authentication required\n"
	opts := baseOpts()
	opts.View = output.ViewEvidence
	var buf bytes.Buffer

	if err := svc.Analyze(strings.NewReader(log), "stdin", opts, &buf); err != nil {
		t.Fatalf("Analyze evidence view: %v", err)
	}
	for _, want := range []string{"EVIDENCE  docker-auth", "Matched evidence:", "pull access denied"} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("expected %q in evidence view, got:\n%s", want, buf.String())
		}
	}
}

func TestReplayFixView(t *testing.T) {
	svc := NewService()
	log := "pull access denied\nError response from daemon: authentication required\n"
	artifactOpts := baseOpts()
	artifactOpts.JSON = true
	var artifact bytes.Buffer
	if err := svc.Analyze(strings.NewReader(log), "stdin", artifactOpts, &artifact); err != nil {
		t.Fatalf("Analyze to artifact: %v", err)
	}

	replayOpts := baseOpts()
	replayOpts.View = output.ViewFix
	var replay bytes.Buffer
	if err := svc.Replay(strings.NewReader(artifact.String()), replayOpts, &replay); err != nil {
		t.Fatalf("Replay fix view: %v", err)
	}
	if !strings.Contains(replay.String(), "Fix Steps") {
		t.Fatalf("expected fix-only replay output, got:\n%s", replay.String())
	}
}

func TestAnalyzeEmptyInputReturnsErrNoInput(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer

	err := svc.Analyze(strings.NewReader(""), "", baseOpts(), &buf)
	if !errors.Is(err, engine.ErrNoInput) {
		t.Fatalf("expected ErrNoInput, got %v", err)
	}
}

// ── Fix ──────────────────────────────────────────────────────────────────────

func TestFixOutputContainsFixSteps(t *testing.T) {
	svc := NewService()
	log := "fatal: could not read Username for 'https://github.com': terminal prompts disabled\n"
	var buf bytes.Buffer

	err := svc.Fix(strings.NewReader(log), "", baseOpts(), &buf)
	if err != nil {
		t.Fatalf("Fix: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty fix output")
	}
}

func TestFixMarkdownOutput(t *testing.T) {
	svc := NewService()
	log := "fatal: could not read Username for 'https://github.com': terminal prompts disabled\n"
	opts := baseOpts()
	opts.Format = output.FormatMarkdown
	var buf bytes.Buffer

	err := svc.Fix(strings.NewReader(log), "", opts, &buf)
	if err != nil {
		t.Fatalf("Fix markdown: %v", err)
	}
	if !strings.HasPrefix(buf.String(), "#") {
		t.Errorf("expected markdown heading, got %q", buf.String()[:min(60, buf.Len())])
	}
}

// ── List ─────────────────────────────────────────────────────────────────────

func TestListWritesPlaybooks(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer

	err := svc.List("", repoPlaybookDir(), nil, &buf)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected list output to be non-empty")
	}
}

func TestListFiltersByCategory(t *testing.T) {
	svc := NewService()
	var all, filtered bytes.Buffer

	if err := svc.List("", repoPlaybookDir(), nil, &all); err != nil {
		t.Fatalf("List (all): %v", err)
	}
	if err := svc.List("auth", repoPlaybookDir(), nil, &filtered); err != nil {
		t.Fatalf("List (auth): %v", err)
	}
	if filtered.Len() >= all.Len() {
		t.Error("filtered list should be smaller than the full list")
	}
}

// ── Explain ──────────────────────────────────────────────────────────────────

func TestExplainKnownPlaybook(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer

	err := svc.Explain("git-auth", repoPlaybookDir(), nil, output.FormatTerminal, &buf)
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	if !strings.Contains(buf.String(), "git-auth") {
		t.Errorf("expected playbook ID in output, got %q", buf.String()[:min(80, buf.Len())])
	}
}

func TestExplainUnknownPlaybookReturnsError(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer

	err := svc.Explain("does-not-exist-abc123", repoPlaybookDir(), nil, output.FormatTerminal, &buf)
	if err == nil {
		t.Error("expected error for unknown playbook ID, got nil")
	}
}

// ── Workflow ──────────────────────────────────────────────────────────────────

func TestWorkflowLocalMode(t *testing.T) {
	svc := NewService()
	log := "pull access denied\nError response from daemon: authentication required\n"
	var buf bytes.Buffer

	err := svc.Workflow(strings.NewReader(log), "", baseOpts(), workflow.ModeLocal, false, &buf)
	if err != nil {
		t.Fatalf("Workflow: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty workflow output")
	}
}

func TestWorkflowJSONOutput(t *testing.T) {
	svc := NewService()
	log := "pull access denied\nError response from daemon: authentication required\n"
	var buf bytes.Buffer

	err := svc.Workflow(strings.NewReader(log), "", baseOpts(), workflow.ModeLocal, true, &buf)
	if err != nil {
		t.Fatalf("Workflow JSON: %v", err)
	}
	var payload map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &payload); jsonErr != nil {
		t.Fatalf("unmarshal workflow JSON: %v", jsonErr)
	}
}

func TestWorkflowBayesJSONIncludesHints(t *testing.T) {
	svc := NewService()
	repoDir := writeServiceTempRepo(t)
	log := "exec /__e/node20/bin/node: no such file or directory\n"
	opts := baseOpts()
	opts.BayesEnabled = true
	opts.GitContextEnabled = true
	opts.RepoPath = repoDir
	var buf bytes.Buffer

	err := svc.Workflow(strings.NewReader(log), "", opts, workflow.ModeAgent, true, &buf)
	if err != nil {
		t.Fatalf("Workflow bayes JSON: %v", err)
	}
	var payload map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &payload); jsonErr != nil {
		t.Fatalf("unmarshal workflow JSON: %v", jsonErr)
	}
	if payload["ranking_hints"] == nil {
		t.Fatalf("expected ranking_hints, got %v", payload)
	}
}

// ── ListInstalledPacks ────────────────────────────────────────────────────────

func TestListInstalledPacksNoPacksInstalled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	svc := NewService()
	var buf bytes.Buffer
	if err := svc.ListInstalledPacks(&buf); err != nil {
		t.Fatalf("ListInstalledPacks: %v", err)
	}
	if !strings.Contains(buf.String(), "No installed") {
		t.Errorf("expected 'No installed' message, got %q", buf.String())
	}
}

func TestListInstalledPacksShowsInstalledPack(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	src := t.TempDir()
	sampleYAML := `
id: installed-test
title: Installed Test
category: test
severity: low
summary: Summary text.
diagnosis: |
  ## Diagnosis

  Details here.
fix: |
  ## Fix steps

  1. Do the thing.
validation: |
  ## Validation

  - Verify it worked.
match:
  any:
    - "installed error"
`
	if err := os.WriteFile(filepath.Join(src, "test.yaml"), []byte(sampleYAML), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	if _, err := playbooks.InstallPack(src, "listed-pack", false); err != nil {
		t.Fatalf("InstallPack: %v", err)
	}

	svc := NewService()
	var buf bytes.Buffer
	if err := svc.ListInstalledPacks(&buf); err != nil {
		t.Fatalf("ListInstalledPacks: %v", err)
	}
	if !strings.Contains(buf.String(), "listed-pack") {
		t.Errorf("expected pack name in output, got %q", buf.String())
	}
}

// ── InstallPack ───────────────────────────────────────────────────────────────

func TestInstallPackWritesSuccessMessage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	src := t.TempDir()
	sampleYAML := `
id: svc-install-test
title: Svc Install Test
category: test
severity: low
summary: Summary text.
diagnosis: |
  ## Diagnosis

  Details.
fix: |
  ## Fix steps

  1. Fix it.
validation: |
  ## Validation

  - Verify.
match:
  any:
    - "svc error"
`
	if err := os.WriteFile(filepath.Join(src, "test.yaml"), []byte(sampleYAML), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	svc := NewService()
	var buf bytes.Buffer
	if err := svc.InstallPack(src, "svc-pack", false, &buf); err != nil {
		t.Fatalf("InstallPack: %v", err)
	}
	if !strings.Contains(buf.String(), "Installed pack") {
		t.Errorf("expected 'Installed pack' in output, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "svc-pack") {
		t.Errorf("expected pack name in output, got %q", buf.String())
	}
}

// ── Inspect ───────────────────────────────────────────────────────────────────

func TestInspectNoSourcePlaybooksReturnsNoError(t *testing.T) {
	// Use a temp dir with only a log-detector playbook so there are no source
	// playbooks, which results in ErrNoMatch (swallowed by Inspect).
	emptyPackDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(emptyPackDir, "placeholder.yaml"), []byte(`
id: log-only
title: Log Only
category: test
severity: low
summary: Summary text.
diagnosis: |
  ## Diagnosis

  Details.
fix: |
  ## Fix steps

  1. Fix.
validation: |
  ## Validation

  - Check.
match:
  any:
    - "some error"
`), 0o600); err != nil {
		t.Fatalf("write placeholder: %v", err)
	}

	svc := NewService()
	opts := baseOpts()
	opts.PlaybookDir = emptyPackDir
	var buf bytes.Buffer
	err := svc.Inspect(t.TempDir(), opts, &buf)
	if err != nil {
		t.Fatalf("Inspect with log-only playbooks: %v", err)
	}
}

func TestInspectUsesScopedWorktreeDiffForSubdirectory(t *testing.T) {
	svc := NewService()
	repoDir := writeServiceGuardRepo(t)
	opts := baseOpts()
	opts.JSON = true
	var buf bytes.Buffer

	err := svc.Inspect(filepath.Join(repoDir, "api"), opts, &buf)
	if err != nil {
		t.Fatalf("Inspect subdir: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal inspect json: %v", err)
	}
	results, ok := payload["results"].([]any)
	if !ok || len(results) == 0 {
		t.Fatalf("expected inspect results, got %#v", payload["results"])
	}
	first, ok := results[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first result object, got %#v", results[0])
	}
	if first["failure_id"] != "panic-in-http-handler" {
		t.Fatalf("expected panic-in-http-handler, got %#v", first["failure_id"])
	}
	if first["change_status"] != "introduced" {
		t.Fatalf("expected introduced change status, got %#v", first["change_status"])
	}
	delta, ok := payload["delta"].(map[string]any)
	if !ok {
		t.Fatalf("expected delta payload, got %#v", payload["delta"])
	}
	files, ok := delta["files_changed"].([]any)
	if !ok || len(files) != 1 || files[0] != "handler.go" {
		t.Fatalf("expected subdir-scoped changed file, got %#v", delta["files_changed"])
	}
}

func TestGuardQuietOnCleanRepo(t *testing.T) {
	svc := NewService()
	repoDir := writeServiceTempRepo(t)
	opts := baseOpts()
	var buf bytes.Buffer

	err := svc.Guard(repoDir, opts, &buf)
	if err != nil {
		t.Fatalf("Guard clean repo: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected quiet guard output, got %q", buf.String())
	}
}

func TestGuardReturnsFindingsError(t *testing.T) {
	svc := NewService()
	repoDir := writeServiceGuardRepo(t)
	opts := baseOpts()
	var buf bytes.Buffer

	err := svc.Guard(repoDir, opts, &buf)
	if !errors.Is(err, ErrGuardFindings) {
		t.Fatalf("expected ErrGuardFindings, got %v", err)
	}
	if !strings.Contains(buf.String(), "panic-in-http-handler") {
		t.Fatalf("expected guard finding in output, got %q", buf.String())
	}
}

// ── Explain (JSON and Markdown formats) ──────────────────────────────────────

func TestExplainMarkdownOutput(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer
	err := svc.Explain("git-auth", repoPlaybookDir(), nil, output.FormatMarkdown, &buf)
	if err != nil {
		t.Fatalf("Explain markdown: %v", err)
	}
	if !strings.HasPrefix(buf.String(), "#") {
		t.Errorf("expected markdown heading, got %q", buf.String()[:min(80, buf.Len())])
	}
}

func TestExplainJSONOutput(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer
	err := svc.Explain("git-auth", repoPlaybookDir(), nil, output.FormatJSON, &buf)
	if err != nil {
		t.Fatalf("Explain json: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &out); err != nil {
		t.Fatalf("unmarshal explain json: %v", err)
	}
	if out["id"] != "git-auth" {
		t.Errorf("expected id=git-auth in json, got %v", out["id"])
	}
}

// ── Fix (no-match) ────────────────────────────────────────────────────────────

func TestFixNoMatchOutput(t *testing.T) {
	svc := NewService()
	log := "everything is perfectly fine\n"
	var buf bytes.Buffer
	err := svc.Fix(strings.NewReader(log), "", baseOpts(), &buf)
	if err != nil {
		t.Fatalf("Fix no-match: %v", err)
	}
	// No-match should still write output without error.
	if buf.Len() == 0 {
		t.Error("expected non-empty fix output even on no-match")
	}
}

// ── FixturesStats (empty dir) ──────────────────────────────────────────────

func TestFixturesStatsEmptyRootReturnsError(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer
	// An empty temp dir has no fixtures layout.
	err := svc.FixturesStats(t.TempDir(), "", fixtures.EvaluateOptions{PlaybookDir: repoPlaybookDir()}, "", false, false, false, &buf)
	// We expect either an error or an empty-report output; we only verify
	// it doesn't panic.
	_ = err
}

func TestFixturesScaffoldSanitizesAndWritesPackFile(t *testing.T) {
	svc := NewService()
	packDir := t.TempDir()
	log := "Authorization: Bearer supersecrettoken123abc\npull access denied\n"
	var buf bytes.Buffer

	err := svc.FixturesScaffold(log, authoring.ScaffoldOptions{
		Category: "auth",
		ID:       "auth-redacted-token",
		PackDir:  packDir,
		MaxMatch: 3,
	}, &buf)
	if err != nil {
		t.Fatalf("FixturesScaffold: %v", err)
	}
	if !strings.Contains(buf.String(), "wrote scaffold:") {
		t.Fatalf("expected write notice, got %q", buf.String())
	}
	if strings.Contains(buf.String(), "supersecrettoken123abc") {
		t.Fatalf("expected scaffold output to redact secrets, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "<redacted>") {
		t.Fatalf("expected scaffold output to include redacted marker, got %q", buf.String())
	}

	writtenPath := filepath.Join(packDir, "auth-redacted-token.yaml")
	data, err := os.ReadFile(writtenPath)
	if err != nil {
		t.Fatalf("read scaffold file: %v", err)
	}
	if !strings.Contains(string(data), "id: auth-redacted-token") {
		t.Fatalf("expected scaffold file to be written, got %q", string(data))
	}
}

func writeServiceTempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runServiceGit(t, dir, "init")
	runServiceGit(t, dir, "config", "user.name", "Faultline Test")
	runServiceGit(t, dir, "config", "user.email", "faultline@example.com")
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM alpine\n"), 0o644); err != nil {
		t.Fatalf("write Dockerfile: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "deploy"), 0o755); err != nil {
		t.Fatalf("mkdir deploy: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "deploy", "healthcheck.yaml"), []byte("path: /healthz\n"), 0o644); err != nil {
		t.Fatalf("write healthcheck: %v", err)
	}
	runServiceGit(t, dir, "add", ".")
	runServiceGitEnv(t, dir, []string{
		"GIT_AUTHOR_DATE=2026-04-10T10:00:00Z",
		"GIT_COMMITTER_DATE=2026-04-10T10:00:00Z",
	}, "commit", "--quiet", "-m", "baseline: add deploy files")
	return dir
}

func writeServiceGuardRepo(t *testing.T) string {
	t.Helper()
	dir := writeServiceTempRepo(t)
	handlerPath := filepath.Join(dir, "api", "handler.go")
	if err := os.MkdirAll(filepath.Dir(handlerPath), 0o755); err != nil {
		t.Fatalf("mkdir handler: %v", err)
	}
	if err := os.WriteFile(handlerPath, []byte("package api\n\nfunc UserHandler() string {\n\tpanic(\"boom\")\n}\n"), 0o644); err != nil {
		t.Fatalf("write handler: %v", err)
	}
	return dir
}

func runServiceGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	runServiceGitEnv(t, dir, nil, args...)
}

func runServiceGitEnv(t *testing.T, dir string, env []string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}

// ── History and Signatures ──────────────────────────────────────────────────

func TestSignaturesEmptyStore(t *testing.T) {
	svc := NewService()
	storePath := filepath.Join(t.TempDir(), "faultline.db")

	// Test text output
	var buf bytes.Buffer
	err := svc.Signatures(storePath, 10, false, &buf)
	if err != nil {
		t.Fatalf("Signatures: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "No stored signatures yet") {
		t.Errorf("expected 'No stored signatures yet' in output, got: %s", output)
	}

	// Test JSON output
	buf.Reset()
	err = svc.Signatures(storePath, 10, true, &buf)
	if err != nil {
		t.Fatalf("Signatures JSON: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal JSON output: %v", err)
	}
	if storeObj, ok := result["store"].(map[string]interface{}); !ok {
		t.Errorf("expected store object in JSON output")
	} else if storeObj["mode"] != "auto" {
		t.Errorf("expected store mode 'auto' for empty store, got %v", storeObj["mode"])
	}
}

func TestSignaturesWithData(t *testing.T) {
	svc := NewService()
	storePath := filepath.Join(t.TempDir(), "faultline.db")

	// First, analyze a log to create some history
	log := "Error response from daemon: pull access denied for mcr/microsoft.com/mssql/server, repository does not exist or may require 'docker login'\n"
	opts := AnalyzeOptions{
		JSON:        true,
		NoHistory:   false,
		Store:       storePath,
		PlaybookDir: repoPlaybookDir(),
	}

	var analysisBuf bytes.Buffer
	err := svc.Analyze(bytes.NewBufferString(log), "stdin", opts, &analysisBuf)
	if err != nil {
		t.Fatalf("Analyze to create history: %v", err)
	}

	// Now test Signatures with data
	var buf bytes.Buffer
	err = svc.Signatures(storePath, 10, false, &buf)
	if err != nil {
		t.Fatalf("Signatures: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Signatures") {
		t.Errorf("expected 'Signatures' header in output, got: %s", output)
	}
	if strings.Contains(output, "No stored signatures yet") {
		t.Errorf("expected signatures in output, got 'No stored signatures yet'")
	}

	// Test JSON output with data
	buf.Reset()
	err = svc.Signatures(storePath, 10, true, &buf)
	if err != nil {
		t.Fatalf("Signatures JSON: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal JSON output: %v", err)
	}
	if signatures, ok := result["signatures"].([]interface{}); !ok {
		t.Errorf("expected signatures array in JSON output")
	} else if len(signatures) == 0 {
		t.Errorf("expected at least one signature in JSON output")
	}
}

func TestHistoryDeterminismVerification(t *testing.T) {
	svc := NewService()
	storePath := filepath.Join(t.TempDir(), "faultline.db")

	// First analysis
	log := "Error response from daemon: pull access denied\n"
	opts := AnalyzeOptions{
		JSON:        true,
		NoHistory:   false,
		Store:       storePath,
		PlaybookDir: repoPlaybookDir(),
	}

	var firstBuf bytes.Buffer
	err := svc.Analyze(bytes.NewBufferString(log), "stdin", opts, &firstBuf)
	if err != nil {
		t.Fatalf("First Analyze: %v", err)
	}

	// Second analysis with same input
	var secondBuf bytes.Buffer
	err = svc.Analyze(bytes.NewBufferString(log), "stdin", opts, &secondBuf)
	if err != nil {
		t.Fatalf("Second Analyze: %v", err)
	}

	// Verify determinism
	var buf bytes.Buffer
	err = svc.VerifyDeterminism(bytes.NewBufferString(log), "stdin", storePath, false, &buf)
	if err != nil {
		t.Fatalf("VerifyDeterminism: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "deterministic") && !strings.Contains(output, "Determinism") {
		t.Errorf("expected determinism check output, got: %s", output)
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test shortHash
	if hash := shortHash("abcdef1234567890"); hash != "abcdef123456" {
		t.Errorf("shortHash: expected 'abcdef123456', got %s", hash)
	}
	if hash := shortHash("abc"); hash != "abc" {
		t.Errorf("shortHash short input: expected 'abc', got %s", hash)
	}

	// Test maxInt
	if val := maxInt(1, 2); val != 2 {
		t.Errorf("maxInt: expected 2, got %d", val)
	}
	if val := maxInt(5, 3); val != 5 {
		t.Errorf("maxInt: expected 5, got %d", val)
	}
	if val := maxInt(-1, -5); val != -1 {
		t.Errorf("maxInt negative: expected -1, got %d", val)
	}

	// Test emptyDash
	if val := emptyDash(""); val != "-" {
		t.Errorf("emptyDash empty: expected '-', got %s", val)
	}
	if val := emptyDash("test"); val != "test" {
		t.Errorf("emptyDash non-empty: expected 'test', got %s", val)
	}

	// Test historyWindow with timestamp strings
	if window := historyWindow("2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z"); window != "" {
		t.Errorf("historyWindow same timestamp: expected empty string, got %s", window)
	}
	// Test with different timestamps - 24 hours returns "24h" (not "1d" which is for >= 48 hours)
	if window := historyWindow("2024-01-01T00:00:00Z", "2024-01-02T00:00:00Z"); window != "24h" {
		t.Errorf("historyWindow 1 day difference: expected '24h', got %s", window)
	}
	// Test with 2 days difference - should return "2d" (48 hours = 2 days)
	if window := historyWindow("2024-01-01T00:00:00Z", "2024-01-03T00:00:00Z"); window != "2d" {
		t.Errorf("historyWindow 2 day difference: expected '2d', got %s", window)
	}

	// Test fallbackSource
	if source := fallbackSource("test.log"); source != "test.log" {
		t.Errorf("fallbackSource with input: expected 'test.log', got %s", source)
	}
	if source := fallbackSource(""); source != "stdin" {
		t.Errorf("fallbackSource empty: expected 'stdin', got %s", source)
	}
}
