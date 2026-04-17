package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"text/tabwriter"
	"time"

	"faultline/internal/detectors"
	"faultline/internal/engine"
	"faultline/internal/fixtures"
	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/playbooks"
	"faultline/internal/renderer"
	"faultline/internal/repo"
	"faultline/internal/workflow"
)

// Service owns app-level orchestration for CLI commands.
type Service struct{}

var ErrGuardFindings = errors.New("guard findings emitted")

// guardMinConfidence and guardMinScore are the thresholds used by the guard
// command to filter source-detector results down to high-confidence findings
// only. Lower values increase noise; higher values reduce recall.
const (
	guardMinConfidence = 0.75
	guardMinScore      = 3.5
)

// NewService returns the default CLI application service.
func NewService() Service {
	return Service{}
}

// Analyze performs log analysis and writes formatted output to w.
func (Service) Analyze(r io.Reader, source string, opts AnalyzeOptions, w io.Writer) error {
	a, err := analyzeLog(r, source, opts)
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}
	return writeAnalysis(a, opts, w)
}

// Fix performs log analysis and writes only the ranked fix steps to w.
func (Service) Fix(r io.Reader, source string, opts AnalyzeOptions, w io.Writer) error {
	a, err := analyzeLog(r, source, opts)
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}
	if opts.JSON || opts.Format == output.FormatJSON {
		data, err := output.FormatAnalysisJSON(a, 1)
		if err != nil {
			return err
		}
		_, werr := fmt.Fprint(w, data)
		return werr
	}
	if opts.Format == output.FormatMarkdown {
		_, werr := fmt.Fprint(w, output.FormatFixMarkdown(a))
		return werr
	}
	_, werr := fmt.Fprint(w, output.FormatFix(a, renderer.DetectOptions(w)))
	return werr
}

// List loads all playbooks and writes a formatted list to w.
func (Service) List(category, playbookDir string, playbookPacks []string, w io.Writer) error {
	pbs, err := engine.New(engine.Options{
		PlaybookDir:      playbookDir,
		PlaybookPackDirs: playbookPacks,
		NoHistory:        true,
	}).List()
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, output.FormatPlaybookList(pbs, category, renderer.DetectOptions(w)))
	return err
}

// Explain fetches a single playbook by id and writes its details to w.
func (Service) Explain(id, playbookDir string, playbookPacks []string, format output.Format, w io.Writer) error {
	pb, err := engine.New(engine.Options{
		PlaybookDir:      playbookDir,
		PlaybookPackDirs: playbookPacks,
		NoHistory:        true,
	}).Explain(id)
	if err != nil {
		return err
	}
	if format == output.FormatMarkdown {
		_, err = fmt.Fprint(w, output.FormatPlaybookDetailsMarkdown(pb))
		return err
	}
	if format == output.FormatJSON {
		data, err := output.FormatPlaybookDetailsJSON(pb)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, data)
		return err
	}
	_, err = fmt.Fprint(w, output.FormatPlaybookDetails(pb, renderer.DetectOptions(w)))
	return err
}

// ListInstalledPacks prints the user-installed extra packs.
func (Service) ListInstalledPacks(w io.Writer) error {
	packs, err := playbooks.ListInstalledPacks()
	if err != nil {
		return err
	}
	if len(packs) == 0 {
		_, err := fmt.Fprintln(w, "No installed playbook packs.")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tPLAYBOOKS\tPATH"); err != nil {
		return err
	}
	for _, pack := range packs {
		if _, err := fmt.Fprintf(tw, "%s\t%d\t%s\n", pack.Name, pack.PlaybookCount, pack.Root); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// InstallPack installs a playbook pack into the user's persistent Faultline directory.
func (Service) InstallPack(srcDir, name string, force bool, w io.Writer) error {
	pack, err := playbooks.InstallPack(srcDir, name, force)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "Installed pack %s with %d playbooks at %s\n", pack.Name, pack.PlaybookCount, pack.Root)
	return err
}

// Workflow analyzes the log and emits a deterministic follow-up workflow.
func (Service) Workflow(r io.Reader, source string, opts AnalyzeOptions, mode workflow.Mode, jsonOut bool, w io.Writer) error {
	a, err := analyzeLog(r, source, opts)
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}

	plan := workflow.BuildWithOptions(a, mode, workflow.BuildOptions{
		RepoPath: opts.RepoPath,
	})
	if jsonOut {
		data, err := output.FormatWorkflowJSON(plan)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, data)
		return err
	}

	_, err = fmt.Fprint(w, output.FormatWorkflowText(plan))
	return err
}

// Inspect scans a repository tree with source-detector playbooks.
func (Service) Inspect(root string, opts AnalyzeOptions, w io.Writer) error {
	a, err := engine.New(engine.Options{
		PlaybookDir:      opts.PlaybookDir,
		PlaybookPackDirs: opts.PlaybookPackDirs,
		NoHistory:        opts.NoHistory,
		GitSince:         opts.GitSince,
		RepoPath:         opts.RepoPath,
		BayesEnabled:     opts.BayesEnabled,
	}).AnalyzeRepository(root, detectors.ChangeSet{})
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}
	return writeAnalysis(a, opts, w)
}

