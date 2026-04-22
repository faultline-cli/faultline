package playbooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/model"
)

func TestLoadDefaultUsesEnvVar(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "test.yaml", `
id: env-test
title: Env Test
category: test
severity: low
match:
  any:
    - "env error"
`)
	t.Setenv(envKey, dir)
	pbs, err := LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault with env: %v", err)
	}
	if len(pbs) == 0 {
		t.Fatal("expected playbooks from env-configured dir")
	}
	found := false
	for _, pb := range pbs {
		if pb.ID == "env-test" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find env-test playbook, got %v", pbs)
	}
}

func TestLoadDefaultEnvVarInvalidDirErrors(t *testing.T) {
	t.Setenv(envKey, "/nonexistent/path/that/does/not/exist")
	_, err := LoadDefault()
	if err == nil {
		t.Fatal("expected error for invalid env-configured dir")
	}
}

func TestLoadDefaultFallsBackToRepoPlaybooks(t *testing.T) {
	// Point to the actual bundled playbooks directory to verify LoadDefault works
	// when the env var points to a valid directory with yaml files.
	bundledDir := "../../playbooks/bundled"
	t.Setenv(envKey, bundledDir)
	pbs, err := LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault with bundled dir: %v", err)
	}
	if len(pbs) == 0 {
		t.Fatal("expected bundled playbooks to be found")
	}
}

func TestDefaultDirUsesEnvVar(t *testing.T) {
	dir := t.TempDir()
	// Create a yaml file so the directory is treated as valid
	path := filepath.Join(dir, "stub.yaml")
	if err := os.WriteFile(path, []byte("id: stub\n"), 0o600); err != nil {
		t.Fatalf("write stub: %v", err)
	}
	t.Setenv(envKey, dir)
	got, err := DefaultDir()
	if err != nil {
		t.Fatalf("DefaultDir with env: %v", err)
	}
	if !strings.HasPrefix(got, dir) {
		t.Errorf("expected DefaultDir to return %q, got %q", dir, got)
	}
}

func TestDefaultDirErrorsForNonexistentEnvPath(t *testing.T) {
	t.Setenv(envKey, "/no/such/dir")
	_, err := DefaultDir()
	if err == nil {
		t.Fatal("expected error for nonexistent FAULTLINE_PLAYBOOK_DIR")
	}
}

func TestLoadDirPreservesMatchNone(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "sample.yaml", `
id: sample
title: Sample
category: test
severity: low
match:
  any:
    - "primary error"
  none:
    - "ignore this"
`)

	pbs, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(pbs) != 1 {
		t.Fatalf("expected 1 playbook, got %d", len(pbs))
	}
	if got := pbs[0].Match.None; len(got) != 1 || got[0] != "ignore this" {
		t.Fatalf("expected match.none to be preserved, got %#v", got)
	}
}

func TestLoadDirRejectsMatchNoneOverlap(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "sample.yaml", `
id: sample
title: Sample
category: test
severity: low
match:
  any:
    - "primary error"
  none:
    - "PRIMARY ERROR"
`)

	_, err := LoadDir(dir)
	if err == nil {
		t.Fatal("expected LoadDir to reject overlapping match.none pattern")
	}
	if !strings.Contains(err.Error(), "match.none") {
		t.Fatalf("expected match.none validation error, got %v", err)
	}
}

func TestLoadDirRejectsEmptyMatchPattern(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "sample.yaml", `
id: sample
title: Sample
category: test
severity: low
match:
  any:
    - "primary error"
    - "   "
`)

	_, err := LoadDir(dir)
	if err == nil {
		t.Fatal("expected LoadDir to reject empty match.any pattern")
	}
	if !strings.Contains(err.Error(), "must not be empty") {
		t.Fatalf("expected empty-pattern validation error, got %v", err)
	}
}

func TestLoadDirSupportsSourceDetector(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "sample.yaml", `
id: source-sample
title: Source Sample
category: runtime
detector: source
severity: high
source:
  triggers:
    - id: outbound
      patterns:
        - client.Do(
`)

	pbs, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(pbs) != 1 {
		t.Fatalf("expected 1 playbook, got %d", len(pbs))
	}
	if pbs[0].Detector != "source" {
		t.Fatalf("expected source detector, got %q", pbs[0].Detector)
	}
	if len(pbs[0].Source.Triggers) != 1 {
		t.Fatalf("expected source trigger to load, got %#v", pbs[0].Source.Triggers)
	}
}

