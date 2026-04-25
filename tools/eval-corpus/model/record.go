// Package model defines the shared data types for the eval-corpus pipeline.
package model

// Record is a single row extracted from an input source before normalization.
type Record struct {
	// ID is extracted from the id_field config, or generated from line number.
	ID string
	// Raw is the log text extracted from the log_field config.
	Raw string
	// Timestamp is the value of the timestamp_field config, if present.
	Timestamp string
	// Fields holds all remaining columns from the source row.
	Fields map[string]string
	// Source identifies the dataset (file name or config name).
	Source string
	// LineNum is the 1-based row index in the source file.
	LineNum int
}
