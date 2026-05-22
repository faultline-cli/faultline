package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/workflow"
)

func newWorkflowCommand() *cobra.Command {
	var (
		playbookDir    string
		playbookPacks  []string
		history        bool
		noHistory      bool
		noStore        bool
		storePath      string
		gitContext     bool
		gitSince       string
		repoPath       string
		mode           string
		jsonOut        bool
		bayes          bool
		metricsHistory string
	)

	cmd := &cobra.Command{
		Use:   "workflow [file]",
		Short: "Generate a deterministic remediation handoff",
		Long: joinLines(
			"Analyze a CI log and turn the top diagnosis into a deterministic",
			"handoff with evidence, likely files, local repro commands, and verification steps.",
		),
		Example: joinLines(
			"  faultline workflow build.log",
			"  faultline workflow build.log --json --mode agent",
		),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			if mode != "local" && mode != "agent" {
				return fmt.Errorf("--mode must be %q or %q", "local", "agent")
			}
			resolvedStore := resolveStoreSetting(history, noHistory, noStore, storePath)
			return app.NewService().Workflow(cmd.Context(), input.Reader, input.Source, app.AnalyzeOptions{
				OutputOptions: app.OutputOptions{
					Top: 1,
				},
				ProviderOptions: app.ProviderOptions{
					GitContextEnabled: gitContext,
					GitSince:          gitSince,
					RepoPath:          repoPath,
				},
				PlaybookDir:        playbookDir,
				PlaybookPackDirs:   playbookPacks,
				Store:              resolvedStore,
				BayesEnabled:       bayes,
				MetricsHistoryFile: metricsHistory,
			}, appWorkflowMode(mode), jsonOut, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&mode, "mode", "local", "workflow mode: local|agent")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(&history, "history", false, "read local history for this run")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Flags().BoolVar(&gitContext, "git", true, "enrich the workflow with recent local git repository context (enabled by default; pass --git=false to disable)")
	cmd.Flags().StringVar(&gitSince, "since", "30d", "git history window for --git (for example 7d, 2w, 1 month ago)")
	cmd.Flags().StringVar(&repoPath, "repo", ".", "repository path to scan when --git is enabled")
	cmd.Flags().BoolVar(&bayes, "bayes", true, "rerank deterministic matches with the Bayesian-inspired scoring layer before building the workflow (enabled by default; pass --bayes=false to disable)")
	cmd.Flags().StringVar(&metricsHistory, "metrics-history", "", "read explicit metrics history JSONL for reliability metrics")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")
	_ = cmd.Flags().MarkHidden("metrics-history")

	return cmd
}

func appWorkflowMode(value string) workflow.Mode {
	if value == "agent" {
		return workflow.ModeAgent
	}
	return workflow.ModeLocal
}
