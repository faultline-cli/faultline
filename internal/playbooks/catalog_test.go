package playbooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogUsesEnvExtraPacks(t *testing.T) {
	bundled := t.TempDir()
	extra := t.TempDir()
	if err := os.MkdirAll(filepath.Join(bundled, "log"), 0o755); err != nil {
		t.Fatalf("mkdir bundled log dir: %v", err)
	}
	writePlaybookFixture(t, filepath.Join(bundled, "log"), "base.yaml", `
id: bundled-only
title: Bundled
category: test
severity: low
match:
  any:
    - "bundled error"
`)
	writePlaybookFixture(t, extra, "extra.yaml", `
id: extra-only
title: Extra
category: test
severity: low
match:
  any:
    - "extra error"
`)

	t.Setenv(envKey, bundled)
	t.Setenv(packsEnvKey, extra)

	packs, err := NewCatalog("").Packs()
	if err != nil {
		t.Fatalf("Packs: %v", err)
	}
	if len(packs) != 2 {
		t.Fatalf("expected bundled and extra pack, got %#v", packs)
	}
	if packs[0].Name != BundledPackName || packs[1].Root != extra || packs[1].Name != filepath.Base(extra) {
		t.Fatalf("unexpected resolved packs: %#v", packs)
	}
}

func TestCatalogOverrideIgnoresExtraPackEnv(t *testing.T) {
	override := t.TempDir()
	writePlaybookFixture(t, override, "override.yaml", `
id: override-only
title: Override
category: test
severity: low
match:
  any:
    - "override error"
`)

	t.Setenv(packsEnvKey, filepath.Join(string(os.PathSeparator), "ignored"))

	packs, err := NewCatalog(override).Packs()
	if err != nil {
		t.Fatalf("Packs: %v", err)
	}
	if len(packs) != 1 || packs[0].Name != CustomPackName || packs[0].Root != override {
		t.Fatalf("unexpected override packs: %#v", packs)
	}
}

func TestCatalogRejectsOverrideWithExplicitExtraPacks(t *testing.T) {
	override := t.TempDir()
	extra := t.TempDir()
	writePlaybookFixture(t, override, "override.yaml", `
id: override-only
title: Override
category: test
severity: low
match:
  any:
    - "override error"
`)
	writePlaybookFixture(t, extra, "extra.yaml", `
id: extra-only
title: Extra
category: test
severity: low
match:
  any:
    - "extra error"
`)

	_, err := NewCatalogWithOptions(CatalogOptions{
		OverrideDir:   override,
		ExtraPackDirs: []string{extra},
	}).Packs()
	if err == nil {
		t.Fatal("expected override-plus-extra pack error")
	}
}

func TestCatalogIncludesInstalledPacks(t *testing.T) {
	bundled := t.TempDir()
	home := t.TempDir()
	installedRoot := filepath.Join(home, ".faultline", installedPacksSubdir, "team-pack")
	if err := os.MkdirAll(filepath.Join(bundled, "log"), 0o755); err != nil {
		t.Fatalf("mkdir bundled log dir: %v", err)
	}
	writePlaybookFixture(t, filepath.Join(bundled, "log"), "base.yaml", `
id: bundled-only
title: Bundled
category: test
severity: low
match:
  any:
    - "bundled error"
`)
	if err := os.MkdirAll(installedRoot, 0o755); err != nil {
		t.Fatalf("mkdir installed root: %v", err)
	}
	writePlaybookFixture(t, installedRoot, "extra.yaml", `
id: extra-only
title: Team Pack
category: test
severity: low
match:
  any:
		- "extra error"
`)

	t.Setenv(envKey, bundled)
	t.Setenv("HOME", home)

	packs, err := NewCatalog("").Packs()
	if err != nil {
		t.Fatalf("Packs: %v", err)
	}
	if len(packs) != 2 {
		t.Fatalf("expected bundled and installed pack, got %#v", packs)
	}
	if packs[1].Name != "team-pack" || packs[1].Root != installedRoot {
		t.Fatalf("unexpected installed pack: %#v", packs[1])
	}
}

