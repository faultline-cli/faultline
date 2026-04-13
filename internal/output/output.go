// Package output formats analysis results for humans, automation, and CI
// annotation consumers.  All functions accept a *model.Analysis that may be
// nil (when no log was provided) or have an empty Results slice (when no
// playbook matched).
//
// Formatting is split across focused sub-files:
//   - output.go           — types, shared helpers, terminal-text wrappers
//   - output_json.go      — JSON analysis formatting
//   - output_markdown.go  — markdown formatting + helpers
//   - output_annotations.go — GitHub Actions CI annotation formatting
//   - output_workflow.go  — workflow text and JSON formatting
package output

import (
	"strings"

	"faultline/internal/model"
	"faultline/internal/renderer"
)

// Mode selects the verbosity of human-readable output.
type Mode string

const (
	ModeQuick    Mode = "quick"
	ModeDetailed Mode = "detailed"
)

// Format selects the human-readable output shape.
type Format string

const (
	FormatRaw      Format = "raw"
	FormatMarkdown Format = "markdown"
)

// Valid reports whether f is a recognised output format.
func (f Format) Valid() bool {
	switch f {
	case FormatRaw, FormatMarkdown:
		return true
	default:
		return false
	}
}

// ── Human-readable text ──────────────────────────────────────────────────────

// FormatAnalysisText formats an analysis for human consumption.
// top limits the number of results shown (0 or negative means show all).
func FormatAnalysisText(a *model.Analysis, top int, mode Mode, opts renderer.Options) string {
	return renderer.New(opts).RenderAnalyze(a, top, mode == ModeDetailed)
}

// FormatFix formats only the fix steps for the top result.
func FormatFix(a *model.Analysis, opts renderer.Options) string {
	return renderer.New(opts).RenderFix(a)
}

// FormatPlaybookList formats a tab-aligned table of available playbooks.
// When category is non-empty only matching playbooks are shown.
func FormatPlaybookList(playbooks []model.Playbook, category string, opts renderer.Options) string {
	return renderer.New(opts).RenderList(playbooks, category)
}

// FormatPlaybookDetails formats all fields of a single playbook for the
// explain command.
func FormatPlaybookDetails(pb model.Playbook, opts renderer.Options) string {
	return renderer.New(opts).RenderExplain(pb)
}

// ── shared helpers ────────────────────────────────────────────────────────────

// topN returns the first n results, or all results when n <= 0.
func topN(results []model.Result, n int) []model.Result {
	if n <= 0 || n > len(results) {
		return results
	}
	return results[:n]
}

// displayPackName returns the pack name to show in output, suppressing the
// built-in "starter" and "custom" labels which add no user value.
func displayPackName(pb model.Playbook) string {
	name := strings.TrimSpace(pb.Metadata.PackName)
	if name == "" || name == "starter" || name == "custom" {
		return ""
	}
	return name
}
