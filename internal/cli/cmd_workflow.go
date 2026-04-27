package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/workflow"
	workflowexec "faultline/internal/workflow/execute"
)

func newWorkflowCommand() *cobra.Command {
	var (
		playbookDir   string
		playbookPacks []string
		noHistory     bool
		noStore       bool
		storePath     string
		gitContext    bool
		gitSince      string
		repoPath      string
		mode          string
		jsonOut       bool
		bayes         bool
	)

	cmd := &cobra.Command{
		Use:   "workflow [file]",
		Short: "Explain and execute deterministic remediation workflows",
		Long: joinLines(
			"Analyze a CI log or load a saved analysis artifact, then resolve the recommended typed remediation workflow.",
			"",
			"`faultline workflow <file>` remains a compatibility alias for `faultline workflow explain <file>`.",
		),
		Example: joinLines(
			"  faultline workflow build.log",
			"  faultline workflow explain build.log --json",
			"  faultline workflow apply build.log --dry-run",
			"  faultline workflow apply build.log --allow-environment-mutation",
		),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			if mode != "local" && mode != "agent" {
				return fmt.Errorf("--mode must be %q or %q", "local", "agent")
			}
			resolvedStore := firstNonEmpty(storePath, os.Getenv(storeEnv))
			if noHistory || noStore {
				resolvedStore = "off"
			}
			return app.NewService().Workflow(input.Reader, input.Source, app.AnalyzeOptions{
				OutputOptions: app.OutputOptions{
					Top: 1,
				},
				ProviderOptions: app.ProviderOptions{
					GitContextEnabled: gitContext,
					GitSince:          gitSince,
					RepoPath:          repoPath,
				},
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
				Store:            resolvedStore,
				BayesEnabled:     bayes,
			}, appWorkflowMode(mode), jsonOut, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&mode, "mode", "local", "workflow mode: local|agent")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Flags().BoolVar(&gitContext, "git", true, "enrich the workflow with recent local git repository context (enabled by default; pass --git=false to disable)")
	cmd.Flags().StringVar(&gitSince, "since", "30d", "git history window for --git (for example 7d, 2w, 1 month ago)")
	cmd.Flags().StringVar(&repoPath, "repo", ".", "repository path to scan when --git is enabled")
	cmd.Flags().BoolVar(&bayes, "bayes", false, "rerank deterministic matches with the Bayesian-inspired scoring layer before building the workflow")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")

	cmd.AddCommand(newWorkflowExplainCommand())
	cmd.AddCommand(newWorkflowApplyCommand())
	cmd.AddCommand(newWorkflowShowCommand())
	cmd.AddCommand(newWorkflowHistoryCommand())
	return cmd
}

func newWorkflowExplainCommand() *cobra.Command {
	var (
		playbookDir   string
		playbookPacks []string
		noHistory     bool
		noStore       bool
		storePath     string
		gitContext    bool
		gitSince      string
		repoPath      string
		jsonOut       bool
		bayes         bool
		workflowRef   string
	)
	cmd := &cobra.Command{
		Use:   "explain [file]",
		Short: "Explain the recommended remediation workflow",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()
			resolvedStore := firstNonEmpty(storePath, os.Getenv(storeEnv))
			if noHistory || noStore {
				resolvedStore = "off"
			}
			return app.NewService().WorkflowExplain(input.Reader, input.Source, app.AnalyzeOptions{
				OutputOptions: app.OutputOptions{
					Top: 1,
				},
				ProviderOptions: app.ProviderOptions{
					GitContextEnabled: gitContext,
					GitSince:          gitSince,
					RepoPath:          repoPath,
				},
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
				Store:            resolvedStore,
				BayesEnabled:     bayes,
			}, workflowRef, jsonOut, cmd.OutOrStdout())
		},
	}
	addWorkflowInputFlags(cmd, &playbookDir, &playbookPacks, &noHistory, &noStore, &storePath, &gitContext, &gitSince, &repoPath, &jsonOut, &bayes, &workflowRef)
	return cmd
}

