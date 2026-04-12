// Package output formats analysis results for humans, automation, and CI
// annotation consumers.  All functions accept a *model.Analysis that may be
// nil (when no log was provided) or have an empty Results slice (when no
// playbook matched).
package output

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"text/tabwriter"

	"faultline/internal/model"
	"faultline/internal/workflow"
)

// Mode selects the verbosity of human-readable output.
type Mode string

const (
	ModeQuick    Mode = "quick"
	ModeDetailed Mode = "detailed"
)

// ── Human-readable text ──────────────────────────────────────────────────────

// FormatAnalysisText formats an analysis for human consumption.
// top limits the number of results shown (0 or negative means show all).
func FormatAnalysisText(a *model.Analysis, top int, mode Mode) string {
	if a == nil || len(a.Results) == 0 {
		return noMatchText()
	}
	results := topN(a.Results, top)

	var b strings.Builder
	for i, r := range results {
		if i > 0 {
			b.WriteString("\n")
		}
		if mode == ModeDetailed {
			writeDetailed(&b, a, r, i, len(results))
		} else {
			writeQuick(&b, r, i, len(results))
		}
	}
	return b.String()
}

// FormatFix formats only the fix steps for the top result.
func FormatFix(a *model.Analysis) string {
	if a == nil || len(a.Results) == 0 {
		return noMatchText()
	}
	r := a.Results[0]
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %s\n\n", r.Playbook.ID, r.Playbook.Title)
	if len(r.Playbook.Fix) == 0 {
		fmt.Fprintln(&b, "No fix steps defined for this playbook.")
		return b.String()
	}
	fmt.Fprintln(&b, "Fix:")
	for i, step := range r.Playbook.Fix {
		fmt.Fprintf(&b, "  %d. %s\n", i+1, step)
	}
	return b.String()
}

func writeQuick(b *strings.Builder, r model.Result, rank, total int) {
	pb := r.Playbook

	// Header line
	severity := pb.Severity
	if severity == "" {
		severity = "unknown"
	}
	if total > 1 {
		fmt.Fprintf(b, "#%d  %s · %s  [%s · %s · %d%%]\n",
			rank+1, pb.ID, pb.Title, pb.Category, severity,
			int(math.Round(r.Confidence*100)))
	} else {
		fmt.Fprintf(b, "DETECTED  %s · %s  [%s · %s · %d%%]\n",
			pb.ID, pb.Title, pb.Category, severity,
			int(math.Round(r.Confidence*100)))
	}
	if r.Detector != "" && r.Detector != "log" {
		fmt.Fprintf(b, "Detector: %s", r.Detector)
		if r.ChangeStatus != "" {
			fmt.Fprintf(b, "  Change: %s", r.ChangeStatus)
		}
		b.WriteString("\n")
	}

	// Fix steps
	if len(pb.Fix) > 0 {
		fmt.Fprintln(b, "FIX")
		for i, step := range pb.Fix {
			fmt.Fprintf(b, "  %d. %s\n", i+1, step)
		}
	}
}

const ruler = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

