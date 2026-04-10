package repo

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Commit represents a single parsed git commit.
type Commit struct {
	Hash    string
	Time    time.Time
	Subject string
	Files   []string // paths changed in this commit
}

// LoadHistory fetches commits from HEAD that are newer than the given duration
// string (e.g. "30d", "7d"). Returns commits in reverse-chronological order.
func LoadHistory(s *Scanner, since string) ([]Commit, error) {
	sinceArg, err := parseSince(since)
	if err != nil {
		return nil, err
	}

	// Each commit starts with \x1f followed by metadata separated by \x1e.
	// The remaining lines in the block come from --name-only and are file paths.
	formatArg := "--format=%x1f%H%x1e%at%x1e%s"
	logOut, err := s.Run("log", "--no-merges", "--name-only", sinceArg, formatArg)
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	if strings.TrimSpace(logOut) == "" {
		return nil, nil
	}

	rawBlocks := strings.Split(logOut, "\x1f")
	commits := make([]Commit, 0, len(rawBlocks))
	for _, block := range rawBlocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		commit, ok := parseCommitBlock(block)
		if !ok {
			continue
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func parseCommitBlock(block string) (Commit, bool) {
	lines := strings.Split(block, "\n")
	if len(lines) == 0 {
		return Commit{}, false
	}

	header := strings.SplitN(strings.TrimSpace(lines[0]), "\x1e", 3)
	if len(header) != 3 {
		return Commit{}, false
	}

	unixTS, err := strconv.ParseInt(strings.TrimSpace(header[1]), 10, 64)
	if err != nil {
		return Commit{}, false
	}

	files := make([]string, 0, len(lines)-1)
	seen := make(map[string]struct{}, len(lines)-1)
	for _, line := range lines[1:] {
		f := filepath.ToSlash(strings.TrimSpace(line))
		if f != "" {
			if _, ok := seen[f]; ok {
				continue
			}
			seen[f] = struct{}{}
			files = append(files, f)
		}
	}

	return Commit{
		Hash:    strings.TrimSpace(header[0]),
		Time:    time.Unix(unixTS, 0).UTC(),
		Subject: strings.TrimSpace(header[2]),
		Files:   files,
	}, true
}

// parseSince converts a human-readable duration string to a git --since flag.
// Supported: Nd (days), Nw (weeks), Nm (months), or a bare integer (days).
// Also accepts Ny (years) and passes through git-native phrases with spaces.
// Invalid input falls back to "30 days ago" so git context stays best-effort.
func parseSince(s string) (string, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return "--since=30 days ago", nil
	}
	// Accept git native strings like "1 week ago" unchanged.
	if strings.Contains(s, " ") {
		return "--since=" + s, nil
	}

	suffix := s[len(s)-1:]
	numStr := s[:len(s)-1]

	var unit string
	switch suffix {
	case "d":
		unit = "days"
	case "w":
		unit = "weeks"
	case "m":
		unit = "months"
	case "y":
		unit = "years"
	default:
		numStr = s
		unit = "days"
	}

	n, err := strconv.Atoi(numStr)
	if err != nil || n <= 0 {
		return "--since=30 days ago", nil
	}
	unitWord := unit
	if n == 1 {
		unitWord = strings.TrimSuffix(unit, "s")
	}
	return fmt.Sprintf("--since=%d %s ago", n, unitWord), nil
}
