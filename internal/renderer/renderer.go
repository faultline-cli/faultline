package renderer

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	glamour "charm.land/glamour/v2"
	lipgloss "charm.land/lipgloss/v2"

	"faultline/internal/model"
)

var leadingHeadingPattern = regexp.MustCompile(`\A(?s)#{1,6}[ \t]+([^\n]+)\n+`)

type Renderer struct {
	opts   Options
	styles styles
}

func New(opts Options) Renderer {
	if opts.Width == 0 {
		opts.Width = defaultWidth
	}
	return Renderer{
		opts:   opts,
		styles: newStyles(),
	}
}

func (r Renderer) RenderNoMatch() string {
	return "No known playbook matched this input.\n" +
		"  - Run \"faultline list\" to see available playbooks.\n" +
		"  - Pass --json for machine-readable output.\n"
}

func (r Renderer) RenderAnalyze(a *model.Analysis, top int, detailed bool) string {
	if a == nil || len(a.Results) == 0 {
		return r.RenderNoMatch()
	}
	if !detailed {
		return r.renderAnalyzeQuick(a, top)
	}
	results := topN(a.Results, top)
	var parts []string
	for i, result := range results {
		parts = append(parts, r.renderAnalyzeResult(a, result, i, len(results), detailed))
	}
	if detailed {
		if delta := r.renderDelta(a.Delta); delta != "" {
			parts = append(parts, r.renderDetailPanel("Delta Diagnosis", delta, "repo"))
		}
		if repo := r.renderRepoContext(a.RepoContext); repo != "" {
			parts = append(parts, r.renderDetailPanel("Repo Context", repo, "repo"))
		}
	}
	return strings.Join(parts, "\n\n")
}

func (r Renderer) renderAnalyzeQuick(a *model.Analysis, top int) string {
	results := topN(a.Results, top)
	if len(results) == 0 {
		return r.RenderNoMatch()
	}
	topResult := results[0]
	sections := []string{
		r.renderQuickDiagnosis(topResult),
	}

	if summary := r.renderMarkdownSection("Summary", topResult.Playbook.Summary); summary != "" {
		sections = append(sections, r.renderSection("Summary", summary))
	}
	if evidence := r.renderQuickEvidence(topResult.Evidence); evidence != "" {
		sections = append(sections, r.renderSection("Matched Evidence", evidence))
	}
	if actions := r.renderQuickActions(topResult); actions != "" {
		sections = append(sections, r.renderSection("Recommended Action", actions))
	}
	if alternatives := r.renderQuickAlternatives(results); alternatives != "" {
		sections = append(sections, r.renderSection("Other Likely Matches", alternatives))
	}
	if next := r.renderQuickNextSteps(len(results) > 1); next != "" {
		sections = append(sections, r.renderSection("More", next))
	}
	return strings.TrimSpace(strings.Join(filterEmpty(sections), "\n\n")) + "\n"
}

func (r Renderer) RenderFix(a *model.Analysis) string {
	if a == nil || len(a.Results) == 0 {
		return r.RenderNoMatch()
	}
	result := a.Results[0]
	header := r.renderHeader(result.Playbook, fmt.Sprintf("%d%% confidence", int(math.Round(result.Confidence*100))), "")
	body := r.renderMarkdownSection("Fix Steps", result.Playbook.Fix)
	if strings.TrimSpace(body) == "" {
		body = "No fix steps defined for this playbook."
	}
	return strings.TrimSpace(strings.Join([]string{
		header,
		r.renderMetaRows([]string{
			metaRow("ID", result.Playbook.ID),
			metaRow("Category", result.Playbook.Category),
		}),
		r.renderSection("Fix Steps", body),
	}, "\n\n")) + "\n"
}

