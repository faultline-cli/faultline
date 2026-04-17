// Package topology provides repo topology awareness for Faultline.
// It parses CODEOWNERS files, builds an ownership graph from the directory
// structure, and derives structural signals from that graph.
package topology

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// OwnerRule maps a gitignore-style path pattern to one or more owner handles.
type OwnerRule struct {
	Pattern string
	Owners  []string
}

// ParseCODEOWNERS reads a CODEOWNERS file from any of the standard locations
// inside root (root/CODEOWNERS, root/.github/CODEOWNERS, root/docs/CODEOWNERS)
// and returns the ordered list of rules. If no CODEOWNERS file is found, an
// empty slice is returned without error. Later rules in the file take
// precedence (GitHub semantics), so callers should iterate in reverse when
// finding the first match.
func ParseCODEOWNERS(root string) ([]OwnerRule, error) {
	candidates := []string{
		filepath.Join(root, "CODEOWNERS"),
		filepath.Join(root, ".github", "CODEOWNERS"),
		filepath.Join(root, "docs", "CODEOWNERS"),
	}
	for _, path := range candidates {
		rules, err := parseCODEOWNERSFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		return rules, nil
	}
	return nil, nil
}

// OwnersFor returns the owners for the given file path based on the provided
// rules. It applies GitHub CODEOWNERS semantics: the last matching rule wins.
// Path should be relative to the repository root and use forward slashes.
func OwnersFor(rules []OwnerRule, path string) []string {
	path = filepath.ToSlash(path)
	var match []string
	for _, rule := range rules {
		if matchPattern(rule.Pattern, path) {
			match = rule.Owners
		}
	}
	return match
}

// parseCODEOWNERSFile reads a single file and returns the parsed rules.
func parseCODEOWNERSFile(path string) ([]OwnerRule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rules []OwnerRule
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		rules = append(rules, OwnerRule{
			Pattern: fields[0],
			Owners:  fields[1:],
		})
	}
	return rules, scanner.Err()
}

// matchPattern tests whether path matches a CODEOWNERS-style pattern.
// Patterns follow gitignore semantics:
//   - A pattern without a leading "/" or containing "/" matches anywhere in the
//     tree (matched against basename or full path).
//   - A pattern with a leading "/" anchors to the repo root.
//   - A pattern ending with "/" matches directories recursively.
//   - "*" matches anything except "/"; "**" matches across slashes.
func matchPattern(pattern, path string) bool {
	path = filepath.ToSlash(strings.TrimPrefix(path, "/"))
	pattern = strings.TrimPrefix(pattern, "/")

	// Directory-only pattern (e.g. "docs/") - match path prefix.
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(path, strings.TrimSuffix(pattern, "/"))
	}

	// Patterns without "/" match against the basename as well as the full path.
	if !strings.Contains(pattern, "/") {
		base := filepath.Base(path)
		if globMatch(pattern, base) {
			return true
		}
	}

	// Try matching against the full path.
	if globMatch(pattern, path) {
		return true
	}

	// If the pattern has no leading slash and no wildcard, also check prefixes
	// so that "docs" matches "docs/api/handler.go".
	if !strings.ContainsAny(pattern, "*?") {
		return strings.HasPrefix(path, pattern+"/") || path == pattern
	}

	return false
}

// globMatch is a simple glob that supports "*" (non-separator) and "**" (any).
func globMatch(pattern, name string) bool {
	// Replace "**" placeholder with a path-crossing wildcard handled below.
	// We do two-step conversion: ** -> \x00 so the single * logic is clean.
	pattern = strings.ReplaceAll(pattern, "**", "\x00")

	for {
		if len(pattern) == 0 {
			return len(name) == 0
		}

		if pattern[0] == '\x00' {
			// ** matches any sequence including slashes.
			pattern = pattern[1:]
			for i := 0; i <= len(name); i++ {
				if globMatch(pattern, name[i:]) {
					return true
				}
			}
			return false
		}

		if pattern[0] == '*' {
			// * matches any non-slash sequence.
			pattern = pattern[1:]
			for i := 0; i <= len(name); i++ {
				if i > 0 && name[i-1] == '/' {
					break
				}
				if globMatch(pattern, name[i:]) {
					return true
				}
			}
			return false
		}

		if pattern[0] == '?' {
			if len(name) == 0 || name[0] == '/' {
				return false
			}
			pattern = pattern[1:]
			name = name[1:]
			continue
		}

		if len(name) == 0 || pattern[0] != name[0] {
			return false
		}
		pattern = pattern[1:]
		name = name[1:]
	}
}
