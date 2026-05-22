package output

import (
	"encoding/json"
	"strings"
	"testing"

	analysiscompare "faultline/internal/compare"
	"faultline/internal/model"
)

// makeCompareReport builds a simple compare.Report for testing.
func makeCompareReport(prevID, curID string, diagChanged bool) analysiscompare.Report {
	prev := &analysiscompare.Candidate{
		FailureID:  prevID,
		Title:      prevID + " title",
		Confidence: 0.9,
	}
	cur := &analysiscompare.Candidate{
		FailureID:  curID,
		Title:      curID + " title",
		Confidence: 0.85,
	}
	return analysiscompare.Report{
		LeftSource:       prevID + ".json",
		RightSource:      curID + ".json",
		DiagnosisChanged: diagChanged,
		Changed:          diagChanged,
		Previous:         prev,
		Current:          cur,
		Summary:          []string{"summary line"},
	}
}

// ── FormatCompareText ─────────────────────────────────────────────────────────

func TestFormatCompareTextContainsSources(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	out := FormatCompareText(report)
	if !strings.Contains(out, "docker-auth.json") {
		t.Errorf("expected source in compare text, got:\n%s", out)
	}
	if !strings.Contains(out, "COMPARE") {
		t.Errorf("expected COMPARE header, got:\n%s", out)
	}
}

func TestFormatCompareTextDiagnosisSection(t *testing.T) {
	report := makeCompareReport("docker-auth", "permission-denied", true)
	out := FormatCompareText(report)
	if !strings.Contains(out, "permission-denied") {
		t.Errorf("expected current failure ID in compare text, got:\n%s", out)
	}
}

func TestFormatCompareTextEvidenceChanges(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	report.Evidence = analysiscompare.StringDelta{
		Added:   []string{"new evidence"},
		Removed: []string{"old evidence"},
	}
	out := FormatCompareText(report)
	if !strings.Contains(out, "+ new evidence") {
		t.Errorf("expected '+ new evidence' in compare text, got:\n%s", out)
	}
	if !strings.Contains(out, "- old evidence") {
		t.Errorf("expected '- old evidence' in compare text, got:\n%s", out)
	}
}

func TestFormatCompareTextFixStepChanges(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	report.FixSteps = analysiscompare.StringDelta{
		Added: []string{"run docker login"},
	}
	out := FormatCompareText(report)
	if !strings.Contains(out, "Fix Step Changes") {
		t.Errorf("expected 'Fix Step Changes' section, got:\n%s", out)
	}
	if !strings.Contains(out, "run docker login") {
		t.Errorf("expected fix step in output, got:\n%s", out)
	}
}

func TestFormatCompareTextDominantSignalChanges(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	report.DominantSignals = analysiscompare.StringDelta{
		Added: []string{"auth-signal"},
	}
	out := FormatCompareText(report)
	if !strings.Contains(out, "Dominant Signal Changes") {
		t.Errorf("expected 'Dominant Signal Changes' section, got:\n%s", out)
	}
}

func TestFormatCompareTextDeltaTestChanges(t *testing.T) {
	report := makeCompareReport("test-fail", "test-fail", false)
	report.DeltaTests = analysiscompare.StringDelta{
		Added: []string{"TestFoo"},
	}
	out := FormatCompareText(report)
	if !strings.Contains(out, "Delta Test Changes") {
		t.Errorf("expected 'Delta Test Changes' section, got:\n%s", out)
	}
}

func TestFormatCompareTextDeltaErrorChanges(t *testing.T) {
	report := makeCompareReport("test-fail", "test-fail", false)
	report.DeltaErrors = analysiscompare.StringDelta{
		Added: []string{"ERR1"},
	}
	out := FormatCompareText(report)
	if !strings.Contains(out, "Delta Error Changes") {
		t.Errorf("expected 'Delta Error Changes' section, got:\n%s", out)
	}
}

func TestFormatCompareTextEndsWithNewline(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	out := FormatCompareText(report)
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("expected trailing newline, got:\n%q", out)
	}
}

// ── FormatCompareMarkdown ─────────────────────────────────────────────────────

func TestFormatCompareMarkdownHeader(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	out := FormatCompareMarkdown(report)
	if !strings.Contains(out, "# Faultline Compare") {
		t.Errorf("expected markdown header, got:\n%s", out)
	}
}

func TestFormatCompareMarkdownContainsSources(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	out := FormatCompareMarkdown(report)
	if !strings.Contains(out, "`docker-auth.json`") {
		t.Errorf("expected source in markdown, got:\n%s", out)
	}
}

func TestFormatCompareMarkdownDiagnosisSection(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	out := FormatCompareMarkdown(report)
	if !strings.Contains(out, "## Diagnosis") {
		t.Errorf("expected Diagnosis section in markdown, got:\n%s", out)
	}
}

