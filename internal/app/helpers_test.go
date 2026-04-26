package app

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/store"
)

// ── writeStoreInfoText ────────────────────────────────────────────────────────

func TestWriteStoreInfoTextDisabled(t *testing.T) {
	var b strings.Builder
	writeStoreInfoText(&b, store.Info{})
	if !strings.Contains(b.String(), "disabled") {
		t.Errorf("expected 'disabled' in output, got %q", b.String())
	}
}

func TestWriteStoreInfoTextBackendWithPath(t *testing.T) {
	var b strings.Builder
	writeStoreInfoText(&b, store.Info{Backend: "sqlite", Path: "/tmp/test.db"})
	got := b.String()
	if !strings.Contains(got, "sqlite") {
		t.Errorf("expected backend name 'sqlite', got %q", got)
	}
	if !strings.Contains(got, "/tmp/test.db") {
		t.Errorf("expected path '/tmp/test.db', got %q", got)
	}
}

func TestWriteStoreInfoTextBackendWithoutPath(t *testing.T) {
	var b strings.Builder
	writeStoreInfoText(&b, store.Info{Backend: "sqlite"})
	got := b.String()
	if !strings.Contains(got, "sqlite") {
		t.Errorf("expected backend name 'sqlite', got %q", got)
	}
}

func TestWriteStoreInfoTextEmptyBackendFallsBackToStoreLabel(t *testing.T) {
	// When Backend is empty but Path is set, the label should default to "store".
	var b strings.Builder
	writeStoreInfoText(&b, store.Info{Path: "/tmp/test.db"})
	got := b.String()
	if !strings.Contains(got, "store") {
		t.Errorf("expected 'store' label, got %q", got)
	}
	if !strings.Contains(got, "/tmp/test.db") {
		t.Errorf("expected path, got %q", got)
	}
}

func TestWriteStoreInfoTextDegradedShowsWarning(t *testing.T) {
	var b strings.Builder
	writeStoreInfoText(&b, store.Info{
		Backend:  "sqlite",
		Path:     "/tmp/test.db",
		Degraded: true,
		Warning:  "file is corrupt",
	})
	got := b.String()
	if !strings.Contains(got, "file is corrupt") {
		t.Errorf("expected warning message in output, got %q", got)
	}
}

// A degraded store without a warning message should not emit "Warning:".
func TestWriteStoreInfoTextDegradedWithoutWarningNoWarningLine(t *testing.T) {
	var b strings.Builder
	writeStoreInfoText(&b, store.Info{Backend: "sqlite", Degraded: true, Warning: ""})
	if strings.Contains(b.String(), "Warning:") {
		t.Errorf("unexpected Warning: line when warning is empty, got %q", b.String())
	}
}

// ── historyWindow ─────────────────────────────────────────────────────────────

func TestHistoryWindowInvalidFirstSeenAt(t *testing.T) {
	if got := historyWindow("not-a-date", "2026-04-22T10:00:00Z"); got != "" {
		t.Errorf("expected empty for invalid firstSeenAt, got %q", got)
	}
}

func TestHistoryWindowInvalidLastSeenAt(t *testing.T) {
	if got := historyWindow("2026-04-22T10:00:00Z", "not-a-date"); got != "" {
		t.Errorf("expected empty for invalid lastSeenAt, got %q", got)
	}
}

func TestHistoryWindowEndBeforeStart(t *testing.T) {
	if got := historyWindow("2026-04-22T12:00:00Z", "2026-04-22T10:00:00Z"); got != "" {
		t.Errorf("expected empty when end is before start, got %q", got)
	}
}

func TestHistoryWindowDays(t *testing.T) {
	got := historyWindow("2026-04-20T10:00:00Z", "2026-04-23T10:00:00Z")
	if got != "3d" {
		t.Errorf("expected '3d', got %q", got)
	}
}

func TestHistoryWindowHours(t *testing.T) {
	got := historyWindow("2026-04-22T10:00:00Z", "2026-04-22T13:00:00Z")
	if got != "3h" {
		t.Errorf("expected '3h', got %q", got)
	}
}

func TestHistoryWindowMinutes(t *testing.T) {
	got := historyWindow("2026-04-22T10:00:00Z", "2026-04-22T10:30:00Z")
	if got != "30m" {
		t.Errorf("expected '30m', got %q", got)
	}
}

func TestHistoryWindowSubMinuteReturnsEmpty(t *testing.T) {
	got := historyWindow("2026-04-22T10:00:00Z", "2026-04-22T10:00:30Z")
	if got != "" {
		t.Errorf("expected empty for sub-minute duration, got %q", got)
	}
}

// ── guardFindings ─────────────────────────────────────────────────────────────

func TestGuardFindingsNilAnalysisReturnsEmptyResults(t *testing.T) {
	result := guardFindings(nil, 5)
	if result == nil {
		t.Fatal("expected non-nil analysis from nil input")
	}
	if len(result.Results) != 0 {
		t.Errorf("expected 0 results for nil analysis, got %d", len(result.Results))
	}
}

func TestGuardFindingsFiltersLowConfidence(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{Confidence: 0.50, Score: 4.0},
			{Confidence: 0.80, Score: 4.0},
		},
	}
	result := guardFindings(a, 10)
	if len(result.Results) != 1 {
		t.Errorf("expected 1 result after confidence filtering, got %d", len(result.Results))
	}
}

func TestGuardFindingsFiltersLowScore(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{Confidence: 0.80, Score: 2.0},
			{Confidence: 0.80, Score: 4.0},
		},
	}
	result := guardFindings(a, 10)
	if len(result.Results) != 1 {
		t.Errorf("expected 1 result after score filtering, got %d", len(result.Results))
	}
}

