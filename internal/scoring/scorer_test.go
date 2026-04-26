package scoring

import (
	"testing"

	"faultline/internal/model"
)

func TestScoreAddsRankingAndDelta(t *testing.T) {
	results, delta, err := Score(Inputs{
		Context: model.Context{Stage: "build", CommandHint: "npm ci"},
		Lines: []model.Line{
			{Original: "npm ERR! ERESOLVE unable to resolve dependency tree", Normalized: normalizeText("npm ERR! ERESOLVE unable to resolve dependency tree")},
		},
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:         "npm-peer-dependency-conflict",
					Title:      "npm peer dependency conflict",
					Category:   "build",
					StageHints: []string{"build"},
					Tags:       []string{"npm", "dependency"},
					Match: model.MatchSpec{
						Any: []string{"ERESOLVE unable to resolve dependency tree"},
					},
					Workflow: model.WorkflowSpec{
						LikelyFiles: []string{"package.json", "package-lock.json"},
					},
				},
				Detector:   "log",
				Score:      3.2,
				Confidence: 0.81,
				Evidence:   []string{"npm ERR! ERESOLVE unable to resolve dependency tree"},
				EvidenceBy: model.EvidenceBundle{Triggers: []model.Evidence{{Detail: "npm ERR! ERESOLVE unable to resolve dependency tree"}}},
			},
		},
		RepoState: &RepoState{
			ChangedFiles: []string{"package.json", "package-lock.json"},
			RecentFiles:  []string{"package.json", "package-lock.json"},
		},
	})
	if err != nil {
		t.Fatalf("Score: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Ranking == nil {
		t.Fatal("expected ranking payload")
	}
	if results[0].Ranking.Mode != ModeBayes {
		t.Fatalf("expected ranking mode %q, got %q", ModeBayes, results[0].Ranking.Mode)
	}
	if delta == nil || len(delta.Causes) == 0 {
		t.Fatalf("expected delta causes, got %#v", delta)
	}
}

func TestScoreTieBreaksByPlaybookIDWhenSignalsMatch(t *testing.T) {
	results, _, err := Score(Inputs{
		Results: []model.Result{
			{
				Playbook:   model.Playbook{ID: "b", Title: "B", Category: "build"},
				Detector:   "log",
				Score:      1,
				Confidence: 0.5,
			},
			{
				Playbook:   model.Playbook{ID: "a", Title: "A", Category: "build"},
				Detector:   "log",
				Score:      1,
				Confidence: 0.5,
			},
		},
	})
	if err != nil {
		t.Fatalf("Score: %v", err)
	}
	if results[0].Playbook.ID != "a" {
		t.Fatalf("expected playbook a first, got %s", results[0].Playbook.ID)
	}
}

func TestBuildDeltaClassifiesChangedFiles(t *testing.T) {
	delta := buildDelta(&RepoState{
		Provider:          "github-actions",
		ChangedFiles:      []string{".github/workflows/ci.yml", "package.json", "Dockerfile"},
		TestsNewlyFailing: []string{"TestLockfile"},
		ErrorsAdded:       []string{"npm ERR! package-lock.json is not in sync"},
	})
	if delta == nil || len(delta.Causes) == 0 {
		t.Fatalf("expected delta output, got %#v", delta)
	}
	if delta.Causes[0].Kind == "" {
		t.Fatalf("expected populated delta cause, got %#v", delta.Causes[0])
	}
	if delta.Provider != "github-actions" {
		t.Fatalf("expected provider to be preserved, got %#v", delta)
	}
	if len(delta.Signals) == 0 {
		t.Fatalf("expected delta signals, got %#v", delta)
	}
	if len(delta.TestsNewlyFailing) != 1 || delta.TestsNewlyFailing[0] != "TestLockfile" {
		t.Fatalf("expected structured failing tests, got %#v", delta)
	}
}

