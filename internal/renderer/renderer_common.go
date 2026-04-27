package renderer

import (
	"fmt"
	"strings"

	glamour "charm.land/glamour/v2"
	lipgloss "charm.land/lipgloss/v2"

	"faultline/internal/model"
)

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