func TestGuardFindingsRespectTopLimit(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{Confidence: 0.80, Score: 4.0},
			{Confidence: 0.80, Score: 4.0},
			{Confidence: 0.80, Score: 4.0},
		},
	}
	result := guardFindings(a, 2)
	if len(result.Results) != 2 {
		t.Errorf("expected 2 results for top=2, got %d", len(result.Results))
	}
}

func TestGuardFindingsZeroTopNoLimit(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{Confidence: 0.80, Score: 4.0},
			{Confidence: 0.80, Score: 4.0},
			{Confidence: 0.80, Score: 4.0},
		},
	}
	result := guardFindings(a, 0)
	if len(result.Results) != 3 {
		t.Errorf("expected 3 results for top=0 (no limit), got %d", len(result.Results))
	}
}

func TestGuardFindingsPreservesMetadata(t *testing.T) {
	a := &model.Analysis{
		Fingerprint: "fp-123",
		Source:      "stdin",
		Results:     []model.Result{{Confidence: 0.80, Score: 4.0}},
	}
	result := guardFindings(a, 10)
	if result.Fingerprint != "fp-123" {
		t.Errorf("expected Fingerprint preserved, got %q", result.Fingerprint)
	}
	if result.Source != "stdin" {
		t.Errorf("expected Source preserved, got %q", result.Source)
	}
}

// ── writeGuardNoFindings ──────────────────────────────────────────────────────

func TestWriteGuardNoFindingsJSONWritesEmptyAnalysis(t *testing.T) {
	opts := AnalyzeOptions{JSON: true, NoHistory: true, PlaybookDir: repoPlaybookDir()}
	var buf bytes.Buffer
	if err := writeGuardNoFindings("/tmp/repo", opts, &buf); err != nil {
		t.Fatalf("writeGuardNoFindings JSON: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty JSON output")
	}
	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
}

func TestWriteGuardNoFindingsJSONFormatWritesEmptyAnalysis(t *testing.T) {
	opts := AnalyzeOptions{
		Format:      output.FormatJSON,
		NoHistory:   true,
		PlaybookDir: repoPlaybookDir(),
	}
	var buf bytes.Buffer
	if err := writeGuardNoFindings("/tmp/repo", opts, &buf); err != nil {
		t.Fatalf("writeGuardNoFindings FormatJSON: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty JSON output for FormatJSON")
	}
}

func TestWriteGuardNoFindingsTextProducesNoOutput(t *testing.T) {
	// For non-JSON format, writeGuardNoFindings should not write anything.
	opts := AnalyzeOptions{NoHistory: true, PlaybookDir: repoPlaybookDir()}
	var buf bytes.Buffer
	if err := writeGuardNoFindings("/tmp/repo", opts, &buf); err != nil {
		t.Fatalf("writeGuardNoFindings text: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for non-JSON format, got %q", buf.String())
	}
}

// ── tracePlaybookID ───────────────────────────────────────────────────────────

func TestTracePlaybookIDFromTracePlaybookOption(t *testing.T) {
	opts := AnalyzeOptions{TracePlaybook: "my-playbook"}
	id, err := tracePlaybookID(nil, opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "my-playbook" {
		t.Errorf("expected 'my-playbook', got %q", id)
	}
}

func TestTracePlaybookIDSelectWithNilAnalysisErrors(t *testing.T) {
	opts := AnalyzeOptions{Select: 1}
	_, err := tracePlaybookID(nil, opts)
	if err == nil {
		t.Fatal("expected error for nil analysis with Select > 0")
	}
}

func TestTracePlaybookIDSelectWithEmptyResultsErrors(t *testing.T) {
	opts := AnalyzeOptions{Select: 1}
	_, err := tracePlaybookID(&model.Analysis{Results: []model.Result{}}, opts)
	if err == nil {
		t.Fatal("expected error for empty results with Select > 0")
	}
}

func TestTracePlaybookIDSelectOutOfRangeErrors(t *testing.T) {
	a := &model.Analysis{Results: []model.Result{
		{Playbook: model.Playbook{ID: "p1"}},
	}}
	opts := AnalyzeOptions{Select: 5}
	_, err := tracePlaybookID(a, opts)
	if err == nil {
		t.Fatal("expected error for out-of-range select")
	}
	if !strings.Contains(err.Error(), "--select 5") {
		t.Errorf("expected --select 5 in error message, got %q", err.Error())
	}
}

func TestTracePlaybookIDSelectValidReturnsID(t *testing.T) {
	a := &model.Analysis{Results: []model.Result{
		{Playbook: model.Playbook{ID: "p1"}},
		{Playbook: model.Playbook{ID: "p2"}},
	}}
	opts := AnalyzeOptions{Select: 2}
	id, err := tracePlaybookID(a, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "p2" {
		t.Errorf("expected 'p2', got %q", id)
	}
}

func TestTracePlaybookIDAutoFirstResult(t *testing.T) {
	a := &model.Analysis{Results: []model.Result{
		{Playbook: model.Playbook{ID: "first"}},
		{Playbook: model.Playbook{ID: "second"}},
	}}
	id, err := tracePlaybookID(a, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "first" {
		t.Errorf("expected 'first', got %q", id)
	}
}

func TestTracePlaybookIDNoResultsReturnsEmpty(t *testing.T) {
	id, err := tracePlaybookID(&model.Analysis{Results: []model.Result{}}, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "" {
		t.Errorf("expected empty string, got %q", id)
	}
}

func TestTracePlaybookIDNilAnalysisReturnsEmpty(t *testing.T) {
	id, err := tracePlaybookID(nil, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "" {
		t.Errorf("expected empty for nil analysis, got %q", id)
	}
}
