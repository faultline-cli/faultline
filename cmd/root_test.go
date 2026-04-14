package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func repoPlaybookDir(_ *testing.T) string {
	return "../playbooks/bundled"
}

func writeTempLog(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp log: %v", err)
	}
	return path
}

func writeTempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, nil, "init")
	runGit(t, dir, nil, "config", "user.name", "Faultline Test")
	runGit(t, dir, nil, "config", "user.email", "faultline@example.com")

	filePath := filepath.Join(dir, "deploy", "healthcheck.yaml")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filePath, []byte("path: /healthz\n"), 0o644); err != nil {
		t.Fatalf("write repo file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM alpine\n"), 0o644); err != nil {
		t.Fatalf("write Dockerfile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".dockerignore"), []byte("dist/\n"), 0o644); err != nil {
		t.Fatalf("write .dockerignore: %v", err)
	}
	runGit(t, dir, nil, "add", ".")
	runGit(t, dir, []string{
		"GIT_AUTHOR_DATE=2026-04-10T10:00:00Z",
		"GIT_COMMITTER_DATE=2026-04-10T10:00:00Z",
	}, "commit", "--quiet", "-m", "hotfix: adjust healthcheck config")
	return dir
}

func runGit(t *testing.T, dir string, env []string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}

// ── analyze ───────────────────────────────────────────────────────────────────

func TestAnalyzeFileText(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze: %v", err)
	}
	if !strings.Contains(out.String(), "Docker registry authentication failure") {
		t.Fatalf("expected docker auth result, got %q", out.String())
	}
}

func TestAnalyzeFileJSON(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--json", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze --json: %v", err)
	}
	if strings.Contains(out.String(), "\x1b[") {
		t.Fatalf("json output should not contain ANSI sequences, got %q", out.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if payload["matched"] != true {
		t.Errorf("expected matched=true, got %v", payload["matched"])
	}
}

func TestAnalyzeStdinJSON(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	oldStdin := os.Stdin
	t.Cleanup(func() { os.Stdin = oldStdin })

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	_, _ = writer.WriteString("missing go.sum entry for module providing package\n")
	writer.Close()
	os.Stdin = reader

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--json", "--no-history"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze stdin: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	results, ok := payload["results"].([]interface{})
	if !ok || len(results) == 0 {
		t.Fatalf("expected results in JSON, got %v", payload)
	}
	r := results[0].(map[string]interface{})
	if r["failure_id"] != "go-sum-missing" {
		t.Fatalf("expected go-sum-missing, got %v", r["failure_id"])
	}
}

func TestAnalyzeTopNText(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t,
		"pull access denied\nauthentication required\ncould not read username for 'https://github.com': terminal prompts disabled\n",
	)

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--top", "3", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze --top: %v", err)
	}
	if !strings.Contains(out.String(), "#1") {
		t.Fatalf("expected ranked output with #1, got %q", out.String())
	}
}

func TestAnalyzeDetailedMode(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--mode", "detailed", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze --mode detailed: %v", err)
	}
	if !strings.Contains(out.String(), "Summary") {
		t.Fatalf("expected detailed summary section, got %q", out.String())
	}
}

func TestAnalyzeMarkdownFormat(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--format", "markdown", "--mode", "detailed", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze --format markdown: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "# Docker registry authentication failure") {
		t.Fatalf("expected markdown heading, got %q", got)
	}
	if !strings.Contains(got, "- ID: `docker-auth`") {
		t.Fatalf("expected markdown metadata, got %q", got)
	}
	if !strings.Contains(got, "## Triggered By") {
		t.Fatalf("expected detailed markdown sections, got %q", got)
	}
}

func TestAnalyzeFormatJSON(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--format", "json", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze --format json: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if payload["matched"] != true {
		t.Fatalf("expected matched=true, got %v", payload["matched"])
	}
}

func TestAnalyzeRejectsInvalidFormat(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--format", "html", "--no-history", logPath})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid format error")
	}
	if !strings.Contains(err.Error(), "--format must be \"terminal\", \"markdown\", or \"json\"") {
		t.Fatalf("unexpected invalid format error: %v", err)
	}
}

