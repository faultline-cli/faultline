package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type sqliteStore struct {
	db *sql.DB
}

func openSQLite(path string) (Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create store directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite store: %w", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable sqlite foreign keys: %w", err)
	}
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("set sqlite busy timeout: %w", err)
	}
	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable sqlite wal mode: %w", err)
	}
	store := &sqliteStore{db: db}
	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *sqliteStore) BeginRun(ctx context.Context, params BeginRunParams) (RunHandle, error) {
	if params.StartedAt.IsZero() {
		params.StartedAt = time.Now().UTC()
	}
	result, err := s.db.ExecContext(ctx, `
INSERT INTO analysis_runs (
	surface, source_kind, source, input_hash, started_at, completed_at
) VALUES (?, ?, ?, ?, ?, ?)
`,
		strings.TrimSpace(params.Surface),
		strings.TrimSpace(params.SourceKind),
		nullableString(params.Source),
		nullableString(params.InputHash),
		params.StartedAt.UTC().Format(time.RFC3339),
		params.StartedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return RunHandle{}, fmt.Errorf("insert analysis run: %w", err)
	}
	runID, err := result.LastInsertId()
	if err != nil {
		return RunHandle{}, fmt.Errorf("resolve analysis run id: %w", err)
	}
	return RunHandle{ID: runID}, nil
}

