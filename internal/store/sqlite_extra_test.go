package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"faultline/internal/model"
)

// ── LookupSignatureHistory blank hash ─────────────────────────────────────────

func TestLookupSignatureHistoryBlankHash(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	history, err := st.LookupSignatureHistory(context.Background(), "   ")
	if err != nil {
		t.Fatalf("LookupSignatureHistory with blank hash: %v", err)
	}
	if history.SeenBefore {
		t.Error("expected SeenBefore=false for blank hash")
	}
}

func TestLookupSignatureHistoryNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	history, err := st.LookupSignatureHistory(context.Background(), "deadbeefdeadbeef")
	if err != nil {
		t.Fatalf("LookupSignatureHistory for unknown hash: %v", err)
	}
	if history.SeenBefore {
		t.Error("expected SeenBefore=false for unknown signature")
	}
	if history.OccurrenceCount != 0 {
		t.Errorf("expected OccurrenceCount=0, got %d", history.OccurrenceCount)
	}
}

// ── CountSeenFailure blank id ─────────────────────────────────────────────────

func TestCountSeenFailureBlankID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	count, err := st.CountSeenFailure(context.Background(), "   ")
	if err != nil {
		t.Fatalf("CountSeenFailure with blank id: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for blank id, got %d", count)
	}
}

// ── ListSignatures ────────────────────────────────────────────────────────────

func TestListSignaturesEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	sigs, err := st.ListSignatures(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListSignatures empty: %v", err)
	}
	if len(sigs) != 0 {
		t.Errorf("expected 0 signatures for empty db, got %d", len(sigs))
	}
}

func TestListSignaturesDefaultLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	// Zero limit should default to 10 – should not error.
	_, err = st.ListSignatures(context.Background(), 0)
	if err != nil {
		t.Fatalf("ListSignatures with limit=0: %v", err)
	}
}

func TestListSignaturesAfterRuns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)

	sig := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Evidence: []string{"pull access denied for registry.example.com"},
	}).Hash

	handle, err := st.BeginRun(ctx, BeginRunParams{
		Surface:    "analyze",
		SourceKind: "log",
		Source:     "stdin",
		InputHash:  "list-sig-input",
		StartedAt:  now,
	})
	if err != nil {
		t.Fatalf("BeginRun: %v", err)
	}
	if err := st.CompleteRun(ctx, handle, CompleteRunParams{
		CompletedAt: now,
		Analysis: &model.Analysis{
			Source:     "stdin",
			InputHash:  "list-sig-input",
			OutputHash: "list-sig-output",
			Results: []model.Result{{
				Playbook:      model.Playbook{ID: "docker-auth", Title: "Docker Auth", Category: "auth"},
				Detector:      "log",
				Score:         4.0,
				Confidence:    0.85,
				Evidence:      []string{"pull access denied for registry.example.com"},
				SignatureHash: sig,
			}},
		},
	}); err != nil {
		t.Fatalf("CompleteRun: %v", err)
	}

	sigs, err := st.ListSignatures(ctx, 10)
	if err != nil {
		t.Fatalf("ListSignatures: %v", err)
	}
	if len(sigs) != 1 {
		t.Fatalf("expected 1 signature, got %d: %v", len(sigs), sigs)
	}
	if sigs[0].FailureID != "docker-auth" {
		t.Errorf("FailureID = %q, want docker-auth", sigs[0].FailureID)
	}
	if sigs[0].OccurrenceCount != 1 {
		t.Errorf("OccurrenceCount = %d, want 1", sigs[0].OccurrenceCount)
	}
	if sigs[0].Title != "Docker Auth" {
		t.Errorf("Title = %q, want %q", sigs[0].Title, "Docker Auth")
	}
}

// ── ListPlaybookStats ─────────────────────────────────────────────────────────

func TestListPlaybookStatsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	stats, err := st.ListPlaybookStats(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListPlaybookStats empty: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 stats for empty db, got %d", len(stats))
	}
}

func TestListPlaybookStatsDefaultLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	// Zero limit should default to 10 – should not error.
	_, err = st.ListPlaybookStats(context.Background(), 0)
	if err != nil {
		t.Fatalf("ListPlaybookStats with limit=0: %v", err)
	}
}

