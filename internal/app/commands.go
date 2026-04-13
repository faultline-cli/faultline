// Package app implements CLI-facing application services and command helpers.
package app

import (
	"fmt"
	"io"

	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/renderer"
)

// AnalyzeOptions collects all flags that influence the analyze and fix commands.
type AnalyzeOptions struct {
	// Top is the maximum number of ranked results to show (0 = all).
	Top int
	// Mode selects quick or detailed human-readable output.
	Mode output.Mode
	// Format selects the human-readable output shape.
	Format output.Format
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
	if opts.Format == output.FormatMarkdown {
		_, err := fmt.Fprint(w, output.FormatAnalysisMarkdown(a, opts.Top, opts.Mode))
		return err
	}
	_, err := fmt.Fprint(w, output.FormatAnalysisText(a, opts.Top, opts.Mode, renderOpts))
	return err
}
