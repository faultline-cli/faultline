package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"faultline/tools/eval-corpus/report"
)

func newReportCommand() *cobra.Command {
	var (
		resultsPath string
		comparePath string
		format      string
		topN        int
	)

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a coverage report from evaluation results",
		Long: `report reads a JSONL results file produced by 'faultline-eval run' and
prints coverage metrics, match distribution, and unmatched clusters.

When --compare is provided, the two result files are compared field-by-field
(excluding timing) to verify that the engine is deterministic.`,
		Example: `  faultline-eval report --results results.jsonl
  faultline-eval report --results results.jsonl --format json
  faultline-eval report --results run1.jsonl --compare run2.jsonl`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReport(resultsPath, comparePath, format, topN)
		},
	}

	cmd.Flags().StringVar(&resultsPath, "results", "", "results JSONL file (required)")
	cmd.Flags().StringVar(&comparePath, "compare", "", "second results JSONL for determinism check")
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	cmd.Flags().IntVar(&topN, "top", 20, "maximum entries in distribution and cluster tables")
	_ = cmd.MarkFlagRequired("results")

	return cmd
}

func runReport(resultsPath, comparePath, format string, topN int) error {
	switch format {
	case "text", "json":
	default:
		return fmt.Errorf("--format must be %q or %q", "text", "json")
	}

	results, err := report.LoadResults(resultsPath)
	if err != nil {
		return fmt.Errorf("load results: %w", err)
	}

	rpt := report.Compute(results, topN)

	if comparePath != "" {
		compare, err := report.LoadResults(comparePath)
		if err != nil {
			return fmt.Errorf("load compare results: %w", err)
		}
		dr := report.CheckDeterminism(results, compare)
		rpt.Determinism = &dr
	}

	switch format {
	case "json":
		return report.PrintJSON(os.Stdout, rpt)
	default:
		report.PrintText(os.Stdout, rpt)
	}

	return nil
}
