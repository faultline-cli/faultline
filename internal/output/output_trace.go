package output

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"faultline/internal/model"
	"faultline/internal/signature"
	tracereport "faultline/internal/trace"
)

// FormatTraceText formats a deterministic playbook trace for terminal output.
func FormatTraceText(report tracereport.Report, showEvidence, showScoring, showRejected bool) string {
	status := "not matched"
	if report.Matched {
		status = "matched"
		if report.Rank > 0 {
			status = fmt.Sprintf("matched and ranked #%d", report.Rank)
		}
	}

	lines := []string{
		fmt.Sprintf("TRACE  %s  [%s]", report.Playbook.ID, fallback(report.Detector, fallback(report.Playbook.Detector, "log"))),
		"Source: " + fallback(report.Source, "stdin"),
		"Outcome: " + status,
	}
	if report.Score > 0 || report.Confidence > 0 {
		lines = append(lines, fmt.Sprintf("Score: %.2f  Confidence: %d%%", report.Score, int(report.Confidence*100+0.5)))
	}

	sections := []string{strings.Join(lines, "\n")}
	sections = append(sections, renderTraceRulesText(report.Rules))
	if len(report.Why) > 0 {
		sections = append(sections, joinTraceSection("Why This Result", bulletLines(report.Why)))
	}
	if sig := renderTraceSignatureText(report.Signature); sig != "" {
		sections = append(sections, joinTraceSection("Signature", sig))
	}
	if history := renderTraceHistoryText(report); history != "" {
		sections = append(sections, joinTraceSection("History", history))
	}
	if hooks := renderTraceHooksText(report.Hooks); hooks != "" {
		sections = append(sections, joinTraceSection("Hooks", hooks))
	}
	if showScoring {
		if scoring := renderTraceScoringText(report); scoring != "" {
			sections = append(sections, joinTraceSection("Score", scoring))
		}
	}
	if showEvidence {
		if evidence := renderTraceEvidenceText(report); evidence != "" {
			sections = append(sections, joinTraceSection("Raw Evidence", evidence))
		}
	}
	if showRejected {
		if competing := renderTraceCompetingText(report.Competing); competing != "" {
			sections = append(sections, joinTraceSection("Competing Matches", competing))
		}
	}
	return strings.TrimSpace(strings.Join(sections, "\n\n")) + "\n"
}

// FormatTraceMarkdown formats a deterministic playbook trace as markdown.
func FormatTraceMarkdown(report tracereport.Report, showEvidence, showScoring, showRejected bool) string {
	status := "not matched"
	if report.Matched {
		status = "matched"
		if report.Rank > 0 {
			status = fmt.Sprintf("matched and ranked #%d", report.Rank)
		}
	}

	sections := []string{
		"# Faultline Trace",
		"",
		strings.Join([]string{
			"- Playbook: `" + report.Playbook.ID + "`",
			"- Title: " + report.Playbook.Title,
			"- Source: `" + fallback(report.Source, "stdin") + "`",
			"- Outcome: " + status,
		}, "\n"),
	}
	if report.Score > 0 || report.Confidence > 0 {
		sections = append(sections, "", "## Outcome", "", strings.Join([]string{
			fmt.Sprintf("- Score: %.2f", report.Score),
			fmt.Sprintf("- Confidence: %d%%", int(report.Confidence*100+0.5)),
		}, "\n"))
	}
	sections = append(sections, "", "## Rule Evaluation", "", renderTraceRulesMarkdown(report.Rules))
	if len(report.Why) > 0 {
		sections = append(sections, "", "## Why This Result", "", bulletLines(report.Why))
	}
	if sig := renderTraceSignatureMarkdown(report.Signature); sig != "" {
		sections = append(sections, "", "## Signature", "", sig)
	}
	if history := renderTraceHistoryMarkdown(report); history != "" {
		sections = append(sections, "", "## History", "", history)
	}
	if hooks := renderTraceHooksMarkdown(report.Hooks); hooks != "" {
		sections = append(sections, "", "## Hooks", "", hooks)
	}
	if showScoring {
		if scoring := renderTraceScoringMarkdown(report); scoring != "" {
			sections = append(sections, "", "## Score", "", scoring)
		}
	}
	if showEvidence {
		if evidence := renderTraceEvidenceMarkdown(report); evidence != "" {
			sections = append(sections, "", "## Raw Evidence", "", evidence)
		}
	}
	if showRejected {
		if competing := renderTraceCompetingMarkdown(report.Competing); competing != "" {
			sections = append(sections, "", "## Competing Matches", "", competing)
		}
	}
	return strings.TrimSpace(strings.Join(sections, "\n")) + "\n"
}

