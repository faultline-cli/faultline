package cli

// Package cli handles command-line argument parsing and log input acquisition.
// It has no dependencies on other Faultline internal packages.

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ErrNoInput is returned when no log input can be obtained — either no file
// path was given and stdin is an interactive terminal.
var ErrNoInput = errors.New("no log input provided; pass a file path or pipe stdin")

// Input holds a log reader and a human-readable label for its source.
type Input struct {
	Reader io.ReadCloser
	Source string
}

// Close releases any underlying file handle opened by ReadInput.
func (i *Input) Close() error {
	if i.Reader != nil {
		return i.Reader.Close()
	}
	return nil
}

// ReadInput resolves the log source from args:
//
//   - If args contains at least one element, args[0] is treated as a file path.
//   - Otherwise stdin is used, provided it is not an interactive TTY.
//
// The caller is responsible for closing the returned Input.
func ReadInput(args []string) (*Input, error) {
	if len(args) >= 1 {
		path, err := filepath.Abs(args[0])
		if err != nil {
			return nil, fmt.Errorf("resolve input path: %w", err)
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		return &Input{Reader: f, Source: path}, nil
	}
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect stdin: %w", err)
	}
	if stat.Mode()&os.ModeCharDevice != 0 {
		return nil, ErrNoInput
	}
	return &Input{Reader: io.NopCloser(os.Stdin), Source: "stdin"}, nil
}
