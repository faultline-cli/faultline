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

	entries, err := os.ReadDir(pbDir)
	if err != nil {
		t.Fatalf("read playbook dir: %v", err)
	}
	playbookCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			playbookCount++
		}
	}
	if playbookCount != 3 {
		t.Fatalf("expected exactly 3 playbooks in catalog setup, got %d", playbookCount)
	}

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
		Robustness             []struct {
			PlaybookID        string `json:"playbook_id"`
			ConfidenceScore   int    `json:"confidence_score"`
			FalsePositiveRisk string `json:"false_positive_risk"`
			FalseNegativeRisk string `json:"false_negative_risk"`
		} `json:"robustness"`
	}
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("unmarshal JSON output: %v\nraw: %s", err, buf.String())
	}
	var raw map[string]any
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal JSON output (raw): %v\nraw: %s", err, buf.String())
	}

	if report.FixtureMode != "fixture-corpus" {
		t.Fatalf("fixture_mode = %q, want fixture-corpus", report.FixtureMode)
	}
	if report.WithFixtures != 2 || report.PositiveFixtureCount != 2 {
		t.Fatalf("expected two positive fixture-backed playbooks, got with=%d positive=%d", report.WithFixtures, report.PositiveFixtureCount)
	}

	seenExpected := map[string]bool{
		"npm-registry-auth": false,
		"install-failure":   false,
	}
	if cats, ok := raw["by_category"].([]any); ok {
		for _, cat := range cats {
			if m, ok := cat.(map[string]any); ok {
				if ids, ok := m["playbook_ids"].([]any); ok {
					for _, id := range ids {
						if s, ok := id.(string); ok {
							if _, want := seenExpected[s]; want {
								seenExpected[s] = true
							}
						}
					}
				}
			}
		}
	}
	for id, seen := range seenExpected {
		if !seen {
			t.Fatalf("coverage report missing expected playbook id %q in by_category playbook_ids; raw: %s", id, buf.String())
		}
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
	if len(report.Robustness) != 3 {
		t.Fatalf("robustness length = %d, want 3", len(report.Robustness))
	}
	if report.Robustness[0].ConfidenceScore > report.Robustness[len(report.Robustness)-1].ConfidenceScore {
		t.Fatalf("robustness entries are not sorted by ascending score: %#v", report.Robustness)
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
		"--json",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("coverage command: %v\noutput: %s", err, buf.String())
	}
	out := buf.String()
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("parse json output: %v\n%s", err, out)
	}
	mode, ok := payload["fixture_mode"].(string)
	if !ok {
		t.Fatalf("expected fixture_mode string in output: %s", out)
	}
	if mode != "log-stems" {
		t.Fatalf("expected fixture_mode log-stems, got %q\n%s", mode, out)
	}
	missingFixtures, ok := payload["missing_fixtures"].([]any)
	if !ok {
		t.Fatalf("expected missing_fixtures array in output: %s", out)
	}
	found := false
	for _, v := range missingFixtures {
		if s, ok := v.(string); ok && s == "missing-positive" {
			found = true
			break
		}
	}
	if !found {
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