func TestFormatCompareMarkdownEvidenceSection(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	report.Evidence = analysiscompare.StringDelta{
		Added:   []string{"new evidence"},
		Removed: []string{"old evidence"},
	}
	out := FormatCompareMarkdown(report)
	if !strings.Contains(out, "## Evidence Changes") {
		t.Errorf("expected '## Evidence Changes' in markdown, got:\n%s", out)
	}
	if !strings.Contains(out, "Added: new evidence") {
		t.Errorf("expected 'Added: new evidence' in markdown, got:\n%s", out)
	}
	if !strings.Contains(out, "Removed: old evidence") {
		t.Errorf("expected 'Removed: old evidence' in markdown, got:\n%s", out)
	}
}

func TestFormatCompareMarkdownFixStepsSection(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	report.FixSteps = analysiscompare.StringDelta{Added: []string{"run docker login"}}
	out := FormatCompareMarkdown(report)
	if !strings.Contains(out, "## Fix Step Changes") {
		t.Errorf("expected '## Fix Step Changes' in markdown, got:\n%s", out)
	}
}

func TestFormatCompareMarkdownDominantSignalsSection(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	report.DominantSignals = analysiscompare.StringDelta{Added: []string{"auth-signal"}}
	out := FormatCompareMarkdown(report)
	if !strings.Contains(out, "## Dominant Signal Changes") {
		t.Errorf("expected '## Dominant Signal Changes' in markdown, got:\n%s", out)
	}
}

func TestFormatCompareMarkdownRepoContextSection(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	report.RepoFiles = analysiscompare.StringDelta{Added: []string{"new-file.go"}}
	out := FormatCompareMarkdown(report)
	if !strings.Contains(out, "## Repo Context Changes") {
		t.Errorf("expected '## Repo Context Changes' in markdown, got:\n%s", out)
	}
}

func TestFormatCompareMarkdownDeltaFilesSection(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	report.DeltaFiles = analysiscompare.StringDelta{Added: []string{"main.go"}}
	out := FormatCompareMarkdown(report)
	if !strings.Contains(out, "## Delta File Changes") {
		t.Errorf("expected '## Delta File Changes' in markdown, got:\n%s", out)
	}
}

func TestFormatCompareMarkdownDeltaTestsSection(t *testing.T) {
	report := makeCompareReport("test-fail", "test-fail", false)
	report.DeltaTests = analysiscompare.StringDelta{Added: []string{"TestFoo"}}
	out := FormatCompareMarkdown(report)
	if !strings.Contains(out, "## Delta Test Changes") {
		t.Errorf("expected '## Delta Test Changes' in markdown, got:\n%s", out)
	}
}

func TestFormatCompareMarkdownDeltaErrorsSection(t *testing.T) {
	report := makeCompareReport("test-fail", "test-fail", false)
	report.DeltaErrors = analysiscompare.StringDelta{Added: []string{"ERR1"}}
	out := FormatCompareMarkdown(report)
	if !strings.Contains(out, "## Delta Error Changes") {
		t.Errorf("expected '## Delta Error Changes' in markdown, got:\n%s", out)
	}
}

func TestFormatCompareMarkdownEndsWithNewline(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	out := FormatCompareMarkdown(report)
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("expected trailing newline, got:\n%q", out)
	}
}

func TestFormatCompareMarkdownNilCandidates(t *testing.T) {
	// Report with nil Previous/Current should not panic and should still have header.
	report := analysiscompare.Report{
		LeftSource:  "left.json",
		RightSource: "right.json",
		Summary:     []string{"no material differences were found"},
	}
	out := FormatCompareMarkdown(report)
	if !strings.Contains(out, "# Faultline Compare") {
		t.Errorf("expected header for nil-candidate report, got:\n%s", out)
	}
}

// ── FormatCompareJSON ─────────────────────────────────────────────────────────

