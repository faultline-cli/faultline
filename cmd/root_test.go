package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

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

func writeTempAnalysisArtifact(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "analysis.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp analysis artifact: %v", err)
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

func writeTempGuardRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, nil, "init")
	runGit(t, dir, nil, "config", "user.name", "Faultline Test")
	runGit(t, dir, nil, "config", "user.email", "faultline@example.com")

	handlerPath := filepath.Join(dir, "api", "handler.go")
	if err := os.MkdirAll(filepath.Dir(handlerPath), 0o755); err != nil {
		t.Fatalf("mkdir handler dir: %v", err)
	}
	if err := os.WriteFile(handlerPath, []byte("package api\n\nfunc UserHandler() string { return \"ok\" }\n"), 0o644); err != nil {
		t.Fatalf("write handler file: %v", err)
	}
	runGit(t, dir, nil, "add", ".")
	runGit(t, dir, []string{
		"GIT_AUTHOR_DATE=2026-04-10T10:00:00Z",
		"GIT_COMMITTER_DATE=2026-04-10T10:00:00Z",
	}, "commit", "--quiet", "-m", "baseline: add handler")

	if err := os.WriteFile(handlerPath, []byte("package api\n\nfunc UserHandler() string {\n\tpanic(\"boom\")\n}\n"), 0o644); err != nil {
		t.Fatalf("rewrite handler file: %v", err)
	}
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

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	prov, ok := payload["pack_provenance"].([]interface{})
	if !ok || len(prov) == 0 {
		t.Fatalf("expected non-empty pack_provenance in JSON, got %v", payload["pack_provenance"])
	}
	first := prov[0].(map[string]interface{})
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

