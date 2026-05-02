package coverage

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildCountsFixtureExpectationsAndNegativeAssertions(t *testing.T) {
	pbDir := t.TempDir()
	fixtureRoot := filepath.Join(t.TempDir(), "fixtures")
	if err := os.MkdirAll(filepath.Join(fixtureRoot, "minimal"), 0o755); err != nil {
		t.Fatalf("create fixture corpus: %v", err)
	}

	writePlaybook(t, pbDir, "npm-registry-auth", "auth", "package")
	writePlaybook(t, pbDir, "install-failure", "build", "package")
	writePlaybook(t, pbDir, "uncovered-playbook", "build", "")
	writeFixture(t, filepath.Join(fixtureRoot, "minimal", "positive.yaml"), `
id: npm-registry-auth-positive
fixture_class: minimal
normalized_log: |
  npm ERR! code E401
expectation:
  expected_playbook: npm-registry-auth
  strict_top1: true
`)
	writeFixture(t, filepath.Join(fixtureRoot, "minimal", "near-miss.yaml"), `
id: npm-registry-auth-near-miss
fixture_class: minimal
normalized_log: |
  npm ERR! code E404
expectation:
  expected_playbook: install-failure
  disallowed_playbooks:
    - npm-registry-auth
`)

	report, err := Build(Options{PlaybookDir: pbDir, FixtureRoot: fixtureRoot})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if report.FixtureMode != "fixture-corpus" {
		t.Fatalf("FixtureMode = %q, want fixture-corpus", report.FixtureMode)
	}
	if report.TotalPlaybooks != 3 {
		t.Fatalf("TotalPlaybooks = %d, want 3", report.TotalPlaybooks)
	}
	if report.WithFixtures != 2 {
		t.Fatalf("WithFixtures = %d, want 2", report.WithFixtures)
	}
	if report.PositiveFixtureCount != 2 {
		t.Fatalf("PositiveFixtureCount = %d, want 2", report.PositiveFixtureCount)
	}
	if report.NegativeAssertionCount != 1 {
		t.Fatalf("NegativeAssertionCount = %d, want 1", report.NegativeAssertionCount)
	}
	if report.StrictTop1FixtureCount != 1 {
		t.Fatalf("StrictTop1FixtureCount = %d, want 1", report.StrictTop1FixtureCount)
	}
	if strings.Join(report.MissingFixtures, ",") != "uncovered-playbook" {
		t.Fatalf("MissingFixtures = %v, want [uncovered-playbook]", report.MissingFixtures)
	}

	auth := categoryByName(t, report, "auth")
	if auth.PositiveFixtures != 1 || auth.NegativeAssertions != 1 {
		t.Fatalf("auth evidence = positive %d negative %d, want 1/1", auth.PositiveFixtures, auth.NegativeAssertions)
	}
	build := categoryByName(t, report, "build")
	if build.Count != 2 || build.WithFixtures != 1 {
		t.Fatalf("build category count/with = %d/%d, want 2/1", build.Count, build.WithFixtures)
	}
}

func TestBuildFallsBackToLegacyLogStemFixtures(t *testing.T) {
	pbDir := t.TempDir()
	fixtureDir := t.TempDir()
	writePlaybook(t, pbDir, "docker-auth", "auth", "")
	writePlaybook(t, pbDir, "missing-positive", "build", "")
	if err := os.WriteFile(filepath.Join(fixtureDir, "docker-auth.log"), []byte("auth failure\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	report, err := Build(Options{PlaybookDir: pbDir, FixtureRoot: fixtureDir})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if report.FixtureMode != "log-stems" {
		t.Fatalf("FixtureMode = %q, want log-stems", report.FixtureMode)
	}
	if report.WithFixtures != 1 || report.PositiveFixtureCount != 1 {
		t.Fatalf("expected one legacy positive fixture, got with=%d positive=%d", report.WithFixtures, report.PositiveFixtureCount)
	}
	if strings.Join(report.MissingFixtures, ",") != "missing-positive" {
		t.Fatalf("MissingFixtures = %v, want [missing-positive]", report.MissingFixtures)
	}
}

func TestWriteJSONKeepsStableEmptySlices(t *testing.T) {
	var buf bytes.Buffer
	err := WriteJSON(&buf, Report{
		ByCategory:      []Category{},
		ByDomain:        []Domain{},
		MissingFixtures: []string{},
		DuplicateIDs:    []string{},
	})
	if err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"by_category", "by_domain", "missing_fixtures", "duplicate_ids"} {
		if string(raw[key]) == "null" {
			t.Fatalf("%s rendered as null; want []", key)
		}
	}
}

func categoryByName(t *testing.T, report Report, name string) Category {
	t.Helper()
	for _, category := range report.ByCategory {
		if category.Category == name {
			return category
		}
	}
	t.Fatalf("category %q not found in %#v", name, report.ByCategory)
	return Category{}
}

func writePlaybook(t *testing.T, dir, id, category, domain string) {
	t.Helper()
	content := `id: ` + id + `
title: ` + id + ` title
category: ` + category + `
domain: ` + domain + `
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

func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}