// Guard inspects changed repository files and only emits high-confidence findings.
func (Service) Guard(root string, opts AnalyzeOptions, w io.Writer) error {
	scanner, err := repo.NewScanner(root)
	if err != nil {
		return writeGuardNoFindings(root, opts, w)
	}
	changeSet, err := repo.LoadWorktreeChangeSet(scanner)
	if err != nil {
		return err
	}
	if len(changeSet.ChangedFiles) == 0 {
		return writeGuardNoFindings(scanner.Root, opts, w)
	}

	a, err := engine.New(engine.Options{
		PlaybookDir:      opts.PlaybookDir,
		PlaybookPackDirs: opts.PlaybookPackDirs,
		NoHistory:        true,
		GitSince:         opts.GitSince,
		RepoPath:         scanner.Root,
		BayesEnabled:     true,
	}).AnalyzeRepository(scanner.Root, changeSet)
	if errors.Is(err, engine.ErrNoInput) || errors.Is(err, engine.ErrNoMatch) {
		return writeGuardNoFindings(scanner.Root, opts, w)
	}
	if err != nil {
		return err
	}

	filtered := guardFindings(a, opts.Top)
	if len(filtered.Results) == 0 {
		return writeGuardNoFindings(scanner.Root, opts, w)
	}
	if err := writeAnalysis(filtered, opts, w); err != nil {
		return err
	}
	return ErrGuardFindings
}

func analyzeLog(r io.Reader, source string, opts AnalyzeOptions) (*model.Analysis, error) {
	a, err := engine.New(engine.Options{
		PlaybookDir:       opts.PlaybookDir,
		PlaybookPackDirs:  opts.PlaybookPackDirs,
		NoHistory:         opts.NoHistory,
		GitContextEnabled: opts.GitContextEnabled,
		GitSince:          opts.GitSince,
		RepoPath:          opts.RepoPath,
		BayesEnabled:      opts.BayesEnabled,
		DeltaProvider:     opts.DeltaProvider,
		GitHubRepository:  opts.GitHubRepository,
		GitHubBranch:      opts.GitHubBranch,
		GitHubRunID:       opts.GitHubRunID,
		GitHubToken:       opts.GitHubToken,
	}).AnalyzeReader(r)
	if a != nil {
		a.Source = source
	}
	return a, err
}

func guardFindings(a *model.Analysis, top int) *model.Analysis {
	if a == nil {
		return &model.Analysis{Results: []model.Result{}}
	}
	filtered := make([]model.Result, 0, len(a.Results))
	for _, result := range a.Results {
		if result.Confidence < guardMinConfidence {
			continue
		}
		if result.Score < guardMinScore {
			continue
		}
		filtered = append(filtered, result)
	}
	if top > 0 && len(filtered) > top {
		filtered = filtered[:top]
	}
	return &model.Analysis{
		Results:     filtered,
		Context:     a.Context,
		Fingerprint: a.Fingerprint,
		Source:      a.Source,
		RepoContext: a.RepoContext,
		Delta:       a.Delta,
	}
}

func writeGuardNoFindings(root string, opts AnalyzeOptions, w io.Writer) error {
	if opts.JSON || opts.Format == output.FormatJSON {
		return writeAnalysis(&model.Analysis{
			Results: []model.Result{},
			Source:  root,
		}, opts, w)
	}
	return nil
}

func (Service) FixturesIngest(root string, opts fixtures.IngestOptions, jsonOut bool, w io.Writer) error {
	layout, err := fixtures.ResolveLayout(root)
	if err != nil {
		return err
	}
	result, err := fixtures.Ingest(context.Background(), layout, opts)
	if err != nil {
		return err
	}
	formatted, err := fixtures.FormatIngestResult(result, jsonOut)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, formatted)
	return err
}

func (Service) FixturesReview(root string, opts fixtures.EvaluateOptions, jsonOut bool, w io.Writer) error {
	layout, err := fixtures.ResolveLayout(root)
	if err != nil {
		return err
	}
	report, err := fixtures.Review(layout, opts)
	if err != nil {
		return err
	}
	formatted, err := fixtures.FormatReviewReport(report, jsonOut)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, formatted)
	return err
}

func (Service) FixturesPromote(root string, ids []string, opts fixtures.PromoteOptions, w io.Writer) error {
	layout, err := fixtures.ResolveLayout(root)
	if err != nil {
		return err
	}
	if opts.PromotedAt.IsZero() {
		opts.PromotedAt = time.Now().UTC()
	}
	promoted, err := fixtures.Promote(layout, ids, opts)
	if err != nil {
		return err
	}
	for _, fixture := range promoted {
		if _, err := fmt.Fprintf(w, "promoted %s -> %s\n", fixture.ID, fixture.Expectation.ExpectedPlaybook); err != nil {
			return err
		}
	}
	return nil
}

func (Service) FixturesStats(root string, class fixtures.Class, opts fixtures.EvaluateOptions, baselinePath string, jsonOut, checkBaseline, updateBaseline bool, w io.Writer) error {
	layout, err := fixtures.ResolveLayout(root)
	if err != nil {
		return err
	}
	if baselinePath != "" && !filepath.IsAbs(baselinePath) {
		baselinePath = filepath.Join(layout.Root, baselinePath)
	}
	report, err := fixtures.Evaluate(layout, class, opts)
	if err != nil {
		return err
	}
	if baselinePath != "" {
		report.AppliedBaselinePath = baselinePath
	}
	if updateBaseline {
		thresholds := fixtures.Thresholds{MinTop1: 0.65, MinTop3: 0.85, MaxUnmatched: 0.15, MaxFalsePositive: 0.35, MaxWeakMatch: 0.15}
		if err := fixtures.WriteBaseline(baselinePath, report.Baseline(thresholds)); err != nil {
			return err
		}
	}
	if checkBaseline {
		baseline, err := fixtures.LoadBaseline(baselinePath)
		if err != nil {
			return err
		}
		if err := fixtures.CheckBaseline(&report, baseline); err != nil {
			formatted, ferr := fixtures.FormatStatsReport(report, jsonOut)
			if ferr == nil {
				_, _ = fmt.Fprint(w, formatted)
			}
			return err
		}
	}
	formatted, err := fixtures.FormatStatsReport(report, jsonOut)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, formatted)
	return err
}
