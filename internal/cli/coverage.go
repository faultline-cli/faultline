package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"faultline/internal/coverage"
)

// newCoverageCommand reports behavioral fixture evidence for the resolved
// playbook catalog.
func newCoverageCommand() *cobra.Command {
	var (
		playbookDir   string
		playbookPacks []string
		fixtureRoot   string
		jsonOut       bool
	)

	cmd := &cobra.Command{
		Use:   "coverage",
		Short: "Report playbook fixture evidence and catalog coverage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := coverage.Build(coverage.Options{
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
				FixtureRoot:      fixtureRoot,
			})
			if err != nil {
				return fmt.Errorf("build coverage report: %w", err)
			}
			if jsonOut {
				return coverage.WriteJSON(cmd.OutOrStdout(), report)
			}
			return coverage.WriteText(cmd.OutOrStdout(), report)
		},
	}

	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().StringVar(&fixtureRoot, "fixture-dir", "", "fixture corpus root or legacy .log fixture directory")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	return cmd
}
