package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"faultline/internal/model"
)

// ── nullableBool ──────────────────────────────────────────────────────────────

func TestNullableBoolNilReturnsNil(t *testing.T) {
	got := nullableBool(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestNullableBoolTrueReturnsOne(t *testing.T) {
	v := true
	got := nullableBool(&v)
	if got != 1 {
		t.Errorf("expected 1, got %v", got)
	}
}

func TestNullableBoolFalseReturnsZero(t *testing.T) {
	v := false
	got := nullableBool(&v)
	if got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

// ── firstEvidenceLines ────────────────────────────────────────────────────────

func TestFirstEvidenceLinesReturnsFirstNonEmpty(t *testing.T) {
	got := firstEvidenceLines([]string{"  \n  ", "first line\nsecond line"})
	if len(got) != 1 || got[0] != "first line" {
		t.Errorf("expected [\"first line\"], got %v", got)
	}
}

func TestFirstEvidenceLinesEmptyInput(t *testing.T) {
	got := firstEvidenceLines(nil)
	if len(got) != 0 {
		t.Errorf("expected empty slice for nil input, got %v", got)
	}
}

func TestFirstEvidenceLinesAllEmpty(t *testing.T) {
	got := firstEvidenceLines([]string{"   ", "\n\n\n", "\r\n"})
	if len(got) != 0 {
		t.Errorf("expected empty slice for all-whitespace inputs, got %v", got)
	}
}

func TestFirstEvidenceLinesNormalizesCRLF(t *testing.T) {
	got := firstEvidenceLines([]string{"line1\r\nline2"})
	if len(got) != 1 || got[0] != "line1" {
		t.Errorf("expected [\"line1\"], got %v", got)
	}
}

func TestFirstEvidenceLinesNormalizesCR(t *testing.T) {
	got := firstEvidenceLines([]string{"lineA\rlineB"})
	if len(got) != 1 || got[0] != "lineA" {
		t.Errorf("expected [\"lineA\"], got %v", got)
	}
}

// ── firstStoredEvidence ───────────────────────────────────────────────────────

func TestFirstStoredEvidenceEmpty(t *testing.T) {
	got := firstStoredEvidence("")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestFirstStoredEvidenceWhitespaceOnly(t *testing.T) {
	got := firstStoredEvidence("   ")
	if got != "" {
		t.Errorf("expected empty for whitespace-only, got %q", got)
	}
}

func TestFirstStoredEvidenceValidJSON(t *testing.T) {
	got := firstStoredEvidence(`["pull access denied","other error"]`)
	if got != "pull access denied" {
		t.Errorf("expected %q, got %q", "pull access denied", got)
	}
}

func TestFirstStoredEvidenceInvalidJSON(t *testing.T) {
	got := firstStoredEvidence("not-json")
	if got != "" {
		t.Errorf("expected empty for invalid JSON, got %q", got)
	}
}

func TestFirstStoredEvidenceAllBlankValues(t *testing.T) {
	got := firstStoredEvidence(`["  ","","  "]`)
	if got != "" {
		t.Errorf("expected empty for all-blank JSON values, got %q", got)
	}
}

func TestFirstStoredEvidenceMultilineValue(t *testing.T) {
	got := firstStoredEvidence(`["line1\nline2"]`)
	if got != "line1" {
		t.Errorf("expected first non-empty line, got %q", got)
	}
}

// ── noopStore.ListFailureReports ──────────────────────────────────────────────

func TestNoopStoreListFailureReports(t *testing.T) {
	st := Noop()
	reports, err := st.ListFailureReports(context.Background(), 10)
	if err != nil {
		t.Fatalf("noopStore.ListFailureReports: %v", err)
	}
	if reports != nil {
		t.Errorf("expected nil reports from noop store, got %v", reports)
	}
}

// ── SQLite ListFailureReports ─────────────────────────────────────────────────

func TestSQLiteListFailureReportsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	reports, err := st.ListFailureReports(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListFailureReports empty: %v", err)
	}
	if len(reports) != 0 {
		t.Errorf("expected 0 reports for empty db, got %d", len(reports))
	}
}

func TestSQLiteListFailureReportsAfterRuns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 26, 12, 0, 0, 0, time.UTC)

	sig := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Evidence: []string{"pull access denied for repo"},
	}).Hash

	// Insert two runs with the same failure ID to generate a count.
	for i := range 2 {
		handle, err := st.BeginRun(ctx, BeginRunParams{
			Surface:    "analyze",
			SourceKind: "log",
			Source:     "stdin",
			InputHash:  "report-input-" + string(rune('a'+i)),
			StartedAt:  now.Add(time.Duration(i) * time.Minute),
		})
		if err != nil {
			t.Fatalf("BeginRun: %v", err)
		}
		if err := st.CompleteRun(ctx, handle, CompleteRunParams{
			CompletedAt: now.Add(time.Duration(i) * time.Minute),
			Analysis: &model.Analysis{
				Source:     "stdin",
				InputHash:  "report-input-" + string(rune('a'+i)),
				OutputHash: "report-output",
				Results: []model.Result{{
					Playbook:      model.Playbook{ID: "docker-auth", Title: "Docker Auth", Category: "auth"},
					Detector:      "log",
					Score:         4.5,
					Confidence:    0.9,
					Evidence:      []string{"pull access denied for repo"},
					SignatureHash: sig,
				}},
			},
		}); err != nil {
			t.Fatalf("CompleteRun: %v", err)
		}
	}

	reports, err := st.ListFailureReports(ctx, 10)
	if err != nil {
		t.Fatalf("ListFailureReports: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d: %v", len(reports), reports)
	}
	if reports[0].FailureID != "docker-auth" {
		t.Errorf("FailureID = %q, want docker-auth", reports[0].FailureID)
	}
	if reports[0].Count != 2 {
		t.Errorf("Count = %d, want 2", reports[0].Count)
	}
	if reports[0].ExampleEvidence == "" {
		t.Error("expected non-empty ExampleEvidence")
	}
}

func TestSQLiteListFailureReportsWithZeroLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	// Zero limit means no LIMIT clause; should return all (in this case zero).
	reports, err := st.ListFailureReports(context.Background(), 0)
	if err != nil {
		t.Fatalf("ListFailureReports with limit=0: %v", err)
	}
	if reports != nil && len(reports) != 0 {
		t.Errorf("expected empty reports for empty db, got %d", len(reports))
	}
}

func TestSQLiteListFailureReportsWithLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 27, 8, 0, 0, 0, time.UTC)

	// Insert 3 distinct failures.
	failures := []string{"docker-auth", "missing-exec", "network-timeout"}
	for i, id := range failures {
		sig := SignatureForResult(model.Result{
			Playbook: model.Playbook{ID: id},
			Evidence: []string{"evidence for " + id},
		}).Hash
		handle, err := st.BeginRun(ctx, BeginRunParams{
			Surface:    "analyze",
			SourceKind: "log",
			Source:     "stdin",
			InputHash:  id + "-hash",
			StartedAt:  now.Add(time.Duration(i) * time.Minute),
		})
		if err != nil {
			t.Fatalf("BeginRun %s: %v", id, err)
		}
		if err := st.CompleteRun(ctx, handle, CompleteRunParams{
			CompletedAt: now.Add(time.Duration(i) * time.Minute),
			Analysis: &model.Analysis{
				Source:     "stdin",
				InputHash:  id + "-hash",
				OutputHash: id + "-out",
				Results: []model.Result{{
					Playbook:      model.Playbook{ID: id, Title: id, Category: "test"},
					Detector:      "log",
					Score:         3.0,
					Confidence:    0.80,
					Evidence:      []string{"evidence for " + id},
					SignatureHash: sig,
				}},
			},
		}); err != nil {
			t.Fatalf("CompleteRun %s: %v", id, err)
		}
	}

	// Limit to 2 – should get exactly 2 reports back.
	reports, err := st.ListFailureReports(ctx, 2)
	if err != nil {
		t.Fatalf("ListFailureReports with limit=2: %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports (limit), got %d: %v", len(reports), reports)
	}
}
