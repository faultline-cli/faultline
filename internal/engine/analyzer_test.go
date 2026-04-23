package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadLinesNormal(t *testing.T) {
	lines, err := ReadLines(strings.NewReader("line one\nline two\nline three\n"))
	if err != nil {
		t.Fatalf("ReadLines: %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %#v", len(lines), lines)
	}
	if lines[0].Original != "line one" {
		t.Errorf("expected first line 'line one', got %q", lines[0].Original)
	}
	if lines[2].Original != "line three" {
		t.Errorf("expected third line 'line three', got %q", lines[2].Original)
	}
}

func TestReadLinesSkipsBlanks(t *testing.T) {
	lines, err := ReadLines(strings.NewReader("\n\nline one\n\n  \nline two\n\n"))
	if err != nil {
		t.Fatalf("ReadLines: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 non-blank lines, got %d: %#v", len(lines), lines)
	}
}

func TestReadLinesEmpty(t *testing.T) {
	lines, err := ReadLines(strings.NewReader(""))
	if err != nil {
		t.Fatalf("ReadLines empty: %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(lines))
	}
}

func TestReadLinesPreservesLineNumbers(t *testing.T) {
	lines, err := ReadLines(strings.NewReader("first\nsecond\nthird\n"))
	if err != nil {
		t.Fatalf("ReadLines: %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	for i, line := range lines {
		if line.Number == 0 {
			t.Errorf("line[%d] has zero line number", i)
		}
	}
}

func TestAnalyzeReaderFindsBestMatch(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	a, err := e.AnalyzeReader(strings.NewReader(
		"fatal: could not read Username for 'https://github.com': terminal prompts disabled\n",
	))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(a.Results) == 0 {
		t.Fatal("expected at least one result")
	}
	if a.Results[0].Playbook.ID != "git-auth" {
		t.Fatalf("expected git-auth, got %s", a.Results[0].Playbook.ID)
	}
}

func TestAnalyzeReaderReturnsNoMatch(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	_, err := e.AnalyzeReader(strings.NewReader("all checks passed\n"))
	if err != ErrNoMatch {
		t.Fatalf("expected ErrNoMatch, got %v", err)
	}
}

func TestAnalyzeReaderMultipleResults(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	log := "pull access denied\nauthentication required\ncould not read username for 'https://github.com': terminal prompts disabled\n"
	a, err := e.AnalyzeReader(strings.NewReader(log))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(a.Results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(a.Results))
	}
}

func TestAnalyzeReaderHypothesisPrefersDependencyDriftWhenCacheRestoreIsAbsent(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	log := "checksum mismatch\nfailed to resolve dependencies\n"
	a, err := e.AnalyzeReader(strings.NewReader(log))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(a.Results) < 2 {
		t.Fatalf("expected competing results, got %#v", a.Results)
	}
	if a.Results[0].Playbook.ID != "dependency-drift" {
		t.Fatalf("expected dependency-drift to win differential ranking, got %s", a.Results[0].Playbook.ID)
	}
	if a.Differential == nil || len(a.Differential.Alternatives) == 0 {
		t.Fatalf("expected populated differential diagnosis, got %#v", a.Differential)
	}
	alternative := a.Differential.Alternatives[0]
	if alternative.FailureID != "cache-corruption" {
		t.Fatalf("expected cache-corruption as the main alternative, got %#v", alternative)
	}
	if len(alternative.WhyLessLikely) == 0 || !strings.Contains(strings.ToLower(strings.Join(alternative.WhyLessLikely, " ")), "cache restore") {
		t.Fatalf("expected alternative to explain missing cache restore evidence, got %#v", alternative.WhyLessLikely)
	}
}

func TestAnalyzeReaderHypothesisRulesOutDependencyDriftOnHashMismatch(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	log := "THESE PACKAGES DO NOT MATCH THE HASHES FROM THE REQUIREMENTS FILE\nfailed to resolve dependencies\n"
	a, err := e.AnalyzeReader(strings.NewReader(log))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(a.Results) == 0 {
		t.Fatal("expected at least one result")
	}
	if a.Results[0].Playbook.ID != "pip-hash-mismatch" {
		t.Fatalf("expected pip-hash-mismatch to win, got %s", a.Results[0].Playbook.ID)
	}
	if a.Differential == nil || len(a.Differential.RuledOut) == 0 {
		t.Fatalf("expected ruled-out rival in differential diagnosis, got %#v", a.Differential)
	}
	found := false
	for _, item := range a.Differential.RuledOut {
		if item.FailureID == "dependency-drift" {
			found = true
			if len(item.RuledOutBy) == 0 {
				t.Fatalf("expected ruled-out reason for dependency-drift, got %#v", item)
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected dependency-drift to be ruled out, got %#v", a.Differential.RuledOut)
	}
}

func TestAnalyzeReaderDockerAuthDoesNotMatchGenericPermissionDenied(t *testing.T) {
	e := New(Options{
		PlaybookDir:  repoPlaybookDir(t),
		NoHistory:    true,
		BayesEnabled: true,
		GitSince:     "30d",
		RepoPath:     t.TempDir(),
	})

	log := "Error response from daemon: pull access denied for mcr/microsoft.com/mssql/server, repository does not exist or may require 'docker login'\n"
	a, err := e.AnalyzeReader(strings.NewReader(log))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(a.Results) == 0 {
		t.Fatal("expected at least one result")
	}
	if a.Results[0].Playbook.ID != "docker-auth" {
		t.Fatalf("expected docker-auth, got %s", a.Results[0].Playbook.ID)
	}
	for _, result := range a.Results {
		if result.Playbook.ID == "permission-denied" {
			t.Fatalf("permission-denied should be excluded for registry auth failures: %#v", a.Results)
		}
	}
}

func TestAnalyzeReaderBayesDoesNotAttachDeltaWithoutGitContext(t *testing.T) {
	e := New(Options{
		PlaybookDir:  repoPlaybookDir(t),
		NoHistory:    true,
		BayesEnabled: true,
	})

	a, err := e.AnalyzeReader(strings.NewReader("Go Version: go1.26.0\n"))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if a.Delta != nil {
		t.Fatalf("expected no delta without git context, got %#v", a.Delta)
	}
	if a.RepoContext != nil {
		t.Fatalf("expected no repo context without git context, got %#v", a.RepoContext)
	}
	if len(a.Results) == 0 || a.Results[0].Ranking == nil {
		t.Fatalf("expected bayes ranking without git context, got %#v", a.Results)
	}
}

func TestAnalyzeReaderOOMKilled(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	a, err := e.AnalyzeReader(strings.NewReader("Process exited with exit code 137\nout of memory\n"))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if a.Results[0].Playbook.ID != "oom-killed" {
		t.Fatalf("expected oom-killed, got %s", a.Results[0].Playbook.ID)
	}
}

func TestAnalyzeReaderYarnLockfile(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	a, err := e.AnalyzeReader(strings.NewReader(
		"Your lockfile needs to be updated, but yarn was run with `--frozen-lockfile`.\n",
	))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if a.Results[0].Playbook.ID != "yarn-lockfile" {
		t.Fatalf("expected yarn-lockfile, got %s", a.Results[0].Playbook.ID)
	}
}

func TestAnalyzeReaderContextExtracted(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	log := "$ docker push ghcr.io/example/app\npull access denied\n"
	a, err := e.AnalyzeReader(strings.NewReader(log))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if a.Context.CommandHint == "" {
		t.Error("expected a command hint to be extracted")
	}
}

func TestAnalyzeReaderDockerBuildContext(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	a, err := e.AnalyzeReader(strings.NewReader(
		"failed to solve with frontend dockerfile.v0: failed to read Dockerfile: open Dockerfile: no such file or directory\n",
	))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if a.Results[0].Playbook.ID != "docker-build-context" {
		t.Fatalf("expected docker-build-context, got %s", a.Results[0].Playbook.ID)
	}
}

func TestAnalyzeReaderImagePullBackoff(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	a, err := e.AnalyzeReader(strings.NewReader(
		"Warning Failed pod/app-123 Failed to pull image \"ghcr.io/acme/app:missing\": manifest unknown\nBack-off pulling image\n",
	))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if a.Results[0].Playbook.ID != "image-pull-backoff" {
		t.Fatalf("expected image-pull-backoff, got %s", a.Results[0].Playbook.ID)
	}
}

func TestAnalyzeReaderSnapshotMismatch(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	a, err := e.AnalyzeReader(strings.NewReader(
		"Received value does not match stored snapshot\nRun with -u to update snapshots\n",
	))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if a.Results[0].Playbook.ID != "snapshot-mismatch" {
		t.Fatalf("expected snapshot-mismatch, got %s", a.Results[0].Playbook.ID)
	}
}

func TestAnalyzeReaderWithAdditionalPack(t *testing.T) {
	extra := t.TempDir()
	if err := os.WriteFile(filepath.Join(extra, "custom.yaml"), []byte(`
id: extra-custom
title: Extra Custom
category: test
severity: low
summary: |
  Custom summary.
diagnosis: |
  ## Diagnosis

  Custom diagnosis.
fix: |
  ## Fix steps

  1. Custom fix.
validation: |
  ## Validation

  - Custom validation.
match:
  any:
    - "totally custom failure marker"
`), 0o600); err != nil {
		t.Fatalf("write custom pack: %v", err)
	}

	e := New(Options{
		PlaybookDir:      "",
		PlaybookPackDirs: []string{extra},
		NoHistory:        true,
	})
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", repoPlaybookDir(t))

	a, err := e.AnalyzeReader(strings.NewReader("totally custom failure marker\n"))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if a.Results[0].Playbook.ID != "extra-custom" {
		t.Fatalf("expected extra-custom, got %s", a.Results[0].Playbook.ID)
	}
	if a.Results[0].Playbook.Metadata.PackName != filepath.Base(extra) {
		t.Fatalf("expected pack metadata %q, got %#v", filepath.Base(extra), a.Results[0].Playbook.Metadata)
	}
}

func TestAnalyzePathFindsBestMatch(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "test-*.log")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString("fatal: could not read Username for 'https://github.com': terminal prompts disabled\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	a, err := e.AnalyzePath(f.Name())
	if err != nil {
		t.Fatalf("AnalyzePath: %v", err)
	}
	if len(a.Results) == 0 {
		t.Fatal("expected at least one result")
	}
	if a.Results[0].Playbook.ID != "git-auth" {
		t.Fatalf("expected git-auth, got %s", a.Results[0].Playbook.ID)
	}
	if a.Source != f.Name() {
		t.Errorf("expected source=%s, got %s", f.Name(), a.Source)
	}
}

func TestAnalyzePathMissingFileReturnsError(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	_, err := e.AnalyzePath("/nonexistent/does-not-exist.log")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestEngineListReturnsPlaybooks(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	pbs, err := e.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(pbs) == 0 {
		t.Fatal("expected non-empty playbook list")
	}
}

func TestEngineExplainKnownPlaybook(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	pb, err := e.Explain("git-auth")
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	if pb.ID != "git-auth" {
		t.Errorf("expected git-auth, got %s", pb.ID)
	}
}

func TestEngineExplainUnknownReturnsError(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	_, err := e.Explain("totally-unknown-playbook-xyz")
	if err == nil {
		t.Fatal("expected error for unknown playbook ID")
	}
}

func TestLooksLikeSourceFile(t *testing.T) {
	wantTrue := []string{
		"main.go", "app.js", "component.jsx", "types.ts", "page.tsx",
		"script.py", "Main.java", "helper.rb", "controller.php",
		"service.cs", "activity.kt", "config.yaml", "config.yml",
	}
	for _, path := range wantTrue {
		if !looksLikeSourceFile(path) {
			t.Errorf("looksLikeSourceFile(%q) = false, want true", path)
		}
	}
	wantFalse := []string{
		"README.md", "binary.exe", "archive.zip", "image.png",
		"Makefile", ".gitignore", "data.json", "notes.txt",
	}
	for _, path := range wantFalse {
		if looksLikeSourceFile(path) {
			t.Errorf("looksLikeSourceFile(%q) = true, want false", path)
		}
	}
}

func TestLoadSourceFilesScansDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Readme\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.ts"), []byte("const x = 1;\n"), 0o644); err != nil {
		t.Fatalf("write app.ts: %v", err)
	}

	files, err := loadSourceFiles(dir)
	if err != nil {
		t.Fatalf("loadSourceFiles: %v", err)
	}
	paths := make(map[string]struct{}, len(files))
	for _, f := range files {
		paths[f.Path] = struct{}{}
	}
	if _, ok := paths["main.go"]; !ok {
		t.Error("expected main.go in source files")
	}
	if _, ok := paths["app.ts"]; !ok {
		t.Error("expected app.ts in source files")
	}
	if _, ok := paths["README.md"]; ok {
		t.Error("README.md should not be in source files")
	}
}

func TestLoadSourceFilesSkipsGitDir(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "config.go"), []byte("package git\n"), 0o644); err != nil {
		t.Fatalf("write config.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write app.go: %v", err)
	}

	files, err := loadSourceFiles(dir)
	if err != nil {
		t.Fatalf("loadSourceFiles: %v", err)
	}
	for _, f := range files {
		if strings.Contains(f.Path, ".git") {
			t.Errorf("expected .git directory to be skipped, got %q", f.Path)
		}
	}
}