func (s *sqliteStore) CompleteRun(ctx context.Context, handle RunHandle, params CompleteRunParams) error {
	if handle.ID == 0 || params.Analysis == nil {
		return nil
	}
	completedAt := params.CompletedAt
	if completedAt.IsZero() {
		completedAt = time.Now().UTC()
	}
	analysis := params.Analysis
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin store transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var topFailureID, topSignatureHash, fingerprint string
	artifactJSON := ""
	if len(analysis.Results) > 0 {
		top := analysis.Results[0]
		topFailureID = top.Playbook.ID
		topSignatureHash = strings.TrimSpace(top.SignatureHash)
		fingerprint = strings.TrimSpace(analysis.Fingerprint)
	}
	if analysis.Artifact != nil {
		data, merr := json.Marshal(analysis.Artifact)
		if merr != nil {
			return fmt.Errorf("marshal analysis artifact: %w", merr)
		}
		artifactJSON = string(data)
	}
	_, err = tx.ExecContext(ctx, `
UPDATE analysis_runs
SET matched = ?, output_hash = ?, top_failure_id = ?, top_signature_hash = ?, fingerprint = ?, artifact_json = ?, completed_at = ?
WHERE id = ?
`,
		boolToInt(len(analysis.Results) > 0),
		nullableString(analysis.OutputHash),
		nullableString(topFailureID),
		nullableString(topSignatureHash),
		nullableString(fingerprint),
		nullableString(artifactJSON),
		completedAt.UTC().Format(time.RFC3339),
		handle.ID,
	)
	if err != nil {
		return fmt.Errorf("update analysis run: %w", err)
	}

	for i, result := range analysis.Results {
		signature := strings.TrimSpace(result.SignatureHash)
		normalizedSignature := ""
		if signature != "" {
			normalizedSignature = SignatureForResult(result).Normalized
		}
		evidenceJSON, merr := json.Marshal(firstEvidenceLines(result.Evidence))
		if merr != nil {
			return fmt.Errorf("marshal finding evidence: %w", merr)
		}
		_, err = tx.ExecContext(ctx, `
INSERT INTO findings (
	run_id, rank, failure_id, title, category, detector, score, confidence,
	fingerprint, signature_hash, normalized_signature, evidence_excerpt_json, seen_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
			handle.ID,
			i+1,
			result.Playbook.ID,
			nullableString(result.Playbook.Title),
			nullableString(result.Playbook.Category),
			nullableString(result.Detector),
			result.Score,
			result.Confidence,
			nullableString(analysis.Fingerprint),
			nullableString(signature),
			nullableString(normalizedSignature),
			string(evidenceJSON),
			completedAt.UTC().Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("insert finding: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
INSERT INTO playbook_matches (
	run_id, rank, playbook_id, detector, score, confidence, matched_at
) VALUES (?, ?, ?, ?, ?, ?, ?)
`,
			handle.ID,
			i+1,
			result.Playbook.ID,
			nullableString(result.Detector),
			result.Score,
			result.Confidence,
			completedAt.UTC().Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("insert playbook match: %w", err)
		}

		if i == 0 && signature != "" {
			_, err = tx.ExecContext(ctx, `
INSERT INTO signatures (
	signature_hash, failure_id, normalized_signature, first_seen_at, last_seen_at, occurrence_count
) VALUES (?, ?, ?, ?, ?, 1)
ON CONFLICT(signature_hash) DO UPDATE SET
	failure_id = excluded.failure_id,
	normalized_signature = excluded.normalized_signature,
	last_seen_at = excluded.last_seen_at,
	occurrence_count = signatures.occurrence_count + 1
`,
				signature,
				result.Playbook.ID,
				normalizedSignature,
				completedAt.UTC().Format(time.RFC3339),
				completedAt.UTC().Format(time.RFC3339),
			)
			if err != nil {
				return fmt.Errorf("upsert signature: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit store transaction: %w", err)
	}
	return nil
}

func (s *sqliteStore) LookupSignatureHistory(ctx context.Context, signatureHash string) (SignatureHistory, error) {
	signatureHash = strings.TrimSpace(signatureHash)
	if signatureHash == "" {
		return SignatureHistory{}, nil
	}
	var history SignatureHistory
	err := s.db.QueryRowContext(ctx, `
SELECT signature_hash, first_seen_at, last_seen_at, occurrence_count
FROM signatures
WHERE signature_hash = ?
`, signatureHash).Scan(&history.SignatureHash, &history.FirstSeenAt, &history.LastSeenAt, &history.OccurrenceCount)
	if errors.Is(err, sql.ErrNoRows) {
		return SignatureHistory{}, nil
	}
	if err != nil {
		return SignatureHistory{}, fmt.Errorf("lookup signature history: %w", err)
	}
	history.SeenBefore = history.OccurrenceCount > 0
	return history, nil
}

func (s *sqliteStore) CountSeenFailure(ctx context.Context, failureID string) (int, error) {
	failureID = strings.TrimSpace(failureID)
	if failureID == "" {
		return 0, nil
	}
	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM analysis_runs
WHERE top_failure_id = ?
`, failureID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count seen failure: %w", err)
	}
	return count, nil
}

func (s *sqliteStore) RecentTopFailures(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT top_failure_id
FROM analysis_runs
WHERE matched = 1 AND top_failure_id IS NOT NULL AND top_failure_id != ''
ORDER BY completed_at DESC
LIMIT ?
`, limit)
	if err != nil {
		return nil, fmt.Errorf("query recent failures: %w", err)
	}
	defer rows.Close()
	var values []string
	for rows.Next() {
		var failureID string
		if err := rows.Scan(&failureID); err != nil {
			return nil, fmt.Errorf("scan recent failures: %w", err)
		}
		values = append(values, failureID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent failures: %w", err)
	}
	return values, nil
}

func (s *sqliteStore) ListFailureReports(ctx context.Context, limit int) ([]FailureReport, error) {
	query := `
SELECT
	grouped.failure_id,
	grouped.count,
	grouped.last_seen_at,
	COALESCE(latest.evidence_excerpt_json, '')
FROM (
	SELECT failure_id, COUNT(*) AS count, MAX(seen_at) AS last_seen_at
	FROM findings
	WHERE rank = 1
	GROUP BY failure_id
) grouped
JOIN findings latest ON latest.id = (
	SELECT f.id
	FROM findings f
	WHERE f.rank = 1 AND f.failure_id = grouped.failure_id
	ORDER BY f.seen_at DESC, f.run_id DESC, f.id DESC
	LIMIT 1
)
ORDER BY grouped.count DESC, grouped.last_seen_at DESC, grouped.failure_id ASC`
	args := []any{}
	if limit > 0 {
		query += "\nLIMIT ?"
		args = append(args, limit)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failure report: %w", err)
	}
	defer rows.Close()

	var out []FailureReport
	for rows.Next() {
		var (
			item         FailureReport
			evidenceJSON string
		)
		if err := rows.Scan(&item.FailureID, &item.Count, &item.LastSeenAt, &evidenceJSON); err != nil {
			return nil, fmt.Errorf("scan failure report: %w", err)
		}
		item.ExampleEvidence = firstStoredEvidence(evidenceJSON)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate failure report: %w", err)
	}
	return out, nil
}

func (s *sqliteStore) ListSignatures(ctx context.Context, limit int) ([]SignatureSummary, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT
	s.signature_hash,
	s.failure_id,
	COALESCE((
		SELECT f.title
		FROM findings f
		WHERE f.rank = 1 AND f.signature_hash = s.signature_hash
		ORDER BY f.seen_at DESC, f.run_id DESC
		LIMIT 1
	), ''),
	COALESCE((
		SELECT f.category
		FROM findings f
		WHERE f.rank = 1 AND f.signature_hash = s.signature_hash
		ORDER BY f.seen_at DESC, f.run_id DESC
		LIMIT 1
	), ''),
	s.occurrence_count,
	s.first_seen_at,
	s.last_seen_at
FROM signatures s
ORDER BY s.occurrence_count DESC, s.last_seen_at DESC, s.signature_hash ASC
LIMIT ?
`, limit)
	if err != nil {
		return nil, fmt.Errorf("query signatures: %w", err)
	}
	defer rows.Close()

	var out []SignatureSummary
	for rows.Next() {
		var item SignatureSummary
		if err := rows.Scan(
			&item.SignatureHash,
			&item.FailureID,
			&item.Title,
			&item.Category,
			&item.OccurrenceCount,
			&item.FirstSeenAt,
			&item.LastSeenAt,
		); err != nil {
			return nil, fmt.Errorf("scan signature summary: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate signatures: %w", err)
	}
	return out, nil
}

func (s *sqliteStore) GetRecentFindingsBySignature(ctx context.Context, signatureHash string, limit int) ([]FindingSummary, error) {
	signatureHash = strings.TrimSpace(signatureHash)
	if signatureHash == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT run_id, failure_id, title, category, signature_hash, seen_at
FROM findings
WHERE rank = 1 AND signature_hash = ?
ORDER BY seen_at DESC
LIMIT ?
`, signatureHash, limit)
	if err != nil {
		return nil, fmt.Errorf("query findings by signature: %w", err)
	}
	defer rows.Close()
	var findings []FindingSummary
	for rows.Next() {
		var item FindingSummary
		if err := rows.Scan(&item.RunID, &item.FailureID, &item.Title, &item.Category, &item.SignatureHash, &item.SeenAt); err != nil {
			return nil, fmt.Errorf("scan finding by signature: %w", err)
		}
		findings = append(findings, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate findings by signature: %w", err)
	}
	return findings, nil
}

func (s *sqliteStore) ListPlaybookStats(ctx context.Context, limit int) ([]PlaybookStats, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT
	f.failure_id,
	COALESCE(MAX(f.title), ''),
	COALESCE(MAX(f.category), ''),
	COUNT(*) AS selected_count,
	COALESCE(pm.match_count, COUNT(*)) AS matched_count,
	COALESCE(pm.avg_rank, 1.0) AS avg_rank,
	SUM(CASE WHEN COALESCE(s.occurrence_count, 0) > 1 THEN 1 ELSE 0 END) AS recurring_run_count,
	COUNT(DISTINCT CASE WHEN COALESCE(s.occurrence_count, 0) > 1 THEN f.signature_hash END) AS recurring_signatures,
	AVG(f.confidence) AS avg_confidence,
	COALESCE(MAX(f.seen_at), '')
FROM findings f
LEFT JOIN signatures s ON s.signature_hash = f.signature_hash
LEFT JOIN (
	SELECT
		playbook_id,
		COUNT(*) AS match_count,
		AVG(CAST(rank AS REAL)) AS avg_rank
	FROM playbook_matches
	GROUP BY playbook_id
) pm ON pm.playbook_id = f.failure_id
WHERE f.rank = 1
GROUP BY f.failure_id, pm.match_count, pm.avg_rank
ORDER BY selected_count DESC, MAX(f.seen_at) DESC, f.failure_id ASC
LIMIT ?
`, limit)
	if err != nil {
		return nil, fmt.Errorf("query playbook stats: %w", err)
	}
	defer rows.Close()

	var out []PlaybookStats
	for rows.Next() {
		var item PlaybookStats
		if err := rows.Scan(
			&item.FailureID,
			&item.Title,
			&item.Category,
			&item.SelectedCount,
			&item.MatchedCount,
			&item.AvgRank,
			&item.RecurringRunCount,
			&item.RecurringSignatures,
			&item.AvgConfidence,
			&item.LastSeenAt,
		); err != nil {
			return nil, fmt.Errorf("scan playbook stats: %w", err)
		}
		if item.MatchedCount > item.SelectedCount {
			item.NonSelectedCount = item.MatchedCount - item.SelectedCount
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate playbook stats: %w", err)
	}
	return out, nil
}

func (s *sqliteStore) Close() error {
	return s.db.Close()
}

func (s *sqliteStore) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	applied_at TEXT NOT NULL
);
`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	applied := map[int]struct{}{}
	rows, err := s.db.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version ASC`)
	if err != nil {
		return fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("scan schema migration: %w", err)
		}
		applied[version] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate schema migrations: %w", err)
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})
	for _, migration := range migrations {
		if _, ok := applied[migration.version]; ok {
			continue
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", migration.version, err)
		}
		if _, err := tx.ExecContext(ctx, migration.sql); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %d (%s): %w", migration.version, migration.name, err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO schema_migrations(version, name, applied_at) VALUES (?, ?, ?)
`, migration.version, migration.name, time.Now().UTC().Format(time.RFC3339)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %d (%s): %w", migration.version, migration.name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d (%s): %w", migration.version, migration.name, err)
		}
	}
	return nil
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullableBool(value *bool) any {
	if value == nil {
		return nil
	}
	if *value {
		return 1
	}
	return 0
}

func firstEvidenceLines(values []string) []string {
	for _, value := range values {
		normalized := strings.ReplaceAll(value, "\r\n", "\n")
		normalized = strings.ReplaceAll(normalized, "\r", "\n")
		for _, line := range strings.Split(normalized, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				return []string{line}
			}
		}
	}
	return []string{}
}

func firstStoredEvidence(encoded string) string {
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return ""
	}
	var values []string
	if err := json.Unmarshal([]byte(encoded), &values); err == nil {
		for _, value := range values {
			if line := firstEvidenceLines([]string{value}); len(line) > 0 {
				return line[0]
			}
		}
	}
	return ""
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
