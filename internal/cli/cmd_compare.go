package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/output"
)

func newReplayCommand() *cobra.Command {
	var (
		jsonOut    bool
		top        int
		selectRank int
		mode       string
		format     string
		view       string
	)

	cmd := &cobra.Command{
		Use:   "replay <analysis.json>",
		Short: "Re-render a saved analysis artifact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputMode(mode); err != nil {
				return err
			}
			resolvedView, err := validateView(view)
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
			if resolvedView == output.ViewTrace {
				return fmt.Errorf("replay trace is not supported from analysis artifacts; replay a saved trace artifact or use `faultline trace` on the original log")
			}
			if resolvedJSON && resolvedView != output.ViewDefault {
				return fmt.Errorf("--view cannot be combined with --json")
			}

			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			return app.NewService().Replay(input.Reader, app.AnalyzeOptions{
				OutputOptions: app.OutputOptions{
					Top:    top,
					Select: selectRank,
					Mode:   output.Mode(mode),
					Format: resolvedFormat,
					View:   resolvedView,
					JSON:   resolvedJSON,
				},
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().IntVar(&top, "top", 1, "show top N ranked results")
	cmd.Flags().IntVar(&selectRank, "select", 0, "render only the Nth ranked result (1-based)")
	cmd.Flags().StringVar(&mode, "mode", string(output.ModeQuick), "output mode: quick|detailed")
	cmd.Flags().StringVar(&format, "format", string(output.FormatTerminal), "output format: terminal|markdown|json")
	cmd.Flags().StringVar(&view, "view", string(output.ViewDefault), "focused output view: summary|evidence|fix|raw|trace")
	return cmd
}
