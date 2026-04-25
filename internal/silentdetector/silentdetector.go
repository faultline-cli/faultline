// Package silentdetector provides built-in deterministic detectors for
// silent/misleading CI failures — cases where CI appears to succeed but
// important work was silently skipped, suppressed, or missing.
//
// All detectors in this package are internal and conservative: they prefer
// precision over recall.  No public plugin or DSL API is exposed.
package silentdetector

import (
	"strings"

	"faultline/internal/model"
)

// AnalysisInput is the read-only view of a log that silent detectors receive.
type AnalysisInput struct {
	// Lines is the normalised log as returned by engine.ReadLines.
	Lines []model.Line
	// RawLog is the full original log text joined together.
	RawLog string
}

// detector is the internal interface for a single built-in silent detector.
type detector interface {
	id() string
	match(input AnalysisInput) *model.SilentFinding
}

// all is the ordered list of built-in silent detectors.
// Order is stable and deterministic; it also defines the tiebreak priority
// when multiple findings share the same severity and confidence.
var all = []detector{
	ignoredExitCodeDetector{},
	continueOnErrorDetector{},
	zeroTestsExecutedDetector{},
	artifactMissingDetector{},
	cacheMissNonFatalDetector{},
	skippedCriticalStepDetector{},
}

// Run executes all built-in silent detectors against input and returns the
// resulting findings in deterministic order.  Each detector contributes at
// most one finding.
func Run(input AnalysisInput) []model.SilentFinding {
	var out []model.SilentFinding
	for _, d := range all {
		if f := d.match(input); f != nil {
			out = append(out, *f)
		}
	}
	return out
}

// SelectPrimary picks the highest-priority finding from a non-empty slice
// using a deterministic ranking: severity (high > medium > low) then
// confidence (high > medium > low) then detector order (already stable).
func SelectPrimary(findings []model.SilentFinding) *model.SilentFinding {
	if len(findings) == 0 {
		return nil
	}
	best := findings[0]
	for _, f := range findings[1:] {
		if severityRank(f.Severity) > severityRank(best.Severity) {
			best = f
			continue
		}
		if severityRank(f.Severity) == severityRank(best.Severity) &&
			confidenceRank(f.Confidence) > confidenceRank(best.Confidence) {
			best = f
		}
	}
	return &best
}

