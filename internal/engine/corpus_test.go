package engine

import (
	"errors"
	"faultline/internal/playbooks"
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
			name:      "parallelism conflict noisy test log",
			file:      "parallelism-conflict-noisy.log",
			wantTopID: "parallelism-conflict",
			wantStage: "test",
		},
		{
			name:      "pytest fixture error noisy test log",
			file:      "pytest-fixture-error-noisy.log",
			wantTopID: "pytest-fixture-error",
			wantStage: "test",
			absentIDs: []string{"pytest-no-tests"},
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
			name:      "go compile error noisy build log",
			file:      "go-compile-error-noisy.log",
			wantTopID: "go-compile-error",
			wantStage: "build",
		},
		{
			name:      "typescript compile noisy build log",
			file:      "typescript-compile-noisy.log",
			wantTopID: "typescript-compile",
			wantStage: "build",
		},
		{
			name:      "test timeout noisy test log",
			file:      "test-timeout-noisy.log",
			wantTopID: "test-timeout",
			wantStage: "test",
		},
		{
			name:      "testcontainer startup noisy test log",
			file:      "testcontainer-startup-noisy.log",
			wantTopID: "testcontainer-startup",
			wantStage: "test",
			absentIDs: []string{"docker-daemon-unavailable"},
		},
		{
			name:      "go data race noisy test log",
			file:      "go-data-race-noisy.log",
			wantTopID: "go-data-race",
			wantStage: "test",
		},
		{
			name:      "disk full noisy runtime log",
			file:      "disk-full-noisy.log",
			wantTopID: "disk-full",
			wantStage: "build",
		},
		{
			name:      "permission denied noisy runtime log",
			file:      "permission-denied-noisy.log",
			wantTopID: "permission-denied",
		},
		{
			name:      "missing executable noisy build log",
			file:      "missing-executable-noisy.log",
			wantTopID: "missing-executable",
			wantStage: "build",
		},
		{
			name:      "runtime mismatch noisy build log",
			file:      "runtime-mismatch-noisy.log",
			wantTopID: "runtime-mismatch",
		},
		{
			name:      "ssl cert error noisy network log",
			file:      "ssl-cert-error-noisy.log",
			wantTopID: "ssl-cert-error",
			wantStage: "test",
		},
		{
			name:      "git auth noisy auth log",
			file:      "git-auth-noisy.log",
			wantTopID: "git-auth",
		},
		{
			name:      "git shallow checkout noisy ci log",
			file:      "git-shallow-checkout-noisy.log",
			wantTopID: "git-shallow-checkout",
		},
		{
			name:      "build input file missing noisy build log",
			file:      "build-input-file-missing-noisy.log",
			wantTopID: "node-gyp-missing-build-tools",
			wantStage: "build",
		},
		{
			name:      "buildkit session lost noisy build log",
			file:      "buildkit-session-lost-noisy.log",
			wantTopID: "buildkit-session-lost",
			wantStage: "build",
		},
		{
			name:      "npm registry auth noisy build log",
			file:      "npm-registry-auth-noisy.log",
			wantTopID: "npm-registry-auth",
			absentIDs: []string{"install-failure"},
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

func TestAnalyzeReaderExtraPackCorpus(t *testing.T) {
	// Skip test if extra pack is not available
	extraDir := ""
	for _, path := range []string{
		"../../playbooks/packs/extra-local",
		"../../../faultline-extra-pack",
	} {
		if _, err := os.Stat(path); err == nil {
			extraDir = path
			break
		}
	}
	if extraDir == "" {
		t.Skip("extra pack repository is not available locally")
	}

	t.Setenv("FAULTLINE_PLAYBOOK_DIR", repoPlaybookDir(t))
	e := New(Options{PlaybookPackDirs: []string{extraDir}, NoHistory: true})

	data, err := os.ReadFile(filepath.Join("testdata", "corpus", "terraform-state-lock-noisy.log"))
	if err != nil {
		t.Fatalf("read corpus fixture terraform-state-lock-noisy.log: %v", err)
	}

	analysis, err := e.AnalyzeReader(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("analyze terraform-state-lock-noisy.log: %v", err)
	}
	if len(analysis.Results) == 0 {
		t.Fatal("expected ranked results for extra-pack corpus fixture")
	}
	if got := analysis.Results[0].Playbook.ID; got != "terraform-state-lock" {
		t.Fatalf("expected top match terraform-state-lock, got %s", got)
	}
	if analysis.Results[0].Playbook.Metadata.PackName == "" || analysis.Results[0].Playbook.Metadata.PackName == playbooks.BundledPackName {
		t.Fatalf("expected extra-pack metadata, got %#v", analysis.Results[0].Playbook.Metadata)
	}
}
