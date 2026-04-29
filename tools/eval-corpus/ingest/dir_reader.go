package ingest

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"faultline/tools/eval-corpus/model"
)

// DirReader walks a directory tree and returns one Record per file whose
// extension matches the configured filter.  Files are opened and read lazily
// so the memory footprint scales with the size of a single log, not the whole
// corpus — important when ingesting tens-of-thousands of files.
type DirReader struct {
	source  string
	ext     string   // lower-case extension filter, e.g. ".txt"; empty = all files
	paths   []string // collected by WalkDir upfront (paths only, no open handles)
	pos     int      // next index into paths
	lineNum int
}

// NewDirReader creates a DirReader that walks root and yields files whose
// extension matches ext (case-insensitive).  ext should include the leading
// dot (e.g. ".txt") or be empty to accept all regular files.
func NewDirReader(root, source, ext string) (*DirReader, error) {
	if root == "" {
		return nil, &ErrEmptyPath{Type: "dir"}
	}

	// Normalise extension filter.
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	ext = strings.ToLower(ext)

	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if ext != "" && strings.ToLower(filepath.Ext(path)) != ext {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &DirReader{
		source: source,
		ext:    ext,
		paths:  paths,
	}, nil
}

// Next reads the next file and returns it as a Record.
// Returns nil, io.EOF when all files have been consumed.
func (d *DirReader) Next() (*model.Record, error) {
	if d.pos >= len(d.paths) {
		return nil, io.EOF
	}
	path := d.paths[d.pos]
	d.pos++
	d.lineNum++

	data, err := os.ReadFile(path) // #nosec G304 -- operator-provided directory
	if err != nil {
		return nil, err
	}

	id := filepath.Base(path)
	// Strip extension from the base name to use as a clean ID.
	if ext := filepath.Ext(id); ext != "" {
		id = id[:len(id)-len(ext)]
	}

	return &model.Record{
		ID:      id,
		Raw:     string(data),
		Source:  d.source,
		LineNum: d.lineNum,
		Fields:  map[string]string{"path": path},
	}, nil
}

// Close is a no-op because DirReader holds no open file handles between calls.
func (d *DirReader) Close() error { return nil }

// Len returns the total number of files discovered during Walk.  Useful for
// progress reporting.
func (d *DirReader) Len() int { return len(d.paths) }

// ErrEmptyPath is returned when a required path argument is empty.
type ErrEmptyPath struct{ Type string }

func (e *ErrEmptyPath) Error() string { return e.Type + " input path is required" }
