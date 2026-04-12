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

// ── Human-readable text ──────────────────────────────────────────────────────

// FormatAnalysisText formats an analysis for human consumption.
// top limits the number of results shown (0 or negative means show all).
func FormatAnalysisText(a *model.Analysis, top int, mode Mode, opts renderer.Options) string {
	return renderer.New(opts).RenderAnalyze(a, top, mode == ModeDetailed)
}

// FormatFix formats only the fix steps for the top result.
func FormatFix(a *model.Analysis, opts renderer.Options) string {
	return renderer.New(opts).RenderFix(a)
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
