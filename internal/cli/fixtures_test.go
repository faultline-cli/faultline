package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadScaffoldLogFromLogFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "build.log")
	const want = "pull access denied\n"
	if err := os.WriteFile(logPath, []byte(want), 0o644); err != nil {
		t.Fatalf("write log file: %v", err)
	}

	got, err := readScaffoldLog("", logPath, "", nil, strings.NewReader("stdin"))
	if err != nil {
		t.Fatalf("readScaffoldLog: %v", err)
	}
	if got != want {
		t.Fatalf("readScaffoldLog = %q, want %q", got, want)
	}
}

func TestReadScaffoldLogFromPositionalArg(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "build.log")
	const want = "permission denied\n"
	if err := os.WriteFile(logPath, []byte(want), 0o644); err != nil {
		t.Fatalf("write log file: %v", err)
	}

	got, err := readScaffoldLog("", "", "", []string{logPath}, strings.NewReader("stdin"))
	if err != nil {
		t.Fatalf("readScaffoldLog: %v", err)
	}
	if got != want {
		t.Fatalf("readScaffoldLog = %q, want %q", got, want)
	}
}

func TestReadScaffoldLogFromStdin(t *testing.T) {
	const want = "exec /__e/node20/bin/node: no such file or directory\n"

	got, err := readScaffoldLog("", "", "", nil, strings.NewReader(want))
	if err != nil {
		t.Fatalf("readScaffoldLog: %v", err)
	}
	if got != want {
		t.Fatalf("readScaffoldLog = %q, want %q", got, want)
	}
}

func TestReadScaffoldLogRejectsMultipleSources(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "build.log")
	if err := os.WriteFile(logPath, []byte("duplicate source"), 0o644); err != nil {
		t.Fatalf("write log file: %v", err)
	}

	_, err := readScaffoldLog("", logPath, "", []string{logPath}, strings.NewReader("stdin"))
	if err == nil {
		t.Fatal("expected multiple input sources error")
	}
	if !strings.Contains(err.Error(), "choose exactly one log source") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadStagingFixtureLogUsesRawLogFirst(t *testing.T) {
	root := t.TempDir()
	writeStagingFixture(t, root, "fixture-raw", `
id: fixture-raw
raw_log: |
  original secret-bearing log
normalized_log: |
  normalized fallback log
`)

	got, err := loadStagingFixtureLog(root, "fixture-raw")
	if err != nil {
		t.Fatalf("loadStagingFixtureLog: %v", err)
	}
	if !strings.Contains(got, "original secret-bearing log") {
		t.Fatalf("expected raw_log to be preferred, got %q", got)
	}
}

func TestLoadStagingFixtureLogFallsBackToNormalizedLog(t *testing.T) {
	root := t.TempDir()
	writeStagingFixture(t, root, "fixture-normalized", `
id: fixture-normalized
normalized_log: |
  canonical failure line
`)

	got, err := loadStagingFixtureLog(root, "fixture-normalized")
	if err != nil {
		t.Fatalf("loadStagingFixtureLog: %v", err)
	}
	if !strings.Contains(got, "canonical failure line") {
		t.Fatalf("expected normalized_log fallback, got %q", got)
	}
}

func TestLoadStagingFixtureLogMissingFixture(t *testing.T) {
	root := t.TempDir()

	_, err := loadStagingFixtureLog(root, "missing-fixture")
	if err == nil {
		t.Fatal("expected fixture lookup error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeStagingFixture(t *testing.T, root, name, content string) {
	t.Helper()
	dir := filepath.Join(root, "fixtures", "staging")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir staging fixtures: %v", err)
	}
	path := filepath.Join(dir, name+".yaml")
	if err := os.WriteFile(path, []byte(strings.TrimLeft(content, "\n")), 0o644); err != nil {
		t.Fatalf("write staging fixture: %v", err)
	}
}
