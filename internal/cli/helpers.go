package cli

import (
	"strconv"
	"strings"
)

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func firstInt64(values ...interface{}) int64 {
	for _, value := range values {
		switch v := value.(type) {
		case int64:
			if v != 0 {
				return v
			}
		case string:
			if strings.TrimSpace(v) == "" {
				continue
			}
			parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if err == nil && parsed != 0 {
				return parsed
			}
		}
	}
	return 0
}

func deriveGitLabAPIBaseURL(serverURL string) string {
	serverURL = strings.TrimSpace(serverURL)
	if serverURL == "" {
		return ""
	}
	return strings.TrimRight(serverURL, "/") + "/api/v4"
}

// joinLines joins strings with newlines, used for Long/Example in command descriptions.
func joinLines(lines ...string) string {
	return strings.Join(lines, "\n")
}