func (r Renderer) RenderExplain(pb model.Playbook) string {
	sections := []string{
		r.renderHeader(pb, pb.ID, displayPackName(pb)),
		r.renderMetaRows([]string{
			metaRow("Category", pb.Category),
			metaRow("Severity", fallback(pb.Severity, "unknown")),
			metaRow("Detector", fallback(pb.Detector, "log")),
			metaRow("Tags", strings.Join(pb.Tags, ", ")),
			metaRow("Stages", strings.Join(pb.StageHints, ", ")),
		}),
	}

	if summary := r.renderMarkdownSection("Summary", pb.Summary); summary != "" {
		sections = append(sections, r.renderSection("Summary", summary))
	}
	if diagnosis := r.renderMarkdownSection("Diagnosis", pb.Diagnosis); diagnosis != "" {
		sections = append(sections, r.renderSection("Diagnosis", diagnosis))
	}
	if why := r.renderMarkdownSection("Why It Matters", pb.WhyItMatters); why != "" {
		sections = append(sections, r.renderSection("Why It Matters", why))
	}
	if fix := r.renderMarkdownSection("Fix Steps", pb.Fix); fix != "" {
		sections = append(sections, r.renderSection("Fix Steps", fix))
	}
	if validation := r.renderMarkdownSection("Validation", pb.Validation); validation != "" {
		sections = append(sections, r.renderSection("Validation", validation))
	}
	sections = append(sections, r.renderCallout("Match rules decide; markdown explains.\n\n"+r.renderMatchSummary(pb)))
	return strings.TrimSpace(strings.Join(filterEmpty(sections), "\n\n")) + "\n"
}

func (r Renderer) RenderList(playbooks []model.Playbook, category string) string {
	filter := strings.ToLower(strings.TrimSpace(category))
	lines := []string{r.renderListHeader()}
	for _, pb := range playbooks {
		if filter != "" && strings.ToLower(pb.Category) != filter {
			continue
		}
		lines = append(lines, r.renderListRow(pb))
	}
	return strings.Join(lines, "\n") + "\n"
}

func (r Renderer) renderAnalyzeResult(a *model.Analysis, result model.Result, rank, total int, detailed bool) string {
	subtitle := fmt.Sprintf("%d%% confidence", int(math.Round(result.Confidence*100)))
	if total > 1 {
		subtitle = fmt.Sprintf("#%d · %s", rank+1, subtitle)
	}

	parts := []string{
		r.renderHeader(result.Playbook, subtitle, displayPackName(result.Playbook)),
		r.renderMetaRows([]string{
			metaRow("Category", result.Playbook.Category),
			metaRow("Severity", fallback(result.Playbook.Severity, "unknown")),
			metaRow("Score", fmt.Sprintf("%.2f", result.Score)),
			metaRow("Detector", fallback(result.Detector, "log")),
			metaRow("Stage", a.Context.Stage),
		}),
	}

	if summary := r.renderMarkdownSection("Summary", result.Playbook.Summary); summary != "" {
		parts = append(parts, r.renderSection("Summary", summary))
	}
	if detailed && len(result.Evidence) > 0 {
		parts = append(parts, r.renderDetailPanel("Evidence", r.renderBulletLines(result.Evidence), "evidence"))
	}
	if detailed {
		if rank == 0 {
			if differential := r.renderDifferential(a); differential != "" {
				parts = append(parts, r.renderDetailPanel("Differential Diagnosis", differential, "signal"))
			}
			if confidence := r.renderConfidenceBreakdown(a.Results, result); confidence != "" {
				parts = append(parts, r.renderDetailPanel("Confidence Breakdown", confidence, "score"))
			}
		}
		if !sameTrimmedLines(result.Explanation.TriggeredBy, result.Evidence) {
			if triggered := r.renderExplanation("Triggered by", result.Explanation.TriggeredBy); triggered != "" {
				parts = append(parts, triggered)
			}
		}
		if amplified := r.renderExplanation("Amplified by", result.Explanation.AmplifiedBy); amplified != "" {
			parts = append(parts, amplified)
		}
		if mitigated := r.renderExplanation("Mitigated by", result.Explanation.MitigatedBy); mitigated != "" {
			parts = append(parts, mitigated)
		}
		if breakdown := r.renderScoreBreakdown(result.Breakdown); breakdown != "" {
			parts = append(parts, r.renderDetailPanel("Score Breakdown", breakdown, "score"))
		}
		if rank == 0 {
			if fix := r.renderMarkdownSection("Suggested Fix", result.Playbook.Fix); fix != "" {
				parts = append(parts, r.renderSection("Suggested Fix", fix))
			}
		}
	}
	return strings.TrimSpace(strings.Join(filterEmpty(parts), "\n\n"))
}

