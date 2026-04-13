package engine

import (
	"regexp"
	"strings"
)

var (
	ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
	rfc3339Pattern    = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}[tT ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:z|[+-]\d{2}:?\d{2})?\b`)
	dateTimePattern   = regexp.MustCompile(`\b\d{4}/\d{2}/\d{2}[ t]\d{2}:\d{2}:\d{2}(?:\.\d+)?\b`)
	timeOnlyPattern   = regexp.MustCompile(`\b\d{2}:\d{2}:\d{2}(?:\.\d+)?\b`)
	tmpPathPattern    = regexp.MustCompile(`(?i)(?:/tmp|/var/folders|/private/var/folders|/home/runner/work/_temp|c:\\users\\[^\\\s]+\\appdata\\local\\temp)[^\s:]*`)
	uuidPattern       = regexp.MustCompile(`\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)
)

// CanonicalizeLog strips unstable noise while preserving the original line
// structure used for deterministic fixture storage and analysis.
func CanonicalizeLog(raw string) string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")
	parts := strings.Split(raw, "\n")
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		line := sanitizeLogFragment(part)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func NormalizeLine(line string) string {
	return strings.ToLower(sanitizeLogFragment(line))
}

func sanitizeLogFragment(line string) string {
	line = ansiEscapePattern.ReplaceAllString(line, "")
	line = rfc3339Pattern.ReplaceAllString(line, "<timestamp>")
	line = dateTimePattern.ReplaceAllString(line, "<timestamp>")
	line = timeOnlyPattern.ReplaceAllString(line, "<timestamp>")
	line = tmpPathPattern.ReplaceAllString(line, "<tmp>")
	line = uuidPattern.ReplaceAllString(line, "<id>")
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	return strings.Join(strings.Fields(line), " ")
}