func TestAnalyzeWithGitContextJSON(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	repoDir := writeTempRepo(t)
	logPath := writeTempLog(t, "Readiness probe failed: HTTP probe failed with statuscode: 500\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--json", "--no-history", "--git", "--repo", repoDir, logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze --git --json: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	repoCtx, ok := payload["repo_context"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected repo_context in JSON payload, got %v", payload["repo_context"])
	}
	if repoCtx["repo_root"] != repoDir {
		t.Fatalf("expected repo_root %q, got %v", repoDir, repoCtx["repo_root"])
	}
}

// ── fix ───────────────────────────────────────────────────────────────────────

func TestFixCommand(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"fix", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute fix: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "docker-auth") {
		t.Fatalf("expected fix output to reference docker-auth, got %q", got)
	}
	if !strings.Contains(got, "Fix steps") && !strings.Contains(got, "Verify the registry username") {
		t.Fatalf("expected markdown fix content in fix output, got %q", got)
	}
}

func TestFixCommandMarkdownFormat(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"fix", "--format", "markdown", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute fix --format markdown: %v", err)
	}
	if !strings.Contains(out.String(), "## Fix") {
		t.Fatalf("expected markdown fix heading, got %q", out.String())
	}
}

// ── list ─────────────────────────────────────────────────────────────────────

func TestListCommand(t *testing.T) {
	cmd := newRootCommand()
	cmd.SetArgs([]string{"list"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", repoPlaybookDir(t))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute list: %v", err)
	}
	if !strings.Contains(out.String(), "docker-auth") {
		t.Fatalf("expected docker-auth in list, got %q", out.String())
	}
}

func TestListCommandWithAdditionalPack(t *testing.T) {
	extra := t.TempDir()
	if err := os.WriteFile(filepath.Join(extra, "extra.yaml"), []byte(`
id: list-extra
title: List Extra
category: test
severity: low
summary: |
  Extra summary.
diagnosis: |
  ## Diagnosis

  Extra diagnosis.
fix: |
  ## Fix steps

  1. Extra fix.
validation: |
  ## Validation

  - Extra validation.
match:
  any:
    - "extra marker"
`), 0o600); err != nil {
		t.Fatalf("write extra pack: %v", err)
	}

	cmd := newRootCommand()
	cmd.SetArgs([]string{"list", "--playbook-pack", extra})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", repoPlaybookDir(t))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute list with playbook pack: %v", err)
	}
	if !strings.Contains(out.String(), "list-extra") {
		t.Fatalf("expected list-extra in list output, got %q", out.String())
	}
}

