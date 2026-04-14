package renderer

import (
	"strings"
	"testing"

	"faultline/internal/model"
)

func samplePlaybook() model.Playbook {
	return model.Playbook{
		ID:           "go-sum-missing",
		Title:        "Missing go.sum entry",
		Category:     "build",
		Severity:     "medium",
		Summary:      "The build needs a checksum that is missing from `go.sum`.",
		Diagnosis:    "## Diagnosis\n\nThe dependency graph references a module without a checksum entry.",
		Fix:          "## Fix steps\n\n1. Run `go mod tidy`\n2. Commit `go.mod` and `go.sum`",
		Validation:   "## Validation\n\n- Run `go test ./...`",
		WhyItMatters: "## Why it matters\n\nChecksum drift breaks reproducible builds.",
		Match:        model.MatchSpec{Any: []string{"missing go.sum entry"}},
		Workflow:     model.WorkflowSpec{Verify: []string{"go test ./..."}},
	}
}

func TestRenderExplainPlain(t *testing.T) {
	out := New(Options{Plain: true, Width: 88}).RenderExplain(samplePlaybook())
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("plain output should not contain ANSI: %q", out)
	}
	for _, want := range []string{"Summary", "Diagnosis", "Fix", "Validation", "match.any"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in plain output, got:\n%s", want, out)
		}
	}
}

func TestRenderExplainStyled(t *testing.T) {
	out := New(Options{Plain: false, Width: 88, DarkBackground: true}).RenderExplain(samplePlaybook())
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("styled output should contain ANSI escape sequences, got:\n%s", out)
	}
	if !strings.Contains(out, "Diagnosis") {
		t.Fatalf("expected markdown section in styled output, got:\n%s", out)
	}
	for _, unwanted := range []string{"## Diagnosis", "## Why it matters", "## Fix steps", "## Validation"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected duplicate markdown heading %q to be stripped, got:\n%s", unwanted, out)
		}
	}
}

func TestRenderFixUsesSectionStyling(t *testing.T) {
	a := &model.Analysis{Results: []model.Result{{
		Playbook: model.Playbook{
			Title:    "Missing go.sum entry",
			Category: "build",
			Fix:      "## Fix steps\n\n1. Run `go mod tidy`\n2. Commit `go.sum`",
		},
		Confidence: 0.82,
	}}}
	out := New(Options{Plain: true, Width: 88}).RenderFix(a)
	if !strings.Contains(out, "Fix Steps") {
		t.Fatalf("expected fix section header, got:\n%s", out)
	}
	if strings.Contains(out, "## Fix steps") {
		t.Fatalf("expected duplicate markdown heading to be removed, got:\n%s", out)
	}
}
