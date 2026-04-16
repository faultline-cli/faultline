package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadWorktreeChangeSet(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Faultline Test")
	runGit(t, dir, "config", "user.email", "faultline@example.com")
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "--quiet", "-m", "baseline")

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

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}
