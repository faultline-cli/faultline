package cli

import (
	"os"
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

func resolveStoreSetting(history, noHistory, noStore bool, storePath string) string {
	if noHistory || noStore {
		return "off"
	}
	if explicit := firstNonEmpty(storePath, os.Getenv(storeEnv)); explicit != "" {
		return explicit
	}
	if history {
		return "auto"
	}
	return "off"
}

// joinLines joins strings with newlines, used for Long/Example in command descriptions.
func joinLines(lines ...string) string {
	return strings.Join(lines, "\n")
}
