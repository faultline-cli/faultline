package app

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"faultline/internal/model"
	"faultline/internal/store"
)

// ── truncateReportCell ────────────────────────────────────────────────────────

func TestTruncateReportCell(t *testing.T) {
	cases := []struct {
		name  string
		value string
		limit int
		want  string
	}{
		{"empty string", "", 10, ""},
		{"within limit", "hello", 10, "hello"},
		{"exactly at limit", "hello", 5, "hello"},
		{"whitespace trimmed before limit check", "  hi  ", 10, "hi"},
		{"over limit trims with ellipsis", "hello world", 8, "hello..."},
		{"over limit pads ellipsis correctly", "abcdefghij", 7, "abcd..."},
		{"limit of 3 returns prefix without ellipsis", "abcdefg", 3, "abc"},
		{"limit of 2 returns prefix without ellipsis", "abcdefg", 2, "ab"},
		{"limit of 1 returns single char", "abc", 1, "a"},
		{"zero limit returns unchanged", "hello", 0, "hello"},
		{"negative limit returns unchanged", "hello", -1, "hello"},
		{"unicode runes count correctly", "hëllo wörld", 8, "hëllo..."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := truncateReportCell(tc.value, tc.limit)
			if got != tc.want {
				t.Errorf("truncateReportCell(%q, %d) = %q, want %q", tc.value, tc.limit, got, tc.want)
			}
		})
	}
}

// ── formatReportText ──────────────────────────────────────────────────────────

func TestFormatReportTextEmpty(t *testing.T) {
	got := formatReportText(nil)
	if !strings.Contains(got, "No stored failures yet.") {
		t.Errorf("expected 'No stored failures yet.' in output for nil input, got %q", got)
	}
	got2 := formatReportText([]store.FailureReport{})
	if !strings.Contains(got2, "No stored failures yet.") {
		t.Errorf("expected 'No stored failures yet.' in output for empty slice, got %q", got2)
	}
}

func TestFormatReportTextWithFailures(t *testing.T) {
	failures := []store.FailureReport{
		{FailureID: "missing-executable", Count: 5, LastSeenAt: "2025-01-01", ExampleEvidence: "exec: foo: not found"},
		{FailureID: "timeout", Count: 12, LastSeenAt: "2025-02-01", ExampleEvidence: "exceeded time limit"},
	}
	got := formatReportText(failures)

	// Must have report header
	if !strings.Contains(got, "Faultline Report") {
		t.Errorf("expected 'Faultline Report' header, got %q", got)
	}
	// Must have column headers
	if !strings.Contains(got, "Failure ID") {
		t.Errorf("expected 'Failure ID' column header, got %q", got)
	}
	if !strings.Contains(got, "Count") {
		t.Errorf("expected 'Count' column header, got %q", got)
	}
	// Must contain both failure IDs
	if !strings.Contains(got, "missing-executable") {
		t.Errorf("expected 'missing-executable' in output, got %q", got)
	}
	if !strings.Contains(got, "timeout") {
		t.Errorf("expected 'timeout' in output, got %q", got)
	}
	// Must show counts
	if !strings.Contains(got, "5") {
		t.Errorf("expected count '5' in output, got %q", got)
	}
	if !strings.Contains(got, "12") {
		t.Errorf("expected count '12' in output, got %q", got)
	}
	// Separator line of dashes must appear
	if !strings.Contains(got, "---") {
		t.Errorf("expected separator line in output, got %q", got)
	}
}

func TestFormatReportTextTruncatesLongEvidence(t *testing.T) {
	longEvidence := strings.Repeat("a", 100)
	failures := []store.FailureReport{
		{FailureID: "test", Count: 1, LastSeenAt: "2025-01-01", ExampleEvidence: longEvidence},
	}
	got := formatReportText(failures)
	if strings.Contains(got, longEvidence) {
		t.Errorf("expected long evidence to be truncated in output")
	}
	if !strings.Contains(got, "...") {
		t.Errorf("expected truncation indicator '...' in output, got %q", got)
	}
}

// ── writeReportRow ────────────────────────────────────────────────────────────

func TestWriteReportRow(t *testing.T) {
	var b strings.Builder
	cells := []string{"alpha", "42", "date", "evidence"}
	widths := []int{5, 5, 6, 8}
	writeReportRow(&b, cells, widths)
	line := b.String()

	// Must end with newline
	if !strings.HasSuffix(line, "\n") {
		t.Errorf("expected row to end with newline, got %q", line)
	}
	// Count (index 1) must be right-aligned — "42" right-padded to width 5 = "   42"
	if !strings.Contains(line, "   42") {
		t.Errorf("expected count column to be right-aligned with spaces, got %q", line)
	}
	// First cell must appear left-aligned
	if !strings.Contains(line, "alpha") {
		t.Errorf("expected first cell in output, got %q", line)
	}
}

func TestWriteReportRowSingleCell(t *testing.T) {
	var b strings.Builder
	writeReportRow(&b, []string{"only"}, []int{6})
	got := b.String()
	if !strings.HasPrefix(got, "only") {
		t.Errorf("expected 'only' at start of row, got %q", got)
	}
}

