package fixtures

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func FormatIngestResult(result IngestResult, jsonOut bool) (string, error) {
	if jsonOut {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("Written: %d", len(result.Written)))
	for _, fixture := range result.Written {
		lines = append(lines, fmt.Sprintf("- %s (%s)", fixture.ID, fixture.Source.URL))
	}
	if len(result.Skipped) > 0 {
		lines = append(lines, fmt.Sprintf("Skipped: %d", len(result.Skipped)))
		for _, entry := range result.Skipped {
			lines = append(lines, "- "+entry)
		}
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func FormatReviewReport(report ReviewReport, jsonOut bool) (string, error) {
	if jsonOut {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	}
	var lines []string
	for _, item := range report.Items {
		prediction := "no-match"
		if item.PredictedTopID != "" {
			prediction = item.PredictedTopID
		}
		lines = append(lines, fmt.Sprintf("%s [%s] top=%s", item.Fixture.ID, item.Status, prediction))
		if item.DuplicateOf != "" {
			lines = append(lines, fmt.Sprintf("  duplicate_of: %s", item.DuplicateOf))
		}
		if len(item.NearDuplicates) > 0 {
			lines = append(lines, fmt.Sprintf("  near_duplicates: %s", strings.Join(item.NearDuplicates, ", ")))
		}
		if len(item.PredictedTop3) > 0 {
			lines = append(lines, fmt.Sprintf("  top3: %s", strings.Join(item.PredictedTop3, ", ")))
		}
		if item.Fixture.Source.URL != "" {
			lines = append(lines, fmt.Sprintf("  source: %s", item.Fixture.Source.URL))
		}
	}
	if len(lines) == 0 {
		return "No staging fixtures found.\n", nil
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func FormatStatsReport(report Report, jsonOut bool) (string, error) {
	if jsonOut {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	}
	lines := []string{
		fmt.Sprintf("class: %s", report.Class),
		fmt.Sprintf("fixtures: %d", report.FixtureCount),
		fmt.Sprintf("top_1: %.3f", report.Top1Rate()),
		fmt.Sprintf("top_3: %.3f", report.Top3Rate()),
		fmt.Sprintf("unmatched: %.3f", report.UnmatchedRate()),
		fmt.Sprintf("false_positive: %.3f", report.FalsePositiveRate()),
		fmt.Sprintf("weak_match: %.3f", report.WeakMatchRate()),
	}
	if len(report.UnmatchedFixtureIDs) > 0 {
		lines = append(lines, "unmatched_ids: "+strings.Join(report.UnmatchedFixtureIDs, ", "))
	}
	if len(report.FalsePositiveFixtureIDs) > 0 {
		lines = append(lines, "false_positive_ids: "+strings.Join(report.FalsePositiveFixtureIDs, ", "))
	}
	if len(report.WeakMatchFixtureIDs) > 0 {
		lines = append(lines, "weak_ids: "+strings.Join(report.WeakMatchFixtureIDs, ", "))
	}
	if len(report.RecurringPatterns) > 0 {
		keys := sortedKeys(report.RecurringPatterns)
		patterns := make([]string, 0, len(keys))
		for _, key := range keys {
			patterns = append(patterns, fmt.Sprintf("%s=%d", key, report.RecurringPatterns[key]))
		}
		lines = append(lines, "patterns: "+strings.Join(patterns, ", "))
	}
	if len(report.Providers) > 0 {
		keys := sortedKeys(report.Providers)
		providers := make([]string, 0, len(keys))
		for _, key := range keys {
			providers = append(providers, fmt.Sprintf("%s=%d", key, report.Providers[key]))
		}
		lines = append(lines, "providers: "+strings.Join(providers, ", "))
	}
	if len(report.Adapters) > 0 {
		keys := sortedKeys(report.Adapters)
		adapters := make([]string, 0, len(keys))
		for _, key := range keys {
			adapters = append(adapters, fmt.Sprintf("%s=%d", key, report.Adapters[key]))
		}
		lines = append(lines, "adapters: "+strings.Join(adapters, ", "))
	}
	if len(report.ThresholdViolations) > 0 {
		sorted := append([]string(nil), report.ThresholdViolations...)
		sort.Strings(sorted)
		lines = append(lines, "violations: "+strings.Join(sorted, " | "))
	}
	return strings.Join(lines, "\n") + "\n", nil
}
