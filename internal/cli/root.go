package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

const storeEnv = "FAULTLINE_STORE"

// NewRootCommand builds the Faultline CLI command tree.
func NewRootCommand(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "faultline",
		Short: "Deterministic CI failure diagnosis from logs",
		Long: strings.Join([]string{
			"Faultline turns CI logs into deterministic diagnoses.",
			"It returns evidence-backed explanations, concrete fixes, and stable output for automation.",
			"",
			"The core release flow is: analyze a failing log, inspect source when the risk is repository-visible, and generate deterministic follow-up output.",
		}, "\n\n"),
		Example: strings.Join([]string{
			"  faultline analyze build.log",
			"  cat build.log | faultline analyze --json",
			"  faultline batch build-1.log build-2.log --json",
			"  faultline inspect .",
			"  faultline workflow build.log --json --mode agent",
			"  faultline explain docker-auth",
			"  faultline list --category auth",
		}, "\n"),
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newAnalyzeCommand())
	cmd.AddCommand(newWorkflowCommand())
	cmd.AddCommand(newExplainCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newFixCommand())
	cmd.AddCommand(newReportCommand())
	cmd.AddCommand(newInspectCommand())
	cmd.AddCommand(newFixturesCommand())
	cmd.AddCommand(newBatchCommand())
	cmd.AddCommand(newAuthCommand())
	return cmd
}
