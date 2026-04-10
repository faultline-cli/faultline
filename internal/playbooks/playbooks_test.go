package playbooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func writePlaybookFixture(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o600); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}
