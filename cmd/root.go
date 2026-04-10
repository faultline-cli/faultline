package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/cli"
	"faultline/internal/output"
	"faultline/internal/workflow"
)

// version is stamped at build time: -ldflags "-X main.version=x.y.z".
var version = "dev"

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "faultline",
		Short:         "Deterministic CI failure diagnosis from logs",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newAnalyzeCommand())
	cmd.AddCommand(newFixCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newExplainCommand())
	cmd.AddCommand(newWorkflowCommand())
	return cmd
}

func newAnalyzeCommand() *cobra.Command {
	var (
		jsonOut       bool
		top           int
		mode          string
		playbookDir   string
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
			if mode != string(output.ModeQuick) && mode != string(output.ModeDetailed) {
				return fmt.Errorf("--mode must be %q or %q", output.ModeQuick, output.ModeDetailed)
			}

			input, err := cli.ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			return app.RunAnalyze(input.Reader, input.Source, app.AnalyzeOptions{
				Top:               top,
				Mode:              output.Mode(mode),
				JSON:              jsonOut,
				CIAnnotations:     ciAnnotations,
				NoHistory:         noHistory,
				PlaybookDir:       playbookDir,
				GitContextEnabled: gitContext,
				GitSince:          gitSince,
				RepoPath:          repoPath,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON output")
	cmd.Flags().IntVar(&top, "top", 1, "show top N ranked results")
	cmd.Flags().StringVar(&mode, "mode", string(output.ModeQuick), "output mode: quick|detailed")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	cmd.Flags().BoolVar(&ciAnnotations, "ci-annotations", false, "emit GitHub Actions ::warning:: annotations")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&gitContext, "git", false, "enrich results with recent local git repository context")
	cmd.Flags().StringVar(&gitSince, "since", "30d", "git history window for --git (for example 7d, 2w, 1 month ago)")
	cmd.Flags().StringVar(&repoPath, "repo", ".", "repository path to scan when --git is enabled")
	return cmd
}

func newFixCommand() *cobra.Command {
	var (
		playbookDir string
		noHistory   bool
	)

	cmd := &cobra.Command{
		Use:   "fix [file]",
		Short: "Show fix steps for the top diagnosis",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := cli.ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			return app.RunFix(input.Reader, input.Source, app.AnalyzeOptions{
				Top:         1,
				NoHistory:   noHistory,
				PlaybookDir: playbookDir,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	return cmd
}

func newListCommand() *cobra.Command {
	var (
		category    string
		playbookDir string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available failure playbooks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.RunList(category, playbookDir, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "filter by failure category (e.g. auth, build, deploy)")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	return cmd
}

func newExplainCommand() *cobra.Command {
	var playbookDir string

	cmd := &cobra.Command{
		Use:   "explain <id>",
		Short: "Show full details for a failure playbook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.RunExplain(args[0], playbookDir, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	return cmd
}

func newWorkflowCommand() *cobra.Command {
	var (
		playbookDir string
		noHistory   bool
		gitContext  bool
		gitSince    string
		repoPath    string
		mode        string
		jsonOut     bool
	)

	cmd := &cobra.Command{
		Use:   "workflow [file]",
		Short: "Generate a deterministic local or agent-ready follow-up workflow",
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
			if mode != string(workflow.ModeLocal) && mode != string(workflow.ModeAgent) {
				return fmt.Errorf("--mode must be %q or %q", workflow.ModeLocal, workflow.ModeAgent)
			}

			input, err := cli.ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			return app.RunWorkflow(input.Reader, input.Source, app.AnalyzeOptions{
				Top:               1,
				NoHistory:         noHistory,
				PlaybookDir:       playbookDir,
				GitContextEnabled: gitContext,
				GitSince:          gitSince,
				RepoPath:          repoPath,
			}, workflow.Mode(mode), jsonOut, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&mode, "mode", string(workflow.ModeLocal), "workflow mode: local|agent")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON output")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&gitContext, "git", false, "enrich the workflow with recent local git repository context")
	cmd.Flags().StringVar(&gitSince, "since", "30d", "git history window for --git (for example 7d, 2w, 1 month ago)")
	cmd.Flags().StringVar(&repoPath, "repo", ".", "repository path to scan when --git is enabled")
	return cmd
}
