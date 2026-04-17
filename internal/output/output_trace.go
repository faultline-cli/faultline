package output

import (
	"encoding/json"
	"fmt"
	"strings"

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
		Rules       []tracereport.Rule      `json:"rules"`
		Why         []string                `json:"why,omitempty"`
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
		Rules:       report.Rules,
		Why:         report.Why,
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

func joinTraceSection(title, body string) string {
	if strings.TrimSpace(body) == "" {
		return ""
	}
	return title + "\n" + strings.Repeat("-", len(title)) + "\n\n" + body
}
