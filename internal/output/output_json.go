package output

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"faultline/internal/model"
)

// ── JSON types ────────────────────────────────────────────────────────────────

// analysisJSON is the stable JSON schema emitted by FormatAnalysisJSON.
type analysisJSON struct {
	Matched               bool                         `json:"matched"`
	Status                model.ArtifactStatus         `json:"status,omitempty"`
	Source                string                       `json:"source,omitempty"`
	Fingerprint           string                       `json:"fingerprint,omitempty"`
	InputHash             string                       `json:"input_hash,omitempty"`
	OutputHash            string                       `json:"output_hash,omitempty"`
	Context               ctxJSON                      `json:"context"`
	Results               []resultJSON                 `json:"results"`
	RepoContext           *repoCtxJSON                 `json:"repo_context,omitempty"`
	Delta                 *model.Delta                 `json:"delta,omitempty"`
	Differential          *model.DifferentialDiagnosis `json:"differential,omitempty"`
	PackProvenance        []model.PackProvenance       `json:"pack_provenance,omitempty"`
	Metrics               *model.Metrics               `json:"metrics,omitempty"`
	Policy                *model.Policy                `json:"policy,omitempty"`
	CandidateClusters     []model.CandidateCluster     `json:"candidate_clusters,omitempty"`
	DominantSignals       []string                     `json:"dominant_signals,omitempty"`
	SuggestedPlaybookSeed *model.SuggestedPlaybookSeed `json:"suggested_playbook_seed,omitempty"`
	Artifact              *model.FailureArtifact       `json:"artifact,omitempty"`
	Message               string                       `json:"message,omitempty"`
}

// ctxJSON is the stable representation of model.Context in JSON output.
// It is shared by analysisJSON and workflowJSON.
type ctxJSON struct {
	Stage       string `json:"stage,omitempty"`
	CommandHint string `json:"command_hint,omitempty"`
	Step        string `json:"step,omitempty"`
}

type resultJSON struct {
	Rank               int                         `json:"rank"`
	FailureID          string                      `json:"failure_id"`
	Title              string                      `json:"title"`
	Category           string                      `json:"category"`
	Pack               string                      `json:"pack,omitempty"`
	Severity           string                      `json:"severity,omitempty"`
	Detector           string                      `json:"detector,omitempty"`
	Score              float64                     `json:"score"`
	Confidence         float64                     `json:"confidence"`
	Summary            string                      `json:"summary,omitempty"`
	Diagnosis          string                      `json:"diagnosis,omitempty"`
	WhyItMatters       string                      `json:"why_it_matters,omitempty"`
	Fix                string                      `json:"fix,omitempty"`
	Validation         string                      `json:"validation,omitempty"`
	Evidence           []string                    `json:"evidence"`
	EvidenceBy         model.EvidenceBundle        `json:"evidence_by,omitempty"`
	Explanation        model.ResultExplanation     `json:"explanation,omitempty"`
	Breakdown          model.ScoreBreakdown        `json:"breakdown,omitempty"`
	ChangeStatus       string                      `json:"change_status,omitempty"`
	SeenCount          int                         `json:"seen_count"`
	SignatureHash      string                      `json:"signature_hash,omitempty"`
	SeenBefore         bool                        `json:"seen_before,omitempty"`
	OccurrenceCount    int                         `json:"occurrence_count,omitempty"`
	FirstSeenAt        string                      `json:"first_seen_at,omitempty"`
	LastSeenAt         string                      `json:"last_seen_at,omitempty"`
	HookHistorySummary *model.HookHistorySummary   `json:"hook_history_summary,omitempty"`
	Ranking            *model.Ranking              `json:"ranking,omitempty"`
	Hypothesis         *model.HypothesisAssessment `json:"hypothesis,omitempty"`
	Hooks              *model.HookReport           `json:"hooks,omitempty"`
}

type repoCtxJSON struct {
	RepoRoot           string           `json:"repo_root"`
	RecentFiles        []string         `json:"recent_files,omitempty"`
	RelatedCommits     []repoCommitJSON `json:"related_commits,omitempty"`
	HotspotDirectories []string         `json:"hotspot_directories,omitempty"`
	CoChangeHints      []string         `json:"co_change_hints,omitempty"`
	HotfixSignals      []string         `json:"hotfix_signals,omitempty"`
	DriftSignals       []string         `json:"drift_signals,omitempty"`
	ConfigDriftSignals []string         `json:"config_drift_signals,omitempty"`
	CIChangeSignals    []string         `json:"ci_change_signals,omitempty"`
	LargeCommitSignals []string         `json:"large_commit_signals,omitempty"`
}

type repoCommitJSON struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
}

// ── JSON formatters ───────────────────────────────────────────────────────────

