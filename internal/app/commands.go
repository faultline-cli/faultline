// Package app implements the command handlers for each Faultline sub-command.
// Handlers receive a pre-opened log reader, structured options, and a writer;
// they coordinate the engine and output packages and write the final result.
package app

import (
	"errors"
	"fmt"
	"io"

	"faultline/internal/engine"
	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/workflow"
)

// AnalyzeOptions collects all flags that influence the analyze and fix commands.
type AnalyzeOptions struct {
	// Top is the maximum number of ranked results to show (0 = all).
	Top int
	// Mode selects quick or detailed human-readable output.
	Mode output.Mode
	// JSON overrides Mode and emits machine-readable JSON.
	JSON bool
	// CIAnnotations emits GitHub Actions-style ::warning:: lines.
	CIAnnotations bool
	// NoHistory skips reading and writing the local history store.
	NoHistory bool
	// PlaybookDir overrides the default playbook directory.
	PlaybookDir string
	// GitContextEnabled enriches diagnosis results with local git history.
	GitContextEnabled bool
	// GitSince limits git history scanning to recent commits.
	GitSince string
	// RepoPath overrides the repository path used for git context scanning.
	RepoPath string
}

// RunAnalyze performs log analysis and writes formatted output to w.
//
// engine.ErrNoInput is propagated so the caller can exit non-zero.
// A no-match result is written as informational output (exit 0 convention).
func RunAnalyze(r io.Reader, source string, opts AnalyzeOptions, w io.Writer) error {
	a, err := doAnalyze(r, source, opts)
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}
	// ErrNoMatch: a is non-nil with empty Results; render informational output.
	return writeAnalysis(a, opts, w)
}

// RunFix performs log analysis and writes only the ranked fix steps to w.
func RunFix(r io.Reader, source string, opts AnalyzeOptions, w io.Writer) error {
	a, err := doAnalyze(r, source, opts)
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}
	_, werr := fmt.Fprint(w, output.FormatFix(a))
	return werr
}

// RunList loads all playbooks and writes a formatted list to w.
func RunList(category, playbookDir string, w io.Writer) error {
	pbs, err := engine.New(engine.Options{
		PlaybookDir: playbookDir,
		NoHistory:   true,
	}).List()
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, output.FormatPlaybookList(pbs, category))
	return err
}

// RunExplain fetches a single playbook by id and writes its details to w.
func RunExplain(id, playbookDir string, w io.Writer) error {
	pb, err := engine.New(engine.Options{
		PlaybookDir: playbookDir,
		NoHistory:   true,
	}).Explain(id)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, output.FormatPlaybookDetails(pb))
	return err
}

// RunWorkflow analyzes the log and emits a deterministic local or agent-ready
// follow-up workflow based on the top diagnosis.
func RunWorkflow(r io.Reader, source string, opts AnalyzeOptions, mode workflow.Mode, jsonOut bool, w io.Writer) error {
	a, err := doAnalyze(r, source, opts)
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

// doAnalyze runs the engine and attaches the source label to the analysis.
func doAnalyze(r io.Reader, source string, opts AnalyzeOptions) (*model.Analysis, error) {
	a, err := engine.New(engine.Options{
		PlaybookDir:       opts.PlaybookDir,
		NoHistory:         opts.NoHistory,
		GitContextEnabled: opts.GitContextEnabled,
		GitSince:          opts.GitSince,
		RepoPath:          opts.RepoPath,
	}).AnalyzeReader(r)
	if a != nil {
		a.Source = source
	}
	return a, err
}

// writeAnalysis dispatches to the appropriate formatter based on opts.
func writeAnalysis(a *model.Analysis, opts AnalyzeOptions, w io.Writer) error {
	if opts.JSON {
		data, err := output.FormatAnalysisJSON(a, opts.Top)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, data)
		return err
	}

	if opts.CIAnnotations {
		_, err := fmt.Fprint(w, output.FormatCIAnnotations(a, opts.Top))
		return err
	}

	_, err := fmt.Fprint(w, output.FormatAnalysisText(a, opts.Top, opts.Mode))
	return err
}
