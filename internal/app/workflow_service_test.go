package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/store"
	workflowexec "faultline/internal/workflow/execute"
)

// ── helpers ───────────────────────────────────────────────────────────────────

const (
	// workflowTestMarker is a distinctive log line that ONLY matches our
	// test playbook so no bundled playbook interferes.
	workflowTestMarker = "FAULTLINE_TEST_WF_MARKER_UNIQUE_9F3A"

	// workflowTestRef is the ID of the noop workflow used by tests.
	workflowTestRef = "test-wf"
)

// workflowTestDirs creates:
//   - wfDir: temp dir with a noop workflow YAML (test-wf.yaml)
//   - pbDir: temp dir with a minimal playbook that matches workflowTestMarker
//     and references workflowTestRef as its remediation workflow
//
// The caller must set:
//
// t.Setenv("FAULTLINE_WORKFLOW_DIR", wfDir)
// opts.PlaybookDir = pbDir
func workflowTestDirs(t *testing.T) (wfDir, pbDir string) {
	t.Helper()

	wfDir = t.TempDir()
	wfContent := strings.Join([]string{
		"schema_version: workflow.v1",
		"workflow_id: " + workflowTestRef,
		"title: Test Workflow",
		"description: A test workflow with no side effects.",
		"steps:",
		"  - id: step1",
		"    type: noop",
		"    args: {}",
		"verification:",
		"  - id: verify1",
		"    type: noop",
		"    args: {}",
		"    expect: {}",
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(wfDir, workflowTestRef+".yaml"), []byte(wfContent), 0o644); err != nil {
		t.Fatalf("write test workflow: %v", err)
	}

	pbDir = t.TempDir()
	pbContent := strings.Join([]string{
		"id: test-wf-playbook",
		"title: Test Workflow Playbook",
		"category: test",
		"severity: high",
		"base_score: 1.0",
		"match:",
		"  any:",
		"    - " + workflowTestMarker,
		"summary: Test-only playbook for unit tests.",
		"diagnosis: Test diagnosis.",
		"fix: Test fix.",
		"validation: Test validation.",
		"remediation:",
		"  workflows:",
		"    - ref: " + workflowTestRef,
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(pbDir, "test-playbook.yaml"), []byte(pbContent), 0o644); err != nil {
		t.Fatalf("write test playbook: %v", err)
	}

	return wfDir, pbDir
}

// workflowTestLog returns a log string that will match the test playbook.
func workflowTestLog() string {
	return "some build output\n" + workflowTestMarker + "\nmore output\n"
}

// workflowNoStoreOpts returns base options configured for workflow tests with store disabled.
func workflowNoHistoryOpts() AnalyzeOptions {
	return AnalyzeOptions{
		OutputOptions: OutputOptions{Top: 1},
		Store:         "off",
	}
}

// ── WorkflowExplain ───────────────────────────────────────────────────────────

func TestWorkflowExplainTextFromLog(t *testing.T) {
	svc := NewService()
	wfDir, pbDir := workflowTestDirs(t)
	t.Setenv("FAULTLINE_WORKFLOW_DIR", wfDir)

	opts := workflowNoHistoryOpts()
	opts.PlaybookDir = pbDir

	var buf bytes.Buffer
	err := svc.WorkflowExplain(strings.NewReader(workflowTestLog()), "stdin", opts, "", false, &buf)
	if err != nil {
		t.Fatalf("WorkflowExplain text: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty text output")
	}
}

func TestWorkflowExplainJSONFromLog(t *testing.T) {
	svc := NewService()
	wfDir, pbDir := workflowTestDirs(t)
	t.Setenv("FAULTLINE_WORKFLOW_DIR", wfDir)

	opts := workflowNoHistoryOpts()
	opts.PlaybookDir = pbDir

	var buf bytes.Buffer
	err := svc.WorkflowExplain(strings.NewReader(workflowTestLog()), "stdin", opts, "", true, &buf)
	if err != nil {
		t.Fatalf("WorkflowExplain JSON: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &payload); err != nil {
		t.Fatalf("unmarshal WorkflowExplain JSON: %v", err)
	}
	if payload["workflow_id"] != workflowTestRef {
		t.Errorf("expected workflow_id %q, got %v", workflowTestRef, payload["workflow_id"])
	}
	if payload["mode"] != "explain" {
		t.Errorf("expected mode 'explain', got %v", payload["mode"])
	}
}

func TestWorkflowExplainWithWorkflowRefOverride(t *testing.T) {
	svc := NewService()
	wfDir, pbDir := workflowTestDirs(t)
	t.Setenv("FAULTLINE_WORKFLOW_DIR", wfDir)

	opts := workflowNoHistoryOpts()
	opts.PlaybookDir = pbDir

	var buf bytes.Buffer
	err := svc.WorkflowExplain(strings.NewReader(workflowTestLog()), "stdin", opts, workflowTestRef, true, &buf)
	if err != nil {
		t.Fatalf("WorkflowExplain with ref override: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["workflow_id"] != workflowTestRef {
		t.Errorf("expected workflow_id %q, got %v", workflowTestRef, payload["workflow_id"])
	}
}

func TestWorkflowExplainUnknownRefReturnsError(t *testing.T) {
	svc := NewService()
	wfDir, pbDir := workflowTestDirs(t)
	t.Setenv("FAULTLINE_WORKFLOW_DIR", wfDir)

	opts := workflowNoHistoryOpts()
	opts.PlaybookDir = pbDir

	var buf bytes.Buffer
	// The playbook recommends workflowTestRef; requesting a different ref should error.
	err := svc.WorkflowExplain(strings.NewReader(workflowTestLog()), "stdin", opts, "nonexistent-workflow-xyz", false, &buf)
	if err == nil {
		t.Fatal("expected error for unrecognised workflow ref")
	}
}

// ── WorkflowApply (dry-run) ───────────────────────────────────────────────────

func TestWorkflowApplyDryRunText(t *testing.T) {
	svc := NewService()
	wfDir, pbDir := workflowTestDirs(t)
	t.Setenv("FAULTLINE_WORKFLOW_DIR", wfDir)

	opts := workflowNoHistoryOpts()
	opts.PlaybookDir = pbDir

	var buf bytes.Buffer
	err := svc.WorkflowApply(
		strings.NewReader(workflowTestLog()), "stdin",
		opts, "", true, workflowexec.Policy{}, false, &buf,
	)
	if err != nil {
		t.Fatalf("WorkflowApply dry-run text: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty dry-run text output")
	}
}

func TestWorkflowApplyDryRunJSON(t *testing.T) {
	svc := NewService()
	wfDir, pbDir := workflowTestDirs(t)
	t.Setenv("FAULTLINE_WORKFLOW_DIR", wfDir)

	opts := workflowNoHistoryOpts()
	opts.PlaybookDir = pbDir

	var buf bytes.Buffer
	err := svc.WorkflowApply(
		strings.NewReader(workflowTestLog()), "stdin",
		opts, "", true, workflowexec.Policy{}, true, &buf,
	)
	if err != nil {
		t.Fatalf("WorkflowApply dry-run JSON: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &payload); err != nil {
		t.Fatalf("unmarshal dry-run JSON: %v", err)
	}
	if payload["mode"] != "dry-run" {
		t.Errorf("expected mode 'dry-run', got %v", payload["mode"])
	}
}

// ── WorkflowHistory ───────────────────────────────────────────────────────────

func TestWorkflowHistoryTextEmptyStore(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer

	err := svc.WorkflowHistory(workflowNoHistoryOpts(), 10, false, &buf)
	if err != nil {
		t.Fatalf("WorkflowHistory text: %v", err)
	}
	if !strings.Contains(buf.String(), "No workflow executions recorded") {
		t.Errorf("expected empty-history message, got %q", buf.String())
	}
}

func TestWorkflowHistoryJSONEmptyStore(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer

	err := svc.WorkflowHistory(workflowNoHistoryOpts(), 10, true, &buf)
	if err != nil {
		t.Fatalf("WorkflowHistory JSON: %v", err)
	}
	trimmed := strings.TrimSpace(buf.String())
	// noop store returns empty slice; JSON must be a valid array.
	if trimmed != "[]" && trimmed != "null" {
		var arr []interface{}
		if err := json.Unmarshal([]byte(trimmed), &arr); err != nil {
			t.Fatalf("WorkflowHistory JSON not a valid array: %v — got %q", err, trimmed)
		}
	}
}

// ── WorkflowShow ─────────────────────────────────────────────────────────────

func TestWorkflowShowNotFoundReturnsError(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer

	err := svc.WorkflowShow("does-not-exist", workflowNoHistoryOpts(), false, &buf)
	if err == nil {
		t.Fatal("expected error for unknown execution ID")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got %v", err)
	}
}

// ── loadWorkflowAnalysis ──────────────────────────────────────────────────────

func TestLoadWorkflowAnalysisFromLog(t *testing.T) {
	opts := workflowNoHistoryOpts()
	opts.PlaybookDir = repoPlaybookDir()
	logInput := "pull access denied\nError response from daemon: authentication required\n"

	a, err := loadWorkflowAnalysis(strings.NewReader(logInput), "stdin", opts)
	if err != nil {
		t.Fatalf("loadWorkflowAnalysis from log: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil analysis")
	}
	if len(a.Results) == 0 {
		t.Error("expected at least one matched result from docker-auth log")
	}
}

func TestLoadWorkflowAnalysisFromArtifactJSON(t *testing.T) {
	// Build a real analysis via log analysis, serialize to JSON, then verify
	// loadWorkflowAnalysis can round-trip it through the JSON parsing branch.
	_, pbDir := workflowTestDirs(t)
	opts := workflowNoHistoryOpts()
	opts.PlaybookDir = pbDir

	// First pass: produce a real analysis via log analysis.
	a, err := loadWorkflowAnalysis(strings.NewReader(workflowTestLog()), "stdin", opts)
	if err != nil {
		t.Fatalf("first loadWorkflowAnalysis: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil analysis on first pass")
	}

	// Serialize to the stable JSON schema.
	data, err := output.FormatAnalysisJSON(a, 1)
	if err != nil {
		t.Fatalf("serialize analysis: %v", err)
	}

	// Second pass: load from JSON — exercises the JSON-first branch.
	loaded, err := loadWorkflowAnalysis(strings.NewReader(data), "", opts)
	if err != nil {
		t.Fatalf("loadWorkflowAnalysis from artifact JSON: %v", err)
	}
	if loaded == nil || loaded.Artifact == nil {
		t.Fatal("expected loaded analysis with populated artifact")
	}
}

func TestLoadWorkflowAnalysisEmptyInputReturnsError(t *testing.T) {
	opts := workflowNoHistoryOpts()
	opts.PlaybookDir = repoPlaybookDir()

	_, err := loadWorkflowAnalysis(strings.NewReader(""), "stdin", opts)
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

// ── openWorkflowStore ─────────────────────────────────────────────────────────

func TestOpenWorkflowStoreOffReturnsNoop(t *testing.T) {
	opts := workflowNoHistoryOpts() // Store = "off"

	st, info, err := openWorkflowStore(opts)
	if err != nil {
		t.Fatalf("openWorkflowStore: %v", err)
	}
	defer st.Close()
	if info.Mode != store.ModeOff {
		t.Errorf("expected mode 'off' for Store=off, got %q", info.Mode)
	}
}

func TestOpenWorkflowStoreTempDB(t *testing.T) {
	opts := workflowNoHistoryOpts()
	opts.Store = filepath.Join(t.TempDir(), "wf.db")

	st, info, err := openWorkflowStore(opts)
	if err != nil {
		t.Fatalf("openWorkflowStore with temp db: %v", err)
	}
	defer st.Close()
	if info.Mode == store.ModeOff {
		t.Error("expected a real store mode for explicit path, got 'off'")
	}
}

// ── persistWorkflowExecution ─────────────────────────────────────────────────

func TestPersistWorkflowExecutionNilRecordIsNoop(t *testing.T) {
	opts := workflowNoHistoryOpts()

	result, err := persistWorkflowExecution(nil, opts)
	if err != nil {
		t.Fatalf("persistWorkflowExecution(nil): %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for nil input, got %v", result)
	}
}

func TestPersistWorkflowExecutionStoresRecord(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "wf.db")
	opts := AnalyzeOptions{Store: dbPath}

	record := &model.WorkflowExecutionRecord{
		SchemaVersion:      "workflow_execution.v1",
		WorkflowID:         workflowTestRef,
		Title:              "Test workflow",
		Mode:               model.WorkflowExecutionModeApply,
		Status:             model.WorkflowExecutionStatusSucceeded,
		VerificationStatus: model.WorkflowVerificationStatusPassed,
		StartedAt:          "2024-01-01T00:00:00Z",
		FinishedAt:         "2024-01-01T00:00:01Z",
	}

	got, err := persistWorkflowExecution(record, opts)
	if err != nil {
		t.Fatalf("persistWorkflowExecution: %v", err)
	}
	if got == nil || got.ExecutionID == "" {
		t.Errorf("expected record with execution ID, got %v", got)
	}
}

// ── WorkflowShow with real store ──────────────────────────────────────────────

func persistTestRecord(t *testing.T, dbPath string) string {
	t.Helper()
	st, _, err := store.OpenBestEffort(store.Config{Mode: store.ModeAuto, Path: dbPath})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()
	record, err := st.RecordWorkflowExecution(context.Background(), &model.WorkflowExecutionRecord{
		SchemaVersion:      "workflow_execution.v1",
		WorkflowID:         workflowTestRef,
		Title:              "Test workflow",
		Mode:               model.WorkflowExecutionModeApply,
		Status:             model.WorkflowExecutionStatusSucceeded,
		VerificationStatus: model.WorkflowVerificationStatusPassed,
		StartedAt:          "2024-01-01T00:00:00Z",
		FinishedAt:         "2024-01-01T00:00:01Z",
	})
	if err != nil {
		t.Fatalf("RecordWorkflowExecution: %v", err)
	}
	return record.ExecutionID
}

func TestWorkflowShowTextFromStore(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "wf.db")
	executionID := persistTestRecord(t, dbPath)

	svc := NewService()
	var buf bytes.Buffer
	err := svc.WorkflowShow(executionID, AnalyzeOptions{Store: dbPath}, false, &buf)
	if err != nil {
		t.Fatalf("WorkflowShow text: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty text output")
	}
}

func TestWorkflowShowJSONFromStore(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "wf.db")
	executionID := persistTestRecord(t, dbPath)

	svc := NewService()
	var buf bytes.Buffer
	err := svc.WorkflowShow(executionID, AnalyzeOptions{Store: dbPath}, true, &buf)
	if err != nil {
		t.Fatalf("WorkflowShow JSON: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &payload); err != nil {
		t.Fatalf("WorkflowShow JSON unmarshal: %v — got %q", err, buf.String())
	}
	if payload["workflow_id"] != workflowTestRef {
		t.Errorf("unexpected workflow_id: %v", payload["workflow_id"])
	}
}

// ── WorkflowApply live (non-dry-run) ─────────────────────────────────────────

func TestWorkflowApplyLiveNoopText(t *testing.T) {
	wfDir, pbDir := workflowTestDirs(t)
	t.Setenv("FAULTLINE_WORKFLOW_DIR", wfDir)
	opts := workflowNoHistoryOpts()
	opts.PlaybookDir = pbDir

	svc := NewService()
	var buf bytes.Buffer
	err := svc.WorkflowApply(strings.NewReader(workflowTestLog()), "stdin", opts, workflowTestRef, false, workflowexec.Policy{}, false, &buf)
	if err != nil {
		t.Fatalf("WorkflowApply live text: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty text output")
	}
}

func TestWorkflowApplyLiveNoopJSON(t *testing.T) {
	wfDir, pbDir := workflowTestDirs(t)
	t.Setenv("FAULTLINE_WORKFLOW_DIR", wfDir)
	opts := workflowNoHistoryOpts()
	opts.PlaybookDir = pbDir

	svc := NewService()
	var buf bytes.Buffer
	err := svc.WorkflowApply(strings.NewReader(workflowTestLog()), "stdin", opts, workflowTestRef, false, workflowexec.Policy{}, true, &buf)
	if err != nil {
		t.Fatalf("WorkflowApply live JSON: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &payload); err != nil {
		t.Fatalf("WorkflowApply live JSON unmarshal: %v — got %q", err, buf.String())
	}
	if payload["mode"] != string(model.WorkflowExecutionModeApply) {
		t.Errorf("expected mode %q, got %v", model.WorkflowExecutionModeApply, payload["mode"])
	}
}
