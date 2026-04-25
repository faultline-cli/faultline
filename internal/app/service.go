package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"text/tabwriter"
	"time"

	"faultline/internal/artifact"
	"faultline/internal/authoring"
	analysiscompare "faultline/internal/compare"
	"faultline/internal/detectors"
	"faultline/internal/engine"
	"faultline/internal/fixtures"
	"faultline/internal/hooks"
	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/playbooks"
	"faultline/internal/renderer"
	"faultline/internal/repo"
	tracereport "faultline/internal/trace"
	"faultline/internal/workflow"
	workflowexec "faultline/internal/workflow/execute"
	workflowrender "faultline/internal/workflow/render"
)

// Service owns app-level orchestration for CLI commands.
type Service struct{}

var ErrGuardFindings = errors.New("guard findings emitted")

// ErrSilentFailure is returned by Analyze when --fail-on-silent is set and a
// silent failure is detected.  The error message is not printed to stderr;
// the analysis output already describes the finding.
var ErrSilentFailure = errors.New("silent failure detected")

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
	if opts.View == output.ViewTrace {
		opts.TraceEnabled = true
		opts.View = output.ViewDefault
	}
	if opts.TraceEnabled || opts.TracePlaybook != "" {
		return Service{}.Trace(r, source, opts, w)
	}
	a, err := analyzeLog(r, source, opts, "analyze", true)
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}
	if writeErr := writeAnalysis(a, opts, w); writeErr != nil {
		return writeErr
	}
	if opts.FailOnSilent && a != nil && len(a.SilentFindings) > 0 {
		return ErrSilentFailure
	}
	return nil
}

// Trace performs log analysis and renders a deterministic playbook trace.
func (Service) Trace(r io.Reader, source string, opts AnalyzeOptions, w io.Writer) error {
	loaded, err := loadAnalysisInput(r, source, opts)
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}

	playbooks, err := engine.New(engine.Options{
		PlaybookDir:      opts.PlaybookDir,
		PlaybookPackDirs: opts.PlaybookPackDirs,
		NoHistory:        true,
	}).List()
	if err != nil {
		return err
	}

	playbookID, err := tracePlaybookID(loaded.Analysis, opts)
	if err != nil {
		return err
	}
	if playbookID == "" {
		return writeAnalysis(loaded.Analysis, AnalyzeOptions{Top: 1, Mode: output.ModeQuick, Format: opts.Format, JSON: opts.JSON}, w)
	}

	report, err := tracereport.Build(loaded.Analysis, loaded.Lines, playbooks, playbookID, opts.ShowRejected)
	if err != nil {
		return err
	}
	if hookReport := hooks.NewExecutor(hooks.HookPolicy{Mode: opts.HookMode}).Execute(context.Background(), report.Playbook, report.Confidence, hookWorkDir(opts)); hookReport != nil {
		report.Hooks = hookReport
		if report.Matched {
			report.Confidence = hookReport.FinalConfidence
		}
	}

	switch {
	case opts.JSON || opts.Format == output.FormatJSON:
		data, err := output.FormatTraceJSON(report, opts.ShowEvidence, opts.ShowScoring, opts.ShowRejected)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, data)
		return err
	case opts.Format == output.FormatMarkdown:
		_, err := fmt.Fprint(w, output.FormatTraceMarkdown(report, opts.ShowEvidence, opts.ShowScoring, opts.ShowRejected))
		return err
	default:
		_, err := fmt.Fprint(w, output.FormatTraceText(report, opts.ShowEvidence, opts.ShowScoring, opts.ShowRejected))
		return err
	}
}

// Replay re-renders a saved analysis artifact using the current deterministic
// output surfaces. Replay currently supports the stable analysis JSON schema.
func (Service) Replay(r io.Reader, opts AnalyzeOptions, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read analysis artifact: %w", err)
	}
	a, err := output.ParseAnalysisJSON(data)
	if err != nil {
		return err
	}
	a = artifact.Sync(a)
	if opts.View == output.ViewTrace {
		return fmt.Errorf("replay trace is not supported from analysis artifacts; replay a saved trace artifact or use `faultline trace` on the original log")
	}
	if opts.TraceEnabled || opts.TracePlaybook != "" {
		return fmt.Errorf("replay trace is not supported from analysis artifacts; replay a saved trace artifact or use `faultline trace` on the original log")
	}
	return writeAnalysis(a, opts, w)
}