func (r Renderer) renderQuickDiagnosis(result model.Result) string {
	meta := []string{
		fmt.Sprintf("Confidence: %s (%d%%)", confidenceLabel(result.Confidence), int(math.Round(result.Confidence*100))),
		metaRow("Category", result.Playbook.Category),
		metaRow("Severity", fallback(result.Playbook.Severity, "unknown")),
	}
	body := strings.Join(filterEmpty([]string{
		result.Playbook.ID + "  " + result.Playbook.Title,
		metaRow("Pack", displayPackName(result.Playbook)),
		joinMetaInline(filterEmpty(meta)...),
	}), "\n")
	return r.renderSection("Most Likely Diagnosis", body)
}

func (r Renderer) renderQuickEvidence(lines []string) string {
	lines = trimLines(lines)
	if len(lines) == 0 {
		return ""
	}
	if len(lines) > 3 {
		lines = lines[:3]
	}
	if r.opts.Plain {
		var out []string
		for i, line := range lines {
			out = append(out, fmt.Sprintf("%d. %s", i+1, line))
		}
		return strings.Join(out, "\n")
	}
	return r.renderBulletLines(lines)
}

func (r Renderer) renderQuickActions(result model.Result) string {
	items := markdownListItems(result.Playbook.Fix)
	if len(items) == 0 {
		return "Use `faultline fix` to show the playbook remediation steps."
	}
	if len(items) > 2 {
		items = items[:2]
	}
	if r.opts.Plain {
		var out []string
		for i, item := range items {
			out = append(out, fmt.Sprintf("%d. %s", i+1, trimTerminalPunctuation(item)))
		}
		return strings.Join(out, "\n")
	}
	return r.renderBulletLines(items)
}

func (r Renderer) renderQuickAlternatives(results []model.Result) string {
	if len(results) < 2 {
		return ""
	}
	var lines []string
	for i, result := range results[1:] {
		lines = append(lines, fmt.Sprintf("#%d %s (%d%%)", i+2, result.Playbook.ID, int(math.Round(result.Confidence*100))))
	}
	if r.opts.Plain {
		return strings.Join(lines, "\n")
	}
	return r.renderBulletLines(lines)
}

func (r Renderer) renderQuickNextSteps(hasAlternatives bool) string {
	lines := []string{
		"Use `faultline fix` for remediation-only output.",
		"Use `faultline workflow` for deterministic follow-through steps.",
		"Use `faultline analyze --mode detailed` for full reasoning and scoring detail.",
	}
	if hasAlternatives {
		lines = append(lines, "Use `faultline analyze --select <n>` to inspect another ranked candidate.")
	}
	if r.opts.Plain {
		return strings.Join(lines, "\n")
	}
	return r.renderBulletLines(lines)
}

func (r Renderer) renderHeader(pb model.Playbook, subtitle, pack string) string {
	severity := fallback(pb.Severity, "unknown")
	if r.opts.Plain {
		header := fmt.Sprintf("%s (%s)", pb.Title, pb.ID)
		if subtitle != "" {
			header += " [" + subtitle + "]"
		}
		lines := []string{header}
		if pack != "" {
			lines = append(lines, "Pack: "+pack)
		}
		lines = append(lines, "Severity: "+severity)
		return strings.Join(lines, "\n")
	}

	title := r.styles.title.Render(pb.Title)
	severityStyle, ok := r.styles.severity[severity]
	if !ok {
		severityStyle = r.styles.severity["unknown"]
	}
	meta := []string{severityStyle.Render(strings.ToUpper(severity))}
	if subtitle != "" {
		meta = append(meta, r.styles.confidence.Render(subtitle))
	}
	if pack != "" {
		meta = append(meta, r.styles.muted.Render(pack))
	}
	body := title + "\n" + joinHorizontalWithGap("  ", meta...)
	return r.styles.card.Width(r.opts.Width - 2).Render(body)
}

func (r Renderer) renderMetaRows(rows []string) string {
	rows = filterEmpty(rows)
	if len(rows) == 0 {
		return ""
	}
	if r.opts.Plain {
		return r.renderMetaRowsPlain(rows)
	}
	return r.renderMetaRowsStyled(rows)
}