func TestLoadSourceFilesSkipsVirtualEnvDir(t *testing.T) {
	dir := t.TempDir()
	venvDir := filepath.Join(dir, ".venv", "lib", "python3.13", "site-packages", "pip")
	if err := os.MkdirAll(venvDir, 0o755); err != nil {
		t.Fatalf("mkdir .venv: %v", err)
	}
	if err := os.WriteFile(filepath.Join(venvDir, "service.go"), []byte("package pip\n"), 0o644); err != nil {
		t.Fatalf("write service.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write app.go: %v", err)
	}

	files, err := loadSourceFiles(dir)
	if err != nil {
		t.Fatalf("loadSourceFiles: %v", err)
	}
	for _, f := range files {
		if strings.Contains(f.Path, ".venv/") {
			t.Errorf("expected .venv directory to be skipped, got %q", f.Path)
		}
	}
}

func TestAnalyzeReaderEmptyInput(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	_, err := e.AnalyzeReader(strings.NewReader(""))
	if err != ErrNoInput {
		t.Fatalf("expected ErrNoInput, got %v", err)
	}
}

func repoPlaybookDir(_ testing.TB) string {
	return "../../playbooks/bundled"
}

func repoExtraPackDir(_ testing.TB) string {
	return "../../playbooks/packs/extra-local"
}

func requireExtraPack(t testing.TB) string {
	t.Helper()
	for _, path := range []string{
		repoExtraPackDir(t),
		"../../../faultline-extra-pack",
	} {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			return resolved
		}
		if abs, err := filepath.Abs(path); err == nil {
			return abs
		}
		return path
	}
	t.Skip("extra pack repository is not available locally")
	return ""
}