// FormatTraceJSON serializes a deterministic trace report.
func FormatTraceJSON(report tracereport.Report, showEvidence, showScoring, showRejected bool) (string, error) {
	payload := struct {
		Source      string                  `json:"source,omitempty"`
		Fingerprint string                  `json:"fingerprint,omitempty"`
		Context     interface{}             `json:"context,omitempty"`
		Playbook    string                  `json:"playbook_id"`
		Title       string                  `json:"title"`
		Matched     bool                    `json:"matched"`
		Rank        int                     `json:"rank,omitempty"`
		Score       float64                 `json:"score,omitempty"`
		Confidence  float64                 `json:"confidence,omitempty"`
		Detector    string                  `json:"detector,omitempty"`
		Signature   interface{}             `json:"signature,omitempty"`
		History     interface{}             `json:"history,omitempty"`
		Rules       []tracereport.Rule      `json:"rules"`
		Why         []string                `json:"why,omitempty"`
		Hooks       *model.HookReport       `json:"hooks,omitempty"`
		Scoring     interface{}             `json:"scoring,omitempty"`
		Ranking     interface{}             `json:"ranking,omitempty"`
		Competing   []tracereport.Candidate `json:"competing,omitempty"`
	}{
		Source:      report.Source,
		Fingerprint: report.Fingerprint,
		Context:     report.Context,
		Playbook:    report.Playbook.ID,
		Title:       report.Playbook.Title,
		Matched:     report.Matched,
		Rank:        report.Rank,
		Score:       report.Score,
		Confidence:  report.Confidence,
		Detector:    fallback(report.Detector, fallback(report.Playbook.Detector, "log")),
		Signature:   traceSignatureJSON(report.Signature),
		History:     traceHistoryJSON(report),
		Rules:       report.Rules,
		Why:         report.Why,
		Hooks:       report.Hooks,
	}
	if showScoring {
		payload.Scoring = report.Scoring
		payload.Ranking = report.Ranking
	}
	if showRejected {
		payload.Competing = report.Competing
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal trace JSON: %w", err)
	}
	return string(data) + "\n", nil
}

func renderTraceRulesText(rules []tracereport.Rule) string {
	var lines []string
	for _, rule := range rules {
		lines = append(lines, fmt.Sprintf("%-7s %s[%d]", strings.ToUpper(string(rule.Status)), rule.Group, rule.Index))
		lines = append(lines, "      pattern: "+rule.Pattern)
		if len(rule.LineMatches) > 0 {
			for _, match := range rule.LineMatches {
				lines = append(lines, fmt.Sprintf("      evidence: line %d: %s", match.Number, match.Text))
			}
		} else {
			lines = append(lines, "      evidence: none")
		}
		if rule.Note != "" {
			lines = append(lines, "      note: "+rule.Note)
		}
	}
	return joinTraceSection("Rule Evaluation", strings.Join(lines, "\n"))
}

