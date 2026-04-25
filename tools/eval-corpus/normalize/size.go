package normalize

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseSize converts a human-readable byte size string to a byte count.
// Supported suffixes (case-insensitive): gb, mb, kb, b.
// An empty string or "0" returns 0 (no limit).
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" || s == "0" {
		return 0, nil
	}

	type entry struct {
		suffix string
		mult   int64
	}
	// Longest suffix first to avoid "b" matching "mb".
	suffixes := []entry{
		{"gb", 1 << 30},
		{"mb", 1 << 20},
		{"kb", 1 << 10},
		{"b", 1},
	}
	for _, e := range suffixes {
		if strings.HasSuffix(s, e.suffix) {
			num := strings.TrimSpace(strings.TrimSuffix(s, e.suffix))
			n, err := strconv.ParseInt(num, 10, 64)
			if err != nil || n < 0 {
				return 0, fmt.Errorf("invalid size %q", s)
			}
			return n * e.mult, nil
		}
	}
	// Plain integer with no suffix.
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid size %q: expected a non-negative integer with optional suffix (b, kb, mb, gb)", s)
	}
	return n, nil
}

// truncateSentinel is appended when a log is trimmed by ApplyMaxSize.
const truncateSentinel = "\n...[truncated]"

// ApplyMaxSize truncates s to maxBytes, appending a sentinel suffix.
// When maxBytes is 0, s is returned unchanged.
func ApplyMaxSize(s string, maxBytes int64) string {
	if maxBytes <= 0 || int64(len(s)) <= maxBytes {
		return s
	}
	return s[:maxBytes] + truncateSentinel
}
