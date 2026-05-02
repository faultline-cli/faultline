package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRelPlaybookLinkFromCategoryPage(t *testing.T) {
	got := relPlaybookLink("playbooks/bundled/log/auth/aws-credentials.yaml", "auth")
	want := "../../../playbooks/bundled/log/auth/aws-credentials.yaml"
	if got != want {
		t.Fatalf("relPlaybookLink() = %q, want %q", got, want)
	}
}

func TestRelPlaybookLinkFromNestedCategoryPage(t *testing.T) {
	got := relPlaybookLink("playbooks/bundled/log/silent/artifact-missing.yaml", "silent/failure")
	want := "../../../../playbooks/bundled/log/silent/artifact-missing.yaml"
	if got != want {
		t.Fatalf("relPlaybookLink() = %q, want %q", got, want)
	}
}

func TestCheckFilesReportsOrphanedGeneratedDoc(t *testing.T) {
	dir := t.TempDir()
	files := []generatedFile{
		{RelPath: "build/missing-executable.md", Content: []byte("current\n")},
		{RelPath: "catalog/README.md", Content: []byte("catalog\n")},
	}
	writeDoc(t, dir, "build/missing-executable.md", "current\n")
	writeDoc(t, dir, "catalog/README.md", "catalog\n")
	writeDoc(t, dir, "build/removed-playbook.md", "stale\n")
	writeDoc(t, dir, "README.md", "manual\n")

	stale, err := checkFiles(dir, files)
	if err != nil {
		t.Fatalf("checkFiles: %v", err)
	}
	got := strings.Join(stale, "\n")
	if !strings.Contains(got, "build/removed-playbook.md (stale generated doc)") {
		t.Fatalf("expected stale generated doc, got %q", got)
	}
	if strings.Contains(got, "README.md (stale generated doc)") {
		t.Fatalf("manual README should not be stale, got %q", got)
	}
}

func TestWriteFilesPrunesOrphanedGeneratedDoc(t *testing.T) {
	dir := t.TempDir()
	files := []generatedFile{
		{RelPath: "build/missing-executable.md", Content: []byte("current\n")},
		{RelPath: "catalog/README.md", Content: []byte("catalog\n")},
	}
	writeDoc(t, dir, "build/removed-playbook.md", "stale\n")
	writeDoc(t, dir, "README.md", "manual\n")

	if err := writeFiles(dir, files); err != nil {
		t.Fatalf("writeFiles: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "build", "removed-playbook.md")); !os.IsNotExist(err) {
		t.Fatalf("expected stale doc to be removed, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "README.md")); err != nil {
		t.Fatalf("manual README should remain: %v", err)
	}
}

func writeDoc(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
