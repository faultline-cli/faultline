// Package ingest provides config-driven readers for heterogeneous input sources.
package ingest

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config drives the full ingestion pipeline for a single dataset.
type Config struct {
	// Name is an optional human label for the dataset. Defaults to the input filename.
	Name       string           `yaml:"name"`
	Input      InputConfig      `yaml:"input"`
	Parsing    ParsingConfig    `yaml:"parsing"`
	Processing ProcessingConfig `yaml:"processing"`
	Output     OutputConfig     `yaml:"output"`
}

// InputConfig selects the source type and path.
type InputConfig struct {
	// Type is the input format. Supported: "csv". Future: "jsonl", "ndjson".
	Type string `yaml:"type"`
	// Path is the default input file. Can be overridden by the --input CLI flag.
	Path string `yaml:"path"`
}

// ParsingConfig tells the reader how to extract log content from each row.
type ParsingConfig struct {
	// LogField is the column name that contains the log message. Required.
	LogField string `yaml:"log_field"`
	// IDField is the column name to use as a stable row identifier. Optional.
	// When empty, the SHA-256 of the log content is used.
	IDField string `yaml:"id_field"`
	// TimestampField is the column name for the event timestamp. Optional.
	TimestampField string `yaml:"timestamp_field"`
	// JoinLines controls whether embedded newlines inside a field are preserved
	// (true) or collapsed to spaces (false).
	JoinLines bool `yaml:"join_lines"`
}

// ProcessingConfig applies transformations to each extracted record.
type ProcessingConfig struct {
	// Dedupe drops records whose normalized content has already been seen.
	Dedupe bool `yaml:"dedupe"`
	// MaxLogSize is a human-readable byte limit per log (e.g. "200kb", "1mb").
	// Records exceeding this limit are truncated with a sentinel suffix.
	// Zero or empty means no limit.
	MaxLogSize string `yaml:"max_log_size"`
	// Redact applies privacy scrubbing before content hashing and storage.
	Redact RedactConfig `yaml:"redact"`
}

// RedactConfig enables pattern-based scrubbing of sensitive data.
type RedactConfig struct {
	// Emails replaces email addresses with "<email>".
	Emails bool `yaml:"emails"`
	// Tokens replaces bearer tokens and common API key patterns with "<token>".
	Tokens bool `yaml:"tokens"`
}

// OutputConfig declares where and how the corpus is written.
type OutputConfig struct {
	// Format is the output encoding. Only "jsonl" is supported.
	Format string `yaml:"format"`
	// Path is the default output file. Can be overridden by the --out CLI flag.
	Path string `yaml:"path"`
}

// LoadConfig reads and validates a YAML config file from path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &cfg, nil
}

// Validate checks that the required config fields are present.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.Parsing.LogField) == "" {
		return fmt.Errorf("parsing.log_field is required")
	}
	t := strings.TrimSpace(strings.ToLower(c.Input.Type))
	if t != "" && t != "csv" {
		return fmt.Errorf("input.type %q is not supported; supported types: csv", c.Input.Type)
	}
	return nil
}

// EffectiveName returns the dataset name for labelling fixtures, falling back
// to the basename of the input path when Name is empty.
func (c *Config) EffectiveName(inputPath string) string {
	if c.Name != "" {
		return c.Name
	}
	if inputPath != "" {
		base := inputPath
		for i := len(inputPath) - 1; i >= 0; i-- {
			if inputPath[i] == '/' || inputPath[i] == '\\' {
				base = inputPath[i+1:]
				break
			}
		}
		// strip extension
		if dot := strings.LastIndex(base, "."); dot > 0 {
			base = base[:dot]
		}
		return base
	}
	return "dataset"
}
