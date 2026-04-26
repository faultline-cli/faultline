package model

// Line is a single processed log line with its original and normalised forms.
type Line struct {
	Original   string
	Normalized string
	Number     int
}

// Context holds lightweight inferences about the log extracted by heuristics.
type Context struct {
	Stage       string `json:"stage,omitempty"`
	CommandHint string `json:"command_hint,omitempty"`
	Step        string `json:"step,omitempty"`
}

type EvidenceKind string

const (
	EvidenceTrigger     EvidenceKind = "trigger"
	EvidenceAmplifier   EvidenceKind = "amplifier"
	EvidenceMitigation  EvidenceKind = "mitigation"
	EvidenceSuppression EvidenceKind = "suppression"
	EvidenceContext     EvidenceKind = "context"
)

type Evidence struct {
	Kind       EvidenceKind `json:"kind"`
	SignalID   string       `json:"signal_id,omitempty"`
	Label      string       `json:"label,omitempty"`
	Detail     string       `json:"detail,omitempty"`
	File       string       `json:"file,omitempty"`
	Line       int          `json:"line,omitempty"`
	PathClass  string       `json:"path_class,omitempty"`
	Scope      string       `json:"scope,omitempty"`
	ScopeName  string       `json:"scope_name,omitempty"`
	Proximity  string       `json:"proximity,omitempty"`
	Distance   int          `json:"distance,omitempty"`
	Weight     float64      `json:"weight,omitempty"`
	Suppressed bool         `json:"suppressed,omitempty"`
	ExpiresOn  string       `json:"expires_on,omitempty"`
	Reason     string       `json:"reason,omitempty"`
	Source     string       `json:"source,omitempty"`
}

type EvidenceBundle struct {
	Triggers     []Evidence `json:"triggers,omitempty"`
	Amplifiers   []Evidence `json:"amplifiers,omitempty"`
	Mitigations  []Evidence `json:"mitigations,omitempty"`
	Suppressions []Evidence `json:"suppressions,omitempty"`
	Context      []Evidence `json:"context,omitempty"`
}

type ScoreBreakdown struct {
	BaseSignalScore            float64 `json:"base_signal_score"`
	CompoundSignalBonus        float64 `json:"compound_signal_bonus"`
	BlastRadiusMultiplier      float64 `json:"blast_radius_multiplier"`
	HotPathMultiplier          float64 `json:"hot_path_multiplier"`
	ChangeIntroducedBonus      float64 `json:"change_introduced_bonus"`
	MitigatingEvidenceDiscount float64 `json:"mitigating_evidence_discount"`
	ExplicitExceptionDiscount  float64 `json:"explicit_exception_discount"`
	SafeContextDiscount        float64 `json:"safe_context_discount"`
	FinalScore                 float64 `json:"final_score"`
}

type ResultExplanation struct {
	TriggeredBy    []string `json:"triggered_by,omitempty"`
	AmplifiedBy    []string `json:"amplified_by,omitempty"`
	MitigatedBy    []string `json:"mitigated_by,omitempty"`
	SuppressedBy   []string `json:"suppressed_by,omitempty"`
	Contextualized []string `json:"contextualized_by,omitempty"`
	ChangeStatus   string   `json:"change_status,omitempty"`
}

type RankingContribution struct {
	Feature      string   `json:"feature"`
	Value        float64  `json:"value"`
	Weight       float64  `json:"weight"`
	Contribution float64  `json:"contribution"`
	Direction    string   `json:"direction,omitempty"`
	Reason       string   `json:"reason,omitempty"`
	EvidenceRefs []string `json:"evidence_refs,omitempty"`
}

type Ranking struct {
	Mode              string                `json:"mode,omitempty"`
	Version           string                `json:"version,omitempty"`
	BaselineScore     float64               `json:"baseline_score,omitempty"`
	Prior             float64               `json:"prior,omitempty"`
	FinalScore        float64               `json:"final_score,omitempty"`
	Contributions     []RankingContribution `json:"contributions,omitempty"`
	StrongestPositive []string              `json:"strongest_positive,omitempty"`
	StrongestNegative []string              `json:"strongest_negative,omitempty"`
}

type DeltaCause struct {
	Kind    string   `json:"kind"`
	Score   float64  `json:"score"`
	Reasons []string `json:"reasons,omitempty"`
}

