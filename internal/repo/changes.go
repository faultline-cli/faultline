package repo

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"faultline/internal/detectors"
)

// LoadWorktreeChangeSet returns the current local worktree changes in a stable
// detector.ChangeSet shape. Line-level data is intentionally omitted; source
// detection treats listed files as changed scopes.
func LoadWorktreeChangeSet(s *Scanner) (detectors.ChangeSet, error) {
	out, err := s.Run("status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return detectors.ChangeSet{}, fmt.Errorf("git status: %w", err)
	}
	changeSet := detectors.ChangeSet{ChangedFiles: map[string]detectors.FileChange{}}
	if strings.TrimSpace(out) == "" {
		return changeSet, nil
	}
	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" || len(line) < 4 {
			continue
		}
		status := strings.TrimSpace(line[:2])
		path := strings.TrimSpace(line[3:])
		if strings.Contains(path, " -> ") {
			parts := strings.Split(path, " -> ")
			path = parts[len(parts)-1]
		}
		path = filepath.ToSlash(path)
		if path == "" {
			continue
		}
		changeSet.ChangedFiles[path] = detectors.FileChange{
			Status: normalizeChangeStatus(status),
			Lines:  map[int]struct{}{},
		}
	}
	return changeSet, nil
}

func normalizeChangeStatus(status string) string {
	status = strings.TrimSpace(status)
	switch {
	case strings.Contains(status, "?"), strings.Contains(status, "A"):
		return "added"
	case strings.Contains(status, "D"):
		return "deleted"
	default:
		return "modified"
	}
}

func SortedChangedFiles(changeSet detectors.ChangeSet) []string {
	files := make([]string, 0, len(changeSet.ChangedFiles))
	for file := range changeSet.ChangedFiles {
		files = append(files, file)
	}
	sort.Strings(files)
	return files
}
