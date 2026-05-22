package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFixturesCommandRegistersHiddenSubcommands(t *testing.T) {
	cmd := newFixturesCommand()
	if !cmd.Hidden {
		t.Fatal("fixtures command should be hidden")
	}

	want := map[string]bool{
		"ingest":        false,
		"review":        false,
		"promote":       false,
		"stats":         false,
		"sanitize":      false,
		"compare-modes": false,
	}
	for _, child := range cmd.Commands() {
		if _, ok := want[child.Name()]; ok {
			want[child.Name()] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("missing subcommand %q", name)
		}
	}
}

func TestFixturesIngestCommandValidatesRequiredFlags(t *testing.T) {
	t.Run("adapter is required", func(t *testing.T) {
		cmd := newFixturesIngestCommand()
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))

		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "--adapter is required") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("url is required", func(t *testing.T) {
		cmd := newFixturesIngestCommand()
		cmd.SetArgs([]string{"--adapter", "github-issue"})
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))

		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "at least one --url is required") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestFixturesPromoteCommandRequiresExpectedPlaybook(t *testing.T) {
	cmd := newFixturesPromoteCommand()
	cmd.SetArgs([]string{"fixture-123"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--expected-playbook is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFixturesStatsCommandRejectsInvalidClass(t *testing.T) {
	cmd := newFixturesStatsCommand()
	cmd.SetArgs([]string{"--class", "bogus"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid fixture class") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFixturesSanitizeCommandRequiresArgs(t *testing.T) {
	cmd := newFixturesSanitizeCommand()
	cmd.SetArgs([]string{})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no staging IDs are provided")
	}
}

// setBundledPlaybookDir points FAULTLINE_PLAYBOOK_DIR at the real bundled
// catalog for tests that need a working playbook catalog. When running inside
// the internal/cli package directory, DefaultDir() may incorrectly resolve to
// internal/playbooks (the Go source package) before finding playbooks/bundled.
func setBundledPlaybookDir(t *testing.T) {
	t.Helper()
	bundledDir, err := filepath.Abs("../../playbooks/bundled")
	if err != nil {
		t.Fatalf("resolve bundled playbook dir: %v", err)
	}
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", bundledDir)
}

// ── newFixturesPatternsCommand ────────────────────────────────────────────────

func TestFixturesPatternsCommandRejectsMissingBaseline(t *testing.T) {
	dir := t.TempDir()
	cmd := newFixturesPatternsCommand()
	cmd.SetArgs([]string{"--baseline", dir + "/nonexistent-baseline.txt"})
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when baseline file is missing")
	}
}

func TestFixturesPatternsCommandUpdateBaseline(t *testing.T) {
	setBundledPlaybookDir(t)
	dir := t.TempDir()
	baseline := dir + "/baseline.txt"

	cmd := newFixturesPatternsCommand()
	cmd.SetArgs([]string{"--baseline", baseline, "--update-baseline"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("update-baseline: %v", err)
	}
	if !strings.Contains(out.String(), "updated playbook review baseline") {
		t.Errorf("expected success message, got: %s", out.String())
	}
	// Baseline file should now exist
	if _, err := os.Stat(baseline); err != nil {
		t.Errorf("expected baseline file to be created, got: %v", err)
	}
}

func TestFixturesPatternsCommandMatchesOwnBaseline(t *testing.T) {
	setBundledPlaybookDir(t)
	// First update a baseline, then verify it passes
	dir := t.TempDir()
	baseline := dir + "/baseline.txt"

	updateCmd := newFixturesPatternsCommand()
	updateCmd.SetArgs([]string{"--baseline", baseline, "--update-baseline"})
	updateCmd.SetOut(new(bytes.Buffer))
	updateCmd.SetErr(new(bytes.Buffer))
	if err := updateCmd.Execute(); err != nil {
		t.Fatalf("update-baseline: %v", err)
	}

	checkCmd := newFixturesPatternsCommand()
	checkCmd.SetArgs([]string{"--baseline", baseline})
	var out bytes.Buffer
	checkCmd.SetOut(&out)
	checkCmd.SetErr(new(bytes.Buffer))
	if err := checkCmd.Execute(); err != nil {
		t.Fatalf("patterns check against own baseline: %v\noutput: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "playbook review passed") {
		t.Errorf("expected pass message, got: %s", out.String())
	}
}

func TestFixturesPatternsCommandVerbosePrintsReport(t *testing.T) {
	setBundledPlaybookDir(t)
	dir := t.TempDir()
	baseline := dir + "/baseline.txt"

	updateCmd := newFixturesPatternsCommand()
	updateCmd.SetArgs([]string{"--baseline", baseline, "--update-baseline"})
	updateCmd.SetOut(new(bytes.Buffer))
	updateCmd.SetErr(new(bytes.Buffer))
	if err := updateCmd.Execute(); err != nil {
		t.Fatalf("update-baseline: %v", err)
	}

	checkCmd := newFixturesPatternsCommand()
	checkCmd.SetArgs([]string{"--baseline", baseline, "--verbose"})
	var out bytes.Buffer
	checkCmd.SetOut(&out)
	checkCmd.SetErr(new(bytes.Buffer))
	if err := checkCmd.Execute(); err != nil {
		t.Fatalf("patterns --verbose: %v\noutput: %s", err, out.String())
	}
	// verbose should output the full report even when passing
	// (output may be empty if there are no conflicts, that's valid)
}

// ── newFixturesPackCheckCommand ───────────────────────────────────────────────

func TestFixturesPackCheckCommandRequiresPack(t *testing.T) {
	cmd := newFixturesPackCheckCommand()
	cmd.SetArgs([]string{})
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--pack") {
		t.Fatalf("expected --pack required error, got: %v", err)
	}
}

func TestFixturesPackCheckCommandInvalidPackPath(t *testing.T) {
	cmd := newFixturesPackCheckCommand()
	cmd.SetArgs([]string{"--pack", "/nonexistent/path/to/pack"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent pack path")
	}
}

func TestFixturesPackCheckCommandWithRealPack(t *testing.T) {
	setBundledPlaybookDir(t)

	// Create a minimal extra pack with a single unique playbook so it can
	// be composed alongside the bundled catalog without duplicate-ID errors.
	packDir := t.TempDir()
	playbookYAML := `id: test-pack-check-unique
title: Test pack-check unique playbook
category: test
severity: low
base_score: 0.5
tags: [test]
summary: Placeholder for pack-check integration test.
diagnosis: placeholder
fix: placeholder
validation: placeholder
match:
  any:
    - pack-check-unique-sentinel-pattern
`
	if err := os.WriteFile(filepath.Join(packDir, "test-pack-check.yaml"), []byte(playbookYAML), 0o600); err != nil {
		t.Fatalf("write temp playbook: %v", err)
	}

	cmd := newFixturesPackCheckCommand()
	cmd.SetArgs([]string{"--pack", packDir})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("pack-check with extra pack: %v\noutput: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "loaded") {
		t.Errorf("expected 'loaded' in output, got: %s", out.String())
	}
}

func TestFixturesPackCheckCommandWithReview(t *testing.T) {
	setBundledPlaybookDir(t)

	packDir := t.TempDir()
	playbookYAML := `id: test-pack-review-unique
title: Test pack-review unique playbook
category: test
severity: low
base_score: 0.5
tags: [test]
summary: Placeholder for pack-check review integration test.
diagnosis: placeholder
fix: placeholder
validation: placeholder
match:
  any:
    - pack-review-unique-sentinel-pattern
`
	if err := os.WriteFile(filepath.Join(packDir, "test-pack-review.yaml"), []byte(playbookYAML), 0o600); err != nil {
		t.Fatalf("write temp playbook: %v", err)
	}

	cmd := newFixturesPackCheckCommand()
	cmd.SetArgs([]string{"--pack", packDir, "--review"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("pack-check --review: %v\noutput: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "pattern conflicts") {
		t.Errorf("expected 'pattern conflicts' in output, got: %s", out.String())
	}
}