func writeDetailed(b *strings.Builder, a *model.Analysis, r model.Result, rank, total int) {
	pb := r.Playbook
	fmt.Fprintln(b, ruler)

	if total > 1 {
		fmt.Fprintf(b, "  Diagnosis #%d of %d — %s\n", rank+1, total, pb.Title)
	} else {
		fmt.Fprintf(b, "  Diagnosis: %s\n", pb.Title)
	}
	fmt.Fprintln(b, ruler)

	// Metadata grid
	if a.Source != "" {
		fmt.Fprintf(b, "  Source:      %s\n", a.Source)
	}
	if a.Context.Stage != "" {
		fmt.Fprintf(b, "  Stage:       %s\n", a.Context.Stage)
	}
	if a.Context.CommandHint != "" {
		fmt.Fprintf(b, "  Command:     %s\n", a.Context.CommandHint)
	}
	if a.Context.Step != "" {
		fmt.Fprintf(b, "  Step:        %s\n", a.Context.Step)
	}
	fmt.Fprintf(b, "  Category:    %s\n", pb.Category)
	if r.Detector != "" {
		fmt.Fprintf(b, "  Detector:    %s\n", r.Detector)
	}
	if pb.Severity != "" {
		fmt.Fprintf(b, "  Severity:    %s\n", pb.Severity)
	}
	fmt.Fprintf(b, "  Score:       %.2f\n", r.Score)
	fmt.Fprintf(b, "  Confidence:  %d%%\n", int(math.Round(r.Confidence*100)))
	if r.ChangeStatus != "" {
		fmt.Fprintf(b, "  Change:      %s\n", r.ChangeStatus)
	}
	if r.SeenCount > 0 {
		fmt.Fprintf(b, "  Seen before: %d time", r.SeenCount)
		if r.SeenCount != 1 {
			b.WriteString("s")
		}
		b.WriteString("\n")
	}

	// Narrative sections
	writeSection(b, "Cause", []string{pb.Explain})
	writeSection(b, "Why this happens", []string{pb.Why})

	if len(pb.Fix) > 0 {
		fmt.Fprintln(b, "\n  Fix")
		fmt.Fprintln(b, "  "+strings.Repeat("─", 40))
		for i, step := range pb.Fix {
			fmt.Fprintf(b, "  %d. %s\n", i+1, step)
		}
	}
	if len(pb.Prevent) > 0 {
		fmt.Fprintln(b, "\n  Prevent")
		fmt.Fprintln(b, "  "+strings.Repeat("─", 40))
		for i, step := range pb.Prevent {
			fmt.Fprintf(b, "  %d. %s\n", i+1, step)
		}
	}
	if len(r.Evidence) > 0 {
		fmt.Fprintln(b, "\n  Evidence")
		fmt.Fprintln(b, "  "+strings.Repeat("─", 40))
		for _, e := range r.Evidence {
			fmt.Fprintf(b, "  › %s\n", e)
		}
	}
	writeReasonSection(b, "Triggered by", r.Explanation.TriggeredBy)
	writeReasonSection(b, "Amplified by", r.Explanation.AmplifiedBy)
	writeReasonSection(b, "Mitigated by", r.Explanation.MitigatedBy)
	writeReasonSection(b, "Suppressions", r.Explanation.SuppressedBy)
	writeReasonSection(b, "Context", r.Explanation.Contextualized)
	writeScoreBreakdown(b, r.Breakdown)

	// Alternatives: the remaining results listed briefly.
	if rank == 0 && len(a.Results) > 1 {
		fmt.Fprintln(b, "\n  Alternatives")
		fmt.Fprintln(b, "  "+strings.Repeat("─", 40))
		for i, alt := range a.Results[1:] {
			fmt.Fprintf(b, "  %d. %s: %s (%d%%)\n",
				i+2, alt.Playbook.ID, alt.Playbook.Title,
				int(math.Round(alt.Confidence*100)))
		}
	}

	writeRepoContext(b, a.RepoContext)

	fmt.Fprintln(b, ruler)
}

func writeSection(b *strings.Builder, header string, lines []string) {
	// Skip sections whose content is completely empty.
	nonEmpty := false
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty = true
			break
		}
	}
	if !nonEmpty {
		return
	}
	fmt.Fprintf(b, "\n  %s\n", header)
	fmt.Fprintln(b, "  "+strings.Repeat("─", 40))
	for _, l := range lines {
		if l != "" {
			fmt.Fprintf(b, "  %s\n", l)
		}
	}
}

func writeReasonSection(b *strings.Builder, header string, lines []string) {
	if len(lines) == 0 {
		return
	}
	fmt.Fprintf(b, "\n  %s\n", header)
	fmt.Fprintln(b, "  "+strings.Repeat("─", 40))
	for _, line := range lines {
		fmt.Fprintf(b, "  - %s\n", line)
	}
}