func TestRankingContributionOrderingPrefersLargestAbsoluteContribution(t *testing.T) {
	ranking := rankingFromFeatures("bayes_v1", 2.0, 0.1, []feature{
		{Name: "small", Value: 0.1, Weight: 0.5, Reason: "small reason"},
		{Name: "large", Value: 1.0, Weight: 0.9, Reason: "large reason"},
	})
	if len(ranking.Contributions) < 2 {
		t.Fatalf("expected contributions, got %#v", ranking)
	}
	if ranking.Contributions[0].Feature != "large" {
		t.Fatalf("expected largest contribution first, got %#v", ranking.Contributions)
	}
}

func TestBaselineCandidateSeparationOnlyRewardsUniqueLeader(t *testing.T) {
	baseline := []model.Result{
		{Playbook: model.Playbook{ID: "top"}, Score: 4},
		{Playbook: model.Playbook{ID: "runner-up"}, Score: 3},
		{Playbook: model.Playbook{ID: "third"}, Score: 2},
	}

	if got := baselineCandidateSeparation(baseline, 0); got <= 0 {
		t.Fatalf("expected positive separation for unique leader, got %.2f", got)
	}
	if got := baselineCandidateSeparation(baseline, 1); got != 0 {
		t.Fatalf("expected no separation bonus for runner-up, got %.2f", got)
	}
	if got := baselineCandidateSeparation([]model.Result{
		{Playbook: model.Playbook{ID: "a"}, Score: 2},
		{Playbook: model.Playbook{ID: "b"}, Score: 2},
	}, 0); got != 0 {
		t.Fatalf("expected tied leader to get no separation bonus, got %.2f", got)
	}
}

func TestScoreDeltaBoostImprovesSpecificPlaybookRanking(t *testing.T) {
	results := []model.Result{
		{
			Playbook: model.Playbook{
				ID:            "dependency-drift",
				Title:         "Dependency drift",
				Category:      "build",
				RequiresDelta: true,
				DeltaBoost: []model.DeltaBoost{
					{Signal: "delta.dependency.changed", Weight: 0.4},
				},
			},
			Detector:   "log",
			Score:      2.2,
			Confidence: 0.70,
		},
		{
			Playbook: model.Playbook{
				ID:            "npm-ci-lockfile",
				Title:         "npm ci lockfile mismatch",
				Category:      "build",
				RequiresDelta: true,
				DeltaBoost: []model.DeltaBoost{
					{Signal: "delta.dependency.changed", Weight: 1.2},
					{Signal: "delta.scope.changed", Weight: 0.8},
				},
				Workflow: model.WorkflowSpec{
					LikelyFiles: []string{"package.json", "package-lock.json"},
				},
			},
			Detector:   "log",
			Score:      2.0,
			Confidence: 0.72,
		},
	}

	withoutDelta, _, err := Score(Inputs{
		Results:        results,
		DeltaRequested: false,
	})
	if err != nil {
		t.Fatalf("Score without delta: %v", err)
	}
	if withoutDelta[0].Playbook.ID != "dependency-drift" {
		t.Fatalf("expected baseline leader to remain first without delta, got %s", withoutDelta[0].Playbook.ID)
	}

	withDelta, delta, err := Score(Inputs{
		Results: results,
		RepoState: &RepoState{
			Provider:     "github-actions",
			ChangedFiles: []string{"package.json", "package-lock.json"},
		},
		DeltaRequested: true,
	})
	if err != nil {
		t.Fatalf("Score with delta: %v", err)
	}
	if delta == nil {
		t.Fatal("expected structured delta")
	}
	if withDelta[0].Playbook.ID != "npm-ci-lockfile" {
		t.Fatalf("expected delta-aware ranking to prefer npm-ci-lockfile, got %s", withDelta[0].Playbook.ID)
	}
}

func TestDiagnoseDeltaReturnsNilForNilState(t *testing.T) {
	if got := DiagnoseDelta(nil); got != nil {
		t.Fatalf("expected nil delta for nil state, got %#v", got)
	}
}

