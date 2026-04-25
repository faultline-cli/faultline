package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"faultline/tools/eval-corpus/badge"
	"faultline/tools/eval-corpus/report"
)

func newBadgeCommand() *cobra.Command {
	var (
		resultsPath   string
		outPath       string
		corpusVersion string
		corpusHash    string
		deterministic string
		topN          int
	)

	cmd := &cobra.Command{
		Use:   "badge",
		Short: "Generate a lightweight coverage summary artifact",
		Long: `badge produces a machine-readable JSON summary and a Markdown snippet
from an evaluation results file. Suitable for README badges, release notes,
or CI PR comments.

Example:

  faultline-eval badge \
    --results results.jsonl \
    --corpus-version ci-realworld-v1 \
    --deterministic pass \
    --out coverage-summary`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBadge(resultsPath, outPath, badge.Options{
				CorpusVersion: corpusVersion,
				CorpusHash:    corpusHash,
				Deterministic: deterministic,
				TopN:          topN,
			})
		},
	}

	cmd.Flags().StringVar(&resultsPath, "results", "", "path to results JSONL")
	cmd.Flags().StringVar(&outPath, "out", "coverage-summary", "output path prefix (produces .json and .md)")
	cmd.Flags().StringVar(&corpusVersion, "corpus-version", "", "human-readable corpus version label")
	cmd.Flags().StringVar(&corpusHash, "corpus-hash", "", "corpus content hash (from 'faultline-eval manifest')")
	cmd.Flags().StringVar(&deterministic, "deterministic", "", "determinism status: pass, fail, or empty for unknown")
	cmd.Flags().IntVar(&topN, "top", 5, "number of entries to show in top-covered and top-gaps")

	_ = cmd.MarkFlagRequired("results")

	return cmd
}

func runBadge(resultsPath, outPath string, opts badge.Options) error {
	results, err := report.LoadResults(resultsPath)
	if err != nil {
		return fmt.Errorf("load results: %w", err)
	}

	s := badge.Compute(results, opts)

	// Write JSON.
	jsonPath := outPath + ".json"
	jsonFile, err := os.Create(jsonPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", jsonPath, err)
	}
	defer jsonFile.Close()
	if err := badge.WriteJSON(bufio.NewWriter(jsonFile), s); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}

	// Write Markdown.
	mdPath := outPath + ".md"
	mdFile, err := os.Create(mdPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", mdPath, err)
	}
	defer mdFile.Close()
	bw := bufio.NewWriter(mdFile)
	badge.WriteMarkdown(bw, s)
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush markdown: %w", err)
	}

	fmt.Fprintf(os.Stderr, "badge: written to %s.json and %s.md\n", outPath, outPath)
	return nil
}