// Compare renders a deterministic comparison between two saved analysis artifacts.
func (Service) Compare(left, right io.Reader, opts AnalyzeOptions, w io.Writer) error {
	leftData, err := io.ReadAll(left)
	if err != nil {
		return fmt.Errorf("read left analysis artifact: %w", err)
	}
	rightData, err := io.ReadAll(right)
	if err != nil {
		return fmt.Errorf("read right analysis artifact: %w", err)
	}
	leftAnalysis, err := output.ParseAnalysisJSON(leftData)
	if err != nil {
		return err
	}
	leftAnalysis = artifact.Sync(leftAnalysis)
	rightAnalysis, err := output.ParseAnalysisJSON(rightData)
	if err != nil {
		return err
	}
	rightAnalysis = artifact.Sync(rightAnalysis)

	report := analysiscompare.Build(leftAnalysis, rightAnalysis)
	switch {
	case opts.JSON || opts.Format == output.FormatJSON:
		data, err := output.FormatCompareJSON(report)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, data)
		return err
	case opts.Format == output.FormatMarkdown:
		_, err := fmt.Fprint(w, output.FormatCompareMarkdown(report))
		return err
	default:
		_, err := fmt.Fprint(w, output.FormatCompareText(report))
		return err
	}
}

// Fix performs log analysis and writes only the ranked fix steps to w.
func (Service) Fix(r io.Reader, source string, opts AnalyzeOptions, w io.Writer) error {
	a, err := analyzeLog(r, source, opts, "fix", false)
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
	if _, err := fmt.Fprintln(tw, "NAME\tPLAYBOOKS\tVERSION\tPINNED REF\tPATH"); err != nil {
		return err
	}
	for _, pack := range packs {
		version := pack.Version
		if version == "" {
			version = "-"
		}
		ref := pack.PinnedRef
		if ref == "" {
			ref = "-"
		}
		if _, err := fmt.Fprintf(tw, "%s\t%d\t%s\t%s\t%s\n", pack.Name, pack.PlaybookCount, version, ref, pack.Root); err != nil {
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

// Workflow analyzes the log and emits the legacy deterministic follow-up workflow.
func (Service) Workflow(r io.Reader, source string, opts AnalyzeOptions, mode workflow.Mode, jsonOut bool, w io.Writer) error {
	a, err := analyzeLog(r, source, opts, "workflow", false)
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

func (Service) WorkflowExplain(r io.Reader, source string, opts AnalyzeOptions, workflowRef string, jsonOut bool, w io.Writer) error {
	analysis, err := loadWorkflowAnalysis(r, source, opts)
	if err != nil {
		return err
	}
	doc, err := workflow.Explain(context.Background(), analysis, workflow.Options{
		WorkflowRef: workflowRef,
		RepoPath:    opts.RepoPath,
	})
	if err != nil {
		return err
	}
	if jsonOut {
		data, err := workflow.MarshalPlanJSON(doc)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, data)
		return err
	}
	_, err = fmt.Fprint(w, workflowrender.PlanText(doc))
	return err
}

func (Service) WorkflowApply(r io.Reader, source string, opts AnalyzeOptions, workflowRef string, dryRun bool, policy workflowexec.Policy, jsonOut bool, w io.Writer) error {
	analysis, err := loadWorkflowAnalysis(r, source, opts)
	if err != nil {
		return err
	}
	if dryRun {
		doc, err := workflow.DryRun(context.Background(), analysis, workflow.Options{
			WorkflowRef: workflowRef,
			RepoPath:    opts.RepoPath,
		})
		if err != nil {
			return err
		}
		if jsonOut {
			data, err := workflow.MarshalPlanJSON(doc)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(w, data)
			return err
		}
		_, err = fmt.Fprint(w, workflowrender.PlanText(doc))
		return err
	}
	record, err := workflow.Apply(context.Background(), analysis, workflow.Options{
		WorkflowRef: workflowRef,
		RepoPath:    opts.RepoPath,
	}, policy)
	if record == nil && err != nil {
		return err
	}
	record, persistErr := persistWorkflowExecution(record, opts)
	if persistErr != nil {
		return persistErr
	}
	if jsonOut {
		data, err := workflow.MarshalExecutionJSON(record)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, data)
		return err
	}
	_, writeErr := fmt.Fprint(w, workflowrender.ExecutionText(record))
	if err != nil {
		return err
	}
	return writeErr
}

func (Service) WorkflowShow(executionID string, opts AnalyzeOptions, jsonOut bool, w io.Writer) error {
	record, err := loadWorkflowExecution(executionID, opts)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("workflow execution %s not found", executionID)
	}
	if jsonOut {
		data, err := workflow.MarshalExecutionJSON(record)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, data)
		return err
	}
	_, err = fmt.Fprint(w, workflowrender.ExecutionText(record))
	return err
}