func (r Renderer) renderSection(title, body string) string {
	if strings.TrimSpace(body) == "" {
		return ""
	}
	if r.opts.Plain {
		return title + "\n" + strings.Repeat("-", len(title)) + "\n\n" + body
	}
	header := r.styles.subtitle.Render(title)
	divider := r.styles.divider.Render(strings.Repeat("─", min(r.opts.Width-2, 32)))
	return header + "\n" + divider + "\n\n" + body
}

func (r Renderer) renderDetailPanel(title, body, tone string) string {
	if strings.TrimSpace(body) == "" {
		return ""
	}
	if r.opts.Plain {
		return r.renderSection(title, body)
	}

	borderColor, badge := r.detailPanelStyles(title, tone)
	panel := r.styles.panel.
		BorderForeground(lipgloss.Color(borderColor)).
		Width(r.opts.Width - 2)

	return panel.Render(badge.Render(strings.ToUpper(title)) + "\n\n" + body)
}

func (r Renderer) renderCallout(body string) string {
	if strings.TrimSpace(body) == "" {
		return ""
	}
	if r.opts.Plain {
		return body
	}
	return r.styles.callout.Width(r.opts.Width - 2).Render(body)
}

func (r Renderer) renderMarkdown(markdown string) string {
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return ""
	}
	if r.opts.Plain {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithStandardStyle("notty"),
			glamour.WithWordWrap(r.markdownWidth()),
		)
		if err != nil {
			return markdown
		}
		defer renderer.Close()

		out, err := renderer.Render(markdown)
		if err != nil {
			return markdown
		}
		return strings.TrimRight(out, "\n")
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(markdownStyles(r.opts.DarkBackground)),
		glamour.WithWordWrap(r.markdownWidth()),
	)
	if err != nil {
		return markdown
	}
	defer renderer.Close()

	out, err := renderer.Render(markdown)
	if err != nil {
		return markdown
	}
	return strings.TrimRight(out, "\n")
}

func (r Renderer) renderMarkdownSection(title, markdown string) string {
	return r.renderMarkdown(trimRedundantHeading(markdown, title))
}

func (r Renderer) renderMatchSummary(pb model.Playbook) string {
	lines := []string{}
	if len(pb.Match.Any) > 0 {
		lines = append(lines, "match.any")
		for _, item := range pb.Match.Any {
			lines = append(lines, "- "+item)
		}
	}
	if len(pb.Match.All) > 0 {
		lines = append(lines, "match.all")
		for _, item := range pb.Match.All {
			lines = append(lines, "- "+item)
		}
	}
	if len(pb.Match.None) > 0 {
		lines = append(lines, "match.none")
		for _, item := range pb.Match.None {
			lines = append(lines, "- "+item)
		}
	}
	if len(pb.Workflow.Verify) > 0 {
		lines = append(lines, "workflow.verify")
		for _, item := range pb.Workflow.Verify {
			lines = append(lines, "- "+item)
		}
	}
	return strings.Join(lines, "\n")
}

func (r Renderer) renderBulletLines(lines []string) string {
	var b strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("- ")
		b.WriteString(line)
	}
	return b.String()
}

func (r Renderer) renderExplanation(title string, lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return r.renderDetailPanel(title, r.renderBulletLines(lines), "signal")
}

func (r Renderer) renderDifferential(a *model.Analysis) string {
	if a != nil && a.Differential != nil {
		if body := r.renderDifferentialSummary(a.Differential); body != "" {
			return body
		}
	}
	if a == nil || len(a.Results) < 2 {
		return ""
	}
	top := a.Results[0]
	runnerUp := a.Results[1]
	lines := []string{
		fmt.Sprintf("- Top candidate: %s (%s)", top.Playbook.ID, top.Playbook.Title),
		fmt.Sprintf("- Runner-up: %s (%s)", runnerUp.Playbook.ID, runnerUp.Playbook.Title),
	}
	gap := roundScore(top.Score - runnerUp.Score)
	if gap <= 0 {
		lines = append(lines, "- Score gap: tied on score; stable ordering kept the top candidate first.")
	} else {
		lines = append(lines, fmt.Sprintf("- Score gap: +%.2f over the runner-up", gap))
	}
	if reason := higherRankedReason(top, runnerUp); reason != "" {
		lines = append(lines, "- Higher-ranked because: "+reason)
	}
	if reason := alternateReason(top, runnerUp); reason != "" {
		lines = append(lines, "- Alternate remains plausible because: "+reason)
	}
	return strings.Join(lines, "\n")
}

