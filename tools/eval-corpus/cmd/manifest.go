package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"faultline/tools/eval-corpus/manifest"
)

func newManifestCommand() *cobra.Command {
	var (
		corpusPath  string
		outPath     string
		corpusID    string
		configPath  string
		toolVersion string
	)

	cmd := &cobra.Command{
		Use:   "manifest",
		Short: "Generate a stable versioning manifest for an ingested corpus",
		Long: `manifest produces a deterministic JSON manifest for a corpus JSONL file.
The overall_corpus_hash is stable across machines and file orderings, making it
suitable for cache keys, release tracking, and corpus version pinning.

Example:

  faultline-eval manifest \
    --corpus corpus.jsonl \
    --corpus-id ci-realworld-v1 \
    --config config.yaml \
    --out corpus.manifest.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runManifest(corpusPath, outPath, corpusID, configPath, toolVersion)
		},
	}

	cmd.Flags().StringVar(&corpusPath, "corpus", "", "path to corpus JSONL")
	cmd.Flags().StringVar(&outPath, "out", "corpus.manifest.json", "output manifest path")
	cmd.Flags().StringVar(&corpusID, "corpus-id", "", "human-readable corpus identifier")
	cmd.Flags().StringVar(&configPath, "config", "", "ingest config YAML path (hashed for the manifest)")
	cmd.Flags().StringVar(&toolVersion, "tool-version", "", "tool version string (e.g. git SHA)")

	_ = cmd.MarkFlagRequired("corpus")

	return cmd
}

func runManifest(corpusPath, outPath, corpusID, configPath, toolVersion string) error {
	configHash := ""
	if configPath != "" {
		h, err := manifest.HashFile(configPath)
		if err != nil {
			return fmt.Errorf("hash config: %w", err)
		}
		configHash = h
	}

	m, err := manifest.BuildFromFile(corpusPath, corpusID, configHash, toolVersion)
	if err != nil {
		return fmt.Errorf("build manifest: %w", err)
	}

	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	defer out.Close()

	bw := bufio.NewWriter(out)
	if err := manifest.Write(bw, m); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush manifest: %w", err)
	}

	fmt.Fprintf(os.Stderr, "manifest: %d fixtures, corpus hash %s, written to %s\n",
		m.FixtureCount, m.OverallCorpusHash[:12], outPath)
	return nil
}