func TestLoadDirSupportsMarkdownContentFields(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "sample.yaml", `
id: markdown-sample
title: Markdown Sample
category: build
severity: medium
summary: |
  Short summary.
diagnosis: |
  ## Diagnosis

  Detailed markdown.
fix: |
  1. Run fix
validation: |
  - Verify fix
match:
  any:
    - "primary error"
workflow:
  verify:
    - go test ./...
`)

	pbs, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(pbs) != 1 {
		t.Fatalf("expected 1 playbook, got %d", len(pbs))
	}
	pb := pbs[0]
	if pb.Summary != "Short summary." {
		t.Fatalf("expected summary, got %q", pb.Summary)
	}
	if !strings.Contains(pb.Diagnosis, "## Diagnosis") {
		t.Fatalf("expected markdown diagnosis, got %q", pb.Diagnosis)
	}
	if !strings.Contains(pb.Validation, "Verify fix") {
		t.Fatalf("expected validation markdown, got %q", pb.Validation)
	}
}

func TestLoadDirSupportsHypothesisMetadata(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "sample.yaml", `
id: hypothesis-sample
title: Hypothesis Sample
category: build
severity: medium
summary: |
  Summary.
diagnosis: |
  Diagnosis.
fix: |
  Fix.
validation: |
  Validate.
match:
  any:
    - "primary error"
hypothesis:
  supports:
    - signal: dependency.resolution.conflict
      weight: 0.7
  contradicts:
    - signal: cache.restore.absent
      weight: -0.4
  discriminators:
    - description: Resolver wording is specific.
      signal: dependency.resolution.conflict
  excludes:
    - signal: dependency.hash.mismatch
`)

	pbs, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(pbs) != 1 {
		t.Fatalf("expected 1 playbook, got %d", len(pbs))
	}
	if len(pbs[0].Hypothesis.Supports) != 1 || pbs[0].Hypothesis.Supports[0].Signal != "dependency.resolution.conflict" {
		t.Fatalf("expected hypothesis supports to load, got %#v", pbs[0].Hypothesis)
	}
	if len(pbs[0].Hypothesis.Excludes) != 1 || pbs[0].Hypothesis.Excludes[0].Signal != "dependency.hash.mismatch" {
		t.Fatalf("expected hypothesis excludes to load, got %#v", pbs[0].Hypothesis)
	}
}

func TestLoadDirRejectsUnknownHypothesisSignal(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "sample.yaml", `
id: invalid-hypothesis
title: Invalid Hypothesis
category: build
severity: medium
summary: |
  Summary.
diagnosis: |
  Diagnosis.
fix: |
  Fix.
validation: |
  Validate.
match:
  any:
    - "primary error"
hypothesis:
  supports:
    - signal: not.a.real.signal
`)

	_, err := LoadDir(dir)
	if err == nil {
		t.Fatal("expected unknown hypothesis signal to fail validation")
	}
	if !strings.Contains(err.Error(), "unknown signal") {
		t.Fatalf("expected unknown signal error, got %v", err)
	}
}

func TestLoadDirRejectsMissingMarkdownFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.yaml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(`
id: invalid-sample
title: Invalid Sample
category: build
severity: medium
summary: |
  Summary only.
match:
  any:
    - "primary error"
`)+"\n"), 0o600); err != nil {
		t.Fatalf("write invalid fixture: %v", err)
	}

	_, err := LoadDir(dir)
	if err == nil {
		t.Fatal("expected missing markdown fields to fail validation")
	}
	if !strings.Contains(err.Error(), "diagnosis") {
		t.Fatalf("expected diagnosis validation error, got %v", err)
	}
}

func TestLoadPacksRejectsDuplicateIDsAcrossPacks(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	writePlaybookFixture(t, first, "one.yaml", `
id: shared
title: Shared
category: test
severity: low
match:
  any:
    - "primary error"
`)
	writePlaybookFixture(t, second, "two.yaml", `
id: shared
title: Shared Again
category: test
severity: low
match:
  any:
    - "secondary error"
`)

	_, err := LoadPacks([]Pack{
		{Name: "starter", Root: first},
		{Name: "extra", Root: second},
	})
	if err == nil {
		t.Fatal("expected duplicate ID error across packs")
	}
	if !strings.Contains(err.Error(), "across packs") {
		t.Fatalf("expected cross-pack duplicate error, got %v", err)
	}
}