func TestCatalogLoadReturnsSortedPlaybooks(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "zzz.yaml", `
id: zzz-last
title: ZZZ Last
category: test
severity: low
match:
  any:
    - "zzz error"
`)
	writePlaybookFixture(t, dir, "aaa.yaml", `
id: aaa-first
title: AAA First
category: test
severity: low
match:
  any:
    - "aaa error"
`)

	catalog := NewCatalog(dir)
	pbs, err := catalog.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(pbs) != 2 {
		t.Fatalf("expected 2 playbooks, got %d", len(pbs))
	}
	if pbs[0].ID != "aaa-first" || pbs[1].ID != "zzz-last" {
		t.Errorf("expected alphabetical order by ID, got %s, %s", pbs[0].ID, pbs[1].ID)
	}
}

func TestCatalogListMatchesLoad(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "sample.yaml", `
id: list-sample
title: List Sample
category: build
severity: medium
match:
  any:
    - "build error"
`)

	catalog := NewCatalog(dir)
	loaded, err := catalog.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	listed, err := catalog.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(loaded) != len(listed) {
		t.Errorf("Load returned %d, List returned %d", len(loaded), len(listed))
	}
}

func TestCatalogExplainKnownID(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "sample.yaml", `
id: explain-sample
title: Explain Sample
category: auth
severity: high
match:
  any:
    - "auth error"
`)

	catalog := NewCatalog(dir)
	pb, err := catalog.Explain("explain-sample")
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	if pb.ID != "explain-sample" {
		t.Errorf("expected explain-sample, got %s", pb.ID)
	}
}

func TestCatalogExplainUnknownIDReturnsError(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "sample.yaml", `
id: known
title: Known
category: test
severity: low
match:
  any:
    - "some error"
`)

	catalog := NewCatalog(dir)
	_, err := catalog.Explain("does-not-exist-xyz")
	if err == nil {
		t.Fatal("expected error for unknown playbook ID")
	}
}

func TestExtraPackEnvKey(t *testing.T) {
	key := ExtraPackEnvKey()
	if key == "" {
		t.Error("ExtraPackEnvKey() should return a non-empty string")
	}
	if key != packsEnvKey {
		t.Errorf("ExtraPackEnvKey() = %q, want %q", key, packsEnvKey)
	}
}

func TestCatalogDirSinglePack(t *testing.T) {
	dir := t.TempDir()
	writePlaybookFixture(t, dir, "sample.yaml", `
id: dir-sample
title: Dir Sample
category: test
severity: low
match:
  any:
    - "some error"
`)

	catalog := NewCatalog(dir)
	got, err := catalog.Dir()
	if err != nil {
		t.Fatalf("Dir: %v", err)
	}
	if got != dir {
		t.Errorf("Dir() = %q, want %q", got, dir)
	}
}

func TestCatalogDirMultiplePacksReturnsError(t *testing.T) {
	bundled := t.TempDir()
	extra := t.TempDir()
	writePlaybookFixture(t, bundled, "base.yaml", `
id: bundled-dir
title: Bundled
category: test
severity: low
match:
  any:
    - "bundled error"
`)
	writePlaybookFixture(t, extra, "extra.yaml", `
id: extra-dir
title: Extra
category: test
severity: low
match:
  any:
    - "extra error"
`)

	catalog := NewCatalogWithOptions(CatalogOptions{
		OverrideDir:   "",
		ExtraPackDirs: []string{extra},
	})
	t.Setenv(envKey, bundled)

	_, err := catalog.Dir()
	if err == nil {
		t.Fatal("expected error when catalog spans multiple packs")
	}
}

func TestCleanPackDirsDeduplicates(t *testing.T) {
	dirs := []string{"a", "b", "a", "  ", "c", "b"}
	got := cleanPackDirs(dirs)
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("cleanPackDirs = %v, want %v", got, want)
	}
	for i, d := range want {
		if got[i] != d {
			t.Errorf("cleanPackDirs[%d] = %q, want %q", i, got[i], d)
		}
	}
}

func TestPackNameFromRoot(t *testing.T) {
	tests := []struct {
		root string
		want string
	}{
		{"/home/user/.faultline/packs/team-pack", "team-pack"},
		{"/some/path/my-pack", "my-pack"},
	}
	for _, tt := range tests {
		got := packNameFromRoot(tt.root)
		if got != tt.want {
			t.Errorf("packNameFromRoot(%q) = %q, want %q", tt.root, got, tt.want)
		}
	}
}
