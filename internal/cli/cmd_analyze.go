package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/output"
)

func newAnalyzeCommand() *cobra.Command {
	var (
		jsonOut       bool
		top           int
		mode          string
		format        string
		view          string
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
		selectRank    int
		showEvidence  bool
		showScoring   bool
		failOnSilent  bool
	)

	cmd := &cobra.Command{
		Use:   "analyze [file]",
		Short: "Analyze a CI log from a file or stdin",
		Long: joinLines(
			"Analyze a CI log and rank matching playbooks using deterministic rules.",
			"",
			"Faultline inspects recent local git history by default to correlate the",
			"likely failure with recently changed files, commits, churn hotspots,",
			"and simple hotfix or drift signals.",
		),
		Example: joinLines(
			"  faultline analyze build.log",
			"  faultline analyze build.log --mode detailed",
			"  faultline analyze build.log --git",
			"  faultline analyze build.log --git --since 30d --repo .",
			"  cat build.log | faultline analyze --json --git",
		),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputMode(mode); err != nil {
				return err
			}
			resolvedView, err := validateAnalyzeView(view)
			if err != nil {
				return err
			}
			if err := validateSelect(selectRank); err != nil {
				return err
			}
			resolvedFormat, resolvedJSON, err := resolveOutputSelection(format, jsonOut)
			if err != nil {
				return err
			}
			if resolvedJSON && resolvedView != output.ViewDefault {
				return fmt.Errorf("--view cannot be combined with --json")
			}

			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			resolvedStore := resolveAnalyzeStoreSetting(noHistory, noStore, storePath)
			return app.NewService().Analyze(cmd.Context(), input.Reader, input.Source, app.AnalyzeOptions{
				OutputOptions: app.OutputOptions{
					Top:          top,
					Mode:         output.Mode(mode),
					Format:       resolvedFormat,
					View:         resolvedView,
					JSON:         resolvedJSON,
					Select:       selectRank,
					ShowEvidence: showEvidence,
					ShowScoring:  showScoring,
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
				History:          resolveStoreHistoryOutput(history, noHistory, noStore, storePath),
				FailOnSilent:     failOnSilent,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().IntVar(&top, "top", 1, "show top N ranked results")
	cmd.Flags().StringVar(&mode, "mode", string(output.ModeQuick), "output mode: quick|detailed")
	cmd.Flags().StringVar(&format, "format", string(output.FormatTerminal), "output format: terminal|markdown|json")
	cmd.Flags().StringVar(&view, "view", string(output.ViewDefault), "focused output view: summary|evidence|fix|raw")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(&history, "history", false, "read and write local history for this run")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Flags().BoolVar(&gitContext, "git", true, "enrich results with recent local git repository context (enabled by default; pass --git=false to disable)")
	cmd.Flags().StringVar(&gitSince, "since", "30d", "git history window for --git (for example 7d, 2w, 1 month ago)")
	cmd.Flags().StringVar(&repoPath, "repo", ".", "repository path to scan when --git is enabled")
	cmd.Flags().BoolVar(&bayes, "bayes", true, "rerank deterministic matches with the Bayesian-inspired scoring layer (enabled by default; pass --bayes=false to disable)")
	cmd.Flags().IntVar(&selectRank, "select", 0, "render only the Nth ranked result (1-based)")
	cmd.Flags().BoolVar(&showEvidence, "show-evidence", false, "include a raw evidence appendix when supported")
	cmd.Flags().BoolVar(&showScoring, "show-scoring", false, "include scoring detail when supported")
	cmd.Flags().BoolVar(&failOnSilent, "fail-on-silent", false, "exit non-zero when a silent failure is detected")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")
	return cmd
}

func validateAnalyzeView(value string) (output.View, error) {
	view, ok := output.ParseView(value)
	if !ok {
		return "", fmt.Errorf("--view must be %q, %q, %q, or %q", output.ViewSummary, output.ViewEvidence, output.ViewFix, output.ViewRaw)
	}
	if view == output.ViewTrace {
		return "", fmt.Errorf("--view trace was removed from analyze; use --show-evidence or --show-scoring for diagnostics")
	}
	return view, nil
}
