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
		ChangedFiles: []string{".github/workflows/ci.yml", "package.json", "Dockerfile"},
	})
	if delta == nil || len(delta.Causes) == 0 {
		t.Fatalf("expected delta output, got %#v", delta)
	}
	if delta.Causes[0].Kind == "" {
		t.Fatalf("expected populated delta cause, got %#v", delta.Causes[0])
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
