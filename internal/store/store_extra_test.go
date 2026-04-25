package store

import (
	"os"
	"path/filepath"
	"testing"
)

// --- DefaultPath ---

func TestDefaultPathReturnsPathUnderHome(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	home, _ := os.UserHomeDir()
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}
	// Should be under the user's home directory
	rel, err := filepath.Rel(home, path)
	if err != nil || rel == "" {
		t.Errorf("expected path under home dir %q, got %q", home, path)
	}
}

// --- OpenBestEffort strict mode ---

func TestOpenBestEffortStrictModeCorruptFileErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.db")
	if err := os.WriteFile(path, []byte("not-sqlite"), 0o600); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}
	_, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path, Strict: true})
	if err == nil {
		t.Fatal("expected error in strict mode for corrupt db")
	}
}

func TestOpenBestEffortNonStrictCorruptFileDegrades(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.db")
	if err := os.WriteFile(path, []byte("not-sqlite"), 0o600); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}
	st, info, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path, Strict: false})
	if err != nil {
		t.Fatalf("unexpected error in non-strict mode: %v", err)
	}
	defer st.Close()
	if !info.Degraded {
		t.Errorf("expected degraded=true, got %#v", info)
	}
	if info.Warning == "" {
		t.Error("expected non-empty warning in degraded info")
	}
}

func TestOpenBestEffortAutoModeWithValidPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "good.db")
	st, info, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()
	if info.Mode != ModeAuto {
		t.Errorf("expected auto mode, got %v", info.Mode)
	}
	if info.Backend != "sqlite" {
		t.Errorf("expected sqlite backend, got %q", info.Backend)
	}
	if info.Path != path {
		t.Errorf("expected path %q, got %q", path, info.Path)
	}
	if info.Degraded {
		t.Error("expected non-degraded for valid path")
	}
}

// --- ResolveConfig whitespace handling ---

func TestResolveConfigWhitespaceTrimmed(t *testing.T) {
	// Whitespace-only should be treated as empty → auto
	cfg, err := ResolveConfig("   ", false)
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Mode != ModeAuto {
		t.Errorf("expected auto for whitespace, got %v", cfg.Mode)
	}
	if cfg.Path != "" {
		t.Errorf("expected empty path for whitespace, got %q", cfg.Path)
	}
}

func TestResolveConfigDisableOverridesExplicitPath(t *testing.T) {
	cfg, err := ResolveConfig("/some/path.db", true)
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Mode != ModeOff {
		t.Errorf("expected off when disable=true, got %v", cfg.Mode)
	}
	if cfg.Path != "" {
		t.Errorf("expected empty path when disable=true, got %q", cfg.Path)
	}
}
