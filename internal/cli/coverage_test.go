package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/model"
)

func TestNewCoverageCommandIsNotHidden(t *testing.T) {
	cmd := newCoverageCommand()
	if cmd.Hidden {
		t.Fatal("coverage command must not be hidden")
	}
}

func TestNewCoverageCommandAcceptsNoArgs(t *testing.T) {
	cmd := newCoverageCommand()
	if err := cmd.Args(cmd, nil); err != nil {
		t.Fatalf("expected no args to be valid: %v", err)
	}
}

func TestNewCoverageCommandRejectsPositionalArgs(t *testing.T) {
	cmd := newCoverageCommand()
	if err := cmd.Args(cmd, []string{"unexpected"}); err == nil {
		t.Fatal("expected error for positional args, got nil")
	}
}

// stubPlaybooks returns a small slice of model.Playbook for use in unit tests.
func stubPlaybooks() []model.Playbook {
	return []model.Playbook{
		{ID: "docker-auth", Category: "docker"},
		{ID: "npm-install", Category: "node"},
		{ID: "go-build", Category: "go"},
		{ID: "go-vet", Category: "go"},
		{ID: "uncategorized-pb"},
	}
}

func TestPrintCoverageTextStructure(t *testing.T) {
	pbs := stubPlaybooks()
	byCategory := map[string][]string{
		"docker":        {"docker-auth"},
		"node":          {"npm-install"},
		"go":            {"go-build", "go-vet"},
		"uncategorized": {"uncategorized-pb"},
	}
	missing := []string{"uncategorized-pb"}
	duplicates := []string{}

	var buf bytes.Buffer
	if err := printCoverageText(&buf, pbs, byCategory, missing, duplicates, ""); err != nil {
		t.Fatalf("printCoverageText: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "Total playbooks") {
		t.Error("expected 'Total playbooks' header")
	}
	if !strings.Contains(out, "5") {
		t.Error("expected total playbook count '5'")
	}
	if !strings.Contains(out, "By category:") {
		t.Error("expected 'By category:' section")
	}
	if !strings.Contains(out, "docker") {
		t.Error("expected 'docker' category")
	}
	if !strings.Contains(out, "go") {
		t.Error("expected 'go' category")
	}
	if !strings.Contains(out, "Playbooks missing fixtures") {
		t.Error("expected 'Playbooks missing fixtures' section")
	}
	if !strings.Contains(out, "uncategorized-pb") {
		t.Error("expected missing fixture 'uncategorized-pb'")
	}
	if !strings.Contains(out, "No duplicate IDs detected") {
		t.Error("expected 'No duplicate IDs detected'")
	}
}

func TestPrintCoverageTextReportsDuplicates(t *testing.T) {
	pbs := stubPlaybooks()
	byCategory := map[string][]string{
		"go": {"go-build"},
	}
	duplicates := []string{"go-build (×2)"}
	missing := []string{}

	var buf bytes.Buffer
	if err := printCoverageText(&buf, pbs, byCategory, missing, duplicates, "/tmp/fixtures"); err != nil {
		t.Fatalf("printCoverageText: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Duplicate IDs") {
		t.Error("expected 'Duplicate IDs' section")
	}
	if !strings.Contains(out, "go-build (×2)") {
		t.Error("expected duplicate entry 'go-build (×2)'")
	}
}

func TestPrintCoverageTextIncludesFixtureDir(t *testing.T) {
	pbs := stubPlaybooks()
	byCategory := map[string][]string{"go": {"go-build"}}
	var buf bytes.Buffer
	if err := printCoverageText(&buf, pbs, byCategory, nil, nil, "/custom/fixtures"); err != nil {
		t.Fatalf("printCoverageText: %v", err)
	}
	if !strings.Contains(buf.String(), "/custom/fixtures") {
		t.Error("expected fixture dir to appear in output")
	}
}

func TestPrintCoverageJSON(t *testing.T) {
	pbs := stubPlaybooks()
	byCategory := map[string][]string{
		"docker": {"docker-auth"},
		"go":     {"go-build", "go-vet"},
	}
	missing := []string{"go-vet"}
	duplicates := []string{}

	var buf bytes.Buffer
	if err := printCoverageJSON(&buf, pbs, byCategory, missing, duplicates, "/tmp/fixtures"); err != nil {
		t.Fatalf("printCoverageJSON: %v", err)
	}

	var report coverageReportJSON
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}

	if report.TotalPlaybooks != len(pbs) {
		t.Errorf("total_playbooks = %d, want %d", report.TotalPlaybooks, len(pbs))
	}
	if report.FixtureDir != "/tmp/fixtures" {
		t.Errorf("fixture_dir = %q, want %q", report.FixtureDir, "/tmp/fixtures")
	}
	if len(report.MissingFixtures) != 1 || report.MissingFixtures[0] != "go-vet" {
		t.Errorf("missing_fixtures = %v, want [go-vet]", report.MissingFixtures)
	}
	if len(report.DuplicateIDs) != 0 {
		t.Errorf("duplicate_ids should be empty, got %v", report.DuplicateIDs)
	}
	if len(report.ByCategory) != 2 {
		t.Errorf("by_category count = %d, want 2", len(report.ByCategory))
	}
}

func TestPrintCoverageJSONNullSafety(t *testing.T) {
	// Ensure slices are never null in JSON output (should be [] not null).
	pbs := stubPlaybooks()
	var buf bytes.Buffer
	if err := printCoverageJSON(&buf, pbs, map[string][]string{}, nil, nil, ""); err != nil {
		t.Fatalf("printCoverageJSON: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"missing_fixtures", "duplicate_ids", "by_category"} {
		val := string(raw[key])
		if val == "null" {
			t.Errorf("JSON key %q must not be null", key)
		}
	}
}

func TestCoverageCommandRunsWithMinimalPlaybookDir(t *testing.T) {
	// Set up a minimal playbook directory with one valid playbook YAML and one fixture.
	pbDir := t.TempDir()
	fixtureDir := t.TempDir()

	writeTestPlaybook(t, pbDir, "pb-test-one.yaml", minimalPlaybook("pb-test-one", "test"))
	writeTestPlaybook(t, pbDir, "pb-test-two.yaml", minimalPlaybook("pb-test-two", "test"))

	// Only provide a fixture for pb-test-one.
	if err := os.WriteFile(filepath.Join(fixtureDir, "pb-test-one.log"), []byte("error occurred\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
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
	if !strings.Contains(out, "pb-test-two") {
		t.Errorf("expected missing fixture 'pb-test-two' in output\n%s", out)
	}
	if !strings.Contains(out, "2") {
		t.Errorf("expected total count 2 in output\n%s", out)
	}
}

func TestCoverageCommandJSONRunsWithMinimalPlaybookDir(t *testing.T) {
	pbDir := t.TempDir()
	fixtureDir := t.TempDir()

	writeTestPlaybook(t, pbDir, "pb-json.yaml", minimalPlaybook("pb-json", "test"))

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
		t.Fatalf("coverage command --json: %v\noutput: %s", err, buf.String())
	}

	var report coverageReportJSON
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("unmarshal JSON output: %v\nraw: %s", err, buf.String())
	}

	if report.TotalPlaybooks != 1 {
		t.Errorf("total_playbooks = %d, want 1", report.TotalPlaybooks)
	}
	if report.WithFixtures != 0 {
		t.Errorf("with_fixtures = %d, want 0 (no fixture file provided)", report.WithFixtures)
	}
}

// minimalPlaybook returns a YAML string with all required playbook fields populated.
func minimalPlaybook(id, category string) string {
	return `id: ` + id + `
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
}

// writeTestPlaybook writes a YAML playbook file into dir.
func writeTestPlaybook(t *testing.T, dir, filename, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatalf("write playbook %s: %v", filename, err)
	}
}

func TestResolveDefaultFixtureDirFindsRepoDir(t *testing.T) {
	// resolveDefaultFixtureDir walks upward from cwd looking for
	// internal/engine/testdata/fixtures. Running from the repo root
	// (or any subdirectory) it should find the bundled fixture directory.
	got := resolveDefaultFixtureDir()
	if got == "" {
		t.Fatal("bundled fixture directory not found from current working directory")
	}
	if !filepath.IsAbs(got) {
		t.Errorf("expected absolute path, got %q", got)
	}
	if info, err := os.Stat(got); err != nil || !info.IsDir() {
		t.Errorf("resolved path does not exist or is not a directory: %q", got)
	}
}

func TestResolveDefaultFixtureDirReturnEmptyOutsideRepo(t *testing.T) {
	// Change the working directory to a temp dir that does NOT contain the
	// engine testdata tree.  resolveDefaultFixtureDir should return "".
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if got := resolveDefaultFixtureDir(); got != "" {
		t.Errorf("expected empty string outside repo, got %q", got)
	}
}

func TestLoadPlaybooksForCoverageDefaultPath(t *testing.T) {
	// Make default playbook resolution deterministic for this test by
	// pointing the default directory environment variable at a temp fixture.
	pbDir := t.TempDir()
	writeTestPlaybook(t, pbDir, "pb-default.yaml", minimalPlaybook("pb-default", "test"))
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", pbDir)

	pbs, err := loadPlaybooksForCoverage("", nil)
	if err != nil {
		t.Fatalf("loadPlaybooksForCoverage (default path): %v", err)
	}
	if len(pbs) != 1 || pbs[0].ID != "pb-default" {
		t.Errorf("expected exactly one playbook pb-default, got %v", pbs)
	}
}

func TestLoadPlaybooksForCoverageExplicitDir(t *testing.T) {
	pbDir := t.TempDir()
	writeTestPlaybook(t, pbDir, "pb-explicit.yaml", minimalPlaybook("pb-explicit", "test"))

	pbs, err := loadPlaybooksForCoverage(pbDir, nil)
	if err != nil {
		t.Fatalf("loadPlaybooksForCoverage (explicit dir): %v", err)
	}
	if len(pbs) != 1 || pbs[0].ID != "pb-explicit" {
		t.Errorf("expected exactly one playbook pb-explicit, got %v", pbs)
	}
}

func TestLoadPlaybooksForCoverageWithExtraPacks(t *testing.T) {
	// Supply an explicit playbook dir AND an extra pack dir.
	pbDir := t.TempDir()
	packDir := t.TempDir()
	writeTestPlaybook(t, pbDir, "pb-base.yaml", minimalPlaybook("pb-base", "test"))
	writeTestPlaybook(t, packDir, "pb-pack.yaml", minimalPlaybook("pb-pack", "test"))

	pbs, err := loadPlaybooksForCoverage(pbDir, []string{packDir})
	if err != nil {
		t.Fatalf("loadPlaybooksForCoverage (with pack): %v", err)
	}
	if len(pbs) < 2 {
		t.Errorf("expected at least 2 playbooks (base + pack), got %d", len(pbs))
	}
	ids := make(map[string]bool)
	for _, pb := range pbs {
		ids[pb.ID] = true
	}
	for _, want := range []string{"pb-base", "pb-pack"} {
		if !ids[want] {
			t.Errorf("expected playbook %q in loaded set", want)
		}
	}
}

