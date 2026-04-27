package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

const experimentalProviderDeltaEnv = "FAULTLINE_EXPERIMENTAL_PROVIDER_DELTA"
const experimentalGitHubDeltaEnv = "FAULTLINE_EXPERIMENTAL_GITHUB_DELTA"
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
			"The core release flow is: analyze a failing log, inspect the top playbook,",
			"and generate a deterministic follow-up workflow when you need handoff-ready output.",
		}, "\n\n"),
		Example: strings.Join([]string{
			"  faultline analyze build.log",
			"  cat build.log | faultline analyze --json",
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
	cmd.AddCommand(newCompareCommand())
	cmd.AddCommand(newReplayCommand())
	cmd.AddCommand(newTraceCommand())
	cmd.AddCommand(newInspectCommand())
	cmd.AddCommand(newGuardCommand())
	cmd.AddCommand(newPacksCommand())
	cmd.AddCommand(newHistoryCommand())
	cmd.AddCommand(newSignaturesCommand())
	cmd.AddCommand(newVerifyDeterminismCommand())
	cmd.AddCommand(newFixturesCommand())
	cmd.AddCommand(newCoverageCommand())
	return cmd
}
