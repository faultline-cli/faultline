package playbooks

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const installedPacksSubdir = "packs"

// InstalledPack describes one user-installed playbook pack.
type InstalledPack struct {
	Name          string
	Root          string
	PlaybookCount int
	Version       string // from faultline-pack.yaml; empty when not recorded
	PinnedRef     string // git ref at install time; empty when not available
	SourceURL     string // install-time source path or URL; empty when not recorded
}

// InstalledPackRoot returns the per-user directory used for installed packs.
func InstalledPackRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".faultline", installedPacksSubdir), nil
}

// DiscoverInstalledPackRoots returns installed pack directories in stable order.
func DiscoverInstalledPackRoots() ([]string, error) {
	root, err := InstalledPackRoot()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read installed pack directory: %w", err)
	}
	roots := make([]string, 0, len(entries))
	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		info, statErr := os.Stat(path)
		if statErr != nil {
			continue
		}
		if !info.IsDir() {
			continue
		}
		roots = append(roots, path)
	}
	sort.Strings(roots)
	return roots, nil
}

// ListInstalledPacks returns user-installed packs with playbook counts and provenance.
func ListInstalledPacks() ([]InstalledPack, error) {
	roots, err := DiscoverInstalledPackRoots()
	if err != nil {
		return nil, err
	}
	packs := make([]InstalledPack, 0, len(roots))
	for _, root := range roots {
		pbs, loadErr := LoadDir(root)
		if loadErr != nil {
			return nil, fmt.Errorf("load installed pack %q: %w", filepath.Base(root), loadErr)
		}
		meta, _, _ := ReadPackMeta(root) // best-effort; missing manifest is fine
		packs = append(packs, InstalledPack{
			Name:          filepath.Base(root),
			Root:          root,
			PlaybookCount: len(pbs),
			Version:       meta.Version,
			PinnedRef:     meta.PinnedRef,
			SourceURL:     meta.SourceURL,
		})
	}
	return packs, nil
}

// InstallPack copies a pack into the per-user install directory so it can be
// auto-loaded on future runs and survive binary upgrades.
func InstallPack(srcDir, name string, force bool) (InstalledPack, error) {
	root, err := validateDir(srcDir)
	if err != nil {
		return InstalledPack{}, err
	}
	pbs, err := LoadDir(root)
	if err != nil {
		return InstalledPack{}, err
	}
	installRoot, err := InstalledPackRoot()
	if err != nil {
		return InstalledPack{}, err
	}
	if err := os.MkdirAll(installRoot, 0o700); err != nil {
		return InstalledPack{}, fmt.Errorf("create installed pack directory: %w", err)
	}
	packName := strings.TrimSpace(name)
	if packName == "" {
		packName = packNameFromRoot(root)
	}
	if packName == "" {
		return InstalledPack{}, fmt.Errorf("could not determine installed pack name for %q", root)
	}
	if packName == "." || packName == ".." || strings.ContainsAny(packName, `/\`) {
		return InstalledPack{}, fmt.Errorf("invalid installed pack name %q", packName)
	}
	if packName != filepath.Base(packName) {
		return InstalledPack{}, fmt.Errorf("invalid installed pack name %q", packName)
	}
	dest := filepath.Join(installRoot, packName)
	if _, err := os.Stat(dest); err == nil {
		if !force {
			return InstalledPack{}, fmt.Errorf("installed pack %q already exists; re-run with --force to replace it", packName)
		}
		if err := os.RemoveAll(dest); err != nil {
			return InstalledPack{}, fmt.Errorf("replace installed pack %q: %w", packName, err)
		}
	} else if !os.IsNotExist(err) {
		return InstalledPack{}, fmt.Errorf("check installed pack %q: %w", packName, err)
	}
	// Read provenance from the source pack if it has a manifest.
	srcMeta, _, _ := ReadPackMeta(root) // best-effort

	if err := copyTree(root, dest); err != nil {
		return InstalledPack{}, err
	}
	if _, err := LoadDir(dest); err != nil {
		_ = os.RemoveAll(dest)
		return InstalledPack{}, fmt.Errorf("validate installed pack %q: %w", packName, err)
	}

	// Derive pinned ref from the source directory's git HEAD if available.
	pinnedRef := resolveGitHead(root)

	// Inherit version from the source manifest if present; mark local installs explicitly.
	version := srcMeta.Version
	if version == "" {
		version = "0.0.0+local"
	}

	// Absolute source path is the canonical install-time source URL for local directories.
	sourceURL := srcMeta.SourceURL
	if sourceURL == "" {
		abs, absErr := filepath.Abs(root)
		if absErr == nil {
			sourceURL = abs
		}
	}

	installMeta := PackMeta{
		Name:      packName,
		Version:   version,
		SourceURL: sourceURL,
		PinnedRef: pinnedRef,
	}
	// Best-effort: a manifest write failure should not prevent the pack from loading.
	_ = WritePackMeta(dest, installMeta)

	return InstalledPack{
		Name:          packName,
		Root:          dest,
		PlaybookCount: len(pbs),
		Version:       installMeta.Version,
		PinnedRef:     installMeta.PinnedRef,
		SourceURL:     installMeta.SourceURL,
	}, nil
}

// resolveGitHead returns the short SHA of HEAD in dir, or an empty string
// when dir is not a git repository or git is not available.
func resolveGitHead(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func copyTree(src, dest string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("resolve installed pack path: %w", err)
		}
		target := dest
		if rel != "." {
			target = filepath.Join(dest, rel)
		}
		if d.IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("create installed pack directory: %w", err)
			}
			return nil
		}
		if !isPlaybookFile(d.Name()) {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("install pack: symlinks are not supported (%s)", path)
		}
		return copyFile(path, target)
	})
}

func isPlaybookFile(name string) bool {
	if name == PackMetaFileName {
		return false
	}
	return strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("create %s: %w", dest, err)
	}
	defer func() {
		if out != nil {
			_ = out.Close()
		}
	}()
	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %s: %w", src, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close %s: %w", dest, err)
	}
	out = nil
	return nil
}
