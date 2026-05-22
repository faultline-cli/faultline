package app

import (
	"context"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/model"
	"faultline/internal/output"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// minimalAnalysis returns a one-result analysis suitable for writeAnalysis tests.
func minimalAnalysis() *model.Analysis {
	return &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:       "docker-auth",
					Title:    "Docker Auth Failure",
					Summary:  "Image pull failed because registry credentials are missing or invalid.",
					Severity: "high",
				},
				Evidence:   []string{"pull access denied", "authentication required"},
				Confidence: 0.9,
				Score:      4.5,
			},
		},
	}
}

// ── writeAnalysis – CIAnnotations ────────────────────────────────────────────

func TestWriteAnalysisCIAnnotations(t *testing.T) {
	a := minimalAnalysis()
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, CIAnnotations: true}}
	var buf bytes.Buffer

	if err := writeAnalysis(a, opts, &buf); err != nil {
		t.Fatalf("writeAnalysis CIAnnotations: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "::") {
		t.Errorf("expected GitHub Actions annotation syntax in output, got %q", got)
	}
}

// ── writeAnalysis – ViewSummary ───────────────────────────────────────────────

func TestWriteAnalysisViewSummary(t *testing.T) {
	a := minimalAnalysis()
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, View: output.ViewSummary}}
	var buf bytes.Buffer

	if err := writeAnalysis(a, opts, &buf); err != nil {
		t.Fatalf("writeAnalysis ViewSummary: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty summary output")
	}
}

// ── writeAnalysis – ViewRaw ───────────────────────────────────────────────────

func TestWriteAnalysisViewRaw(t *testing.T) {
	a := minimalAnalysis()
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, View: output.ViewRaw}}
	var buf bytes.Buffer

	if err := writeAnalysis(a, opts, &buf); err != nil {
		t.Fatalf("writeAnalysis ViewRaw: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty raw output")
	}
}

// ── writeAnalysis – FormatMarkdown ───────────────────────────────────────────

func TestWriteAnalysisMarkdownFormat(t *testing.T) {
	a := minimalAnalysis()
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, Format: output.FormatMarkdown}}
	var buf bytes.Buffer

	if err := writeAnalysis(a, opts, &buf); err != nil {
		t.Fatalf("writeAnalysis markdown: %v", err)
	}
	if !strings.Contains(buf.String(), "#") {
		t.Errorf("expected markdown heading in output, got %q", buf.String())
	}
}

// ── writeAnalysis – ViewRaw + FormatMarkdown (path coverage) ─────────────────

func TestWriteAnalysisViewRawMarkdown(t *testing.T) {
	a := minimalAnalysis()
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, View: output.ViewRaw, Format: output.FormatMarkdown}}
	var buf bytes.Buffer

	if err := writeAnalysis(a, opts, &buf); err != nil {
		t.Fatalf("writeAnalysis ViewRaw+Markdown: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output for ViewRaw+Markdown")
	}
}

// ── writeAnalysis – ViewSummary + FormatMarkdown ─────────────────────────────

func TestWriteAnalysisViewSummaryMarkdown(t *testing.T) {
	a := minimalAnalysis()
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, View: output.ViewSummary, Format: output.FormatMarkdown}}
	var buf bytes.Buffer

	if err := writeAnalysis(a, opts, &buf); err != nil {
		t.Fatalf("writeAnalysis ViewSummary+Markdown: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output for ViewSummary+Markdown")
	}
}

// ── selectAnalysisResults – out of range ─────────────────────────────────────

func TestSelectAnalysisResultsOutOfRange(t *testing.T) {
	a := minimalAnalysis() // 1 result
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, Select: 99}}

	_, err := selectAnalysisResults(a, opts)
	if err == nil {
		t.Fatal("expected error when Select is out of range")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Errorf("expected 'out of range' in error, got %v", err)
	}
}

// ── selectAnalysisResults – nil input ────────────────────────────────────────

func TestSelectAnalysisResultsNilInput(t *testing.T) {
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, Select: 1}}

	result, err := selectAnalysisResults(nil, opts)
	if err != nil {
		t.Fatalf("expected no error for nil input, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for nil input, got %v", result)
	}
}

// ── selectAnalysisResults – valid Select ─────────────────────────────────────

func TestSelectAnalysisResultsValidSelection(t *testing.T) {
	a := minimalAnalysis()
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, Select: 1}}

	result, err := selectAnalysisResults(a, opts)
	if err != nil {
		t.Fatalf("selectAnalysisResults: %v", err)
	}
	if len(result.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(result.Results))
	}
}

// ── writeAnalysis – ModeDetailed forced by ShowEvidence ──────────────────────

func TestWriteAnalysisShowEvidenceForcesDetailed(t *testing.T) {
	a := minimalAnalysis()
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, ShowEvidence: true}}
	var buf bytes.Buffer

	if err := writeAnalysis(a, opts, &buf); err != nil {
		t.Fatalf("writeAnalysis ShowEvidence: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output for ShowEvidence")
	}
}

// ── writeAnalysis – ViewFix ───────────────────────────────────────────────────

func TestWriteAnalysisViewFixText(t *testing.T) {
	a := minimalAnalysis()
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, View: output.ViewFix}}
	var buf bytes.Buffer

	if err := writeAnalysis(a, opts, &buf); err != nil {
		t.Fatalf("writeAnalysis ViewFix text: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output for ViewFix text")
	}
}

func TestWriteAnalysisViewFixMarkdown(t *testing.T) {
	a := minimalAnalysis()
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, View: output.ViewFix, Format: output.FormatMarkdown}}
	var buf bytes.Buffer

	if err := writeAnalysis(a, opts, &buf); err != nil {
		t.Fatalf("writeAnalysis ViewFix markdown: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output for ViewFix markdown")
	}
}

// ── writeAnalysis – ViewEvidence ──────────────────────────────────────────────

func TestWriteAnalysisViewEvidenceText(t *testing.T) {
	a := minimalAnalysis()
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, View: output.ViewEvidence}}
	var buf bytes.Buffer

	if err := writeAnalysis(a, opts, &buf); err != nil {
		t.Fatalf("writeAnalysis ViewEvidence text: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output for ViewEvidence text")
	}
}

func TestWriteAnalysisViewEvidenceMarkdown(t *testing.T) {
	a := minimalAnalysis()
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, View: output.ViewEvidence, Format: output.FormatMarkdown}}
	var buf bytes.Buffer

	if err := writeAnalysis(a, opts, &buf); err != nil {
		t.Fatalf("writeAnalysis ViewEvidence markdown: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output for ViewEvidence markdown")
	}
}

// ── writeAnalysis – select error propagation ────────────────────────────────

func TestWriteAnalysisSelectErrorPropagates(t *testing.T) {
	a := minimalAnalysis() // 1 result
	opts := AnalyzeOptions{OutputOptions: OutputOptions{Top: 1, Select: 99}}
	var buf bytes.Buffer

	err := writeAnalysis(a, opts, &buf)
	if err == nil {
		t.Fatal("expected error when Select is out of range")
	}
}

// ── VerifyDeterminism ─────────────────────────────────────────────────────────

func TestVerifyDeterminismNoRunsText(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "det.db")
	svc := NewService()
	var buf bytes.Buffer

	err := svc.VerifyDeterminism(context.Background(), strings.NewReader("some log line\n"), "stdin", storePath, false, &buf)
	if err != nil {
		t.Fatalf("VerifyDeterminism: %v", err)
	}
	if !strings.Contains(buf.String(), "no stored runs") {
		t.Errorf("expected 'no stored runs' in output, got %q", buf.String())
	}
}
