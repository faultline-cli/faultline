package cli

import (
	"github.com/spf13/cobra"

	"faultline/internal/app"
)

func newPacksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packs",
		Short: "Manage optional playbook packs",
		Long: joinLines(
			"Install and inspect optional playbook packs that should be loaded automatically.",
			"",
			"Installed packs live under ~/.faultline/packs so they persist across CLI updates",
			"and can be mounted into Docker containers using the same path convention.",
			"The bundled catalog works on its own; packs are for extra or team-specific coverage.",
		),
	}
	cmd.AddCommand(newPacksInstallCommand())
	cmd.AddCommand(newPacksListCommand())
	return cmd
}

func newPacksInstallCommand() *cobra.Command {
	var (
		name  string
		force bool
	)

	cmd := &cobra.Command{
		Use:   "install <dir>",
		Short: "Install a playbook pack into the local Faultline directory",
		Example: joinLines(
			"  faultline packs install ./playbooks/company-pack",
			"  faultline packs install ./playbooks/extended-pack --force",
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().InstallPack(args[0], name, force, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "installed pack name override")
	cmd.Flags().BoolVar(&force, "force", false, "replace an existing installed pack with the same name")
	return cmd
}

func newPacksListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List locally installed playbook packs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().ListInstalledPacks(cmd.OutOrStdout())
		},
	}
}
