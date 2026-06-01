package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"faultline/internal/app"
)

func newSyncCommand() *cobra.Command {
	var (
		projectID string
		source    string
		branch    string
		commitSHA string
	)

	cmd := &cobra.Command{
		Use:    "sync [file]",
		Short:  "Push a failure artifact to Faultline Teams",
		Hidden: true,
		Long: joinLines(
			"Push the JSON output of 'faultline analyze --json' to Faultline Teams.",
			"",
			"Reads from a file argument or stdin. The artifact is deduplicated server-side;",
			"sending the same failure again increments the occurrence count without",
			"creating a duplicate record.",
			"",
			"Authentication: credentials stored by 'faultline auth login' are used",
			"automatically. In CI, set FAULTLINE_TOKEN and FAULTLINE_PROJECT instead.",
		),
		Example: joinLines(
			"  faultline analyze build.log --json | faultline sync --project $FAULTLINE_PROJECT",
			"  faultline sync --project proj123 result.json",
			"  faultline sync --project proj123 --branch main --commit-sha abc1234 result.json",
			"  FAULTLINE_TOKEN=ft_abc123 faultline sync --project $FAULTLINE_PROJECT result.json",
		),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := syncInput(cmd, args)
			if err != nil {
				return err
			}
			defer r.Close()

			return app.NewService().Sync(cmd.Context(), r, app.SyncOptions{
				ProjectID: projectID,
				Source:    source,
				Branch:    branch,
				CommitSHA: commitSHA,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&projectID, "project", "", "project ID (or set FAULTLINE_PROJECT)")
	cmd.Flags().StringVar(&source, "source", "", `CI provider label, e.g. "github-actions" or "local"`)
	cmd.Flags().StringVar(&branch, "branch", "", "git branch (overrides artifact.environment.branch)")
	cmd.Flags().StringVar(&commitSHA, "commit-sha", "", "commit SHA (overrides artifact.environment.commit_sha)")
	return cmd
}

func syncInput(cmd *cobra.Command, args []string) (io.ReadCloser, error) {
	if len(args) >= 1 {
		f, err := os.Open(args[0])
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", args[0], err)
		}
		return f, nil
	}
	in := cmd.InOrStdin()
	if f, ok := in.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return nil, fmt.Errorf("inspect stdin: %w", err)
		}
		if stat.Mode()&os.ModeCharDevice != 0 {
			return nil, fmt.Errorf("no artifact provided; pass a file path or pipe from 'faultline analyze --json'")
		}
	}
	return io.NopCloser(in), nil
}
