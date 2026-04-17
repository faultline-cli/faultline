package playbooks

import (
	"strings"
	"testing"
)

func TestUpwardDirsIncludesBundledAndLegacyPaths(t *testing.T) {
	dirs := upwardDirs("/some/project/sub")
	if len(dirs) == 0 {
		t.Fatal("expected at least one candidate dir")
	}
	foundBundled := false
	foundLegacy := false
	for _, dir := range dirs {
		if strings.HasSuffix(dir, "playbooks/bundled") {
			foundBundled = true
		}
		if strings.HasSuffix(dir, "/playbooks") && !strings.HasSuffix(dir, "playbooks/bundled") {
			foundLegacy = true
		}
	}
	if !foundBundled {
		t.Errorf("expected at least one 'playbooks/bundled' candidate, got %v", dirs)
	}
	if !foundLegacy {
		t.Errorf("expected at least one legacy 'playbooks' candidate, got %v", dirs)
	}
}

func TestUpwardDirsTerminatesAtFilesystemRoot(t *testing.T) {
	// Call with a shallow path — should still terminate and return a finite list.
	dirs := upwardDirs("/")
	if len(dirs) == 0 {
		t.Fatal("expected at least one candidate for root dir")
	}
	// All returned paths should be absolute
	for _, dir := range dirs {
		if !strings.HasPrefix(dir, "/") {
			t.Errorf("expected absolute path, got %q", dir)
		}
	}
}

func TestUpwardDirsIncludesAllAncestors(t *testing.T) {
	dirs := upwardDirs("/a/b/c")
	// Should include candidates for /a/b/c, /a/b, /a, /
	// (2 entries per level × at least 4 levels = at least 8)
	if len(dirs) < 8 {
		t.Errorf("expected at least 8 candidates for 4-level path, got %d: %v", len(dirs), dirs)
	}
}