// FormatAnalysisJSON serialises an analysis to the stable JSON schema.
func FormatAnalysisJSON(a *model.Analysis, top int) (string, error) {
	payload := analysisPayload(a, top)

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal analysis JSON: %w", err)
	}
	return string(data) + "\n", nil
}

// FormatPlaybookDetailsJSON serialises a single playbook using the stable
// model JSON tags.
func FormatPlaybookDetailsJSON(pb model.Playbook) (string, error) {
	data, err := json.MarshalIndent(pb, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal playbook JSON: %w", err)
	}
	return string(data) + "\n", nil
}

// ParseAnalysisJSON deserializes the stable analysis JSON schema back into an
// in-memory Analysis so saved artifacts can be deterministically replayed.
func ParseAnalysisJSON(data []byte) (*model.Analysis, error) {
	var payload analysisJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse analysis JSON: %w", err)
	}

	a := &model.Analysis{
		Results:               make([]model.Result, 0, len(payload.Results)),
		Source:                payload.Source,
		Fingerprint:           payload.Fingerprint,
		InputHash:             payload.InputHash,
		OutputHash:            payload.OutputHash,
		Context:               model.Context(payload.Context),
		Delta:                 payload.Delta,
		Differential:          payload.Differential,
		PackProvenances:       payload.PackProvenance,
		Metrics:               payload.Metrics,
		Policy:                payload.Policy,
		Status:                payload.Status,
		CandidateClusters:     payload.CandidateClusters,
		DominantSignals:       payload.DominantSignals,
		SuggestedPlaybookSeed: payload.SuggestedPlaybookSeed,
		Artifact:              payload.Artifact,
	}
	a.RepoContext = parseRepoContextJSON(payload.RepoContext)

	for _, item := range payload.Results {
		a.Results = append(a.Results, model.Result{
			Playbook: model.Playbook{
				ID:           item.FailureID,
				Title:        item.Title,
				Category:     item.Category,
				Severity:     item.Severity,
				Summary:      item.Summary,
				Diagnosis:    item.Diagnosis,
				WhyItMatters: item.WhyItMatters,
				Fix:          item.Fix,
				Validation:   item.Validation,
				Metadata: model.PlaybookMeta{
					PackName: item.Pack,
				},
			},
			Detector:           item.Detector,
			Score:              item.Score,
			Confidence:         item.Confidence,
			Evidence:           item.Evidence,
			EvidenceBy:         item.EvidenceBy,
			Explanation:        item.Explanation,
			Breakdown:          item.Breakdown,
			ChangeStatus:       item.ChangeStatus,
			SeenCount:          item.SeenCount,
			SignatureHash:      item.SignatureHash,
			SeenBefore:         item.SeenBefore,
			OccurrenceCount:    item.OccurrenceCount,
			FirstSeenAt:        item.FirstSeenAt,
			LastSeenAt:         item.LastSeenAt,
			HookHistorySummary: item.HookHistorySummary,
			Ranking:            item.Ranking,
			Hypothesis:         item.Hypothesis,
			Hooks:              item.Hooks,
		})
	}

	if !payload.Matched && len(a.Results) == 0 {
		a.Results = []model.Result{}
	}

	return a, nil
}