// ── writeReportSeparator ──────────────────────────────────────────────────────

func TestWriteReportSeparator(t *testing.T) {
	var b strings.Builder
	widths := []int{10, 5, 8}
	writeReportSeparator(&b, widths)
	line := b.String()

	if !strings.HasSuffix(line, "\n") {
		t.Errorf("expected separator to end with newline, got %q", line)
	}
	// Must contain dashes for each column
	if !strings.Contains(line, strings.Repeat("-", 10)) {
		t.Errorf("expected 10-dash segment in separator, got %q", line)
	}
	if !strings.Contains(line, strings.Repeat("-", 5)) {
		t.Errorf("expected 5-dash segment in separator, got %q", line)
	}
}

func TestWriteReportSeparatorSingleColumn(t *testing.T) {
	var b strings.Builder
	writeReportSeparator(&b, []int{4})
	got := strings.TrimSpace(b.String())
	if got != "----" {
		t.Errorf("expected '----', got %q", got)
	}
}

// ── Report (integration) ──────────────────────────────────────────────────────

// openTestStore creates a fresh SQLite store in a temp dir and returns its path.
func openTestStore(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "faultline.db")
}

func TestReportEmptyStoreText(t *testing.T) {
	svc := Service{}
	storePath := openTestStore(t)

	var buf bytes.Buffer
	if err := svc.Report(storePath, false, &buf); err != nil {
		t.Fatalf("Report: %v", err)
	}
	if !strings.Contains(buf.String(), "No stored failures yet.") {
		t.Errorf("expected empty-store message, got %q", buf.String())
	}
}

func TestReportEmptyStoreJSON(t *testing.T) {
	svc := Service{}
	storePath := openTestStore(t)

	var buf bytes.Buffer
	if err := svc.Report(storePath, true, &buf); err != nil {
		t.Fatalf("Report JSON: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, buf.String())
	}
	if _, ok := result["failures"]; !ok {
		t.Errorf("expected 'failures' key in JSON output, got: %s", buf.String())
	}
}

func seedRun(t *testing.T, st store.Store, analysis *model.Analysis, ts time.Time) {
	t.Helper()
	ctx := context.Background()
	handle, err := st.BeginRun(ctx, store.BeginRunParams{
		Surface:    "analyze",
		SourceKind: "log",
		Source:     analysis.Source,
		InputHash:  analysis.InputHash,
		StartedAt:  ts,
	})
	if err != nil {
		t.Fatalf("BeginRun: %v", err)
	}
	if err := st.CompleteRun(ctx, handle, store.CompleteRunParams{
		CompletedAt: ts,
		Analysis:    analysis,
	}); err != nil {
		t.Fatalf("CompleteRun: %v", err)
	}
}

func TestReportWithDataText(t *testing.T) {
	storePath := openTestStore(t)

	// Seed the store with a result.
	st, _, err := openHistoryStore(storePath)
	if err != nil {
		t.Fatalf("openHistoryStore: %v", err)
	}
	analysis := &model.Analysis{
		Source:      "ci.log",
		InputHash:   "input-report-text",
		OutputHash:  "output-report-text",
		Fingerprint: "fingerprint-abc",
		Results: []model.Result{
			{
				Playbook:      model.Playbook{ID: "missing-executable", Title: "Missing executable"},
				Score:         0.9,
				Confidence:    0.85,
				SignatureHash: "sig-report-text",
			},
		},
	}
	ts := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	seedRun(t, st, analysis, ts)
	st.Close()

	svc := Service{}
	var buf bytes.Buffer
	if err := svc.Report(storePath, false, &buf); err != nil {
		t.Fatalf("Report: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "missing-executable") {
		t.Errorf("expected failure ID 'missing-executable' in report output, got %q", got)
	}
}

func TestReportWithDataJSON(t *testing.T) {
	storePath := openTestStore(t)

	st, _, err := openHistoryStore(storePath)
	if err != nil {
		t.Fatalf("openHistoryStore: %v", err)
	}
	analysis := &model.Analysis{
		Source:      "ci.log",
		InputHash:   "input-report-json",
		OutputHash:  "output-report-json",
		Fingerprint: "fingerprint-xyz",
		Results: []model.Result{
			{
				Playbook:      model.Playbook{ID: "timeout", Title: "Job timeout"},
				Score:         0.8,
				Confidence:    0.75,
				SignatureHash: "sig-report-json",
			},
		},
	}
	ts := time.Date(2025, 2, 20, 10, 0, 0, 0, time.UTC)
	seedRun(t, st, analysis, ts)
	st.Close()

	svc := Service{}
	var buf bytes.Buffer
	if err := svc.Report(storePath, true, &buf); err != nil {
		t.Fatalf("Report JSON: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	failures, ok := result["failures"].([]interface{})
	if !ok || len(failures) == 0 {
		t.Fatalf("expected non-empty failures array in JSON, got: %s", buf.String())
	}
}
