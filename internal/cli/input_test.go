package cli

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestInputCloseWithNilReader(t *testing.T) {
	var input Input
	if err := input.Close(); err != nil {
		t.Fatalf("Close() returned error for nil reader: %v", err)
	}
}

func TestReadInputFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "build.log")
	want := "example failure\n"
	if err := os.WriteFile(path, []byte(want), 0o600); err != nil {
		t.Fatalf("write temp log: %v", err)
	}

	input, err := ReadInput([]string{path})
	if err != nil {
		t.Fatalf("ReadInput(file): %v", err)
	}
	defer input.Close()

	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	if input.Source != absPath {
		t.Fatalf("Source = %q, want %q", input.Source, absPath)
	}

	got, err := io.ReadAll(input.Reader)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != want {
		t.Fatalf("reader contents = %q, want %q", got, want)
	}
}

func TestReadInputFromStdin(t *testing.T) {
	oldStdin := os.Stdin
	t.Cleanup(func() { os.Stdin = oldStdin })

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	t.Cleanup(func() { reader.Close() })
	if _, err := writer.WriteString("stdin failure\n"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close writer: %v", err)
	}
	os.Stdin = reader

	input, err := ReadInput(nil)
	if err != nil {
		t.Fatalf("ReadInput(stdin): %v", err)
	}

	if input.Source != "stdin" {
		t.Fatalf("Source = %q, want stdin", input.Source)
	}
	got, err := io.ReadAll(input.Reader)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != "stdin failure\n" {
		t.Fatalf("reader contents = %q", got)
	}
}

func TestReadInputWithoutArgsOrPipeReturnsErrNoInput(t *testing.T) {
	oldStdin := os.Stdin
	t.Cleanup(func() { os.Stdin = oldStdin })

	devNull, err := os.Open("/dev/null")
	if err != nil {
		t.Fatalf("open /dev/null: %v", err)
	}
	defer devNull.Close()
	os.Stdin = devNull

	_, err = ReadInput(nil)
	if !errors.Is(err, ErrNoInput) {
		t.Fatalf("ReadInput() error = %v, want %v", err, ErrNoInput)
	}
}
