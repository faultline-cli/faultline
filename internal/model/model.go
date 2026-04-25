package model

// Package model defines the shared data types used across all Faultline packages.
// No other internal packages should be imported here.

// Playbook is a failure definition loaded from a YAML file.
type Playbook struct {
	ID               string          `yaml:"id" json:"id"`
	Extends          string          `yaml:"extends,omitempty" json:"extends,omitempty"`
	Title            string          `yaml:"title" json:"title"`
	Category         string          `yaml:"category" json:"category"`
	Severity         string          `yaml:"severity" json:"severity"`
	Detector         string          `yaml:"detector,omitempty" json:"detector,omitempty"`
	BaseScore        float64         `yaml:"base_score" json:"base_score"`
	Tags             []string        `yaml:"tags" json:"tags"`
	StageHints       []string        `yaml:"stage_hints" json:"stage_hints"`
	Match            MatchSpec       `yaml:"match" json:"match"`
	Source           SourceSpec      `yaml:"source,omitempty" json:"source,omitempty"`
	Summary          string          `yaml:"summary,omitempty" json:"summary,omitempty"`
	Diagnosis        string          `yaml:"diagnosis,omitempty" json:"diagnosis,omitempty"`
	Fix              string          `yaml:"fix,omitempty" json:"fix,omitempty"`
	Validation       string          `yaml:"validation,omitempty" json:"validation,omitempty"`
	WhyItMatters     string          `yaml:"why_it_matters,omitempty" json:"why_it_matters,omitempty"`
	RequiresDelta    bool            `yaml:"requires_delta,omitempty" json:"requires_delta,omitempty"`
	DeltaBoost       []DeltaBoost    `yaml:"delta_boost,omitempty" json:"delta_boost,omitempty"`
	RequiresTopology bool            `yaml:"requires_topology,omitempty" json:"requires_topology,omitempty"`
	TopologyBoost    []TopologyBoost `yaml:"topology_boost,omitempty" json:"topology_boost,omitempty"`
	Workflow         WorkflowSpec    `yaml:"workflow" json:"workflow"`
	Remediation      RemediationSpec `yaml:"remediation,omitempty" json:"remediation,omitempty"`
	Hooks            PlaybookHooks   `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	Metadata         PlaybookMeta    `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Scoring          ScoringConfig   `yaml:"scoring,omitempty" json:"scoring,omitempty"`
	Contextual       ContextPolicy   `yaml:"context_filters,omitempty" json:"context_filters,omitempty"`
	Hypothesis       HypothesisSpec  `yaml:"hypothesis,omitempty" json:"hypothesis,omitempty"`
}

type DeltaBoost struct {
	Signal string  `yaml:"signal,omitempty" json:"signal,omitempty"`
	Weight float64 `yaml:"weight,omitempty" json:"weight,omitempty"`
}

// TopologyBoost amplifies (or discounts) a playbook score when the given
// topology signal is active for the current repository context.
type TopologyBoost struct {
	Signal string  `yaml:"signal,omitempty" json:"signal,omitempty"`
	Weight float64 `yaml:"weight,omitempty" json:"weight,omitempty"`
}

// TopologyNode represents a single path element in the repository ownership
// graph. It is derived from CODEOWNERS and the top-level directory structure.
type TopologyNode struct {
	Path   string   `json:"path"`
	Owners []string `json:"owners,omitempty"`
}

// TopologySignals holds the structural signals derived from the repository
// ownership graph for a given analysis context. Only populated when --git
// is active and a CODEOWNERS file is present.
type TopologySignals struct {
	ActiveSignals     []string `json:"active_signals,omitempty"`
	OwnerZones        []string `json:"owner_zones,omitempty"`
	BoundaryCrossed   bool     `json:"boundary_crossed,omitempty"`
	UpstreamChanged   bool     `json:"upstream_changed,omitempty"`
	OwnershipMismatch bool     `json:"ownership_mismatch,omitempty"`
	FailureClustered  bool     `json:"failure_clustered,omitempty"`
}

// MatchSpec holds declarative match patterns for a Playbook.
// Any is matched as OR: at least one pattern must appear in the log.
// All is matched as AND: every pattern must appear in the log.
type MatchSpec struct {
	Any     []string            `yaml:"any" json:"any"`
	All     []string            `yaml:"all" json:"all"`
	None    []string            `yaml:"none" json:"none,omitempty"`
	Use     []string            `yaml:"use,omitempty" json:"use,omitempty"`
	Partial []PartialMatchGroup `yaml:"partial,omitempty" json:"partial,omitempty"`
}

