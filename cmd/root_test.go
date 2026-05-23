package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"faultline/internal/app"
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

type analyzeJSONResult struct {
	FailureID       string `json:"failure_id"`
	SignatureHash   string `json:"signature_hash"`
	SeenBefore      bool   `json:"seen_before"`
	OccurrenceCount int    `json:"occurrence_count"`
}

type analyzeJSONPayload struct {
	Results []analyzeJSONResult `json:"results"`
}

func runAnalyzeJSONCommand(t *testing.T, logPath string, args ...string) (string, analyzeJSONPayload) {
	t.Helper()
	cmd := newRootCommand()
	fullArgs := append([]string{"analyze", "--json", "--git=false"}, args...)
	fullArgs = append(fullArgs, logPath)
	cmd.SetArgs(fullArgs)
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", repoPlaybookDir(t))
	// Redirect HOME so the auto-store writes to a temp dir rather than ~/.faultline during tests.
	t.Setenv("HOME", filepath.Dir(logPath))
	t.Setenv("FAULTLINE_STORE", "")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze %v: %v", fullArgs, err)
	}
	raw := strings.TrimSpace(out.String())
	var payload analyzeJSONPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("unmarshal analyze JSON: %v\n%s", err, raw)
	}
	if len(payload.Results) == 0 {
		t.Fatalf("expected at least one result in analyze JSON: %s", raw)
	}
	return raw, payload
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
	commitDate := recentGitDate(2)
	runGit(t, dir, []string{
		"GIT_AUTHOR_DATE=" + commitDate,
		"GIT_COMMITTER_DATE=" + commitDate,
	}, "commit", "--quiet", "-m", "hotfix: adjust healthcheck config")
	return dir
}

func recentGitDate(daysAgo int) string {
	return time.Now().UTC().AddDate(0, 0, -daysAgo).Format(time.RFC3339)
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

	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if payload["matched"] != true {
		t.Errorf("expected matched=true, got %v", payload["matched"])
	}
}

func TestAnalyzeDefaultJSONIsHistoryFreeAndStable(t *testing.T) {
	logPath := writeTempLog(t, "exec /__e/node20/bin/node: no such file or directory\n")

	firstRaw, first := runAnalyzeJSONCommand(t, logPath)
	secondRaw, second := runAnalyzeJSONCommand(t, logPath)

	if firstRaw != secondRaw {
		t.Fatalf("expected repeated default analyze JSON to be identical\nfirst:  %s\nsecond: %s", firstRaw, secondRaw)
	}
	if first.Results[0].OccurrenceCount != 0 || first.Results[0].SeenBefore {
		t.Fatalf("expected first default run to omit recurrence, got %#v", first.Results[0])
	}
	if second.Results[0].OccurrenceCount != 0 || second.Results[0].SeenBefore {
		t.Fatalf("expected second default run to omit recurrence, got %#v", second.Results[0])
	}
}

func TestAnalyzeHistoryFlagRecordsAndEmitsRecurrence(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	logPath := writeTempLog(t, "exec /__e/node20/bin/node: no such file or directory\n")

	_, first := runAnalyzeJSONCommand(t, logPath, "--history")
	_, second := runAnalyzeJSONCommand(t, logPath, "--history")

	if first.Results[0].OccurrenceCount != 1 || first.Results[0].SeenBefore {
		t.Fatalf("expected first history run to record occurrence 1 without seen_before, got %#v", first.Results[0])
	}
	if second.Results[0].OccurrenceCount != 2 || !second.Results[0].SeenBefore {
		t.Fatalf("expected second history run to emit recurrence, got %#v", second.Results[0])
	}
}

func TestAnalyzeStorePathOptsInAndNoHistoryForcesOff(t *testing.T) {
	logPath := writeTempLog(t, "exec /__e/node20/bin/node: no such file or directory\n")
	storePath := filepath.Join(t.TempDir(), "faultline.db")

	_, first := runAnalyzeJSONCommand(t, logPath, "--store", storePath)
	_, second := runAnalyzeJSONCommand(t, logPath, "--store", storePath)
	if first.Results[0].OccurrenceCount != 1 || first.Results[0].SeenBefore {
		t.Fatalf("expected explicit store first run to record occurrence 1, got %#v", first.Results[0])
	}
	if second.Results[0].OccurrenceCount != 2 || !second.Results[0].SeenBefore {
		t.Fatalf("expected explicit store second run to emit recurrence, got %#v", second.Results[0])
	}

	offStorePath := filepath.Join(t.TempDir(), "off.db")
	firstRaw, firstOff := runAnalyzeJSONCommand(t, logPath, "--history", "--store", offStorePath, "--no-history")
	secondRaw, secondOff := runAnalyzeJSONCommand(t, logPath, "--history", "--store", offStorePath, "--no-history")
	if firstRaw != secondRaw {
		t.Fatalf("expected --no-history output to stay stable\nfirst:  %s\nsecond: %s", firstRaw, secondRaw)
	}
	if firstOff.Results[0].OccurrenceCount != 0 || firstOff.Results[0].SeenBefore {
		t.Fatalf("expected --no-history first run to force recurrence off, got %#v", firstOff.Results[0])
	}
	if secondOff.Results[0].OccurrenceCount != 0 || secondOff.Results[0].SeenBefore {
		t.Fatalf("expected --no-history second run to force recurrence off, got %#v", secondOff.Results[0])
	}
}