type DeltaEnvChange struct {
	Baseline string `json:"baseline,omitempty"`
	Current  string `json:"current,omitempty"`
}

type DeltaSignal struct {
	ID     string  `json:"id"`
	Detail string  `json:"detail,omitempty"`
	Weight float64 `json:"weight,omitempty"`
}

type Delta struct {
	Version           string                    `json:"version,omitempty"`
	Provider          string                    `json:"provider,omitempty"`
	FilesChanged      []string                  `json:"files_changed,omitempty"`
	TestsNewlyFailing []string                  `json:"tests_newly_failing,omitempty"`
	ErrorsAdded       []string                  `json:"errors_added,omitempty"`
	EnvDiff           map[string]DeltaEnvChange `json:"env_diff,omitempty"`
	Signals           []DeltaSignal             `json:"signals,omitempty"`
	Causes            []DeltaCause              `json:"causes,omitempty"`
}

type DifferentialCandidate struct {
	FailureID       string   `json:"failure_id,omitempty"`
	Title           string   `json:"title,omitempty"`
	Category        string   `json:"category,omitempty"`
	Confidence      float64  `json:"confidence,omitempty"`
	ConfidenceText  string   `json:"confidence_text,omitempty"`
	HypothesisScore float64  `json:"hypothesis_score,omitempty"`
	Why             []string `json:"why,omitempty"`
	WhyLessLikely   []string `json:"why_less_likely,omitempty"`
	RuledOutBy      []string `json:"ruled_out_by,omitempty"`
	DisproofChecks  []string `json:"disproof_checks,omitempty"`
}

type DifferentialDiagnosis struct {
	Version      string                  `json:"version,omitempty"`
	Likely       *DifferentialCandidate  `json:"likely,omitempty"`
	Alternatives []DifferentialCandidate `json:"alternatives,omitempty"`
	RuledOut     []DifferentialCandidate `json:"ruled_out,omitempty"`
}

// Result is a single ranked playbook match with its scoring detail.
type Result struct {
	Playbook           Playbook              `json:"playbook"`
	Detector           string                `json:"detector,omitempty"`
	Score              float64               `json:"score"`
	Confidence         float64               `json:"confidence"`
	Evidence           []string              `json:"evidence"`
	EvidenceBy         EvidenceBundle        `json:"evidence_by,omitempty"`
	Explanation        ResultExplanation     `json:"explanation,omitempty"`
	Breakdown          ScoreBreakdown        `json:"breakdown,omitempty"`
	ChangeStatus       string                `json:"change_status,omitempty"`
	SeenCount          int                   `json:"seen_count"`
	SignatureHash      string                `json:"signature_hash,omitempty"`
	SeenBefore         bool                  `json:"seen_before,omitempty"`
	OccurrenceCount    int                   `json:"occurrence_count,omitempty"`
	FirstSeenAt        string                `json:"first_seen_at,omitempty"`
	LastSeenAt         string                `json:"last_seen_at,omitempty"`
	HookHistorySummary *HookHistorySummary   `json:"hook_history_summary,omitempty"`
	Ranking            *Ranking              `json:"ranking,omitempty"`
	Hypothesis         *HypothesisAssessment `json:"hypothesis,omitempty"`
	Hooks              *HookReport           `json:"hooks,omitempty"`
}

type ArtifactStatus string

const (
	ArtifactStatusMatched ArtifactStatus = "matched"
	ArtifactStatusUnknown ArtifactStatus = "unknown"
)

type ArtifactPlaybook struct {
	ID       string `json:"id,omitempty"`
	Title    string `json:"title,omitempty"`
	Category string `json:"category,omitempty"`
	Severity string `json:"severity,omitempty"`
	Detector string `json:"detector,omitempty"`
	Pack     string `json:"pack,omitempty"`
}

type ArtifactEnvironment struct {
	Source         string           `json:"source,omitempty"`
	Context        Context          `json:"context"`
	RepoRoot       string           `json:"repo_root,omitempty"`
	DeltaProvider  string           `json:"delta_provider,omitempty"`
	PackProvenance []PackProvenance `json:"pack_provenance,omitempty"`
	RecentFiles    []string         `json:"recent_files,omitempty"`
	RelatedCommits []RepoCommit     `json:"related_commits,omitempty"`
}

