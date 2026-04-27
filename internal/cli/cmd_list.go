package cli

import (
	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/output"
)

func newListCommand() *cobra.Command {
	var (
		category      string
		playbookDir   string
		playbookPacks []string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available playbooks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().List(category, playbookDir, playbookPacks, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "filter by category (for example auth, build, deploy)")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	return cmd
}

func newExplainCommand() *cobra.Command {
	var (
		format        string
		playbookDir   string
		playbookPacks []string
	)

	cmd := &cobra.Command{
		Use:   "explain <id>",
		Short: "Show full details for a playbook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedFormat, err := validateOutputFormat(format)
			if err != nil {
				return err
			}
			return app.NewService().Explain(args[0], playbookDir, playbookPacks, resolvedFormat, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&format, "format", string(output.FormatTerminal), "output format: terminal|markdown|json")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	return cmd
}