func TestAnalyzeJSONIncludesPackProvenance(t *testing.T) {
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

	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	prov, ok := payload["pack_provenance"].([]any)
	if !ok || len(prov) == 0 {
		t.Fatalf("expected non-empty pack_provenance in JSON, got %v", payload["pack_provenance"])
	}
	first := prov[0].(map[string]any)
	if first["name"] == "" || first["name"] == nil {
		t.Errorf("expected pack name in provenance entry, got %v", first)
	}
	if first["playbook_count"] == nil {
		t.Errorf("expected playbook_count in provenance entry, got %v", first)
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

	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	results, ok := payload["results"].([]any)
	if !ok || len(results) == 0 {
		t.Fatalf("expected results in JSON, got %v", payload)
	}
	r := results[0].(map[string]any)
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
	if !strings.Contains(out.String(), "Other Likely Matches") || !strings.Contains(out.String(), "#2") {
		t.Fatalf("expected ranked alternatives output, got %q", out.String())
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

func TestAnalyzeEvidenceView(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--view", "evidence", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze --view evidence: %v", err)
	}
	if !strings.Contains(out.String(), "EVIDENCE  docker-auth") {
		t.Fatalf("expected evidence view output, got %q", out.String())
	}
}

func TestAnalyzeTraceViewRemoved(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "exec /__e/node20/bin/node: no such file or directory\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--view", "trace", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected analyze --view trace to be rejected")
	}
	if !strings.Contains(err.Error(), "--view trace was removed from analyze") {
		t.Fatalf("expected removed trace view guidance, got %v", err)
	}
}

