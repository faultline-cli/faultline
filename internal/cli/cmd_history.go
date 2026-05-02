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
			return app.NewService().History(signatureHash, firstNonEmpty(storePath, os.Getenv(storeEnv)), limit, jsonOut, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().IntVar(&limit, "limit", 10, "maximum number of items to show per section")
	cmd.Flags().StringVar(&signatureHash, "signature", "", "show detailed history for one signature hash")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	return cmd
}

func newSignaturesCommand() *cobra.Command {
	var (
		jsonOut   bool
		limit     int
		storePath string
	)

	cmd := &cobra.Command{
		Use:   "signatures",
		Short: "List stored recurring signatures",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().Signatures(firstNonEmpty(storePath, os.Getenv(storeEnv)), limit, jsonOut, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum number of signatures to show")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Hidden = true
	return cmd
}

func newVerifyDeterminismCommand() *cobra.Command {
	var (
		jsonOut   bool
		storePath string
	)

	cmd := &cobra.Command{
		Use:   "verify-determinism [file]",
		Short: "Check whether one input hash has produced stable stored output",
		Long: joinLines(
			"Compute the deterministic input hash for a log and compare it with stored output hashes.",
			"",
			"This is an explicit verification surface over the local forensic store.",
		),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()
			return app.NewService().VerifyDeterminism(input.Reader, input.Source, firstNonEmpty(storePath, os.Getenv(storeEnv)), jsonOut, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Hidden = true
	return cmd
}
