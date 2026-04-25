package ingest

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"faultline/tools/eval-corpus/model"
)

// CSVReader streams rows from a CSV file and extracts log records according
// to the field mapping in ParsingConfig. It implements RecordReader.
type CSVReader struct {
	f       *os.File
	r       *csv.Reader
	cfg     ParsingConfig
	source  string
	header  map[string]int // column name → zero-based index
	lineNum int
}

// NewCSVReader opens path and prepares to stream records.
// The first row must be a header row specifying column names.
func NewCSVReader(path, source string, cfg ParsingConfig) (*CSVReader, error) {
	if path == "" {
		return nil, fmt.Errorf("csv input path is required")
	}
	f, err := os.Open(path) // #nosec G304 -- path is provided by the operator, not end-user input
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	r := csv.NewReader(f)
	r.ReuseRecord = false
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	// Read and index the header row.
	header, err := r.Read()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("read csv header: %w", err)
	}
	index := make(map[string]int, len(header))
	for i, col := range header {
		index[strings.TrimSpace(col)] = i
	}

	if cfg.LogField != "" {
		if _, ok := index[cfg.LogField]; !ok {
			f.Close()
			return nil, fmt.Errorf("log_field %q not found in csv header %v", cfg.LogField, header)
		}
	}

	return &CSVReader{
		f:      f,
		r:      r,
		cfg:    cfg,
		source: source,
		header: index,
	}, nil
}

// Next reads and returns the next record. Returns nil, io.EOF when the file
// is exhausted. Other errors indicate a read failure.
func (c *CSVReader) Next() (*model.Record, error) {
	row, err := c.r.Read()
	if err != nil {
		return nil, err // includes io.EOF
	}
	c.lineNum++

	raw := c.fieldValue(row, c.cfg.LogField)
	if !c.cfg.JoinLines {
		// Collapse embedded newlines to a single space.
		raw = strings.ReplaceAll(raw, "\r\n", " ")
		raw = strings.ReplaceAll(raw, "\n", " ")
		raw = strings.ReplaceAll(raw, "\r", " ")
	}

	id := c.fieldValue(row, c.cfg.IDField)
	ts := c.fieldValue(row, c.cfg.TimestampField)

	// Collect all remaining columns as extra metadata fields.
	extras := make(map[string]string, len(c.header))
	for col, idx := range c.header {
		if col == c.cfg.LogField || col == c.cfg.IDField || col == c.cfg.TimestampField {
			continue
		}
		if idx < len(row) {
			extras[col] = row[idx]
		}
	}

	return &model.Record{
		ID:        id,
		Raw:       raw,
		Timestamp: ts,
		Fields:    extras,
		Source:    c.source,
		LineNum:   c.lineNum,
	}, nil
}

// Close releases the underlying file.
func (c *CSVReader) Close() error {
	return c.f.Close()
}

// fieldValue returns the value at the named column, or "" when the column
// is unset, the row is too short, or the field name is empty.
func (c *CSVReader) fieldValue(row []string, field string) string {
	if field == "" {
		return ""
	}
	idx, ok := c.header[field]
	if !ok || idx >= len(row) {
		return ""
	}
	return row[idx]
}

// Ensure CSVReader satisfies RecordReader at compile time.
var _ RecordReader = (*CSVReader)(nil)
var _ io.Closer = (*CSVReader)(nil)
