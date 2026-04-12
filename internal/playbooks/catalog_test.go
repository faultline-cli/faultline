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
