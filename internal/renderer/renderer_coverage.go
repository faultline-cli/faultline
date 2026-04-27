package renderer

import (
	"fmt"
	"math"
	"strings"

	"faultline/internal/model"
)

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