type ArtifactHistoryContext struct {
	SeenCount          int                 `json:"seen_count,omitempty"`
	SignatureHash      string              `json:"signature_hash,omitempty"`
	SeenBefore         bool                `json:"seen_before,omitempty"`
	OccurrenceCount    int                 `json:"occurrence_count,omitempty"`
	FirstSeenAt        string              `json:"first_seen_at,omitempty"`
	LastSeenAt         string              `json:"last_seen_at,omitempty"`
	HookHistorySummary *HookHistorySummary `json:"hook_history_summary,omitempty"`
}

type CandidateCluster struct {
	Key            string   `json:"key,omitempty"`
	Summary        string   `json:"summary,omitempty"`
	LikelyCategory string   `json:"likely_category,omitempty"`
	Confidence     float64  `json:"confidence,omitempty"`
	Signals        []string `json:"signals,omitempty"`
	Evidence       []string `json:"evidence,omitempty"`
}

type SuggestedPlaybookSeed struct {
	Category  string       `json:"category,omitempty"`
	Title     string       `json:"title,omitempty"`
	MatchAny  []string     `json:"match_any,omitempty"`
	MatchNone []string     `json:"match_none,omitempty"`
	Workflow  WorkflowSpec `json:"workflow,omitempty"`
}

type RemediationCommand struct {
	ID        string   `json:"id,omitempty"`
	Phase     string   `json:"phase,omitempty"`
	Command   []string `json:"command,omitempty"`
	WorkDir   string   `json:"workdir,omitempty"`
	Rationale string   `json:"rationale,omitempty"`
}

type PatchSuggestion struct {
	TargetFile string   `json:"target_file,omitempty"`
	Summary    string   `json:"summary,omitempty"`
	Actions    []string `json:"actions,omitempty"`
}

type CIConfigDiff struct {
	TargetFile string   `json:"target_file,omitempty"`
	Summary    string   `json:"summary,omitempty"`
	Before     []string `json:"before,omitempty"`
	After      []string `json:"after,omitempty"`
}

type RemediationPlan struct {
	Commands         []RemediationCommand `json:"commands,omitempty"`
	PatchSuggestions []PatchSuggestion    `json:"patch_suggestions,omitempty"`
	CIConfigDiffs    []CIConfigDiff       `json:"ci_config_diffs,omitempty"`
}

type ArtifactWorkflowRecommendation struct {
	Ref    string            `json:"ref,omitempty"`
	Inputs map[string]string `json:"inputs,omitempty"`
}

type FailureArtifact struct {
	SchemaVersion           string                           `json:"schema_version,omitempty"`
	Status                  ArtifactStatus                   `json:"status,omitempty"`
	Fingerprint             string                           `json:"fingerprint,omitempty"`
	MatchedPlaybook         *ArtifactPlaybook                `json:"matched_playbook,omitempty"`
	Evidence                []string                         `json:"evidence,omitempty"`
	Confidence              float64                          `json:"confidence,omitempty"`
	Environment             ArtifactEnvironment              `json:"environment"`
	HistoryContext          *ArtifactHistoryContext          `json:"history_context,omitempty"`
	FixSteps                []string                         `json:"fix_steps,omitempty"`
	CandidateClusters       []CandidateCluster               `json:"candidate_clusters,omitempty"`
	DominantSignals         []string                         `json:"dominant_signals,omitempty"`
	Facts                   map[string]string                `json:"facts,omitempty"`
	SuggestedPlaybookSeed   *SuggestedPlaybookSeed           `json:"suggested_playbook_seed,omitempty"`
	Remediation             *RemediationPlan                 `json:"remediation,omitempty"`
	WorkflowRecommendations []ArtifactWorkflowRecommendation `json:"workflow_recommendations,omitempty"`
}

// SilentFinding is a single finding produced by a built-in silent-failure
// detector. Silent findings are attached to Analysis.SilentFindings.
//
// Precedence rule: silent findings supplement normal playbook matches. When a
// normal playbook match exists, that match remains primary and silent findings
// are reported as secondary findings. Silent findings are promoted to the
// primary failure classification only when no normal playbook match exists.
type SilentFinding struct {
	// ID is the detector identifier (e.g. "zero-tests-executed").
	ID string `json:"id"`
	// Class is always "silent_failure".
	Class string `json:"class"`
	// Severity is one of "high", "medium", or "low".
	Severity string `json:"severity"`
	// Confidence is one of "high", "medium", or "low".
	Confidence string `json:"confidence"`
	// Explanation is a short human-readable description of the finding.
	Explanation string `json:"explanation"`
	// Evidence lists the log lines (or patterns) that triggered this finding.
	Evidence []string `json:"evidence,omitempty"`
}

