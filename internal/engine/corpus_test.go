package engine

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAnalyzeReaderCorpusReleaseGate(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	tests := []struct {
		name      string
		file      string
		wantTopID string
		wantStage string
		wantErr   error
		absentIDs []string
	}{
		{
			name:      "database isolation noisy test log",
			file:      "database-test-isolation-noisy.log",
			wantTopID: "database-test-isolation",
			wantStage: "test",
			absentIDs: []string{"order-dependency"},
		},
		{
			name:      "docker auth noisy build log",
			file:      "docker-auth-noisy.log",
			wantTopID: "docker-auth",
			wantStage: "build",
			absentIDs: []string{"image-pull-backoff"},
		},
		{
			name:      "image pull backoff noisy deploy log",
			file:      "image-pull-backoff-noisy.log",
			wantTopID: "image-pull-backoff",
			wantStage: "deploy",
			absentIDs: []string{"docker-auth"},
		},
		{
			name:      "network timeout noisy build log",
			file:      "network-timeout-noisy.log",
			wantTopID: "network-timeout",
			wantStage: "build",
			absentIDs: []string{"dns-resolution"},
		},
		{
			name:      "connection refused noisy test log",
			file:      "connection-refused-noisy.log",
			wantTopID: "connection-refused",
			wantStage: "test",
			absentIDs: []string{"network-timeout"},
		},
		{
			name:      "pipeline timeout noisy build log",
			file:      "pipeline-timeout-noisy.log",
			wantTopID: "pipeline-timeout",
			wantStage: "build",
		},
		{
			name:      "terraform state lock noisy deploy log",
			file:      "terraform-state-lock-noisy.log",
			wantTopID: "terraform-state-lock",
			wantStage: "deploy",
			absentIDs: []string{"terraform-init", "terraform-apply-error"},
		},
		{
			name:      "parallelism conflict noisy test log",
			file:      "parallelism-conflict-noisy.log",
			wantTopID: "parallelism-conflict",
			wantStage: "test",
		},
		{
			name:      "missing test fixture noisy test log",
			file:      "missing-test-fixture-noisy.log",
			wantTopID: "missing-test-fixture",
			wantStage: "test",
			absentIDs: []string{"working-directory"},
		},
		{
			name:      "python module missing noisy test log",
			file:      "python-module-missing-noisy.log",
			wantTopID: "python-module-missing",
			wantStage: "test",
			absentIDs: []string{"path-case-mismatch", "typescript-compile"},
		},
		{
			name:      "config mismatch noisy deploy log",
			file:      "config-mismatch-noisy.log",
			wantTopID: "config-mismatch",
			wantStage: "deploy",
			absentIDs: []string{"missing-env", "env-var-missing", "missing-test-fixture"},
		},
		{
			name:      "port in use noisy deploy log",
			file:      "port-in-use-noisy.log",
			wantTopID: "port-in-use",
			wantStage: "deploy",
		},
		{
			name:      "container crash noisy deploy log",
			file:      "container-crash-noisy.log",
			wantTopID: "k8s-crashloopbackoff",
			wantStage: "deploy",
			absentIDs: []string{"oom-killed"},
		},
		{
			name:      "snapshot mismatch noisy test log",
			file:      "snapshot-mismatch-noisy.log",
			wantTopID: "snapshot-mismatch",
			wantStage: "test",
		},
		{
			name:    "no match success log",
			file:    "no-match-success.log",
			wantErr: ErrNoMatch,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", "corpus", tc.file))
			if err != nil {
				t.Fatalf("read corpus fixture %s: %v", tc.file, err)
			}

			analysis, err := e.AnalyzeReader(strings.NewReader(string(data)))
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				if analysis == nil {
					t.Fatal("expected no-match analysis payload")
				}
				if len(analysis.Results) != 0 {
					t.Fatalf("expected no results for %s, got %v", tc.file, resultIDs(analysis.Results))
				}

				again, err := e.AnalyzeReader(strings.NewReader(string(data)))
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected repeat error %v, got %v", tc.wantErr, err)
				}
				if !reflect.DeepEqual(analysis.Results, again.Results) || analysis.Context != again.Context {
					t.Fatalf("expected deterministic no-match analysis for %s", tc.file)
				}
				return
			}

			if err != nil {
				t.Fatalf("analyze %s: %v", tc.file, err)
			}
			if len(analysis.Results) == 0 {
				t.Fatalf("expected ranked results for %s", tc.file)
			}
			if got := analysis.Results[0].Playbook.ID; got != tc.wantTopID {
				t.Fatalf("expected top match %s, got %s (all results: %v)", tc.wantTopID, got, resultIDs(analysis.Results))
			}
			if tc.wantStage != "" && analysis.Context.Stage != tc.wantStage {
				t.Fatalf("expected stage %s, got %s", tc.wantStage, analysis.Context.Stage)
			}
			if len(analysis.Results[0].Evidence) == 0 {
				t.Fatalf("expected evidence for %s", tc.wantTopID)
			}
			for _, absentID := range tc.absentIDs {
				if containsPlaybook(analysis.Results, absentID) {
					t.Fatalf("expected %s to be excluded for %s, got %v", absentID, tc.file, resultIDs(analysis.Results))
				}
			}

			again, err := e.AnalyzeReader(strings.NewReader(string(data)))
			if err != nil {
				t.Fatalf("repeat analyze %s: %v", tc.file, err)
			}
			if !reflect.DeepEqual(analysis.Results, again.Results) || analysis.Context != again.Context || analysis.Fingerprint != again.Fingerprint {
				t.Fatalf("expected deterministic analysis for %s", tc.file)
			}
		})
	}
}
