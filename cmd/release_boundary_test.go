package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeHelpOmitsExperimentalDeltaFlags(t *testing.T) {
	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--help"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze --help: %v", err)
	}
	if strings.Contains(out.String(), "delta-provider") {
		t.Fatalf("expected analyze help to omit hidden experimental delta flags, got %q", out.String())
	}
}

func TestAnalyzeExperimentalDeltaRequiresExplicitOptIn(t *testing.T) {
	playbookDir, err := filepath.Abs("../playbooks/bundled")
	if err != nil {
		t.Fatalf("abs playbook dir: %v", err)
	}
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)
	t.Setenv("FAULTLINE_EXPERIMENTAL_GITHUB_DELTA", "")

	cmd := newRootCommand()
	cmd.SetArgs([]string{"analyze", "--delta-provider", "github-actions", "--no-history", "../examples/docker-auth.log"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err = cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "FAULTLINE_EXPERIMENTAL_GITHUB_DELTA") {
		t.Fatalf("expected experimental opt-in error, got %v", err)
	}
}

func TestExampleMarkdownSnapshots(t *testing.T) {
	playbookDir, err := filepath.Abs("../playbooks/bundled")
	if err != nil {
		t.Fatalf("abs playbook dir: %v", err)
	}
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)
	repoRoot := filepath.Clean("..")

	cases := []struct {
		logPath      string
		expectedPath string
	}{
		{logPath: "examples/docker-auth.log", expectedPath: "examples/docker-auth.expected.md"},
		{logPath: "examples/missing-executable.log", expectedPath: "examples/missing-executable.expected.md"},
		{logPath: "examples/runtime-mismatch.log", expectedPath: "examples/runtime-mismatch.expected.md"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(filepath.Base(tc.logPath), func(t *testing.T) {
			out := runRootCommand(t, repoRoot, "analyze", "--format", "markdown", "--no-history", tc.logPath)

			want, readErr := os.ReadFile(filepath.Join(repoRoot, tc.expectedPath))
			if readErr != nil {
				t.Fatalf("read expected output %s: %v", tc.expectedPath, readErr)
			}
			if out != string(want) {
				t.Fatalf("snapshot mismatch for %s", tc.logPath)
			}
		})
	}
}

func TestExampleWorkflowSnapshots(t *testing.T) {
	playbookDir, err := filepath.Abs("../playbooks/bundled")
	if err != nil {
		t.Fatalf("abs playbook dir: %v", err)
	}
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)
	repoRoot := filepath.Clean("..")

	logData, err := os.ReadFile(filepath.Join(repoRoot, "examples/missing-executable.log"))
	if err != nil {
		t.Fatalf("read example log: %v", err)
	}

	t.Run("local text", func(t *testing.T) {
		got := runRootCommandWithStdin(t, repoRoot, string(logData), "workflow", "--no-history")
		want, readErr := os.ReadFile(filepath.Join(repoRoot, "examples/missing-executable.workflow.local.txt"))
		if readErr != nil {
			t.Fatalf("read local workflow snapshot: %v", readErr)
		}
		if got != string(want) {
			t.Fatal("local workflow snapshot mismatch")
		}
	})

	t.Run("agent json", func(t *testing.T) {
		got := runRootCommandWithStdin(t, repoRoot, string(logData), "workflow", "--json", "--mode", "agent", "--no-history")
		want, readErr := os.ReadFile(filepath.Join(repoRoot, "examples/missing-executable.workflow.agent.json"))
		if readErr != nil {
			t.Fatalf("read agent workflow snapshot: %v", readErr)
		}
		if got != string(want) {
			t.Fatal("agent workflow snapshot mismatch")
		}
	})
}

func runRootCommand(t *testing.T, workdir string, args ...string) string {
	t.Helper()

	restore := chdirForTest(t, workdir)
	defer restore()

	cmd := newRootCommand()
	cmd.SetArgs(args)
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute %v: %v", args, err)
	}
	return out.String()
}

func runRootCommandWithStdin(t *testing.T, workdir, input string, args ...string) string {
	t.Helper()

	restore := chdirForTest(t, workdir)
	defer restore()

	oldStdin := os.Stdin
	t.Cleanup(func() { os.Stdin = oldStdin })

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdin pipe: %v", err)
	}
	if _, err := writer.WriteString(input); err != nil {
		t.Fatalf("write stdin pipe: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdin writer: %v", err)
	}
	os.Stdin = reader

	cmd := newRootCommand()
	cmd.SetArgs(args)
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute %v: %v", args, err)
	}
	return out.String()
}

func chdirForTest(t *testing.T, dir string) func() {
	t.Helper()

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	return func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Fatalf("restore wd %s: %v", oldWD, err)
		}
	}
}
