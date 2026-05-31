// Package repo provides local git repository scanning for Faultline.
// It discovers the repository root, executes git commands, and returns
// structured data that other repo sub-packages can use for analysis.
package repo

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Scanner discovers a git repository and executes git commands within it.
type Scanner struct {
	// Root is the absolute path to the repository root (the directory
	// containing the .git entry).
	Root string
}

// NewScanner returns a Scanner rooted at the nearest git repository ancestor
// of dir. If dir is empty or ".", the current working directory is used.
// Returns an error if no git repository can be found.
func NewScanner(dir string) (*Scanner, error) {
	if dir == "" || dir == "." {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve repo path: %w", err)
	}

	root, err := findGitRoot(abs)
	if err != nil {
		return nil, err
	}
	return &Scanner{Root: root}, nil
}

// gitCommandTimeout caps how long a single git invocation may run.
const gitCommandTimeout = 30 * time.Second

// gitRootSearchDepth limits how many directory levels are walked upward when
// searching for the git repository root.
const gitRootSearchDepth = 50

// Run executes a git command with the given arguments inside the repository
// root and returns the combined stdout output. Non-zero exit codes produce an
// error that includes the stderr text. Each invocation is bounded by
// gitCommandTimeout to prevent hangs on slow or unresponsive git operations.
func (s *Scanner) Run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = s.Root
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s",
			strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// findGitRoot walks up from dir until it finds a directory that contains a
// .git entry (file or directory). Returns an error if none is found or if
// the search exceeds gitRootSearchDepth directory levels.
func findGitRoot(dir string) (string, error) {
	start := dir
	for depth := 0; depth < gitRootSearchDepth; depth++ {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf(
				"no git repository found (searched from %s upward)", start)
		}
		dir = parent
	}
	return "", fmt.Errorf(
		"no git repository found (searched from %s upward, max depth %d)", start, gitRootSearchDepth)
}