func TestDiagnoseDeltaPopulatesCausesFromFiles(t *testing.T) {
	delta := DiagnoseDelta(&RepoState{
		ChangedFiles: []string{"package.json", ".github/workflows/ci.yml"},
	})
	if delta == nil {
		t.Fatal("expected non-nil delta")
	}
	if len(delta.Causes) == 0 {
		t.Fatalf("expected at least one cause, got %#v", delta)
	}
}

func TestDebugStringEmptyForNilRanking(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{ID: "some-playbook"},
	}
	if got := DebugString(result); got != "" {
		t.Fatalf("expected empty string for nil ranking, got %q", got)
	}
}

func TestDebugStringFormatsIDAndScore(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{ID: "my-playbook"},
		Ranking: &model.Ranking{
			FinalScore: 3.14159,
		},
	}
	got := DebugString(result)
	if got == "" {
		t.Fatal("expected non-empty debug string")
	}
	if got != "my-playbook 3.14" {
		t.Fatalf("unexpected debug string: %q", got)
	}
}

func TestCloneEnvDiffReturnsNilForEmpty(t *testing.T) {
	if got := cloneEnvDiff(nil); got != nil {
		t.Fatalf("expected nil for nil input, got %#v", got)
	}
	if got := cloneEnvDiff(map[string]model.DeltaEnvChange{}); got != nil {
		t.Fatalf("expected nil for empty input, got %#v", got)
	}
}

func TestCloneEnvDiffCopiesEntries(t *testing.T) {
	in := map[string]model.DeltaEnvChange{
		"NODE_VERSION": {Baseline: "14", Current: "18"},
		"  ":           {Baseline: "skip"},
	}
	out := cloneEnvDiff(in)
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if _, ok := out["NODE_VERSION"]; !ok {
		t.Fatal("expected NODE_VERSION entry to be preserved")
	}
	if _, ok := out["  "]; ok {
		t.Fatal("expected blank key to be dropped")
	}
}

// --- ratio ---

func TestRatioZeroDenominatorReturnsZero(t *testing.T) {
	if got := ratio(5, 0); got != 0 {
		t.Errorf("expected 0 for zero denominator, got %f", got)
	}
}

func TestRatioNegativeDenominatorReturnsZero(t *testing.T) {
	if got := ratio(3, -1); got != 0 {
		t.Errorf("expected 0 for negative denominator, got %f", got)
	}
}

func TestRatioNormalCase(t *testing.T) {
	if got := ratio(1, 2); got != 0.5 {
		t.Errorf("expected 0.5, got %f", got)
	}
}

func TestRatioClampsToOne(t *testing.T) {
	if got := ratio(10, 3); got != 1.0 {
		t.Errorf("expected 1.0 for num > denom, got %f", got)
	}
}

// --- classifyDeltaFile ---

func TestClassifyDeltaFileRuntimeToolchain(t *testing.T) {
	kind, reason, score := classifyDeltaFile(".nvmrc")
	if kind != "runtime_toolchain" {
		t.Errorf("expected runtime_toolchain for .nvmrc, got %q", kind)
	}
	if reason == "" {
		t.Error("expected non-empty reason")
	}
	if score == 0 {
		t.Error("expected non-zero score")
	}
}

func TestClassifyDeltaFileEnvironment(t *testing.T) {
	kind, _, score := classifyDeltaFile(".env")
	if kind != "environment" {
		t.Errorf("expected environment for .env, got %q", kind)
	}
	if score == 0 {
		t.Error("expected non-zero score for environment file")
	}
}

func TestClassifyDeltaFileTestFile(t *testing.T) {
	kind, reason, score := classifyDeltaFile("main_test.go")
	if kind != "test_data" {
		t.Errorf("expected test_data for main_test.go, got %q", kind)
	}
	if reason == "" {
		t.Error("expected non-empty reason")
	}
	if score == 0 {
		t.Error("expected non-zero score")
	}
}

