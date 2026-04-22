package output

import (
	"fmt"
	"strings"

	"faultline/internal/model"
)

func FormatHookSummariesText(a *model.Analysis) string {
	lines := analysisHookLines(a)
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func FormatHookSummariesMarkdown(a *model.Analysis) string {
	lines := analysisHookLines(a)
	if len(lines) == 0 {
		return ""
	}
	return "## Hooks\n\n" + bulletLines(lines) + "\n"
}

func analysisHookLines(a *model.Analysis) []string {
	if a == nil || len(a.Results) == 0 {
		return nil
	}
	var lines []string
	for _, result := range a.Results {
		if result.Hooks == nil {
			continue
		}
		prefix := result.Playbook.ID
		for _, item := range hookSummaryLines(result.Hooks) {
			lines = append(lines, fmt.Sprintf("%s: %s", prefix, item))
		}
	}
	return lines
}

func hookSummaryLines(report *model.HookReport) []string {
	if report == nil {
		return nil
	}
	lines := []string{
		fmt.Sprintf("mode: %s", report.Mode),
		fmt.Sprintf("confidence: %.2f -> %.2f (%+.2f)", report.BaseConfidence, report.FinalConfidence, report.ConfidenceDelta),
	}
	for _, item := range report.Results {
		line := fmt.Sprintf("%s/%s: %s", item.Category, item.ID, item.Status)
		switch {
		case item.Passed != nil && *item.Passed:
			line += " (passed)"
		case item.Passed != nil:
			line += " (failed check)"
		}
		if item.Reason != "" {
			line += " - " + item.Reason
		}
		lines = append(lines, line)
	}
	return lines
}
