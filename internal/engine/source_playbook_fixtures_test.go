package engine

import (
	"path/filepath"
	"testing"

	"faultline/internal/detectors"
	"faultline/internal/model"
)

func TestBundledSourcePlaybookFixtures(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t)})

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
		{
			name:   "continue on error critical step",
			dir:    filepath.Join("testdata", "source", "continue-on-error-critical-step-positive"),
			wantID: "continue-on-error-critical-step",
		},
		{
			name:   "continue on error critical step noisy",
			dir:    filepath.Join("testdata", "source", "continue-on-error-critical-step-noisy"),
			wantID: "continue-on-error-critical-step",
		},
		{
			name:   "ignored shell exit in ci",
			dir:    filepath.Join("testdata", "source", "ignored-shell-exit-in-ci-positive"),
			wantID: "ignored-shell-exit-in-ci",
		},
		{
			name:   "ignored shell exit in ci noisy",
			dir:    filepath.Join("testdata", "source", "ignored-shell-exit-in-ci-noisy"),
			wantID: "ignored-shell-exit-in-ci",
		},
		{
			name:   "floating docker base image",
			dir:    filepath.Join("testdata", "source", "floating-docker-base-image-positive"),
			wantID: "floating-docker-base-image",
		},
		{
			name:   "shell dialect mismatch",
			dir:    filepath.Join("testdata", "source", "shell-dialect-mismatch-positive"),
			wantID: "shell-dialect-mismatch",
		},
		{
			name:   "unawaited promise db client",
			dir:    filepath.Join("testdata", "source", "unawaited-promise-db-positive"),
			wantID: "unawaited-promise",
		},
		{
			name:   "hardcoded secret",
			dir:    filepath.Join("testdata", "source", "hardcoded-secret-positive"),
			wantID: "hardcoded-secret",
		},
		{
			name:   "goroutine leak",
			dir:    filepath.Join("testdata", "source", "goroutine-leak-positive"),
			wantID: "goroutine-leak",
		},
		{
			name:   "insecure tls skip verify",
			dir:    filepath.Join("testdata", "source", "insecure-tls-skip-verify-positive"),
			wantID: "insecure-tls-skip-verify",
		},
		{
			name:   "missing transaction rollback",
			dir:    filepath.Join("testdata", "source", "missing-transaction-rollback-positive"),
			wantID: "missing-transaction-rollback",
		},
		{
			name:   "http client no timeout",
			dir:    filepath.Join("testdata", "source", "http-client-no-timeout-positive"),
			wantID: "http-client-no-timeout",
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

func TestBundledSourcePlaybookSafeFixturesDoNotMatch(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t)})

	tests := []struct {
		name string
		dir  string
	}{
		{name: "missing error propagation checked error", dir: filepath.Join("testdata", "source", "missing-error-propagation-safe")},
		{name: "missing error propagation ignores javascript noise", dir: filepath.Join("testdata", "source", "missing-error-propagation-js-noise")},
		{name: "panic in http handler recovered", dir: filepath.Join("testdata", "source", "panic-in-http-handler-safe")},
		{name: "unawaited promise awaited", dir: filepath.Join("testdata", "source", "unawaited-promise-safe")},
		{name: "unawaited promise returned", dir: filepath.Join("testdata", "source", "unawaited-promise-return-safe")},
		{name: "unawaited promise caught", dir: filepath.Join("testdata", "source", "unawaited-promise-catch-safe")},
		{name: "continue on error optional step", dir: filepath.Join("testdata", "source", "continue-on-error-critical-step-safe")},
		{name: "ignored shell exit optional probe", dir: filepath.Join("testdata", "source", "ignored-shell-exit-in-ci-safe")},
		{name: "floating docker base image pinned", dir: filepath.Join("testdata", "source", "floating-docker-base-image-safe")},
		{name: "shell dialect mismatch uses bash", dir: filepath.Join("testdata", "source", "shell-dialect-mismatch-safe")},
		{name: "continue on error notification step only", dir: filepath.Join("testdata", "source", "continue-on-error-critical-step-noisy2")},
		{name: "hardcoded secret env var lookup", dir: filepath.Join("testdata", "source", "hardcoded-secret-env-safe")},
		{name: "goroutine with context cancel", dir: filepath.Join("testdata", "source", "goroutine-leak-context-safe")},
		{name: "insecure tls verify enabled", dir: filepath.Join("testdata", "source", "insecure-tls-skip-verify-safe")},
		{name: "transaction with deferred rollback", dir: filepath.Join("testdata", "source", "missing-transaction-rollback-safe")},
		{name: "http client with timeout", dir: filepath.Join("testdata", "source", "http-client-no-timeout-safe")},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assertSourceFixtureNoMatch(t, e, tc.dir)
		})
	}
}

func TestAnalyzeRepositoryIgnoresVirtualEnvNoise(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t)})

	analysis, err := e.AnalyzeRepository(
		filepath.Join("testdata", "source", "missing-error-propagation-venv-noise"),
		detectors.ChangeSet{},
	)
	if err != ErrNoMatch {
		t.Fatalf("expected ErrNoMatch for virtualenv-only source risk, got %v", err)
	}
	if analysis == nil {
		t.Fatal("expected non-nil analysis for ignored virtualenv noise fixture")
	}
	if len(analysis.Results) != 0 {
		t.Fatalf("expected no source results for virtualenv noise fixture, got %v", resultIDs(analysis.Results))
	}
}

func TestAnalyzeRepositoryIgnoresTestOnlyPanicNoise(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t)})

	analysis, err := e.AnalyzeRepository(
		filepath.Join("testdata", "source", "panic-in-http-handler-test-only-safe"),
		detectors.ChangeSet{},
	)
	if err != ErrNoMatch {
		t.Fatalf("expected ErrNoMatch for test-only panic fixture, got %v", err)
	}
	if analysis == nil {
		t.Fatal("expected non-nil analysis for test-only panic fixture")
	}
	if len(analysis.Results) != 0 {
		t.Fatalf("expected no source results for test-only panic fixture, got %v", resultIDs(analysis.Results))
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

func assertSourceFixtureNoMatch(t *testing.T, e *Engine, dir string) {
	t.Helper()

	analysis, err := e.AnalyzeRepository(dir, detectors.ChangeSet{})
	if err != ErrNoMatch {
		t.Fatalf("expected ErrNoMatch for source fixture %s, got %v", dir, err)
	}
	if analysis == nil {
		t.Fatalf("expected non-nil analysis for source fixture %s", dir)
	}
	if len(analysis.Results) != 0 {
		t.Fatalf("expected source fixture %s to stay unmatched, got %v", dir, resultIDs(analysis.Results))
	}
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
