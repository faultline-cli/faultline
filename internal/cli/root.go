package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/output"
	"faultline/internal/workflow"
)

// NewRootCommand builds the Faultline CLI command tree.
func NewRootCommand(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "faultline",
		Short: "Deterministic CI failure diagnosis from logs",
		Long: strings.Join([]string{
			"Faultline analyzes CI logs and repository trees using deterministic playbook matching.",
			"It returns evidence-backed diagnoses, concrete fixes, and stable output for automation.",
		}, "\n\n"),
		Example: strings.Join([]string{
			"  faultline analyze build.log",
			"  cat build.log | faultline analyze --json",
			"  faultline inspect .",
			"  faultline workflow build.log --mode agent --git --repo .",
		}, "\n"),
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newAnalyzeCommand())
	cmd.AddCommand(newInspectCommand())
	cmd.AddCommand(newFixCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newExplainCommand())
	cmd.AddCommand(newWorkflowCommand())
	cmd.AddCommand(newPacksCommand())
	cmd.AddCommand(newFixturesCommand())
	return cmd
}

func validateOutputFormat(value string) error {
	format := output.Format(value)
	if !format.Valid() {
		return fmt.Errorf("--format must be %q or %q", output.FormatRaw, output.FormatMarkdown)
	}
	return nil
}

func validateOutputMode(value string) error {
	if value != string(output.ModeQuick) && value != string(output.ModeDetailed) {
		return fmt.Errorf("--mode must be %q or %q", output.ModeQuick, output.ModeDetailed)
	}
	return nil
}

func validateWorkflowMode(value string) error {
	if value != string(workflow.ModeLocal) && value != string(workflow.ModeAgent) {
		return fmt.Errorf("--mode must be %q or %q", workflow.ModeLocal, workflow.ModeAgent)
	}
	return nil
}

