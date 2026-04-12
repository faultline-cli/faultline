package playbooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestOptional(t *testing.T) {
	dir := t.TempDir()
	_, ok, err := loadManifest(dir)
	if err != nil {
		t.Fatalf("loadManifest: %v", err)
	}
	if ok {
		t.Fatal("expected missing manifest to be optional")
	}
}

func TestLoadManifestValidatesRequiredFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, packManifestFile)
	if err := os.WriteFile(path, []byte("name: starter\n"), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	_, _, err := loadManifest(dir)
	if err == nil {
		t.Fatal("expected manifest validation error")
	}
	if !strings.Contains(err.Error(), "version") {
		t.Fatalf("expected version validation error, got %v", err)
	}
}

func TestLoadPacksValidatesManifestDetectorCoverage(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, packManifestFile), []byte(strings.TrimSpace(`
name: source-only
version: 1.0.0
detectors:
  - source
`)+"\n"), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	writePlaybookFixture(t, dir, "sample.yaml", `
id: log-sample
title: Log Sample
category: test
severity: low
match:
  any:
    - "primary error"
`)

	_, err := LoadPacks([]Pack{{Root: dir}})
	if err == nil {
		t.Fatal("expected detector coverage error")
	}
	if !strings.Contains(err.Error(), "does not allow detector") {
		t.Fatalf("expected detector coverage error, got %v", err)
	}
}
