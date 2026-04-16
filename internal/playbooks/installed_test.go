package playbooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsPlaybookFile(t *testing.T) {
	for _, name := range []string{"playbook.yaml", "rule.yml", "sub.yaml", "sub.yml"} {
		if !isPlaybookFile(name) {
			t.Errorf("isPlaybookFile(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"README.md", "main.go", "config.json", ".gitignore", "Makefile"} {
		if isPlaybookFile(name) {
			t.Errorf("isPlaybookFile(%q) = true, want false", name)
		}
	}
}

func TestCopyFileBasic(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	src := filepath.Join(srcDir, "source.yaml")
	dst := filepath.Join(dstDir, "dest.yaml")

	content := "id: test-copy\n"
	if err := os.WriteFile(src, []byte(content), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != content {
		t.Errorf("copyFile content = %q, want %q", string(got), content)
	}
}

func TestCopyFileMissingSrcReturnsError(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "out.yaml")
	err := copyFile("/nonexistent/missing.yaml", dst)
	if err == nil {
		t.Fatal("expected error for missing source file")
	}
}

func TestCopyTreeCopiesYAMLFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writePlaybookFixture(t, src, "rule.yaml", `
id: copy-tree-test
title: Copy Tree
category: test
severity: low
summary: Summary.
diagnosis: |
  ## Diagnosis

  Details.
fix: |
  ## Fix steps

  1. Fix.
validation: |
  ## Validation

  - Check.
match:
  any:
    - "tree error"
`)
	if err := os.WriteFile(filepath.Join(src, "README.md"), []byte("# readme\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	if err := copyTree(src, dst); err != nil {
		t.Fatalf("copyTree: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, "rule.yaml")); err != nil {
		t.Errorf("expected rule.yaml in dst, got: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "README.md")); err == nil {
		t.Error("README.md should not have been copied (not a playbook file)")
	}
}

func TestCopyTreeSkipsGitDir(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	gitDir := filepath.Join(src, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "config.yaml"), []byte("not-a-playbook\n"), 0o644); err != nil {
		t.Fatalf("write .git/config.yaml: %v", err)
	}
	writePlaybookFixture(t, src, "real.yaml", `
id: real-playbook
title: Real
category: test
severity: low
summary: Summary.
diagnosis: |
  ## Diagnosis

  Details.
fix: |
  ## Fix steps

  1. Fix.
validation: |
  ## Validation

  - Check.
match:
  any:
    - "real error"
`)

	if err := copyTree(src, dst); err != nil {
		t.Fatalf("copyTree: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, ".git")); err == nil {
		t.Error(".git directory should not have been copied")
	}
	if _, err := os.Stat(filepath.Join(dst, "real.yaml")); err != nil {
		t.Errorf("expected real.yaml in dst, got: %v", err)
	}
}

func TestDiscoverInstalledPackRootsEmptyDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	roots, err := DiscoverInstalledPackRoots()
	if err != nil {
		t.Fatalf("DiscoverInstalledPackRoots: %v", err)
	}
	if len(roots) != 0 {
		t.Errorf("expected empty roots, got %v", roots)
	}
}

func TestDiscoverInstalledPackRootsFindsPackDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	packDir := filepath.Join(home, ".faultline", installedPacksSubdir, "my-pack")
	if err := os.MkdirAll(packDir, 0o755); err != nil {
		t.Fatalf("mkdir pack dir: %v", err)
	}

	roots, err := DiscoverInstalledPackRoots()
	if err != nil {
		t.Fatalf("DiscoverInstalledPackRoots: %v", err)
	}
	if len(roots) != 1 || roots[0] != packDir {
		t.Errorf("expected [%s], got %v", packDir, roots)
	}
}

func TestListInstalledPacksNoPacksInstalled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	packs, err := ListInstalledPacks()
	if err != nil {
		t.Fatalf("ListInstalledPacks: %v", err)
	}
	if len(packs) != 0 {
		t.Errorf("expected no installed packs, got %v", packs)
	}
}

func TestInstallPackBasic(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	src := t.TempDir()
	writePlaybookFixture(t, src, "rule.yaml", `
id: install-test
title: Install Test
category: test
severity: low
summary: Summary.
diagnosis: |
  ## Diagnosis

  Details.
fix: |
  ## Fix steps

  1. Do the thing.
validation: |
  ## Validation

  - Check.
match:
  any:
    - "install error"
`)

	pack, err := InstallPack(src, "my-test-pack", false)
	if err != nil {
		t.Fatalf("InstallPack: %v", err)
	}
	if pack.Name != "my-test-pack" {
		t.Errorf("Name = %q, want %q", pack.Name, "my-test-pack")
	}
	if pack.PlaybookCount != 1 {
		t.Errorf("PlaybookCount = %d, want 1", pack.PlaybookCount)
	}
	if _, err := os.Stat(filepath.Join(pack.Root, "rule.yaml")); err != nil {
		t.Errorf("expected rule.yaml at install root: %v", err)
	}
}

func TestInstallPackAlreadyExistsReturnsError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	src := t.TempDir()
	writePlaybookFixture(t, src, "rule.yaml", `
id: dup-install
title: Dup
category: test
severity: low
summary: Summary.
diagnosis: |
  ## Diagnosis

  Details.
fix: |
  ## Fix steps

  1. Fix.
validation: |
  ## Validation

  - Check.
match:
  any:
    - "dup error"
`)

	if _, err := InstallPack(src, "dup-pack", false); err != nil {
		t.Fatalf("first InstallPack: %v", err)
	}
	_, err := InstallPack(src, "dup-pack", false)
	if err == nil {
		t.Fatal("expected error when pack already exists without --force")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got %v", err)
	}
}

