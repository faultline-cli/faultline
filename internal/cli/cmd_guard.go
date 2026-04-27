package cli

import (
	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/output"
)

func newGuardCommand() *cobra.Command {
	var (
		jsonOut       bool
		top           int
		mode          string
		format        string
		playbookDir   string
		playbookPacks []string
		gitSince      string
	)

	cmd := &cobra.Command{
		Use:   "guard [path]",
		Short: "Run quiet high-confidence local prevention checks on changed files",
		Long: joinLines(
			"Inspect changed repository files and emit only high-confidence deterministic findings.",
			"",
			"Guard stays quiet when the worktree is clean or when no strong preventive signal is present.",
		),
		Args: cobra.MaximumNArgs(1),
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
			return app.NewService().Guard(root, app.AnalyzeOptions{
				OutputOptions: app.OutputOptions{
					Top:    top,
					Mode:   output.Mode(mode),
					Format: resolvedFormat,
					JSON:   resolvedJSON,
				},
				ProviderOptions: app.ProviderOptions{
					GitSince: gitSince,
				},
				Store:            "off",
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().IntVar(&top, "top", 3, "show up to N guard findings")
	cmd.Flags().StringVar(&mode, "mode", string(output.ModeQuick), "output mode: quick|detailed")
	cmd.Flags().StringVar(&format, "format", string(output.FormatTerminal), "output format: terminal|markdown|json")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().StringVar(&gitSince, "since", "30d", "git history window used for deterministic drift hints")
	return cmd
}
