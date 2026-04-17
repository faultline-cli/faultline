package output

import (
	"fmt"
	"strings"

	"faultline/internal/model"
)

// FormatAnalysisEvidenceText formats only the top result plus extracted evidence.
func FormatAnalysisEvidenceText(a *model.Analysis) string {
	if a == nil || len(a.Results) == 0 {
		return "No known playbook matched this input.\n"
	}
	result := a.Results[0]
	lines := []string{
		fmt.Sprintf("EVIDENCE  %s · %s", result.Playbook.ID, result.Playbook.Title),
	}
	if a.Source != "" {
		lines = append(lines, "Source: "+a.Source)
	}
	if len(result.Evidence) == 0 {
		lines = append(lines, "", "No extracted evidence lines were recorded.")
		return strings.Join(lines, "\n") + "\n"
	}
	lines = append(lines, "", "Matched evidence:")
	for i, item := range result.Evidence {
		lines = append(lines, fmt.Sprintf("  %d. %s", i+1, item))
	}
	return strings.Join(lines, "\n") + "\n"
}

// FormatAnalysisEvidenceMarkdown formats only the top result plus extracted evidence as markdown.
func FormatAnalysisEvidenceMarkdown(a *model.Analysis) string {
	if a == nil || len(a.Results) == 0 {
		return "# No Match\n\nNo known playbook matched this input.\n"
	}
	result := a.Results[0]
	sections := []string{
		"# Faultline Evidence",
		"",
		strings.Join(filterEmpty([]string{
			"- ID: `" + result.Playbook.ID + "`",
			joinMetadataListItem("Title", result.Playbook.Title),
			joinMetadataListItem("Source", wrapMarkdownCode(a.Source)),
		}), "\n"),
		"",
		"## Matched Evidence",
	}
	if len(result.Evidence) == 0 {
		sections = append(sections, "", "No extracted evidence lines were recorded.")
		return strings.TrimSpace(strings.Join(sections, "\n")) + "\n"
	}
	sections = append(sections, "", "```text", strings.Join(result.Evidence, "\n"), "```")
	return strings.TrimSpace(strings.Join(sections, "\n")) + "\n"
}

func wrapMarkdownCode(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return "`" + value + "`"
}
