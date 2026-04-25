// Package silentdetector provides built-in deterministic detectors for
// silent/misleading CI failures — cases where CI appears to succeed but
// important work was silently skipped, suppressed, or missing.
//
// All detectors in this package are internal and conservative: they prefer
// precision over recall.  No public plugin or DSL API is exposed.
package silentdetector

import (
	"sort"
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

// Finding is the internal detector finding type. It aliases the shared model so
// detectors stay minimal while callers can use the stable analysis schema.
type Finding = model.SilentFinding

// Detector is the internal interface for a single built-in silent detector.
// It is intentionally small and is not a public plugin API.
type Detector interface {
	ID() string
	Class() string
	Match(input AnalysisInput) []Finding
}

// all is the ordered list of built-in silent detectors.
// Order is stable and deterministic; it also defines the tiebreak priority
// when multiple findings share the same severity and confidence.
var all = []Detector{
	ignoredExitCodeDetector{},
	continueOnErrorDetector{},
	zeroTestsExecutedDetector{},
	artifactMissingDetector{},
	cacheMissNonFatalDetector{},
	skippedCriticalStepDetector{},
	emptyDeploymentTargetDetector{},
	emptyQualityCheckDetector{},
}

// Run executes all built-in silent detectors against input and returns the
// resulting findings in deterministic order.  Each detector contributes at
// most one finding.
func Run(input AnalysisInput) []model.SilentFinding {
	var out []model.SilentFinding
	for _, d := range all {
		out = append(out, d.Match(input)...)
	}
	order := detectorOrder()
	sort.SliceStable(out, func(i, j int) bool {
		if severityRank(out[i].Severity) != severityRank(out[j].Severity) {
			return severityRank(out[i].Severity) > severityRank(out[j].Severity)
		}
		if confidenceRank(out[i].Confidence) != confidenceRank(out[j].Confidence) {
			return confidenceRank(out[i].Confidence) > confidenceRank(out[j].Confidence)
		}
		return order[out[i].ID] < order[out[j].ID]
	})
	return out
}

func detectorOrder() map[string]int {
	order := make(map[string]int, len(all))
	for i, d := range all {
		order[d.ID()] = i
	}
	return order
}

// SelectPrimary picks the highest-priority finding from a non-empty slice.
// It assumes the slice is already ranked by Run (severity desc, confidence
// desc, detector order asc) and simply returns the first element.
func SelectPrimary(findings []model.SilentFinding) *model.SilentFinding {
	if len(findings) == 0 {
		return nil
	}
	best := findings[0]
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

func (ignoredExitCodeDetector) ID() string    { return "ignored-exit-code" }
func (ignoredExitCodeDetector) Class() string { return "silent_failure" }

func (ignoredExitCodeDetector) Match(input AnalysisInput) []Finding {
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
	return []Finding{{
		ID:          ignoredExitCodeDetector{}.ID(),
		Class:       ignoredExitCodeDetector{}.Class(),
		Severity:    "high",
		Confidence:  "high",
		Explanation: "A command failure was explicitly suppressed (|| true, set +e, or similar), allowing the CI job to continue without surfacing the error.",
		Evidence:    evidence,
	}}
}

// ── Detector B: continue-on-error ────────────────────────────────────────────

type continueOnErrorDetector struct{}

func (continueOnErrorDetector) ID() string    { return "continue-on-error" }
func (continueOnErrorDetector) Class() string { return "silent_failure" }

func (continueOnErrorDetector) Match(input AnalysisInput) []Finding {
	triggers := []string{
		"continue-on-error: true",
		"continueOnError: true",
		"allow_failure: true",
	}
	evidence := matchingLines(input.Lines, 3, triggers...)
	if len(evidence) == 0 {
		return nil
	}
	return []Finding{{
		ID:          continueOnErrorDetector{}.ID(),
		Class:       continueOnErrorDetector{}.Class(),
		Severity:    "high",
		Confidence:  "high",
		Explanation: "A CI step is configured to continue on error (continue-on-error: true or allow_failure: true), which can mask real failures.",
		Evidence:    evidence,
	}}
}

// ── Detector C: zero-tests-executed ──────────────────────────────────────────

type zeroTestsExecutedDetector struct{}

func (zeroTestsExecutedDetector) ID() string    { return "zero-tests-executed" }
func (zeroTestsExecutedDetector) Class() string { return "silent_failure" }

func (zeroTestsExecutedDetector) Match(input AnalysisInput) []Finding {
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
	return []Finding{{
		ID:          zeroTestsExecutedDetector{}.ID(),
		Class:       zeroTestsExecutedDetector{}.Class(),
		Severity:    "high",
		Confidence:  "high",
		Explanation: "A test command appeared to run, but no tests were discovered or executed.",
		Evidence:    evidence,
	}}
}

// ── Detector D: artifact-missing ─────────────────────────────────────────────

type artifactMissingDetector struct{}

func (artifactMissingDetector) ID() string    { return "artifact-missing" }
func (artifactMissingDetector) Class() string { return "silent_failure" }

func (artifactMissingDetector) Match(input AnalysisInput) []Finding {
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
	return []Finding{{
		ID:          artifactMissingDetector{}.ID(),
		Class:       artifactMissingDetector{}.Class(),
		Severity:    "high",
		Confidence:  "high",
		Explanation: "An artifact upload or report step ran, but no files were found or uploaded.",
		Evidence:    evidence,
	}}
}

// ── Detector E: cache-miss-non-fatal ─────────────────────────────────────────

type cacheMissNonFatalDetector struct{}

func (cacheMissNonFatalDetector) ID() string    { return "cache-miss-non-fatal" }
func (cacheMissNonFatalDetector) Class() string { return "silent_failure" }

func (cacheMissNonFatalDetector) Match(input AnalysisInput) []Finding {
	// "cache miss" is intentionally absent here — it also appears in
	// failSignals and would satisfy the AND on a single line.
	cacheSignals := []string{
		"restore cache", "save cache",
		"actions/cache", "cache restore",
		"cache hit",
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
	return []Finding{{
		ID:          cacheMissNonFatalDetector{}.ID(),
		Class:       cacheMissNonFatalDetector{}.Class(),
		Severity:    "medium",
		Confidence:  "medium",
		Explanation: "A cache restore or save step failed, but the job continued without surfacing this as a failure. Repeated cache misses degrade CI performance and may mask dependency issues.",
		Evidence:    evidence,
	}}
}

// ── Detector F: skipped-critical-step ────────────────────────────────────────

type skippedCriticalStepDetector struct{}

func (skippedCriticalStepDetector) ID() string    { return "skipped-critical-step" }
func (skippedCriticalStepDetector) Class() string { return "silent_failure" }

func (skippedCriticalStepDetector) Match(input AnalysisInput) []Finding {
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
	return []Finding{{
		ID:          skippedCriticalStepDetector{}.ID(),
		Class:       skippedCriticalStepDetector{}.Class(),
		Severity:    "high",
		Confidence:  "medium",
		Explanation: "A CI step was skipped due to a condition, and the context suggests it may have been a critical step (build, test, deploy, security, coverage, or report).",
		Evidence:    evidence,
	}}
}

// ── Detector G: empty-deployment-target ──────────────────────────────────────

type emptyDeploymentTargetDetector struct{}

func (emptyDeploymentTargetDetector) ID() string    { return "empty-deployment-target" }
func (emptyDeploymentTargetDetector) Class() string { return "silent_failure" }

func (emptyDeploymentTargetDetector) Match(input AnalysisInput) []Finding {
	// Prefer explicit tool names to avoid matching the generic token "deploy"
	// inside emptySignals phrases such as "Nothing to deploy".
	deploySignals := []string{
		"kubectl apply", "helm upgrade", "helm install", "terraform apply",
		"serverless deploy",
	}
	emptySignals := []string{
		"nothing to deploy",
		"no files to deploy",
		"no manifests found",
		"no objects passed to apply",
		"0 artifacts to publish",
		"no packages found to publish",
		"no deployable artifacts",
	}
	if !containsAny(input.RawLog, deploySignals...) || !containsAny(input.RawLog, emptySignals...) {
		return nil
	}
	evidence := matchingLines(input.Lines, 5, append(deploySignals, emptySignals...)...)
	if len(evidence) == 0 {
		return nil
	}
	return []Finding{{
		ID:          emptyDeploymentTargetDetector{}.ID(),
		Class:       emptyDeploymentTargetDetector{}.Class(),
		Severity:    "high",
		Confidence:  "medium",
		Explanation: "A deploy or publish step ran, but the logs indicate there was nothing deployable to apply, publish, or release.",
		Evidence:    evidence,
	}}
}

// ── Detector H: empty-quality-check ──────────────────────────────────────────

type emptyQualityCheckDetector struct{}

func (emptyQualityCheckDetector) ID() string    { return "empty-quality-check" }
func (emptyQualityCheckDetector) Class() string { return "silent_failure" }

func (emptyQualityCheckDetector) Match(input AnalysisInput) []Finding {
	checkSignals := []string{
		"eslint", "golangci-lint", "ruff", "flake8", "shellcheck",
		"semgrep", "trivy", "snyk", "bandit", "codeql", "coverage",
		"lint", "scan", "security scan",
	}
	emptySignals := []string{
		"no files to lint",
		"0 files checked",
		"0 files scanned",
		"0 packages scanned",
		"nothing to scan",
		"no source files found",
		"0 files analyzed",
		"coverage report not generated",
		"no files were analyzed",
	}

	var hasCheckContext bool
	var hasEmptyResult bool
	for _, l := range input.Lines {
		line := strings.TrimSpace(l.Normalized)
		if line == "" {
			continue
		}

		matchesCheck := containsAny(line, checkSignals...)
		matchesEmpty := containsAny(line, emptySignals...)

		if matchesCheck && !matchesEmpty {
			hasCheckContext = true
		}
		if matchesEmpty {
			hasEmptyResult = true
		}
		if hasCheckContext && hasEmptyResult {
			break
		}
	}

	if !hasCheckContext || !hasEmptyResult {
		return nil
	}
	evidence := matchingLines(input.Lines, 5, append(checkSignals, emptySignals...)...)
	if len(evidence) == 0 {
		return nil
	}
	return []Finding{{
		ID:          emptyQualityCheckDetector{}.ID(),
		Class:       emptyQualityCheckDetector{}.Class(),
		Severity:    "medium",
		Confidence:  "medium",
		Explanation: "A lint, scan, or coverage step ran, but it processed no files or generated no usable report.",
		Evidence:    evidence,
	}}
}