func TestFormatCompareJSONIsValidJSON(t *testing.T) {
	report := makeCompareReport("docker-auth", "permission-denied", true)
	out, err := FormatCompareJSON(report)
	if err != nil {
		t.Fatalf("FormatCompareJSON: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := m["diagnosis_changed"]; !ok {
		t.Error("expected 'diagnosis_changed' key in JSON output")
	}
}

func TestFormatCompareJSONEndsWithNewline(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	out, err := FormatCompareJSON(report)
	if err != nil {
		t.Fatalf("FormatCompareJSON: %v", err)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("expected trailing newline, got:\n%q", out)
	}
}

// ── joinCompareSection ────────────────────────────────────────────────────────

func TestJoinCompareSectionEmptyBodyReturnsEmpty(t *testing.T) {
	got := joinCompareSection("Title", "")
	if got != "" {
		t.Errorf("expected empty for empty body, got %q", got)
	}
	got = joinCompareSection("Title", "   ")
	if got != "" {
		t.Errorf("expected empty for whitespace-only body, got %q", got)
	}
}

func TestJoinCompareSectionNonEmpty(t *testing.T) {
	got := joinCompareSection("Summary", "line 1\nline 2")
	if !strings.HasPrefix(got, "Summary\n") {
		t.Errorf("expected title as first line, got %q", got)
	}
	if !strings.Contains(got, strings.Repeat("-", len("Summary"))) {
		t.Errorf("expected underline in section header, got %q", got)
	}
	if !strings.Contains(got, "line 1") {
		t.Errorf("expected body content, got %q", got)
	}
}

// ── compareOverviewText ───────────────────────────────────────────────────────

func TestCompareOverviewTextBothCandidates(t *testing.T) {
	report := makeCompareReport("docker-auth", "permission-denied", true)
	got := compareOverviewText(report)
	if !strings.Contains(got, "Previous: docker-auth") {
		t.Errorf("expected previous in overview, got %q", got)
	}
	if !strings.Contains(got, "Current: permission-denied") {
		t.Errorf("expected current in overview, got %q", got)
	}
	if !strings.Contains(got, "Diagnosis changed: yes") {
		t.Errorf("expected 'Diagnosis changed: yes', got %q", got)
	}
}

func TestCompareOverviewTextNoCandidates(t *testing.T) {
	report := analysiscompare.Report{DiagnosisChanged: false}
	got := compareOverviewText(report)
	// No candidates → no Previous/Current lines; still shows status lines.
	if strings.Contains(got, "Previous:") {
		t.Errorf("did not expect 'Previous:' when Previous is nil, got %q", got)
	}
	if strings.Contains(got, "Current:") {
		t.Errorf("did not expect 'Current:' when Current is nil, got %q", got)
	}
}

func TestCompareOverviewTextStatusChanged(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	report.StatusChanged = true
	got := compareOverviewText(report)
	if !strings.Contains(got, "Artifact status changed: yes") {
		t.Errorf("expected 'Artifact status changed: yes', got %q", got)
	}
}

// ── wrapCode ──────────────────────────────────────────────────────────────────

func TestWrapCodeEmpty(t *testing.T) {
	if got := wrapCode(""); got != "" {
		t.Errorf("expected empty for empty value, got %q", got)
	}
	if got := wrapCode("   "); got != "" {
		t.Errorf("expected empty for whitespace, got %q", got)
	}
}

func TestWrapCodeNonEmpty(t *testing.T) {
	if got := wrapCode("file.json"); got != "`file.json`" {
		t.Errorf("expected backtick-wrapped, got %q", got)
	}
}

// ── Status field in compareOverviewMarkdown ───────────────────────────────────

func TestCompareOverviewMarkdownNoDiagnosisChange(t *testing.T) {
	report := makeCompareReport("docker-auth", "docker-auth", false)
	got := compareOverviewMarkdown(report)
	if !strings.Contains(got, "Diagnosis changed: no") {
		t.Errorf("expected 'Diagnosis changed: no', got %q", got)
	}
	if !strings.Contains(got, "Artifact status changed: no") {
		t.Errorf("expected 'Artifact status changed: no', got %q", got)
	}
}

// ── compareDeltaMarkdown ──────────────────────────────────────────────────────

func TestCompareDeltaMarkdownEmpty(t *testing.T) {
	got := compareDeltaMarkdown(analysiscompare.StringDelta{})
	if got != "" {
		t.Errorf("expected empty for empty delta, got %q", got)
	}
}

func TestCompareDeltaMarkdownWithChanges(t *testing.T) {
	delta := analysiscompare.StringDelta{
		Added:   []string{"added-item"},
		Removed: []string{"removed-item"},
	}
	got := compareDeltaMarkdown(delta)
	if !strings.Contains(got, "Added: added-item") {
		t.Errorf("expected 'Added: added-item', got %q", got)
	}
	if !strings.Contains(got, "Removed: removed-item") {
		t.Errorf("expected 'Removed: removed-item', got %q", got)
	}
}

// ── model.FailureArtifact used via Analysis.Artifact ─────────────────────────

func TestFormatCompareTextWithArtifactFixAndSignals(t *testing.T) {
	a1 := &model.Analysis{
		Results: []model.Result{{
			Playbook: model.Playbook{ID: "docker-auth", Title: "Docker auth"},
			Score:    0.9,
		}},
		Artifact: &model.FailureArtifact{
			FixSteps:        []string{"old fix"},
			DominantSignals: []string{"old signal"},
		},
	}
	a2 := &model.Analysis{
		Results: []model.Result{{
			Playbook: model.Playbook{ID: "docker-auth", Title: "Docker auth"},
			Score:    0.9,
		}},
		Artifact: &model.FailureArtifact{
			FixSteps:        []string{"new fix"},
			DominantSignals: []string{"new signal"},
		},
	}
	report := analysiscompare.Build(a1, a2)
	out := FormatCompareText(report)
	if !strings.Contains(out, "Fix Step Changes") {
		t.Errorf("expected Fix Step Changes section when fix steps differ, got:\n%s", out)
	}
	if !strings.Contains(out, "Dominant Signal Changes") {
		t.Errorf("expected Dominant Signal Changes section, got:\n%s", out)
	}
}