func TestAnalyzeTraceView(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "exec /__e/node20/bin/node: no such file or directory\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--view", "trace", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze --view trace: %v", err)
	}
	if !strings.Contains(out.String(), "TRACE") || !strings.Contains(out.String(), "Rule Evaluation") {
		t.Fatalf("expected trace view output, got %q", out.String())
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

	var payload map[string]interface{}
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

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	results, ok := payload["results"].([]interface{})
	if !ok || len(results) == 0 {
		t.Fatalf("expected results, got %v", payload["results"])
	}
	first := results[0].(map[string]interface{})
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

func TestReplayCommandMarkdown(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	svc := app.NewService()
	var artifact bytes.Buffer
	if err := svc.Analyze(strings.NewReader("pull access denied\nError response from daemon: authentication required\n"), "stdin", app.AnalyzeOptions{
		Top:         1,
		Mode:        "quick",
		Format:      "json",
		JSON:        true,
		NoHistory:   true,
		PlaybookDir: playbookDir,
	}, &artifact); err != nil {
		t.Fatalf("build analysis artifact: %v", err)
	}
	artifactPath := writeTempAnalysisArtifact(t, artifact.String())

	cmd := newRootCommand()
	cmd.SetArgs([]string{"replay", "--format", "markdown", "--mode", "detailed", artifactPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute replay: %v", err)
	}
	if !strings.Contains(out.String(), "# Docker registry authentication failure") {
		t.Fatalf("expected replay markdown heading, got %q", out.String())
	}
}

func TestReplayCommandJSONSelect(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	svc := app.NewService()
	var artifact bytes.Buffer
	if err := svc.Analyze(strings.NewReader("pull access denied\nError response from daemon: authentication required\n"), "stdin", app.AnalyzeOptions{
		Top:         2,
		Mode:        "quick",
		Format:      "json",
		JSON:        true,
		NoHistory:   true,
		PlaybookDir: playbookDir,
	}, &artifact); err != nil {
		t.Fatalf("build analysis artifact: %v", err)
	}
	artifactPath := writeTempAnalysisArtifact(t, artifact.String())

	cmd := newRootCommand()
	cmd.SetArgs([]string{"replay", "--json", "--select", "2", artifactPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute replay --select: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal replay JSON: %v", err)
	}
	results, ok := payload["results"].([]interface{})
	if !ok || len(results) != 1 {
		t.Fatalf("expected one replay-selected result, got %v", payload["results"])
	}
}

func TestReplayCommandFixView(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	svc := app.NewService()
	var artifact bytes.Buffer
	if err := svc.Analyze(strings.NewReader("pull access denied\nError response from daemon: authentication required\n"), "stdin", app.AnalyzeOptions{
		Top:         1,
		Mode:        "quick",
		Format:      "json",
		JSON:        true,
		NoHistory:   true,
		PlaybookDir: playbookDir,
	}, &artifact); err != nil {
		t.Fatalf("build analysis artifact: %v", err)
	}
	artifactPath := writeTempAnalysisArtifact(t, artifact.String())

	cmd := newRootCommand()
	cmd.SetArgs([]string{"replay", "--view", "fix", artifactPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute replay --view fix: %v", err)
	}
	if !strings.Contains(out.String(), "Fix Steps") {
		t.Fatalf("expected fix view output, got %q", out.String())
	}
}

func TestReplayCommandRejectsTraceView(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	svc := app.NewService()
	var artifact bytes.Buffer
	if err := svc.Analyze(strings.NewReader("pull access denied\nError response from daemon: authentication required\n"), "stdin", app.AnalyzeOptions{
		Top:         1,
		Mode:        "quick",
		Format:      "json",
		JSON:        true,
		NoHistory:   true,
		PlaybookDir: playbookDir,
	}, &artifact); err != nil {
		t.Fatalf("build analysis artifact: %v", err)
	}
	artifactPath := writeTempAnalysisArtifact(t, artifact.String())

	cmd := newRootCommand()
	cmd.SetArgs([]string{"replay", "--view", "trace", artifactPath})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected replay trace view error")
	}
	if !strings.Contains(err.Error(), "replay trace is not supported") {
		t.Fatalf("unexpected replay trace view error: %v", err)
	}
}

func TestCompareCommandMarkdown(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	svc := app.NewService()

	makeArtifact := func(log string) string {
		var artifact bytes.Buffer
		if err := svc.Analyze(strings.NewReader(log), "stdin", app.AnalyzeOptions{
			Top:         1,
			Mode:        "quick",
			Format:      "json",
			JSON:        true,
			NoHistory:   true,
			PlaybookDir: playbookDir,
		}, &artifact); err != nil {
			t.Fatalf("build analysis artifact: %v", err)
		}
		return writeTempAnalysisArtifact(t, artifact.String())
	}

	left := makeArtifact("pull access denied\nError response from daemon: authentication required\n")
	right := makeArtifact("pull access denied\npermission denied\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"compare", "--format", "markdown", left, right})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute compare: %v", err)
	}
	if !strings.Contains(out.String(), "# Faultline Compare") || !strings.Contains(out.String(), "## Diagnosis") {
		t.Fatalf("expected compare markdown output, got %q", out.String())
	}
}

func TestCompareCommandJSON(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	svc := app.NewService()

	makeArtifact := func(log string) string {
		var artifact bytes.Buffer
		if err := svc.Analyze(strings.NewReader(log), "stdin", app.AnalyzeOptions{
			Top:         1,
			Mode:        "quick",
			Format:      "json",
			JSON:        true,
			NoHistory:   true,
			PlaybookDir: playbookDir,
		}, &artifact); err != nil {
			t.Fatalf("build analysis artifact: %v", err)
		}
		return writeTempAnalysisArtifact(t, artifact.String())
	}

	left := makeArtifact("pull access denied\nError response from daemon: authentication required\n")
	right := makeArtifact("pull access denied\npermission denied\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"compare", "--json", left, right})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute compare --json: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal compare JSON: %v", err)
	}
	if payload["changed"] != true {
		t.Fatalf("expected changed=true, got %v", payload["changed"])
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
	// VERSION and PINNED REF columns should be present in the header.
	if !strings.Contains(packsOut.String(), "VERSION") || !strings.Contains(packsOut.String(), "PINNED REF") {
		t.Fatalf("expected VERSION and PINNED REF columns in packs list, got %q", packsOut.String())
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

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal workflow JSON: %v", err)
	}
	if payload["ranking_hints"] == nil {
		t.Fatalf("expected ranking_hints, got %v", payload)
	}
}

func TestGuardCommandQuietOnCleanRepo(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	repoDir := writeTempRepo(t)

	cmd := newRootCommand()
	cmd.SetArgs([]string{"guard", repoDir})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute guard on clean repo: %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected quiet guard output on clean repo, got %q", out.String())
	}
}

func TestGuardCommandJSONNoFindings(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	repoDir := writeTempRepo(t)

	cmd := newRootCommand()
	cmd.SetArgs([]string{"guard", "--json", repoDir})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute guard --json on clean repo: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal guard json: %v", err)
	}
	if payload["matched"] != false {
		t.Fatalf("expected matched=false, got %v", payload["matched"])
	}
}

