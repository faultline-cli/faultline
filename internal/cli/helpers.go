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

func firstInt64(flagValue int64, envVar string) int64 {
	if flagValue != 0 {
		return flagValue
	}
	if s := strings.TrimSpace(envVar); s != "" {
		if parsed, err := strconv.ParseInt(s, 10, 64); err == nil && parsed != 0 {
			return parsed
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

// resolveStoreSetting is used by fix and workflow, which default the store to
// off and require --history or an explicit --store path to opt in. analyze
// uses resolveAnalyzeStoreSetting instead, which defaults to auto so that
// faultline report can see data from routine analyze runs.
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

func resolveAnalyzeStoreSetting(noHistory, noStore bool, storePath string) string {
	if noHistory || noStore {
		return "off"
	}
	if explicit := firstNonEmpty(storePath, os.Getenv(storeEnv)); explicit != "" {
		return explicit
	}
	return "auto"
}

func resolveStoreHistoryOutput(history, noHistory, noStore bool, storePath string) bool {
	if noHistory || noStore {
		return false
	}
	explicit := firstNonEmpty(storePath, os.Getenv(storeEnv))
	if strings.EqualFold(explicit, "off") {
		return false
	}
	return history || explicit != ""
}

// joinLines joins strings with newlines, used for Long/Example in command descriptions.
func joinLines(lines ...string) string {
	return strings.Join(lines, "\n")
}
