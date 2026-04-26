package model

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
