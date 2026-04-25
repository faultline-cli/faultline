// Package normalize provides deterministic transformations for log content
// before fixture storage and evaluation.
package normalize

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// FixtureID returns the lower-case hex SHA-256 of the content string.
// The hash is computed over the raw UTF-8 bytes after trimming surrounding
// whitespace, ensuring identical logs produce the same ID regardless of
// leading or trailing whitespace from the source format.
func FixtureID(content string) string {
	normalized := strings.TrimSpace(content)
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
