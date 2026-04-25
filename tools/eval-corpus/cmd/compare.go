package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"faultline/tools/eval-corpus/compare"
	"faultline/tools/eval-corpus/report"
)

func newCompareCommand() *cobra.Command {
	var (
		baselinePath            string
		currentPath             string
		outPath                 string
		minMatchRate            float64
		failOnNewNondeterminism bool
		failOnCoverageDrop      bool
	)

	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare evaluation results against a baseline (CI regression gate)",
		Long: `compare checks whether current evaluation results meet coverage quality gates
compared to a stored baseline.

Exit code 1 indicates a gate failure. Zero means all gates passed.

Example:

  faultline-eval compare \
    --baseline baseline-results.jsonl \
    --current  results.jsonl \
    --min-match-rate 0.72 \
    --fail-on-coverage-drop \
    --out comparison`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompare(
				baselinePath, currentPath, outPath,
				compare.Options{
					MinMatchRate:            minMatchRate,
					FailOnNewNondeterminism: failOnNewNondeterminism,
					FailOnCoverageDrop:      failOnCoverageDrop,
				},
			)
		},
	}

	cmd.Flags().StringVar(&baselinePath, "baseline", "", "path to baseline results JSONL")
	cmd.Flags().StringVar(&currentPath, "current", "", "path to current results JSONL")
	cmd.Flags().StringVar(&outPath, "out", "comparison", "output path prefix (produces .json and .md)")
	cmd.Flags().Float64Var(&minMatchRate, "min-match-rate", 0, "minimum acceptable match rate (0–1); 0 disables check")
	cmd.Flags().BoolVar(&failOnNewNondeterminism, "fail-on-new-nondeterminism", false, "fail when new non-deterministic fixtures are detected")
	cmd.Flags().BoolVar(&failOnCoverageDrop, "fail-on-coverage-drop", false, "fail when current match rate is lower than baseline")

	_ = cmd.MarkFlagRequired("baseline")
	_ = cmd.MarkFlagRequired("current")

	return cmd
}

func runCompare(baselinePath, currentPath, outPath string, opts compare.Options) error {
	baseline, err := report.LoadResults(baselinePath)
	if err != nil {
		return fmt.Errorf("load baseline: %w", err)
	}
	current, err := report.LoadResults(currentPath)
	if err != nil {
		return fmt.Errorf("load current: %w", err)
	}

	res := compare.Compare(baseline, current, opts)

	if opts.FailOnNewNondeterminism {
		det := report.CheckDeterminism(baseline, current)
		compare.AttachNondeterminism(&res, det, opts)
	}

	// Write JSON.
	jsonPath := outPath + ".json"
	jsonFile, err := os.Create(jsonPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", jsonPath, err)
	}
	defer jsonFile.Close()
	if err := compare.WriteJSON(bufio.NewWriter(jsonFile), res); err != nil {
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
	compare.PrintMarkdownReport(bw, res)
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush markdown: %w", err)
	}

	// Print text to stdout.
	compare.PrintTextReport(os.Stdout, res)

	fmt.Fprintf(os.Stderr, "compare: written to %s.json and %s.md\n", outPath, outPath)

	if !res.Pass {
		return fmt.Errorf("gate failed: %v", res.FailReasons)
	}
	return nil
}
