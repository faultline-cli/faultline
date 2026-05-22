package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"faultline/internal/app"
	"faultline/internal/fixtures"
	"faultline/internal/playbooks"
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
	cmd.AddCommand(newFixturesPatternsCommand())
	cmd.AddCommand(newFixturesPackCheckCommand())
	// Hide children explicitly so they are suppressed in tab-completion and any
	// programmatic traversal of cmd.Commands(), not just top-level help.
	// (The parent is already Hidden: true, but children remain visible to
	// subcommand-aware renderers unless individually hidden.)
	for _, child := range cmd.Commands() {
		child.Hidden = true
	}
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
			return app.NewService().FixturesIngest(cmd.Context(), root, fixtures.IngestOptions{
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
			}, baselinePath, jsonOut, checkBaseline, updateBaseline, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "repository root containing fixtures/")
	cmd.Flags().StringVar(&classValue, "class", string(fixtures.ClassAll), "fixture class to evaluate: minimal|real|noisy|all")
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
			}, jsonOut, failOnRegression, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "repository root containing fixtures/")
	cmd.Flags().StringVar(&classValue, "class", string(fixtures.ClassReal), "fixture class to evaluate: minimal|real|noisy|all")
	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "custom playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON output")
	cmd.Flags().BoolVar(&failOnRegression, "fail-on-regression", false, "exit non-zero when Bayes mode regresses any fixture's rank")
	return cmd
}

func newFixturesPatternsCommand() *cobra.Command {
	var (
		baselinePath   string
		updateBaseline bool
		verbose        bool
	)
	cmd := &cobra.Command{
		Use:   "patterns",
		Short: "Verify bundled playbook pattern conflicts against the checked-in baseline",
		Long: strings.Join([]string{
			"patterns compares the current bundled-playbook pattern-conflict report against a",
			"checked-in baseline file. It exits non-zero when the report has drifted.",
			"",
			"Use --update-baseline to rewrite the baseline after an intentional change.",
		}, "\n"),
		Example: strings.Join([]string{
			"  faultline fixtures patterns",
			"  faultline fixtures patterns --verbose",
			"  faultline fixtures patterns --update-baseline",
		}, "\n"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			pbs, err := playbooks.NewCatalog("").Load()
			if err != nil {
				return err
			}
			conflicts := playbooks.FindPatternConflicts(pbs)
			report := []byte(playbooks.FormatPatternConflicts(conflicts))
			if verbose {
				fmt.Fprint(cmd.OutOrStdout(), string(report))
			}
			if updateBaseline {
				if err := os.MkdirAll(filepath.Dir(baselinePath), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(baselinePath, report, 0o644); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "updated playbook review baseline: %s\n", baselinePath)
				return nil
			}
			baseline, err := os.ReadFile(baselinePath)
			if err != nil {
				return fmt.Errorf("read playbook review baseline: %w", err)
			}
			if !bytes.Equal(report, baseline) {
				fmt.Fprintf(cmd.ErrOrStderr(), "playbook pattern conflicts drifted from %s\n", baselinePath)
				fmt.Fprintln(cmd.ErrOrStderr(), "Run `make review-update` after reviewing intentional conflict changes.")
				if !verbose {
					fmt.Fprintln(cmd.ErrOrStderr(), "Use `make review-verbose` to print the full report.")
				}
				return fmt.Errorf("pattern conflict drift detected")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "playbook review passed (%d classified conflict patterns)\n", len(conflicts))
			return nil
		},
	}
	cmd.Flags().StringVar(&baselinePath, "baseline", "playbooks/bundled/pattern-conflicts.baseline.txt", "checked-in pattern conflict baseline")
	cmd.Flags().BoolVar(&updateBaseline, "update-baseline", false, "rewrite the pattern conflict baseline")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "print the full conflict report")
	return cmd
}

func newFixturesPackCheckCommand() *cobra.Command {
	var (
		packs  []string
		review bool
	)
	cmd := &cobra.Command{
		Use:   "pack-check",
		Short: "Validate and optionally review an extra playbook pack composed with the bundled catalog",
		Example: strings.Join([]string{
			"  faultline fixtures pack-check --pack ./playbooks/company-pack",
			"  faultline fixtures pack-check --pack ./playbooks/company-pack --review",
		}, "\n"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(packs) == 0 {
				return fmt.Errorf("at least one --pack path is required")
			}
			catalog := playbooks.NewCatalogWithOptions(playbooks.CatalogOptions{ExtraPackDirs: packs})
			pbs, err := catalog.Load()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "loaded %d playbooks across composed packs\n", len(pbs))
			if review {
				conflicts := playbooks.FindPatternConflicts(pbs)
				fmt.Fprintf(cmd.OutOrStdout(), "found %d pattern conflicts across composed packs\n", len(conflicts))
				fmt.Fprint(cmd.OutOrStdout(), playbooks.FormatPatternConflicts(conflicts))
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&packs, "pack", nil, "external playbook pack directory to compose with the bundled catalog; repeatable")
	cmd.Flags().BoolVar(&review, "review", false, "print deterministic overlap review for the composed catalog")
	return cmd
}