// PartialMatchGroup defines a deterministic sub-pattern cluster where a
// configurable minimum number of patterns must match before the group is
// considered satisfied.
type PartialMatchGroup struct {
	ID       string   `yaml:"id,omitempty" json:"id,omitempty"`
	Label    string   `yaml:"label,omitempty" json:"label,omitempty"`
	Minimum  int      `yaml:"minimum,omitempty" json:"minimum,omitempty"`
	Patterns []string `yaml:"patterns,omitempty" json:"patterns,omitempty"`
}

// SourceSpec defines a reusable source-code detection schema.
type SourceSpec struct {
	Triggers           []SignalMatcher   `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	Amplifiers         []SignalMatcher   `yaml:"amplifiers,omitempty" json:"amplifiers,omitempty"`
	Mitigations        []SignalMatcher   `yaml:"mitigations,omitempty" json:"mitigations,omitempty"`
	Suppressions       []SuppressionRule `yaml:"suppressions,omitempty" json:"suppressions,omitempty"`
	Context            []SignalMatcher   `yaml:"context,omitempty" json:"context,omitempty"`
	CompoundSignals    []CompoundSignal  `yaml:"compound_signals,omitempty" json:"compound_signals,omitempty"`
	LocalConsistency   []ConsistencyRule `yaml:"local_consistency,omitempty" json:"local_consistency,omitempty"`
	PathClasses        []PathClassRule   `yaml:"path_classes,omitempty" json:"path_classes,omitempty"`
	ChangeSensitivity  ChangeSensitivity `yaml:"change_sensitivity,omitempty" json:"change_sensitivity,omitempty"`
	SafeContextClasses []SafeContextRule `yaml:"safe_context,omitempty" json:"safe_context,omitempty"`
}

type SignalMatcher struct {
	ID           string   `yaml:"id,omitempty" json:"id,omitempty"`
	Label        string   `yaml:"label,omitempty" json:"label,omitempty"`
	Description  string   `yaml:"description,omitempty" json:"description,omitempty"`
	Patterns     []string `yaml:"patterns,omitempty" json:"patterns,omitempty"`
	PathIncludes []string `yaml:"path_includes,omitempty" json:"path_includes,omitempty"`
	PathExcludes []string `yaml:"path_excludes,omitempty" json:"path_excludes,omitempty"`
	Scopes       []string `yaml:"scopes,omitempty" json:"scopes,omitempty"`
	Weight       float64  `yaml:"weight,omitempty" json:"weight,omitempty"`
	Required     bool     `yaml:"required,omitempty" json:"required,omitempty"`
}

type CompoundSignal struct {
	ID             string   `yaml:"id,omitempty" json:"id,omitempty"`
	Label          string   `yaml:"label,omitempty" json:"label,omitempty"`
	Require        []string `yaml:"require,omitempty" json:"require,omitempty"`
	Scope          string   `yaml:"scope,omitempty" json:"scope,omitempty"`
	Bonus          float64  `yaml:"bonus,omitempty" json:"bonus,omitempty"`
	Required       bool     `yaml:"required,omitempty" json:"required,omitempty"`
	AllowMitigated bool     `yaml:"allow_mitigated,omitempty" json:"allow_mitigated,omitempty"`
}

type ConsistencyRule struct {
	ID                string   `yaml:"id,omitempty" json:"id,omitempty"`
	Label             string   `yaml:"label,omitempty" json:"label,omitempty"`
	BaselineSignalIDs []string `yaml:"baseline_signal_ids,omitempty" json:"baseline_signal_ids,omitempty"`
	ExpectedSignalID  string   `yaml:"expected_signal_id,omitempty" json:"expected_signal_id,omitempty"`
	Scope             string   `yaml:"scope,omitempty" json:"scope,omitempty"`
	MinimumPeers      int      `yaml:"minimum_peers,omitempty" json:"minimum_peers,omitempty"`
	Threshold         float64  `yaml:"threshold,omitempty" json:"threshold,omitempty"`
	Amplifier         float64  `yaml:"amplifier,omitempty" json:"amplifier,omitempty"`
}

type SuppressionRule struct {
	Style       string   `yaml:"style,omitempty" json:"style,omitempty"`
	Pattern     string   `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Paths       []string `yaml:"paths,omitempty" json:"paths,omitempty"`
	Playbooks   []string `yaml:"playbooks,omitempty" json:"playbooks,omitempty"`
	Reason      string   `yaml:"reason,omitempty" json:"reason,omitempty"`
	ExpiresOn   string   `yaml:"expires_on,omitempty" json:"expires_on,omitempty"`
	Discount    float64  `yaml:"discount,omitempty" json:"discount,omitempty"`
	SuppressAll bool     `yaml:"suppress_all,omitempty" json:"suppress_all,omitempty"`
}

