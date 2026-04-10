package model

// Package model defines the shared data types used across all Faultline packages.
// No other internal packages should be imported here.

// Playbook is a failure definition loaded from a YAML file.
type Playbook struct {
	ID         string       `yaml:"id" json:"id"`
	Title      string       `yaml:"title" json:"title"`
	Category   string       `yaml:"category" json:"category"`
	Severity   string       `yaml:"severity" json:"severity"`
	BaseScore  float64      `yaml:"base_score" json:"base_score"`
	Tags       []string     `yaml:"tags" json:"tags"`
	StageHints []string     `yaml:"stage_hints" json:"stage_hints"`
	Match      MatchSpec    `yaml:"match" json:"match"`
	Explain    string       `yaml:"explain" json:"explain"`
	Why        string       `yaml:"why" json:"why"`
	Fix        []string     `yaml:"fix" json:"fix"`
	Prevent    []string     `yaml:"prevent" json:"prevent"`
	Workflow   WorkflowSpec `yaml:"workflow" json:"workflow"`
}

// MatchSpec holds declarative match patterns for a Playbook.
// Any is matched as OR: at least one pattern must appear in the log.
// All is matched as AND: every pattern must appear in the log.
type MatchSpec struct {
	Any  []string `yaml:"any" json:"any"`
	All  []string `yaml:"all" json:"all"`
	None []string `yaml:"none" json:"none,omitempty"`
}

// WorkflowSpec defines deterministic local follow-up metadata for a playbook.
type WorkflowSpec struct {
	LikelyFiles []string `yaml:"likely_files" json:"likely_files,omitempty"`
	LocalRepro  []string `yaml:"local_repro" json:"local_repro,omitempty"`
	Verify      []string `yaml:"verify" json:"verify,omitempty"`
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

// Result is a single ranked playbook match with its scoring detail.
type Result struct {
	Playbook   Playbook `json:"playbook"`
	Score      float64  `json:"score"`
	Confidence float64  `json:"confidence"`
	Evidence   []string `json:"evidence"`
	SeenCount  int      `json:"seen_count"`
}

// RepoContext holds git repository context enrichment from a recent commit window.
// Only populated when the --git flag is used.
type RepoContext struct {
	RepoRoot           string       `json:"repo_root"`
	RecentFiles        []string     `json:"recent_files,omitempty"`
	RelatedCommits     []RepoCommit `json:"related_commits,omitempty"`
	HotspotDirectories []string     `json:"hotspot_directories,omitempty"`
	CoChangeHints      []string     `json:"co_change_hints,omitempty"`
	HotfixSignals      []string     `json:"hotfix_signals,omitempty"`
	DriftSignals       []string     `json:"drift_signals,omitempty"`
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
	Results     []Result     `json:"results"`
	Context     Context      `json:"context"`
	Fingerprint string       `json:"fingerprint,omitempty"`
	Source      string       `json:"source,omitempty"`
	RepoContext *RepoContext `json:"repo_context,omitempty"`
}
