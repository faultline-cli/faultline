package scoring

import (
	"testing"

	"faultline/internal/model"
)

// --- mitigationPresence ---

func TestMitigationPresenceNoNegativesReturnsZero(t *testing.T) {
	result := model.Result{
		EvidenceBy: model.EvidenceBundle{
			Triggers: []model.Evidence{{Detail: "some trigger"}},
		},
	}
	got := mitigationPresence(result)
	if got != 0 {
		t.Errorf("expected 0 with no mitigations, got %f", got)
	}
}

func TestMitigationPresenceWithMitigations(t *testing.T) {
	result := model.Result{
		EvidenceBy: model.EvidenceBundle{
			Triggers: []model.Evidence{
				{Detail: "trigger1"},
				{Detail: "trigger2"},
			},
			Mitigations: []model.Evidence{
				{Detail: "mitigation1"},
			},
		},
	}
	got := mitigationPresence(result)
	// 1 mitigation / 2 triggers = 0.5
	if got != 0.5 {
		t.Errorf("expected 0.5, got %f", got)
	}
}

func TestMitigationPresenceWithSuppressions(t *testing.T) {
	result := model.Result{
		EvidenceBy: model.EvidenceBundle{
			Suppressions: []model.Evidence{
				{Detail: "suppression1"},
				{Detail: "suppression2"},
			},
		},
	}
	got := mitigationPresence(result)
	// 2 suppressions / 1 (base=0 so clamped to 1) = 1.0 clamped
	if got != 1.0 {
		t.Errorf("expected 1.0 when suppressions exceed base, got %f", got)
	}
}

func TestMitigationPresenceMixedClampsToOne(t *testing.T) {
	result := model.Result{
		EvidenceBy: model.EvidenceBundle{
			Triggers:    []model.Evidence{{Detail: "t1"}},
			Mitigations: []model.Evidence{{Detail: "m1"}, {Detail: "m2"}, {Detail: "m3"}},
		},
	}
	got := mitigationPresence(result)
	if got > 1.0 {
		t.Errorf("expected clamped to 1.0, got %f", got)
	}
}

// --- matchesFileHints ---

func TestMatchesFileHintsExactMatch(t *testing.T) {
	if !matchesFileHints("Dockerfile", []string{"Dockerfile"}) {
		t.Error("expected exact match for Dockerfile")
	}
}

func TestMatchesFileHintsBasenameMatch(t *testing.T) {
	if !matchesFileHints("build/Dockerfile", []string{"Dockerfile"}) {
		t.Error("expected basename match for build/Dockerfile")
	}
}

func TestMatchesFileHintsSuffixMatch(t *testing.T) {
	if !matchesFileHints("config/app.yaml", []string{"app.yaml"}) {
		t.Error("expected suffix match for config/app.yaml")
	}
}

func TestMatchesFileHintsGlobMatch(t *testing.T) {
	if !matchesFileHints(".github/workflows/ci.yml", []string{"*.yml"}) {
		t.Error("expected glob match for .yml extension")
	}
}

func TestMatchesFileHintsContainsMatch(t *testing.T) {
	if !matchesFileHints("internal/config/settings.go", []string{"config"}) {
		t.Error("expected contains match for 'config'")
	}
}

func TestMatchesFileHintsNoMatch(t *testing.T) {
	if matchesFileHints("main.go", []string{"Dockerfile", "package.json"}) {
		t.Error("expected no match for main.go against Docker/npm hints")
	}
}

func TestMatchesFileHintsEmptyHints(t *testing.T) {
	if matchesFileHints("main.go", nil) {
		t.Error("expected false for empty hints")
	}
}

func TestMatchesFileHintsEmptyHint(t *testing.T) {
	if matchesFileHints("main.go", []string{"", "  "}) {
		t.Error("expected false for blank hints")
	}
}

// --- normalizeAgainst ---

func TestNormalizeAgainstZeroTopReturnsZero(t *testing.T) {
	got := normalizeAgainst(5, 0)
	if got != 0 {
		t.Errorf("expected 0, got %f", got)
	}
}

func TestNormalizeAgainstClampsToOne(t *testing.T) {
	got := normalizeAgainst(10, 5)
	if got != 1.0 {
		t.Errorf("expected 1.0 for value > top, got %f", got)
	}
}

func TestNormalizeAgainstHalf(t *testing.T) {
	got := normalizeAgainst(2.5, 5)
	if got != 0.5 {
		t.Errorf("expected 0.5, got %f", got)
	}
}

// --- candidateSeparation ---

func TestCandidateSeparationZeroScore(t *testing.T) {
	got := candidateSeparation(0, 0)
	if got != 0 {
		t.Errorf("expected 0 for zero score, got %f", got)
	}
}

func TestCandidateSeparationNextGeScore(t *testing.T) {
	got := candidateSeparation(3, 5)
	if got != 0 {
		t.Errorf("expected 0 when next >= score, got %f", got)
	}
}

func TestCandidateSeparationPositiveSeparation(t *testing.T) {
	got := candidateSeparation(10, 5)
	// (10-5)/10 = 0.5
	if got != 0.5 {
		t.Errorf("expected 0.5, got %f", got)
	}
}

