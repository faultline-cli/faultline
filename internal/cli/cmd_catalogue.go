package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"faultline/internal/catalogue"
)

func newCatalogueCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "catalogue",
		Short:  "Generate and validate the public failure catalogue export",
		Hidden: true,
	}
	cmd.AddCommand(newCatalogueExportCommand())
	for _, child := range cmd.Commands() {
		child.Hidden = true
	}
	return cmd
}

func newCatalogueExportCommand() *cobra.Command {
	var (
		srcDir  string
		outDir  string
		repo    string
		commit  string
		version string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export the failure catalogue to a static-site-ready directory",
		Long: joinLines(
			"Generate a static-site-ready failure catalogue from the bundled playbooks.",
			"",
			"Produces:",
			"  <out>/failures/<slug>.md         — one Markdown file per failure (Astro frontmatter)",
			"  <out>/catalogue.json             — full failure index",
			"  <out>/catalogue.manifest.json    — provenance and generation metadata",
			"",
			"The output is deterministic: repeated runs with the same playbook set produce",
			"identical files (except for the generated_at timestamp in the manifest).",
		),
		Example: joinLines(
			"  faultline catalogue export --out ./catalogue",
			"  faultline catalogue export --src playbooks/bundled --out ./catalogue",
			"  faultline catalogue export --out ./catalogue --repo org/faultline --commit $GITHUB_SHA",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if outDir == "" {
				outDir = "catalogue"
			}
			if commit == "" {
				commit = resolveGitSHA()
			}
			if repo == "" {
				repo = firstNonEmpty(os.Getenv("GITHUB_REPOSITORY"), "faultline")
			}
			opts := catalogue.ExportOptions{
				SrcDir:           srcDir,
				OutDir:           outDir,
				SourceRepo:       repo,
				SourceCommit:     commit,
				GeneratorVersion: version,
			}
			if err := catalogue.Export(opts); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "catalogue export: wrote output to %s\n", outDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&srcDir, "src", "playbooks/bundled", "root of the playbook tree to export")
	cmd.Flags().StringVar(&outDir, "out", "catalogue", "output directory for the generated catalogue")
	cmd.Flags().StringVar(&repo, "repo", "", "source repository name stamped into the manifest (default: $GITHUB_REPOSITORY or \"faultline\")")
	cmd.Flags().StringVar(&commit, "commit", "", "source commit SHA stamped into the manifest (default: git rev-parse HEAD)")
	cmd.Flags().StringVar(&version, "version", "", "generator version stamped into the manifest")
	return cmd
}

// resolveGitSHA returns the current HEAD commit SHA by running git.
// Returns an empty string when git is unavailable or the working directory
// is not a git repository.
func resolveGitSHA() string {
	sha := strings.TrimSpace(os.Getenv("GITHUB_SHA"))
	if sha != "" {
		return sha
	}
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
