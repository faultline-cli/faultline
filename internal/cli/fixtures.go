package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/fixtures"
)

func newFixturesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "fixtures",
		Short:  "Manage minimal, staging, and real fixture corpora",
		Hidden: true,
	}
	cmd.AddCommand(newFixturesIngestCommand())
	cmd.AddCommand(newFixturesReviewCommand())
	cmd.AddCommand(newFixturesPromoteCommand())
	cmd.AddCommand(newFixturesStatsCommand())
	cmd.AddCommand(newFixturesSanitizeCommand())
	cmd.AddCommand(newFixturesCompareModesCommand())
	return cmd
}

func newFixturesIngestCommand() *cobra.Command {
	var (
		root    string
		adapter string
		urls    []string
		jsonOut bool
		force   bool
	)
	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Fetch public CI failure snippets into fixtures/staging",
		Example: strings.Join([]string{
			"  faultline fixtures ingest --adapter github-issue --url https://github.com/owner/repo/issues/123",
			"  faultline fixtures ingest --adapter gitlab-issue --url https://gitlab.com/group/project/-/issues/456",
			"  faultline fixtures ingest --adapter stackexchange-question --url https://stackoverflow.com/questions/12345/example",
			"  faultline fixtures ingest --adapter discourse-topic --url https://meta.discourse.org/t/example/12345",
			"  faultline fixtures ingest --adapter reddit-post --url https://www.reddit.com/r/docker/comments/1fbi7v2/ssh_docker_daemon/",
		}, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if adapter == "" {
				return fmt.Errorf("--adapter is required")
			}
			if len(urls) == 0 {
				return fmt.Errorf("at least one --url is required")
			}
			return app.NewService().FixturesIngest(root, fixtures.IngestOptions{
				Adapter: adapter,
				URLs:    urls,
				Force:   force,
			}, jsonOut, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "repository root containing fixtures/")
	cmd.Flags().StringVar(&adapter, "adapter", "", "source adapter: github-issue|gitlab-issue|stackexchange-question|discourse-topic|reddit-post")
	cmd.Flags().StringSliceVar(&urls, "url", nil, "public issue URL to ingest; repeat for batches")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON output")
	cmd.Flags().BoolVar(&force, "force", false, "write fixtures even when an existing fingerprint matches")
	return cmd
}

func newFixturesReviewCommand() *cobra.Command {
	var (
		root          string
		playbookDir   string
		playbookPacks []string
		jsonOut       bool
	)
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Review staging fixtures with predicted matches and duplicate hints",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().FixturesReview(root, fixtures.EvaluateOptions{
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
				NoHistory:        true,
			}, jsonOut, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "repository root containing fixtures/")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON output")
	return cmd
}

