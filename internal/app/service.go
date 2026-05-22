package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"faultline/internal/artifact"
	"faultline/internal/authoring"
	"faultline/internal/engine"
	"faultline/internal/fixtures"
	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/renderer"
	tracereport "faultline/internal/trace"
	"faultline/internal/workflow"
)

// Service owns app-level orchestration for CLI commands.
type Service struct{}

const (
	defaultFixtureBaselineMinTop1          = 0.65
	defaultFixtureBaselineMinTop3          = 0.85
	defaultFixtureBaselineMaxUnmatched     = 0.15
	defaultFixtureBaselineMaxFalsePositive = 0.35
	defaultFixtureBaselineMaxWeakMatch     = 0.15
)

// ErrSilentFailure is returned by Analyze when --fail-on-silent is set and a
// silent failure is detected.  The error message is not printed to stderr;
// the analysis output already describes the finding.
var ErrSilentFailure = errors.New("silent failure detected")

// ErrBatchUnmatched is returned by Batch when one or more input logs did not
// match any playbook. It is a sentinel: output has already been written, and
// the exit code signals that the run produced a partial diagnosis.
var ErrBatchUnmatched = errors.New("one or more logs did not match any playbook")

// NewService returns the default CLI application service.
func NewService() Service {
	return Service{}
}

