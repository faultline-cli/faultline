package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGuardCommandRejectsInvalidMode(t *testing.T) {
	cmd := newGuardCommand()
	cmd.SetArgs([]string{"--mode", "verbose"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid mode error")
	}
}

func TestGuardCommandRejectsJSONMarkdownCombination(t *testing.T) {
	cmd := newGuardCommand()
	cmd.SetArgs([]string{"--format", "markdown", "--json"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected json/markdown conflict")
	}
}

func TestFixCommandRejectsInvalidFormat(t *testing.T) {
	cmd := newFixCommand()
	cmd.SetArgs([]string{"--format", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid format error")
	}
}

func TestFixCommandRunsWithBundledPlaybooks(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "build.log")
	logText := "npm ERR! code EUSAGE\nnpm ERR! npm ci can only install packages when your package.json and package-lock.json are in sync.\n"
	if err := os.WriteFile(logPath, []byte(logText), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	cmd := newFixCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--playbooks", cliTestPlaybookDir(),
		"--no-store",
		logPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("fix command: %v\noutput: %s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "npm-ci-lockfile") {
		t.Fatalf("expected fix output to mention npm-ci-lockfile, got:\n%s", out)
	}
}

func TestTraceCommandRejectsNegativeSelect(t *testing.T) {
	cmd := newTraceCommand()
	cmd.SetArgs([]string{"--select", "-1"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected negative select error")
	}
}

func TestTraceCommandRejectsInvalidHookMode(t *testing.T) {
	cmd := newTraceCommand()
	cmd.SetArgs([]string{"--hooks", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid hook mode error")
	}
}

func TestTraceCommandRunsWithBundledPlaybooks(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "build.log")
	logText := "fatal: could not read Username for 'https://github.com': terminal prompts disabled\n"
	if err := os.WriteFile(logPath, []byte(logText), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	cmd := newTraceCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--playbooks", cliTestPlaybookDir(),
		"--no-store",
		logPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("trace command: %v\noutput: %s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "TRACE  git-auth") {
		t.Fatalf("expected trace output to mention git-auth, got:\n%s", out)
	}
}
