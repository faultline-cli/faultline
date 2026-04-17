package model

// Package model defines the shared data types used across all Faultline packages.
// No other internal packages should be imported here.

// Playbook is a failure definition loaded from a YAML file.
type Playbook struct {
	ID               string          `yaml:"id" json:"id"`
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
	Any  []string `yaml:"any" json:"any"`
	All  []string `yaml:"all" json:"all"`
	None []string `yaml:"none" json:"none,omitempty"`
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
	Playbook     Playbook              `json:"playbook"`
	Detector     string                `json:"detector,omitempty"`
	Score        float64               `json:"score"`
	Confidence   float64               `json:"confidence"`
	Evidence     []string              `json:"evidence"`
	EvidenceBy   EvidenceBundle        `json:"evidence_by,omitempty"`
	Explanation  ResultExplanation     `json:"explanation,omitempty"`
	Breakdown    ScoreBreakdown        `json:"breakdown,omitempty"`
	ChangeStatus string                `json:"change_status,omitempty"`
	SeenCount    int                   `json:"seen_count"`
	Ranking      *Ranking              `json:"ranking,omitempty"`
	Hypothesis   *HypothesisAssessment `json:"hypothesis,omitempty"`
}

// RepoContext holds git repository context enrichment from a recent commit window.
// Only populated when the --git flag is used.
type RepoContext struct {
	RepoRoot           string           `json:"repo_root"`
	RecentFiles        []string         `json:"recent_files,omitempty"`
	RelatedCommits     []RepoCommit     `json:"related_commits,omitempty"`
	HotspotDirectories []string         `json:"hotspot_directories,omitempty"`
	CoChangeHints      []string         `json:"co_change_hints,omitempty"`
	HotfixSignals      []string         `json:"hotfix_signals,omitempty"`
	DriftSignals       []string         `json:"drift_signals,omitempty"`
	Topology           *TopologySignals `json:"topology,omitempty"`
}

// RepoCommit is a trimmed commit for output.
type RepoCommit struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
}

// Analysis is the complete output of a log analysis run.
// Results is empty (not nil) when no playbook matched.
type Analysis struct {
	Results      []Result               `json:"results"`
	Context      Context                `json:"context"`
	Fingerprint  string                 `json:"fingerprint,omitempty"`
	Source       string                 `json:"source,omitempty"`
	RepoContext  *RepoContext           `json:"repo_context,omitempty"`
	Delta        *Delta                 `json:"delta,omitempty"`
	Differential *DifferentialDiagnosis `json:"differential,omitempty"`
}