func writeScoreBreakdown(b *strings.Builder, breakdown model.ScoreBreakdown) {
	if breakdown.FinalScore == 0 {
		return
	}
	fmt.Fprintln(b, "\n  Score Breakdown")
	fmt.Fprintln(b, "  "+strings.Repeat("─", 40))
	fmt.Fprintf(b, "  base: %.2f\n", breakdown.BaseSignalScore)
	if breakdown.CompoundSignalBonus != 0 {
		fmt.Fprintf(b, "  compound: +%.2f\n", breakdown.CompoundSignalBonus)
	}
	if breakdown.BlastRadiusMultiplier != 0 {
		fmt.Fprintf(b, "  blast radius: +%.2f\n", breakdown.BlastRadiusMultiplier)
	}
	if breakdown.HotPathMultiplier != 0 {
		fmt.Fprintf(b, "  hot path: +%.2f\n", breakdown.HotPathMultiplier)
	}
	if breakdown.ChangeIntroducedBonus != 0 {
		fmt.Fprintf(b, "  change bonus: %.2f\n", breakdown.ChangeIntroducedBonus)
	}
	if breakdown.MitigatingEvidenceDiscount != 0 {
		fmt.Fprintf(b, "  mitigations: -%.2f\n", breakdown.MitigatingEvidenceDiscount)
	}
	if breakdown.ExplicitExceptionDiscount != 0 {
		fmt.Fprintf(b, "  suppressions: -%.2f\n", breakdown.ExplicitExceptionDiscount)
	}
	if breakdown.SafeContextDiscount != 0 {
		fmt.Fprintf(b, "  safe context: -%.2f\n", breakdown.SafeContextDiscount)
	}
	fmt.Fprintf(b, "  final: %.2f\n", breakdown.FinalScore)
}

func fallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func noMatchText() string {
	return "No known failure pattern matched.\n" +
		"  » Run \"faultline list\" to see all available playbooks.\n" +
		"  » Pass --json for machine-readable output.\n"
}

func writeRepoContext(b *strings.Builder, repoCtx *model.RepoContext) {
	if repoCtx == nil {
		return
	}

	if repoCtx.RepoRoot == "" &&
		len(repoCtx.RecentFiles) == 0 &&
		len(repoCtx.RelatedCommits) == 0 &&
		len(repoCtx.HotspotDirectories) == 0 &&
		len(repoCtx.CoChangeHints) == 0 &&
		len(repoCtx.HotfixSignals) == 0 &&
		len(repoCtx.DriftSignals) == 0 {
		return
	}

	fmt.Fprintln(b, "\n  Repo Context")
	fmt.Fprintln(b, "  "+strings.Repeat("─", 40))
	if repoCtx.RepoRoot != "" {
		fmt.Fprintf(b, "  Repo root: %s\n", repoCtx.RepoRoot)
	}
	for _, item := range repoCtx.RecentFiles {
		fmt.Fprintf(b, "  Recent file: %s\n", item)
	}
	for _, commit := range repoCtx.RelatedCommits {
		fmt.Fprintf(b, "  Related commit: %s  %s  %s\n", commit.Date, commit.Hash, commit.Subject)
	}
	for _, dir := range repoCtx.HotspotDirectories {
		fmt.Fprintf(b, "  Hotspot area: %s\n", dir)
	}
	for _, hint := range repoCtx.CoChangeHints {
		fmt.Fprintf(b, "  Co-change: %s\n", hint)
	}
	for _, signal := range repoCtx.HotfixSignals {
		fmt.Fprintf(b, "  Hotfix signal: %s\n", signal)
	}
	for _, signal := range repoCtx.DriftSignals {
		fmt.Fprintf(b, "  Drift hint: %s\n", signal)
	}
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
	Rank         int                     `json:"rank"`
	FailureID    string                  `json:"failure_id"`
	Title        string                  `json:"title"`
	Category     string                  `json:"category"`
	Severity     string                  `json:"severity,omitempty"`
	Detector     string                  `json:"detector,omitempty"`
	Score        float64                 `json:"score"`
	Confidence   float64                 `json:"confidence"`
	Explain      string                  `json:"explain,omitempty"`
	Why          string                  `json:"why,omitempty"`
	Fix          []string                `json:"fix,omitempty"`
	Prevent      []string                `json:"prevent,omitempty"`
	Evidence     []string                `json:"evidence"`
	EvidenceBy   model.EvidenceBundle    `json:"evidence_by,omitempty"`
	Explanation  model.ResultExplanation `json:"explanation,omitempty"`
	Breakdown    model.ScoreBreakdown    `json:"breakdown,omitempty"`
	ChangeStatus string                  `json:"change_status,omitempty"`
	SeenCount    int                     `json:"seen_count"`
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
				Rank:         i + 1,
				FailureID:    r.Playbook.ID,
				Title:        r.Playbook.Title,
				Category:     r.Playbook.Category,
				Severity:     r.Playbook.Severity,
				Detector:     r.Detector,
				Score:        r.Score,
				Confidence:   r.Confidence,
				Explain:      r.Playbook.Explain,
				Why:          r.Playbook.Why,
				Fix:          r.Playbook.Fix,
				Prevent:      r.Playbook.Prevent,
				Evidence:     r.Evidence,
				EvidenceBy:   r.EvidenceBy,
				Explanation:  r.Explanation,
				Breakdown:    r.Breakdown,
				ChangeStatus: r.ChangeStatus,
				SeenCount:    r.SeenCount,
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
		if len(r.Playbook.Fix) > 0 {
			fix = " Fix: " + r.Playbook.Fix[0]
		}
		fmt.Fprintf(&b, "::warning title=%s::%s.%s\n",
			r.Playbook.ID, r.Playbook.Title, fix)
	}
	return b.String()
}

