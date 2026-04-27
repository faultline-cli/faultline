package renderer

import (
	"fmt"
	"math"
	"strings"
	"time"

	"faultline/internal/model"
)

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
	if history := r.renderHistorySummary(topResult); history != "" {
		sections = append(sections, r.renderSection("History", history))
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
	if history := r.renderHistorySummary(result); history != "" {
		parts = append(parts, r.renderDetailPanel("History", history, "repo"))
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

func (r Renderer) renderHistorySummary(result model.Result) string {
	lines := historySummaryLines(result)
	if len(lines) == 0 {
		return ""
	}
	if r.opts.Plain {
		return strings.Join(lines, "\n")
	}
	return r.renderBulletLines(lines)
}

func historySummaryLines(result model.Result) []string {
	var lines []string
	if sig := shortHistorySignature(result.SignatureHash); sig != "" {
		lines = append(lines, "history available for signature "+sig)
	}
	switch {
	case result.OccurrenceCount > 1:
		if span := historyWindow(result.FirstSeenAt, result.LastSeenAt); span != "" {
			lines = append(lines, fmt.Sprintf("seen %d times over %s in local history", result.OccurrenceCount, span))
		} else {
			lines = append(lines, fmt.Sprintf("seen %d times in local history", result.OccurrenceCount))
		}
	case result.OccurrenceCount == 1:
		lines = append(lines, "first recorded occurrence in local history")
	}
	if result.FirstSeenAt != "" {
		lines = append(lines, "first seen: "+result.FirstSeenAt)
	}
	if result.LastSeenAt != "" {
		lines = append(lines, "last seen: "+result.LastSeenAt)
	}
	if summary := hookHistorySummaryLine(result.HookHistorySummary); summary != "" {
		lines = append(lines, summary)
	}
	return lines
}

func hookHistorySummaryLine(summary *model.HookHistorySummary) string {
	if summary == nil || summary.TotalCount == 0 {
		return ""
	}
	parts := []string{fmt.Sprintf("hook verification history: %d run(s)", summary.TotalCount)}
	if summary.ExecutedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d executed", summary.ExecutedCount))
	}
	if summary.PassedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d passed", summary.PassedCount))
	}
	if summary.FailedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", summary.FailedCount))
	}
	if summary.BlockedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d blocked", summary.BlockedCount))
	}
	if summary.SkippedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", summary.SkippedCount))
	}
	if summary.LastSeenAt != "" {
		parts = append(parts, "last "+summary.LastSeenAt)
	}
	return strings.Join(parts, ", ")
}

func historyWindow(firstSeenAt, lastSeenAt string) string {
	start, err := time.Parse(time.RFC3339, strings.TrimSpace(firstSeenAt))
	if err != nil {
		return ""
	}
	end, err := time.Parse(time.RFC3339, strings.TrimSpace(lastSeenAt))
	if err != nil || end.Before(start) {
		return ""
	}
	duration := end.Sub(start)
	switch {
	case duration >= 48*time.Hour:
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	case duration >= time.Hour:
		return fmt.Sprintf("%dh", int(duration.Hours()))
	case duration >= time.Minute:
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	default:
		return ""
	}
}

func shortHistorySignature(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) > 12 {
		return value[:12]
	}
	return value
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
