package cli

import (
	"os"

	"github.com/spf13/cobra"

	"faultline/internal/app"
)

func newHistoryCommand() *cobra.Command {
	var (
		jsonOut       bool
		limit         int
		signatureHash string
		storePath     string
	)

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Inspect local recurrence history and quality summaries",
		Long: joinLines(
			"Read the local forensic store without changing diagnosis logic.",
			"",
			"By default this prints recurring signatures plus playbook quality summaries.",
			"Use --signature to inspect one stored signature in detail.",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().History(cmd.Context(), signatureHash, firstNonEmpty(storePath, os.Getenv(storeEnv)), limit, jsonOut, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().IntVar(&limit, "limit", 10, "maximum number of items to show per section")
	cmd.Flags().StringVar(&signatureHash, "signature", "", "show detailed history for one signature hash")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	return cmd
}

