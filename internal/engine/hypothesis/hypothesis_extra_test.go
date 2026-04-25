package hypothesis

import (
	"testing"

	"faultline/internal/model"
)

// makeLines creates Line slices for testing.
func makeLines(originals ...string) []model.Line {
	out := make([]model.Line, len(originals))
	for i, s := range originals {
		out[i] = model.Line{
			Original:   s,
			Normalized: normalize(s),
		}
	}
	return out
}

// --- environment.evaluate for prefixed signal forms ---

func newTestEnv(lines []model.Line, stage string, deltaSignals map[string]model.DeltaSignal) *environment {
	return newEnvironment(lines, model.Context{Stage: stage}, &model.Delta{Signals: func() []model.DeltaSignal {
		var out []model.DeltaSignal
		for _, v := range deltaSignals {
			out = append(out, v)
		}
		return out
	}()})
}

func TestEvaluateLogContainsMatches(t *testing.T) {
	env := newTestEnv(makeLines("npm ERR! missing module"), "", nil)
	result := env.evaluate("log.contains:npm ERR")
	if !result.matched {
		t.Fatal("expected log.contains to match")
	}
	if len(result.evidence) == 0 {
		t.Error("expected evidence for log.contains match")
	}
}

func TestEvaluateLogContainsNoMatch(t *testing.T) {
	env := newTestEnv(makeLines("everything is fine"), "", nil)
	result := env.evaluate("log.contains:fatal error")
	if result.matched {
		t.Fatal("expected log.contains to not match")
	}
}

func TestEvaluateLogAbsentMatchesWhenAbsent(t *testing.T) {
	env := newTestEnv(makeLines("all good"), "", nil)
	result := env.evaluate("log.absent:fatal")
	if !result.matched {
		t.Fatal("expected log.absent to match when pattern absent")
	}
}

func TestEvaluateLogAbsentNoMatchWhenPresent(t *testing.T) {
	env := newTestEnv(makeLines("fatal error occurred"), "", nil)
	result := env.evaluate("log.absent:fatal")
	if result.matched {
		t.Fatal("expected log.absent to not match when pattern present")
	}
}

func TestEvaluateDeltaSignalMatches(t *testing.T) {
	env := newEnvironment(nil, model.Context{}, &model.Delta{
		Signals: []model.DeltaSignal{{ID: "dep.changed", Detail: "package.json changed"}},
	})
	result := env.evaluate("delta.signal:dep.changed")
	if !result.matched {
		t.Fatal("expected delta.signal to match")
	}
	if len(result.evidence) == 0 {
		t.Error("expected evidence from delta signal detail")
	}
}

func TestEvaluateDeltaSignalNoMatch(t *testing.T) {
	env := newEnvironment(nil, model.Context{}, nil)
	result := env.evaluate("delta.signal:nonexistent.signal")
	if result.matched {
		t.Fatal("expected delta.signal to not match for absent signal")
	}
}

func TestEvaluateDeltaAbsentMatchesWhenAbsent(t *testing.T) {
	env := newEnvironment(nil, model.Context{}, nil)
	result := env.evaluate("delta.absent:some.signal")
	if !result.matched {
		t.Fatal("expected delta.absent to match when signal is absent")
	}
}

func TestEvaluateDeltaAbsentNoMatchWhenPresent(t *testing.T) {
	env := newEnvironment(nil, model.Context{}, &model.Delta{
		Signals: []model.DeltaSignal{{ID: "some.signal"}},
	})
	result := env.evaluate("delta.absent:some.signal")
	if result.matched {
		t.Fatal("expected delta.absent to not match when signal is present")
	}
}

func TestEvaluateContextStageMatches(t *testing.T) {
	env := newEnvironment(nil, model.Context{Stage: "build"}, nil)
	result := env.evaluate("context.stage:build")
	if !result.matched {
		t.Fatal("expected context.stage to match")
	}
}

func TestEvaluateContextStageCaseInsensitive(t *testing.T) {
	env := newEnvironment(nil, model.Context{Stage: "BUILD"}, nil)
	result := env.evaluate("context.stage:build")
	if !result.matched {
		t.Fatal("expected context.stage to match case-insensitively")
	}
}

func TestEvaluateContextStageNoMatch(t *testing.T) {
	env := newEnvironment(nil, model.Context{Stage: "test"}, nil)
	result := env.evaluate("context.stage:build")
	if result.matched {
		t.Fatal("expected context.stage not to match different stage")
	}
}

