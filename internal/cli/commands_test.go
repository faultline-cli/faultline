package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/app"
	"faultline/internal/output"
)

// cliTestPlaybookDir returns the canonical bundled playbook directory relative
// to this package. Tests in the cli package run from internal/cli, so we walk
// up two levels.
func cliTestPlaybookDir() string {
	return "../../playbooks/bundled"
}

// buildAnalysisArtifact runs svc.Analyze on the given log text and returns
// the JSON artifact as a string. It uses the bundled playbook directory
// relative to this package (two levels up).
func buildAnalysisArtifact(t *testing.T, log string) string {
	t.Helper()
	svc := app.NewService()
	opts := app.AnalyzeOptions{
		OutputOptions: app.OutputOptions{
			Top:    3,
			Mode:   output.ModeQuick,
			Format: output.FormatJSON,
			JSON:   true,
		},
		Store:       "off",
		PlaybookDir: cliTestPlaybookDir(),
	}
	var buf bytes.Buffer
	if err := svc.Analyze(context.Background(), strings.NewReader(log), "stdin", opts, &buf); err != nil {
		t.Fatalf("buildAnalysisArtifact: %v", err)
	}
	return strings.TrimSpace(buf.String())
}

// ── newReplayCommand ──────────────────────────────────────────────────────────

func TestReplayCommandRejectsWrongArgCount(t *testing.T) {
	cmd := newReplayCommand()
	if err := cmd.Args(cmd, nil); err == nil {
		t.Fatal("expected error for zero args, got nil")
	}
	if err := cmd.Args(cmd, []string{"a", "b"}); err == nil {
		t.Fatal("expected error for two args, got nil")
	}
	if err := cmd.Args(cmd, []string{"only-one"}); err != nil {
		t.Fatalf("expected no error for exactly 1 arg, got %v", err)
	}
}

func TestReplayCommandRunsWithArtifactFile(t *testing.T) {
	log := "pull access denied\nError response from daemon: authentication required\n"
	artifact := buildAnalysisArtifact(t, log)

	dir := t.TempDir()
	artifactFile := filepath.Join(dir, "analysis.json")
	if err := os.WriteFile(artifactFile, []byte(artifact), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	cmd := newReplayCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{artifactFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("replay command: %v\noutput: %s", err, buf.String())
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty replay output")
	}
}

func TestReplayCommandMarkdownOutput(t *testing.T) {
	log := "pull access denied\nError response from daemon: authentication required\n"
	artifact := buildAnalysisArtifact(t, log)

	dir := t.TempDir()
	artifactFile := filepath.Join(dir, "analysis.json")
	if err := os.WriteFile(artifactFile, []byte(artifact), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	cmd := newReplayCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--format", "markdown", artifactFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("replay command --format markdown: %v\noutput: %s", err, buf.String())
	}
	if !strings.HasPrefix(buf.String(), "#") {
		t.Errorf("expected markdown heading in replay output, got %q", buf.String()[:min(80, buf.Len())])
	}
}

func TestReplayCommandRejectsInvalidView(t *testing.T) {
	dir := t.TempDir()
	artifactFile := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(artifactFile, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	cmd := newReplayCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--view", "notaview", artifactFile})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for invalid --view, got nil")
	}
}

// ── newPacksListCommand ───────────────────────────────────────────────────────

func TestPacksListCommandRunsWithNoInstalledPacks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cmd := newPacksListCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("packs list: %v", err)
	}
	if !strings.Contains(buf.String(), "No installed") {
		t.Errorf("expected 'No installed' message, got %q", buf.String())
	}
}