// --- playbookLikelyClasses ---

func TestPlaybookLikelyClassesDependency(t *testing.T) {
	pb := model.Playbook{ID: "npm-ci-lockfile", Category: "build"}
	classes := playbookLikelyClasses(pb)
	found := false
	for _, c := range classes {
		if c == "dependency" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected dependency class, got %v", classes)
	}
}

func TestPlaybookLikelyClassesCICategory(t *testing.T) {
	pb := model.Playbook{ID: "github-actions-syntax", Category: "ci"}
	classes := playbookLikelyClasses(pb)
	found := false
	for _, c := range classes {
		if c == "ci_config" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected ci_config class, got %v", classes)
	}
}

func TestPlaybookLikelyClassesSourceDetector(t *testing.T) {
	pb := model.Playbook{ID: "missing-error-check", Detector: "source", Category: "build"}
	classes := playbookLikelyClasses(pb)
	found := false
	for _, c := range classes {
		if c == "source_code" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected source_code class, got %v", classes)
	}
}

// --- mitigationRefs ---

func TestMitigationRefsExtractsDetails(t *testing.T) {
	result := model.Result{
		EvidenceBy: model.EvidenceBundle{
			Mitigations: []model.Evidence{
				{Detail: "uses secure pattern"},
			},
			Suppressions: []model.Evidence{
				{Detail: "faultline:ignore test-id reason=expected"},
			},
		},
	}
	refs := mitigationRefs(result)
	if len(refs) != 2 {
		t.Errorf("expected 2 refs, got %v", refs)
	}
}

func TestMitigationRefsSkipsBlankDetails(t *testing.T) {
	result := model.Result{
		EvidenceBy: model.EvidenceBundle{
			Mitigations: []model.Evidence{
				{Detail: ""},
				{Detail: "  "},
				{Detail: "real detail"},
			},
		},
	}
	refs := mitigationRefs(result)
	if len(refs) != 1 {
		t.Errorf("expected 1 ref (blank details skipped), got %v", refs)
	}
}

// --- deltaSignalMap ---

func TestDeltaSignalMapNilDelta(t *testing.T) {
	out := deltaSignalMap(nil)
	if out != nil {
		t.Errorf("expected nil for nil delta, got %v", out)
	}
}

func TestDeltaSignalMapBuildsMap(t *testing.T) {
	delta := &model.Delta{
		Signals: []model.DeltaSignal{
			{ID: "dep_changed", Detail: "package.json changed"},
			{ID: "", Detail: "should be skipped"},
		},
	}
	out := deltaSignalMap(delta)
	if len(out) != 1 {
		t.Errorf("expected 1 signal (blank ID skipped), got %v", out)
	}
	if out["dep_changed"] != "package.json changed" {
		t.Errorf("unexpected signal detail: %v", out)
	}
}

// --- hasDeltaBoostSignal ---

func TestHasDeltaBoostSignalEmptySignals(t *testing.T) {
	boosts := []model.DeltaBoost{{Signal: "dep_changed"}}
	if hasDeltaBoostSignal(nil, boosts) {
		t.Error("expected false for empty signals map")
	}
}

func TestHasDeltaBoostSignalMatch(t *testing.T) {
	signals := map[string]string{"dep_changed": "package.json changed"}
	boosts := []model.DeltaBoost{{Signal: "dep_changed"}}
	if !hasDeltaBoostSignal(signals, boosts) {
		t.Error("expected true when boost signal is present")
	}
}

func TestHasDeltaBoostSignalNoMatch(t *testing.T) {
	signals := map[string]string{"other_signal": "something"}
	boosts := []model.DeltaBoost{{Signal: "dep_changed"}}
	if hasDeltaBoostSignal(signals, boosts) {
		t.Error("expected false when boost signal not present")
	}
}

// --- historicalFixtureSupport ---

func TestHistoricalFixtureSupportZeroCount(t *testing.T) {
	weights := weightsFile{PlaybookCounts: map[string]int{"other-id": 5}}
	pb := model.Playbook{ID: "git-auth"}
	got := historicalFixtureSupport(weights, pb)
	if got != 0 {
		t.Errorf("expected 0 for missing playbook, got %f", got)
	}
}

func TestHistoricalFixtureSupportNormalized(t *testing.T) {
	weights := weightsFile{PlaybookCounts: map[string]int{
		"git-auth":    10,
		"docker-auth": 5,
	}}
	pb := model.Playbook{ID: "git-auth"}
	got := historicalFixtureSupport(weights, pb)
	// 10/10 = 1.0
	if got != 1.0 {
		t.Errorf("expected 1.0 for max count playbook, got %f", got)
	}
}

func TestHistoricalFixtureSupportPartial(t *testing.T) {
	weights := weightsFile{PlaybookCounts: map[string]int{
		"git-auth":    5,
		"docker-auth": 10,
	}}
	pb := model.Playbook{ID: "git-auth"}
	got := historicalFixtureSupport(weights, pb)
	// 5/10 = 0.5
	if got != 0.5 {
		t.Errorf("expected 0.5, got %f", got)
	}
}
