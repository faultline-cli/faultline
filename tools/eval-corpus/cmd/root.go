// Package cmd wires the faultline-eval CLI commands.
package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCommand returns the root cobra command for the faultline-eval binary.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "faultline-eval",
		Short: "Deterministic log ingestion and Faultline coverage evaluation",
		Long: `faultline-eval is a data pipeline for evaluating Faultline coverage
over large, heterogeneous log datasets.

Typical workflow:

  1. Ingest a dataset into a normalised corpus:
       faultline-eval ingest --config config.yaml --out corpus.jsonl

  2. (Optional) Generate a corpus manifest for versioning:
       faultline-eval manifest --corpus corpus.jsonl --corpus-id ci-v1

  3. Run Faultline over the corpus:
       faultline-eval run --corpus corpus.jsonl --out results.jsonl

  4. Generate a coverage report:
       faultline-eval report --results results.jsonl

  5. Identify missing playbooks from unmatched results:
       faultline-eval gaps --results results.jsonl --fixtures corpus.jsonl

  6. Compare against a baseline (CI gate):
       faultline-eval compare --baseline baseline.jsonl --current results.jsonl \
         --fail-on-coverage-drop

  7. Generate a badge / summary artifact:
       faultline-eval badge --results results.jsonl --corpus-version ci-v1`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newIngestCommand())
	cmd.AddCommand(newManifestCommand())
	cmd.AddCommand(newRunCommand())
	cmd.AddCommand(newReportCommand())
	cmd.AddCommand(newGapsCommand())
	cmd.AddCommand(newCompareCommand())
	cmd.AddCommand(newBadgeCommand())

	return cmd
}