func HashAnalysisOutput(a *model.Analysis) (string, error) {
	payload := analysisPayload(stableHashAnalysis(a), 0)
	payload.OutputHash = ""
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal analysis JSON for hash: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func stableHashAnalysis(a *model.Analysis) *model.Analysis {
	if a == nil {
		return nil
	}
	clone := *a
	clone.Results = append([]model.Result(nil), a.Results...)
	clone.CandidateClusters = append([]model.CandidateCluster(nil), a.CandidateClusters...)
	clone.DominantSignals = append([]string(nil), a.DominantSignals...)
	if a.SuggestedPlaybookSeed != nil {
		seed := *a.SuggestedPlaybookSeed
		clone.SuggestedPlaybookSeed = &seed
	}
	if a.Artifact != nil {
		artifact := *a.Artifact
		clone.Artifact = &artifact
		if a.Artifact.MatchedPlaybook != nil {
			matched := *a.Artifact.MatchedPlaybook
			clone.Artifact.MatchedPlaybook = &matched
		}
		if a.Artifact.HistoryContext != nil {
			history := *a.Artifact.HistoryContext
			clone.Artifact.HistoryContext = &history
		}
		if a.Artifact.SuggestedPlaybookSeed != nil {
			seed := *a.Artifact.SuggestedPlaybookSeed
			clone.Artifact.SuggestedPlaybookSeed = &seed
		}
		if a.Artifact.Remediation != nil {
			remediation := *a.Artifact.Remediation
			clone.Artifact.Remediation = &remediation
		}
	}
	clone.Metrics = nil
	clone.Policy = nil
	for i := range clone.Results {
		result := clone.Results[i]
		result.SeenCount = 0
		result.SeenBefore = false
		result.OccurrenceCount = 0
		result.FirstSeenAt = ""
		result.LastSeenAt = ""
		result.HookHistorySummary = nil
		clone.Results[i] = result
	}
	if clone.Artifact != nil && clone.Artifact.HistoryContext != nil {
		clone.Artifact.HistoryContext.SeenCount = 0
		clone.Artifact.HistoryContext.SeenBefore = false
		clone.Artifact.HistoryContext.OccurrenceCount = 0
		clone.Artifact.HistoryContext.FirstSeenAt = ""
		clone.Artifact.HistoryContext.LastSeenAt = ""
		clone.Artifact.HistoryContext.HookHistorySummary = nil
	}
	return &clone
}

func analysisPayload(a *model.Analysis, top int) analysisJSON {
	payload := analysisJSON{
		Matched: a != nil && len(a.Results) > 0,
	}

	if a == nil {
		payload.Message = "No known playbook matched this input."
		payload.Results = []resultJSON{}
		return payload
	}

	payload.Source = a.Source
	payload.Status = a.Status
	payload.Fingerprint = a.Fingerprint
	payload.InputHash = a.InputHash
	payload.OutputHash = a.OutputHash
	payload.Context = ctxJSON{
		Stage:       a.Context.Stage,
		CommandHint: a.Context.CommandHint,
		Step:        a.Context.Step,
	}
	payload.RepoContext = repoContextJSON(a.RepoContext)
	payload.Delta = a.Delta
	payload.Differential = a.Differential
	payload.PackProvenance = a.PackProvenances
	payload.Metrics = a.Metrics
	payload.Policy = a.Policy
	payload.CandidateClusters = a.CandidateClusters
	payload.DominantSignals = a.DominantSignals
	payload.SuggestedPlaybookSeed = a.SuggestedPlaybookSeed
	payload.Artifact = a.Artifact

	if !payload.Matched {
		payload.Message = "No known playbook matched this input."
		payload.Results = []resultJSON{}
		return payload
	}

	results := topN(a.Results, top)
	payload.Results = make([]resultJSON, len(results))
	for i, r := range results {
		payload.Results[i] = resultJSON{
			Rank:               i + 1,
			FailureID:          r.Playbook.ID,
			Title:              r.Playbook.Title,
			Category:           r.Playbook.Category,
			Pack:               displayPackName(r.Playbook),
			Severity:           r.Playbook.Severity,
			Detector:           r.Detector,
			Score:              r.Score,
			Confidence:         r.Confidence,
			Summary:            r.Playbook.Summary,
			Diagnosis:          r.Playbook.Diagnosis,
			WhyItMatters:       r.Playbook.WhyItMatters,
			Fix:                r.Playbook.Fix,
			Validation:         r.Playbook.Validation,
			Evidence:           r.Evidence,
			EvidenceBy:         r.EvidenceBy,
			Explanation:        r.Explanation,
			Breakdown:          r.Breakdown,
			ChangeStatus:       r.ChangeStatus,
			SeenCount:          r.SeenCount,
			SignatureHash:      r.SignatureHash,
			SeenBefore:         r.SeenBefore,
			OccurrenceCount:    r.OccurrenceCount,
			FirstSeenAt:        r.FirstSeenAt,
			LastSeenAt:         r.LastSeenAt,
			HookHistorySummary: r.HookHistorySummary,
			Ranking:            r.Ranking,
			Hypothesis:         r.Hypothesis,
			Hooks:              r.Hooks,
		}
	}
	return payload
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
		ConfigDriftSignals: repoCtx.ConfigDriftSignals,
		CIChangeSignals:    repoCtx.CIChangeSignals,
		LargeCommitSignals: repoCtx.LargeCommitSignals,
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

func parseRepoContextJSON(repoCtx *repoCtxJSON) *model.RepoContext {
	if repoCtx == nil {
		return nil
	}
	out := &model.RepoContext{
		RepoRoot:           repoCtx.RepoRoot,
		RecentFiles:        repoCtx.RecentFiles,
		HotspotDirectories: repoCtx.HotspotDirectories,
		CoChangeHints:      repoCtx.CoChangeHints,
		HotfixSignals:      repoCtx.HotfixSignals,
		DriftSignals:       repoCtx.DriftSignals,
		ConfigDriftSignals: repoCtx.ConfigDriftSignals,
		CIChangeSignals:    repoCtx.CIChangeSignals,
		LargeCommitSignals: repoCtx.LargeCommitSignals,
	}
	if len(repoCtx.RelatedCommits) > 0 {
		out.RelatedCommits = make([]model.RepoCommit, len(repoCtx.RelatedCommits))
		for i, commit := range repoCtx.RelatedCommits {
			out.RelatedCommits[i] = model.RepoCommit{
				Hash:    commit.Hash,
				Subject: commit.Subject,
				Date:    commit.Date,
			}
		}
	}
	return out
}
