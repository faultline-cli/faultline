package fixtures

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// sanitizeRule pairs a compiled pattern with a named label and replacement template.
// replacement may reference regexp submatches using $1, $2, etc.
type sanitizeRule struct {
	name        string
	pattern     *regexp.Regexp
	replacement string
}

// sanitizeRules is the deterministic, ordered set of masking rules applied by
// ApplySanitizeRules. Rules are conservative: they target high-confidence
// credential and identity patterns only to avoid over-redaction.
var sanitizeRules = []sanitizeRule{
	// GitHub personal access tokens and app tokens.
	{
		name:        "github-token",
		pattern:     regexp.MustCompile(`(?i)\b(ghp|ghs|gho|github_pat)_[A-Za-z0-9_]{20,}\b`),
		replacement: "<redacted-github-token>",
	},
	// AWS access key IDs.
	{
		name:        "aws-key",
		pattern:     regexp.MustCompile(`\b(AKIA|AIPA|ASIA|AROA)[A-Z0-9]{16}\b`),
		replacement: "<redacted-aws-key>",
	},
	// Authorization header values (Bearer, Token, Basic).
	{
		name:        "auth-header",
		pattern:     regexp.MustCompile(`(?i)(Authorization:\s*(?:Bearer|Token|Basic)\s+)\S+`),
		replacement: "${1}<redacted>",
	},
	// Credentials embedded in URLs: https://user:password@host
	{
		name:        "url-credentials",
		pattern:     regexp.MustCompile(`(?i)(https?://)([^:\s@]+):([^@\s]+)@`),
		replacement: "${1}<redacted>:<redacted>@",
	},
	// Generic secret key-value assignments (conservative: key=value or key: value).
	{
		name:        "credential-kv",
		pattern:     regexp.MustCompile(`(?i)((?:password|passwd|secret|api[_-]?key|access[_-]?token|auth[_-]?token|private[_-]?key)\s*[=:]\s*)(\S+)`),
		replacement: "${1}<redacted>",
	},
	// JWT tokens: three base64url segments starting with eyJ.
	{
		name:        "jwt",
		pattern:     regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`),
		replacement: "<redacted-jwt>",
	},
	// PEM-encoded private key blocks.
	{
		name:        "pem-key",
		pattern:     regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*(?:PRIVATE|RSA|EC|DSA) KEY-----.*?-----END [A-Z ]*(?:PRIVATE|RSA|EC|DSA) KEY-----`),
		replacement: "<redacted-pem-key>",
	},
	// Email addresses.
	{
		name:        "email",
		pattern:     regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`),
		replacement: "<redacted-email>",
	},
}

// Replacement records how many times a named rule matched in a single field.
type Replacement struct {
	Pattern string
	Field   string
	Count   int
}

// SanitizeResult summarises the rules applied to a single staging fixture.
type SanitizeResult struct {
	FixtureID    string
	Path         string
	Replacements []Replacement
	DryRun       bool
}

// TotalReplacements returns the sum of all replacement counts.
func (r SanitizeResult) TotalReplacements() int {
	total := 0
	for _, rep := range r.Replacements {
		total += rep.Count
	}
	return total
}

// SanitizeOptions controls sanitize behaviour.
type SanitizeOptions struct {
	// DryRun reports what would be replaced without modifying the file.
	DryRun bool
}

// Sanitize loads each named staging fixture, applies masking rules to the
// text fields that carry log content, and writes the result back in place
// unless opts.DryRun is true.
func Sanitize(layout Layout, ids []string, opts SanitizeOptions) ([]SanitizeResult, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("at least one staging fixture ID is required")
	}
	results := make([]SanitizeResult, 0, len(ids))
	for _, id := range ids {
		path := filepath.Join(layout.StagingDir, id+".yaml")
		if _, err := os.Stat(path); err != nil {
			return results, fmt.Errorf("staging fixture %q not found at %s", id, path)
		}
		staging, err := Load(layout, ClassStaging)
		if err != nil {
			return results, err
		}
		var target *Fixture
		for i := range staging {
			if staging[i].ID == id {
				target = &staging[i]
				break
			}
		}
		if target == nil {
			return results, fmt.Errorf("staging fixture %q not found", id)
		}

		var allReplacements []Replacement
		target.RawLog, allReplacements = applySanitizeRulesField(target.RawLog, "raw_log", allReplacements)
		target.NormalizedLog, allReplacements = applySanitizeRulesField(target.NormalizedLog, "normalized_log", allReplacements)

		result := SanitizeResult{
			FixtureID:    id,
			Path:         path,
			Replacements: allReplacements,
			DryRun:       opts.DryRun,
		}
		results = append(results, result)

		if !opts.DryRun {
			if err := writeFixture(path, *target); err != nil {
				return results, fmt.Errorf("write sanitized fixture %q: %w", id, err)
			}
		}
	}
	return results, nil
}

// ApplySanitizeRules applies the masking rule set to a single text string and
// returns the redacted text plus the set of replacements that were made. It is
// exposed so callers can apply sanitization to raw log text before ingestion.
func ApplySanitizeRules(text string) (string, []Replacement) {
	var reps []Replacement
	result, reps := applySanitizeRulesField(text, "", reps)
	return result, reps
}

// applySanitizeRulesField applies all rules to the given text, labelling each
// Replacement with the given field name, and appends the non-zero results to
// existing.
func applySanitizeRulesField(text, field string, existing []Replacement) (string, []Replacement) {
	if text == "" {
		return text, existing
	}
	for _, rule := range sanitizeRules {
		replaced := rule.pattern.ReplaceAllString(text, rule.replacement)
		if replaced == text {
			continue
		}
		count := countMatches(rule.pattern, text)
		text = replaced
		existing = append(existing, Replacement{Pattern: rule.name, Field: field, Count: count})
	}
	return text, existing
}

// countMatches returns the number of non-overlapping matches of pat in s.
func countMatches(pat *regexp.Regexp, s string) int {
	return len(pat.FindAllString(s, -1))
}

// jsonSanitizeReplacement is the JSON shape for a single replacement record.
type jsonSanitizeReplacement struct {
	Pattern string `json:"pattern"`
	Field   string `json:"field,omitempty"`
	Count   int    `json:"count"`
}

// jsonSanitizeResult is the JSON shape for a single fixture sanitize result.
type jsonSanitizeResult struct {
	FixtureID    string                    `json:"fixture_id"`
	Path         string                    `json:"path"`
	Replacements []jsonSanitizeReplacement `json:"replacements"`
	DryRun       bool                      `json:"dry_run"`
	Total        int                       `json:"total_replacements"`
}

// FormatSanitizeResults returns a human-readable or JSON summary of a batch of
// sanitize results.
func FormatSanitizeResults(results []SanitizeResult, jsonOut bool) (string, error) {
	if jsonOut {
		out := make([]jsonSanitizeResult, 0, len(results))
		for _, r := range results {
			jr := jsonSanitizeResult{
				FixtureID:    r.FixtureID,
				Path:         r.Path,
				DryRun:       r.DryRun,
				Total:        r.TotalReplacements(),
				Replacements: []jsonSanitizeReplacement{},
			}
			for _, rep := range r.Replacements {
				jr.Replacements = append(jr.Replacements, jsonSanitizeReplacement{
					Pattern: rep.Pattern,
					Field:   rep.Field,
					Count:   rep.Count,
				})
			}
			out = append(out, jr)
		}
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	}

	if len(results) == 0 {
		return "No staging fixtures sanitized.\n", nil
	}
	var lines []string
	for _, r := range results {
		total := r.TotalReplacements()
		suffix := ""
		if r.DryRun {
			suffix = " (dry-run)"
		}
		if total == 0 {
			lines = append(lines, fmt.Sprintf("%s: no replacements%s", r.FixtureID, suffix))
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %d replacement(s)%s", r.FixtureID, total, suffix))
		// Sort replacements for deterministic output.
		sorted := make([]Replacement, len(r.Replacements))
		copy(sorted, r.Replacements)
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].Field != sorted[j].Field {
				return sorted[i].Field < sorted[j].Field
			}
			return sorted[i].Pattern < sorted[j].Pattern
		})
		for _, rep := range sorted {
			if rep.Field != "" {
				lines = append(lines, fmt.Sprintf("  - %s [%s]: %d", rep.Pattern, rep.Field, rep.Count))
			} else {
				lines = append(lines, fmt.Sprintf("  - %s: %d", rep.Pattern, rep.Count))
			}
		}
	}
	return strings.Join(lines, "\n") + "\n", nil
}