func TestEvaluateContextStageAbsentMatchesWhenDifferent(t *testing.T) {
	env := newEnvironment(nil, model.Context{Stage: "test"}, nil)
	result := env.evaluate("context.stage.absent:build")
	if !result.matched {
		t.Fatal("expected context.stage.absent to match when stage differs")
	}
}

func TestEvaluateContextStageAbsentNoMatchWhenSame(t *testing.T) {
	env := newEnvironment(nil, model.Context{Stage: "build"}, nil)
	result := env.evaluate("context.stage.absent:build")
	if result.matched {
		t.Fatal("expected context.stage.absent to not match when stage matches")
	}
}

func TestEvaluateEmptySignalReturnsEmpty(t *testing.T) {
	env := newEnvironment(nil, model.Context{}, nil)
	result := env.evaluate("")
	if result.matched {
		t.Error("expected empty signal to not match")
	}
}

func TestEvaluateCachesResult(t *testing.T) {
	env := newTestEnv(makeLines("npm ERR! missing"), "", nil)
	r1 := env.evaluate("log.contains:npm ERR")
	r2 := env.evaluate("log.contains:npm ERR")
	if r1.matched != r2.matched {
		t.Error("expected cached result to match first result")
	}
}

// --- newEnvironment with delta signals ---

func TestNewEnvironmentWithDelta(t *testing.T) {
	delta := &model.Delta{
		Signals: []model.DeltaSignal{
			{ID: "delta.test.failure.introduced", Detail: "TestFoo failed"},
		},
	}
	env := newEnvironment(nil, model.Context{}, delta)
	if _, ok := env.deltaSignals["delta.test.failure.introduced"]; !ok {
		t.Fatal("expected delta signal in environment")
	}
}

func TestNewEnvironmentWithNilDelta(t *testing.T) {
	env := newEnvironment(nil, model.Context{}, nil)
	if len(env.deltaSignals) != 0 {
		t.Error("expected no delta signals for nil delta")
	}
}

// --- assess with discriminators and excludes ---

func TestAssessWithDiscriminator(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{
			ID: "dep-conflict",
			Hypothesis: model.HypothesisSpec{
				Discriminators: []model.HypothesisDiscriminator{
					{Signal: "log.contains:dependency conflict", Weight: 0.6},
				},
			},
		},
		Score: 1.0,
	}
	env := newTestEnv(makeLines("dependency conflict detected"), "", nil)
	assessment := assess(result, env)
	if len(assessment.Discriminators) == 0 {
		t.Fatal("expected discriminator match")
	}
	if assessment.FinalScore <= assessment.BaseScore {
		t.Error("expected final score higher than base with discriminator")
	}
}

func TestAssessWithDiscriminatorCustomDescription(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{
			ID: "dep-conflict",
			Hypothesis: model.HypothesisSpec{
				Discriminators: []model.HypothesisDiscriminator{
					{Signal: "log.contains:dependency conflict", Description: "custom description", Weight: 0.5},
				},
			},
		},
		Score: 1.0,
	}
	env := newTestEnv(makeLines("dependency conflict detected"), "", nil)
	assessment := assess(result, env)
	if len(assessment.Discriminators) == 0 {
		t.Fatal("expected discriminator match")
	}
	if assessment.Discriminators[0].Description != "custom description" {
		t.Errorf("expected custom description, got %q", assessment.Discriminators[0].Description)
	}
}

func TestAssessWithExcludeRulesOut(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{
			ID: "dep-conflict",
			Hypothesis: model.HypothesisSpec{
				Excludes: []model.HypothesisSignal{
					{Signal: "log.contains:dependency conflict"},
				},
			},
		},
		Score: 1.0,
	}
	env := newTestEnv(makeLines("dependency conflict detected"), "", nil)
	assessment := assess(result, env)
	if !assessment.Eliminated {
		t.Fatal("expected hypothesis to be eliminated by exclude signal")
	}
	if len(assessment.Excludes) == 0 {
		t.Error("expected excludes entry in assessment")
	}
}