func (r Renderer) renderDifferentialSummary(diff *model.DifferentialDiagnosis) string {
	if diff == nil {
		return ""
	}
	lines := make([]string, 0)
	if diff.Likely != nil {
		lines = append(lines, fmt.Sprintf("- Likely cause: %s (%s)", diff.Likely.FailureID, diff.Likely.Title))
		if diff.Likely.ConfidenceText != "" {
			lines = append(lines, "- Confidence: "+diff.Likely.ConfidenceText)
		}
		for _, item := range diff.Likely.Why {
			lines = append(lines, "- Evidence: "+item)
		}
		for _, item := range diff.Likely.DisproofChecks {
			lines = append(lines, "- Disproof check: "+item)
			break
		}
	}
	for _, item := range diff.Alternatives {
		lines = append(lines, fmt.Sprintf("- Alternative: %s (%s)", item.FailureID, item.Title))
		for _, reason := range item.WhyLessLikely {
			lines = append(lines, "- Why less likely: "+reason)
		}
	}
	for _, item := range diff.RuledOut {
		lines = append(lines, fmt.Sprintf("- Ruled out: %s (%s)", item.FailureID, item.Title))
		for _, reason := range item.RuledOutBy {
			lines = append(lines, "- Reason: "+reason)
		}
	}
	return strings.Join(lines, "\n")
}

func (r Renderer) renderConfidenceBreakdown(results []model.Result, result model.Result) string {
	lines := []string{
		fmt.Sprintf("- Reported confidence: %d%%", int(math.Round(result.Confidence*100))),
	}
	if result.Ranking == nil {
		lines = append(lines, fmt.Sprintf("- Detector score: %.2f", result.Score))
		return strings.Join(lines, "\n")
	}

	lines = append(lines,
		fmt.Sprintf("- Detector baseline: %.2f", result.Ranking.BaselineScore),
		fmt.Sprintf("- Final reranked score: %.2f", result.Ranking.FinalScore),
	)
	if result.Ranking.Prior != 0 {
		lines = append(lines, fmt.Sprintf("- Conservative prior: %+.2f", result.Ranking.Prior))
	}
	for _, item := range topSignalContributions(result.Ranking, 3) {
		lines = append(lines, fmt.Sprintf("- %+.2f %s", item.Contribution, fallback(item.Reason, item.Feature)))
	}
	if len(results) > 1 {
		gap := roundScore(result.Score - results[1].Score)
		if gap > 0 {
			lines = append(lines, fmt.Sprintf("- Margin over #2: +%.2f", gap))
		}
	}
	return strings.Join(lines, "\n")
}

