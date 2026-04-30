package cli

import (
	"os"

	"github.com/spf13/cobra"

	"faultline/internal/app"
)

func newReportCommand() *cobra.Command {
	var (
		jsonOut   bool
		storePath string
	)

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Show recurring local failure classes",
		Long: joinLines(
			"Read the local Faultline store and group stored analyze runs by failure class.",
			"",
			"The report is local-only and deterministic. It never contacts external services.",
		),
		Example: joinLines(
			"  faultline report",
			"  faultline report --json",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().Report(firstNonEmpty(storePath, os.Getenv(storeEnv)), jsonOut, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local store: auto|off|/path/to/store.db")
	_ = cmd.Flags().MarkHidden("store")
	return cmd
}