type PathClassRule struct {
	Class    string   `yaml:"class,omitempty" json:"class,omitempty"`
	Paths    []string `yaml:"paths,omitempty" json:"paths,omitempty"`
	Adjust   float64  `yaml:"adjust,omitempty" json:"adjust,omitempty"`
	HotPath  bool     `yaml:"hot_path,omitempty" json:"hot_path,omitempty"`
	Critical bool     `yaml:"critical,omitempty" json:"critical,omitempty"`
}

type SafeContextRule struct {
	ID       string   `yaml:"id,omitempty" json:"id,omitempty"`
	Label    string   `yaml:"label,omitempty" json:"label,omitempty"`
	Paths    []string `yaml:"paths,omitempty" json:"paths,omitempty"`
	Patterns []string `yaml:"patterns,omitempty" json:"patterns,omitempty"`
	Discount float64  `yaml:"discount,omitempty" json:"discount,omitempty"`
}

type ChangeSensitivity struct {
	NewFileBonus        float64 `yaml:"new_file_bonus,omitempty" json:"new_file_bonus,omitempty"`
	ModifiedLineBonus   float64 `yaml:"modified_line_bonus,omitempty" json:"modified_line_bonus,omitempty"`
	LegacyDiscount      float64 `yaml:"legacy_discount,omitempty" json:"legacy_discount,omitempty"`
	PreferChangedScopes bool    `yaml:"prefer_changed_scopes,omitempty" json:"prefer_changed_scopes,omitempty"`
}

type PlaybookMeta struct {
	SchemaVersion string `yaml:"schema_version,omitempty" json:"schema_version,omitempty"`
	PackName      string `yaml:"-" json:"pack_name,omitempty"`
	PackRoot      string `yaml:"-" json:"pack_root,omitempty"`
	PackVersion   string `yaml:"-" json:"pack_version,omitempty"`
	PackSourceURL string `yaml:"-" json:"pack_source_url,omitempty"`
	PackPinnedRef string `yaml:"-" json:"pack_pinned_ref,omitempty"`
	SourceFile    string `yaml:"-" json:"source_file,omitempty"`
}

type ScoringConfig struct {
	BaseTriggerWeight          float64 `yaml:"base_trigger_weight,omitempty" json:"base_trigger_weight,omitempty"`
	DefaultAmplifierWeight     float64 `yaml:"default_amplifier_weight,omitempty" json:"default_amplifier_weight,omitempty"`
	DefaultMitigationDiscount  float64 `yaml:"default_mitigation_discount,omitempty" json:"default_mitigation_discount,omitempty"`
	DefaultSuppressionDiscount float64 `yaml:"default_suppression_discount,omitempty" json:"default_suppression_discount,omitempty"`
	HotPathBonus               float64 `yaml:"hot_path_bonus,omitempty" json:"hot_path_bonus,omitempty"`
	BlastRadiusBonus           float64 `yaml:"blast_radius_bonus,omitempty" json:"blast_radius_bonus,omitempty"`
	SafeContextDiscount        float64 `yaml:"safe_context_discount,omitempty" json:"safe_context_discount,omitempty"`
}

type ContextPolicy struct {
	PathIncludes []string `yaml:"path_includes,omitempty" json:"path_includes,omitempty"`
	PathExcludes []string `yaml:"path_excludes,omitempty" json:"path_excludes,omitempty"`
}

// WorkflowSpec defines deterministic local follow-up metadata for a playbook.
type WorkflowSpec struct {
	LikelyFiles []string `yaml:"likely_files" json:"likely_files,omitempty"`
	LocalRepro  []string `yaml:"local_repro" json:"local_repro,omitempty"`
	Verify      []string `yaml:"verify" json:"verify,omitempty"`
}

