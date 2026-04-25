package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"faultline/tools/eval-corpus/ingest"
	"faultline/tools/eval-corpus/model"
	"faultline/tools/eval-corpus/normalize"
)

func newIngestCommand() *cobra.Command {
	var (
		configPath string
		inputPath  string
		outPath    string
		sourceName string
	)

	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Ingest a dataset into a normalised fixture corpus",
		Long: `ingest reads an input dataset (CSV, etc.) according to a YAML config,
normalises and optionally deduplicates each log record, and writes the
resulting corpus to a JSONL file where each line is a Fixture.

The output file can then be evaluated with:
  faultline-eval run --corpus <out>`,
		Example: `  faultline-eval ingest --config config.yaml --out corpus.jsonl
  faultline-eval ingest --config config.yaml --input other.csv --out corpus.jsonl
  faultline-eval ingest --config config.yaml --out corpus.jsonl --source mydata`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runIngest(configPath, inputPath, outPath, sourceName, cmd)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to ingest config YAML (required)")
	cmd.Flags().StringVar(&inputPath, "input", "", "override input file path from config")
	cmd.Flags().StringVar(&outPath, "out", "", "output corpus JSONL path (required)")
	cmd.Flags().StringVar(&sourceName, "source", "", "dataset source label (overrides config name)")
	_ = cmd.MarkFlagRequired("config")
	_ = cmd.MarkFlagRequired("out")

	return cmd
}

func runIngest(configPath, inputPath, outPath, sourceName string, cmd *cobra.Command) error {
	cfg, err := ingest.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Resolve effective input path.
	effectiveInput := inputPath
	if effectiveInput == "" {
		effectiveInput = cfg.Input.Path
	}
	if effectiveInput == "" {
		return fmt.Errorf("no input path: set input.path in config or pass --input")
	}

	// Resolve effective source label.
	if sourceName == "" {
		sourceName = cfg.EffectiveName(effectiveInput)
	}

	// Parse max log size limit.
	maxBytes, err := normalize.ParseSize(cfg.Processing.MaxLogSize)
	if err != nil {
		return fmt.Errorf("processing.max_log_size: %w", err)
	}

	// Build redact options from config.
	redactOpts := normalize.RedactOptions{
		Emails: cfg.Processing.Redact.Emails,
		Tokens: cfg.Processing.Redact.Tokens,
	}

	// Open the record reader.
	reader, err := ingest.NewReader(cfg, effectiveInput, sourceName)
	if err != nil {
		return fmt.Errorf("open reader: %w", err)
	}
	defer reader.Close()

	// Open the output file.
	outFile, err := os.Create(outPath) // #nosec G304 -- operator-provided path
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer outFile.Close()

	bw := bufio.NewWriterSize(outFile, 1<<20) // 1 MiB write buffer
	enc := json.NewEncoder(bw)

	dedupe := normalize.NewDeduplicator()

	var (
		total   int
		written int
		skipped int
	)

	for {
		rec, err := reader.Next()
		if err != nil {
			if isEOF(err) {
				break
			}
			return fmt.Errorf("read record: %w", err)
		}
		total++

		raw := rec.Raw

		// Apply size limit.
		raw = normalize.ApplyMaxSize(raw, maxBytes)

		// Apply redaction.
		raw = normalize.Redact(redactOpts, raw)

		// Compute deterministic fixture ID.
		id := normalize.FixtureID(raw)

		// Deduplicate.
		if cfg.Processing.Dedupe && dedupe.Seen(id) {
			skipped++
			continue
		}
		dedupe.Mark(id)

		fix := model.Fixture{
			ID:     id,
			Raw:    raw,
			Source: rec.Source,
			Metadata: model.FixtureMetadata{
				Timestamp: rec.Timestamp,
				Fields:    rec.Fields,
				LineNum:   rec.LineNum,
			},
		}

		if err := enc.Encode(fix); err != nil {
			return fmt.Errorf("write fixture: %w", err)
		}
		written++
	}

	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush output: %w", err)
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "ingest: source=%s total=%d written=%d skipped=%d\n",
		sourceName, total, written, skipped)

	return nil
}
