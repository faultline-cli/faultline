package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"faultline/internal/model"
	"faultline/internal/output"
)

// writeTempLogFile creates a temporary log file with the given content and
// returns its path. Cleaned up automatically via t.TempDir.
func writeTempLogFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	f, err := os.CreateTemp(dir, "batch-*.log")
	if err != nil {
		t.Fatalf("create temp log: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		_ = f.Close()
		t.Fatalf("write temp log: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp log: %v", err)
	}
	return f.Name()
}

// batchOpts returns minimal AnalyzeOptions suitable for Batch tests.
func batchOpts() AnalyzeOptions {
	return AnalyzeOptions{
		OutputOptions: OutputOptions{
			Top:    1,
			Mode:   output.ModeQuick,
			Format: output.FormatTerminal,
		},
		Store:       "off",
		PlaybookDir: repoPlaybookDir(),
	}
}

const (
	batchDockerAuthLog = "pull access denied\nError response from daemon: authentication required\n"
	batchGitAuthLog    = "fatal: could not read Username for 'https://github.com': terminal prompts disabled\n"
	batchUnmatchedLog  = "everything is perfectly fine and nothing is wrong here\n"
)

// ── Batch: all sources match ──────────────────────────────────────────────────

func TestBatchAllMatchedTextOutput(t *testing.T) {
	svc := NewService()
	f1 := writeTempLogFile(t, batchDockerAuthLog)
	f2 := writeTempLogFile(t, batchDockerAuthLog)
	var buf bytes.Buffer

	err := svc.Batch([]string{f1, f2}, batchOpts(), &buf)
	if err != nil {
		t.Fatalf("Batch all-matched: expected nil, got %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "docker-auth") {
		t.Errorf("expected docker-auth pattern in output, got %q", out)
	}
	if !strings.Contains(out, "2/2 matched") {
		t.Errorf("expected '2/2 matched' in output, got %q", out)
	}
	// Single distinct pattern → "1 pattern" (singular)
	if !strings.Contains(out, "1 pattern") {
		t.Errorf("expected '1 pattern' in output, got %q", out)
	}
}

func TestBatchAllMatchedJSONOutput(t *testing.T) {
	svc := NewService()
	f1 := writeTempLogFile(t, batchDockerAuthLog)
	opts := batchOpts()
	opts.JSON = true
	var buf bytes.Buffer

	err := svc.Batch([]string{f1}, opts, &buf)
	if err != nil {
		t.Fatalf("Batch JSON all-matched: expected nil, got %v", err)
	}
	var result model.BatchResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &result); err != nil {
		t.Fatalf("unmarshal batch JSON: %v", err)
	}
	if result.SchemaVersion != "batch.v1" {
		t.Errorf("expected schema_version batch.v1, got %q", result.SchemaVersion)
	}
	if result.Matched != 1 {
		t.Errorf("expected matched=1, got %d", result.Matched)
	}
	if result.Total != 1 {
		t.Errorf("expected total=1, got %d", result.Total)
	}
	if result.Unmatched != 0 {
		t.Errorf("expected unmatched=0, got %d", result.Unmatched)
	}
	if len(result.Patterns) != 1 {
		t.Errorf("expected 1 pattern, got %d", len(result.Patterns))
	}
}

// ── Batch: FormatJSON flag triggers same JSON path ────────────────────────────