// ── Playbook list & details ──────────────────────────────────────────────────

// FormatPlaybookList formats a tab-aligned table of available playbooks.
// When category is non-empty only matching playbooks are shown.
func FormatPlaybookList(playbooks []model.Playbook, category string) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCATEGORY\tSEVERITY\tTITLE")
	filter := strings.ToLower(strings.TrimSpace(category))
	for _, pb := range playbooks {
		if filter != "" && strings.ToLower(pb.Category) != filter {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			pb.ID, pb.Category, pb.Severity, pb.Title)
	}
	_ = w.Flush()
	return b.String()
}

// FormatPlaybookDetails formats all fields of a single playbook for the
// explain command.
func FormatPlaybookDetails(pb model.Playbook) string {
	var b strings.Builder
	const sep = "═══════════════════════════════════════════════════════"

	fmt.Fprintln(&b, sep)
	fmt.Fprintf(&b, "  %s — %s\n", pb.ID, pb.Title)
	fmt.Fprintln(&b, sep)
	fmt.Fprintf(&b, "  Category:  %s\n", pb.Category)
	if pb.Severity != "" {
		fmt.Fprintf(&b, "  Severity:  %s\n", pb.Severity)
	}
	if len(pb.Tags) > 0 {
		fmt.Fprintf(&b, "  Tags:      %s\n", strings.Join(pb.Tags, ", "))
	}
	if len(pb.StageHints) > 0 {
		fmt.Fprintf(&b, "  Stages:    %s\n", strings.Join(pb.StageHints, ", "))
	}

	writeSection(&b, "Explanation", []string{pb.Explain})
	writeSection(&b, "Why this happens", []string{pb.Why})

	if len(pb.Fix) > 0 {
		fmt.Fprintln(&b, "\n  Fix")
		fmt.Fprintln(&b, "  "+strings.Repeat("─", 40))
		for i, step := range pb.Fix {
			fmt.Fprintf(&b, "  %d. %s\n", i+1, step)
		}
	}
	if len(pb.Prevent) > 0 {
		fmt.Fprintln(&b, "\n  Prevent")
		fmt.Fprintln(&b, "  "+strings.Repeat("─", 40))
		for i, step := range pb.Prevent {
			fmt.Fprintf(&b, "  %d. %s\n", i+1, step)
		}
	}

	// Match patterns summary
	fmt.Fprintln(&b, "\n  Match Patterns")
	fmt.Fprintln(&b, "  "+strings.Repeat("─", 40))
	if len(pb.Match.Any) > 0 {
		fmt.Fprintf(&b, "  any: %s\n", strings.Join(pb.Match.Any, " | "))
	}
	if len(pb.Match.All) > 0 {
		fmt.Fprintf(&b, "  all: %s\n", strings.Join(pb.Match.All, " + "))
	}

	fmt.Fprintln(&b, sep)
	return b.String()
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

// topN returns the first n results. When n <= 0 or n > len(results), all
// results are returned.
func topN(results []model.Result, n int) []model.Result {
	if n <= 0 || n > len(results) {
		return results
	}
	return results[:n]
}
