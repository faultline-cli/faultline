package cli

import (
	"os"

	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/output"
)

func newInspectCommand() *cobra.Command {
	var (
		jsonOut       bool
		top           int
		mode          string
		format        string
		playbookDir   string
		playbookPacks []string
		noHistory     bool
		noStore       bool
		storePath     string
		bayes         bool
	)

	cmd := &cobra.Command{
		Use:   "inspect [path]",
		Short: "Inspect a repository tree for source-level failure risks",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputMode(mode); err != nil {
				return err
			}
			resolvedFormat, resolvedJSON, err := resolveOutputSelection(format, jsonOut)
			if err != nil {
				return err
			}
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			resolvedStore := firstNonEmpty(storePath, os.Getenv(storeEnv))
			if noHistory || noStore {
				resolvedStore = "off"
			}
			return app.NewService().Inspect(root, app.AnalyzeOptions{
				OutputOptions: app.OutputOptions{
					Top:    top,
					Mode:   output.Mode(mode),
					Format: resolvedFormat,
					JSON:   resolvedJSON,
				},
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
				Store:            resolvedStore,
				BayesEnabled:     bayes,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().IntVar(&top, "top", 1, "show top N ranked results")
	cmd.Flags().StringVar(&mode, "mode", string(output.ModeQuick), "output mode: quick|detailed")
	cmd.Flags().StringVar(&format, "format", string(output.FormatTerminal), "output format: terminal|markdown|json")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Flags().BoolVar(&bayes, "bayes", false, "rerank deterministic findings with the Bayesian-inspired scoring layer")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")
	return cmd
}
