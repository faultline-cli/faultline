package cli

import (
	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/output"
)

func newTraceCommand() *cobra.Command {
	var (
		jsonOut       bool
		format        string
		playbookDir   string
		playbookPacks []string
		history       bool
		noHistory     bool
		noStore       bool
		storePath     string
		gitContext    bool
		gitSince      string
		repoPath      string
		bayes         bool
		playbookID    string
		selectRank    int
		showRejected  bool
		showEvidence  bool
		showScoring   bool
	)

	cmd := &cobra.Command{
		Use:   "trace [file]",
		Short: "Show deterministic rule-by-rule evaluation for a playbook",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateSelect(selectRank); err != nil {
				return err
			}
			resolvedFormat, resolvedJSON, err := resolveOutputSelection(format, jsonOut)
			if err != nil {
				return err
			}
			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			resolvedStore := resolveStoreSetting(history, noHistory, noStore, storePath)
			return app.NewService().Trace(input.Reader, input.Source, app.AnalyzeOptions{
				OutputOptions: app.OutputOptions{
					Top:          1,
					Select:       selectRank,
					Format:       resolvedFormat,
					JSON:         resolvedJSON,
					ShowRejected: showRejected,
					ShowEvidence: showEvidence,
					ShowScoring:  showScoring,
				},
				TraceOptions: app.TraceOptions{
					TraceEnabled:  true,
					TracePlaybook: playbookID,
				},
				ProviderOptions: app.ProviderOptions{
					GitContextEnabled: gitContext,
					GitSince:          gitSince,
					RepoPath:          repoPath,
				},
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
				BayesEnabled:     bayes,
				Store:            resolvedStore,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().StringVar(&format, "format", string(output.FormatTerminal), "output format: terminal|markdown|json")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(&history, "history", false, "read and write local history for this run")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Flags().BoolVar(&gitContext, "git", true, "enrich results with recent local git repository context (enabled by default; pass --git=false to disable)")
	cmd.Flags().StringVar(&gitSince, "since", "30d", "git history window for --git (for example 7d, 2w, 1 month ago)")
	cmd.Flags().StringVar(&repoPath, "repo", ".", "repository path to scan when --git is enabled")
	cmd.Flags().BoolVar(&bayes, "bayes", false, "rerank deterministic matches with the Bayesian-inspired scoring layer")
	cmd.Flags().StringVar(&playbookID, "playbook", "", "trace the named playbook even if it did not win the ranking")
	cmd.Flags().IntVar(&selectRank, "select", 0, "trace the Nth ranked result instead of the winner (1-based)")
	cmd.Flags().BoolVar(&showRejected, "show-rejected", false, "include competing candidates and rejection context")
	cmd.Flags().BoolVar(&showEvidence, "show-evidence", false, "include a raw evidence appendix")
	cmd.Flags().BoolVar(&showScoring, "show-scoring", false, "include scoring detail")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")
	return cmd
}