func TestAssessWithContradictionDefaultWeight(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{
			ID: "dep-conflict",
			Hypothesis: model.HypothesisSpec{
				Contradicts: []model.HypothesisSignal{
					{Signal: "log.contains:dependency conflict", Weight: 0}, // use default
				},
			},
		},
		Score: 1.5,
	}
	env := newTestEnv(makeLines("dependency conflict detected"), "", nil)
	assessment := assess(result, env)
	if len(assessment.Contradicts) == 0 {
		t.Fatal("expected contradiction match")
	}
	if assessment.FinalScore >= assessment.BaseScore {
		t.Error("expected final score reduced by contradiction")
	}
}

func TestAssessWithSupportDefaultWeight(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{
			ID: "dep-conflict",
			Hypothesis: model.HypothesisSpec{
				Supports: []model.HypothesisSignal{
					{Signal: "log.contains:dependency conflict", Weight: 0}, // use default
				},
			},
		},
		Score: 1.0,
	}
	env := newTestEnv(makeLines("dependency conflict detected"), "", nil)
	assessment := assess(result, env)
	if len(assessment.Supports) == 0 {
		t.Fatal("expected support match")
	}
	if assessment.FinalScore <= assessment.BaseScore {
		t.Error("expected final score increased by support")
	}
}

// --- buildDifferential ---

func TestBuildDifferentialReturnsNilForEmpty(t *testing.T) {
	if got := buildDifferential(nil, 3); got != nil {
		t.Errorf("expected nil for empty results, got %v", got)
	}
}

func TestBuildDifferentialAllEliminatedUsesFirst(t *testing.T) {
	results := []model.Result{
		{
			Playbook: model.Playbook{ID: "eliminated"},
			Hypothesis: &model.HypothesisAssessment{
				Eliminated: true,
				FinalScore: 0.5,
			},
		},
	}
	diff := buildDifferential(results, 3)
	if diff == nil {
		t.Fatal("expected non-nil diff")
	}
	if diff.Likely == nil {
		t.Fatal("expected likely result even when all eliminated")
	}
	if diff.Likely.FailureID != "eliminated" {
		t.Errorf("expected eliminated candidate as likely fallback, got %q", diff.Likely.FailureID)
	}
}

func TestBuildDifferentialLimitsAlternatives(t *testing.T) {
	results := make([]model.Result, 5)
	for i := range results {
		results[i] = model.Result{
			Playbook: model.Playbook{ID: "result", Category: "cat"},
			Hypothesis: &model.HypothesisAssessment{
				FinalScore: float64(5 - i),
			},
			Evidence: []string{"evidence unique " + string(rune('a'+i))},
		}
		results[i].Playbook.ID = "result-" + string(rune('a'+i))
	}
	diff := buildDifferential(results, 2)
	if diff == nil {
		t.Fatal("expected non-nil diff")
	}
	if len(diff.Alternatives) > 1 {
		t.Errorf("expected at most 1 alternative (limit-1), got %d", len(diff.Alternatives))
	}
}

// --- summarizeMatches dedup ---

func TestSummarizeMatchesDedupesDescriptions(t *testing.T) {
	items := []model.HypothesisMatch{
		{Description: "same description"},
		{Description: "same description"},
		{Description: "different description"},
	}
	got := summarizeMatches(items, nil, 5)
	if len(got) != 2 {
		t.Errorf("expected 2 unique descriptions, got %v", got)
	}
}

func TestSummarizeMatchesFiltersEmpty(t *testing.T) {
	items := []model.HypothesisMatch{
		{Description: ""},
		{Description: "  "},
		{Description: "valid"},
	}
	got := summarizeMatches(items, nil, 5)
	if len(got) != 1 || got[0] != "valid" {
		t.Errorf("expected [valid], got %v", got)
	}
}

func TestSummarizeMatchesPullsFromSecondaryWhenPrimaryShort(t *testing.T) {
	primary := []model.HypothesisMatch{
		{Description: "from primary"},
	}
	secondary := []model.HypothesisMatch{
		{Description: "from secondary"},
	}
	got := summarizeMatches(primary, secondary, 3)
	if len(got) != 2 {
		t.Errorf("expected 2 items from primary+secondary, got %v", got)
	}
}

// --- disproofChecks ---

func TestDisproofChecksCollectsFromExcludes(t *testing.T) {
	spec := model.HypothesisSpec{
		Excludes: []model.HypothesisSignal{
			{Signal: "log.contains:dependency conflict"},
		},
	}
	got := disproofChecks(spec, 3)
	if len(got) == 0 {
		t.Fatal("expected disproof checks from excludes")
	}
}

