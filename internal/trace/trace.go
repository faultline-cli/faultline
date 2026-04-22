package trace

import (
	"fmt"
	"strings"

	"faultline/internal/engine"
	"faultline/internal/model"
	"faultline/internal/signature"
)

type RuleStatus string

const (
	StatusMatched RuleStatus = "matched"
	StatusMissing RuleStatus = "missing"
	StatusClear   RuleStatus = "clear"
	StatusBlocked RuleStatus = "blocked"
)

type LineMatch struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
}

type Rule struct {
	Group       string      `json:"group"`
	Index       int         `json:"index"`
	Pattern     string      `json:"pattern"`
	Status      RuleStatus  `json:"status"`
	Matched     bool        `json:"matched"`
	Relevance   string      `json:"relevance,omitempty"`
	Note        string      `json:"note,omitempty"`
	LineMatches []LineMatch `json:"line_matches,omitempty"`
}

type Candidate struct {
	Status         string   `json:"status"`
	FailureID      string   `json:"failure_id"`
	Title          string   `json:"title"`
	Confidence     float64  `json:"confidence,omitempty"`
	ConfidenceText string   `json:"confidence_text,omitempty"`
	Reasons        []string `json:"reasons,omitempty"`
}

type Report struct {
	Source      string                     `json:"source,omitempty"`
	Fingerprint string                     `json:"fingerprint,omitempty"`
	Context     model.Context              `json:"context,omitempty"`
	Playbook    model.Playbook             `json:"playbook"`
	Signature   *signature.ResultSignature `json:"signature,omitempty"`
	Rank        int                        `json:"rank,omitempty"`
	Matched     bool                       `json:"matched"`
	Score       float64                    `json:"score,omitempty"`
	Confidence  float64                    `json:"confidence,omitempty"`
	Detector    string                     `json:"detector,omitempty"`
	Summary     string                     `json:"summary,omitempty"`
	Rules       []Rule                     `json:"rules"`
	Why         []string                   `json:"why,omitempty"`
	Scoring     *model.ScoreBreakdown      `json:"scoring,omitempty"`
	Ranking     *model.Ranking             `json:"ranking,omitempty"`
	Competing   []Candidate                `json:"competing,omitempty"`
	Hooks       *model.HookReport          `json:"hooks,omitempty"`
}

func Build(analysis *model.Analysis, lines []model.Line, playbooks []model.Playbook, playbookID string, includeRejected bool) (Report, error) {
	pb, result, rank, err := resolvePlaybook(analysis, playbooks, playbookID)
	if err != nil {
		return Report{}, err
	}

	report := Report{
		Playbook: pb,
		Rules:    buildRules(pb, lines),
	}
	if analysis != nil {
		report.Source = analysis.Source
		report.Fingerprint = analysis.Fingerprint
		report.Context = analysis.Context
	}
	if result != nil {
		report.Rank = rank
		report.Matched = true
		sig := signature.ForResult(*result)
		report.Signature = &sig
		report.Score = result.Score
		report.Confidence = result.Confidence
		report.Detector = result.Detector
		report.Summary = pb.Summary
		scoring := result.Breakdown
		report.Scoring = &scoring
		report.Ranking = result.Ranking
		report.Why = buildWhy(*result, report.Rules)
	} else {
		report.Matched = false
		report.Summary = pb.Summary
		report.Why = buildUnmatchedWhy(report.Rules)
	}
	if includeRejected && analysis != nil {
		report.Competing = buildCompeting(analysis, pb.ID)
	}
	return report, nil
}

func resolvePlaybook(analysis *model.Analysis, playbooks []model.Playbook, playbookID string) (model.Playbook, *model.Result, int, error) {
	if playbookID == "" && analysis != nil && len(analysis.Results) > 0 {
		top := analysis.Results[0]
		return top.Playbook, &top, 1, nil
	}

	if analysis != nil {
		for i := range analysis.Results {
			if analysis.Results[i].Playbook.ID == playbookID {
				return analysis.Results[i].Playbook, &analysis.Results[i], i + 1, nil
			}
		}
	}

	for _, pb := range playbooks {
		if pb.ID == playbookID {
			return pb, nil, 0, nil
		}
	}

	if playbookID == "" {
		return model.Playbook{}, nil, 0, fmt.Errorf("no matched playbook is available to trace")
	}
	return model.Playbook{}, nil, 0, fmt.Errorf("playbook %q not found", playbookID)
}

func buildRules(pb model.Playbook, lines []model.Line) []Rule {
	rules := make([]Rule, 0, len(pb.Match.Any)+len(pb.Match.All)+len(pb.Match.None))
	for i, pattern := range pb.Match.Any {
		lineMatches := matchLines(pattern, lines)
		rules = append(rules, Rule{
			Group:       "match.any",
			Index:       i,
			Pattern:     pattern,
			Status:      matchedStatus(len(lineMatches) > 0),
			Matched:     len(lineMatches) > 0,
			Relevance:   "trigger",
			Note:        anyRuleNote(len(lineMatches) > 0),
			LineMatches: lineMatches,
		})
	}
	for i, pattern := range pb.Match.All {
		lineMatches := matchLines(pattern, lines)
		rules = append(rules, Rule{
			Group:       "match.all",
			Index:       i,
			Pattern:     pattern,
			Status:      matchedStatus(len(lineMatches) > 0),
			Matched:     len(lineMatches) > 0,
			Relevance:   "required",
			Note:        allRuleNote(len(lineMatches) > 0),
			LineMatches: lineMatches,
		})
	}
	for i, pattern := range pb.Match.None {
		lineMatches := matchLines(pattern, lines)
		blocked := len(lineMatches) > 0
		rules = append(rules, Rule{
			Group:       "match.none",
			Index:       i,
			Pattern:     pattern,
			Status:      noneStatus(blocked),
			Matched:     blocked,
			Relevance:   "exclusion",
			Note:        noneRuleNote(blocked),
			LineMatches: lineMatches,
		})
	}
	return rules
}

