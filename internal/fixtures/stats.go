package fixtures

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"faultline/internal/engine"
	"faultline/internal/model"
)

type EvaluateOptions struct {
	PlaybookDir      string
	PlaybookPackDirs []string
	NoHistory        bool
}

type EvaluatedFixture struct {
	Fixture         Fixture
	Analysis        *model.Analysis
	Err             error
	ExpectedRank    int
	Success         bool
	FalsePositive   bool
	WeakMatch       bool
	Unmatched       bool
	StageMismatch   bool
	DisallowedHits  []string
	UnexpectedTopN  int
	PredictedTopIDs []string
}

type Report struct {
	Class               Class
	Fixtures            []EvaluatedFixture
	FixtureCount        int
	Top1Count           int
	Top3Count           int
	UnmatchedCount      int
	FalsePositiveCount  int
	WeakMatchCount      int
	RecurringPatterns   map[string]int
	Providers           map[string]int
	Adapters            map[string]int
	UnmatchedFixtureIDs []string
	WeakMatchFixtureIDs []string
	ThresholdViolations []string
	AppliedThresholds   Thresholds
	AppliedBaselinePath string
	AppliedBaselineHash string
}

func Evaluate(layout Layout, class Class, opts EvaluateOptions) (Report, error) {
	loaded, err := Load(layout, class)
	if err != nil {
		return Report{}, err
	}
	return EvaluateFixtures(layout, class, loaded, opts)
}

func EvaluateFixtures(layout Layout, class Class, loaded []Fixture, opts EvaluateOptions) (Report, error) {
	if issues := validateFixtureMetadata(loaded); len(issues) > 0 {
		return Report{}, fmt.Errorf("fixture metadata validation failed: %s", strings.Join(issues, "; "))
	}

	e := engine.New(engine.Options{
		PlaybookDir:      opts.PlaybookDir,
		PlaybookPackDirs: opts.PlaybookPackDirs,
		NoHistory:        opts.NoHistory,
	})
	report := Report{
		Class:             class,
		RecurringPatterns: map[string]int{},
		Providers:         map[string]int{},
		Adapters:          map[string]int{},
	}
	for _, fixture := range loaded {
		incrementBucket(report.Providers, fixture.Source.Provider)
		incrementBucket(report.Adapters, fixture.Source.Adapter)
		if strings.TrimSpace(fixture.Expectation.ExpectedPlaybook) == "" && fixture.effectiveClass() != ClassStaging {
			return Report{}, fmt.Errorf("fixture %s is missing expectation.expected_playbook", fixture.ID)
		}
		logText, err := fixtureLog(fixture, layout.Root)
		if err != nil {
			return Report{}, err
		}
		analysis, analyzeErr := e.AnalyzeReader(strings.NewReader(logText))
		item := evaluateFixture(fixture, analysis, analyzeErr)
		report.Fixtures = append(report.Fixtures, item)
		if strings.TrimSpace(fixture.Expectation.ExpectedPlaybook) == "" {
			continue
		}
		report.FixtureCount++
		report.RecurringPatterns[fixture.Expectation.ExpectedPlaybook]++
		if item.ExpectedRank == 1 {
			report.Top1Count++
		}
		if item.ExpectedRank > 0 && item.ExpectedRank <= 3 {
			report.Top3Count++
		}
		if item.Unmatched {
			report.UnmatchedCount++
			report.UnmatchedFixtureIDs = append(report.UnmatchedFixtureIDs, fixture.ID)
		}
		if item.FalsePositive {
			report.FalsePositiveCount++
		}
		if item.WeakMatch {
			report.WeakMatchCount++
			report.WeakMatchFixtureIDs = append(report.WeakMatchFixtureIDs, fixture.ID)
		}
	}
	sort.Strings(report.UnmatchedFixtureIDs)
	sort.Strings(report.WeakMatchFixtureIDs)
	return report, nil
}