// Analyze performs log analysis and writes formatted output to w.
func (Service) Analyze(ctx context.Context, r io.Reader, source string, opts AnalyzeOptions, w io.Writer) error {
	if opts.View == output.ViewTrace {
		opts.TraceEnabled = true
		opts.View = output.ViewDefault
	}
	if opts.TraceEnabled || opts.TracePlaybook != "" {
		return Service{}.Trace(ctx, r, source, opts, w)
	}
	a, err := analyzeLog(ctx, r, source, opts, "analyze", true)
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
func (Service) Trace(ctx context.Context, r io.Reader, source string, opts AnalyzeOptions, w io.Writer) error {
	loaded, err := loadAnalysisInput(ctx, r, source, opts)
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}

	playbookID, err := tracePlaybookID(loaded.Analysis, opts)
	if err != nil {
		return err
	}
	if playbookID == "" {
		return writeAnalysis(loaded.Analysis, AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, Mode: output.ModeQuick, Format: opts.Format, JSON: opts.JSON}}, w)
	}

	report, err := tracereport.Build(loaded.Analysis, loaded.Lines, loaded.Playbooks, playbookID, opts.ShowRejected)
	if err != nil {
		return err
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

// Fix performs log analysis and writes only the ranked fix steps to w.
func (Service) Fix(ctx context.Context, r io.Reader, source string, opts AnalyzeOptions, w io.Writer) error {
	a, err := analyzeLog(ctx, r, source, opts, "fix", false)
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
	rendOpts := renderer.DetectOptions(w)
	rendOpts.FixCommandsOnly = opts.FixCommandsOnly
	rendOpts.FixWithPreconditions = opts.FixWithPreconditions
	rendOpts.FixWithRisks = opts.FixWithRisks
	_, werr := fmt.Fprint(w, output.FormatFix(a, rendOpts))
	return werr
}

// List loads all playbooks and writes a formatted list to w.
func (Service) List(category, playbookDir string, playbookPacks []string, w io.Writer) error {
	return catalogService{}.List(category, playbookDir, playbookPacks, w)
}

// Explain fetches a single playbook by id and writes its details to w.
func (Service) Explain(id, playbookDir string, playbookPacks []string, format output.Format, w io.Writer) error {
	return catalogService{}.Explain(id, playbookDir, playbookPacks, format, w)
}

// ListInstalledPacks prints the user-installed extra packs.
func (Service) ListInstalledPacks(w io.Writer) error {
	return catalogService{}.ListInstalledPacks(w)
}

// InstallPack installs a playbook pack into the user's persistent Faultline directory.
func (Service) InstallPack(srcDir, name string, force bool, w io.Writer) error {
	return catalogService{}.InstallPack(srcDir, name, force, w)
}

// Workflow analyzes the log and emits the deterministic follow-up handoff.
func (Service) Workflow(ctx context.Context, r io.Reader, source string, opts AnalyzeOptions, mode workflow.Mode, jsonOut bool, w io.Writer) error {
	a, err := analyzeLog(ctx, r, source, opts, "workflow", false)
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

func (Service) FixturesIngest(ctx context.Context, root string, opts fixtures.IngestOptions, jsonOut bool, w io.Writer) error {
	layout, err := fixtures.ResolveLayout(root)
	if err != nil {
		return err
	}
	result, err := fixtures.Ingest(ctx, layout, opts)
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
		opts.PromotedAt = optionNow(AnalyzeOptions{})
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
	if opts.Now == nil {
		opts.Now = func() time.Time { return optionNow(AnalyzeOptions{}) }
	}
	report, err := fixtures.Evaluate(layout, class, opts)
	if err != nil {
		return err
	}
	if baselinePath != "" {
		report.AppliedBaselinePath = baselinePath
	}
	if updateBaseline {
		thresholds := fixtures.Thresholds{
			MinTop1:          defaultFixtureBaselineMinTop1,
			MinTop3:          defaultFixtureBaselineMinTop3,
			MaxUnmatched:     defaultFixtureBaselineMaxUnmatched,
			MaxFalsePositive: defaultFixtureBaselineMaxFalsePositive,
			MaxWeakMatch:     defaultFixtureBaselineMaxWeakMatch,
		}
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

// Batch analyzes multiple CI log files and groups the results by failure
// pattern to identify recurring root causes across a build matrix.
//
// Exit semantics: if any input did not match a playbook, Batch returns
// ErrBatchUnmatched after writing the full output. Real errors (file open
// failure, engine failure) abort the run and return the wrapped error.
func (Service) Batch(ctx context.Context, sources []string, opts AnalyzeOptions, w io.Writer) error {
	result := &model.BatchResult{
		SchemaVersion: "batch.v1",
		Total:         len(sources),
		Entries:       make([]model.BatchEntry, 0, len(sources)),
	}
	patternMap := map[string]*model.BatchPattern{}

	rootAbs, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	batchEngine := engine.New(logEngineOptions(opts))
	for _, src := range sources {
		validatedPath, err := resolvePathWithinRoot(src, rootAbs)
		if err != nil {
			return fmt.Errorf("invalid source path %s: %w", src, err)
		}

		f, err := os.Open(validatedPath)
		if err != nil {
			return fmt.Errorf("open %s: %w", src, err)
		}
		a, analyzeErr := analyzeLogWithEngine(ctx, batchEngine, f, src, opts, "batch", false)
		_ = f.Close()

		if analyzeErr != nil && !errors.Is(analyzeErr, engine.ErrNoMatch) {
			return fmt.Errorf("analyze %s: %w", src, analyzeErr)
		}

		entry := model.BatchEntry{Source: src}
		if a != nil && len(a.Results) > 0 {
			top := a.Results[0]
			entry.Matched = true
			entry.FailureID = top.Playbook.ID
			entry.Confidence = top.Confidence
			entry.Title = top.Playbook.Title
			entry.Category = top.Playbook.Category
			entry.Severity = top.Playbook.Severity
			result.Matched++

			pat, ok := patternMap[top.Playbook.ID]
			if !ok {
				pat = &model.BatchPattern{
					FailureID: top.Playbook.ID,
					Title:     top.Playbook.Title,
					Category:  top.Playbook.Category,
					Severity:  top.Playbook.Severity,
					Sources:   []string{},
				}
				patternMap[top.Playbook.ID] = pat
			}
			if top.Confidence > pat.Confidence {
				pat.Confidence = top.Confidence
			}
			pat.Count++
			pat.Sources = append(pat.Sources, src)
		} else {
			result.Unmatched++
			result.UnmatchedSources = append(result.UnmatchedSources, src)
		}
		result.Entries = append(result.Entries, entry)
	}

	// Sort patterns by count desc, then failure_id asc for determinism.
	for _, pat := range patternMap {
		result.Patterns = append(result.Patterns, *pat)
	}
	sort.Slice(result.Patterns, func(i, j int) bool {
		if result.Patterns[i].Count != result.Patterns[j].Count {
			return result.Patterns[i].Count > result.Patterns[j].Count
		}
		return result.Patterns[i].FailureID < result.Patterns[j].FailureID
	})

	switch {
	case opts.JSON || opts.Format == output.FormatJSON:
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(data)); err != nil {
			return err
		}
	case opts.Format == output.FormatMarkdown:
		if _, err := fmt.Fprint(w, output.FormatBatchMarkdown(result)); err != nil {
			return err
		}
	default:
		if _, err := fmt.Fprint(w, output.FormatBatchText(result)); err != nil {
			return err
		}
	}

	if result.Unmatched > 0 {
		return ErrBatchUnmatched
	}
	return nil
}

// resolvePathWithinRoot normalizes src and ensures the resulting absolute path
// remains inside rootAbs. It rejects empty paths and traversal attempts (for
// both relative and absolute inputs) that escape the allowed root.
func resolvePathWithinRoot(src, rootAbs string) (string, error) {
	if src == "" {
		return "", errors.New("path is empty")
	}

	cleaned := filepath.Clean(src)
	var candidate string
	if filepath.IsAbs(cleaned) {
		candidate = cleaned
	} else {
		candidate = filepath.Join(rootAbs, cleaned)
	}

	candidateAbs, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}

	rootEval, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", fmt.Errorf("resolve root symlinks: %w", err)
	}
	candidateEval, err := resolvePathForContainment(candidateAbs)
	if err != nil {
		return "", fmt.Errorf("resolve candidate symlinks: %w", err)
	}

	rel, err := filepath.Rel(rootEval, candidateEval)
	if err != nil {
		return "", fmt.Errorf("compute relative path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes allowed root")
	}

	return candidateAbs, nil
}

func resolvePathForContainment(candidateAbs string) (string, error) {
	candidateEval, err := filepath.EvalSymlinks(candidateAbs)
	if err == nil {
		return candidateEval, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	ancestor := filepath.Dir(candidateAbs)
	for {
		if _, statErr := os.Lstat(ancestor); statErr == nil {
			ancestorEval, evalErr := filepath.EvalSymlinks(ancestor)
			if evalErr != nil {
				return "", evalErr
			}
			tail, relErr := filepath.Rel(ancestor, candidateAbs)
			if relErr != nil {
				return "", relErr
			}
			return filepath.Join(ancestorEval, tail), nil
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return "", statErr
		}

		next := filepath.Dir(ancestor)
		// filepath.Dir returns the same value at filesystem root.
		if next == ancestor {
			return candidateAbs, nil
		}
		ancestor = next
	}
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
