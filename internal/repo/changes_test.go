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
	if len(changeSet.ChangedFiles["main.go"].Lines) == 0 {
		t.Fatalf("expected line-level diff info for main.go, got %#v", changeSet.ChangedFiles["main.go"])
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

func TestLoadWorktreeChangeSetIncludesStagedAndUnstagedLineRanges(t *testing.T) {
	dir := initTempRepo(t)
	mainPath := filepath.Join(dir, "main.go")
	utilPath := filepath.Join(dir, "internal", "util.go")
	if err := os.MkdirAll(filepath.Dir(utilPath), 0o755); err != nil {
		t.Fatalf("mkdir util dir: %v", err)
	}
	if err := os.WriteFile(mainPath, []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(utilPath, []byte("package internal\n\nfunc Util() string {\n\treturn \"ok\"\n}\n"), 0o644); err != nil {
		t.Fatalf("write util.go: %v", err)
	}
	runGit(t, dir, nil, "add", ".")
	runGit(t, dir, nil, "commit", "--quiet", "-m", "baseline")

	if err := os.WriteFile(mainPath, []byte("package main\n\nfunc main() {\n\tprintln(\"changed\")\n}\n"), 0o644); err != nil {
		t.Fatalf("rewrite main.go: %v", err)
	}
	if err := os.WriteFile(utilPath, []byte("package internal\n\nfunc Util() string {\n\tprintln(\"staged\")\n\treturn \"ok\"\n}\n"), 0o644); err != nil {
		t.Fatalf("rewrite util.go: %v", err)
	}
	runGit(t, dir, nil, "add", "internal/util.go")

	scanner, err := NewScanner(dir)
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	changeSet, err := LoadWorktreeChangeSet(scanner)
	if err != nil {
		t.Fatalf("LoadWorktreeChangeSet: %v", err)
	}

	if !hasLine(changeSet.ChangedFiles["main.go"].Lines, 4) {
		t.Fatalf("expected unstaged line 4 in main.go diff, got %#v", changeSet.ChangedFiles["main.go"])
	}
	if !hasLine(changeSet.ChangedFiles["internal/util.go"].Lines, 4) {
		t.Fatalf("expected staged line 4 in internal/util.go diff, got %#v", changeSet.ChangedFiles["internal/util.go"])
	}
}

func TestChangeSetRelativeToScopesToSubdirectory(t *testing.T) {
	changeSet := detectors.ChangeSet{
		ChangedFiles: map[string]detectors.FileChange{
			"api/handler.go": {
				Status: "added",
				Lines:  map[int]struct{}{3: {}},
			},
			"deploy/healthcheck.yaml": {
				Status: "modified",
				Lines:  map[int]struct{}{1: {}},
			},
		},
	}

	trimmed := ChangeSetRelativeTo(changeSet, "api")
	if len(trimmed.ChangedFiles) != 1 {
		t.Fatalf("expected 1 scoped file, got %#v", trimmed.ChangedFiles)
	}
	change, ok := trimmed.ChangedFiles["handler.go"]
	if !ok {
		t.Fatalf("expected handler.go after rebasing, got %#v", trimmed.ChangedFiles)
	}
	if change.Status != "added" || !hasLine(change.Lines, 3) {
		t.Fatalf("expected rebased change metadata to survive, got %#v", change)
	}
}

func hasLine(lines map[int]struct{}, want int) bool {
	_, ok := lines[want]
	return ok
}

// ── cloneChangeSet ────────────────────────────────────────────────────────────

func TestCloneChangeSetCopiesAllFiles(t *testing.T) {
	in := detectors.ChangeSet{
		ChangedFiles: map[string]detectors.FileChange{
			"main.go": {
				Status: "modified",
				Lines:  map[int]struct{}{3: {}, 7: {}},
			},
			"util.go": {
				Status: "added",
				Lines:  map[int]struct{}{1: {}},
			},
		},
	}
	out := cloneChangeSet(in)
	if len(out.ChangedFiles) != 2 {
		t.Fatalf("expected 2 files, got %d", len(out.ChangedFiles))
	}
	if out.ChangedFiles["main.go"].Status != "modified" {
		t.Errorf("expected status modified for main.go, got %q", out.ChangedFiles["main.go"].Status)
	}
	if !hasLine(out.ChangedFiles["main.go"].Lines, 3) || !hasLine(out.ChangedFiles["main.go"].Lines, 7) {
		t.Errorf("expected lines 3 and 7 for main.go, got %v", out.ChangedFiles["main.go"].Lines)
	}
}

func TestCloneChangeSetIsIndependent(t *testing.T) {
	in := detectors.ChangeSet{
		ChangedFiles: map[string]detectors.FileChange{
			"app.go": {
				Status: "added",
				Lines:  map[int]struct{}{1: {}},
			},
		},
	}
	out := cloneChangeSet(in)
	// Mutate original
	in.ChangedFiles["app.go"] = detectors.FileChange{Status: "deleted", Lines: map[int]struct{}{}}
	// Clone should not be affected
	if out.ChangedFiles["app.go"].Status != "added" {
		t.Errorf("expected clone to be independent of original, got status %q", out.ChangedFiles["app.go"].Status)
	}
}

func TestCloneChangeSetEmptyInput(t *testing.T) {
	in := detectors.ChangeSet{ChangedFiles: map[string]detectors.FileChange{}}
	out := cloneChangeSet(in)
	if len(out.ChangedFiles) != 0 {
		t.Errorf("expected empty clone, got %v", out.ChangedFiles)
	}
}