func newFixturesPromoteCommand() *cobra.Command {
	var (
		root             string
		expectedPlaybook string
		topN             int
		expectedStage    string
		strictTop1       bool
		disallowed       []string
		minConfidence    float64
		keepStaging      bool
	)
	cmd := &cobra.Command{
		Use:   "promote <staging-id> [<staging-id>...]",
		Short: "Promote reviewed staging fixtures into fixtures/real",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if expectedPlaybook == "" {
				return fmt.Errorf("--expected-playbook is required")
			}
			return app.NewService().FixturesPromote(root, args, fixtures.PromoteOptions{
				ExpectedPlaybook:    expectedPlaybook,
				TopN:                topN,
				ExpectedStage:       expectedStage,
				StrictTop1:          strictTop1,
				DisallowedPlaybooks: disallowed,
				MinConfidence:       minConfidence,
				KeepStaging:         keepStaging,
			}, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "repository root containing fixtures/")
	cmd.Flags().StringVar(&expectedPlaybook, "expected-playbook", "", "expected top playbook ID for promoted fixtures")
	cmd.Flags().IntVar(&topN, "top-n", 3, "maximum acceptable rank for the expected playbook")
	cmd.Flags().StringVar(&expectedStage, "expected-stage", "", "expected inferred stage for the promoted fixture")
	cmd.Flags().BoolVar(&strictTop1, "strict-top-1", false, "require the expected playbook to remain the top result")
	cmd.Flags().StringSliceVar(&disallowed, "disallow", nil, "disallowed playbook ID; repeat to bound false positives")
	cmd.Flags().Float64Var(&minConfidence, "min-confidence", 0.55, "minimum acceptable confidence before a match is reported as weak")
	cmd.Flags().BoolVar(&keepStaging, "keep-staging", false, "leave the original staging fixture in place after promotion")
	return cmd
}

func newFixturesStatsCommand() *cobra.Command {
	var (
		root           string
		classValue     string
		playbookDir    string
		playbookPacks  []string
		baselinePath   string
		jsonOut        bool
		checkBaseline  bool
		updateBaseline bool
	)
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Run deterministic regression stats across minimal, real, or combined corpora",
		RunE: func(cmd *cobra.Command, args []string) error {
			class, err := fixtures.ParseClass(classValue)
			if err != nil {
				return err
			}
			if baselinePath == "" {
				baselinePath = "fixtures/real/baseline.json"
			}
			return app.NewService().FixturesStats(root, class, fixtures.EvaluateOptions{
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
				NoHistory:        true,
			}, baselinePath, jsonOut, checkBaseline, updateBaseline, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "repository root containing fixtures/")
	cmd.Flags().StringVar(&classValue, "class", string(fixtures.ClassAll), "fixture class to evaluate: minimal|real|all")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().StringVar(&baselinePath, "baseline", "fixtures/real/baseline.json", "baseline snapshot path")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON output")
	cmd.Flags().BoolVar(&checkBaseline, "check-baseline", false, "fail if the current report regresses from the baseline snapshot")
	cmd.Flags().BoolVar(&updateBaseline, "update-baseline", false, "write the current report metrics to the baseline snapshot")
	return cmd
}

func newFixturesSanitizeCommand() *cobra.Command {
	var (
		root    string
		dryRun  bool
		jsonOut bool
	)
	cmd := &cobra.Command{
		Use:   "sanitize <staging-id> [<staging-id>...]",
		Short: "Mask secrets and sensitive patterns in staging fixtures before promotion",
		Long: strings.Join([]string{
			"Sanitize applies deterministic masking rules to the raw_log and normalized_log",
			"fields of the named staging fixture(s). Masked patterns include GitHub tokens,",
			"AWS keys, Authorization header values, URL credentials, credential key=value",
			"pairs, JWT tokens, PEM-encoded private keys, and email addresses.",
			"",
			"Sanitization is not a substitute for manual review. Always inspect the results",
			"before promoting fixtures into fixtures/real/.",
		}, "\n"),
		Example: strings.Join([]string{
			"  faultline fixtures sanitize staging-abc123",
			"  faultline fixtures sanitize staging-abc123 staging-def456 --dry-run",
			"  faultline fixtures sanitize staging-abc123 --json",
		}, "\n"),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.NewService().FixturesSanitize(root, args, fixtures.SanitizeOptions{
				DryRun: dryRun,
			}, jsonOut, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "repository root containing fixtures/")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report what would be replaced without modifying files")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON output")
	return cmd
}

func newFixturesCompareModesCommand() *cobra.Command {
	var (
		root             string
		classValue       string
		playbookDir      string
		playbookPacks    []string
		jsonOut          bool
		failOnRegression bool
	)
	cmd := &cobra.Command{
		Use:   "compare-modes",
		Short: "Compare baseline vs Bayes ranking across the fixture corpus",
		Long: strings.Join([]string{
			"compare-modes runs two evaluations over the same fixture corpus — one with the",
			"deterministic baseline scorer and one with the Bayesian-inspired reranker — and",
			"reports the per-fixture rank changes, aggregate rate deltas, and any regressions.",
			"",
			"Use this before promoting --bayes to a default or release-gated path.",
		}, "\n"),
		Example: strings.Join([]string{
			"  faultline fixtures compare-modes",
			"  faultline fixtures compare-modes --class real --fail-on-regression",
			"  faultline fixtures compare-modes --json",
		}, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			class, err := fixtures.ParseClass(classValue)
			if err != nil {
				return err
			}
			return app.NewService().FixturesCompareModes(root, class, fixtures.EvaluateOptions{
				PlaybookDir:      playbookDir,
				PlaybookPackDirs: playbookPacks,
				NoHistory:        true,
			}, jsonOut, failOnRegression, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "repository root containing fixtures/")
	cmd.Flags().StringVar(&classValue, "class", string(fixtures.ClassReal), "fixture class to evaluate: minimal|real|all")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON output")
	cmd.Flags().BoolVar(&failOnRegression, "fail-on-regression", false, "exit non-zero when Bayes mode regresses any fixture's rank")
	return cmd
}
