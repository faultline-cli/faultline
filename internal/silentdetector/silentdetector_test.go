package silentdetector_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/model"
	"faultline/internal/silentdetector"
)

func readLog(t *testing.T, name string) silentdetector.AnalysisInput {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	text := string(raw)
	lines := splitLines(text)
	return silentdetector.AnalysisInput{Lines: lines, RawLog: text}
}

func splitLines(text string) []model.Line {
	raw := strings.Split(text, "\n")
	lines := make([]model.Line, 0, len(raw))
	for i, l := range raw {
		lines = append(lines, model.Line{Original: l, Normalized: strings.ToLower(l), Number: i + 1})
	}
	return lines
}

func findingID(findings []model.SilentFinding, id string) *model.SilentFinding {
	for _, f := range findings {
		if f.ID == id {
			return &f
		}
	}
	return nil
}

// ── Individual detector fixtures ─────────────────────────────────────────────

func TestIgnoredExitCodeDetector(t *testing.T) {
	input := readLog(t, "ignored-exit-code.log")
	findings := silentdetector.Run(input)
	f := findingID(findings, "ignored-exit-code")
	if f == nil {
		t.Fatal("expected ignored-exit-code finding, got none")
	}
	if f.Class != "silent_failure" {
		t.Errorf("expected class silent_failure, got %q", f.Class)
	}
	if f.Severity != "high" {
		t.Errorf("expected severity high, got %q", f.Severity)
	}
	if len(f.Evidence) == 0 {
		t.Error("expected evidence, got none")
	}
}

func TestContinueOnErrorDetector(t *testing.T) {
	input := readLog(t, "continue-on-error.log")
	findings := silentdetector.Run(input)
	f := findingID(findings, "continue-on-error")
	if f == nil {
		t.Fatal("expected continue-on-error finding, got none")
	}
	if f.Severity != "high" {
		t.Errorf("expected severity high, got %q", f.Severity)
	}
}

func TestZeroTestsNPM(t *testing.T) {
	input := readLog(t, "zero-tests-npm.log")
	findings := silentdetector.Run(input)
	f := findingID(findings, "zero-tests-executed")
	if f == nil {
		t.Fatal("expected zero-tests-executed finding, got none")
	}
	if f.Severity != "high" {
		t.Errorf("expected severity high, got %q", f.Severity)
	}
}

func TestZeroTestsPytest(t *testing.T) {
	input := readLog(t, "zero-tests-pytest.log")
	findings := silentdetector.Run(input)
	f := findingID(findings, "zero-tests-executed")
	if f == nil {
		t.Fatal("expected zero-tests-executed finding, got none")
	}
}

func TestArtifactMissingDetector(t *testing.T) {
	input := readLog(t, "artifact-missing.log")
	findings := silentdetector.Run(input)
	f := findingID(findings, "artifact-missing")
	if f == nil {
		t.Fatal("expected artifact-missing finding, got none")
	}
	if f.Severity != "high" {
		t.Errorf("expected severity high, got %q", f.Severity)
	}
}

func TestCacheMissNonFatalDetector(t *testing.T) {
	input := readLog(t, "cache-miss.log")
	findings := silentdetector.Run(input)
	f := findingID(findings, "cache-miss-non-fatal")
	if f == nil {
		t.Fatal("expected cache-miss-non-fatal finding, got none")
	}
	if f.Severity != "medium" {
		t.Errorf("expected severity medium, got %q", f.Severity)
	}
}

func TestSkippedCriticalStepDetector(t *testing.T) {
	input := readLog(t, "skipped-critical-step.log")
	findings := silentdetector.Run(input)
	f := findingID(findings, "skipped-critical-step")
	if f == nil {
		t.Fatal("expected skipped-critical-step finding, got none")
	}
}

func TestEmptyDeploymentTargetDetector(t *testing.T) {
	input := readLog(t, "empty-deployment-target.log")
	findings := silentdetector.Run(input)
	f := findingID(findings, "empty-deployment-target")
	if f == nil {
		t.Fatal("expected empty-deployment-target finding, got none")
	}
	if f.Severity != "high" {
		t.Errorf("expected severity high, got %q", f.Severity)
	}
}

func TestEmptyQualityCheckDetector(t *testing.T) {
	input := readLog(t, "empty-quality-check.log")
	findings := silentdetector.Run(input)
	f := findingID(findings, "empty-quality-check")
	if f == nil {
		t.Fatal("expected empty-quality-check finding, got none")
	}
	if f.Severity != "medium" {
		t.Errorf("expected severity medium, got %q", f.Severity)
	}
}

// ── Combined / integration scenarios ─────────────────────────────────────────

