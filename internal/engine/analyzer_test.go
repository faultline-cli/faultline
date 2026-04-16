package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