// Analysis is the complete output of a log analysis run.
// Results is empty (not nil) when no playbook matched.
type Analysis struct {
	Results               []Result               `json:"results"`
	Context               Context                `json:"context"`
	Fingerprint           string                 `json:"fingerprint,omitempty"`
	InputHash             string                 `json:"input_hash,omitempty"`
	OutputHash            string                 `json:"output_hash,omitempty"`
	Source                string                 `json:"source,omitempty"`
	RepoContext           *RepoContext           `json:"repo_context,omitempty"`
	Delta                 *Delta                 `json:"delta,omitempty"`
	Differential          *DifferentialDiagnosis `json:"differential,omitempty"`
	PackProvenances       []PackProvenance       `json:"pack_provenance,omitempty"`
	Metrics               *Metrics               `json:"metrics,omitempty"`
	Policy                *Policy                `json:"policy,omitempty"`
	Status                ArtifactStatus         `json:"status,omitempty"`
	CandidateClusters     []CandidateCluster     `json:"candidate_clusters,omitempty"`
	DominantSignals       []string               `json:"dominant_signals,omitempty"`
	SuggestedPlaybookSeed *SuggestedPlaybookSeed `json:"suggested_playbook_seed,omitempty"`
	Artifact              *FailureArtifact       `json:"artifact,omitempty"`
	// SilentFindings holds results from the built-in silent-failure detector
	// pass.  Non-nil only when at least one silent finding was detected.
	SilentFindings []SilentFinding `json:"findings,omitempty"`
}

// Metrics is the machine-readable pipeline reliability summary.
// Fields are absent (nil or zero) when insufficient data is available.
// TSS is always the first-class metric; PHI and FPC require an explicit
// history artifact when supplied via --history-file.
type Metrics struct {
	// TSS is the Trace Stability Score [0,1]: fraction of locally-stored
	// analysis runs where the same failure pattern appeared.
	// Absent unless local history contains at least 2 matched entries.
	TSS *float64 `json:"tss,omitempty"`
	// FPC is the Failure Pattern Coverage [0,1]: fraction of all runs in
	// the supplied history file that matched a known playbook.
	// Absent unless the history file contains at least 3 entries.
	FPC *float64 `json:"fpc,omitempty"`
	// PHI is the Pipeline Health Index [0,1]: composite score derived from
	// FPC and the dominant-failure share of the supplied history.
	// Absent unless the history file contains at least 5 entries.
	PHI *float64 `json:"phi,omitempty"`
	// HistoryCount is the number of local history entries used to compute TSS.
	HistoryCount int `json:"history_count,omitempty"`
	// DriftComponents lists factors that are degrading pipeline reliability.
	// Populated when at least one metric falls below a warning threshold.
	DriftComponents []string `json:"drift_components,omitempty"`
}

// MetricsHistoryEntry is a single past analysis run from an explicit history
// file supplied via --history-file. Used to compute FPC and PHI.
type MetricsHistoryEntry struct {
	Matched   bool   `json:"matched"`
	FailureID string `json:"failure_id,omitempty"`
	Severity  string `json:"severity,omitempty"`
}

// Policy is the machine-readable advisory policy recommendation derived from
// reliability metrics. It is purely advisory: Faultline does not trigger
// retries, suite routing, or CI orchestration. When metrics are absent,
// Policy is also absent.
//
// Recommendation values (in increasing urgency):
//   - "ok":         metrics look healthy or there is insufficient history.
//   - "observe":    a pattern is emerging but not yet at quarantine threshold.
//   - "quarantine": persistent recurrence or low pipeline health; recommend
//     isolating the test or pipeline path for review.
//   - "blocking":   high-confidence persistent critical failure that should
//     block the pipeline until resolved.
type Policy struct {
	// Recommendation is one of "ok", "observe", "quarantine", or "blocking".
	Recommendation string `json:"recommendation"`
	// Reason is a short human-readable explanation of why this recommendation
	// was made.
	Reason string `json:"reason,omitempty"`
	// Basis lists the metric names that drove the recommendation (e.g. "tss",
	// "fpc", "phi").
	Basis []string `json:"basis,omitempty"`
}