// RemediationSpec declares one or more typed remediation workflows that can be
// recommended for a matched playbook.
type RemediationSpec struct {
	Workflows []RemediationWorkflowRef `yaml:"workflows,omitempty" json:"workflows,omitempty"`
}

// RemediationWorkflowRef points at a static workflow definition and declares
// how workflow inputs should be resolved from the analysis artifact.
type RemediationWorkflowRef struct {
	Ref    string                             `yaml:"ref,omitempty" json:"ref,omitempty"`
	Inputs map[string]RemediationInputBinding `yaml:"inputs,omitempty" json:"inputs,omitempty"`
}

// RemediationInputBinding resolves a workflow input either from a stable
// artifact/context path or from a literal authored value.
type RemediationInputBinding struct {
	From  string `yaml:"from,omitempty" json:"from,omitempty"`
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
}

type HookCategory string

const (
	HookCategoryVerify    HookCategory = "verify"
	HookCategoryCollect   HookCategory = "collect"
	HookCategoryRemediate HookCategory = "remediate"
)

type HookMode string

const (
	HookModeOff         HookMode = "off"
	HookModeVerifyOnly  HookMode = "verify-only"
	HookModeCollectOnly HookMode = "collect-only"
	HookModeSafe        HookMode = "safe"
	HookModeFull        HookMode = "full"
)

type HookKind string

const (
	HookKindFileExists           HookKind = "file_exists"
	HookKindDirExists            HookKind = "dir_exists"
	HookKindEnvVarPresent        HookKind = "env_var_present"
	HookKindCommandExitZero      HookKind = "command_exit_zero"
	HookKindCommandOutputMatches HookKind = "command_output_matches"
	HookKindCommandOutputCapture HookKind = "command_output_capture"
	HookKindReadFileExcerpt      HookKind = "read_file_excerpt"
)

type PlaybookHooks struct {
	Verify    []HookDefinition `yaml:"verify,omitempty" json:"verify,omitempty"`
	Collect   []HookDefinition `yaml:"collect,omitempty" json:"collect,omitempty"`
	Remediate []HookDefinition `yaml:"remediate,omitempty" json:"remediate,omitempty"`
	Disable   []string         `yaml:"disable,omitempty" json:"disable,omitempty"`
}

