package repo

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewScannerWithExplicitDir verifies that NewScanner resolves the root
// when given a path that is itself the git repository root.
func TestNewScannerWithExplicitDir(t *testing.T) {
	t.Parallel()

	dir := initTempRepo(t)
	s, err := NewScanner(dir)
	if err != nil {
		t.Fatalf("NewScanner(%q): %v", dir, err)
	}
	if s.Root != dir {
		t.Errorf("Scanner.Root = %q, want %q", s.Root, dir)
	}
}

// TestNewScannerWalksUpToRoot verifies the upward-walking logic: when given a
// subdirectory nested inside a git repo, the scanner should resolve to the
// repo root.
func TestNewScannerWalksUpToRoot(t *testing.T) {
	t.Parallel()

	root := initTempRepo(t)
	sub := filepath.Join(root, "internal", "pkg", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", sub, err)
	}

	s, err := NewScanner(sub)
	if err != nil {
		t.Fatalf("NewScanner(%q): %v", sub, err)
	}
	if s.Root != root {
		t.Errorf("Scanner.Root = %q, want %q (repo root)", s.Root, root)
	}
}

// TestNewScannerErrorsOutsideGitRepo verifies that NewScanner returns a
// descriptive error when no .git entry can be found anywhere in the hierarchy.
func TestNewScannerErrorsOutsideGitRepo(t *testing.T) {
	t.Parallel()

	// A plain temp dir has no .git directory.
	dir := t.TempDir()
	_, err := NewScanner(dir)
	if err == nil {
		t.Fatal("NewScanner expected error for non-git directory, got nil")
	}
}

// TestNewScannerEmptyDirUsesCwd verifies that passing "" falls back to the
// working directory.  The test process runs inside the faultline repo so the
// call should succeed.
func TestNewScannerEmptyDirUsesCwd(t *testing.T) {
	t.Parallel()

	s, err := NewScanner("")
	if err != nil {
		t.Fatalf("NewScanner(%q): %v", "", err)
	}
	if s.Root == "" {
		t.Error("Scanner.Root is empty, want non-empty path")
	}
	// Root must contain a .git entry.
	if _, statErr := os.Stat(filepath.Join(s.Root, ".git")); statErr != nil {
		t.Errorf("Scanner.Root %q has no .git: %v", s.Root, statErr)
	}
}

// TestNewScannerDotUsesCwd verifies that passing "." behaves identically to
// passing "".
func TestNewScannerDotUsesCwd(t *testing.T) {
	t.Parallel()

	s, err := NewScanner(".")
	if err != nil {
		t.Fatalf("NewScanner(%q): %v", ".", err)
	}
	if s.Root == "" {
		t.Error("Scanner.Root is empty")
	}
}

// TestScannerRunReturnsOutput verifies that Scanner.Run executes git commands
// correctly inside the repository root.
func TestScannerRunReturnsOutput(t *testing.T) {
	t.Parallel()

	dir := initTempRepo(t)
	s, err := NewScanner(dir)
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}

	out, err := s.Run("rev-parse", "--git-dir")
	if err != nil {
		t.Fatalf("Scanner.Run: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty output from git rev-parse --git-dir")
	}
}

// TestScannerRunErrorOnBadArgs verifies that a bad git command produces an
// error (non-zero exit) rather than silently succeeding.
func TestScannerRunErrorOnBadArgs(t *testing.T) {
	t.Parallel()

	dir := initTempRepo(t)
	s, err := NewScanner(dir)
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}

	_, err = s.Run("not-a-real-git-command")
	if err == nil {
		t.Error("expected error for invalid git subcommand, got nil")
	}
}
