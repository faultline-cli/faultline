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