func TestClassifyDeltaFileSourceCode(t *testing.T) {
	kind, reason, score := classifyDeltaFile("main.go")
	if kind != "source_code" {
		t.Errorf("expected source_code for main.go, got %q", kind)
	}
	if reason == "" {
		t.Error("expected non-empty reason")
	}
	if score == 0 {
		t.Error("expected non-zero score")
	}
}

func TestClassifyDeltaFileUnknownReturnsEmpty(t *testing.T) {
	kind, reason, score := classifyDeltaFile("somefile.xyz")
	if kind != "" || reason != "" || score != 0 {
		t.Errorf("expected empty classification for unknown file, got kind=%q reason=%q score=%f", kind, reason, score)
	}
}

// --- buildDelta edge cases ---

func TestBuildDeltaHotfixSignalPopulatesEnvironmentCause(t *testing.T) {
	delta := buildDelta(&RepoState{
		HotfixSignals: []string{"hotfix: revert broken deploy"},
	})
	if delta == nil {
		t.Fatal("expected non-nil delta for hotfix signals")
	}
	found := false
	for _, cause := range delta.Causes {
		if cause.Kind == "environment" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected environment cause from hotfix signal, got %#v", delta.Causes)
	}
}

func TestBuildDeltaDriftSignalPopulatesEnvironmentCause(t *testing.T) {
	delta := buildDelta(&RepoState{
		DriftSignals: []string{"config drift detected"},
	})
	if delta == nil {
		t.Fatal("expected non-nil delta for drift signals")
	}
	found := false
	for _, cause := range delta.Causes {
		if cause.Kind == "environment" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected environment cause from drift signal, got %#v", delta.Causes)
	}
}

func TestBuildDeltaReturnsNilForAllEmptyState(t *testing.T) {
	delta := buildDelta(&RepoState{})
	if delta != nil {
		t.Fatalf("expected nil delta for empty state, got %#v", delta)
	}
}

func TestBuildDeltaVersionDefaultsWhenWeightsHaveVersion(t *testing.T) {
	delta := buildDelta(&RepoState{
		ChangedFiles: []string{"go.mod"},
	})
	if delta == nil {
		t.Fatal("expected non-nil delta")
	}
	if delta.Version == "" {
		t.Error("expected non-empty version in delta")
	}
}

// --- buildDeltaSignals ---

func TestBuildDeltaSignalsIncludesErrorsAdded(t *testing.T) {
	signals := buildDeltaSignals(&RepoState{
		ErrorsAdded: []string{"error: module not found"},
	})
	found := false
	for _, s := range signals {
		if s.ID == "delta.error.new" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected delta.error.new signal for ErrorsAdded, got %#v", signals)
	}
}

func TestBuildDeltaSignalsNilStateReturnsNil(t *testing.T) {
	if got := buildDeltaSignals(nil); got != nil {
		t.Fatalf("expected nil for nil state, got %#v", got)
	}
}

// --- logMatchCoverage ---

func TestLogMatchCoverageSourceDetector(t *testing.T) {
	result := model.Result{
		Detector: "source",
		Playbook: model.Playbook{
			Source: model.SourceSpec{
				Triggers: []model.SignalMatcher{
					{ID: "missing-error-check"},
					{ID: "unhandled-error"},
				},
			},
		},
		EvidenceBy: model.EvidenceBundle{
			Triggers: []model.Evidence{
				{Detail: "missing-error-check at line 42"},
			},
		},
	}
	got := logMatchCoverage(Inputs{}, result)
	// 1 trigger evidence / 2 total triggers = 0.5
	if got != 0.5 {
		t.Errorf("expected 0.5 for source detector with 1/2 triggers, got %f", got)
	}
}