func TestNoSilentFindingsOnCleanLog(t *testing.T) {
	input := readLog(t, "no-silent.log")
	findings := silentdetector.Run(input)
	if len(findings) > 0 {
		t.Errorf("expected no findings on clean log, got %d: %v", len(findings), findings)
	}
}

// TestMultipleSilentFindings verifies that multiple detectors can fire and that
// SelectPrimary returns the highest-severity finding deterministically.
func TestMultipleSilentFindings(t *testing.T) {
	input := readLog(t, "multiple-silent.log")
	findings := silentdetector.Run(input)
	if len(findings) < 2 {
		t.Fatalf("expected at least 2 findings, got %d", len(findings))
	}
	primary := silentdetector.SelectPrimary(findings)
	if primary == nil {
		t.Fatal("SelectPrimary returned nil")
	}
	// Primary should be high severity (ignored-exit-code or zero-tests or artifact-missing).
	if primary.Severity != "high" {
		t.Errorf("expected primary severity high, got %q", primary.Severity)
	}
}

// TestSelectPrimaryDeterminism verifies that SelectPrimary is stable when called
// twice on the same input.
func TestSelectPrimaryDeterminism(t *testing.T) {
	input := readLog(t, "multiple-silent.log")
	findings := silentdetector.Run(input)
	p1 := silentdetector.SelectPrimary(findings)
	p2 := silentdetector.SelectPrimary(findings)
	if p1 == nil || p2 == nil {
		t.Skip("no findings")
	}
	if p1.ID != p2.ID {
		t.Errorf("SelectPrimary is not deterministic: got %q then %q", p1.ID, p2.ID)
	}
}

func TestSelectPrimaryUsesSeverityAndConfidenceOrdering(t *testing.T) {
	findings := []model.SilentFinding{
		{ID: "cache-miss-non-fatal", Severity: "medium", Confidence: "medium"},
		{ID: "ignored-exit-code", Severity: "high", Confidence: "high"},
	}
	primary := silentdetector.SelectPrimary(findings)
	if primary == nil {
		t.Fatal("expected primary finding")
	}
	if primary.ID != "cache-miss-non-fatal" {
		t.Fatalf("SelectPrimary should return the first finding from a pre-ranked slice; got %q", primary.ID)
	}

	input := readLog(t, "multiple-silent.log")
	ranked := silentdetector.Run(input)
	best := silentdetector.SelectPrimary(ranked)
	if best == nil {
		t.Fatal("expected ranked primary finding")
	}
	if best.Severity != "high" {
		t.Fatalf("expected high severity primary, got %q", best.Severity)
	}
}

// TestSelectPrimaryEmpty ensures nil is returned for empty input.
func TestSelectPrimaryEmpty(t *testing.T) {
	if p := silentdetector.SelectPrimary(nil); p != nil {
		t.Errorf("expected nil for empty findings, got %+v", p)
	}
	if p := silentdetector.SelectPrimary([]model.SilentFinding{}); p != nil {
		t.Errorf("expected nil for empty findings, got %+v", p)
	}
}

// TestEvidenceNotEmpty ensures each detector includes evidence.
func TestEvidenceNotEmpty(t *testing.T) {
	fixtures := []struct {
		file       string
		detectorID string
	}{
		{"ignored-exit-code.log", "ignored-exit-code"},
		{"continue-on-error.log", "continue-on-error"},
		{"zero-tests-npm.log", "zero-tests-executed"},
		{"artifact-missing.log", "artifact-missing"},
		{"cache-miss.log", "cache-miss-non-fatal"},
		{"skipped-critical-step.log", "skipped-critical-step"},
		{"empty-deployment-target.log", "empty-deployment-target"},
		{"empty-quality-check.log", "empty-quality-check"},
	}
	for _, tc := range fixtures {
		t.Run(tc.detectorID, func(t *testing.T) {
			input := readLog(t, tc.file)
			findings := silentdetector.Run(input)
			f := findingID(findings, tc.detectorID)
			if f == nil {
				t.Fatalf("expected finding %s", tc.detectorID)
			}
			if len(f.Evidence) == 0 {
				t.Errorf("finding %s has no evidence", tc.detectorID)
			}
		})
	}
}

// TestSilentFindingClass verifies every detector returns "silent_failure" as the class.
func TestSilentFindingClass(t *testing.T) {
	fixtures := []string{
		"ignored-exit-code.log",
		"continue-on-error.log",
		"zero-tests-npm.log",
		"artifact-missing.log",
		"cache-miss.log",
		"skipped-critical-step.log",
		"empty-deployment-target.log",
		"empty-quality-check.log",
	}
	for _, file := range fixtures {
		input := readLog(t, file)
		findings := silentdetector.Run(input)
		for _, f := range findings {
			if f.Class != "silent_failure" {
				t.Errorf("%s: finding %s has class %q, expected silent_failure", file, f.ID, f.Class)
			}
		}
	}
}