func buildWhy(result model.Result, rules []Rule) []string {
	if result.Hypothesis != nil {
		if len(result.Hypothesis.Why) > 0 {
			return append([]string(nil), result.Hypothesis.Why...)
		}
		if len(result.Hypothesis.WhyLessLikely) > 0 {
			return append([]string(nil), result.Hypothesis.WhyLessLikely...)
		}
	}

	var reasons []string
	matchedAny := countGroup(rules, "match.any", StatusMatched)
	matchedAll := countGroup(rules, "match.all", StatusMatched)
	clearNone := countGroup(rules, "match.none", StatusClear)
	blockedNone := countGroup(rules, "match.none", StatusBlocked)

	if matchedAny > 0 {
		reasons = append(reasons, fmt.Sprintf("%d trigger rule(s) matched explicit log evidence", matchedAny))
	}
	if matchedAll > 0 {
		reasons = append(reasons, fmt.Sprintf("%d required rule(s) matched", matchedAll))
	}
	if clearNone > 0 {
		reasons = append(reasons, fmt.Sprintf("%d exclusion rule(s) stayed clear", clearNone))
	}
	if blockedNone > 0 {
		reasons = append(reasons, fmt.Sprintf("%d exclusion rule(s) were triggered", blockedNone))
	}
	if len(result.Evidence) > 0 {
		reasons = append(reasons, "matched evidence was pulled directly from the input log")
	}
	return reasons
}

func buildUnmatchedWhy(rules []Rule) []string {
	var reasons []string
	matchedAny := countGroup(rules, "match.any", StatusMatched)
	missingAll := countGroup(rules, "match.all", StatusMissing)
	blockedNone := countGroup(rules, "match.none", StatusBlocked)
	if matchedAny == 0 {
		reasons = append(reasons, "no trigger rule matched the input log")
	}
	if missingAll > 0 {
		reasons = append(reasons, fmt.Sprintf("%d required rule(s) were missing", missingAll))
	}
	if blockedNone > 0 {
		reasons = append(reasons, fmt.Sprintf("%d exclusion rule(s) blocked the playbook", blockedNone))
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "the playbook did not reach a ranked match")
	}
	return reasons
}

func buildCompeting(analysis *model.Analysis, currentID string) []Candidate {
	if analysis == nil || analysis.Differential == nil {
		return nil
	}
	var out []Candidate
	diff := analysis.Differential
	if diff.Likely != nil && diff.Likely.FailureID != "" && diff.Likely.FailureID != currentID {
		out = append(out, Candidate{
			Status:         "higher_ranked",
			FailureID:      diff.Likely.FailureID,
			Title:          diff.Likely.Title,
			Confidence:     diff.Likely.Confidence,
			ConfidenceText: diff.Likely.ConfidenceText,
			Reasons:        append([]string(nil), diff.Likely.Why...),
		})
	}
	for _, item := range diff.Alternatives {
		if item.FailureID == currentID {
			continue
		}
		out = append(out, Candidate{
			Status:         "alternative",
			FailureID:      item.FailureID,
			Title:          item.Title,
			Confidence:     item.Confidence,
			ConfidenceText: item.ConfidenceText,
			Reasons:        append([]string(nil), item.WhyLessLikely...),
		})
	}
	for _, item := range diff.RuledOut {
		if item.FailureID == currentID {
			continue
		}
		out = append(out, Candidate{
			Status:         "ruled_out",
			FailureID:      item.FailureID,
			Title:          item.Title,
			Confidence:     item.Confidence,
			ConfidenceText: item.ConfidenceText,
			Reasons:        append([]string(nil), item.RuledOutBy...),
		})
	}
	return out
}

func matchLines(pattern string, lines []model.Line) []LineMatch {
	norm := engine.NormalizeLine(pattern)
	if norm == "" {
		return nil
	}
	matches := make([]LineMatch, 0, 1)
	for _, line := range lines {
		if strings.Contains(line.Normalized, norm) {
			matches = append(matches, LineMatch{
				Number: line.Number,
				Text:   line.Original,
			})
			break
		}
	}
	return matches
}

func matchedStatus(matched bool) RuleStatus {
	if matched {
		return StatusMatched
	}
	return StatusMissing
}

func noneStatus(blocked bool) RuleStatus {
	if blocked {
		return StatusBlocked
	}
	return StatusClear
}

func countGroup(rules []Rule, group string, status RuleStatus) int {
	count := 0
	for _, rule := range rules {
		if rule.Group == group && rule.Status == status {
			count++
		}
	}
	return count
}

func anyRuleNote(matched bool) string {
	if matched {
		return "trigger rule matched the log"
	}
	return "trigger rule did not match"
}

func allRuleNote(matched bool) string {
	if matched {
		return "required rule matched"
	}
	return "required rule was missing"
}

func noneRuleNote(blocked bool) string {
	if blocked {
		return "exclusion rule matched and blocks the playbook"
	}
	return "exclusion rule stayed clear"
}
