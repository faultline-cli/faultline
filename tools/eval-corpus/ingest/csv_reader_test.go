package ingest_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/tools/eval-corpus/ingest"
)

func TestCSVReaderStreamsRows(t *testing.T) {
	path := filepath.Join("..", "testdata", "sample.csv")
	cfg := ingest.ParsingConfig{
		LogField:       "message",
		IDField:        "id",
		TimestampField: "timestamp",
		JoinLines:      false,
	}
	r, err := ingest.NewCSVReader(path, "test", cfg)
	if err != nil {
		t.Fatalf("NewCSVReader: %v", err)
	}
	defer r.Close()

	var records []string
	for {
		rec, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		records = append(records, rec.Raw)
	}

	// sample.csv has 6 data rows.
	if len(records) != 6 {
		t.Errorf("got %d records, want 6", len(records))
	}
}

func TestCSVReaderExtractsFields(t *testing.T) {
	path := filepath.Join("..", "testdata", "sample.csv")
	cfg := ingest.ParsingConfig{
		LogField:       "message",
		IDField:        "id",
		TimestampField: "timestamp",
	}
	r, err := ingest.NewCSVReader(path, "ds", cfg)
	if err != nil {
		t.Fatalf("NewCSVReader: %v", err)
	}
	defer r.Close()

	rec, err := r.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if rec.ID != "row-001" {
		t.Errorf("ID = %q, want %q", rec.ID, "row-001")
	}
	if rec.Timestamp != "2024-01-01T00:00:00Z" {
		t.Errorf("Timestamp = %q, want %q", rec.Timestamp, "2024-01-01T00:00:00Z")
	}
	if !strings.Contains(rec.Raw, "unauthorized") {
		t.Errorf("Raw %q does not contain expected content", rec.Raw)
	}
	if rec.Source != "ds" {
		t.Errorf("Source = %q, want %q", rec.Source, "ds")
	}
}

func TestCSVReaderMissingLogField(t *testing.T) {
	// Create a temp CSV without the required log_field column.
	content := "col_a,col_b\nval1,val2\n"
	tmp := writeTempFile(t, content, "*.csv")
	cfg := ingest.ParsingConfig{LogField: "message"}
	_, err := ingest.NewCSVReader(tmp, "test", cfg)
	if err == nil {
		t.Fatal("expected error for missing log_field, got nil")
	}
}

func TestCSVReaderJoinLinesCollapsesNewlines(t *testing.T) {
	content := "id,message\nr1,\"line one\nline two\"\n"
	tmp := writeTempFile(t, content, "*.csv")

	cfg := ingest.ParsingConfig{LogField: "message", JoinLines: false}
	r, err := ingest.NewCSVReader(tmp, "test", cfg)
	if err != nil {
		t.Fatalf("NewCSVReader: %v", err)
	}
	defer r.Close()

	rec, err := r.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if strings.Contains(rec.Raw, "\n") {
		t.Errorf("JoinLines=false should collapse newlines, got: %q", rec.Raw)
	}
}

func TestCSVReaderJoinLinesPreservesNewlines(t *testing.T) {
	content := "id,message\nr1,\"line one\nline two\"\n"
	tmp := writeTempFile(t, content, "*.csv")

	cfg := ingest.ParsingConfig{LogField: "message", JoinLines: true}
	r, err := ingest.NewCSVReader(tmp, "test", cfg)
	if err != nil {
		t.Fatalf("NewCSVReader: %v", err)
	}
	defer r.Close()

	rec, err := r.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if !strings.Contains(rec.Raw, "\n") {
		t.Errorf("JoinLines=true should preserve newlines, got: %q", rec.Raw)
	}
}

func TestCSVReaderLineNumbers(t *testing.T) {
	content := "id,message\nr1,first\nr2,second\n"
	tmp := writeTempFile(t, content, "*.csv")
	cfg := ingest.ParsingConfig{LogField: "message", IDField: "id"}
	r, err := ingest.NewCSVReader(tmp, "test", cfg)
	if err != nil {
		t.Fatalf("NewCSVReader: %v", err)
	}
	defer r.Close()

	for want := 1; want <= 2; want++ {
		rec, err := r.Next()
		if err != nil {
			t.Fatalf("Next at line %d: %v", want, err)
		}
		if rec.LineNum != want {
			t.Errorf("line %d: LineNum = %d, want %d", want, rec.LineNum, want)
		}
	}
}

func writeTempFile(t *testing.T, content, pattern string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), pattern)
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if _, err := io.WriteString(f, content); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	f.Close()
	return f.Name()
}
