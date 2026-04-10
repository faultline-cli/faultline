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

func TestFindPatternConflictsBundled(t *testing.T) {
	pbs, err := LoadDir("../../playbooks")
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}

	conflicts := FindPatternConflicts(pbs)
	if len(conflicts) == 0 {
		t.Fatal("expected bundled playbooks to produce at least one pattern conflict report")
	}

	assertConflict(t, conflicts, "context deadline exceeded", []string{"network-timeout", "test-timeout"}, nil)
	assertConflict(t, conflicts, "connection refused", []string{"connection-refused", "gradle-build"}, nil)
	assertConflict(t, conflicts, "resource_class", []string{"circleci-resource-class"}, []string{"container-crash", "oom-killed"})

	report := FormatPatternConflicts(conflicts)
	if !strings.Contains(report, "context deadline exceeded") {
		t.Fatalf("expected conflict report to include context deadline exceeded, got %q", report)
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
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o600); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}