func severityRank(s string) int {
	switch strings.ToLower(s) {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func confidenceRank(c string) int {
	switch strings.ToLower(c) {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// containsAny reports whether any of the given substrings (case-insensitive)
// appear in s.
func containsAny(s string, subs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range subs {
		if strings.Contains(lower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

// matchingLines returns original log lines that contain any of the given
// substrings (case-insensitive), deduplicated and limited to maxLines.
func matchingLines(lines []model.Line, maxLines int, subs ...string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, l := range lines {
		if !containsAny(l.Original, subs...) {
			continue
		}
		trimmed := strings.TrimSpace(l.Original)
		if _, dup := seen[trimmed]; dup {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
		if len(out) >= maxLines {
			break
		}
	}
	return out
}

// ── Detector A: ignored-exit-code ────────────────────────────────────────────

type ignoredExitCodeDetector struct{}

func (ignoredExitCodeDetector) id() string { return "ignored-exit-code" }

func (ignoredExitCodeDetector) match(input AnalysisInput) *model.SilentFinding {
	// Patterns that strongly suggest an exit-code was deliberately suppressed.
	triggers := []string{"|| true", "set +e", "failed but continuing", "ignoring error", "exit 0"}
	var evidence []string
	for _, l := range input.Lines {
		if containsAny(l.Original, triggers...) {
			evidence = append(evidence, strings.TrimSpace(l.Original))
			if len(evidence) >= 3 {
				break
			}
		}
	}
	if len(evidence) == 0 {
		return nil
	}
	return &model.SilentFinding{
		ID:          ignoredExitCodeDetector{}.id(),
		Class:       "silent_failure",
		Severity:    "high",
		Confidence:  "high",
		Explanation: "A command failure was explicitly suppressed (|| true, set +e, or similar), allowing the CI job to continue without surfacing the error.",
		Evidence:    evidence,
	}
}

// ── Detector B: continue-on-error ────────────────────────────────────────────

type continueOnErrorDetector struct{}

func (continueOnErrorDetector) id() string { return "continue-on-error" }

func (continueOnErrorDetector) match(input AnalysisInput) *model.SilentFinding {
	triggers := []string{
		"continue-on-error: true",
		"continueOnError: true",
		"allow_failure: true",
	}
	evidence := matchingLines(input.Lines, 3, triggers...)
	if len(evidence) == 0 {
		return nil
	}
	return &model.SilentFinding{
		ID:          continueOnErrorDetector{}.id(),
		Class:       "silent_failure",
		Severity:    "high",
		Confidence:  "high",
		Explanation: "A CI step is configured to continue on error (continue-on-error: true or allow_failure: true), which can mask real failures.",
		Evidence:    evidence,
	}
}

// ── Detector C: zero-tests-executed ──────────────────────────────────────────

type zeroTestsExecutedDetector struct{}

func (zeroTestsExecutedDetector) id() string { return "zero-tests-executed" }

func (zeroTestsExecutedDetector) match(input AnalysisInput) *model.SilentFinding {
	// Must see a test command hint AND a zero-test signal.
	testCommands := []string{
		"npm test", "yarn test", "pnpm test",
		"go test", "pytest", "jest", "mocha",
		"rspec", "cargo test", "mvn test", "gradle test",
		"make test", "bundle exec rspec",
	}
	zeroSignals := []string{
		"0 tests", "0 passed", "0 examples",
		"no tests found", "no test files found",
		"collected 0 items",
		"no test suites found",
		"tests: 0",
	}
	log := input.RawLog
	hasTestCmd := containsAny(log, testCommands...)
	hasZeroSignal := containsAny(log, zeroSignals...)
	if !hasTestCmd || !hasZeroSignal {
		return nil
	}
	evidence := matchingLines(input.Lines, 5, append(testCommands, zeroSignals...)...)
	if len(evidence) == 0 {
		return nil
	}
	return &model.SilentFinding{
		ID:          zeroTestsExecutedDetector{}.id(),
		Class:       "silent_failure",
		Severity:    "high",
		Confidence:  "high",
		Explanation: "A test command appeared to run, but no tests were discovered or executed.",
		Evidence:    evidence,
	}
}

// ── Detector D: artifact-missing ─────────────────────────────────────────────

type artifactMissingDetector struct{}

func (artifactMissingDetector) id() string { return "artifact-missing" }

func (artifactMissingDetector) match(input AnalysisInput) *model.SilentFinding {
	uploadSignals := []string{
		"upload artifact", "upload-artifact",
		"upload report", "publish artifact",
		"actions/upload-artifact",
	}
	missingSignals := []string{
		"no files found",
		"artifact not found",
		"skipping upload",
		"no matching files",
		"path does not exist",
		"0 files uploaded",
		"no artifacts",
	}
	log := input.RawLog
	hasUpload := containsAny(log, uploadSignals...)
	hasMissing := containsAny(log, missingSignals...)
	if !hasUpload || !hasMissing {
		return nil
	}
	evidence := matchingLines(input.Lines, 5, append(uploadSignals, missingSignals...)...)
	if len(evidence) == 0 {
		return nil
	}
	return &model.SilentFinding{
		ID:          artifactMissingDetector{}.id(),
		Class:       "silent_failure",
		Severity:    "high",
		Confidence:  "high",
		Explanation: "An artifact upload or report step ran, but no files were found or uploaded.",
		Evidence:    evidence,
	}
}

// ── Detector E: cache-miss-non-fatal ─────────────────────────────────────────

type cacheMissNonFatalDetector struct{}

func (cacheMissNonFatalDetector) id() string { return "cache-miss-non-fatal" }

func (cacheMissNonFatalDetector) match(input AnalysisInput) *model.SilentFinding {
	cacheSignals := []string{
		"restore cache", "save cache",
		"actions/cache", "cache restore",
		"cache hit", "cache miss",
	}
	failSignals := []string{
		"cache not found",
		"failed to restore cache",
		"failed to save cache",
		"cache miss",
		"no cache found",
	}
	log := input.RawLog
	hasCacheOp := containsAny(log, cacheSignals...)
	hasFail := containsAny(log, failSignals...)
	if !hasCacheOp || !hasFail {
		return nil
	}
	evidence := matchingLines(input.Lines, 5, append(cacheSignals, failSignals...)...)
	if len(evidence) == 0 {
		return nil
	}
	return &model.SilentFinding{
		ID:          cacheMissNonFatalDetector{}.id(),
		Class:       "silent_failure",
		Severity:    "medium",
		Confidence:  "medium",
		Explanation: "A cache restore or save step failed, but the job continued without surfacing this as a failure. Repeated cache misses degrade CI performance and may mask dependency issues.",
		Evidence:    evidence,
	}
}

// ── Detector F: skipped-critical-step ────────────────────────────────────────

type skippedCriticalStepDetector struct{}

func (skippedCriticalStepDetector) id() string { return "skipped-critical-step" }

func (skippedCriticalStepDetector) match(input AnalysisInput) *model.SilentFinding {
	skipSignals := []string{
		"skipping step due to condition",
		"condition evaluated to false",
		"step was skipped",
		"##[warning]step skipped",
		"##[debug]step skipped",
	}
	// Only flag as critical if the skipped step is near a critical domain.
	criticalDomains := []string{
		"test", "deploy", "build", "security",
		"coverage", "report", "publish", "scan",
	}
	log := input.RawLog
	hasSkip := containsAny(log, skipSignals...)
	hasCritical := containsAny(log, criticalDomains...)
	if !hasSkip || !hasCritical {
		return nil
	}
	evidence := matchingLines(input.Lines, 5, skipSignals...)
	if len(evidence) == 0 {
		return nil
	}
	return &model.SilentFinding{
		ID:          skippedCriticalStepDetector{}.id(),
		Class:       "silent_failure",
		Severity:    "high",
		Confidence:  "medium",
		Explanation: "A CI step was skipped due to a condition, and the context suggests it may have been a critical step (build, test, deploy, security, coverage, or report).",
		Evidence:    evidence,
	}
}