func topSignalContributions(ranking *model.Ranking, limit int) []model.RankingContribution {
	if ranking == nil || limit <= 0 {
		return nil
	}
	out := make([]model.RankingContribution, 0, limit)
	for _, item := range ranking.Contributions {
		if contributionIsMetadata(item.Feature) {
			continue
		}
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func higherRankedReason(top, runnerUp model.Result) string {
	for _, item := range contributionDelta(top.Ranking, runnerUp.Ranking) {
		if item.Contribution <= 0 {
			continue
		}
		if contributionIsMetadata(item.Feature) {
			continue
		}
		reason := strings.TrimSpace(item.Reason)
		if reason == "" {
			reason = item.Feature
		}
		return reason
	}
	return ""
}

func alternateReason(top, runnerUp model.Result) string {
	if sharedEvidence(top.Evidence, runnerUp.Evidence) {
		return "it matched the same failing evidence line"
	}
	if len(trimLines(runnerUp.Evidence)) > 0 {
		return "it still matched explicit evidence from the input"
	}
	return ""
}

func contributionDelta(top, runnerUp *model.Ranking) []model.RankingContribution {
	if top == nil {
		return nil
	}
	runnerMap := map[string]model.RankingContribution{}
	if runnerUp != nil {
		for _, item := range runnerUp.Contributions {
			runnerMap[item.Feature] = item
		}
	}
	out := make([]model.RankingContribution, 0, len(top.Contributions))
	for _, item := range top.Contributions {
		if item.Feature == "detector_score" {
			continue
		}
		delta := roundScore(item.Contribution - runnerMap[item.Feature].Contribution)
		if delta == 0 {
			continue
		}
		item.Contribution = delta
		out = append(out, item)
	}
	return out
}

func contributionIsMetadata(feature string) bool {
	switch feature {
	case "detector_score", "historical_fixture_support", "candidate_separation":
		return true
	default:
		return false
	}
}

func sharedEvidence(left, right []string) bool {
	for _, l := range trimLines(left) {
		for _, r := range trimLines(right) {
			if l == r {
				return true
			}
		}
	}
	return false
}

func sameTrimmedLines(left, right []string) bool {
	left = trimLines(left)
	right = trimLines(right)
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

func trimLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}

func roundScore(value float64) float64 {
	return math.Round(value*100) / 100
}

func (r Renderer) renderScoreBreakdown(breakdown model.ScoreBreakdown) string {
	if breakdown.FinalScore == 0 {
		return ""
	}
	if breakdown.CompoundSignalBonus == 0 &&
		breakdown.BlastRadiusMultiplier == 0 &&
		breakdown.HotPathMultiplier == 0 &&
		breakdown.ChangeIntroducedBonus == 0 &&
		breakdown.MitigatingEvidenceDiscount == 0 &&
		breakdown.ExplicitExceptionDiscount == 0 &&
		breakdown.SafeContextDiscount == 0 {
		return ""
	}
	lines := []string{
		fmt.Sprintf("- base: %.2f", breakdown.BaseSignalScore),
		fmt.Sprintf("- final: %.2f", breakdown.FinalScore),
	}
	if breakdown.CompoundSignalBonus != 0 {
		lines = append(lines, fmt.Sprintf("- compound: +%.2f", breakdown.CompoundSignalBonus))
	}
	if breakdown.BlastRadiusMultiplier != 0 {
		lines = append(lines, fmt.Sprintf("- blast radius: +%.2f", breakdown.BlastRadiusMultiplier))
	}
	if breakdown.HotPathMultiplier != 0 {
		lines = append(lines, fmt.Sprintf("- hot path: +%.2f", breakdown.HotPathMultiplier))
	}
	if breakdown.ChangeIntroducedBonus != 0 {
		lines = append(lines, fmt.Sprintf("- change bonus: %+.2f", breakdown.ChangeIntroducedBonus))
	}
	if breakdown.MitigatingEvidenceDiscount != 0 {
		lines = append(lines, fmt.Sprintf("- mitigations: -%.2f", breakdown.MitigatingEvidenceDiscount))
	}
	if breakdown.ExplicitExceptionDiscount != 0 {
		lines = append(lines, fmt.Sprintf("- suppressions: -%.2f", breakdown.ExplicitExceptionDiscount))
	}
	if breakdown.SafeContextDiscount != 0 {
		lines = append(lines, fmt.Sprintf("- safe context: -%.2f", breakdown.SafeContextDiscount))
	}
	return strings.Join(lines, "\n")
}

func (r Renderer) renderRepoContext(repo *model.RepoContext) string {
	if repo == nil {
		return ""
	}
	lines := []string{}
	if repo.RepoRoot != "" {
		lines = append(lines, "- Repo root: "+repo.RepoRoot)
	}
	for _, item := range repo.RecentFiles {
		lines = append(lines, "- Recent file: "+item)
	}
	for _, commit := range repo.RelatedCommits {
		lines = append(lines, fmt.Sprintf("- Related commit: %s %s %s", commit.Date, commit.Hash, commit.Subject))
	}
	for _, item := range repo.HotspotDirectories {
		lines = append(lines, "- Hotspot area: "+item)
	}
	for _, item := range repo.CoChangeHints {
		lines = append(lines, "- Co-change: "+item)
	}
	for _, item := range repo.HotfixSignals {
		lines = append(lines, "- Hotfix signal: "+item)
	}
	for _, item := range repo.DriftSignals {
		lines = append(lines, "- Drift hint: "+item)
	}
	for _, item := range repo.ConfigDriftSignals {
		lines = append(lines, "- Config drift: "+item)
	}
	for _, item := range repo.CIChangeSignals {
		lines = append(lines, "- CI change: "+item)
	}
	for _, item := range repo.LargeCommitSignals {
		lines = append(lines, "- Large commit: "+item)
	}
	return strings.Join(lines, "\n")
}

func (r Renderer) renderRanking(ranking *model.Ranking) string {
	if ranking == nil {
		return ""
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("- mode: %s", ranking.Mode))
	lines = append(lines, fmt.Sprintf("- version: %s", ranking.Version))
	lines = append(lines, fmt.Sprintf("- baseline: %.2f", ranking.BaselineScore))
	lines = append(lines, fmt.Sprintf("- prior: %.2f", ranking.Prior))
	lines = append(lines, fmt.Sprintf("- final: %.2f", ranking.FinalScore))
	for _, item := range ranking.StrongestPositive {
		lines = append(lines, "- positive: "+item)
	}
	for _, item := range ranking.StrongestNegative {
		lines = append(lines, "- negative: "+item)
	}
	return strings.Join(lines, "\n")
}

func (r Renderer) renderDelta(delta *model.Delta) string {
	if delta == nil {
		return ""
	}
	var lines []string
	if strings.TrimSpace(delta.Provider) != "" {
		lines = append(lines, "- provider: "+delta.Provider)
	}
	for _, file := range delta.FilesChanged {
		lines = append(lines, "- changed file: "+file)
	}
	for _, test := range delta.TestsNewlyFailing {
		lines = append(lines, "- new failing test: "+test)
	}
	for _, item := range delta.ErrorsAdded {
		lines = append(lines, "- new error: "+item)
	}
	for _, cause := range delta.Causes {
		lines = append(lines, fmt.Sprintf("- %s: %.2f", cause.Kind, cause.Score))
		for _, reason := range cause.Reasons {
			lines = append(lines, "  - "+reason)
		}
	}
	return strings.Join(lines, "\n")
}

func (r Renderer) detailPanelStyles(title, tone string) (string, lipgloss.Style) {
	switch tone {
	case "evidence":
		return "#8B5A2B", panelTitleStyle("#8B5A2B", "#FFF7ED")
	case "score":
		return "#6D28D9", panelTitleStyle("#6D28D9", "#F5F3FF")
	case "repo":
		return "#0F766E", panelTitleStyle("#0F766E", "#ECFEFF")
	case "signal":
		switch strings.ToLower(strings.TrimSpace(title)) {
		case "triggered by":
			return "#0369A1", panelTitleStyle("#0369A1", "#E0F2FE")
		case "amplified by":
			return "#92400E", panelTitleStyle("#92400E", "#FEF3C7")
		case "mitigated by":
			return "#166534", panelTitleStyle("#166534", "#DCFCE7")
		}
	}
	return "#7C8798", panelTitleStyle("#475569", "#E2E8F0")
}

func (r Renderer) renderListHeader() string {
	if r.opts.Plain {
		return "ID\tCATEGORY\tSEVERITY\tPACK\tTITLE"
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		r.styles.metaLabel.Width(26).Render("ID"),
		r.styles.metaLabel.Width(12).Render("CATEGORY"),
		r.styles.metaLabel.Width(12).Render("SEVERITY"),
		r.styles.metaLabel.Width(22).Render("PACK"),
		r.styles.metaLabel.Render("TITLE"),
	)
}

func (r Renderer) renderListRow(pb model.Playbook) string {
	pack := fallback(displayPackName(pb), "-")
	if r.opts.Plain {
		return fmt.Sprintf("%s\t%s\t%s\t%s\t%s", pb.ID, pb.Category, pb.Severity, pack, pb.Title)
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Width(26).Render(pb.ID),
		r.styles.muted.Width(12).Render(pb.Category),
		r.styles.muted.Width(12).Render(fallback(pb.Severity, "unknown")),
		r.styles.muted.Width(22).Render(pack),
		lipgloss.NewStyle().Render(pb.Title),
	)
}

func (r Renderer) markdownWidth() int {
	width := r.opts.Width - 6
	if width < 48 {
		return 48
	}
	return width
}

func metaRow(label, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return label + ": " + value
}

func joinMetaInline(parts ...string) string {
	parts = filterEmpty(parts)
	return strings.Join(parts, "  ")
}

func (r Renderer) renderMetaRowsPlain(rows []string) string {
	pairs := chunkStrings(rows, 2)
	lines := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		lines = append(lines, strings.Join(pair, "  |  "))
	}
	return strings.Join(lines, "\n")
}

func (r Renderer) renderMetaRowsStyled(rows []string) string {
	columnWidth := (r.opts.Width - 6) / 2
	if columnWidth < 24 {
		columnWidth = 24
	}
	pairs := chunkStrings(rows, 2)
	lines := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		columns := make([]string, 0, len(pair))
		for _, row := range pair {
			columns = append(columns, r.renderMetaEntry(row, columnWidth))
		}
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, columns...))
	}
	return strings.Join(lines, "\n")
}

