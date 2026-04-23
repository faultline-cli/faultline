package engine

import (
	"path/filepath"
	"testing"

	"faultline/internal/detectors"
	"faultline/internal/model"
)

func TestBundledSourcePlaybookFixtures(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	tests := []struct {
		name   string
		dir    string
		wantID string
	}{
		{
			name:   "missing error propagation",
			dir:    filepath.Join("testdata", "source", "missing-error-propagation-positive"),
			wantID: "missing-error-propagation",
		},
		{
			name:   "panic in http handler",
			dir:    filepath.Join("testdata", "source", "panic-in-http-handler-positive"),
			wantID: "panic-in-http-handler",
		},
		{
			name:   "unawaited promise",
			dir:    filepath.Join("testdata", "source", "unawaited-promise-positive"),
			wantID: "unawaited-promise",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			first := analyzeSourceFixture(t, e, tc.dir)
			if got := first.Results[0].Playbook.ID; got != tc.wantID {
				t.Fatalf("expected top source playbook %s, got %s", tc.wantID, got)
			}
			if first.Results[0].Detector != string(detectors.KindSource) {
				t.Fatalf("expected source detector result, got %s", first.Results[0].Detector)
			}
			if len(first.Results[0].Evidence) == 0 {
				t.Fatalf("expected evidence for %s", tc.wantID)
			}

			second := analyzeSourceFixture(t, e, tc.dir)
			assertDeterministicSourceResults(t, first.Results, second.Results)
		})
	}
}

func TestBundledSourcePlaybookMitigationsLowerScore(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	tests := []struct {
		name      string
		playbook  string
		unsafeDir string
		safeDir   string
	}{
		{
			name:      "missing error propagation",
			playbook:  "missing-error-propagation",
			unsafeDir: filepath.Join("testdata", "source", "missing-error-propagation-positive"),
			safeDir:   filepath.Join("testdata", "source", "missing-error-propagation-safe"),
		},
		{
			name:      "panic in http handler",
			playbook:  "panic-in-http-handler",
			unsafeDir: filepath.Join("testdata", "source", "panic-in-http-handler-positive"),
			safeDir:   filepath.Join("testdata", "source", "panic-in-http-handler-safe"),
		},
		{
			name:      "unawaited promise",
			playbook:  "unawaited-promise",
			unsafeDir: filepath.Join("testdata", "source", "unawaited-promise-positive"),
			safeDir:   filepath.Join("testdata", "source", "unawaited-promise-safe"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			unsafeResult := requireSourcePlaybookResult(t, analyzeSourceFixture(t, e, tc.unsafeDir).Results, tc.playbook)
			safeResult := requireSourcePlaybookResult(t, analyzeSourceFixture(t, e, tc.safeDir).Results, tc.playbook)
			if safeResult.Score >= unsafeResult.Score {
				t.Fatalf(
					"expected mitigated repository score %.2f to be lower than unsafe score %.2f for %s",
					safeResult.Score,
					unsafeResult.Score,
					tc.playbook,
				)
			}
		})
	}
}

func analyzeSourceFixture(t *testing.T, e *Engine, dir string) *model.Analysis {
	t.Helper()

	analysis, err := e.AnalyzeRepository(dir, detectors.ChangeSet{})
	if err != nil {
		t.Fatalf("analyze source fixture %s: %v", dir, err)
	}
	if analysis == nil || len(analysis.Results) == 0 {
		t.Fatalf("expected source fixture %s to produce matches", dir)
	}
	return analysis
}

func requireSourcePlaybookResult(t *testing.T, results []model.Result, id string) model.Result {
	t.Helper()

	for _, result := range results {
		if result.Playbook.ID == id {
			return result
		}
	}
	t.Fatalf("expected source result %s in %v", id, resultIDs(results))
	return model.Result{}
}

func assertDeterministicSourceResults(t *testing.T, first, second []model.Result) {
	t.Helper()

	if len(first) != len(second) {
		t.Fatalf("expected deterministic source result count, got %d and %d", len(first), len(second))
	}
	for i := range first {
		if first[i].Playbook.ID != second[i].Playbook.ID ||
			first[i].Score != second[i].Score ||
			first[i].Confidence != second[i].Confidence {
			t.Fatalf("expected deterministic source ranking, got %v and %v", resultIDs(first), resultIDs(second))
		}
	}
}
