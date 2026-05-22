package cli

import (
	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/output"
)

func newBatchCommand() *cobra.Command {
	var (
		jsonOut      bool
		format       string
		playbookDir  string
		playbookPack []string
		history      bool
		noHistory    bool
		noStore      bool
		storePath    string
	)

	cmd := &cobra.Command{
		Use:   "batch <file> [file ...]",
		Short: "Analyze multiple CI logs and group results by failure pattern",
		Long: joinLines(
			"Analyze multiple CI log files and group matched diagnoses by failure",
			"pattern to identify recurring root causes across a build matrix.",
			"",
			"Each file is analyzed independently. Results are deduplicated by failure",
			"ID so you can see which pattern affected how many jobs and which files",
			"did not match any known playbook.",
			"",
			"Exit codes:",
			"  0  all logs matched a playbook",
			"  1  one or more logs did not match any playbook",
			"  2  error opening or processing a log file",
		),
		Example: joinLines(
			"  faultline batch build-1.log build-2.log build-3.log",
			"  faultline batch *.log --json",
			"  faultline batch matrix-*/build.log",
		),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedFormat, resolvedJSON, err := resolveOutputSelection(format, jsonOut)
			if err != nil {
				return err
			}
			opts := app.AnalyzeOptions{
				OutputOptions: app.OutputOptions{
					Format: resolvedFormat,
					JSON:   resolvedJSON,
				},
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPack,
				Store:            resolveStoreSetting(history, noHistory, noStore, storePath),
			}
			return app.NewService().Batch(cmd.Context(), args, opts, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON (batch.v1 schema)")
	cmd.Flags().StringVar(&format, "format", string(output.FormatTerminal), "output format: terminal|json|markdown")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override the default playbook directory")
	cmd.Flags().StringArrayVar(&playbookPack, "playbook-pack", nil, "extra playbook pack directory (repeatable)")
	cmd.Flags().BoolVar(&history, "history", false, "read local history for this run")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "disable local history store for this run")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")

	return cmd
}
