// Package authoring provides deterministic utilities for scaffolding new
// playbooks from sanitized build-log input. It is maintainer-only machinery;
// nothing here ships in the default user narrative.
package authoring

import (
	"sort"
	"strings"
)

// errorKeywords are substrings that increase a line's diagnostic weight.
// Each occurrence contributes +3 to the line's score.
var errorKeywords = []string{
	"error", "failed", "failure", "denied", "unauthorized",
	"forbidden", "not found", "missing", "cannot", "could not",
	"couldn't", "unable to", "invalid", "rejected", "timeout",
	"timed out", "exit code", "exit status", "permission denied",
	"no such", "panic", "fatal", "exception", "connection refused",
	"connection reset", "access denied", "broken pipe",
}

// noisePrefixes identifies line prefixes that carry little diagnostic value.
// Lines whose lowercase form starts with one of these fragments are skipped.
var noisePrefixes = []string{
	"##", "---", "===", "+++",
	"running ", "building ", "downloading ", "installing ",
	"step ", "steps ", "task ", "note:", "info:", "debug:",
}

// ExtractCandidatePatterns extracts up to max candidate match phrases from raw
// log text. Lines are scored by diagnostic weight. Results are returned in
// descending score order with alphabetical tie-breaking so the output is fully
// deterministic for the same input.
//
// Callers should sanitize the log before calling this function to ensure no
// secrets appear in the returned patterns.
func ExtractCandidatePatterns(log string, max int) []string {
	if max <= 0 {
		max = 5
	}

	type scored struct {
		line  string
		score int
	}

	seen := make(map[string]bool)
	var candidates []scored

	for _, raw := range strings.Split(log, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)

		// Skip noise prefixes.
		noise := false
		for _, pfx := range noisePrefixes {
			if strings.HasPrefix(lower, pfx) {
				noise = true
				break
			}
		}
		if noise {
			continue
		}

		// Deduplicate by lowered form.
		if seen[lower] {
			continue
		}
		seen[lower] = true

		// Score the line.
		score := 0
		for _, kw := range errorKeywords {
			if strings.Contains(lower, kw) {
				score += 3
			}
		}
		n := len(line)
		switch {
		case n >= 20 && n <= 100:
			score += 2
		case n < 15:
			score -= 2
		case n > 200:
			score--
		}

		candidates = append(candidates, scored{line: line, score: score})
	}

	// Stable sort: descending score, ascending line for tie-breaking.
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].line < candidates[j].line
	})

	if len(candidates) > max {
		candidates = candidates[:max]
	}

	result := make([]string, len(candidates))
	for i, c := range candidates {
		result[i] = c.line
	}
	return result
}