func TestInstallPackForceReplaces(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	src := t.TempDir()
	writePlaybookFixture(t, src, "rule.yaml", `
id: force-install
title: Force Install
category: test
severity: low
summary: Summary.
diagnosis: |
  ## Diagnosis

  Details.
fix: |
  ## Fix steps

  1. Fix.
validation: |
  ## Validation

  - Check.
match:
  any:
    - "force error"
`)

	if _, err := InstallPack(src, "force-pack", false); err != nil {
		t.Fatalf("first install: %v", err)
	}
	pack, err := InstallPack(src, "force-pack", true)
	if err != nil {
		t.Fatalf("force reinstall: %v", err)
	}
	if pack.Name != "force-pack" {
		t.Errorf("Name = %q, want force-pack", pack.Name)
	}
}

func TestInstallPackInvalidNameReturnsError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	src := t.TempDir()
	writePlaybookFixture(t, src, "rule.yaml", `
id: invalid-name-test
title: Invalid
category: test
severity: low
summary: Summary.
diagnosis: |
  ## Diagnosis

  Details.
fix: |
  ## Fix steps

  1. Fix.
validation: |
  ## Validation

  - Check.
match:
  any:
    - "some error"
`)

	for _, bad := range []string{".", "..", "a/b", `a\b`} {
		_, err := InstallPack(src, bad, false)
		if err == nil {
			t.Errorf("expected error for invalid pack name %q", bad)
		}
	}
}

func TestInstallPackUsesDirectoryNameWhenNoName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	src := t.TempDir()
	writePlaybookFixture(t, src, "rule.yaml", `
id: autoname-test
title: Autoname
category: test
severity: low
summary: Summary.
diagnosis: |
  ## Diagnosis

  Details.
fix: |
  ## Fix steps

  1. Fix.
validation: |
  ## Validation

  - Check.
match:
  any:
    - "autoname error"
`)

	// Pass empty name so pack name is derived from directory name.
	pack, err := InstallPack(src, "", false)
	if err != nil {
		t.Fatalf("InstallPack with empty name: %v", err)
	}
	if pack.Name == "" {
		t.Error("expected non-empty pack name when name is derived from directory")
	}
}