func TestLoadPacksStampsPackMetadata(t *testing.T) {
	first := t.TempDir()
	writePlaybookFixture(t, first, "one.yaml", `
id: shared
title: Shared
category: test
severity: low
match:
  any:
    - "primary error"
`)

	pbs, err := LoadPacks([]Pack{{Name: "team-pack", Root: first}})
	if err != nil {
		t.Fatalf("LoadPacks: %v", err)
	}
	if len(pbs) != 1 {
		t.Fatalf("expected one playbook, got %d", len(pbs))
	}
	if pbs[0].Metadata.PackName != "team-pack" {
		t.Fatalf("expected pack name to be stamped, got %#v", pbs[0].Metadata)
	}
	if pbs[0].Metadata.PackRoot != first {
		t.Fatalf("expected pack root %q, got %#v", first, pbs[0].Metadata)
	}
	if pbs[0].Metadata.SourceFile == "" {
		t.Fatalf("expected source file metadata, got %#v", pbs[0].Metadata)
	}
}

func TestCatalogUsesExplicitCustomPack(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "sample.yaml", `
id: custom-sample
title: Custom Sample
category: test
severity: low
match:
  any:
    - "primary error"
`)

	catalog := NewCatalog(dir)
	packs, err := catalog.Packs()
	if err != nil {
		t.Fatalf("Packs: %v", err)
	}
	if len(packs) != 1 || packs[0].Name != CustomPackName || packs[0].Root != dir {
		t.Fatalf("unexpected custom packs: %#v", packs)
	}
}

func TestFindPatternConflictsBundled(t *testing.T) {
	pbs, err := LoadDir("../../playbooks/bundled")
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}

	conflicts := FindPatternConflicts(pbs)
	if len(conflicts) == 0 {
		t.Fatal("expected bundled playbooks to produce at least one pattern conflict report")
	}

	assertConflict(t, conflicts, "configmap not found", []string{"config-mismatch"}, []string{"env-var-missing", "missing-test-fixture"})

	report := FormatPatternConflicts(conflicts)
	if !strings.Contains(report, "configmap not found") {
		t.Fatalf("expected conflict report to include configmap not found, got %q", report)
	}
}

func assertConflict(t *testing.T, conflicts []PatternConflict, pattern string, wantPositive, wantNegative []string) {
	t.Helper()

	for _, conflict := range conflicts {
		if conflict.Pattern != pattern {
			continue
		}
		gotPos := conflictIDs(conflict.Positive)
		gotNeg := conflictIDs(conflict.Negative)
		if !sameStringSet(gotPos, wantPositive) {
			t.Fatalf("pattern %q positive refs: want %v, got %v", pattern, wantPositive, gotPos)
		}
		if !sameStringSet(gotNeg, wantNegative) {
			t.Fatalf("pattern %q negative refs: want %v, got %v", pattern, wantNegative, gotNeg)
		}
		return
	}
	t.Fatalf("expected conflict for pattern %q", pattern)
}

func conflictIDs(refs []PatternRef) []string {
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		ids = append(ids, ref.PlaybookID)
	}
	return ids
}

func sameStringSet(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	seen := make(map[string]int, len(got))
	for _, s := range got {
		seen[s]++
	}
	for _, s := range want {
		if seen[s] == 0 {
			return false
		}
		seen[s]--
	}
	for _, count := range seen {
		if count != 0 {
			return false
		}
	}
	return true
}

func writePlaybookFixture(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	content = strings.TrimSpace(content)
	defaults := map[string]string{
		"summary:":    "summary: |\n  Test summary.",
		"diagnosis:":  "diagnosis: |\n  ## Diagnosis\n\n  Test diagnosis.",
		"fix:":        "fix: |\n  ## Fix steps\n\n  1. Test fix.",
		"validation:": "validation: |\n  ## Validation\n\n  - Test validation.",
	}
	for key, block := range defaults {
		if !strings.Contains(content, key) {
			content += "\n" + block
		}
	}
	if err := os.WriteFile(path, []byte(content+"\n"), 0o600); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}

// ── pack provenance ───────────────────────────────────────────────────────────

