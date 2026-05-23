package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func cliTestPlaybookDir() string {
	return "../../playbooks/bundled"
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