func evaluateFixture(f Fixture, analysis *model.Analysis, analyzeErr error) EvaluatedFixture {
	item := EvaluatedFixture{Fixture: f, Analysis: analysis, Err: analyzeErr}
	if analysis != nil {
		for i, result := range analysis.Results {
			item.PredictedTopIDs = append(item.PredictedTopIDs, result.Playbook.ID)
			if result.Playbook.ID == f.Expectation.ExpectedPlaybook && item.ExpectedRank == 0 {
				item.ExpectedRank = i + 1
			}
		}
	}
	if strings.TrimSpace(f.Expectation.ExpectedPlaybook) == "" {
		item.Success = analyzeErr == nil && analysis != nil && len(analysis.Results) > 0
		return item
	}
	if analyzeErr != nil && !errors.Is(analyzeErr, engine.ErrNoMatch) {
		item.Unmatched = true
		return item
	}
	allowRank := f.allowedRank()
	if f.isStrictTop1() {
		item.Success = item.ExpectedRank == 1
	} else {
		item.Success = item.ExpectedRank > 0 && item.ExpectedRank <= allowRank
	}
	item.Unmatched = !item.Success
	if analysis != nil && f.Expectation.ExpectedStage != "" && analysis.Context.Stage != f.Expectation.ExpectedStage {
		item.StageMismatch = true
		item.Unmatched = true
		item.Success = false
	}
	if analysis != nil && len(analysis.Results) > 0 && analysis.Results[0].Playbook.ID != f.Expectation.ExpectedPlaybook {
		item.FalsePositive = true
	}
	if analysis != nil {
		topLimit := min(len(analysis.Results), max(allowRank, 3))
		unexpected := 0
		for _, result := range analysis.Results[:topLimit] {
			if result.Playbook.ID != f.Expectation.ExpectedPlaybook {
				unexpected++
			}
			for _, disallowed := range f.Expectation.DisallowedPlaybooks {
				if result.Playbook.ID == disallowed {
					item.DisallowedHits = append(item.DisallowedHits, disallowed)
					item.FalsePositive = true
				}
			}
		}
		item.UnexpectedTopN = unexpected
		if f.Expectation.MaxUnexpectedTopN > 0 && unexpected > f.Expectation.MaxUnexpectedTopN {
			item.FalsePositive = true
		}
		if item.ExpectedRank > 0 {
			confidence := analysis.Results[item.ExpectedRank-1].Confidence
			if item.ExpectedRank > 1 || confidence < f.confidenceFloor() {
				item.WeakMatch = true
			}
		}
	}
	return item
}

func (r Report) Top1Rate() float64 {
	return ratio(r.Top1Count, r.FixtureCount)
}

func (r Report) Top3Rate() float64 {
	return ratio(r.Top3Count, r.FixtureCount)
}

func (r Report) UnmatchedRate() float64 {
	return ratio(r.UnmatchedCount, r.FixtureCount)
}

func (r Report) FalsePositiveRate() float64 {
	return ratio(r.FalsePositiveCount, r.FixtureCount)
}

func (r Report) WeakMatchRate() float64 {
	return ratio(r.WeakMatchCount, r.FixtureCount)
}

func (r Report) BaselineFingerprint() string {
	parts := []string{
		fmt.Sprintf("%s:%d", r.Class, r.FixtureCount),
		fmt.Sprintf("%.4f", r.Top1Rate()),
		fmt.Sprintf("%.4f", r.Top3Rate()),
		fmt.Sprintf("%.4f", r.UnmatchedRate()),
		fmt.Sprintf("%.4f", r.FalsePositiveRate()),
		fmt.Sprintf("%.4f", r.WeakMatchRate()),
	}
	parts = appendCountFingerprint(parts, "playbook", r.RecurringPatterns)
	parts = appendCountFingerprint(parts, "provider", r.Providers)
	parts = appendCountFingerprint(parts, "adapter", r.Adapters)
	return FingerprintForLog(strings.Join(parts, "|"))
}