func TestListPlaybookStatsAfterRuns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC)

	sig := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "missing-exec"},
		Evidence: []string{"executable file not found"},
	}).Hash

	handle, err := st.BeginRun(ctx, BeginRunParams{
		Surface:    "analyze",
		SourceKind: "log",
		Source:     "stdin",
		InputHash:  "pb-stats-input",
		StartedAt:  now,
	})
	if err != nil {
		t.Fatalf("BeginRun: %v", err)
	}
	if err := st.CompleteRun(ctx, handle, CompleteRunParams{
		CompletedAt: now,
		Analysis: &model.Analysis{
			Source:     "stdin",
			InputHash:  "pb-stats-input",
			OutputHash: "pb-stats-output",
			Results: []model.Result{{
				Playbook:      model.Playbook{ID: "missing-exec", Title: "Missing executable", Category: "runtime"},
				Detector:      "log",
				Score:         3.5,
				Confidence:    0.78,
				Evidence:      []string{"executable file not found"},
				SignatureHash: sig,
			}},
		},
	}); err != nil {
		t.Fatalf("CompleteRun: %v", err)
	}

	stats, err := st.ListPlaybookStats(ctx, 10)
	if err != nil {
		t.Fatalf("ListPlaybookStats: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 playbook stat, got %d: %v", len(stats), stats)
	}
	if stats[0].FailureID != "missing-exec" {
		t.Errorf("FailureID = %q, want missing-exec", stats[0].FailureID)
	}
	if stats[0].SelectedCount != 1 {
		t.Errorf("SelectedCount = %d, want 1", stats[0].SelectedCount)
	}
}

// ── CompleteRun with zero handle ─────────────────────────────────────────────

func TestCompleteRunZeroHandleIsNoop(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	// RunHandle with zero ID should be a no-op.
	if err := st.CompleteRun(context.Background(), RunHandle{ID: 0}, CompleteRunParams{
		Analysis: &model.Analysis{Source: "stdin"},
	}); err != nil {
		t.Fatalf("CompleteRun with zero handle: %v", err)
	}
}

func TestCompleteRunNilAnalysisIsNoop(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	handle, err := st.BeginRun(ctx, BeginRunParams{
		Surface:   "analyze",
		InputHash: "nil-analysis-input",
		StartedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("BeginRun: %v", err)
	}

	// Nil analysis should be a no-op.
	if err := st.CompleteRun(ctx, handle, CompleteRunParams{Analysis: nil}); err != nil {
		t.Fatalf("CompleteRun with nil analysis: %v", err)
	}
}

// ── BeginRun with zero timestamp ──────────────────────────────────────────────

func TestBeginRunWithZeroTimestampUsesNow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	// StartedAt zero → should be auto-populated by BeginRun.
	ctx := context.Background()
	handle, err := st.BeginRun(ctx, BeginRunParams{
		Surface:    "analyze",
		SourceKind: "log",
		Source:     "stdin",
		InputHash:  "zero-ts-input",
		// StartedAt is zero value
	})
	if err != nil {
		t.Fatalf("BeginRun with zero timestamp: %v", err)
	}
	if handle.ID == 0 {
		t.Error("expected non-zero run handle ID")
	}
}

// ── openSQLite creates the directory if it doesn't exist ─────────────────────

func TestOpenSQLiteCreatesParentDirectory(t *testing.T) {
	base := t.TempDir()
	// Nested path that doesn't exist yet.
	path := filepath.Join(base, "sub", "dir", "faultline.db")
	st, err := openSQLite(path)
	if err != nil {
		t.Fatalf("openSQLite with missing parent dir: %v", err)
	}
	defer st.Close()

	// Verify the file was created.
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected db file to exist at %s: %v", path, err)
	}
}

// ── migrate is idempotent ─────────────────────────────────────────────────────

func TestMigrateIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	// Open twice — second open will call migrate again on an already-migrated db.
	st1, err := openSQLite(path)
	if err != nil {
		t.Fatalf("first openSQLite: %v", err)
	}
	st1.Close()

	st2, err := openSQLite(path)
	if err != nil {
		t.Fatalf("second openSQLite (idempotent migrate): %v", err)
	}
	defer st2.Close()
}
