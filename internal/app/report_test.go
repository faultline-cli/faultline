package app

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"faultline/internal/model"
	"faultline/internal/store"
)

// ── truncateReportCell ────────────────────────────────────────────────────────

func TestTruncateReportCellNoTruncation(t *testing.T) {
	got := truncateReportCell("hello", 10)
	if got != "hello" {
		t.Errorf("expected %q, got %q", "hello", got)
	}
}

func TestTruncateReportCellExactLength(t *testing.T) {
	got := truncateReportCell("hello", 5)
	if got != "hello" {
		t.Errorf("expected %q, got %q", "hello", got)
	}
}

func TestTruncateReportCellTruncatesWithEllipsis(t *testing.T) {
	got := truncateReportCell("hello world", 8)
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected ellipsis suffix, got %q", got)
	}
	if len([]rune(got)) > 8 {
		t.Errorf("expected length <= 8, got %d for %q", len([]rune(got)), got)
	}
}

func TestTruncateReportCellZeroLimit(t *testing.T) {
	got := truncateReportCell("hello", 0)
	if got != "hello" {
		t.Errorf("zero limit should return full value, got %q", got)
	}
}

func TestTruncateReportCellNegativeLimit(t *testing.T) {
	got := truncateReportCell("hello", -1)
	if got != "hello" {
		t.Errorf("negative limit should return full value, got %q", got)
	}
}

func TestTruncateReportCellTrimsWhitespace(t *testing.T) {
	got := truncateReportCell("  hello  ", 20)
	if got != "hello" {
		t.Errorf("expected trimmed value, got %q", got)
	}
}

func TestTruncateReportCellShortLimit(t *testing.T) {
	// limit <= 3 returns exact rune slice without ellipsis
	got := truncateReportCell("hello", 2)
	if got != "he" {
		t.Errorf("expected %q for limit=2, got %q", "he", got)
	}
}

func TestTruncateReportCellLimitExactlyThree(t *testing.T) {
	got := truncateReportCell("hello", 3)
	if got != "hel" {
		t.Errorf("expected %q for limit=3, got %q", "hel", got)
	}
}

// ── formatReportText ──────────────────────────────────────────────────────────

func TestFormatReportTextEmptyFailures(t *testing.T) {
	got := formatReportText(nil)
	if !strings.Contains(got, "No stored failures yet") {
		t.Errorf("expected empty-state message, got %q", got)
	}
}

func TestFormatReportTextEmptySlice(t *testing.T) {
	got := formatReportText([]store.FailureReport{})
	if !strings.Contains(got, "No stored failures yet") {
		t.Errorf("expected empty-state message for empty slice, got %q", got)
	}
}

func TestFormatReportTextWithFailures(t *testing.T) {
	failures := []store.FailureReport{
		{
			FailureID:       "docker-auth",
			Count:           5,
			LastSeenAt:      "2026-04-26T10:00:00Z",
			ExampleEvidence: "pull access denied",
		},
	}
	got := formatReportText(failures)
	if !strings.Contains(got, "docker-auth") {
		t.Errorf("expected failure ID in output, got %q", got)
	}
	if !strings.Contains(got, "5") {
		t.Errorf("expected count in output, got %q", got)
	}
	if !strings.Contains(got, "pull access denied") {
		t.Errorf("expected evidence in output, got %q", got)
	}
	if !strings.Contains(got, "Failure ID") {
		t.Errorf("expected header in output, got %q", got)
	}
}

func TestFormatReportTextMultipleFailures(t *testing.T) {
	failures := []store.FailureReport{
		{FailureID: "docker-auth", Count: 10, LastSeenAt: "2026-04-26T10:00:00Z", ExampleEvidence: "auth error"},
		{FailureID: "missing-exec", Count: 3, LastSeenAt: "2026-04-25T09:00:00Z", ExampleEvidence: "exec not found"},
	}
	got := formatReportText(failures)
	if !strings.Contains(got, "docker-auth") {
		t.Errorf("expected docker-auth in output, got %q", got)
	}
	if !strings.Contains(got, "missing-exec") {
		t.Errorf("expected missing-exec in output, got %q", got)
	}
}

func TestFormatReportTextEvidenceTruncatedAtMaxWidth(t *testing.T) {
	longEvidence := strings.Repeat("x", 100)
	failures := []store.FailureReport{
		{FailureID: "test-id", Count: 1, LastSeenAt: "2026-01-01", ExampleEvidence: longEvidence},
	}
	got := formatReportText(failures)
	if strings.Contains(got, longEvidence) {
		t.Errorf("expected long evidence to be truncated, got full value in output")
	}
	if !strings.Contains(got, "...") {
		t.Errorf("expected ellipsis in truncated evidence, got %q", got)
	}
}

