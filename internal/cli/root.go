package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/workflow"
	workflowexec "faultline/internal/workflow/execute"
)

const experimentalProviderDeltaEnv = "FAULTLINE_EXPERIMENTAL_PROVIDER_DELTA"
const experimentalGitHubDeltaEnv = "FAULTLINE_EXPERIMENTAL_GITHUB_DELTA"
const storeEnv = "FAULTLINE_STORE"

// NewRootCommand builds the Faultline CLI command tree.
func NewRootCommand(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "faultline",
		Short: "Deterministic CI failure diagnosis from logs",
		Long: strings.Join([]string{
			"Faultline turns CI logs into deterministic diagnoses.",
			"It returns evidence-backed explanations, concrete fixes, and stable output for automation.",
			"",
			"The core release flow is: analyze a failing log, inspect the top playbook,",
			"and generate a deterministic follow-up workflow when you need handoff-ready output.",
		}, "\n\n"),
		Example: strings.Join([]string{
			"  faultline analyze build.log",
			"  cat build.log | faultline analyze --json",
			"  faultline workflow build.log --json --mode agent",
			"  faultline explain docker-auth",
			"  faultline list --category auth",
		}, "\n"),
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newAnalyzeCommand())
	cmd.AddCommand(newWorkflowCommand())
	cmd.AddCommand(newExplainCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newFixCommand())
	cmd.AddCommand(newCompareCommand())
	cmd.AddCommand(newReplayCommand())
	cmd.AddCommand(newTraceCommand())
	cmd.AddCommand(newInspectCommand())
	cmd.AddCommand(newGuardCommand())
	cmd.AddCommand(newPacksCommand())
	cmd.AddCommand(newHistoryCommand())
	cmd.AddCommand(newSignaturesCommand())
	cmd.AddCommand(newVerifyDeterminismCommand())
	cmd.AddCommand(newFixturesCommand())
	cmd.AddCommand(newCoverageCommand())
	return cmd
}

func validateOutputFormat(value string) (output.Format, error) {
	format, ok := output.ParseFormat(value)
	if !ok {
		return "", fmt.Errorf("--format must be %q, %q, or %q", output.FormatTerminal, output.FormatMarkdown, output.FormatJSON)
	}
	return format, nil
}

func validateOutputMode(value string) error {
	if value != string(output.ModeQuick) && value != string(output.ModeDetailed) {
		return fmt.Errorf("--mode must be %q or %q", output.ModeQuick, output.ModeDetailed)
	}
	return nil
}

func validateView(value string) (output.View, error) {
	view, ok := output.ParseView(value)
	if !ok {
		return "", fmt.Errorf("--view must be %q, %q, %q, %q, or %q", output.ViewSummary, output.ViewEvidence, output.ViewFix, output.ViewRaw, output.ViewTrace)
	}
	return view, nil
}

func validateSelect(value int) error {
	if value < 0 {
		return fmt.Errorf("--select must be 1 or greater")
	}
	return nil
}

func resolveOutputSelection(formatValue string, jsonOut bool) (output.Format, bool, error) {
	format, err := validateOutputFormat(formatValue)
	if err != nil {
		return "", false, err
	}
	if jsonOut {
		if format != output.FormatTerminal && format != output.FormatJSON {
			return "", false, fmt.Errorf("--json cannot be combined with --format %q", format)
		}
		format = output.FormatJSON
	}
	return format, format == output.FormatJSON, nil
}

func validateExperimentalDeltaProvider(provider string) error {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		return nil
	}
	if strings.TrimSpace(os.Getenv(experimentalProviderDeltaEnv)) == "1" {
		return nil
	}
	if strings.TrimSpace(os.Getenv(experimentalGitHubDeltaEnv)) == "1" {
		return nil
	}
	return fmt.Errorf("--delta-provider is experimental; set %s=1 (preferred) or %s=1 (legacy) to enable it explicitly", experimentalProviderDeltaEnv, experimentalGitHubDeltaEnv)
}

func validateWorkflowMode(value string) error {
	if value != "local" && value != "agent" {
		return fmt.Errorf("--mode must be %q or %q", "local", "agent")
	}
	return nil
}

