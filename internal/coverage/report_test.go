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

// ── WriteText ─────────────────────────────────────────────────────────────────

func TestWriteTextHeader(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteText(&buf, Report{}); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "Playbook coverage report") {
		t.Errorf("output missing header; got:\n%s", got)
	}
}

func TestWriteTextNoDuplicates(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteText(&buf, Report{}); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "No duplicate IDs detected.") {
		t.Errorf("expected 'No duplicate IDs detected.' in output; got:\n%s", got)
	}
}

func TestWriteTextWithDuplicateIDs(t *testing.T) {
	var buf bytes.Buffer
	r := Report{DuplicateIDs: []string{"dup-a", "dup-b"}}
	if err := WriteText(&buf, r); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "Duplicate IDs (2)") {
		t.Errorf("expected 'Duplicate IDs (2)' in output; got:\n%s", got)
	}
	if !strings.Contains(got, "dup-a") {
		t.Errorf("expected 'dup-a' in output; got:\n%s", got)
	}
	if strings.Contains(got, "No duplicate IDs detected.") {
		t.Errorf("should not contain 'No duplicate IDs detected.' when duplicates exist; got:\n%s", got)
	}
}

func TestWriteTextWithMissingFixtures(t *testing.T) {
	var buf bytes.Buffer
	r := Report{MissingFixtures: []string{"playbook-x", "playbook-y"}}
	if err := WriteText(&buf, r); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "Playbooks missing positive fixtures (2)") {
		t.Errorf("expected missing fixtures section; got:\n%s", got)
	}
	if !strings.Contains(got, "playbook-x") {
		t.Errorf("expected 'playbook-x' in output; got:\n%s", got)
	}
}

func TestWriteTextWithByCategory(t *testing.T) {
	var buf bytes.Buffer
	r := Report{
		TotalPlaybooks: 5,
		WithFixtures:   3,
		ByCategory: []Category{
			{Category: "auth", Count: 3, WithFixtures: 2, PositiveFixtures: 4, NegativeAssertions: 1},
		},
	}
	if err := WriteText(&buf, r); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "auth") {
		t.Errorf("expected category 'auth' in output; got:\n%s", got)
	}
	if !strings.Contains(got, "By category:") {
		t.Errorf("expected 'By category:' header; got:\n%s", got)
	}
}

func TestWriteTextWithByDomain(t *testing.T) {
	var buf bytes.Buffer
	r := Report{
		ByDomain: []Domain{{Domain: "ci", Count: 4}},
	}
	if err := WriteText(&buf, r); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "By domain:") {
		t.Errorf("expected 'By domain:' header; got:\n%s", got)
	}
	if !strings.Contains(got, "ci") {
		t.Errorf("expected domain 'ci' in output; got:\n%s", got)
	}
}

func TestWriteTextWithSourceDetectorPlaybooks(t *testing.T) {
	var buf bytes.Buffer
	r := Report{
		SourceDetectorPlaybooks: []string{"source-pb-1", "source-pb-2"},
	}
	if err := WriteText(&buf, r); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "Source-detector playbooks (2)") {
		t.Errorf("expected source-detector section; got:\n%s", got)
	}
	if !strings.Contains(got, "source-pb-1") {
		t.Errorf("expected 'source-pb-1' in output; got:\n%s", got)
	}
}

func TestWriteTextStrictTop1FixtureCount(t *testing.T) {
	var buf bytes.Buffer
	r := Report{StrictTop1FixtureCount: 7}
	if err := WriteText(&buf, r); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "Strict top-1 fixtures") {
		t.Errorf("expected strict top-1 line in output; got:\n%s", got)
	}
}

func TestWriteTextFixtureRootAndMode(t *testing.T) {
	var buf bytes.Buffer
	r := Report{FixtureRoot: "/path/to/fixtures", FixtureMode: "strict"}
	if err := WriteText(&buf, r); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "/path/to/fixtures") {
		t.Errorf("expected fixture root in output; got:\n%s", got)
	}
	if !strings.Contains(got, "strict") {
		t.Errorf("expected fixture mode in output; got:\n%s", got)
	}
}

// ── normalizeReport ───────────────────────────────────────────────────────────

func TestNormalizeReportConvertsNilSlicesToEmpty(t *testing.T) {
	r := Report{} // all slice fields are nil
	got := normalizeReport(r)
	if got.ByCategory == nil {
		t.Error("ByCategory must be non-nil after normalizeReport")
	}
	if got.ByDomain == nil {
		t.Error("ByDomain must be non-nil after normalizeReport")
	}
	if got.MissingFixtures == nil {
		t.Error("MissingFixtures must be non-nil after normalizeReport")
	}
	if got.SourceDetectorPlaybooks == nil {
		t.Error("SourceDetectorPlaybooks must be non-nil after normalizeReport")
	}
	if got.DuplicateIDs == nil {
		t.Error("DuplicateIDs must be non-nil after normalizeReport")
	}
}

func TestNormalizeReportPreservesExistingSlices(t *testing.T) {
	r := Report{
		ByCategory:      []Category{{Category: "auth", Count: 2}},
		ByDomain:        []Domain{{Domain: "security", Count: 1}},
		MissingFixtures: []string{"playbook-without-fixture"},
	}
	got := normalizeReport(r)
	if len(got.ByCategory) != 1 || got.ByCategory[0].Category != "auth" {
		t.Errorf("ByCategory was modified: %v", got.ByCategory)
	}
	if len(got.ByDomain) != 1 {
		t.Errorf("ByDomain was modified: %v", got.ByDomain)
	}
	if len(got.MissingFixtures) != 1 || got.MissingFixtures[0] != "playbook-without-fixture" {
		t.Errorf("MissingFixtures was modified: %v", got.MissingFixtures)
	}
}

// ── fixtureLayoutForRoot ──────────────────────────────────────────────────────

func TestFixtureLayoutForRootNotFound(t *testing.T) {
	dir := t.TempDir() // empty directory — no fixture class subdirs
	_, ok := fixtureLayoutForRoot(dir)
	if ok {
		t.Errorf("expected ok=false for directory with no fixture class dirs, got true")
	}
}

func TestFixtureLayoutForRootWithDirectClassDirs(t *testing.T) {
	root := t.TempDir()
	// Create at least one fixture class directory directly under root
	if err := os.MkdirAll(filepath.Join(root, "minimal"), 0o755); err != nil {
		t.Fatalf("mkdir minimal: %v", err)
	}
	layout, ok := fixtureLayoutForRoot(root)
	if !ok {
		t.Fatal("expected ok=true when minimal/ exists under root")
	}
	if layout.Fixtures != root {
		t.Errorf("Layout.Fixtures = %q, want %q", layout.Fixtures, root)
	}
}

func TestFixtureLayoutForRootWithChildFixturesDir(t *testing.T) {
	parent := t.TempDir()
	fixtures := filepath.Join(parent, "fixtures")
	if err := os.MkdirAll(filepath.Join(fixtures, "real"), 0o755); err != nil {
		t.Fatalf("mkdir fixtures/real: %v", err)
	}
	layout, ok := fixtureLayoutForRoot(parent)
	if !ok {
		t.Fatal("expected ok=true when fixtures/real/ exists")
	}
	if layout.Root != parent {
		t.Errorf("Layout.Root = %q, want %q", layout.Root, parent)
	}
}
