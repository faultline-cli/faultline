package renderer

import (
	"fmt"
	"math"
	"strings"

	lipgloss "charm.land/lipgloss/v2"

	"faultline/internal/model"
)

func (r Renderer) RenderFix(a *model.Analysis) string {
	if a == nil || len(a.Results) == 0 {
		return r.RenderNoMatch()
	}
	result := a.Results[0]
	fix := result.Playbook.Fix

	if r.opts.FixCommandsOnly {
		return r.renderFixCommandsOnly(result.Playbook.ID, fix)
	}

	// Strip opt-in sections from the primary fix render; they are shown
	// separately only when the corresponding flag is set.
	fixBase := stripFixSections(fix, "Preconditions", "Risks")
	body := r.renderMarkdownSection("Fix Steps", fixBase)
	if strings.TrimSpace(body) == "" {
		body = "No fix steps defined for this playbook."
	}
	sections := []string{
		r.renderHeader(result.Playbook, fmt.Sprintf("%d%% confidence", int(math.Round(result.Confidence*100))), ""),
		r.renderMetaRows([]string{
			metaRow("ID", result.Playbook.ID),
			metaRow("Category", result.Playbook.Category),
		}),
		r.renderSection("Fix Steps", body),
	}
	if r.opts.FixWithPreconditions {
		if sec := extractFixSection(fix, "Preconditions"); sec != "" {
			sections = append(sections, r.renderSection("Preconditions", sec))
		}
	}
	if r.opts.FixWithRisks {
		if sec := extractFixSection(fix, "Risks"); sec != "" {
			sections = append(sections, r.renderSection("Risks", sec))
		}
	}
	return strings.TrimSpace(strings.Join(sections, "\n\n")) + "\n"
}

// renderFixCommandsOnly emits only the runnable code blocks from the fix
// field, preceded by a compact header. It is used when --commands-only is set.
func (r Renderer) renderFixCommandsOnly(playbookID, fix string) string {
	commands := extractFixCodeBlocks(fix)
	var sb strings.Builder
	sb.WriteString(playbookID)
	sb.WriteString(": commands\n")
	sb.WriteString(strings.Repeat("─", min(len(playbookID)+10, r.opts.Width)))
	sb.WriteByte('\n')
	if len(commands) == 0 {
		sb.WriteString("No runnable commands found in fix steps.\n")
		return sb.String()
	}
	for i, cmd := range commands {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(cmd)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// extractFixCodeBlocks returns the content of each fenced code block in the
// fix markdown, without the fence delimiters.
func extractFixCodeBlocks(fix string) []string {
	var blocks []string
	lines := strings.Split(fix, "\n")
	var inBlock bool
	var current []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inBlock {
				inBlock = false
				if len(current) > 0 {
					blocks = append(blocks, strings.Join(current, "\n"))
				}
				current = nil
			} else {
				inBlock = true
			}
			continue
		}
		if inBlock {
			current = append(current, line)
		}
	}
	return blocks
}

// extractFixSection finds a heading-delimited section in fix markdown by its
// title (case-insensitive) and returns its content. Returns empty string when
// not found.
func extractFixSection(fix, title string) string {
	lines := strings.Split(fix, "\n")
	titleLower := strings.ToLower(title)
	var collecting bool
	var headPrefix string
	var result []string
	for _, line := range lines {
		stripped := strings.TrimLeft(line, "#")
		prefix := line[:len(line)-len(stripped)]
		content := strings.TrimSpace(stripped)
		if collecting {
			// Stop at next heading of same or higher level.
			if len(prefix) > 0 && len(prefix) <= len(headPrefix) {
				break
			}
			result = append(result, line)
			continue
		}
		if len(prefix) > 0 && strings.ToLower(content) == titleLower {
			collecting = true
			headPrefix = prefix
		}
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}

// stripFixSections removes named heading-delimited sections from fix markdown.
// This is used to exclude opt-in sections (e.g. Preconditions, Risks) from
// the default fix rendering.
func stripFixSections(fix string, titles ...string) string {
	titleSet := make(map[string]struct{}, len(titles))
	for _, t := range titles {
		titleSet[strings.ToLower(t)] = struct{}{}
	}
	lines := strings.Split(fix, "\n")
	var out []string
	var skipPrefix string
	for _, line := range lines {
		stripped := strings.TrimLeft(line, "#")
		prefix := line[:len(line)-len(stripped)]
		content := strings.ToLower(strings.TrimSpace(stripped))
		if skipPrefix != "" {
			// Inside a skipped section; stop skipping at same or higher heading.
			if len(prefix) > 0 && len(prefix) <= len(skipPrefix) {
				skipPrefix = ""
			} else {
				continue
			}
		}
		if len(prefix) > 0 {
			if _, skip := titleSet[content]; skip {
				skipPrefix = prefix
				continue
			}
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
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