func TestAnalyzeRejectsViewWithJSON(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--json", "--view", "fix", "--no-history", logPath})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid --view/--json combination")
	}
	if !strings.Contains(err.Error(), "--view cannot be combined with --json") {
		t.Fatalf("unexpected error: %v", err)
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
	if !strings.Contains(got, "## Differential Diagnosis") || !strings.Contains(got, "## Confidence Breakdown") || !strings.Contains(got, "## Suggested Fix") {
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

	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if payload["matched"] != true {
		t.Fatalf("expected matched=true, got %v", payload["matched"])
	}
}

func TestAnalyzeBayesJSONIncludesRankingAndDelta(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	repoDir := writeTempRepo(t)
	logPath := writeTempLog(t, "exec /__e/node20/bin/node: no such file or directory\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--json", "--bayes", "--git", "--repo", repoDir, "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze --bayes --json: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	results, ok := payload["results"].([]any)
	if !ok || len(results) == 0 {
		t.Fatalf("expected results, got %v", payload["results"])
	}
	first := results[0].(map[string]any)
	if first["ranking"] == nil {
		t.Fatalf("expected ranking payload, got %v", first)
	}
	if payload["delta"] == nil {
		t.Fatalf("expected delta payload, got %v", payload)
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

	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	repoCtx, ok := payload["repo_context"].(map[string]any)
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

	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if payload["schema_version"] != "workflow.v1" {
		t.Fatalf("expected workflow.v1 schema, got %v", payload["schema_version"])
	}
	if payload["failure_id"] != "snapshot-mismatch" {
		t.Fatalf("expected snapshot-mismatch, got %v", payload["failure_id"])
	}
	if payload["agent_prompt"] == "" {
		t.Fatalf("expected agent_prompt, got %v", payload["agent_prompt"])
	}
}

func TestWorkflowCommandBayesJSONIncludesHints(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	repoDir := writeTempRepo(t)
	logPath := writeTempLog(t, "exec /__e/node20/bin/node: no such file or directory\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"workflow", "--json", "--mode", "agent", "--bayes", "--git", "--repo", repoDir, "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute workflow --bayes --json: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal workflow JSON: %v", err)
	}
	if payload["ranking_hints"] == nil {
		t.Fatalf("expected ranking_hints, got %v", payload)
	}
}

// ── misc ─────────────────────────────────────────────────────────────────────

func TestReportCommandAggregatesDefaultAnalyzeRuns(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	home := t.TempDir()
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	runAnalyze := func() {
		cmd := newRootCommand()
		cmd.SetArgs([]string{"analyze", "--json", "--git=false", logPath})
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)
		t.Setenv("HOME", home)
		t.Setenv("FAULTLINE_STORE", "")
		if err := cmd.Execute(); err != nil {
			t.Fatalf("execute analyze for report setup: %v", err)
		}
	}
	runAnalyze()
	runAnalyze()

	cmd := newRootCommand()
	cmd.SetArgs([]string{"report"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("HOME", home)
	t.Setenv("FAULTLINE_STORE", "")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute report: %v", err)
	}

	got := out.String()
	for _, want := range []string{"Faultline Report", "docker-auth", "2", "pull access denied"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected report output to contain %q, got:\n%s", want, got)
		}
	}
}

func TestReportCommandEmptyStoreFriendlyMessage(t *testing.T) {
	cmd := newRootCommand()
	cmd.SetArgs([]string{"report"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("HOME", t.TempDir())
	t.Setenv("FAULTLINE_STORE", "")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute empty report: %v", err)
	}
	if got := out.String(); !strings.Contains(got, "No stored failures yet.") || !strings.Contains(got, "faultline analyze") {
		t.Fatalf("expected friendly empty report message, got %q", got)
	}
}

func TestReportCommandJSON(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	storePath := filepath.Join(t.TempDir(), "faultline.db")
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	for i := 0; i < 2; i++ {
		cmd := newRootCommand()
		cmd.SetArgs([]string{"analyze", "--json", "--store", storePath, logPath})
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("execute analyze for report JSON setup: %v", err)
		}
	}

	cmd := newRootCommand()
	cmd.SetArgs([]string{"report", "--json", "--store", storePath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute report --json: %v", err)
	}

	var payload struct {
		Failures []struct {
			FailureID       string `json:"failure_id"`
			Count           int    `json:"count"`
			LastSeenAt      string `json:"last_seen_at"`
			ExampleEvidence string `json:"example_evidence"`
		} `json:"failures"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal report JSON: %v", err)
	}
	if len(payload.Failures) != 1 {
		t.Fatalf("expected one failure row, got %#v", payload.Failures)
	}
	if payload.Failures[0].FailureID != "docker-auth" || payload.Failures[0].Count != 2 {
		t.Fatalf("unexpected report JSON row: %#v", payload.Failures[0])
	}
	if payload.Failures[0].LastSeenAt == "" || payload.Failures[0].ExampleEvidence == "" {
		t.Fatalf("expected timestamp and example evidence, got %#v", payload.Failures[0])
	}
}

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

// ── --fail-on-silent ──────────────────────────────────────────────────────────

// TestFailOnSilentReturnsErrSilentFailure verifies that analyze exits with
// ErrSilentFailure when --fail-on-silent is set and a silent finding is detected.
func TestFailOnSilentReturnsErrSilentFailure(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	// Log with || true triggers ignored-exit-code detector.
	logPath := writeTempLog(t, "Run npm test\n> jest\nAll tests ran.\nnpm test || true\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--no-history", "--git=false", "--fail-on-silent", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	err := cmd.Execute()
	if !errors.Is(err, app.ErrSilentFailure) {
		t.Fatalf("expected ErrSilentFailure, got %v", err)
	}
}

// TestFailOnSilentNoFindingsExitsZero verifies that --fail-on-silent does not
// cause a non-zero exit when no silent finding is detected.
func TestFailOnSilentNoFindingsExitsZero(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "Run npm test\nPASS src/utils.test.ts\nTests: 5 passed, 5 total\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--no-history", "--git=false", "--fail-on-silent", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected zero exit with no silent findings, got %v", err)
	}
}

// TestFailOnSilentJSONContainsFaultlineStatus verifies that when a silent
// finding is present, the JSON output includes faultline_status=failure.
func TestFailOnSilentJSONContainsFaultlineStatus(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "Run npm test\n> jest\nNo tests found\nnpm test || true\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--no-history", "--git=false", "--json", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	_ = cmd.Execute() // may return ErrNoMatch; that's OK

	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v (output: %q)", err, out.String())
	}

	if payload["faultline_status"] != "failure" {
		t.Errorf("expected faultline_status=failure, got %v", payload["faultline_status"])
	}
	if payload["failure_class"] != "silent_failure" {
		t.Errorf("expected failure_class=silent_failure, got %v", payload["failure_class"])
	}

	// findings array should be present
	findings, ok := payload["findings"].([]any)
	if !ok || len(findings) == 0 {
		t.Errorf("expected non-empty findings array, got %v", payload["findings"])
	}
}