func (Service) WorkflowHistory(opts AnalyzeOptions, limit int, jsonOut bool, w io.Writer) error {
	st, _, err := openWorkflowStore(opts)
	if err != nil {
		return err
	}
	defer st.Close()
	items, err := st.ListWorkflowExecutions(context.Background(), limit)
	if err != nil {
		return err
	}
	if jsonOut {
		data, err := json.Marshal(items)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "%s\n", data)
		return err
	}
	_, err = fmt.Fprint(w, workflowrender.HistoryText(items))
	return err
}

// Inspect scans a repository tree with source-detector playbooks.
func (Service) Inspect(root string, opts AnalyzeOptions, w io.Writer) error {
	changeSet := detectors.ChangeSet{}
	if scanner, err := repo.NewScanner(root); err == nil {
		if loaded, loadErr := repo.LoadWorktreeChangeSet(scanner); loadErr != nil {
			return loadErr
		} else {
			absRoot, absErr := filepath.Abs(root)
			if absErr != nil {
				return absErr
			}
			prefix, relErr := filepath.Rel(scanner.Root, absRoot)
			if relErr != nil {
				return relErr
			}
			changeSet = repo.ChangeSetRelativeTo(loaded, prefix)
		}
	}
	a, err := engine.New(engine.Options{
		PlaybookDir:      opts.PlaybookDir,
		PlaybookPackDirs: opts.PlaybookPackDirs,
		GitSince:         opts.GitSince,
		RepoPath:         opts.RepoPath,
		BayesEnabled:     opts.BayesEnabled,
	}).AnalyzeRepository(root, changeSet)
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}
	a, prepErr := prepareAnalysisWithStore(a, "", "repository", "inspect", opts, true)
	if prepErr != nil {
		return prepErr
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

func analyzeLog(r io.Reader, source string, opts AnalyzeOptions, surface string, persist bool) (*model.Analysis, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read log input: %w", err)
	}
	a, err := engine.New(engine.Options{
		PlaybookDir:       opts.PlaybookDir,
		PlaybookPackDirs:  opts.PlaybookPackDirs,
		GitContextEnabled: opts.GitContextEnabled,
		GitSince:          opts.GitSince,
		RepoPath:          opts.RepoPath,
		BayesEnabled:      opts.BayesEnabled,
		DeltaProvider:     opts.DeltaProvider,
		GitHubRepository:  opts.GitHubRepository,
		GitHubBranch:      opts.GitHubBranch,
		GitHubRunID:       opts.GitHubRunID,
		GitHubToken:       opts.GitHubToken,
		GitLabProject:     opts.GitLabProject,
		GitLabBranch:      opts.GitLabBranch,
		GitLabPipelineID:  opts.GitLabPipelineID,
		GitLabJobID:       opts.GitLabJobID,
		GitLabToken:       opts.GitLabToken,
		GitLabAPIBaseURL:  opts.GitLabAPIBaseURL,
	}).AnalyzeReader(bytes.NewReader(data))
	if a != nil {
		a.Source = source
	}
	if prepErr := error(nil); a != nil || errors.Is(err, engine.ErrNoMatch) {
		a, prepErr = prepareAnalysisWithStore(a, string(data), "log", surface, opts, persist)
		if prepErr != nil {
			return nil, prepErr
		}
	}
	return a, err
}

