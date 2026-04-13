package output

import (
	"encoding/json"
	"fmt"

	"faultline/internal/model"
)

// ── JSON types ────────────────────────────────────────────────────────────────

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

// ctxJSON is the stable representation of model.Context in JSON output.
// It is shared by analysisJSON and workflowJSON.
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
	Pack         string                  `json:"pack,omitempty"`
	Severity     string                  `json:"severity,omitempty"`
	Detector     string                  `json:"detector,omitempty"`
	Score        float64                 `json:"score"`
	Confidence   float64                 `json:"confidence"`
	Summary      string                  `json:"summary,omitempty"`
	Diagnosis    string                  `json:"diagnosis,omitempty"`
	WhyItMatters string                  `json:"why_it_matters,omitempty"`
	Fix          string                  `json:"fix,omitempty"`
	Validation   string                  `json:"validation,omitempty"`
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

// ── JSON formatters ───────────────────────────────────────────────────────────

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
				Pack:         displayPackName(r.Playbook),
				Severity:     r.Playbook.Severity,
				Detector:     r.Detector,
				Score:        r.Score,
				Confidence:   r.Confidence,
				Summary:      r.Playbook.Summary,
				Diagnosis:    r.Playbook.Diagnosis,
				WhyItMatters: r.Playbook.WhyItMatters,
				Fix:          r.Playbook.Fix,
				Validation:   r.Playbook.Validation,
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