func TestDisproofChecksCollectsFromContradicts(t *testing.T) {
	spec := model.HypothesisSpec{
		Contradicts: []model.HypothesisSignal{
			{Signal: "cache.restore.detected"},
		},
	}
	got := disproofChecks(spec, 3)
	if len(got) == 0 {
		t.Fatal("expected disproof checks from contradicts")
	}
}

func TestDisproofChecksDedupesDescriptions(t *testing.T) {
	spec := model.HypothesisSpec{
		Excludes: []model.HypothesisSignal{
			{Signal: "cache.restore.detected"},
			{Signal: "cache.restore.detected"},
		},
	}
	got := disproofChecks(spec, 5)
	if len(got) != 1 {
		t.Errorf("expected 1 unique check, got %v", got)
	}
}

// --- hypothesisScore ---

func TestHypothesisScoreNilFallback(t *testing.T) {
	got := hypothesisScore(nil, 2.5)
	if got != 2.5 {
		t.Errorf("expected fallback 2.5, got %v", got)
	}
}

func TestHypothesisScoreUsesAssessment(t *testing.T) {
	assessment := &model.HypothesisAssessment{FinalScore: 3.75}
	got := hypothesisScore(assessment, 1.0)
	if got != 3.75 {
		t.Errorf("expected 3.75, got %v", got)
	}
}

// --- confidenceText ---

func TestConfidenceTextHigh(t *testing.T) {
	if got := confidenceText(0.9); got != "High" {
		t.Errorf("expected High, got %q", got)
	}
}

func TestConfidenceTextMedium(t *testing.T) {
	if got := confidenceText(0.6); got != "Medium" {
		t.Errorf("expected Medium, got %q", got)
	}
}

func TestConfidenceTextLow(t *testing.T) {
	if got := confidenceText(0.3); got != "Low" {
		t.Errorf("expected Low, got %q", got)
	}
}

func TestConfidenceTextBoundary(t *testing.T) {
	if got := confidenceText(0.55); got != "Medium" {
		t.Errorf("expected Medium at 0.55, got %q", got)
	}
	if got := confidenceText(0.8); got != "High" {
		t.Errorf("expected High at 0.8, got %q", got)
	}
}

// --- Build with delta signals ---

func TestBuildWithDeltaSignals(t *testing.T) {
	results := []model.Result{
		{
			Playbook: model.Playbook{
				ID: "test-failure",
				Hypothesis: model.HypothesisSpec{
					Supports: []model.HypothesisSignal{
						{Signal: "delta.signal:delta.test.failure.introduced"},
					},
				},
			},
			Score: 1.0,
		},
	}
	delta := &model.Delta{
		Signals: []model.DeltaSignal{
			{ID: "delta.test.failure.introduced", Detail: "TestAuth newly failing"},
		},
	}
	got, _ := Build(Inputs{
		Results: results,
		Delta:   delta,
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	if got[0].Hypothesis == nil {
		t.Fatal("expected hypothesis assessment")
	}
	if len(got[0].Hypothesis.Supports) == 0 {
		t.Error("expected delta signal to match as support")
	}
}

func TestBuildWithContextStage(t *testing.T) {
	results := []model.Result{
		{
			Playbook: model.Playbook{
				ID: "build-failure",
				Hypothesis: model.HypothesisSpec{
					Supports: []model.HypothesisSignal{
						{Signal: "context.stage:build"},
					},
				},
			},
			Score: 1.0,
		},
	}
	got, _ := Build(Inputs{
		Results: results,
		Context: model.Context{Stage: "build"},
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	if got[0].Hypothesis == nil {
		t.Fatal("expected hypothesis assessment")
	}
	if len(got[0].Hypothesis.Supports) == 0 {
		t.Error("expected context.stage signal to match")
	}
}

// --- rivalryKey ---

func TestRivalryKeyFallsBackToPlaybookID(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{
			ID:       "test-id",
			Category: "cat",
		},
	}
	key := rivalryKey(result)
	if key == "" {
		t.Fatal("expected non-empty rivalry key")
	}
}

func TestRivalryKeyUsesEvidenceWhenNoHypothesis(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{
			Category: "cat",
		},
		Evidence: []string{"some evidence line"},
	}
	key := rivalryKey(result)
	if key == "" {
		t.Fatal("expected non-empty rivalry key with evidence")
	}
}