func validateHookMode(value string) (model.HookMode, error) {
	mode := model.HookMode(strings.TrimSpace(value))
	switch mode {
	case "", model.HookModeOff, model.HookModeVerifyOnly, model.HookModeCollectOnly, model.HookModeSafe, model.HookModeFull:
		if mode == "" {
			return model.HookModeOff, nil
		}
		return mode, nil
	default:
		return "", fmt.Errorf("--hooks must be %q, %q, %q, %q, or %q", model.HookModeOff, model.HookModeVerifyOnly, model.HookModeCollectOnly, model.HookModeSafe, model.HookModeFull)
	}
}

func newAnalyzeCommand() *cobra.Command {
	var (
		jsonOut            bool
		top                int
		mode               string
		format             string
		view               string
		playbookDir        string
		playbookPacks      []string
		ciAnnotations      bool
		noHistory          bool
		noStore            bool
		storePath          string
		gitContext         bool
		gitSince           string
		repoPath           string
		bayes              bool
		traceEnabled       bool
		tracePlaybook      string
		selectRank         int
		showRejected       bool
		showEvidence       bool
		showScoring        bool
		deltaProvider      string
		githubRepo         string
		githubBranch       string
		githubRunID        int64
		gitlabProject      string
		gitlabBranch       string
		gitlabPipelineID   int64
		gitlabJobID        int64
		gitlabAPIBaseURL   string
		metricsHistoryFile string
		hookMode           string
		failOnSilent       bool
	)

	cmd := &cobra.Command{
		Use:   "analyze [file]",
		Short: "Analyze a CI log from a file or stdin",
		Long: strings.Join([]string{
			"Analyze a CI log and rank matching playbooks using deterministic rules.",
			"",
			"Faultline inspects recent local git history by default to correlate the",
			"likely failure with recently changed files, commits, churn hotspots,",
			"and simple hotfix or drift signals.",
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
			resolvedView, err := validateView(view)
			if err != nil {
				return err
			}
			if err := validateSelect(selectRank); err != nil {
				return err
			}
			if err := validateExperimentalDeltaProvider(deltaProvider); err != nil {
				return err
			}
			resolvedHookMode, err := validateHookMode(hookMode)
			if err != nil {
				return err
			}
			resolvedFormat, resolvedJSON, err := resolveOutputSelection(format, jsonOut)
			if err != nil {
				return err
			}
			if resolvedView == output.ViewTrace {
				traceEnabled = true
				resolvedView = output.ViewDefault
			}
			if resolvedJSON && resolvedView != output.ViewDefault {
				return fmt.Errorf("--view cannot be combined with --json")
			}

			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			return app.NewService().Analyze(input.Reader, input.Source, app.AnalyzeOptions{
				Top:                top,
				Mode:               output.Mode(mode),
				Format:             resolvedFormat,
				View:               resolvedView,
				JSON:               resolvedJSON,
				CIAnnotations:      ciAnnotations,
				NoHistory:          noHistory || noStore,
				PlaybookDir:        playbookDir,
				PlaybookPackDirs:   playbookPacks,
				GitContextEnabled:  gitContext,
				GitSince:           gitSince,
				RepoPath:           repoPath,
				BayesEnabled:       bayes,
				TraceEnabled:       traceEnabled,
				TracePlaybook:      tracePlaybook,
				Select:             selectRank,
				ShowRejected:       showRejected,
				ShowEvidence:       showEvidence,
				ShowScoring:        showScoring,
				DeltaProvider:      deltaProvider,
				GitHubRepository:   firstNonEmpty(githubRepo, os.Getenv("GITHUB_REPOSITORY")),
				GitHubBranch:       firstNonEmpty(githubBranch, os.Getenv("GITHUB_REF_NAME")),
				GitHubRunID:        firstInt64(githubRunID, os.Getenv("GITHUB_RUN_ID")),
				GitHubToken:        firstNonEmpty(os.Getenv("GITHUB_TOKEN"), os.Getenv("GH_TOKEN")),
				GitLabProject:      firstNonEmpty(gitlabProject, os.Getenv("CI_PROJECT_ID"), os.Getenv("CI_PROJECT_PATH")),
				GitLabBranch:       firstNonEmpty(gitlabBranch, os.Getenv("CI_COMMIT_REF_NAME")),
				GitLabPipelineID:   firstInt64(gitlabPipelineID, os.Getenv("CI_PIPELINE_ID")),
				GitLabJobID:        firstInt64(gitlabJobID, os.Getenv("CI_JOB_ID")),
				GitLabToken:        firstNonEmpty(os.Getenv("GITLAB_TOKEN"), os.Getenv("GITLAB_PRIVATE_TOKEN"), os.Getenv("CI_JOB_TOKEN")),
				GitLabAPIBaseURL:   firstNonEmpty(gitlabAPIBaseURL, os.Getenv("CI_API_V4_URL"), deriveGitLabAPIBaseURL(os.Getenv("CI_SERVER_URL"))),
				MetricsHistoryFile: metricsHistoryFile,
				Store:              firstNonEmpty(storePath, os.Getenv(storeEnv)),
				HookMode:           resolvedHookMode,
				FailOnSilent:       failOnSilent,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().IntVar(&top, "top", 1, "show top N ranked results")
	cmd.Flags().StringVar(&mode, "mode", string(output.ModeQuick), "output mode: quick|detailed")
	cmd.Flags().StringVar(&format, "format", string(output.FormatTerminal), "output format: terminal|markdown|json")
	cmd.Flags().StringVar(&view, "view", string(output.ViewDefault), "focused output view: summary|evidence|fix|raw|trace")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(&ciAnnotations, "ci-annotations", false, "emit GitHub Actions ::warning:: annotations")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Flags().BoolVar(&gitContext, "git", true, "enrich results with recent local git repository context (enabled by default; pass --git=false to disable)")
	cmd.Flags().StringVar(&gitSince, "since", "30d", "git history window for --git (for example 7d, 2w, 1 month ago)")
	cmd.Flags().StringVar(&repoPath, "repo", ".", "repository path to scan when --git is enabled")
	cmd.Flags().BoolVar(&bayes, "bayes", false, "rerank deterministic matches with the Bayesian-inspired scoring layer")
	cmd.Flags().BoolVar(&traceEnabled, "trace", false, "render a deterministic trace for the selected playbook")
	cmd.Flags().StringVar(&tracePlaybook, "trace-playbook", "", "render a deterministic trace for the named playbook")
	cmd.Flags().IntVar(&selectRank, "select", 0, "render only the Nth ranked result (1-based)")
	cmd.Flags().BoolVar(&showRejected, "show-rejected", false, "include competing candidates and rejection context in trace output")
	cmd.Flags().BoolVar(&showEvidence, "show-evidence", false, "include a raw evidence appendix when supported")
	cmd.Flags().BoolVar(&showScoring, "show-scoring", false, "include scoring detail when supported")
	cmd.Flags().StringVar(&deltaProvider, "delta-provider", "", "enable provider-backed failure delta resolution (currently: github-actions|gitlab-ci)")
	cmd.Flags().StringVar(&githubRepo, "github-repo", "", "GitHub repository for --delta-provider github-actions (defaults to GITHUB_REPOSITORY)")
	cmd.Flags().StringVar(&githubBranch, "github-branch", "", "GitHub branch for --delta-provider github-actions (defaults to GITHUB_REF_NAME)")
	cmd.Flags().Int64Var(&githubRunID, "github-run-id", 0, "GitHub Actions run ID for --delta-provider github-actions (defaults to GITHUB_RUN_ID)")
	cmd.Flags().StringVar(&gitlabProject, "gitlab-project", "", "GitLab project path or numeric project ID for --delta-provider gitlab-ci (defaults to CI_PROJECT_ID/CI_PROJECT_PATH)")
	cmd.Flags().StringVar(&gitlabBranch, "gitlab-branch", "", "GitLab ref for --delta-provider gitlab-ci (defaults to CI_COMMIT_REF_NAME)")
	cmd.Flags().Int64Var(&gitlabPipelineID, "gitlab-pipeline-id", 0, "GitLab pipeline ID for --delta-provider gitlab-ci (defaults to CI_PIPELINE_ID)")
	cmd.Flags().Int64Var(&gitlabJobID, "gitlab-job-id", 0, "GitLab job ID for --delta-provider gitlab-ci (defaults to CI_JOB_ID)")
	cmd.Flags().StringVar(&gitlabAPIBaseURL, "gitlab-api-base-url", "", "GitLab API v4 base URL for --delta-provider gitlab-ci (defaults to CI_API_V4_URL)")
	cmd.Flags().StringVar(&metricsHistoryFile, "history-file", "", "path to a JSONL history file for FPC and PHI computation")
	cmd.Flags().StringVar(&hookMode, "hooks", string(model.HookModeOff), "execute constrained playbook hooks: off|verify-only|collect-only|safe|full")
	cmd.Flags().BoolVar(&failOnSilent, "fail-on-silent", false, "exit non-zero when a silent failure is detected")
	_ = cmd.Flags().MarkHidden("delta-provider")
	_ = cmd.Flags().MarkHidden("github-repo")
	_ = cmd.Flags().MarkHidden("github-branch")
	_ = cmd.Flags().MarkHidden("github-run-id")
	_ = cmd.Flags().MarkHidden("gitlab-project")
	_ = cmd.Flags().MarkHidden("gitlab-branch")
	_ = cmd.Flags().MarkHidden("gitlab-pipeline-id")
	_ = cmd.Flags().MarkHidden("gitlab-job-id")
	_ = cmd.Flags().MarkHidden("gitlab-api-base-url")
	_ = cmd.Flags().MarkHidden("history-file")
	_ = cmd.Flags().MarkHidden("hooks")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")
	return cmd
}

func deriveGitLabAPIBaseURL(serverURL string) string {
	serverURL = strings.TrimSpace(serverURL)
	if serverURL == "" {
		return ""
	}
	return strings.TrimRight(serverURL, "/") + "/api/v4"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func firstInt64(values ...interface{}) int64 {
	for _, value := range values {
		switch v := value.(type) {
		case int64:
			if v != 0 {
				return v
			}
		case string:
			if strings.TrimSpace(v) == "" {
				continue
			}
			parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if err == nil && parsed != 0 {
				return parsed
			}
		}
	}
	return 0
}

func newFixCommand() *cobra.Command {
	var (
		format        string
		playbookDir   string
		playbookPacks []string
		noHistory     bool
		noStore       bool
		storePath     string
		bayes         bool
	)

	cmd := &cobra.Command{
		Use:   "fix [file]",
		Short: "Show fix steps for the top diagnosis",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedFormat, err := validateOutputFormat(format)
			if err != nil {
				return err
			}
			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			return app.NewService().Fix(input.Reader, input.Source, app.AnalyzeOptions{
				Top:               1,
				Format:            resolvedFormat,
				NoHistory:         noHistory || noStore,
				PlaybookDir:       playbookDir,
				PlaybookPackDirs:  playbookPacks,
				Store:             firstNonEmpty(storePath, os.Getenv(storeEnv)),
				GitContextEnabled: true,
				BayesEnabled:      bayes,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&format, "format", string(output.FormatTerminal), "output format: terminal|markdown|json")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Flags().BoolVar(&bayes, "bayes", false, "rerank deterministic matches with the Bayesian-inspired scoring layer")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")
	return cmd
}

func newCompareCommand() *cobra.Command {
	var (
		jsonOut bool
		format  string
	)

	cmd := &cobra.Command{
		Use:   "compare <left-analysis.json> <right-analysis.json>",
		Short: "Compare two saved analysis artifacts",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedFormat, resolvedJSON, err := resolveOutputSelection(format, jsonOut)
			if err != nil {
				return err
			}

			left, err := ReadInput(args[:1])
			if err != nil {
				return err
			}
			defer left.Close()
			right, err := ReadInput(args[1:])
			if err != nil {
				return err
			}
			defer right.Close()

			return app.NewService().Compare(left.Reader, right.Reader, app.AnalyzeOptions{
				Format: resolvedFormat,
				JSON:   resolvedJSON,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().StringVar(&format, "format", string(output.FormatTerminal), "output format: terminal|markdown|json")
	return cmd
}

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
				Top:    top,
				Select: selectRank,
				Mode:   output.Mode(mode),
				Format: resolvedFormat,
				View:   resolvedView,
				JSON:   resolvedJSON,
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

func newTraceCommand() *cobra.Command {
	var (
		jsonOut       bool
		format        string
		playbookDir   string
		playbookPacks []string
		noHistory     bool
		noStore       bool
		storePath     string
		gitContext    bool
		gitSince      string
		repoPath      string
		bayes         bool
		playbookID    string
		selectRank    int
		showRejected  bool
		showEvidence  bool
		showScoring   bool
		hookMode      string
	)

	cmd := &cobra.Command{
		Use:   "trace [file]",
		Short: "Show deterministic rule-by-rule evaluation for a playbook",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateSelect(selectRank); err != nil {
				return err
			}
			resolvedHookMode, err := validateHookMode(hookMode)
			if err != nil {
				return err
			}
			resolvedFormat, resolvedJSON, err := resolveOutputSelection(format, jsonOut)
			if err != nil {
				return err
			}
			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()

			return app.NewService().Trace(input.Reader, input.Source, app.AnalyzeOptions{
				Top:               1,
				Select:            selectRank,
				Format:            resolvedFormat,
				JSON:              resolvedJSON,
				NoHistory:         noHistory || noStore,
				PlaybookDir:       playbookDir,
				PlaybookPackDirs:  playbookPacks,
				GitContextEnabled: gitContext,
				GitSince:          gitSince,
				RepoPath:          repoPath,
				BayesEnabled:      bayes,
				TraceEnabled:      true,
				TracePlaybook:     playbookID,
				ShowRejected:      showRejected,
				ShowEvidence:      showEvidence,
				ShowScoring:       showScoring,
				Store:             firstNonEmpty(storePath, os.Getenv(storeEnv)),
				HookMode:          resolvedHookMode,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().StringVar(&format, "format", string(output.FormatTerminal), "output format: terminal|markdown|json")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Flags().BoolVar(&gitContext, "git", true, "enrich results with recent local git repository context (enabled by default; pass --git=false to disable)")
	cmd.Flags().StringVar(&gitSince, "since", "30d", "git history window for --git (for example 7d, 2w, 1 month ago)")
	cmd.Flags().StringVar(&repoPath, "repo", ".", "repository path to scan when --git is enabled")
	cmd.Flags().BoolVar(&bayes, "bayes", false, "rerank deterministic matches with the Bayesian-inspired scoring layer")
	cmd.Flags().StringVar(&playbookID, "playbook", "", "trace the named playbook even if it did not win the ranking")
	cmd.Flags().IntVar(&selectRank, "select", 0, "trace the Nth ranked result instead of the winner (1-based)")
	cmd.Flags().BoolVar(&showRejected, "show-rejected", false, "include competing candidates and rejection context")
	cmd.Flags().BoolVar(&showEvidence, "show-evidence", false, "include a raw evidence appendix")
	cmd.Flags().BoolVar(&showScoring, "show-scoring", false, "include scoring detail")
	cmd.Flags().StringVar(&hookMode, "hooks", string(model.HookModeOff), "execute constrained playbook hooks: off|verify-only|collect-only|safe|full")
	_ = cmd.Flags().MarkHidden("hooks")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")
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
		noStore       bool
		storePath     string
		bayes         bool
	)

	cmd := &cobra.Command{
		Use:   "inspect [path]",
		Short: "Inspect a repository tree for source-level failure risks",
		Args:  cobra.MaximumNArgs(1),
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
			return app.NewService().Inspect(root, app.AnalyzeOptions{
				Top:              top,
				Mode:             output.Mode(mode),
				Format:           resolvedFormat,
				JSON:             resolvedJSON,
				NoHistory:        noHistory || noStore,
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
				Store:            firstNonEmpty(storePath, os.Getenv(storeEnv)),
				BayesEnabled:     bayes,
			}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().IntVar(&top, "top", 1, "show top N ranked results")
	cmd.Flags().StringVar(&mode, "mode", string(output.ModeQuick), "output mode: quick|detailed")
	cmd.Flags().StringVar(&format, "format", string(output.FormatTerminal), "output format: terminal|markdown|json")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(&noHistory, "no-history", false, "skip reading and writing local history")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "disable the local forensic store")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Flags().BoolVar(&bayes, "bayes", false, "rerank deterministic findings with the Bayesian-inspired scoring layer")
	_ = cmd.Flags().MarkHidden("no-history")
	_ = cmd.Flags().MarkHidden("no-store")
	_ = cmd.Flags().MarkHidden("store")
	return cmd
}

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
		Long: strings.Join([]string{
			"Inspect changed repository files and emit only high-confidence deterministic findings.",
			"",
			"Guard stays quiet when the worktree is clean or when no strong preventive signal is present.",
		}, "\n"),
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
				Top:              top,
				Mode:             output.Mode(mode),
				Format:           resolvedFormat,
				JSON:             resolvedJSON,
				NoHistory:        true,
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
				GitSince:         gitSince,
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
		Long: strings.Join([]string{
			"Analyze a CI log or load a saved analysis artifact, then resolve the recommended typed remediation workflow.",
			"",
			"`faultline workflow <file>` remains a compatibility alias for `faultline workflow explain <file>`.",
		}, "\n"),
		Example: strings.Join([]string{
			"  faultline workflow build.log",
			"  faultline workflow explain build.log --json",
			"  faultline workflow apply build.log --dry-run",
			"  faultline workflow apply build.log --allow-environment-mutation",
		}, "\n"),
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
			return app.NewService().Workflow(input.Reader, input.Source, app.AnalyzeOptions{
				Top:               1,
				NoHistory:         noHistory || noStore,
				PlaybookDir:       playbookDir,
				PlaybookPackDirs:  playbookPacks,
				Store:             firstNonEmpty(storePath, os.Getenv(storeEnv)),
				GitContextEnabled: gitContext,
				GitSince:          gitSince,
				RepoPath:          repoPath,
				BayesEnabled:      bayes,
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
			return app.NewService().WorkflowExplain(input.Reader, input.Source, app.AnalyzeOptions{
				Top:               1,
				NoHistory:         noHistory || noStore,
				PlaybookDir:       playbookDir,
				PlaybookPackDirs:  playbookPacks,
				Store:             firstNonEmpty(storePath, os.Getenv(storeEnv)),
				GitContextEnabled: gitContext,
				GitSince:          gitSince,
				RepoPath:          repoPath,
				BayesEnabled:      bayes,
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
			return app.NewService().WorkflowApply(input.Reader, input.Source, app.AnalyzeOptions{
				Top:               1,
				NoHistory:         noHistory || noStore,
				PlaybookDir:       playbookDir,
				PlaybookPackDirs:  playbookPacks,
				Store:             firstNonEmpty(storePath, os.Getenv(storeEnv)),
				GitContextEnabled: gitContext,
				GitSince:          gitSince,
				RepoPath:          repoPath,
				BayesEnabled:      bayes,
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
				NoHistory: noHistory || noStore,
				Store:     firstNonEmpty(storePath, os.Getenv(storeEnv)),
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
				NoHistory: noHistory || noStore,
				Store:     firstNonEmpty(storePath, os.Getenv(storeEnv)),
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

func newPacksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packs",
		Short: "Manage optional playbook packs",
		Long: strings.Join([]string{
			"Install and inspect optional playbook packs that should be loaded automatically.",
			"",
			"Installed packs live under ~/.faultline/packs so they persist across CLI updates",
			"and can be mounted into Docker containers using the same path convention.",
			"The bundled catalog works on its own; packs are for extra or team-specific coverage.",
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
		Short: "Install a playbook pack into the local Faultline directory",
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
		Short: "List locally installed playbook packs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().ListInstalledPacks(cmd.OutOrStdout())
		},
	}
}

func newHistoryCommand() *cobra.Command {
	var (
		jsonOut       bool
		limit         int
		signatureHash string
		storePath     string
	)

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Inspect local recurrence history and quality summaries",
		Long: strings.Join([]string{
			"Read the local forensic store without changing diagnosis logic.",
			"",
			"By default this prints recurring signatures plus playbook and hook quality summaries.",
			"Use --signature to inspect one stored signature in detail.",
		}, "\n"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().History(signatureHash, firstNonEmpty(storePath, os.Getenv(storeEnv)), limit, jsonOut, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().IntVar(&limit, "limit", 10, "maximum number of items to show per section")
	cmd.Flags().StringVar(&signatureHash, "signature", "", "show detailed history for one signature hash")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	return cmd
}

func newSignaturesCommand() *cobra.Command {
	var (
		jsonOut   bool
		limit     int
		storePath string
	)

	cmd := &cobra.Command{
		Use:   "signatures",
		Short: "List stored recurring signatures",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().Signatures(firstNonEmpty(storePath, os.Getenv(storeEnv)), limit, jsonOut, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum number of signatures to show")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Hidden = true
	return cmd
}

func newVerifyDeterminismCommand() *cobra.Command {
	var (
		jsonOut   bool
		storePath string
	)

	cmd := &cobra.Command{
		Use:   "verify-determinism [file]",
		Short: "Check whether one input hash has produced stable stored output",
		Long: strings.Join([]string{
			"Compute the deterministic input hash for a log and compare it with stored output hashes.",
			"",
			"This is an explicit verification surface over the local forensic store.",
		}, "\n"),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := ReadInput(args)
			if err != nil {
				return err
			}
			defer input.Close()
			return app.NewService().VerifyDeterminism(input.Reader, input.Source, firstNonEmpty(storePath, os.Getenv(storeEnv)), jsonOut, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().StringVar(&storePath, "store", "", "configure the local forensic store: auto|off|/path/to/store.db")
	cmd.Hidden = true
	return cmd
}
