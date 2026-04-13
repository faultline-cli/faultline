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
diagnosis_markdown: |
  ## Diagnosis

  Custom diagnosis.
fix_markdown: |
  ## Fix steps

  1. Custom fix.
validation_markdown: |
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

func repoPlaybookDir(_ testing.TB) string {
	return "../../playbooks/bundled"
}

func repoPremiumPackDir(_ testing.TB) string {
	return "../../playbooks/packs/premium-local"
}

func requirePremiumPack(t testing.TB) string {
	t.Helper()
	for _, path := range []string{
		repoPremiumPackDir(t),
		"../../../faultline-premium",
		"../../../faultline-premium-pack",
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
	t.Skip("premium pack repository is not available locally")
	return ""
}
