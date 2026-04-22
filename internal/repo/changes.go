package repo

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"faultline/internal/detectors"
)

var diffHunkPattern = regexp.MustCompile(`@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@`)

// LoadWorktreeChangeSet returns the current local worktree changes in a stable
// detector.ChangeSet shape. When git diff hunks are available, changed line
// numbers are attached so source detection can distinguish introduced versus
// legacy findings more precisely.
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
	for _, args := range [][]string{
		{"diff", "--no-ext-diff", "--unified=0"},
		{"diff", "--cached", "--no-ext-diff", "--unified=0"},
	} {
		lineMap, err := loadDiffLineMap(s, args...)
		if err != nil {
			return detectors.ChangeSet{}, err
		}
		for path, lines := range lineMap {
			change := changeSet.ChangedFiles[path]
			if change.Lines == nil {
				change.Lines = map[int]struct{}{}
			}
			for line := range lines {
				change.Lines[line] = struct{}{}
			}
			if change.Status == "" {
				change.Status = "modified"
			}
			changeSet.ChangedFiles[path] = change
		}
	}
	return changeSet, nil
}

func loadDiffLineMap(s *Scanner, args ...string) (map[string]map[int]struct{}, error) {
	out, err := s.Run(args...)
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	linesByFile := map[string]map[int]struct{}{}
	currentPath := ""
	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "+++ "):
			currentPath = diffPath(strings.TrimSpace(strings.TrimPrefix(line, "+++ ")))
			if currentPath != "" {
				if _, ok := linesByFile[currentPath]; !ok {
					linesByFile[currentPath] = map[int]struct{}{}
				}
			}
		case strings.HasPrefix(line, "@@"):
			start, count, ok := parseDiffHunk(line)
			if !ok || currentPath == "" || count <= 0 {
				continue
			}
			fileLines := linesByFile[currentPath]
			if fileLines == nil {
				fileLines = map[int]struct{}{}
				linesByFile[currentPath] = fileLines
			}
			for n := 0; n < count; n++ {
				fileLines[start+n] = struct{}{}
			}
		}
	}
	return linesByFile, nil
}

func diffPath(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case "", "/dev/null":
		return ""
	}
	value = strings.TrimPrefix(value, "a/")
	value = strings.TrimPrefix(value, "b/")
	return filepath.ToSlash(value)
}

func parseDiffHunk(line string) (int, int, bool) {
	matches := diffHunkPattern.FindStringSubmatch(strings.TrimSpace(line))
	if len(matches) != 3 {
		return 0, 0, false
	}
	start := parsePositiveInt(matches[1], 0)
	if start == 0 {
		return 0, 0, false
	}
	count := 1
	if strings.TrimSpace(matches[2]) != "" {
		count = parsePositiveInt(matches[2], -1)
	}
	if count < 0 {
		return 0, 0, false
	}
	return start, count, true
}

func parsePositiveInt(value string, fallback int) int {
	n := 0
	for _, r := range strings.TrimSpace(value) {
		if r < '0' || r > '9' {
			return fallback
		}
		n = (n * 10) + int(r-'0')
	}
	if n <= 0 {
		return fallback
	}
	return n
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

// ChangeSetRelativeTo trims changed-file paths down to the given repository
// subdirectory. Files outside the subdirectory are dropped.
func ChangeSetRelativeTo(changeSet detectors.ChangeSet, prefix string) detectors.ChangeSet {
	prefix = filepath.ToSlash(strings.TrimSpace(prefix))
	if prefix == "" || prefix == "." {
		return cloneChangeSet(changeSet)
	}
	prefix = strings.Trim(prefix, "/")
	out := detectors.ChangeSet{ChangedFiles: map[string]detectors.FileChange{}}
	for file, change := range changeSet.ChangedFiles {
		file = filepath.ToSlash(strings.TrimSpace(file))
		if file == prefix {
			out.ChangedFiles["."] = cloneFileChange(change)
			continue
		}
		matchPrefix := prefix + "/"
		if !strings.HasPrefix(file, matchPrefix) {
			continue
		}
		out.ChangedFiles[strings.TrimPrefix(file, matchPrefix)] = cloneFileChange(change)
	}
	return out
}

func cloneChangeSet(changeSet detectors.ChangeSet) detectors.ChangeSet {
	out := detectors.ChangeSet{ChangedFiles: map[string]detectors.FileChange{}}
	for file, change := range changeSet.ChangedFiles {
		out.ChangedFiles[file] = cloneFileChange(change)
	}
	return out
}

func cloneFileChange(change detectors.FileChange) detectors.FileChange {
	out := detectors.FileChange{
		Status: change.Status,
		Lines:  map[int]struct{}{},
	}
	for line := range change.Lines {
		out.Lines[line] = struct{}{}
	}
	if len(out.Lines) == 0 {
		out.Lines = map[int]struct{}{}
	}
	return out
}
