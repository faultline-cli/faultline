package cli

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
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

func hideCommandTree(cmd *cobra.Command) {
	for _, child := range cmd.Commands() {
		child.Hidden = true
		hideCommandTree(child)
	}
}