type loadedAnalysisInput struct {
	Analysis *model.Analysis
	Lines    []model.Line
}

func loadAnalysisInput(r io.Reader, source string, opts AnalyzeOptions) (loadedAnalysisInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return loadedAnalysisInput{}, fmt.Errorf("read log input: %w", err)
	}
	lines, err := engine.ReadLines(bytes.NewReader(data))
	if err != nil {
		return loadedAnalysisInput{}, err
	}
	baseOpts := opts
	baseOpts.HookMode = model.HookModeOff
	analysis, err := analyzeLog(bytes.NewReader(data), source, baseOpts, "trace", false)
	return loadedAnalysisInput{
		Analysis: analysis,
		Lines:    lines,
	}, err
}

func tracePlaybookID(a *model.Analysis, opts AnalyzeOptions) (string, error) {
	if opts.TracePlaybook != "" {
		return opts.TracePlaybook, nil
	}
	if opts.Select > 0 {
		if a == nil || len(a.Results) == 0 {
			return "", fmt.Errorf("--select requires at least one matched result")
		}
		if opts.Select > len(a.Results) {
			return "", fmt.Errorf("--select %d is out of range; only %d result(s) available", opts.Select, len(a.Results))
		}
		return a.Results[opts.Select-1].Playbook.ID, nil
	}
	if a != nil && len(a.Results) > 0 {
		return a.Results[0].Playbook.ID, nil
	}
	return "", nil
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

func (Service) FixturesSanitize(root string, ids []string, opts fixtures.SanitizeOptions, jsonOut bool, w io.Writer) error {
	layout, err := fixtures.ResolveLayout(root)
	if err != nil {
		return err
	}
	results, err := fixtures.Sanitize(layout, ids, opts)
	if err != nil {
		return err
	}
	formatted, err := fixtures.FormatSanitizeResults(results, jsonOut)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, formatted)
	return err
}

func (Service) FixturesCompareModes(root string, class fixtures.Class, opts fixtures.EvaluateOptions, jsonOut, failOnRegression bool, w io.Writer) error {
	layout, err := fixtures.ResolveLayout(root)
	if err != nil {
		return err
	}
	baselineOpts := opts
	baselineOpts.BayesEnabled = false
	bayesOpts := opts
	bayesOpts.BayesEnabled = true

	baselineReport, err := fixtures.Evaluate(layout, class, baselineOpts)
	if err != nil {
		return fmt.Errorf("baseline evaluation: %w", err)
	}
	bayesReport, err := fixtures.Evaluate(layout, class, bayesOpts)
	if err != nil {
		return fmt.Errorf("bayes evaluation: %w", err)
	}
	cmp, err := fixtures.CompareReports(baselineReport, bayesReport)
	if err != nil {
		return err
	}
	formatted, err := fixtures.FormatModeComparison(cmp, jsonOut)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, formatted)
	if err != nil {
		return err
	}
	if failOnRegression && cmp.HasRegressions() {
		return fmt.Errorf("bayes mode regressed %d fixture(s)", cmp.Regressed)
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

// FixturesScaffold generates a candidate playbook YAML scaffold from a
// sanitized log. logText is the raw log content; sanitization is applied
// automatically before pattern extraction. The scaffold is written to w (and
// optionally to opts.PackDir when set).
//
// FixturesScaffold is maintainer-only; it is wired under the hidden
// fixtures command and is not part of the default user narrative.
func (Service) FixturesScaffold(logText string, opts authoring.ScaffoldOptions, w io.Writer) error {
	sanitized, _ := fixtures.ApplySanitizeRules(logText)
	result, err := authoring.ScaffoldPlaybook(sanitized, opts)
	if err != nil {
		return err
	}
	if result.OutputPath != "" {
		if _, err := fmt.Fprintf(w, "wrote scaffold: %s\n\n", result.OutputPath); err != nil {
			return err
		}
	}
	_, err = fmt.Fprint(w, result.YAML)
	return err
}
