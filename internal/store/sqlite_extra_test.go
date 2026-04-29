package store

import (
	"context"
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

// ── RecordWorkflowExecution / GetWorkflowExecution / ListWorkflowExecutions ───

func TestRecordWorkflowExecutionNilRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	got, err := st.RecordWorkflowExecution(context.Background(), nil)
	if err != nil {
		t.Fatalf("RecordWorkflowExecution(nil): %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for nil record, got %#v", got)
	}
}

func TestGetWorkflowExecutionBlankID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	got, err := st.GetWorkflowExecution(context.Background(), "   ")
	if err != nil {
		t.Fatalf("GetWorkflowExecution blank: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for blank execution ID, got %#v", got)
	}
}

func TestGetWorkflowExecutionNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	got, err := st.GetWorkflowExecution(context.Background(), "wf-nonexistent")
	if err != nil {
		t.Fatalf("GetWorkflowExecution not-found: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for non-existent execution, got %#v", got)
	}
}

func TestWorkflowExecutionRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 27, 11, 0, 0, 0, time.UTC)

	record, err := st.RecordWorkflowExecution(ctx, &model.WorkflowExecutionRecord{
		SchemaVersion:      "workflow_execution.v1",
		WorkflowID:         "missing-exec.install",
		Title:              "Install missing executable",
		Mode:               model.WorkflowExecutionModeApply,
		SourceFingerprint:  "fp-roundtrip",
		SourceFailureID:    "missing-exec",
		StartedAt:          now.Format(time.RFC3339),
		FinishedAt:         now.Add(time.Second).Format(time.RFC3339),
		ResolvedInputs:     map[string]string{"binary": "node"},
		VerificationStatus: model.WorkflowVerificationStatusPassed,
		Status:             model.WorkflowExecutionStatusSucceeded,
	})
	if err != nil {
		t.Fatalf("RecordWorkflowExecution: %v", err)
	}
	if record == nil || record.ExecutionID == "" {
		t.Fatalf("expected non-nil record with execution ID, got %#v", record)
	}

	// GetWorkflowExecution retrieves the persisted record.
	loaded, err := st.GetWorkflowExecution(ctx, record.ExecutionID)
	if err != nil {
		t.Fatalf("GetWorkflowExecution: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil loaded record")
	}
	if loaded.WorkflowID != "missing-exec.install" {
		t.Errorf("WorkflowID = %q, want %q", loaded.WorkflowID, "missing-exec.install")
	}
	if loaded.Status != model.WorkflowExecutionStatusSucceeded {
		t.Errorf("Status = %q, want succeeded", loaded.Status)
	}
}

func TestListWorkflowExecutionsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	items, err := st.ListWorkflowExecutions(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListWorkflowExecutions empty: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for empty db, got %d", len(items))
	}
}

func TestListWorkflowExecutionsDefaultLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	// Zero limit should default to 20 – should not error.
	_, err = st.ListWorkflowExecutions(context.Background(), 0)
	if err != nil {
		t.Fatalf("ListWorkflowExecutions with limit=0: %v", err)
	}
}

func TestListWorkflowExecutionsAfterRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)

	record, err := st.RecordWorkflowExecution(ctx, &model.WorkflowExecutionRecord{
		WorkflowID:         "test-workflow",
		Title:              "Test workflow",
		Mode:               model.WorkflowExecutionModeApply,
		SourceFingerprint:  "fp-list",
		SourceFailureID:    "test-failure",
		StartedAt:          now.Format(time.RFC3339),
		FinishedAt:         now.Add(time.Second).Format(time.RFC3339),
		VerificationStatus: model.WorkflowVerificationStatusPassed,
		Status:             model.WorkflowExecutionStatusSucceeded,
	})
	if err != nil {
		t.Fatalf("RecordWorkflowExecution: %v", err)
	}

	items, err := st.ListWorkflowExecutions(ctx, 10)
	if err != nil {
		t.Fatalf("ListWorkflowExecutions: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ExecutionID != record.ExecutionID {
		t.Errorf("ExecutionID = %q, want %q", items[0].ExecutionID, record.ExecutionID)
	}
	if items[0].WorkflowID != "test-workflow" {
		t.Errorf("WorkflowID = %q, want test-workflow", items[0].WorkflowID)
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

// ── ListHookStats default limit ───────────────────────────────────────────────

func TestListHookStatsDefaultLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	// Zero limit should default to 10 – should not error.
	_, err = st.ListHookStats(context.Background(), 0)
	if err != nil {
		t.Fatalf("ListHookStats with limit=0: %v", err)
	}
}

// ── VerifyDeterminismForInputHash blank input ─────────────────────────────────

func TestVerifyDeterminismForInputHashBlank(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	summary, err := st.VerifyDeterminismForInputHash(context.Background(), "   ")
	if err != nil {
		t.Fatalf("VerifyDeterminismForInputHash with blank: %v", err)
	}
	if summary.RunCount != 0 {
		t.Errorf("expected RunCount=0 for blank hash, got %d", summary.RunCount)
	}
}