func TestProvenanceFromPlaybooks(t *testing.T) {
	pbs := []model.Playbook{
		{ID: "a", Metadata: model.PlaybookMeta{PackName: "starter", PackRoot: "/p1", PackVersion: ""}},
		{ID: "b", Metadata: model.PlaybookMeta{PackName: "starter", PackRoot: "/p1", PackVersion: ""}},
		{ID: "c", Metadata: model.PlaybookMeta{PackName: "premium", PackRoot: "/p2", PackVersion: "1.0.0", PackSourceURL: "https://example.com/premium.git", PackPinnedRef: "abc1234"}},
	}
	prov := ProvenanceFromPlaybooks(pbs)
	if len(prov) != 2 {
		t.Fatalf("expected 2 provenance entries, got %d", len(prov))
	}
	if prov[0].Name != "starter" || prov[0].PlaybookCount != 2 {
		t.Errorf("starter entry = %+v", prov[0])
	}
	if prov[1].Name != "premium" || prov[1].Version != "1.0.0" || prov[1].SourceURL != "https://example.com/premium.git" || prov[1].PinnedRef != "abc1234" || prov[1].PlaybookCount != 1 {
		t.Errorf("premium entry = %+v", prov[1])
	}
}

func TestProvenanceFromPlaybooksEmpty(t *testing.T) {
	prov := ProvenanceFromPlaybooks(nil)
	if len(prov) != 0 {
		t.Errorf("expected empty provenance for nil playbooks, got %v", prov)
	}
}

func TestLoadPacksPropagatesPackMeta(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "rule.yaml", `
id: provenance-playbook
title: Provenance Playbook
category: test
severity: low
match:
  any:
    - "prov error"
`)
	if err := WritePackMeta(dir, PackMeta{Name: "prov-pack", Version: "2.3.4", SourceURL: "https://example.com/prov-pack.git", PinnedRef: "deadbeef"}); err != nil {
		t.Fatalf("WritePackMeta: %v", err)
	}

	packs := []Pack{{Name: "prov-pack", Root: dir}}
	pbs, err := LoadPacks(packs)
	if err != nil {
		t.Fatalf("LoadPacks: %v", err)
	}
	if len(pbs) != 1 {
		t.Fatalf("expected 1 playbook, got %d", len(pbs))
	}
	if pbs[0].Metadata.PackVersion != "2.3.4" {
		t.Errorf("PackVersion = %q, want 2.3.4", pbs[0].Metadata.PackVersion)
	}
	if pbs[0].Metadata.PackSourceURL != "https://example.com/prov-pack.git" {
		t.Errorf("PackSourceURL = %q, want https://example.com/prov-pack.git", pbs[0].Metadata.PackSourceURL)
	}
	if pbs[0].Metadata.PackPinnedRef != "deadbeef" {
		t.Errorf("PackPinnedRef = %q, want deadbeef", pbs[0].Metadata.PackPinnedRef)
	}
}

func TestLoadPacksUsesExplicitPackMetadataWhenManifestMissing(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "rule.yaml", `
id: explicit-meta-playbook
title: Explicit Meta Playbook
category: test
severity: low
match:
  any:
    - "explicit meta error"
`)

	packs := []Pack{{
		Name:      "explicit-meta-pack",
		Root:      dir,
		Version:   "9.9.9",
		SourceURL: "https://example.com/explicit-meta-pack.git",
		PinnedRef: "cafebabe",
	}}
	pbs, err := LoadPacks(packs)
	if err != nil {
		t.Fatalf("LoadPacks: %v", err)
	}
	if len(pbs) != 1 {
		t.Fatalf("expected 1 playbook, got %d", len(pbs))
	}
	if pbs[0].Metadata.PackVersion != "9.9.9" {
		t.Errorf("PackVersion = %q, want 9.9.9", pbs[0].Metadata.PackVersion)
	}
	if pbs[0].Metadata.PackSourceURL != "https://example.com/explicit-meta-pack.git" {
		t.Errorf("PackSourceURL = %q, want https://example.com/explicit-meta-pack.git", pbs[0].Metadata.PackSourceURL)
	}
	if pbs[0].Metadata.PackPinnedRef != "cafebabe" {
		t.Errorf("PackPinnedRef = %q, want cafebabe", pbs[0].Metadata.PackPinnedRef)
	}
}
