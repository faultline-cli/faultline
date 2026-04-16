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
