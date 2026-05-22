package cli

import (
	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/output"
)

func newFixCommand() *cobra.Command {
	var (
		format        string
		playbookDir   string
		playbookPacks []string
		history       bool
		noHistory     bool
		noStore       bool
		storePath     string
		bayes         bool
		commandsOnly  bool
		withPrecons   bool
		withRisks     bool
	)

	cmd := &cobra.Command{
		Use:   "fix [file]",
		Short: "Show fix steps for the top diagnosis",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedFormat, err := validateOutputFormat(format)
			if err != nil {
				return err
			}
			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			resolvedStore := resolveStoreSetting(history, noHistory, noStore, storePath)
			return app.NewService().Fix(cmd.Context(), input.Reader, input.Source, app.AnalyzeOptions{
				OutputOptions: app.OutputOptions{
					Top:    1,
					Format: resolvedFormat,
				},
				ProviderOptions: app.ProviderOptions{
					GitContextEnabled: true,
				},
				PlaybookDir:          playbookDir,
				PlaybookPackDirs:     playbookPacks,
				Store:                resolvedStore,
				BayesEnabled:         bayes,
				FixCommandsOnly:      commandsOnly,
				FixWithPreconditions: withPrecons,
				FixWithRisks:         withRisks,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&format, "format", string(output.FormatTerminal), "output format: terminal|markdown|json")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(&history, "history", false, "read and write local history for this run")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Flags().BoolVar(&bayes, "bayes", false, "rerank deterministic matches with the Bayesian-inspired scoring layer")
	cmd.Flags().BoolVar(&commandsOnly, "commands-only", false, "show only runnable code blocks from fix steps")
	cmd.Flags().BoolVar(&withPrecons, "with-preconditions", false, "include preconditions section when present in fix steps")
	cmd.Flags().BoolVar(&withRisks, "with-risks", false, "include risks section when present in fix steps")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")
	return cmd
}