func (r Renderer) renderMetaEntry(row string, width int) string {
	label, value, ok := strings.Cut(row, ": ")
	if !ok {
		return r.styles.muted.Width(width).Render(row)
	}
	entry := r.styles.metaLabel.Render(strings.ToUpper(label)) + " " + r.styles.muted.Render(value)
	return lipgloss.NewStyle().Width(width).Render(entry)
}

func chunkStrings(values []string, size int) [][]string {
	if size <= 0 || len(values) == 0 {
		return nil
	}
	out := make([][]string, 0, (len(values)+size-1)/size)
	for start := 0; start < len(values); start += size {
		end := start + size
		if end > len(values) {
			end = len(values)
		}
		out = append(out, values[start:end])
	}
	return out
}

func filterEmpty(values []string) []string {
	out := values[:0]
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out
}

func fallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func joinHorizontalWithGap(gap string, parts ...string) string {
	parts = filterEmpty(parts)
	if len(parts) == 0 {
		return ""
	}
	joined := make([]string, 0, len(parts)*2-1)
	for i, part := range parts {
		if i > 0 {
			joined = append(joined, gap)
		}
		joined = append(joined, part)
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, joined...)
}

func trimRedundantHeading(markdown, title string) string {
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return ""
	}
	matches := leadingHeadingPattern.FindStringSubmatch(markdown)
	if len(matches) < 2 {
		return markdown
	}
	heading := normalizeHeading(matches[1])
	target := normalizeHeading(title)
	if heading == "" || target == "" {
		return markdown
	}
	if heading == target || strings.Contains(heading, target) || strings.Contains(target, heading) {
		return strings.TrimSpace(markdown[len(matches[0]):])
	}
	return markdown
}

func confidenceLabel(confidence float64) string {
	switch {
	case confidence >= 0.8:
		return "high"
	case confidence >= 0.5:
		return "medium"
	case confidence > 0:
		return "low"
	default:
		return "unknown"
	}
}

func markdownListItems(markdown string) []string {
	lines := strings.Split(markdown, "\n")
	items := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "- "):
			items = append(items, strings.TrimSpace(strings.TrimPrefix(line, "- ")))
		case len(line) > 3 && line[1] == '.' && line[2] == ' ' && line[0] >= '0' && line[0] <= '9':
			items = append(items, strings.TrimSpace(line[3:]))
		}
	}
	return items
}

func trimTerminalPunctuation(value string) string {
	return strings.TrimSpace(strings.TrimSuffix(value, "."))
}

func normalizeHeading(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var b strings.Builder
	lastSpace := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if !lastSpace {
			b.WriteByte(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

func displayPackName(pb model.Playbook) string {
	name := strings.TrimSpace(pb.Metadata.PackName)
	if name == "" || name == "starter" || name == "custom" {
		return ""
	}
	return name
}

func topN(results []model.Result, n int) []model.Result {
	if n <= 0 || n > len(results) {
		return results
	}
	return results[:n]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