func renderTraceRulesMarkdown(rules []tracereport.Rule) string {
	var lines []string
	for _, rule := range rules {
		lines = append(lines, fmt.Sprintf("- `%s` `%s[%d]`", strings.ToUpper(string(rule.Status)), rule.Group, rule.Index))
		lines = append(lines, "  pattern: `"+rule.Pattern+"`")
		if len(rule.LineMatches) > 0 {
			for _, match := range rule.LineMatches {
				lines = append(lines, fmt.Sprintf("  evidence: line %d: %s", match.Number, match.Text))
			}
		} else {
			lines = append(lines, "  evidence: none")
		}
		if rule.Note != "" {
			lines = append(lines, "  note: "+rule.Note)
		}
	}
	return strings.Join(lines, "\n")
}

func renderTraceScoringText(report tracereport.Report) string {
	var lines []string
	if report.Scoring != nil {
		lines = append(lines,
			fmt.Sprintf("Base score: %.2f", report.Scoring.BaseSignalScore),
			fmt.Sprintf("Final score: %.2f", report.Scoring.FinalScore),
		)
		if report.Scoring.CompoundSignalBonus != 0 {
			lines = append(lines, fmt.Sprintf("Compound bonus: +%.2f", report.Scoring.CompoundSignalBonus))
		}
	}
	if report.Ranking != nil {
		lines = append(lines, fmt.Sprintf("Ranking mode: %s", report.Ranking.Mode))
		lines = append(lines, fmt.Sprintf("Reranked score: %.2f", report.Ranking.FinalScore))
	}
	return strings.Join(lines, "\n")
}

func renderTraceScoringMarkdown(report tracereport.Report) string {
	lines := strings.Split(renderTraceScoringText(report), "\n")
	return bulletLines(lines)
}

func renderTraceSignatureText(sig *signature.ResultSignature) string {
	if sig == nil {
		return ""
	}
	return strings.Join([]string{
		"Hash: " + sig.Hash,
		"Version: " + sig.Version,
		"Payload: " + sig.Normalized,
	}, "\n")
}

func renderTraceSignatureMarkdown(sig *signature.ResultSignature) string {
	if sig == nil {
		return ""
	}
	return strings.Join([]string{
		"- Hash: `" + sig.Hash + "`",
		"- Version: `" + sig.Version + "`",
		"- Payload:",
		"```json",
		sig.Normalized,
		"```",
	}, "\n")
}

func renderTraceHistoryText(report tracereport.Report) string {
	return strings.Join(traceHistoryLines(report), "\n")
}

func renderTraceHistoryMarkdown(report tracereport.Report) string {
	return bulletLines(traceHistoryLines(report))
}

func traceHistoryLines(report tracereport.Report) []string {
	if report.OccurrenceCount == 0 && !report.SeenBefore && report.FirstSeenAt == "" && report.LastSeenAt == "" && report.HookHistory == nil {
		return nil
	}
	var lines []string
	if sig := shortTraceSignature(report.Signature); sig != "" {
		lines = append(lines, "History available for signature "+sig)
	}
	switch {
	case report.OccurrenceCount > 1:
		if span := traceHistoryWindow(report.FirstSeenAt, report.LastSeenAt); span != "" {
			lines = append(lines, fmt.Sprintf("Seen %d times over %s in local history", report.OccurrenceCount, span))
		} else {
			lines = append(lines, fmt.Sprintf("Seen %d times in local history", report.OccurrenceCount))
		}
	case report.OccurrenceCount == 1:
		lines = append(lines, "First recorded occurrence in local history")
	case report.SeenBefore:
		lines = append(lines, "Seen before in local history")
	}
	if report.FirstSeenAt != "" {
		lines = append(lines, "First seen: "+report.FirstSeenAt)
	}
	if report.LastSeenAt != "" {
		lines = append(lines, "Last seen: "+report.LastSeenAt)
	}
	if report.HookHistory != nil && report.HookHistory.TotalCount > 0 {
		lines = append(lines, hookHistoryTraceLine(report.HookHistory))
	}
	return lines
}

