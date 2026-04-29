package model

// Playbook is a failure definition loaded from a YAML file.
type Playbook struct {
	ID         string    `yaml:"id" json:"id"`
	Extends    string    `yaml:"extends,omitempty" json:"extends,omitempty"`
	Title      string    `yaml:"title" json:"title"`
	Category   string    `yaml:"category" json:"category"`
	Severity   string    `yaml:"severity" json:"severity"`
	Detector   string    `yaml:"detector,omitempty" json:"detector,omitempty"`
	BaseScore  float64   `yaml:"base_score" json:"base_score"`
	Tags       []string  `yaml:"tags" json:"tags"`
	StageHints []string  `yaml:"stage_hints" json:"stage_hints"`
	Domain     string    `yaml:"domain,omitempty" json:"domain,omitempty"`
	Class      string    `yaml:"class,omitempty" json:"class,omitempty"`
	Mode       string    `yaml:"mode,omitempty" json:"mode,omitempty"`
	Match      MatchSpec `yaml:"match" json:"match"`
	// NativeAny holds the match.any patterns that were defined directly in
	// this playbook (not inherited from a parent via extends). It is
	// populated during inheritance resolution and used to limit anyScore
	// to the child's own distinctive patterns, preserving the parent's
	// confidence on logs where only generic parent patterns fire.
	// This field is never serialised to YAML or JSON output.
	NativeAny        []string        `yaml:"-" json:"-"`
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

// MatchSpec holds declarative match patterns for a Playbook.
// Any is matched as OR: at least one pattern must appear in the log.
// All is matched as AND: every pattern must appear in the log.
// WithinLines optionally restricts the compound bonus for match.all: when set
// to a positive value N, the +2.0 compound bonus is only awarded when all
// match.all patterns appear within N lines of each other.  Individual pattern
// scores (+1.5 each) are unaffected.  Zero (the default) disables the check.
type MatchSpec struct {
	Any         []string            `yaml:"any" json:"any"`
	All         []string            `yaml:"all" json:"all"`
	WithinLines int                 `yaml:"within_lines,omitempty" json:"within_lines,omitempty"`
	None        []string            `yaml:"none" json:"none,omitempty"`
	Use         []string            `yaml:"use,omitempty" json:"use,omitempty"`
	Partial     []PartialMatchGroup `yaml:"partial,omitempty" json:"partial,omitempty"`
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