func TestBatchFormatJSONFlag(t *testing.T) {
	svc := NewService()
	f1 := writeTempLogFile(t, batchDockerAuthLog)
	opts := batchOpts()
	opts.Format = output.FormatJSON
	var buf bytes.Buffer

	err := svc.Batch([]string{f1}, opts, &buf)
	if err != nil {
		t.Fatalf("Batch FormatJSON: expected nil, got %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &result); err != nil {
		t.Fatalf("unmarshal FormatJSON batch output: %v", err)
	}
	if result["schema_version"] != "batch.v1" {
		t.Errorf("expected schema_version batch.v1, got %v", result["schema_version"])
	}
}

// ── Batch: partial match ──────────────────────────────────────────────────────

func TestBatchPartialMatchReturnsErrBatchUnmatched(t *testing.T) {
	svc := NewService()
	f1 := writeTempLogFile(t, batchDockerAuthLog)
	f2 := writeTempLogFile(t, batchUnmatchedLog)
	var buf bytes.Buffer

	err := svc.Batch([]string{f1, f2}, batchOpts(), &buf)
	if !errors.Is(err, ErrBatchUnmatched) {
		t.Fatalf("expected ErrBatchUnmatched, got %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "docker-auth") {
		t.Errorf("expected docker-auth pattern in partial-match output, got %q", out)
	}
	// Unmatched section with singular "file" (1 unmatched)
	if !strings.Contains(out, "Unmatched") {
		t.Errorf("expected Unmatched section in output, got %q", out)
	}
	if !strings.Contains(out, "1/2 matched") {
		t.Errorf("expected '1/2 matched' in partial-match output, got %q", out)
	}
	// Footer shows unmatched count
	if !strings.Contains(out, "1 unmatched") {
		t.Errorf("expected '1 unmatched' in footer, got %q", out)
	}
}

// ── Batch: all unmatched ──────────────────────────────────────────────────────

func TestBatchAllUnmatchedTextOutput(t *testing.T) {
	svc := NewService()
	f1 := writeTempLogFile(t, batchUnmatchedLog)
	var buf bytes.Buffer

	err := svc.Batch([]string{f1}, batchOpts(), &buf)
	if !errors.Is(err, ErrBatchUnmatched) {
		t.Fatalf("expected ErrBatchUnmatched for all-unmatched, got %v", err)
	}
	out := buf.String()
	// Matched == 0 early-return path in formatBatchText
	if !strings.Contains(out, "No playbook matched") {
		t.Errorf("expected 'No playbook matched' text for all-unmatched, got %q", out)
	}
	// Singular "file" wording (1 source)
	if !strings.Contains(out, "1 file") {
		t.Errorf("expected '1 file' (singular) for single-source batch, got %q", out)
	}
}

func TestBatchAllUnmatchedJSONOutput(t *testing.T) {
	svc := NewService()
	f1 := writeTempLogFile(t, batchUnmatchedLog)
	opts := batchOpts()
	opts.JSON = true
	var buf bytes.Buffer

	err := svc.Batch([]string{f1}, opts, &buf)
	if !errors.Is(err, ErrBatchUnmatched) {
		t.Fatalf("expected ErrBatchUnmatched for JSON all-unmatched, got %v", err)
	}
	var result model.BatchResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &result); err != nil {
		t.Fatalf("unmarshal JSON all-unmatched: %v", err)
	}
	if result.Matched != 0 {
		t.Errorf("expected matched=0, got %d", result.Matched)
	}
	if result.Unmatched != 1 {
		t.Errorf("expected unmatched=1, got %d", result.Unmatched)
	}
}

// ── Batch: file open error ────────────────────────────────────────────────────

func TestBatchMissingFileReturnsError(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer

	err := svc.Batch([]string{"/nonexistent/path/batch_test_file.log"}, batchOpts(), &buf)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if errors.Is(err, ErrBatchUnmatched) {
		t.Fatalf("expected file-open error (not ErrBatchUnmatched), got ErrBatchUnmatched")
	}
}

// ── Batch: multiple distinct patterns ────────────────────────────────────────

func TestBatchMultipleDistinctPatterns(t *testing.T) {
	svc := NewService()
	f1 := writeTempLogFile(t, batchDockerAuthLog)
	f2 := writeTempLogFile(t, batchGitAuthLog)
	var buf bytes.Buffer

	err := svc.Batch([]string{f1, f2}, batchOpts(), &buf)
	if err != nil {
		t.Fatalf("Batch two patterns: expected nil, got %v", err)
	}
	out := buf.String()
	// Two distinct patterns → "2 distinct patterns"
	if !strings.Contains(out, "2 distinct patterns") {
		t.Errorf("expected '2 distinct patterns' in output, got %q", out)
	}
	if !strings.Contains(out, "2/2 matched") {
		t.Errorf("expected '2/2 matched' in output, got %q", out)
	}
}

// ── Batch: source list truncation (>3 sources per pattern) ───────────────────

func TestBatchSourceListTruncatedAboveThree(t *testing.T) {
	svc := NewService()
	// 4 files matching same pattern — triggers "  +1 more" in formatBatchText
	files := make([]string, 4)
	for i := range files {
		files[i] = writeTempLogFile(t, batchDockerAuthLog)
	}
	var buf bytes.Buffer

	err := svc.Batch(files, batchOpts(), &buf)
	if err != nil {
		t.Fatalf("Batch 4 files same pattern: expected nil, got %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "+1 more") {
		t.Errorf("expected '+1 more' for 4 sources in same pattern, got %q", out)
	}
	if !strings.Contains(out, "4/4 matched") {
		t.Errorf("expected '4/4 matched', got %q", out)
	}
}

// ── Batch: empty source list ──────────────────────────────────────────────────

func TestBatchEmptySourcesReturnsNil(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer

	err := svc.Batch([]string{}, batchOpts(), &buf)
	if err != nil {
		t.Fatalf("Batch empty sources: expected nil, got %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "FAULTLINE  batch") {
		t.Errorf("expected batch header in output, got %q", out)
	}
}

// ── Batch: patterns summary singular/plural ───────────────────────────────────

func TestBatchSinglePatternSingularWording(t *testing.T) {
	svc := NewService()
	f1 := writeTempLogFile(t, batchDockerAuthLog)
	var buf bytes.Buffer

	err := svc.Batch([]string{f1}, batchOpts(), &buf)
	if err != nil {
		t.Fatalf("Batch single file single pattern: expected nil, got %v", err)
	}
	out := buf.String()
	// "1 distinct pattern" → formatBatchText uses "pattern" (singular)
	if !strings.Contains(out, "1 distinct pattern") {
		t.Errorf("expected '1 distinct pattern' in header, got %q", out)
	}
}

// ── Batch: JSON entries include expected fields ───────────────────────────────

func TestBatchJSONIncludesEntries(t *testing.T) {
	svc := NewService()
	f1 := writeTempLogFile(t, batchDockerAuthLog)
	f2 := writeTempLogFile(t, batchUnmatchedLog)
	opts := batchOpts()
	opts.JSON = true
	var buf bytes.Buffer

	_ = svc.Batch([]string{f1, f2}, opts, &buf)

	var result model.BatchResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &result); err != nil {
		t.Fatalf("unmarshal entries JSON: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}
	matched := 0
	for _, e := range result.Entries {
		if e.Matched {
			matched++
		}
	}
	if matched != 1 {
		t.Errorf("expected 1 matched entry, got %d", matched)
	}
}

// ── Batch: markdown format ────────────────────────────────────────────────────

func TestBatchMarkdownFormatAllMatched(t *testing.T) {
	svc := NewService()
	f1 := writeTempLogFile(t, batchDockerAuthLog)
	f2 := writeTempLogFile(t, batchDockerAuthLog)
	opts := batchOpts()
	opts.Format = output.FormatMarkdown
	var buf bytes.Buffer

	err := svc.Batch([]string{f1, f2}, opts, &buf)
	if err != nil {
		t.Fatalf("Batch markdown all-matched: expected nil, got %v", err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "# Faultline Batch") {
		t.Errorf("expected markdown heading, got %q", out[:min(60, len(out))])
	}
	if !strings.Contains(out, "docker-auth") {
		t.Errorf("expected docker-auth pattern in markdown output, got %q", out)
	}
	if !strings.Contains(out, "Matched: 2/2") {
		t.Errorf("expected 'Matched: 2/2' in markdown output, got %q", out)
	}
	if !strings.Contains(out, "## Patterns") {
		t.Errorf("expected '## Patterns' section in markdown output, got %q", out)
	}
}

func TestBatchMarkdownFormatPartialMatch(t *testing.T) {
	svc := NewService()
	f1 := writeTempLogFile(t, batchDockerAuthLog)
	f2 := writeTempLogFile(t, batchUnmatchedLog)
	opts := batchOpts()
	opts.Format = output.FormatMarkdown
	var buf bytes.Buffer

	err := svc.Batch([]string{f1, f2}, opts, &buf)
	if !errors.Is(err, ErrBatchUnmatched) {
		t.Fatalf("expected ErrBatchUnmatched, got %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "## Unmatched") {
		t.Errorf("expected '## Unmatched' section in markdown output, got %q", out)
	}
	if !strings.Contains(out, "Unmatched: 1/2") {
		t.Errorf("expected 'Unmatched: 1/2' in markdown output, got %q", out)
	}
}

func TestBatchMarkdownFormatAllUnmatched(t *testing.T) {
	svc := NewService()
	f1 := writeTempLogFile(t, batchUnmatchedLog)
	opts := batchOpts()
	opts.Format = output.FormatMarkdown
	var buf bytes.Buffer

	err := svc.Batch([]string{f1}, opts, &buf)
	if !errors.Is(err, ErrBatchUnmatched) {
		t.Fatalf("expected ErrBatchUnmatched, got %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "No playbook matched") {
		t.Errorf("expected 'No playbook matched' in all-unmatched markdown output, got %q", out)
	}
}

func TestBatchMarkdownSourceListTruncatedAboveThree(t *testing.T) {
	svc := NewService()
	files := make([]string, 4)
	for i := range files {
		files[i] = writeTempLogFile(t, batchDockerAuthLog)
	}
	opts := batchOpts()
	opts.Format = output.FormatMarkdown
	var buf bytes.Buffer

	err := svc.Batch(files, opts, &buf)
	if err != nil {
		t.Fatalf("Batch markdown 4 files: expected nil, got %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "+1 more") {
		t.Errorf("expected '+1 more' truncation in markdown output, got %q", out)
	}
}
