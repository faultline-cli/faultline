package output

import (
	"encoding/json"
	"fmt"
	"strings"

	analysiscompare "faultline/internal/compare"
)

func FormatCompareText(report analysiscompare.Report) string {
	lines := []string{"COMPARE"}
	if report.LeftSource != "" {
		lines = append(lines, "Previous: "+report.LeftSource)
	}
	if report.RightSource != "" {
		lines = append(lines, "Current: "+report.RightSource)
	}

	sections := []string{strings.Join(lines, "\n")}
	sections = append(sections, joinCompareSection("Summary", bulletLines(report.Summary)))
	if overview := compareOverviewText(report); overview != "" {
		sections = append(sections, joinCompareSection("Diagnosis", overview))
	}
	if body := compareDeltaText("Evidence Changes", report.Evidence); body != "" {
		sections = append(sections, joinCompareSection("Evidence Changes", body))
	}
	if body := compareDeltaText("Repo Context Changes", report.RepoFiles); body != "" {
		sections = append(sections, joinCompareSection("Repo Context Changes", body))
	}
	if body := compareDeltaText("Delta File Changes", report.DeltaFiles); body != "" {
		sections = append(sections, joinCompareSection("Delta File Changes", body))
	}
	if body := compareDeltaText("Delta Test Changes", report.DeltaTests); body != "" {
		sections = append(sections, joinCompareSection("Delta Test Changes", body))
	}
	if body := compareDeltaText("Delta Error Changes", report.DeltaErrors); body != "" {
		sections = append(sections, joinCompareSection("Delta Error Changes", body))
	}
	return strings.TrimSpace(strings.Join(filterEmpty(sections), "\n\n")) + "\n"
}

func FormatCompareMarkdown(report analysiscompare.Report) string {
	sections := []string{
		"# Faultline Compare",
		"",
		strings.Join(filterEmpty([]string{
			joinMetadataListItem("Previous", wrapCode(report.LeftSource)),
			joinMetadataListItem("Current", wrapCode(report.RightSource)),
		}), "\n"),
		"",
		"## Summary",
		"",
		bulletLines(report.Summary),
	}
	if overview := compareOverviewMarkdown(report); overview != "" {
		sections = append(sections, "", "## Diagnosis", "", overview)
	}
	if body := compareDeltaMarkdown(report.Evidence); body != "" {
		sections = append(sections, "", "## Evidence Changes", "", body)
	}
	if body := compareDeltaMarkdown(report.RepoFiles); body != "" {
		sections = append(sections, "", "## Repo Context Changes", "", body)
	}
	if body := compareDeltaMarkdown(report.DeltaFiles); body != "" {
		sections = append(sections, "", "## Delta File Changes", "", body)
	}
	if body := compareDeltaMarkdown(report.DeltaTests); body != "" {
		sections = append(sections, "", "## Delta Test Changes", "", body)
	}
	if body := compareDeltaMarkdown(report.DeltaErrors); body != "" {
		sections = append(sections, "", "## Delta Error Changes", "", body)
	}
	return strings.TrimSpace(strings.Join(filterEmpty(sections), "\n")) + "\n"
}

func FormatCompareJSON(report analysiscompare.Report) (string, error) {
	data, err := json.Marshal(report)
	if err != nil {
		return "", fmt.Errorf("marshal compare JSON: %w", err)
	}
	return string(data) + "\n", nil
}

func compareOverviewText(report analysiscompare.Report) string {
	var lines []string
	if report.Previous != nil {
		lines = append(lines, fmt.Sprintf("Previous: %s (%s) %d%%", report.Previous.FailureID, report.Previous.Title, int(report.Previous.Confidence*100+0.5)))
	}
	if report.Current != nil {
		lines = append(lines, fmt.Sprintf("Current: %s (%s) %d%%", report.Current.FailureID, report.Current.Title, int(report.Current.Confidence*100+0.5)))
	}
	if report.DiagnosisChanged {
		lines = append(lines, "Diagnosis changed: yes")
	} else {
		lines = append(lines, "Diagnosis changed: no")
	}
	return strings.Join(lines, "\n")
}

func compareOverviewMarkdown(report analysiscompare.Report) string {
	var lines []string
	if report.Previous != nil {
		lines = append(lines, fmt.Sprintf("Previous: `%s` (%s) %d%%", report.Previous.FailureID, report.Previous.Title, int(report.Previous.Confidence*100+0.5)))
	}
	if report.Current != nil {
		lines = append(lines, fmt.Sprintf("Current: `%s` (%s) %d%%", report.Current.FailureID, report.Current.Title, int(report.Current.Confidence*100+0.5)))
	}
	if report.DiagnosisChanged {
		lines = append(lines, "Diagnosis changed: yes")
	} else {
		lines = append(lines, "Diagnosis changed: no")
	}
	return bulletLines(lines)
}

func compareDeltaText(_ string, delta analysiscompare.StringDelta) string {
	var lines []string
	for _, item := range delta.Added {
		lines = append(lines, "+ "+item)
	}
	for _, item := range delta.Removed {
		lines = append(lines, "- "+item)
	}
	return strings.Join(lines, "\n")
}

func compareDeltaMarkdown(delta analysiscompare.StringDelta) string {
	var lines []string
	for _, item := range delta.Added {
		lines = append(lines, "- Added: "+item)
	}
	for _, item := range delta.Removed {
		lines = append(lines, "- Removed: "+item)
	}
	return strings.Join(lines, "\n")
}

func joinCompareSection(title, body string) string {
	if strings.TrimSpace(body) == "" {
		return ""
	}
	return title + "\n" + strings.Repeat("-", len(title)) + "\n\n" + body
}

func wrapCode(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return "`" + value + "`"
}
