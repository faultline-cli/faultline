package ingest_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"faultline/tools/eval-corpus/ingest"
)

// makeLogDir creates a temporary directory containing n plain-text log files
// and returns the directory path.
func makeLogDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatalf("write test file: %v", err)
		}
	}
	return dir
}

func TestDirReaderStreamsAllTxtFiles(t *testing.T) {
	dir := makeLogDir(t, map[string]string{
		"a.txt": "log line one\nlog line two\n",
		"b.txt": "another log\n",
		"c.txt": "third log\n",
	})

	r, err := ingest.NewDirReader(dir, "test", ".txt")
	if err != nil {
		t.Fatalf("NewDirReader: %v", err)
	}
	defer r.Close()

	if r.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", r.Len())
	}

	seen := make(map[string]bool)
	for {
		rec, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		seen[rec.ID] = true
		if rec.Source != "test" {
			t.Errorf("Source = %q, want %q", rec.Source, "test")
		}
		if rec.Raw == "" {
			t.Errorf("Raw is empty for record %q", rec.ID)
		}
		if _, ok := rec.Fields["path"]; !ok {
			t.Errorf("Fields[path] missing for record %q", rec.ID)
		}
	}

	if len(seen) != 3 {
		t.Errorf("got %d distinct records, want 3", len(seen))
	}
}

func TestDirReaderFiltersByExtension(t *testing.T) {
	dir := makeLogDir(t, map[string]string{
		"a.txt": "log one",
		"b.log": "log two",
		"c.txt": "log three",
	})

	r, err := ingest.NewDirReader(dir, "src", ".txt")
	if err != nil {
		t.Fatalf("NewDirReader: %v", err)
	}
	defer r.Close()

	if r.Len() != 2 {
		t.Fatalf("Len() = %d, want 2 (only .txt files)", r.Len())
	}
}

func TestDirReaderAcceptsAllFilesWhenExtEmpty(t *testing.T) {
	dir := makeLogDir(t, map[string]string{
		"a.txt": "one",
		"b.log": "two",
		"c":     "three",
	})

	r, err := ingest.NewDirReader(dir, "src", "")
	if err != nil {
		t.Fatalf("NewDirReader: %v", err)
	}
	defer r.Close()

	if r.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", r.Len())
	}
}

func TestDirReaderExtWithoutLeadingDot(t *testing.T) {
	dir := makeLogDir(t, map[string]string{
		"a.txt": "one",
		"b.log": "two",
	})

	r, err := ingest.NewDirReader(dir, "src", "txt") // no leading dot
	if err != nil {
		t.Fatalf("NewDirReader: %v", err)
	}
	defer r.Close()

	if r.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", r.Len())
	}
}

func TestDirReaderEmptyPathError(t *testing.T) {
	_, err := ingest.NewDirReader("", "src", ".txt")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestDirReaderEOFOnEmptyDir(t *testing.T) {
	dir := t.TempDir() // empty

	r, err := ingest.NewDirReader(dir, "src", ".txt")
	if err != nil {
		t.Fatalf("NewDirReader: %v", err)
	}
	defer r.Close()

	if r.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", r.Len())
	}

	rec, err := r.Next()
	if err != io.EOF {
		t.Fatalf("Next() on empty dir: err = %v, rec = %v; want io.EOF", err, rec)
	}
}

func TestDirReaderLineNumMonotonicallyIncreases(t *testing.T) {
	dir := makeLogDir(t, map[string]string{
		"x.txt": "foo",
		"y.txt": "bar",
	})

	r, err := ingest.NewDirReader(dir, "src", ".txt")
	if err != nil {
		t.Fatalf("NewDirReader: %v", err)
	}
	defer r.Close()

	rec1, _ := r.Next()
	rec2, _ := r.Next()

	if rec1.LineNum != 1 {
		t.Errorf("first LineNum = %d, want 1", rec1.LineNum)
	}
	if rec2.LineNum != 2 {
		t.Errorf("second LineNum = %d, want 2", rec2.LineNum)
	}
}