func TestGuardCommandReturnsNonZeroOnFindings(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	repoDir := writeTempGuardRepo(t)

	cmd := newRootCommand()
	cmd.SetArgs([]string{"guard", repoDir})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected guard findings error")
	}
	if err != app.ErrGuardFindings {
		t.Fatalf("expected ErrGuardFindings, got %v", err)
	}
	if !strings.Contains(out.String(), "panic-in-http-handler") {
		t.Fatalf("expected guard finding in output, got %q", out.String())
	}
}

// ── trace ────────────────────────────────────────────────────────────────────

func TestTraceCommandText(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"trace", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute trace: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "TRACE") {
		t.Fatalf("expected TRACE header in output, got %q", got)
	}
	if !strings.Contains(got, "docker-auth") {
		t.Fatalf("expected docker-auth playbook in trace output, got %q", got)
	}
	if !strings.Contains(got, "Rule Evaluation") {
		t.Fatalf("expected Rule Evaluation section in trace output, got %q", got)
	}
}

func TestTraceCommandMarkdownFormat(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"trace", "--format", "markdown", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute trace --format markdown: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "# Faultline Trace") {
		t.Fatalf("expected markdown trace heading, got %q", got)
	}
	if !strings.Contains(got, "## Rule Evaluation") {
		t.Fatalf("expected markdown rule evaluation section, got %q", got)
	}
	if !strings.Contains(got, "docker-auth") {
		t.Fatalf("expected docker-auth in markdown trace output, got %q", got)
	}
}

func TestTraceCommandJSON(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"trace", "--json", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute trace --json: %v", err)
	}
	if strings.Contains(out.String(), "\x1b[") {
		t.Fatalf("trace json output should not contain ANSI sequences, got %q", out.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal trace JSON: %v", err)
	}
	if payload["playbook_id"] != "docker-auth" {
		t.Fatalf("expected playbook_id=docker-auth, got %v", payload["playbook_id"])
	}
	if payload["matched"] != true {
		t.Fatalf("expected matched=true, got %v", payload["matched"])
	}
	rules, ok := payload["rules"].([]interface{})
	if !ok || len(rules) == 0 {
		t.Fatalf("expected non-empty rules array, got %v", payload["rules"])
	}
}

func TestTraceCommandPlaybookFlag(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	// Use a log that triggers go-sum-missing, not docker-auth.
	logPath := writeTempLog(t, "missing go.sum entry for module providing package\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"trace", "--playbook", "docker-auth", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute trace --playbook docker-auth: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "docker-auth") {
		t.Fatalf("expected docker-auth in trace output from --playbook flag, got %q", got)
	}
	if !strings.Contains(got, "not matched") {
		t.Fatalf("expected not matched status for unmatched playbook, got %q", got)
	}
}