type HookDefinition struct {
	ID              string   `yaml:"id,omitempty" json:"id,omitempty"`
	Use             string   `yaml:"use,omitempty" json:"use,omitempty"`
	Extends         string   `yaml:"extends,omitempty" json:"extends,omitempty"`
	Kind            HookKind `yaml:"kind,omitempty" json:"kind,omitempty"`
	Path            string   `yaml:"path,omitempty" json:"path,omitempty"`
	EnvVar          string   `yaml:"env_var,omitempty" json:"env_var,omitempty"`
	Command         []string `yaml:"command,omitempty" json:"command,omitempty"`
	Pattern         string   `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Lines           int      `yaml:"lines,omitempty" json:"lines,omitempty"`
	MaxBytes        int      `yaml:"max_bytes,omitempty" json:"max_bytes,omitempty"`
	ConfidenceDelta float64  `yaml:"confidence_delta,omitempty" json:"confidence_delta,omitempty"`
	Metadata        HookMeta `yaml:"-" json:"metadata,omitempty"`
}

type HookMeta struct {
	SourcePack string `json:"source_pack,omitempty"`
	SourceFile string `json:"source_file,omitempty"`
}

type HookStatus string

const (
	HookStatusExecuted HookStatus = "executed"
	HookStatusSkipped  HookStatus = "skipped"
	HookStatusBlocked  HookStatus = "blocked"
	HookStatusFailed   HookStatus = "failed"
)

type HookFact struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type HookResult struct {
	ID              string       `json:"id"`
	Category        HookCategory `json:"category"`
	Kind            HookKind     `json:"kind,omitempty"`
	Status          HookStatus   `json:"status"`
	Passed          *bool        `json:"passed,omitempty"`
	ConfidenceDelta float64      `json:"confidence_delta,omitempty"`
	Reason          string       `json:"reason,omitempty"`
	Facts           []HookFact   `json:"facts,omitempty"`
	Evidence        []string     `json:"evidence,omitempty"`
	SourcePack      string       `json:"source_pack,omitempty"`
	SourceFile      string       `json:"source_file,omitempty"`
}

type HookReport struct {
	Mode            HookMode     `json:"mode,omitempty"`
	BaseConfidence  float64      `json:"base_confidence,omitempty"`
	ConfidenceDelta float64      `json:"confidence_delta,omitempty"`
	FinalConfidence float64      `json:"final_confidence,omitempty"`
	Results         []HookResult `json:"results,omitempty"`
}

type HookHistorySummary struct {
	TotalCount    int    `json:"total_count,omitempty"`
	ExecutedCount int    `json:"executed_count,omitempty"`
	PassedCount   int    `json:"passed_count,omitempty"`
	FailedCount   int    `json:"failed_count,omitempty"`
	BlockedCount  int    `json:"blocked_count,omitempty"`
	SkippedCount  int    `json:"skipped_count,omitempty"`
	LastSeenAt    string `json:"last_seen_at,omitempty"`
}

type HypothesisSpec struct {
	Supports       []HypothesisSignal        `yaml:"supports,omitempty" json:"supports,omitempty"`
	Contradicts    []HypothesisSignal        `yaml:"contradicts,omitempty" json:"contradicts,omitempty"`
	Discriminators []HypothesisDiscriminator `yaml:"discriminators,omitempty" json:"discriminators,omitempty"`
	Excludes       []HypothesisSignal        `yaml:"excludes,omitempty" json:"excludes,omitempty"`
}

type HypothesisSignal struct {
	Signal string  `yaml:"signal,omitempty" json:"signal,omitempty"`
	Weight float64 `yaml:"weight,omitempty" json:"weight,omitempty"`
}

type HypothesisDiscriminator struct {
	Description string  `yaml:"description,omitempty" json:"description,omitempty"`
	Signal      string  `yaml:"signal,omitempty" json:"signal,omitempty"`
	Weight      float64 `yaml:"weight,omitempty" json:"weight,omitempty"`
}

type HypothesisMatch struct {
	Signal      string   `json:"signal,omitempty"`
	Description string   `json:"description,omitempty"`
	Weight      float64  `json:"weight,omitempty"`
	Evidence    []string `json:"evidence,omitempty"`
}

type HypothesisAssessment struct {
	BaseScore      float64           `json:"base_score,omitempty"`
	FinalScore     float64           `json:"final_score,omitempty"`
	Eliminated     bool              `json:"eliminated,omitempty"`
	Supports       []HypothesisMatch `json:"supports,omitempty"`
	Contradicts    []HypothesisMatch `json:"contradicts,omitempty"`
	Discriminators []HypothesisMatch `json:"discriminators,omitempty"`
	Excludes       []HypothesisMatch `json:"excludes,omitempty"`
	Why            []string          `json:"why,omitempty"`
	WhyLessLikely  []string          `json:"why_less_likely,omitempty"`
	RuledOutBy     []string          `json:"ruled_out_by,omitempty"`
	DisproofChecks []string          `json:"disproof_checks,omitempty"`
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

// RepoContext holds git repository context enrichment from a recent commit window.
// It is populated whenever git enrichment is enabled; the shipped CLI surfaces
// enable that path by default and allow `--git=false` to disable it.
type RepoContext struct {
	RepoRoot           string       `json:"repo_root"`
	RecentFiles        []string     `json:"recent_files,omitempty"`
	RelatedCommits     []RepoCommit `json:"related_commits,omitempty"`
	HotspotDirectories []string     `json:"hotspot_directories,omitempty"`
	CoChangeHints      []string     `json:"co_change_hints,omitempty"`
	HotfixSignals      []string     `json:"hotfix_signals,omitempty"`
	DriftSignals       []string     `json:"drift_signals,omitempty"`
	// ConfigDriftSignals are recently changed dependency or config files
	// (go.mod, Dockerfile, package.json, etc.) relevant to the failure.
	ConfigDriftSignals []string `json:"config_drift_signals,omitempty"`
	// CIChangeSignals are recently changed CI pipeline config files
	// (.github/workflows, Makefile, etc.) relevant to the failure.
	CIChangeSignals []string `json:"ci_change_signals,omitempty"`
	// LargeCommitSignals are subjects of large commits (touching many files)
	// that may indicate a high blast-radius change.
	LargeCommitSignals []string         `json:"large_commit_signals,omitempty"`
	Topology           *TopologySignals `json:"topology,omitempty"`
}

// RepoCommit is a trimmed commit for output.
type RepoCommit struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
}

// PackProvenance records which installed pack contributed playbooks to an analysis.
// Version and PinnedRef are empty for the bundled starter pack.
type PackProvenance struct {
	Name          string `json:"name"`
	Version       string `json:"version,omitempty"`
	SourceURL     string `json:"source_url,omitempty"`
	PinnedRef     string `json:"pinned_ref,omitempty"`
	PlaybookCount int    `json:"playbook_count"`
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

type WorkflowExecutionMode string

const (
	WorkflowExecutionModeExplain WorkflowExecutionMode = "explain"
	WorkflowExecutionModeDryRun  WorkflowExecutionMode = "dry-run"
	WorkflowExecutionModeApply   WorkflowExecutionMode = "apply"
)

type WorkflowExecutionStatus string

const (
	WorkflowExecutionStatusPlanned   WorkflowExecutionStatus = "planned"
	WorkflowExecutionStatusSucceeded WorkflowExecutionStatus = "succeeded"
	WorkflowExecutionStatusFailed    WorkflowExecutionStatus = "failed"
	WorkflowExecutionStatusBlocked   WorkflowExecutionStatus = "blocked"
	WorkflowExecutionStatusSkipped   WorkflowExecutionStatus = "skipped"
)

type WorkflowVerificationStatus string

const (
	WorkflowVerificationStatusPending WorkflowVerificationStatus = "pending"
	WorkflowVerificationStatusPassed  WorkflowVerificationStatus = "passed"
	WorkflowVerificationStatusFailed  WorkflowVerificationStatus = "failed"
)

type WorkflowStepResult struct {
	Phase              string                     `json:"phase,omitempty"`
	StepID             string                     `json:"step_id,omitempty"`
	StepType           string                     `json:"step_type,omitempty"`
	SafetyClass        string                     `json:"safety_class,omitempty"`
	Status             WorkflowExecutionStatus    `json:"status,omitempty"`
	VerificationStatus WorkflowVerificationStatus `json:"verification_status,omitempty"`
	StartedAt          string                     `json:"started_at,omitempty"`
	FinishedAt         string                     `json:"finished_at,omitempty"`
	Changed            *bool                      `json:"changed,omitempty"`
	Message            string                     `json:"message,omitempty"`
	Outputs            map[string]string          `json:"outputs,omitempty"`
	Error              string                     `json:"error,omitempty"`
}

type WorkflowExecutionRecord struct {
	SchemaVersion      string                     `json:"schema_version,omitempty"`
	ExecutionID        string                     `json:"execution_id,omitempty"`
	WorkflowID         string                     `json:"workflow_id,omitempty"`
	Title              string                     `json:"title,omitempty"`
	Mode               WorkflowExecutionMode      `json:"mode,omitempty"`
	SourceFingerprint  string                     `json:"source_fingerprint,omitempty"`
	SourceFailureID    string                     `json:"source_failure_id,omitempty"`
	StartedAt          string                     `json:"started_at,omitempty"`
	FinishedAt         string                     `json:"finished_at,omitempty"`
	ResolvedInputs     map[string]string          `json:"resolved_inputs,omitempty"`
	StepResults        []WorkflowStepResult       `json:"step_results,omitempty"`
	VerificationStatus WorkflowVerificationStatus `json:"verification_status,omitempty"`
	Status             WorkflowExecutionStatus    `json:"status,omitempty"`
}

type WorkflowExecutionSummary struct {
	ExecutionID        string                     `json:"execution_id,omitempty"`
	WorkflowID         string                     `json:"workflow_id,omitempty"`
	Title              string                     `json:"title,omitempty"`
	Mode               WorkflowExecutionMode      `json:"mode,omitempty"`
	SourceFingerprint  string                     `json:"source_fingerprint,omitempty"`
	SourceFailureID    string                     `json:"source_failure_id,omitempty"`
	StartedAt          string                     `json:"started_at,omitempty"`
	FinishedAt         string                     `json:"finished_at,omitempty"`
	VerificationStatus WorkflowVerificationStatus `json:"verification_status,omitempty"`
	Status             WorkflowExecutionStatus    `json:"status,omitempty"`
}
