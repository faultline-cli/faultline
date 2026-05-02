package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCoverageCommandIsNotHidden(t *testing.T) {
	cmd := newCoverageCommand()
	if cmd.Hidden {
		t.Fatal("coverage command must not be hidden")
	}
}

func TestNewCoverageCommandRejectsPositionalArgs(t *testing.T) {
	cmd := newCoverageCommand()
	if err := cmd.Args(cmd, []string{"unexpected"}); err == nil {
		t.Fatal("expected error for positional args, got nil")
	}
}

func TestCoverageCommandUsesFixtureExpectations(t *testing.T) {
	pbDir := t.TempDir()
	fixtureRoot := filepath.Join(t.TempDir(), "fixtures")
	if err := os.MkdirAll(filepath.Join(fixtureRoot, "minimal"), 0o755); err != nil {
		t.Fatalf("create fixtures: %v", err)
	}
	writeCoveragePlaybook(t, pbDir, "npm-registry-auth", "auth")
	writeCoveragePlaybook(t, pbDir, "install-failure", "build")
	writeCoveragePlaybook(t, pbDir, "unused-playbook", "build")
	writeCoverageFixture(t, filepath.Join(fixtureRoot, "minimal", "npm-registry-auth.yaml"), `
id: npm-registry-auth-canonical
fixture_class: minimal
normalized_log: |
  npm ERR! code E401
expectation:
  expected_playbook: npm-registry-auth
  strict_top1: true
`)
	writeCoverageFixture(t, filepath.Join(fixtureRoot, "minimal", "npm-registry-auth-near-miss.yaml"), `
id: npm-registry-auth-near-miss
fixture_class: minimal
normalized_log: |
  npm ERR! code E404
expectation:
  expected_playbook: install-failure
  disallowed_playbooks:
    - npm-registry-auth
`)

	cmd := newCoverageCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--playbooks", pbDir,
		"--fixture-dir", fixtureRoot,
		"--json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("coverage command --json: %v\noutput: %s", err, buf.String())
	}

	var report struct {
		WithFixtures           int      `json:"with_fixtures"`
		PositiveFixtureCount   int      `json:"positive_fixture_count"`
		NegativeAssertionCount int      `json:"negative_assertion_count"`
		StrictTop1FixtureCount int      `json:"strict_top_1_fixture_count"`
		FixtureMode            string   `json:"fixture_mode"`
		MissingFixtures        []string `json:"missing_fixtures"`
	}
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("unmarshal JSON output: %v\nraw: %s", err, buf.String())
	}

	if report.FixtureMode != "fixture-corpus" {
		t.Fatalf("fixture_mode = %q, want fixture-corpus", report.FixtureMode)
	}
	if report.WithFixtures != 2 || report.PositiveFixtureCount != 2 {
		t.Fatalf("expected two positive fixture-backed playbooks, got with=%d positive=%d", report.WithFixtures, report.PositiveFixtureCount)
	}
	if report.NegativeAssertionCount != 1 {
		t.Fatalf("expected one negative assertion, got %d", report.NegativeAssertionCount)
	}
	if report.StrictTop1FixtureCount != 1 {
		t.Fatalf("expected one strict top-1 fixture, got %d", report.StrictTop1FixtureCount)
	}
	if strings.Join(report.MissingFixtures, ",") != "unused-playbook" {
		t.Fatalf("missing_fixtures = %v, want [unused-playbook]", report.MissingFixtures)
	}
}

func TestCoverageCommandStillSupportsLegacyLogFixtureDir(t *testing.T) {
	pbDir := t.TempDir()
	fixtureDir := t.TempDir()
	writeCoveragePlaybook(t, pbDir, "docker-auth", "auth")
	writeCoveragePlaybook(t, pbDir, "missing-positive", "build")
	if err := os.WriteFile(filepath.Join(fixtureDir, "docker-auth.log"), []byte("denied\n"), 0o644); err != nil {
		t.Fatalf("write legacy fixture: %v", err)
	}

	cmd := newCoverageCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--playbooks", pbDir,
		"--fixture-dir", fixtureDir,
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("coverage command: %v\noutput: %s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Fixture mode          : log-stems") {
		t.Fatalf("expected log-stems fixture mode\n%s", out)
	}
	if !strings.Contains(out, "missing-positive") {
		t.Fatalf("expected missing-positive to be reported\n%s", out)
	}
}

func writeCoveragePlaybook(t *testing.T, dir, id, category string) {
	t.Helper()
	content := `id: ` + id + `
title: ` + id + ` title
category: ` + category + `
summary: |
  Minimal playbook for testing.
diagnosis: |
  Diagnosis text.
fix: |
  Run the fix command.
validation: |
  Validation text.
match:
  any:
    - ` + id + `-pattern
`
	if err := os.WriteFile(filepath.Join(dir, id+".yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write playbook %s: %v", id, err)
	}
}

func writeCoverageFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}
