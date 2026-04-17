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
	// Select chooses a single ranked result by 1-based position.
	Select int
	// Mode selects quick or detailed human-readable output.
	Mode output.Mode
	// Format selects the human-readable output shape.
	Format output.Format
	// View selects a focused slice of human-readable output.
	View output.View
	// JSON overrides Mode and emits machine-readable JSON.
	JSON bool
	// CIAnnotations emits GitHub Actions-style ::warning:: lines.
	CIAnnotations bool
	// NoHistory skips reading and writing the local history store.
	NoHistory bool
	// PlaybookDir overrides the default playbook directory.
	PlaybookDir string
	// PlaybookPackDirs adds extra pack roots on top of the bundled catalog.
	PlaybookPackDirs []string
	// GitContextEnabled enriches diagnosis results with local git history.
	GitContextEnabled bool
	// GitSince limits git history scanning to recent commits.
	GitSince string
	// RepoPath overrides the repository path used for git context scanning.
	RepoPath string
	// BayesEnabled enables deterministic Bayesian-inspired reranking.
	BayesEnabled bool
	// TraceEnabled renders a deterministic playbook trace instead of the normal report.
	TraceEnabled bool
	// TracePlaybook renders a deterministic trace for the named playbook.
	TracePlaybook string
	// ShowRejected includes competing candidates and rejection context in trace output.
	ShowRejected bool
	// ShowEvidence includes a raw evidence appendix when supported by the selected view.
	ShowEvidence bool
	// ShowScoring includes scoring detail when supported by the selected view.
	ShowScoring bool
	// DeltaProvider enables provider-backed failure delta resolution.
	DeltaProvider string
	// GitHubRepository identifies the GitHub repository for provider-backed delta resolution.
	GitHubRepository string
	// GitHubBranch identifies the GitHub branch for provider-backed delta resolution.
	GitHubBranch string
	// GitHubRunID identifies the current GitHub Actions run for provider-backed delta resolution.
	GitHubRunID int64
	// GitHubToken authenticates provider-backed GitHub Actions delta resolution.
	GitHubToken string
}

// writeAnalysis dispatches to the appropriate formatter based on opts.
func writeAnalysis(a *model.Analysis, opts AnalyzeOptions, w io.Writer) error {
	selected, err := selectAnalysisResults(a, opts)
	if err != nil {
		return err
	}

	if opts.JSON || opts.Format == output.FormatJSON {
		data, err := output.FormatAnalysisJSON(selected, selectedTop(opts))
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, data)
		return err
	}

	if opts.CIAnnotations {
		_, err := fmt.Fprint(w, output.FormatCIAnnotations(selected, selectedTop(opts)))
		return err
	}

	mode := opts.Mode
	if mode != output.ModeDetailed && (opts.ShowEvidence || opts.ShowScoring || opts.ShowRejected) {
		mode = output.ModeDetailed
	}

	renderOpts := renderer.DetectOptions(w)
	if opts.View == output.ViewFix {
		if opts.Format == output.FormatMarkdown {
			_, err := fmt.Fprint(w, output.FormatFixMarkdown(selected))
			return err
		}
		_, err := fmt.Fprint(w, output.FormatFix(selected, renderOpts))
		return err
	}
	if opts.View == output.ViewEvidence {
		if opts.Format == output.FormatMarkdown {
			_, err := fmt.Fprint(w, output.FormatAnalysisEvidenceMarkdown(selected))
			return err
		}
		_, err := fmt.Fprint(w, output.FormatAnalysisEvidenceText(selected))
		return err
	}
	if opts.View == output.ViewSummary {
		mode = output.ModeQuick
	}
	if opts.View == output.ViewRaw {
		mode = output.ModeDetailed
	}
	if opts.Format == output.FormatMarkdown {
		_, err := fmt.Fprint(w, output.FormatAnalysisMarkdown(selected, selectedTop(opts), mode))
		return err
	}
	_, err = fmt.Fprint(w, output.FormatAnalysisText(selected, selectedTop(opts), mode, renderOpts))
	return err
}

func selectedTop(opts AnalyzeOptions) int {
	if opts.Select > 0 {
		return 1
	}
	return opts.Top
}

func selectAnalysisResults(a *model.Analysis, opts AnalyzeOptions) (*model.Analysis, error) {
	if a == nil {
		return nil, nil
	}
	clone := *a
	if len(a.Results) == 0 {
		clone.Results = []model.Result{}
		return &clone, nil
	}
	if opts.Select <= 0 {
		return &clone, nil
	}
	if opts.Select > len(a.Results) {
		return nil, fmt.Errorf("--select %d is out of range; only %d result(s) available", opts.Select, len(a.Results))
	}
	clone.Results = []model.Result{a.Results[opts.Select-1]}
	if opts.Select != 1 {
		clone.Differential = nil
	}
	return &clone, nil
}