func hookHistoryTraceLine(summary *model.HookHistorySummary) string {
	parts := []string{fmt.Sprintf("Hook verification history: %d run(s)", summary.TotalCount)}
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

func traceHistoryWindow(firstSeenAt, lastSeenAt string) string {
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

func shortTraceSignature(sig *signature.ResultSignature) string {
	if sig == nil {
		return ""
	}
	value := strings.TrimSpace(sig.Hash)
	if len(value) > 12 {
		return value[:12]
	}
	return value
}

func renderTraceEvidenceText(report tracereport.Report) string {
	seen := make(map[string]struct{})
	var lines []string
	for _, rule := range report.Rules {
		for _, match := range rule.LineMatches {
			key := fmt.Sprintf("%d:%s", match.Number, match.Text)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			lines = append(lines, fmt.Sprintf("line %d: %s", match.Number, match.Text))
		}
	}
	return bulletLines(lines)
}

func renderTraceEvidenceMarkdown(report tracereport.Report) string {
	text := renderTraceEvidenceText(report)
	if strings.TrimSpace(text) == "" {
		return ""
	}
	raw := make([]string, 0)
	for _, line := range strings.Split(text, "\n") {
		raw = append(raw, strings.TrimPrefix(line, "- "))
	}
	return "```text\n" + strings.Join(raw, "\n") + "\n```"
}

func renderTraceCompetingText(candidates []tracereport.Candidate) string {
	var lines []string
	for _, item := range candidates {
		lines = append(lines, fmt.Sprintf("%s  %s (%s)", strings.ToUpper(item.Status), item.FailureID, item.Title))
		if item.ConfidenceText != "" {
			lines = append(lines, "  confidence: "+item.ConfidenceText)
		}
		for _, reason := range item.Reasons {
			lines = append(lines, "  - "+reason)
		}
	}
	return strings.Join(lines, "\n")
}

func renderTraceCompetingMarkdown(candidates []tracereport.Candidate) string {
	var lines []string
	for _, item := range candidates {
		lines = append(lines, fmt.Sprintf("- `%s` `%s` - %s", item.Status, item.FailureID, item.Title))
		for _, reason := range item.Reasons {
			lines = append(lines, "  "+reason)
		}
	}
	return strings.Join(lines, "\n")
}

func traceSignatureJSON(sig *signature.ResultSignature) interface{} {
	if sig == nil {
		return nil
	}
	return struct {
		Hash       string      `json:"hash"`
		Version    string      `json:"version"`
		Payload    interface{} `json:"payload"`
		Normalized string      `json:"normalized"`
	}{
		Hash:       sig.Hash,
		Version:    sig.Version,
		Payload:    sig.Payload,
		Normalized: sig.Normalized,
	}
}

func traceHistoryJSON(report tracereport.Report) interface{} {
	if report.OccurrenceCount == 0 && !report.SeenBefore && report.FirstSeenAt == "" && report.LastSeenAt == "" && report.HookHistory == nil {
		return nil
	}
	return struct {
		SeenBefore      bool                      `json:"seen_before,omitempty"`
		OccurrenceCount int                       `json:"occurrence_count,omitempty"`
		FirstSeenAt     string                    `json:"first_seen_at,omitempty"`
		LastSeenAt      string                    `json:"last_seen_at,omitempty"`
		HookHistory     *model.HookHistorySummary `json:"hook_history,omitempty"`
	}{
		SeenBefore:      report.SeenBefore,
		OccurrenceCount: report.OccurrenceCount,
		FirstSeenAt:     report.FirstSeenAt,
		LastSeenAt:      report.LastSeenAt,
		HookHistory:     report.HookHistory,
	}
}

func renderTraceHooksText(report *model.HookReport) string {
	lines := hookSummaryLines(report)
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

func renderTraceHooksMarkdown(report *model.HookReport) string {
	return bulletLines(hookSummaryLines(report))
}

func joinTraceSection(title, body string) string {
	if strings.TrimSpace(body) == "" {
		return ""
	}
	return title + "\n" + strings.Repeat("-", len(title)) + "\n\n" + body
}