func TestListCategoryFlag(t *testing.T) {
	cmd := newRootCommand()
	cmd.SetArgs([]string{"list", "--category", "auth"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", repoPlaybookDir(t))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute list --category: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "docker-auth") {
		t.Fatalf("expected docker-auth in auth category list, got %q", got)
	}
	if strings.Contains(got, "oom-killed") {
		t.Fatalf("oom-killed should not appear in auth category, got %q", got)
	}
}

func TestPacksInstallAndAutoLoad(t *testing.T) {
	home := t.TempDir()
	extra := t.TempDir()
	if err := os.MkdirAll(filepath.Join(extra, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(extra, ".git", "config"), []byte("[core]\nrepositoryformatversion = 0\n"), 0o600); err != nil {
		t.Fatalf("write .git config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(extra, "extra.yaml"), []byte("id: extra-installed\n"+
		"title: Installed Pack\n"+
		"category: auth\n"+
		"severity: high\n"+
		"summary: |\n"+
		"  Installed pack summary.\n"+
		"diagnosis: |\n"+
		"  ## Diagnosis\n\n"+
		"  Installed pack diagnosis.\n"+
		"fix: |\n"+
		"  ## Fix steps\n\n"+
		"  1. Installed pack fix.\n"+
		"validation: |\n"+
		"  ## Validation\n\n"+
		"  - Installed pack validation.\n"+
		"match:\n"+
		"  any:\n"+
		"    - \"extra marker\"\n"), 0o600); err != nil {
		t.Fatalf("write extra pack: %v", err)
	}

	install := newRootCommand()
	install.SetArgs([]string{"packs", "install", extra})
	installOut := &bytes.Buffer{}
	install.SetOut(installOut)
	install.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", repoPlaybookDir(t))
	t.Setenv("HOME", home)

	if err := install.Execute(); err != nil {
		t.Fatalf("execute packs install: %v", err)
	}
	if !strings.Contains(installOut.String(), "Installed pack") {
		t.Fatalf("expected install confirmation, got %q", installOut.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".faultline", "packs", filepath.Base(extra), ".git")); !os.IsNotExist(err) {
		t.Fatalf("expected installed pack to skip .git metadata, got err=%v", err)
	}

	list := newRootCommand()
	list.SetArgs([]string{"list"})
	listOut := &bytes.Buffer{}
	list.SetOut(listOut)
	list.SetErr(new(bytes.Buffer))

	if err := list.Execute(); err != nil {
		t.Fatalf("execute list after pack install: %v", err)
	}
	if !strings.Contains(listOut.String(), "extra-installed") {
		t.Fatalf("expected installed playbook in list output, got %q", listOut.String())
	}
	if !strings.Contains(listOut.String(), filepath.Base(extra)) {
		t.Fatalf("expected installed pack name in list output, got %q", listOut.String())
	}

	packs := newRootCommand()
	packs.SetArgs([]string{"packs", "list"})
	packsOut := &bytes.Buffer{}
	packs.SetOut(packsOut)
	packs.SetErr(new(bytes.Buffer))

	if err := packs.Execute(); err != nil {
		t.Fatalf("execute packs list: %v", err)
	}
	if !strings.Contains(packsOut.String(), filepath.Base(extra)) {
		t.Fatalf("expected installed pack in packs list, got %q", packsOut.String())
	}
}

// ── explain ──────────────────────────────────────────────────────────────────

func TestExplainCommand(t *testing.T) {
	cmd := newRootCommand()
	cmd.SetArgs([]string{"explain", "docker-auth"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", repoPlaybookDir(t))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute explain: %v", err)
	}
	if !strings.Contains(out.String(), "docker-auth") || !strings.Contains(out.String(), "Diagnosis") {
		t.Fatalf("expected explain output for docker-auth, got %q", out.String())
	}
}

func TestExplainCommandMarkdownFormat(t *testing.T) {
	cmd := newRootCommand()
	cmd.SetArgs([]string{"explain", "--format", "markdown", "docker-auth"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", repoPlaybookDir(t))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute explain --format markdown: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "# Docker registry authentication failure") {
		t.Fatalf("expected markdown explain heading, got %q", got)
	}
	if !strings.Contains(got, "## Diagnosis") {
		t.Fatalf("expected markdown diagnosis section, got %q", got)
	}
}

func TestWorkflowCommandLocal(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "failed to solve with frontend dockerfile.v0: failed to read Dockerfile: open Dockerfile: no such file or directory\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"workflow", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute workflow: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "WORKFLOW") || !strings.Contains(got, "docker-build-context") {
		t.Fatalf("expected workflow output, got %q", got)
	}
}

func TestWorkflowCommandResolvesLikelyFilesFromRepo(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	repoDir := writeTempRepo(t)
	logPath := writeTempLog(t, "failed to solve with frontend dockerfile.v0: failed to read Dockerfile: open Dockerfile: no such file or directory\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"workflow", "--no-history", "--repo", repoDir, logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute workflow with repo: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Focus files:") {
		t.Fatalf("expected focus files section, got %q", got)
	}
	if !strings.Contains(got, "Dockerfile") || !strings.Contains(got, ".dockerignore") {
		t.Fatalf("expected repo-resolved likely files, got %q", got)
	}
}

func TestWorkflowCommandAgentJSON(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "Received value does not match stored snapshot\nRun with -u to update snapshots\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"workflow", "--json", "--mode", "agent", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute workflow --json: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if payload["failure_id"] != "snapshot-mismatch" {
		t.Fatalf("expected snapshot-mismatch, got %v", payload["failure_id"])
	}
	if payload["agent_prompt"] == "" {
		t.Fatalf("expected agent_prompt, got %v", payload["agent_prompt"])
	}
}

// ── misc ─────────────────────────────────────────────────────────────────────

func TestVersionFlag(t *testing.T) {
	cmd := newRootCommand()
	cmd.SetArgs([]string{"--version"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute --version: %v", err)
	}
	if !strings.Contains(out.String(), "faultline") {
		t.Fatalf("expected version string to mention faultline, got %q", out.String())
	}
}