func TestTraceCommandSelectRank(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	// Use a log that triggers multiple ranked matches.
	logPath := writeTempLog(t,
		"pull access denied\nauthentication required\ncould not read username for 'https://github.com': terminal prompts disabled\n",
	)

	cmd := newRootCommand()
	cmd.SetArgs([]string{"trace", "--select", "2", "--no-history", logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute trace --select 2: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "TRACE") {
		t.Fatalf("expected TRACE header for rank-2 result, got %q", got)
	}
}

// ── misc ─────────────────────────────────────────────────────────────────────

// ── fixtures scaffold ────────────────────────────────────────────────────────

func TestFixturesScaffoldFromStdin(t *testing.T) {
	logInput := "pull access denied\nError response from daemon: unauthorized: authentication required\n"

	cmd := newRootCommand()
	cmd.SetArgs([]string{"fixtures", "scaffold", "--category", "auth"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetIn(strings.NewReader(logInput))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute fixtures scaffold: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "id: auth-") {
		t.Errorf("expected id starting with auth- in scaffold, got %q", got)
	}
	if !strings.Contains(got, "match:") {
		t.Errorf("expected match section in scaffold, got %q", got)
	}
	if !strings.Contains(got, "TODO") {
		t.Errorf("expected TODO placeholders in scaffold, got %q", got)
	}
}

func TestFixturesScaffoldFromLogFile(t *testing.T) {
	logPath := writeTempLog(t, "fatal: remote error: repository not found\nError: process completed with exit code 128\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"fixtures", "scaffold", "--log", logPath, "--category", "network", "--max-match", "3"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute fixtures scaffold --log: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "category: network") {
		t.Errorf("expected category=network in scaffold, got %q", got)
	}
	if !strings.Contains(got, "match:") {
		t.Errorf("expected match section, got %q", got)
	}
}

func TestFixturesScaffoldIDOverride(t *testing.T) {
	logPath := writeTempLog(t, "error: cannot find module 'react'\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"fixtures", "scaffold", "--log", logPath, "--id", "build-missing-react-module", "--category", "build"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute fixtures scaffold --id: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "id: build-missing-react-module") {
		t.Errorf("expected explicit id in scaffold, got %q", got)
	}
}

func TestFixturesScaffoldWritesToPackDir(t *testing.T) {
	packDir := t.TempDir()
	logPath := writeTempLog(t, "permission denied (publickey)\nAuthentication failed\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"fixtures", "scaffold", "--log", logPath, "--category", "auth", "--pack-dir", packDir})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute fixtures scaffold --pack-dir: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "wrote scaffold:") {
		t.Errorf("expected 'wrote scaffold:' prefix when --pack-dir is set, got %q", got)
	}
}

func TestFixturesScaffoldDeterministic(t *testing.T) {
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: unauthorized: authentication required\n")

	run := func() string {
		cmd := newRootCommand()
		cmd.SetArgs([]string{"fixtures", "scaffold", "--log", logPath, "--category", "auth"})
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(new(bytes.Buffer))
		if err := cmd.Execute(); err != nil {
			t.Fatalf("execute fixtures scaffold: %v", err)
		}
		return out.String()
	}

	first := run()
	second := run()
	if first != second {
		t.Errorf("scaffold output is not deterministic:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestFixturesScaffoldRejectsInvalidCategory(t *testing.T) {
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: unauthorized: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"fixtures", "scaffold", "--log", logPath, "--category", "unicorn"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid category error")
	}
	if !strings.Contains(err.Error(), "invalid category") {
		t.Fatalf("expected invalid category error, got %v", err)
	}
}

func TestFixturesScaffoldRejectsMultipleInputSources(t *testing.T) {
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: unauthorized: authentication required\n")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"fixtures", "scaffold", "--log", logPath, logPath, "--category", "auth"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected multiple input sources error")
	}
	if !strings.Contains(err.Error(), "choose exactly one log source") {
		t.Fatalf("expected multiple input sources error, got %v", err)
	}
}

func TestHistoryCommandJSON(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	storePath := filepath.Join(t.TempDir(), "faultline.db")
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	runAnalyze := func() {
		cmd := newRootCommand()
		cmd.SetArgs([]string{"analyze", "--json", "--store", storePath, logPath})
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(new(bytes.Buffer))
		t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("execute analyze for history setup: %v", err)
		}
	}
	runAnalyze()
	runAnalyze()

	cmd := newRootCommand()
	cmd.SetArgs([]string{"history", "--json", "--store", storePath, "--limit", "5"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute history --json: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal history JSON: %v", err)
	}
	signatures, ok := payload["signatures"].([]interface{})
	if !ok || len(signatures) == 0 {
		t.Fatalf("expected signatures in history payload, got %v", payload["signatures"])
	}
	playbooks, ok := payload["playbooks"].([]interface{})
	if !ok || len(playbooks) == 0 {
		t.Fatalf("expected playbooks in history payload, got %v", payload["playbooks"])
	}
}

func TestSignaturesAndHistorySignatureCommands(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	storePath := filepath.Join(t.TempDir(), "faultline.db")
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	analyze := newRootCommand()
	analyze.SetArgs([]string{"analyze", "--json", "--store", storePath, logPath})
	analyzeOut := &bytes.Buffer{}
	analyze.SetOut(analyzeOut)
	analyze.SetErr(new(bytes.Buffer))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)
	if err := analyze.Execute(); err != nil {
		t.Fatalf("execute analyze: %v", err)
	}

	var analysisPayload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(analyzeOut.String())), &analysisPayload); err != nil {
		t.Fatalf("unmarshal analyze JSON: %v", err)
	}
	results := analysisPayload["results"].([]interface{})
	signatureHash := results[0].(map[string]interface{})["signature_hash"].(string)

	signaturesCmd := newRootCommand()
	signaturesCmd.SetArgs([]string{"signatures", "--json", "--store", storePath})
	signaturesOut := &bytes.Buffer{}
	signaturesCmd.SetOut(signaturesOut)
	signaturesCmd.SetErr(new(bytes.Buffer))
	if err := signaturesCmd.Execute(); err != nil {
		t.Fatalf("execute signatures --json: %v", err)
	}
	if !strings.Contains(signaturesOut.String(), signatureHash) {
		t.Fatalf("expected signatures output to include %s, got %q", signatureHash, signaturesOut.String())
	}

	historyCmd := newRootCommand()
	historyCmd.SetArgs([]string{"history", "--json", "--store", storePath, "--signature", signatureHash})
	historyOut := &bytes.Buffer{}
	historyCmd.SetOut(historyOut)
	historyCmd.SetErr(new(bytes.Buffer))
	if err := historyCmd.Execute(); err != nil {
		t.Fatalf("execute history --signature: %v", err)
	}

	var historyPayload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(historyOut.String())), &historyPayload); err != nil {
		t.Fatalf("unmarshal history detail JSON: %v", err)
	}
	signature, ok := historyPayload["signature"].(map[string]interface{})
	if !ok || signature["signature_hash"] != signatureHash {
		t.Fatalf("expected signature detail payload, got %v", historyPayload["signature"])
	}
	findings, ok := historyPayload["findings"].([]interface{})
	if !ok || len(findings) != 1 {
		t.Fatalf("expected one recent finding, got %v", historyPayload["findings"])
	}
}

func TestVerifyDeterminismCommandJSON(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	storePath := filepath.Join(t.TempDir(), "faultline.db")
	logPath := writeTempLog(t, "pull access denied\nError response from daemon: authentication required\n")

	runAnalyze := func() {
		cmd := newRootCommand()
		cmd.SetArgs([]string{"analyze", "--json", "--store", storePath, logPath})
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(new(bytes.Buffer))
		t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("execute analyze for determinism setup: %v", err)
		}
	}
	runAnalyze()
	runAnalyze()

	cmd := newRootCommand()
	cmd.SetArgs([]string{"verify-determinism", "--json", "--store", storePath, logPath})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute verify-determinism --json: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &payload); err != nil {
		t.Fatalf("unmarshal determinism JSON: %v", err)
	}
	determinism, ok := payload["determinism"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected determinism object, got %v", payload["determinism"])
	}
	if determinism["stable"] != true {
		t.Fatalf("expected stable determinism summary, got %v", determinism)
	}
	if determinism["run_count"] != float64(2) {
		t.Fatalf("expected run_count=2, got %v", determinism["run_count"])
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