func newAnalyzeCommand() *cobra.Command {
	var (
		jsonOut       bool
		top           int
		mode          string
		format        string
		playbookDir   string
		playbookPacks []string
		ciAnnotations bool
		noHistory     bool
		gitContext    bool
		gitSince      string
		repoPath      string
	)

	cmd := &cobra.Command{
		Use:   "analyze [file]",
		Short: "Analyze a CI log from a file or stdin",
		Long: strings.Join([]string{
			"Analyze a CI log using deterministic playbook matching.",
			"",
			"When --git is enabled, Faultline also inspects recent local git history",
			"to correlate the likely failure with recently changed files, commits,",
			"churn hotspots, and simple hotfix or drift signals.",
		}, "\n"),
		Example: strings.Join([]string{
			"  faultline analyze build.log",
			"  faultline analyze build.log --mode detailed",
			"  faultline analyze build.log --git",
			"  faultline analyze build.log --git --since 30d --repo .",
			"  cat build.log | faultline analyze --json --git",
		}, "\n"),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputMode(mode); err != nil {
				return err
			}
			if err := validateOutputFormat(format); err != nil {
				return err
			}

			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			return app.NewService().Analyze(input.Reader, input.Source, app.AnalyzeOptions{
				Top:               top,
				Mode:              output.Mode(mode),
				Format:            output.Format(format),
				JSON:              jsonOut,
				CIAnnotations:     ciAnnotations,
				NoHistory:         noHistory,
				PlaybookDir:       playbookDir,
				PlaybookPackDirs:  playbookPacks,
				GitContextEnabled: gitContext,
				GitSince:          gitSince,
				RepoPath:          repoPath,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON output")
	cmd.Flags().IntVar(&top, "top", 1, "show top N ranked results")
	cmd.Flags().StringVar(&mode, "mode", string(output.ModeQuick), "output mode: quick|detailed")
	cmd.Flags().StringVar(&format, "format", string(output.FormatRaw), "human output format: raw|markdown")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "additional playbook pack directory; repeat to compose with the default catalog")
	cmd.Flags().BoolVar(&ciAnnotations, "ci-annotations", false, "emit GitHub Actions ::warning:: annotations")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&gitContext, "git", false, "enrich results with recent local git repository context")
	cmd.Flags().StringVar(&gitSince, "since", "30d", "git history window for --git (for example 7d, 2w, 1 month ago)")
	cmd.Flags().StringVar(&repoPath, "repo", ".", "repository path to scan when --git is enabled")
	return cmd
}

func newFixCommand() *cobra.Command {
	var (
		format        string
		playbookDir   string
		playbookPacks []string
		noHistory     bool
	)

	cmd := &cobra.Command{
		Use:   "fix [file]",
		Short: "Show fix steps for the top diagnosis",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFormat(format); err != nil {
				return err
			}
			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			return app.NewService().Fix(input.Reader, input.Source, app.AnalyzeOptions{
				Top:              1,
				Format:           output.Format(format),
				NoHistory:        noHistory,
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&format, "format", string(output.FormatRaw), "human output format: raw|markdown")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "additional playbook pack directory; repeat to compose with the default catalog")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	return cmd
}

func newInspectCommand() *cobra.Command {
	var (
		jsonOut       bool
		top           int
		mode          string
		format        string
		playbookDir   string
		playbookPacks []string
		noHistory     bool
	)

	cmd := &cobra.Command{
		Use:   "inspect [path]",
		Short: "Inspect a repository tree for source-level failure risks",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputMode(mode); err != nil {
				return err
			}
			if err := validateOutputFormat(format); err != nil {
				return err
			}
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			return app.NewService().Inspect(root, app.AnalyzeOptions{
				Top:              top,
				Mode:             output.Mode(mode),
				Format:           output.Format(format),
				JSON:             jsonOut,
				NoHistory:        noHistory,
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON output")
	cmd.Flags().IntVar(&top, "top", 1, "show top N ranked results")
	cmd.Flags().StringVar(&mode, "mode", string(output.ModeQuick), "output mode: quick|detailed")
	cmd.Flags().StringVar(&format, "format", string(output.FormatRaw), "human output format: raw|markdown")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "additional playbook pack directory; repeat to compose with the default catalog")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	return cmd
}

func newListCommand() *cobra.Command {
	var (
		category      string
		playbookDir   string
		playbookPacks []string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available failure playbooks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().List(category, playbookDir, playbookPacks, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "filter by failure category (e.g. auth, build, deploy)")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "additional playbook pack directory; repeat to compose with the default catalog")
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
		Short: "Show full details for a failure playbook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFormat(format); err != nil {
				return err
			}
			return app.NewService().Explain(args[0], playbookDir, playbookPacks, output.Format(format), cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&format, "format", string(output.FormatRaw), "human output format: raw|markdown")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "additional playbook pack directory; repeat to compose with the default catalog")
	return cmd
}

func newWorkflowCommand() *cobra.Command {
	var (
		playbookDir   string
		playbookPacks []string
		noHistory     bool
		gitContext    bool
		gitSince      string
		repoPath      string
		mode          string
		jsonOut       bool
	)

	cmd := &cobra.Command{
		Use:   "workflow [file]",
		Short: "Generate a deterministic follow-up workflow from a CI log",
		Long: strings.Join([]string{
			"Analyze a CI log and turn the top diagnosis into a deterministic follow-up workflow.",
			"",
			"`--mode local` prints a practical local triage checklist.",
			"`--mode agent` adds a structured agent prompt for code-assistant handoff.",
		}, "\n"),
		Example: strings.Join([]string{
			"  faultline workflow build.log",
			"  faultline workflow build.log --mode agent",
			"  faultline workflow build.log --mode agent --git --repo .",
			"  cat build.log | faultline workflow --json",
		}, "\n"),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateWorkflowMode(mode); err != nil {
				return err
			}

			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			return app.NewService().Workflow(input.Reader, input.Source, app.AnalyzeOptions{
				Top:               1,
				NoHistory:         noHistory,
				PlaybookDir:       playbookDir,
				PlaybookPackDirs:  playbookPacks,
				GitContextEnabled: gitContext,
				GitSince:          gitSince,
				RepoPath:          repoPath,
			}, workflow.Mode(mode), jsonOut, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&mode, "mode", string(workflow.ModeLocal), "workflow mode: local|agent")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON output")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "additional playbook pack directory; repeat to compose with the default catalog")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&gitContext, "git", false, "enrich the workflow with recent local git repository context")
	cmd.Flags().StringVar(&gitSince, "since", "30d", "git history window for --git (for example 7d, 2w, 1 month ago)")
	cmd.Flags().StringVar(&repoPath, "repo", ".", "repository path to scan when --git is enabled")
	return cmd
}

func newPacksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packs",
		Short: "Manage installed extra playbook packs",
		Long: strings.Join([]string{
			"Install and inspect extra playbook packs that should be loaded automatically.",
			"",
			"Installed packs live under ~/.faultline/packs so they survive binary upgrades",
			"and can be mounted into Docker containers using the same path convention.",
		}, "\n"),
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
		Short: "Install an extra playbook pack into the local Faultline directory",
		Example: strings.Join([]string{
			"  faultline packs install ./playbooks/company-pack",
			"  faultline packs install ./playbooks/extended-pack --force",
		}, "\n"),
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
		Short: "List locally installed extra playbook packs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().ListInstalledPacks(cmd.OutOrStdout())
		},
	}
}
