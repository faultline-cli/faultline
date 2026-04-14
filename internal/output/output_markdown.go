package output

import (
	"fmt"
	"strings"

	"faultline/internal/model"
)

// ── Markdown formatters ───────────────────────────────────────────────────────

// FormatAnalysisMarkdown formats an analysis as raw markdown.
func FormatAnalysisMarkdown(a *model.Analysis, top int, mode Mode) string {
	if a == nil || len(a.Results) == 0 {
		return strings.Join([]string{
			"# No Match",
			"",
			"No known playbook matched this input.",
			"",
			"- Run `faultline list` to see available playbooks.",
			"- Pass `--json` for machine-readable output.",
		}, "\n") + "\n"
	}

	results := topN(a.Results, top)
	sections := make([]string, 0, len(results)+1)
	for i, result := range results {
		sections = append(sections, formatAnalysisMarkdownResult(a, result, i, len(results), mode == ModeDetailed))
	}
	if mode == ModeDetailed {
		if repo := markdownListSection("## Repo Context", repoContextLines(a.RepoContext)); repo != "" {
			sections = append(sections, repo)
		}
	}
	return strings.TrimSpace(strings.Join(sections, "\n\n")) + "\n"
}

// FormatFixMarkdown formats only the fix steps for the top result as markdown.
func FormatFixMarkdown(a *model.Analysis) string {
	if a == nil || len(a.Results) == 0 {
		return FormatAnalysisMarkdown(nil, 1, ModeQuick)
	}
	result := a.Results[0]
	sections := []string{
		"# " + result.Playbook.Title,
		"",
		strings.Join([]string{
			"- ID: `" + result.Playbook.ID + "`",
			"- Category: " + result.Playbook.Category,
		}, "\n"),
	}
	if fix := strings.TrimSpace(result.Playbook.Fix); fix != "" {
		sections = append(sections, "## Fix", "", fix)
	} else {
		sections = append(sections, "## Fix", "", "No fix steps defined for this playbook.")
	}
	return strings.TrimSpace(strings.Join(sections, "\n")) + "\n"
}

// FormatPlaybookDetailsMarkdown formats all fields of a single playbook as markdown.
func FormatPlaybookDetailsMarkdown(pb model.Playbook) string {
	sections := []string{
		"# " + pb.Title,
		"",
		strings.Join(filterEmpty([]string{
			"- ID: `" + pb.ID + "`",
			"- Category: " + pb.Category,
			"- Severity: " + fallback(pb.Severity, "unknown"),
			"- Detector: " + fallback(pb.Detector, "log"),
			joinMetadataListItem("Pack", displayPackName(pb)),
			joinMetadataListItem("Tags", strings.Join(pb.Tags, ", ")),
			joinMetadataListItem("Stages", strings.Join(pb.StageHints, ", ")),
		}), "\n"),
	}

	if summary := strings.TrimSpace(pb.Summary); summary != "" {
		sections = append(sections, "## Summary", "", summary)
	}
	if diagnosis := strings.TrimSpace(pb.Diagnosis); diagnosis != "" {
		sections = append(sections, "## Diagnosis", "", diagnosis)
	}
	if why := strings.TrimSpace(pb.WhyItMatters); why != "" {
		sections = append(sections, "## Why It Matters", "", why)
	}
	if fix := strings.TrimSpace(pb.Fix); fix != "" {
		sections = append(sections, "## Fix", "", fix)
	}
	if validation := strings.TrimSpace(pb.Validation); validation != "" {
		sections = append(sections, "## Validation", "", validation)
	}
	if matchRules := formatMatchSummaryMarkdown(pb); matchRules != "" {
		sections = append(sections, "## Match Rules", "", "Structured fields decide; markdown explains.", "", matchRules)
	}
	return strings.TrimSpace(strings.Join(sections, "\n")) + "\n"
}

// ── markdown helpers ──────────────────────────────────────────────────────────

