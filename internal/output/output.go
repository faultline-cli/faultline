// Package output formats analysis results for humans, automation, and CI
// annotation consumers.  All functions accept a *model.Analysis that may be
// nil (when no log was provided) or have an empty Results slice (when no
// playbook matched).
package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"faultline/internal/model"
	"faultline/internal/renderer"
	"faultline/internal/workflow"
)

// Mode selects the verbosity of human-readable output.
type Mode string

const (
	ModeQuick    Mode = "quick"
	ModeDetailed Mode = "detailed"
)

// Format selects the human-readable output shape.
type Format string

const (
	FormatRaw      Format = "raw"
	FormatMarkdown Format = "markdown"
)

func (f Format) Valid() bool {
	switch f {
	case FormatRaw, FormatMarkdown:
		return true
	default:
		return false
	}
}

// ── Human-readable text ──────────────────────────────────────────────────────

// FormatAnalysisText formats an analysis for human consumption.
// top limits the number of results shown (0 or negative means show all).
func FormatAnalysisText(a *model.Analysis, top int, mode Mode, opts renderer.Options) string {
	return renderer.New(opts).RenderAnalyze(a, top, mode == ModeDetailed)
}

// FormatAnalysisMarkdown formats an analysis as raw markdown.
func FormatAnalysisMarkdown(a *model.Analysis, top int, mode Mode) string {
	if a == nil || len(a.Results) == 0 {
		return strings.Join([]string{
			"# No Match",
			"",
			"No known failure pattern matched.",
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
		if repo := markdownListSection("## Repo Context", repoContextLines(a.RepoContext)); repo != "" {
			sections = append(sections, repo)
		}
	}
	return strings.TrimSpace(strings.Join(sections, "\n\n")) + "\n"
}

// FormatFix formats only the fix steps for the top result.
func FormatFix(a *model.Analysis, opts renderer.Options) string {
	return renderer.New(opts).RenderFix(a)
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
	if fix := strings.TrimSpace(result.Playbook.FixMarkdown); fix != "" {
		sections = append(sections, "## Fix", "", fix)
	} else {
		sections = append(sections, "## Fix", "", "No fix steps defined for this playbook.")
	}
	return strings.TrimSpace(strings.Join(sections, "\n")) + "\n"
}

// ── JSON ─────────────────────────────────────────────────────────────────────

// analysisJSON is the stable JSON schema emitted by FormatAnalysisJSON.
type analysisJSON struct {
	Matched     bool         `json:"matched"`
	Source      string       `json:"source,omitempty"`
	Fingerprint string       `json:"fingerprint,omitempty"`
	Context     ctxJSON      `json:"context"`
	Results     []resultJSON `json:"results"`
	RepoContext *repoCtxJSON `json:"repo_context,omitempty"`
	Message     string       `json:"message,omitempty"`
}

type ctxJSON struct {
	Stage       string `json:"stage,omitempty"`
	CommandHint string `json:"command_hint,omitempty"`
	Step        string `json:"step,omitempty"`
}

type resultJSON struct {
	Rank                 int                     `json:"rank"`
	FailureID            string                  `json:"failure_id"`
	Title                string                  `json:"title"`
	Category             string                  `json:"category"`
	Pack                 string                  `json:"pack,omitempty"`
	Severity             string                  `json:"severity,omitempty"`
	Detector             string                  `json:"detector,omitempty"`
	Score                float64                 `json:"score"`
	Confidence           float64                 `json:"confidence"`
	Summary              string                  `json:"summary,omitempty"`
	DiagnosisMarkdown    string                  `json:"diagnosis_markdown,omitempty"`
	WhyItMattersMarkdown string                  `json:"why_it_matters_markdown,omitempty"`
	FixMarkdown          string                  `json:"fix_markdown,omitempty"`
	ValidationMarkdown   string                  `json:"validation_markdown,omitempty"`
	Evidence             []string                `json:"evidence"`
	EvidenceBy           model.EvidenceBundle    `json:"evidence_by,omitempty"`
	Explanation          model.ResultExplanation `json:"explanation,omitempty"`
	Breakdown            model.ScoreBreakdown    `json:"breakdown,omitempty"`
	ChangeStatus         string                  `json:"change_status,omitempty"`
	SeenCount            int                     `json:"seen_count"`
}

type repoCtxJSON struct {
	RepoRoot           string           `json:"repo_root"`
	RecentFiles        []string         `json:"recent_files,omitempty"`
	RelatedCommits     []repoCommitJSON `json:"related_commits,omitempty"`
	HotspotDirectories []string         `json:"hotspot_directories,omitempty"`
	CoChangeHints      []string         `json:"co_change_hints,omitempty"`
	HotfixSignals      []string         `json:"hotfix_signals,omitempty"`
	DriftSignals       []string         `json:"drift_signals,omitempty"`
}

type repoCommitJSON struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
}

// FormatAnalysisJSON serialises an analysis to the stable JSON schema.
func FormatAnalysisJSON(a *model.Analysis, top int) (string, error) {
	payload := analysisJSON{
		Matched: a != nil && len(a.Results) > 0,
	}

	if a == nil {
		payload.Message = "No known failure pattern matched."
		payload.Results = []resultJSON{}
		data, err := json.Marshal(payload)
		if err != nil {
			return "", fmt.Errorf("marshal analysis JSON: %w", err)
		}
		return string(data) + "\n", nil
	}

	payload.Source = a.Source
	payload.Fingerprint = a.Fingerprint
	payload.Context = ctxJSON{
		Stage:       a.Context.Stage,
		CommandHint: a.Context.CommandHint,
		Step:        a.Context.Step,
	}
	payload.RepoContext = repoContextJSON(a.RepoContext)

	if !payload.Matched {
		payload.Message = "No known failure pattern matched."
		payload.Results = []resultJSON{}
	} else {
		results := topN(a.Results, top)
		payload.Results = make([]resultJSON, len(results))
		for i, r := range results {
			payload.Results[i] = resultJSON{
				Rank:                 i + 1,
				FailureID:            r.Playbook.ID,
				Title:                r.Playbook.Title,
				Category:             r.Playbook.Category,
				Pack:                 displayPackName(r.Playbook),
				Severity:             r.Playbook.Severity,
				Detector:             r.Detector,
				Score:                r.Score,
				Confidence:           r.Confidence,
				Summary:              r.Playbook.Summary,
				DiagnosisMarkdown:    r.Playbook.DiagnosisMarkdown,
				WhyItMattersMarkdown: r.Playbook.WhyItMattersMarkdown,
				FixMarkdown:          r.Playbook.FixMarkdown,
				ValidationMarkdown:   r.Playbook.ValidationMarkdown,
				Evidence:             r.Evidence,
				EvidenceBy:           r.EvidenceBy,
				Explanation:          r.Explanation,
				Breakdown:            r.Breakdown,
				ChangeStatus:         r.ChangeStatus,
				SeenCount:            r.SeenCount,
			}
		}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal analysis JSON: %w", err)
	}
	return string(data) + "\n", nil
}

func repoContextJSON(repoCtx *model.RepoContext) *repoCtxJSON {
	if repoCtx == nil {
		return nil
	}

	out := &repoCtxJSON{
		RepoRoot:           repoCtx.RepoRoot,
		RecentFiles:        repoCtx.RecentFiles,
		HotspotDirectories: repoCtx.HotspotDirectories,
		CoChangeHints:      repoCtx.CoChangeHints,
		HotfixSignals:      repoCtx.HotfixSignals,
		DriftSignals:       repoCtx.DriftSignals,
	}
	if len(repoCtx.RelatedCommits) > 0 {
		out.RelatedCommits = make([]repoCommitJSON, len(repoCtx.RelatedCommits))
		for i, commit := range repoCtx.RelatedCommits {
			out.RelatedCommits[i] = repoCommitJSON{
				Hash:    commit.Hash,
				Subject: commit.Subject,
				Date:    commit.Date,
			}
		}
	}
	return out
}

// ── CI annotations ───────────────────────────────────────────────────────────

// FormatCIAnnotations emits GitHub Actions-compatible ::warning:: annotations,
// one per matched result up to top.
func FormatCIAnnotations(a *model.Analysis, top int) string {
	if a == nil || len(a.Results) == 0 {
		return ""
	}
	var b strings.Builder
	for _, r := range topN(a.Results, top) {
		fix := ""
		if first := firstMarkdownListItem(r.Playbook.FixMarkdown); first != "" {
			fix = " Fix: " + first
		}
		fmt.Fprintf(&b, "::warning title=%s::%s.%s\n",
			r.Playbook.ID, r.Playbook.Title, fix)
	}
	return b.String()
}

// ── Playbook list & details ──────────────────────────────────────────────────

// FormatPlaybookList formats a tab-aligned table of available playbooks.
// When category is non-empty only matching playbooks are shown.
func FormatPlaybookList(playbooks []model.Playbook, category string, opts renderer.Options) string {
	return renderer.New(opts).RenderList(playbooks, category)
}

// FormatPlaybookDetails formats all fields of a single playbook for the
// explain command.
func FormatPlaybookDetails(pb model.Playbook, opts renderer.Options) string {
	return renderer.New(opts).RenderExplain(pb)
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
	if diagnosis := strings.TrimSpace(pb.DiagnosisMarkdown); diagnosis != "" {
		sections = append(sections, "## Diagnosis", "", diagnosis)
	}
	if why := strings.TrimSpace(pb.WhyItMattersMarkdown); why != "" {
		sections = append(sections, "## Why It Matters", "", why)
	}
	if fix := strings.TrimSpace(pb.FixMarkdown); fix != "" {
		sections = append(sections, "## Fix", "", fix)
	}
	if validation := strings.TrimSpace(pb.ValidationMarkdown); validation != "" {
		sections = append(sections, "## Validation", "", validation)
	}
	if matchRules := formatMatchSummaryMarkdown(pb); matchRules != "" {
		sections = append(sections, "## Match Rules", "", "Structured fields decide; markdown explains.", "", matchRules)
	}
	return strings.TrimSpace(strings.Join(sections, "\n")) + "\n"
}

func displayPackName(pb model.Playbook) string {
	name := strings.TrimSpace(pb.Metadata.PackName)
	if name == "" || name == "starter" || name == "custom" {
		return ""
	}
	return name
}

// ── Workflow output ──────────────────────────────────────────────────────────

type workflowJSON struct {
	SchemaVersion string   `json:"schema_version"`
	Mode          string   `json:"mode"`
	FailureID     string   `json:"failure_id,omitempty"`
	Title         string   `json:"title,omitempty"`
	Source        string   `json:"source,omitempty"`
	Context       ctxJSON  `json:"context"`
	Evidence      []string `json:"evidence"`
	Files         []string `json:"files,omitempty"`
	LocalRepro    []string `json:"local_repro,omitempty"`
	Verify        []string `json:"verify,omitempty"`
	Steps         []string `json:"steps"`
	AgentPrompt   string   `json:"agent_prompt,omitempty"`
}

// FormatWorkflowText formats a deterministic workflow follow-up plan.
func FormatWorkflowText(plan workflow.Plan) string {
	var b strings.Builder
	if plan.FailureID == "" {
		fmt.Fprintln(&b, "WORKFLOW")
		for i, step := range plan.Steps {
			fmt.Fprintf(&b, "  %d. %s\n", i+1, step)
		}
		return b.String()
	}

	fmt.Fprintf(&b, "WORKFLOW  %s · %s  [%s · %s]\n", plan.FailureID, plan.Title, plan.Mode, plan.SchemaVersion)
	if plan.Source != "" {
		fmt.Fprintf(&b, "Source: %s\n", plan.Source)
	}
	if plan.Context.Stage != "" {
		fmt.Fprintf(&b, "Stage: %s\n", plan.Context.Stage)
	}
	if plan.Context.CommandHint != "" {
		fmt.Fprintf(&b, "Command: %s\n", plan.Context.CommandHint)
	}
	if plan.Context.Step != "" {
		fmt.Fprintf(&b, "Step: %s\n", plan.Context.Step)
	}
	if len(plan.Evidence) > 0 {
		fmt.Fprintln(&b, "Evidence:")
		for _, line := range plan.Evidence {
			fmt.Fprintf(&b, "  - %s\n", line)
		}
	}
	if len(plan.Files) > 0 {
		fmt.Fprintln(&b, "Focus files:")
		for _, file := range plan.Files {
			fmt.Fprintf(&b, "  - %s\n", file)
		}
	}
	if len(plan.LocalRepro) > 0 {
		fmt.Fprintln(&b, "Local repro:")
		for _, cmd := range plan.LocalRepro {
			fmt.Fprintf(&b, "  - %s\n", cmd)
		}
	}
	if len(plan.Verify) > 0 {
		fmt.Fprintln(&b, "Verify:")
		for _, cmd := range plan.Verify {
			fmt.Fprintf(&b, "  - %s\n", cmd)
		}
	}
	fmt.Fprintln(&b, "Next steps:")
	for i, step := range plan.Steps {
		fmt.Fprintf(&b, "  %d. %s\n", i+1, step)
	}
	if plan.AgentPrompt != "" {
		fmt.Fprintln(&b, "\nAgent prompt:")
		fmt.Fprintln(&b, plan.AgentPrompt)
	}
	return b.String()
}

// FormatWorkflowJSON serializes a workflow plan as stable JSON.
func FormatWorkflowJSON(plan workflow.Plan) (string, error) {
	payload := workflowJSON{
		SchemaVersion: plan.SchemaVersion,
		Mode:          string(plan.Mode),
		FailureID:     plan.FailureID,
		Title:         plan.Title,
		Source:        plan.Source,
		Context: ctxJSON{
			Stage:       plan.Context.Stage,
			CommandHint: plan.Context.CommandHint,
			Step:        plan.Context.Step,
		},
		Evidence:    plan.Evidence,
		Files:       plan.Files,
		LocalRepro:  plan.LocalRepro,
		Verify:      plan.Verify,
		Steps:       plan.Steps,
		AgentPrompt: plan.AgentPrompt,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal workflow JSON: %w", err)
	}
	return string(data) + "\n", nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func topN(results []model.Result, n int) []model.Result {
	if n <= 0 || n > len(results) {
		return results
	}
	return results[:n]
}

func firstMarkdownListItem(markdown string) string {
	for _, line := range strings.Split(markdown, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "- "):
			return strings.TrimSpace(strings.TrimPrefix(line, "- "))
		case len(line) > 3 && line[1] == '.' && line[2] == ' ' && line[0] >= '0' && line[0] <= '9':
			return strings.TrimSpace(line[3:])
		}
	}
	return ""
}

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
		if triggered := markdownListSection("## Triggered By", result.Explanation.TriggeredBy); triggered != "" {
			sections = append(sections, "", triggered)
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
	}
	return strings.TrimSpace(strings.Join(sections, "\n"))
}

func formatMatchSummaryMarkdown(pb model.Playbook) string {
	sections := []string{}
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

func scoreBreakdownLines(breakdown model.ScoreBreakdown) []string {
	if breakdown.FinalScore == 0 {
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
		lines = append(lines, fmt.Sprintf("change bonus: %.2f", breakdown.ChangeIntroducedBonus))
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
	lines := []string{}
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

func joinMetadataListItem(label, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return "- " + label + ": " + value
}

func filterEmpty(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}