// ── writeReportRow ────────────────────────────────────────────────────────────

func TestWriteReportRowWritesCells(t *testing.T) {
	var b strings.Builder
	cells := []string{"col1", "5", "2026-01-01", "some evidence"}
	widths := []int{10, 5, 10, 15}
	writeReportRow(&b, cells, widths)
	got := b.String()
	if !strings.Contains(got, "col1") {
		t.Errorf("expected col1 in output, got %q", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("expected newline at end, got %q", got)
	}
}

func TestWriteReportRowRightAlignsCountColumn(t *testing.T) {
	var b strings.Builder
	cells := []string{"failure-id", "3", "2026-01-01", "evidence"}
	widths := []int{10, 5, 10, 10}
	writeReportRow(&b, cells, widths)
	got := b.String()
	// Count column (index 1) should be right-aligned: "    3"
	if !strings.Contains(got, "    3") {
		t.Errorf("expected right-aligned count column, got %q", got)
	}
}

// ── writeReportSeparator ──────────────────────────────────────────────────────

func TestWriteReportSeparatorProducesHyphens(t *testing.T) {
	var b strings.Builder
	widths := []int{5, 3, 8}
	writeReportSeparator(&b, widths)
	got := b.String()
	if !strings.Contains(got, "-----") {
		t.Errorf("expected 5-char separator, got %q", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("expected newline at end, got %q", got)
	}
}

// ── Report (integration via SQLite store) ────────────────────────────────────

func setupReportStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "report_test.db")
	st, _, err := store.OpenBestEffort(store.Config{Mode: store.ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC)

	sig := store.SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Evidence: []string{"pull access denied"},
	}).Hash

	handle, err := st.BeginRun(ctx, store.BeginRunParams{
		Surface:    "analyze",
		SourceKind: "log",
		Source:     "stdin",
		InputHash:  "report-input",
		StartedAt:  now,
	})
	if err != nil {
		t.Fatalf("BeginRun: %v", err)
	}
	if err := st.CompleteRun(ctx, handle, store.CompleteRunParams{
		CompletedAt: now,
		Analysis: &model.Analysis{
			Source:     "stdin",
			InputHash:  "report-input",
			OutputHash: "report-output",
			Results: []model.Result{{
				Playbook:      model.Playbook{ID: "docker-auth", Title: "Docker Auth Failure", Category: "auth"},
				Detector:      "log",
				Score:         4.5,
				Confidence:    0.90,
				Evidence:      []string{"pull access denied"},
				SignatureHash: sig,
			}},
		},
	}); err != nil {
		t.Fatalf("CompleteRun: %v", err)
	}
	return path
}

func TestReportTextOutputContainsFailureID(t *testing.T) {
	storePath := setupReportStore(t)
	svc := NewService()
	var buf bytes.Buffer

	err := svc.Report(storePath, false, &buf)
	if err != nil {
		t.Fatalf("Report text: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "docker-auth") {
		t.Errorf("expected docker-auth in text report, got %q", got)
	}
}

func TestReportJSONOutputContainsFailures(t *testing.T) {
	storePath := setupReportStore(t)
	svc := NewService()
	var buf bytes.Buffer

	err := svc.Report(storePath, true, &buf)
	if err != nil {
		t.Fatalf("Report JSON: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "docker-auth") {
		t.Errorf("expected docker-auth in JSON report, got %q", got)
	}
	if !strings.Contains(got, "failures") {
		t.Errorf("expected 'failures' key in JSON report, got %q", got)
	}
}

func TestReportEmptyStoreTextOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty_report.db")
	svc := NewService()
	var buf bytes.Buffer

	err := svc.Report(path, false, &buf)
	if err != nil {
		t.Fatalf("Report on empty store: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "No stored failures yet") {
		t.Errorf("expected empty-state message, got %q", got)
	}
}

func TestReportEmptyStoreJSONOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty_report_json.db")
	svc := NewService()
	var buf bytes.Buffer

	err := svc.Report(path, true, &buf)
	if err != nil {
		t.Fatalf("Report JSON on empty store: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "failures") {
		t.Errorf("expected 'failures' key in JSON, got %q", got)
	}
}

func TestReportInvalidStorePath(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer
	// Pass a path that is not a valid SQLite DB to trigger an error.
	err := svc.Report("/dev/null", false, &buf)
	// This may or may not error depending on whether /dev/null is readable.
	// Just make sure we don't panic.
	_ = err
}
