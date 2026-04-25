package ingest

import (
	"fmt"

	"faultline/tools/eval-corpus/model"
)

// RecordReader is the abstraction over all input source adapters.
// Callers iterate with Next until io.EOF is returned, then call Close.
type RecordReader interface {
	// Next returns the next record. Returns nil, io.EOF when exhausted.
	// Any other error is a read failure and should abort the pipeline.
	Next() (*model.Record, error)
	// Close releases any underlying resources.
	Close() error
}

// NewReader returns the appropriate RecordReader for the given config.
// inputPath overrides cfg.Input.Path when non-empty.
// source is the dataset label applied to every record.
func NewReader(cfg *Config, inputPath, source string) (RecordReader, error) {
	if inputPath == "" {
		inputPath = cfg.Input.Path
	}
	switch cfg.Input.Type {
	case "", "csv":
		return NewCSVReader(inputPath, source, cfg.Parsing)
	default:
		return nil, fmt.Errorf("unsupported input type %q", cfg.Input.Type)
	}
}
