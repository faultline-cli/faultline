package renderer

import (
	"fmt"
	"math"
	"strings"

	glamour "charm.land/glamour/v2"
	lipgloss "charm.land/lipgloss/v2"

	"faultline/internal/model"
)

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
	results := topN(a.Results, top)
	var parts []string
	for i, result := range results {
		parts = append(parts, r.renderAnalyzeResult(a, result, i, len(results), detailed))
	}
	if detailed {
		if repo := r.renderRepoContext(a.RepoContext); repo != "" {
			parts = append(parts, r.renderSection("Repo Context", repo))
		}
	}
	return strings.Join(parts, "\n\n")
}

func (r Renderer) RenderFix(a *model.Analysis) string {
	if a == nil || len(a.Results) == 0 {
		return r.RenderNoMatch()
	}
	result := a.Results[0]
	header := r.renderHeader(result.Playbook, fmt.Sprintf("%d%% confidence", int(math.Round(result.Confidence*100))), "")
	body := r.renderMarkdown(result.Playbook.Fix)
	if strings.TrimSpace(body) == "" {
		body = "No fix steps defined for this playbook."
	}
	return strings.TrimSpace(strings.Join([]string{
		header,
		r.renderMetaRows([]string{
			metaRow("ID", result.Playbook.ID),
			metaRow("Category", result.Playbook.Category),
		}),
		body,
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

	if summary := r.renderMarkdown(pb.Summary); summary != "" {
		sections = append(sections, r.renderSection("Summary", summary))
	}
	if diagnosis := r.renderMarkdown(pb.Diagnosis); diagnosis != "" {
		sections = append(sections, r.renderSection("Diagnosis", diagnosis))
	}
	if why := r.renderMarkdown(pb.WhyItMatters); why != "" {
		sections = append(sections, r.renderSection("Why It Matters", why))
	}
	if fix := r.renderMarkdown(pb.Fix); fix != "" {
		sections = append(sections, r.renderSection("Fix", fix))
	}
	if validation := r.renderMarkdown(pb.Validation); validation != "" {
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

	if summary := r.renderMarkdown(result.Playbook.Summary); summary != "" {
		parts = append(parts, r.renderSection("Summary", summary))
	}
	if detailed && len(result.Evidence) > 0 {
		parts = append(parts, r.renderSection("Evidence", r.renderBulletLines(result.Evidence)))
	}
	if detailed {
		if triggered := r.renderExplanation("Triggered by", result.Explanation.TriggeredBy); triggered != "" {
			parts = append(parts, triggered)
		}
		if amplified := r.renderExplanation("Amplified by", result.Explanation.AmplifiedBy); amplified != "" {
			parts = append(parts, amplified)
		}
		if mitigated := r.renderExplanation("Mitigated by", result.Explanation.MitigatedBy); mitigated != "" {
			parts = append(parts, mitigated)
		}
		if breakdown := r.renderScoreBreakdown(result.Breakdown); breakdown != "" {
			parts = append(parts, r.renderSection("Score Breakdown", breakdown))
		}
	}
	return strings.TrimSpace(strings.Join(filterEmpty(parts), "\n\n"))
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
	body := title + "\n" + lipgloss.JoinHorizontal(lipgloss.Left, meta...)
	return r.styles.card.Width(r.opts.Width - 2).Render(body)
}

func (r Renderer) renderMetaRows(rows []string) string {
	rows = filterEmpty(rows)
	if len(rows) == 0 {
		return ""
	}
	if r.opts.Plain {
		return strings.Join(rows, "\n")
	}
	return r.styles.muted.Render(strings.Join(rows, "\n"))
}

func (r Renderer) renderSection(title, body string) string {
	if strings.TrimSpace(body) == "" {
		return ""
	}
	if r.opts.Plain {
		return title + "\n" + strings.Repeat("-", len(title)) + "\n" + body
	}
	header := r.styles.subtitle.Render(title)
	divider := r.styles.divider.Render(strings.Repeat("─", min(r.opts.Width-2, 32)))
	return header + "\n" + divider + "\n" + body
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
	style := "notty"
	if !r.opts.Plain {
		style = "dark"
		if !r.opts.DarkBackground {
			style = "light"
		}
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
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
	return r.renderSection(title, r.renderBulletLines(lines))
}

func (r Renderer) renderScoreBreakdown(breakdown model.ScoreBreakdown) string {
	if breakdown.FinalScore == 0 {
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
		lines = append(lines, fmt.Sprintf("- change bonus: %.2f", breakdown.ChangeIntroducedBonus))
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
	return strings.Join(lines, "\n")
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
