package repo

import (
	"os"
	"path/filepath"
	"testing"

	"faultline/internal/detectors"
)

func TestLoadWorktreeChangeSet(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, nil, "init")
	runGit(t, dir, nil, "config", "user.name", "Faultline Test")
	runGit(t, dir, nil, "config", "user.email", "faultline@example.com")
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, dir, nil, "add", ".")
	runGit(t, dir, nil, "commit", "--quiet", "-m", "baseline")

	if err := os.WriteFile(path, []byte("package main\n\nfunc main(){}\n"), 0o644); err != nil {
		t.Fatalf("rewrite file: %v", err)
	}
	scanner, err := NewScanner(dir)
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	changeSet, err := LoadWorktreeChangeSet(scanner)
	if err != nil {
		t.Fatalf("LoadWorktreeChangeSet: %v", err)
	}
	if _, ok := changeSet.ChangedFiles["main.go"]; !ok {
		t.Fatalf("expected main.go in change set, got %#v", changeSet.ChangedFiles)
	}
}

func TestSortedChangedFilesReturnsAlphabeticOrder(t *testing.T) {
	changeSet := detectors.ChangeSet{
		ChangedFiles: map[string]detectors.FileChange{
			"z.go": {Status: "modified"},
			"a.go": {Status: "modified"},
			"m.go": {Status: "added"},
		},
	}
	files := SortedChangedFiles(changeSet)
	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}
	if files[0] != "a.go" || files[1] != "m.go" || files[2] != "z.go" {
		t.Fatalf("expected sorted order [a.go m.go z.go], got %v", files)
	}
}

func TestSortedChangedFilesEmptyChangeSet(t *testing.T) {
	changeSet := detectors.ChangeSet{ChangedFiles: map[string]detectors.FileChange{}}
	files := SortedChangedFiles(changeSet)
	if len(files) != 0 {
		t.Fatalf("expected empty slice for empty changeset, got %v", files)
	}
}

func TestNormalizeChangeStatusClassifiesCorrectly(t *testing.T) {
	cases := []struct {
		status string
		want   string
	}{
		{"A", "added"},
		{"?", "added"},
		{"??", "added"},
		{"AM", "added"},
		{"D", "deleted"},
		{"MD", "deleted"},
		{"M", "modified"},
		{"MM", "modified"},
		{"R", "modified"},
		{"", "modified"},
	}
	for _, tc := range cases {
		if got := normalizeChangeStatus(tc.status); got != tc.want {
			t.Errorf("normalizeChangeStatus(%q) = %q, want %q", tc.status, got, tc.want)
		}
	}
}
