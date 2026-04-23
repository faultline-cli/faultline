package store

type migration struct {
	version int
	name    string
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		name:    "init",
		sql: `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS analysis_runs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	surface TEXT NOT NULL,
	source_kind TEXT NOT NULL,
	source TEXT,
	input_hash TEXT,
	output_hash TEXT,
	matched INTEGER NOT NULL DEFAULT 0,
	top_failure_id TEXT,
	top_signature_hash TEXT,
	fingerprint TEXT,
	started_at TEXT NOT NULL,
	completed_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_analysis_runs_completed_at
	ON analysis_runs(completed_at DESC);

CREATE INDEX IF NOT EXISTS idx_analysis_runs_input_hash
	ON analysis_runs(input_hash);

CREATE INDEX IF NOT EXISTS idx_analysis_runs_top_failure_id
	ON analysis_runs(top_failure_id);

CREATE TABLE IF NOT EXISTS findings (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	run_id INTEGER NOT NULL,
	rank INTEGER NOT NULL,
	failure_id TEXT NOT NULL,
	title TEXT,
	category TEXT,
	detector TEXT,
	score REAL NOT NULL,
	confidence REAL NOT NULL,
	fingerprint TEXT,
	signature_hash TEXT,
	normalized_signature TEXT,
	evidence_excerpt_json TEXT,
	seen_at TEXT NOT NULL,
	FOREIGN KEY(run_id) REFERENCES analysis_runs(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_findings_run_rank
	ON findings(run_id, rank);

CREATE INDEX IF NOT EXISTS idx_findings_signature_hash
	ON findings(signature_hash);

CREATE INDEX IF NOT EXISTS idx_findings_failure_id
	ON findings(failure_id);

CREATE TABLE IF NOT EXISTS signatures (
	signature_hash TEXT PRIMARY KEY,
	failure_id TEXT NOT NULL,
	normalized_signature TEXT NOT NULL,
	first_seen_at TEXT NOT NULL,
	last_seen_at TEXT NOT NULL,
	occurrence_count INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS playbook_matches (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	run_id INTEGER NOT NULL,
	rank INTEGER NOT NULL,
	playbook_id TEXT NOT NULL,
	detector TEXT,
	score REAL NOT NULL,
	confidence REAL NOT NULL,
	matched_at TEXT NOT NULL,
	FOREIGN KEY(run_id) REFERENCES analysis_runs(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_playbook_matches_run_rank
	ON playbook_matches(run_id, rank);

CREATE INDEX IF NOT EXISTS idx_playbook_matches_playbook_id
	ON playbook_matches(playbook_id);

CREATE TABLE IF NOT EXISTS hook_results (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	run_id INTEGER NOT NULL,
	finding_id INTEGER,
	playbook_id TEXT NOT NULL,
	signature_hash TEXT,
	hook_id TEXT NOT NULL,
	category TEXT NOT NULL,
	kind TEXT,
	status TEXT NOT NULL,
	passed INTEGER,
	confidence_delta REAL NOT NULL DEFAULT 0,
	reason TEXT,
	facts_json TEXT,
	evidence_json TEXT,
	executed_at TEXT NOT NULL,
	FOREIGN KEY(run_id) REFERENCES analysis_runs(id) ON DELETE CASCADE,
	FOREIGN KEY(finding_id) REFERENCES findings(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_hook_results_signature_hash
	ON hook_results(signature_hash);

CREATE INDEX IF NOT EXISTS idx_hook_results_playbook_id
	ON hook_results(playbook_id);
`,
	},
	{
		version: 2,
		name:    "analysis-artifact",
		sql: `
ALTER TABLE analysis_runs ADD COLUMN artifact_json TEXT;
`,
	},
	{
		version: 3,
		name:    "workflow-runs",
		sql: `
CREATE TABLE IF NOT EXISTS workflow_runs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	execution_id TEXT UNIQUE,
	workflow_id TEXT NOT NULL,
	title TEXT,
	mode TEXT NOT NULL,
	source_fingerprint TEXT,
	source_failure_id TEXT,
	started_at TEXT NOT NULL,
	finished_at TEXT NOT NULL,
	verification_status TEXT NOT NULL,
	status TEXT NOT NULL,
	record_json TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_execution_id
	ON workflow_runs(execution_id);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow_id
	ON workflow_runs(workflow_id);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_finished_at
	ON workflow_runs(finished_at DESC);
`,
	},
}