func (r Report) Baseline(thresholds Thresholds) Baseline {
	return Baseline{
		Class:             r.Class,
		FixtureCount:      r.FixtureCount,
		Top1Rate:          r.Top1Rate(),
		Top3Rate:          r.Top3Rate(),
		UnmatchedRate:     r.UnmatchedRate(),
		FalsePositiveRate: r.FalsePositiveRate(),
		WeakMatchRate:     r.WeakMatchRate(),
		Thresholds:        thresholds,
		GeneratedAt:       time.Now().UTC().Format(time.RFC3339),
		Fingerprint:       r.BaselineFingerprint(),
	}
}

func LoadBaseline(path string) (Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Baseline{}, err
	}
	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return Baseline{}, fmt.Errorf("parse baseline %s: %w", path, err)
	}
	return baseline, nil
}

func WriteBaseline(path string, baseline Baseline) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create baseline directory: %w", err)
	}
	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal baseline: %w", err)
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func CheckBaseline(report *Report, baseline Baseline) error {
	thresholds := baseline.Thresholds
	report.AppliedThresholds = thresholds
	report.AppliedBaselineHash = baseline.Fingerprint
	if report.FixtureCount != baseline.FixtureCount {
		report.ThresholdViolations = append(report.ThresholdViolations,
			fmt.Sprintf("fixture count changed: baseline=%d current=%d", baseline.FixtureCount, report.FixtureCount),
		)
	}
	if report.Top1Rate() < thresholds.MinTop1 {
		report.ThresholdViolations = append(report.ThresholdViolations,
			fmt.Sprintf("top-1 rate %.3f is below threshold %.3f", report.Top1Rate(), thresholds.MinTop1),
		)
	}
	if report.Top3Rate() < thresholds.MinTop3 {
		report.ThresholdViolations = append(report.ThresholdViolations,
			fmt.Sprintf("top-3 rate %.3f is below threshold %.3f", report.Top3Rate(), thresholds.MinTop3),
		)
	}
	if report.UnmatchedRate() > thresholds.MaxUnmatched {
		report.ThresholdViolations = append(report.ThresholdViolations,
			fmt.Sprintf("unmatched rate %.3f exceeds threshold %.3f", report.UnmatchedRate(), thresholds.MaxUnmatched),
		)
	}
	if report.FalsePositiveRate() > thresholds.MaxFalsePositive {
		report.ThresholdViolations = append(report.ThresholdViolations,
			fmt.Sprintf("false-positive rate %.3f exceeds threshold %.3f", report.FalsePositiveRate(), thresholds.MaxFalsePositive),
		)
	}
	if thresholds.MaxWeakMatch > 0 && report.WeakMatchRate() > thresholds.MaxWeakMatch {
		report.ThresholdViolations = append(report.ThresholdViolations,
			fmt.Sprintf("weak-match rate %.3f exceeds threshold %.3f", report.WeakMatchRate(), thresholds.MaxWeakMatch),
		)
	}
	if report.Top1Rate() < baseline.Top1Rate {
		report.ThresholdViolations = append(report.ThresholdViolations,
			fmt.Sprintf("top-1 rate regressed from %.3f to %.3f", baseline.Top1Rate, report.Top1Rate()),
		)
	}
	if report.Top3Rate() < baseline.Top3Rate {
		report.ThresholdViolations = append(report.ThresholdViolations,
			fmt.Sprintf("top-3 rate regressed from %.3f to %.3f", baseline.Top3Rate, report.Top3Rate()),
		)
	}
	if report.UnmatchedRate() > baseline.UnmatchedRate {
		report.ThresholdViolations = append(report.ThresholdViolations,
			fmt.Sprintf("unmatched rate regressed from %.3f to %.3f", baseline.UnmatchedRate, report.UnmatchedRate()),
		)
	}
	if report.FalsePositiveRate() > baseline.FalsePositiveRate {
		report.ThresholdViolations = append(report.ThresholdViolations,
			fmt.Sprintf("false-positive rate regressed from %.3f to %.3f", baseline.FalsePositiveRate, report.FalsePositiveRate()),
		)
	}
	if report.WeakMatchRate() > baseline.WeakMatchRate {
		report.ThresholdViolations = append(report.ThresholdViolations,
			fmt.Sprintf("weak-match rate regressed from %.3f to %.3f", baseline.WeakMatchRate, report.WeakMatchRate()),
		)
	}
	if baseline.Fingerprint != "" && report.BaselineFingerprint() != baseline.Fingerprint {
		report.ThresholdViolations = append(report.ThresholdViolations,
			fmt.Sprintf("corpus fingerprint changed: baseline=%s current=%s", baseline.Fingerprint, report.BaselineFingerprint()),
		)
	}
	if len(report.ThresholdViolations) > 0 {
		return errors.New(strings.Join(report.ThresholdViolations, "; "))
	}
	return nil
}

