package output

import (
	"fmt"
	"math"
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
		if delta := markdownListSection("## Delta Diagnosis", deltaLines(a.Delta)); delta != "" {
			sections = append(sections, delta)
		}
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
		if rank == 0 {
			if differential := markdownListSection("## Differential Diagnosis", differentialLines(a.Results)); differential != "" {
				sections = append(sections, "", differential)
			}
			if confidence := markdownListSection("## Confidence Breakdown", confidenceBreakdownLines(a.Results, result)); confidence != "" {
				sections = append(sections, "", confidence)
			}
		}
		if !sameLines(result.Explanation.TriggeredBy, result.Evidence) {
			if triggered := markdownListSection("## Triggered By", result.Explanation.TriggeredBy); triggered != "" {
				sections = append(sections, "", triggered)
			}
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
		if rank == 0 {
			if fix := strings.TrimSpace(result.Playbook.Fix); fix != "" {
				sections = append(sections, "", "## Suggested Fix", "", fix)
			}
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

func sameLines(left, right []string) bool {
	left = filterEmpty(trimmedLines(left))
	right = filterEmpty(trimmedLines(right))
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func sharesLine(left, right []string) bool {
	left = filterEmpty(trimmedLines(left))
	right = filterEmpty(trimmedLines(right))
	for _, l := range left {
		for _, r := range right {
			if l == r {
				return true
			}
		}
	}
	return false
}

func differentialLines(results []model.Result) []string {
	if len(results) < 2 {
		return nil
	}
	top := results[0]
	runnerUp := results[1]
	lines := []string{
		fmt.Sprintf("top candidate: %s (%s)", top.Playbook.ID, top.Playbook.Title),
		fmt.Sprintf("runner-up: %s (%s)", runnerUp.Playbook.ID, runnerUp.Playbook.Title),
	}
	gap := roundMarkdownScore(top.Score - runnerUp.Score)
	if gap <= 0 {
		lines = append(lines, "score gap: tied on score; stable ordering kept the top candidate first")
	} else {
		lines = append(lines, fmt.Sprintf("score gap: +%.2f over the runner-up", gap))
	}
	if reason := higherRankedMarkdownReason(top, runnerUp); reason != "" {
		lines = append(lines, "higher-ranked because: "+reason)
	}
	if sharesLine(top.Evidence, runnerUp.Evidence) {
		lines = append(lines, "alternate remains plausible because: it matched the same failing evidence line")
	} else if len(filterEmpty(trimmedLines(runnerUp.Evidence))) > 0 {
		lines = append(lines, "alternate remains plausible because: it still matched explicit evidence from the input")
	}
	return lines
}

func confidenceBreakdownLines(results []model.Result, result model.Result) []string {
	lines := []string{
		fmt.Sprintf("reported confidence: %d%%", int(result.Confidence*100+0.5)),
	}
	if result.Ranking == nil {
		lines = append(lines, fmt.Sprintf("detector score: %.2f", result.Score))
		return lines
	}
	lines = append(lines,
		fmt.Sprintf("detector baseline: %.2f", result.Ranking.BaselineScore),
		fmt.Sprintf("final reranked score: %.2f", result.Ranking.FinalScore),
	)
	if result.Ranking.Prior != 0 {
		lines = append(lines, fmt.Sprintf("conservative prior: %+.2f", result.Ranking.Prior))
	}
	for _, item := range rankingContributionLines(result.Ranking, 3) {
		lines = append(lines, item)
	}
	if len(results) > 1 {
		gap := roundMarkdownScore(result.Score - results[1].Score)
		if gap > 0 {
			lines = append(lines, fmt.Sprintf("margin over #2: +%.2f", gap))
		}
	}
	return lines
}

func higherRankedMarkdownReason(top, runnerUp model.Result) string {
	if top.Ranking == nil || runnerUp.Ranking == nil {
		return ""
	}
	runnerMap := map[string]model.RankingContribution{}
	for _, item := range runnerUp.Ranking.Contributions {
		runnerMap[item.Feature] = item
	}
	for _, item := range top.Ranking.Contributions {
		if contributionIsMetadata(item.Feature) {
			continue
		}
		if roundMarkdownScore(item.Contribution-runnerMap[item.Feature].Contribution) <= 0 {
			continue
		}
		if strings.TrimSpace(item.Reason) != "" {
			return item.Reason
		}
		return item.Feature
	}
	return ""
}

func rankingContributionLines(ranking *model.Ranking, limit int) []string {
	if ranking == nil || limit <= 0 {
		return nil
	}
	lines := make([]string, 0, limit)
	for _, item := range ranking.Contributions {
		if contributionIsMetadata(item.Feature) {
			continue
		}
		lines = append(lines, fmt.Sprintf("%+.2f %s", item.Contribution, fallback(item.Reason, item.Feature)))
		if len(lines) >= limit {
			break
		}
	}
	return lines
}

func roundMarkdownScore(value float64) float64 {
	return math.Round(value*100) / 100
}

func contributionIsMetadata(feature string) bool {
	switch feature {
	case "detector_score", "historical_fixture_support", "candidate_separation":
		return true
	default:
		return false
	}
}

func scoreBreakdownLines(breakdown model.ScoreBreakdown) []string {
	if breakdown.FinalScore == 0 {
		return nil
	}
	if breakdown.CompoundSignalBonus == 0 &&
		breakdown.BlastRadiusMultiplier == 0 &&
		breakdown.HotPathMultiplier == 0 &&
		breakdown.ChangeIntroducedBonus == 0 &&
		breakdown.MitigatingEvidenceDiscount == 0 &&
		breakdown.ExplicitExceptionDiscount == 0 &&
		breakdown.SafeContextDiscount == 0 {
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
		lines = append(lines, fmt.Sprintf("change bonus: %+.2f", breakdown.ChangeIntroducedBonus))
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

func deltaLines(delta *model.Delta) []string {
	if delta == nil {
		return nil
	}
	var lines []string
	for _, cause := range delta.Causes {
		lines = append(lines, fmt.Sprintf("%s: %.2f", cause.Kind, cause.Score))
		for _, reason := range cause.Reasons {
			lines = append(lines, cause.Kind+" reason: "+reason)
		}
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