func TestLogMatchCoverageSourceDetectorNoTriggers(t *testing.T) {
	result := model.Result{
		Detector: "source",
		Playbook: model.Playbook{
			Source: model.SourceSpec{Triggers: nil},
		},
	}
	got := logMatchCoverage(Inputs{}, result)
	if got != 0 {
		t.Errorf("expected 0 for source detector with no triggers defined, got %f", got)
	}
}

func TestLogMatchCoverageUnknownDetectorReturnsZero(t *testing.T) {
	result := model.Result{Detector: "other"}
	got := logMatchCoverage(Inputs{}, result)
	if got != 0 {
		t.Errorf("expected 0 for unknown detector, got %f", got)
	}
}

// --- deltaPlaybookFeatures ---

func TestDeltaPlaybookFeaturesNotRequestedReturnsNil(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{ID: "some-playbook", RequiresDelta: true},
	}
	got := deltaPlaybookFeatures(Inputs{DeltaRequested: false}, result, nil)
	if got != nil {
		t.Fatalf("expected nil features when DeltaRequested=false, got %#v", got)
	}
}

// --- Bayes scoring regression guard ---
//
// This test pins the ordering and key structural properties of the Bayes
// scorer for a well-defined input. It will fail if the scoring logic regresses
// in a way that changes result ordering or removes ranking metadata.

func TestBayesScoringRegressionGuard(t *testing.T) {
	line := "npm ci can only install packages when your package.json and package-lock.json"
	results := []model.Result{
		{
			Playbook: model.Playbook{
				ID:         "npm-lockfile-mismatch",
				Title:      "npm lockfile out of sync",
				Category:   "build",
				StageHints: []string{"build"},
				Tags:       []string{"npm", "lockfile"},
				Match: model.MatchSpec{
					Any: []string{"npm ci can only install packages when your package.json and package-lock.json"},
				},
				Workflow: model.WorkflowSpec{
					LikelyFiles: []string{"package.json", "package-lock.json"},
				},
			},
			Detector:   "log",
			Score:      4.0,
			Confidence: 0.88,
			Evidence:   []string{line},
		},
		{
			Playbook: model.Playbook{
				ID:       "npm-peer-conflict",
				Title:    "npm peer dependency conflict",
				Category: "build",
				Tags:     []string{"npm"},
			},
			Detector:   "log",
			Score:      1.5,
			Confidence: 0.45,
		},
	}
	scored, delta, err := Score(Inputs{
		Context: model.Context{Stage: "build"},
		Lines: []model.Line{
			{Original: line, Normalized: normalizeText(line)},
		},
		Results: results,
		RepoState: &RepoState{
			ChangedFiles: []string{"package.json", "package-lock.json"},
		},
	})
	if err != nil {
		t.Fatalf("Score: %v", err)
	}
	// Top result must remain npm-lockfile-mismatch after Bayes scoring.
	if scored[0].Playbook.ID != "npm-lockfile-mismatch" {
		t.Fatalf("regression: expected npm-lockfile-mismatch to remain top, got %s", scored[0].Playbook.ID)
	}
	// Bayes must always attach a Ranking payload.
	if scored[0].Ranking == nil {
		t.Fatal("regression: expected Ranking payload on top result")
	}
	if scored[0].Ranking.Mode != ModeBayes {
		t.Fatalf("regression: expected mode %q, got %q", ModeBayes, scored[0].Ranking.Mode)
	}
	// Good-evidence candidate should never score lower than its baseline after Bayes.
	if scored[0].Ranking.FinalScore < scored[0].Ranking.BaselineScore {
		t.Fatalf("regression: FinalScore %.2f < BaselineScore %.2f for well-matched candidate", scored[0].Ranking.FinalScore, scored[0].Ranking.BaselineScore)
	}
	// Delta must be populated when changed files are present.
	if delta == nil {
		t.Fatal("regression: expected non-nil delta for changed repo state")
	}
	// Version field must be stable and non-empty.
	if scored[0].Ranking.Version == "" {
		t.Fatal("regression: Ranking.Version must be set")
	}
}