func formatAnalysisMarkdownResult(a *model.Analysis, result model.Result, rank, total int, detailed bool) string {
	title := result.Playbook.Title
	if total > 1 {
		title = fmt.Sprintf("%d. %s", rank+1, title)
	}

	meta := filterEmpty([]string{
		"- ID: `" + result.Playbook.ID + "`",
		joinMetadataListItem("Pack", displayPackName(result.Playbook)),
		joinMetadataListItem("Confidence", fmt.Sprintf("%d%%", int(result.Confidence*100+0.5))),
		joinMetadataListItem("Category", result.Playbook.Category),
		joinMetadataListItem("Severity", fallback(result.Playbook.Severity, "unknown")),
		joinMetadataListItem("Score", fmt.Sprintf("%.2f", result.Score)),
		joinMetadataListItem("Detector", fallback(result.Detector, "log")),
		joinMetadataListItem("Stage", a.Context.Stage),
	})

	sections := []string{"# " + title, "", strings.Join(meta, "\n")}
	if summary := strings.TrimSpace(result.Playbook.Summary); summary != "" {
		sections = append(sections, "", "## Summary", "", summary)
	}
	if detailed {
		if evidence := markdownListSection("## Evidence", result.Evidence); evidence != "" {
			sections = append(sections, "", evidence)
		}
		if triggered := markdownListSection("## Triggered By", result.Explanation.TriggeredBy); triggered != "" {
			sections = append(sections, "", triggered)
		}
		if amplified := markdownListSection("## Amplified By", result.Explanation.AmplifiedBy); amplified != "" {
			sections = append(sections, "", amplified)
		}
		if mitigated := markdownListSection("## Mitigated By", result.Explanation.MitigatedBy); mitigated != "" {
			sections = append(sections, "", mitigated)
		}
		if breakdown := markdownListSection("## Score Breakdown", scoreBreakdownLines(result.Breakdown)); breakdown != "" {
			sections = append(sections, "", breakdown)
		}
	}
	return strings.TrimSpace(strings.Join(sections, "\n"))
}

func formatMatchSummaryMarkdown(pb model.Playbook) string {
	var sections []string
	if section := markdownListSection("### match.any", pb.Match.Any); section != "" {
		sections = append(sections, section)
	}
	if section := markdownListSection("### match.all", pb.Match.All); section != "" {
		sections = append(sections, section)
	}
	if section := markdownListSection("### match.none", pb.Match.None); section != "" {
		sections = append(sections, section)
	}
	if section := markdownListSection("### workflow.verify", pb.Workflow.Verify); section != "" {
		sections = append(sections, section)
	}
	return strings.Join(sections, "\n\n")
}

func markdownListSection(title string, lines []string) string {
	items := bulletLines(lines)
	if items == "" {
		return ""
	}
	return title + "\n\n" + items
}

func bulletLines(lines []string) string {
	trimmed := filterEmpty(trimmedLines(lines))
	if len(trimmed) == 0 {
		return ""
	}
	for i, line := range trimmed {
		trimmed[i] = "- " + line
	}
	return strings.Join(trimmed, "\n")
}

func trimmedLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, strings.TrimSpace(line))
	}
	return out
}

func scoreBreakdownLines(breakdown model.ScoreBreakdown) []string {
	if breakdown.FinalScore == 0 {
		return nil
	}
	lines := []string{
		fmt.Sprintf("base: %.2f", breakdown.BaseSignalScore),
		fmt.Sprintf("final: %.2f", breakdown.FinalScore),
	}
	if breakdown.CompoundSignalBonus != 0 {
		lines = append(lines, fmt.Sprintf("compound: +%.2f", breakdown.CompoundSignalBonus))
	}
	if breakdown.BlastRadiusMultiplier != 0 {
		lines = append(lines, fmt.Sprintf("blast radius: +%.2f", breakdown.BlastRadiusMultiplier))
	}
	if breakdown.HotPathMultiplier != 0 {
		lines = append(lines, fmt.Sprintf("hot path: +%.2f", breakdown.HotPathMultiplier))
	}
	if breakdown.ChangeIntroducedBonus != 0 {
		lines = append(lines, fmt.Sprintf("change bonus: %.2f", breakdown.ChangeIntroducedBonus))
	}
	if breakdown.MitigatingEvidenceDiscount != 0 {
		lines = append(lines, fmt.Sprintf("mitigations: -%.2f", breakdown.MitigatingEvidenceDiscount))
	}
	if breakdown.ExplicitExceptionDiscount != 0 {
		lines = append(lines, fmt.Sprintf("suppressions: -%.2f", breakdown.ExplicitExceptionDiscount))
	}
	if breakdown.SafeContextDiscount != 0 {
		lines = append(lines, fmt.Sprintf("safe context: -%.2f", breakdown.SafeContextDiscount))
	}
	return lines
}

func repoContextLines(repo *model.RepoContext) []string {
	if repo == nil {
		return nil
	}
	var lines []string
	if repo.RepoRoot != "" {
		lines = append(lines, "Repo root: "+repo.RepoRoot)
	}
	for _, item := range repo.RecentFiles {
		lines = append(lines, "Recent file: "+item)
	}
	for _, commit := range repo.RelatedCommits {
		lines = append(lines, fmt.Sprintf("Related commit: %s %s %s", commit.Date, commit.Hash, commit.Subject))
	}
	for _, item := range repo.HotspotDirectories {
		lines = append(lines, "Hotspot area: "+item)
	}
	for _, item := range repo.CoChangeHints {
		lines = append(lines, "Co-change: "+item)
	}
	for _, item := range repo.HotfixSignals {
		lines = append(lines, "Hotfix signal: "+item)
	}
	for _, item := range repo.DriftSignals {
		lines = append(lines, "Drift hint: "+item)
	}
	return lines
}

func joinMetadataListItem(label, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return "- " + label + ": " + value
}

func filterEmpty(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return out
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}
