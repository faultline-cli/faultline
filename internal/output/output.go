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
	FormatTerminal Format = "terminal"
	FormatMarkdown Format = "markdown"
	FormatJSON     Format = "json"
)

// View selects a focused slice of human-readable analysis output.
type View string

const (
	ViewDefault  View = ""
	ViewSummary  View = "summary"
	ViewEvidence View = "evidence"
	ViewFix      View = "fix"
	ViewRaw      View = "raw"
	ViewTrace    View = "trace"
)

// Valid reports whether f is a recognised output format.
func (f Format) Valid() bool {
	_, ok := ParseFormat(string(f))
	return ok
}

// ParseFormat resolves a user-provided format string into the canonical format.
func ParseFormat(value string) (Format, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(FormatTerminal):
		return FormatTerminal, true
	case string(FormatMarkdown):
		return FormatMarkdown, true
	case string(FormatJSON):
		return FormatJSON, true
	default:
		return "", false
	}
}

// Valid reports whether v is a recognised output view.
func (v View) Valid() bool {
	_, ok := ParseView(string(v))
	return ok
}

// ParseView resolves a user-provided view string into the canonical view.
func ParseView(value string) (View, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return ViewDefault, true
	case string(ViewSummary):
		return ViewSummary, true
	case string(ViewEvidence):
		return ViewEvidence, true
	case string(ViewFix):
		return ViewFix, true
	case string(ViewRaw):
		return ViewRaw, true
	case string(ViewTrace):
		return ViewTrace, true
	default:
		return "", false
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
