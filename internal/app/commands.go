// Package app implements CLI-facing application services and command helpers.
package app

import (
	"fmt"
	"io"

	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/renderer"
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
	// PlaybookPackDirs adds extra pack roots on top of the bundled starter pack.
	PlaybookPackDirs []string
	// GitContextEnabled enriches diagnosis results with local git history.
	GitContextEnabled bool
	// GitSince limits git history scanning to recent commits.
	GitSince string
	// RepoPath overrides the repository path used for git context scanning.
	RepoPath string
}

func RunAnalyze(r io.Reader, source string, opts AnalyzeOptions, w io.Writer) error {
	return NewService().Analyze(r, source, opts, w)
}

func RunFix(r io.Reader, source string, opts AnalyzeOptions, w io.Writer) error {
	return NewService().Fix(r, source, opts, w)
}

func RunList(category, playbookDir string, w io.Writer) error {
	return NewService().List(category, playbookDir, nil, w)
}

func RunExplain(id, playbookDir string, w io.Writer) error {
	return NewService().Explain(id, playbookDir, nil, w)
}

func RunWorkflow(r io.Reader, source string, opts AnalyzeOptions, mode workflow.Mode, jsonOut bool, w io.Writer) error {
	return NewService().Workflow(r, source, opts, mode, jsonOut, w)
}

func RunInspect(root string, opts AnalyzeOptions, w io.Writer) error {
	return NewService().Inspect(root, opts, w)
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

	renderOpts := renderer.DetectOptions(w)
	_, err := fmt.Fprint(w, output.FormatAnalysisText(a, opts.Top, opts.Mode, renderOpts))
	return err
}