func newWorkflowApplyCommand() *cobra.Command {
	var (
		playbookDir              string
		playbookPacks            []string
		noHistory                bool
		noStore                  bool
		storePath                string
		gitContext               bool
		gitSince                 string
		repoPath                 string
		jsonOut                  bool
		bayes                    bool
		workflowRef              string
		dryRun                   bool
		allowLocalMutation       bool
		allowEnvironmentMutation bool
		allowExternalSideEffect  bool
	)
	cmd := &cobra.Command{
		Use:   "apply [file]",
		Short: "Dry-run or apply the recommended remediation workflow",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()
			resolvedStore := firstNonEmpty(storePath, os.Getenv(storeEnv))
			if noHistory || noStore {
				resolvedStore = "off"
			}
			return app.NewService().WorkflowApply(input.Reader, input.Source, app.AnalyzeOptions{
				OutputOptions: app.OutputOptions{
					Top: 1,
				},
				ProviderOptions: app.ProviderOptions{
					GitContextEnabled: gitContext,
					GitSince:          gitSince,
					RepoPath:          repoPath,
				},
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
				Store:            resolvedStore,
				BayesEnabled:     bayes,
			}, workflowRef, dryRun, workflowexec.Policy{
				AllowLocalMutation:       allowLocalMutation,
				AllowEnvironmentMutation: allowEnvironmentMutation,
				AllowExternalSideEffect:  allowExternalSideEffect,
			}, jsonOut, cmd.OutOrStdout())
		},
	}
	addWorkflowInputFlags(cmd, &playbookDir, &playbookPacks, &noHistory, &noStore, &storePath, &gitContext, &gitSince, &repoPath, &jsonOut, &bayes, &workflowRef)
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "resolve the workflow and render the concrete execution plan without applying it")
	cmd.Flags().BoolVar(&allowLocalMutation, "allow-local-mutation", false, "allow steps that modify local repository files")
	cmd.Flags().BoolVar(&allowEnvironmentMutation, "allow-environment-mutation", false, "allow steps that mutate the local environment such as package installation")
	cmd.Flags().BoolVar(&allowExternalSideEffect, "allow-external-side-effect", false, "allow steps with external side effects")
	return cmd
}

func newWorkflowShowCommand() *cobra.Command {
	var (
		jsonOut   bool
		noHistory bool
		noStore   bool
		storePath string
	)
	cmd := &cobra.Command{
		Use:   "show <execution-id>",
		Short: "Show a persisted workflow execution record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().WorkflowShow(args[0], app.AnalyzeOptions{
				Store: func() string {
					if noHistory || noStore {
						return "off"
					}
					return firstNonEmpty(storePath, os.Getenv(storeEnv))
				}(),
			}, jsonOut, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading the local forensic store")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")
	return cmd
}

func newWorkflowHistoryCommand() *cobra.Command {
	var (
		jsonOut   bool
		noHistory bool
		noStore   bool
		storePath string
		limit     int
	)
	cmd := &cobra.Command{
		Use:   "history",
		Short: "List persisted workflow executions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().WorkflowHistory(app.AnalyzeOptions{
				Store: func() string {
					if noHistory || noStore {
						return "off"
					}
					return firstNonEmpty(storePath, os.Getenv(storeEnv))
				}(),
			}, limit, jsonOut, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading the local forensic store")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum number of executions to show")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")
	return cmd
}

func addWorkflowInputFlags(cmd *cobra.Command, playbookDir *string, playbookPacks *[]string, noHistory, noStore *bool, storePath *string, gitContext *bool, gitSince, repoPath *string, jsonOut, bayes *bool, workflowRef *string) {
	cmd.Flags().BoolVar(jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().StringVar(workflowRef, "workflow", "", "select a specific recommended workflow by id")
	cmd.Flags().StringVar(playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Flags().BoolVar(gitContext, "git", true, "enrich the workflow with recent local git repository context (enabled by default; pass --git=false to disable)")
	cmd.Flags().StringVar(gitSince, "since", "30d", "git history window for --git (for example 7d, 2w, 1 month ago)")
	cmd.Flags().StringVar(repoPath, "repo", ".", "repository path to scan when --git is enabled")
	cmd.Flags().BoolVar(bayes, "bayes", false, "rerank deterministic matches with the Bayesian-inspired scoring layer before building the workflow")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")
}

func appWorkflowMode(value string) workflow.Mode {
	if value == "agent" {
		return workflow.ModeAgent
	}
	return workflow.ModeLocal
}