func incrementBucket(target map[string]int, key string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	target[key]++
}

func appendCountFingerprint(parts []string, prefix string, counts map[string]int) []string {
	for _, key := range sortedKeys(counts) {
		parts = append(parts, fmt.Sprintf("%s:%s:%d", prefix, key, counts[key]))
	}
	return parts
}

func validateFixtureMetadata(loaded []Fixture) []string {
	issues := make([]string, 0)
	seenIDs := make(map[string]string, len(loaded))
	seenFingerprints := make(map[string]string, len(loaded))

	for _, fixture := range loaded {
		if prev, ok := seenIDs[fixture.ID]; ok {
			issues = append(issues, fmt.Sprintf("duplicate fixture id %s also seen in %s", fixture.ID, prev))
		} else {
			seenIDs[fixture.ID] = fixture.FilePath
		}

		fingerprint := strings.TrimSpace(fixture.Fingerprint)
		if fingerprint == "" {
			issues = append(issues, fmt.Sprintf("fixture %s is missing fingerprint", fixture.ID))
		} else if prev, ok := seenFingerprints[fingerprint]; ok {
			issues = append(issues, fmt.Sprintf("duplicate fingerprint %s shared by %s and %s", fingerprint, prev, fixture.ID))
		} else {
			seenFingerprints[fingerprint] = fixture.ID
		}

		switch fixture.effectiveClass() {
		case ClassReal:
			if strings.TrimSpace(fixture.Title) == "" {
				issues = append(issues, fmt.Sprintf("real fixture %s is missing title", fixture.ID))
			}
			if strings.TrimSpace(fixture.Source.Adapter) == "" {
				issues = append(issues, fmt.Sprintf("real fixture %s is missing source.adapter", fixture.ID))
			}
			if strings.TrimSpace(fixture.Source.Provider) == "" {
				issues = append(issues, fmt.Sprintf("real fixture %s is missing source.provider", fixture.ID))
			}
			if strings.TrimSpace(fixture.Source.URL) == "" {
				issues = append(issues, fmt.Sprintf("real fixture %s is missing source.url", fixture.ID))
			}
			if strings.TrimSpace(fixture.Review.Status) != "promoted" {
				issues = append(issues, fmt.Sprintf("real fixture %s review.status must be promoted", fixture.ID))
			}
			if strings.TrimSpace(fixture.Review.PromotedAt) == "" {
				issues = append(issues, fmt.Sprintf("real fixture %s is missing review.promoted_at", fixture.ID))
			}
		case ClassStaging:
			if strings.TrimSpace(fixture.Source.Adapter) == "" {
				issues = append(issues, fmt.Sprintf("staging fixture %s is missing source.adapter", fixture.ID))
			}
			if strings.TrimSpace(fixture.Source.Provider) == "" {
				issues = append(issues, fmt.Sprintf("staging fixture %s is missing source.provider", fixture.ID))
			}
			if strings.TrimSpace(fixture.Source.URL) == "" {
				issues = append(issues, fmt.Sprintf("staging fixture %s is missing source.url", fixture.ID))
			}
		}
	}

	sort.Strings(issues)
	return issues
}

func ratio(value, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(value) / float64(total)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
