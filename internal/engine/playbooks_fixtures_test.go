package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/matcher"
	"faultline/internal/playbooks"
)

func TestBundledPlaybookFixtures(t *testing.T) {
	pbs, err := playbooks.LoadDir(repoPlaybookDir(t))
	if err != nil {
		t.Fatalf("load playbooks: %v", err)
	}

	tests := []struct {
		name   string
		file   string
		wantID string
	}{
		{name: "docker auth", file: "docker-auth.log", wantID: "docker-auth"},
		{name: "database test isolation", file: "database-test-isolation.log", wantID: "database-test-isolation"},
		{name: "image pull backoff", file: "image-pull-backoff.log", wantID: "image-pull-backoff"},
		{name: "git auth", file: "git-auth.log", wantID: "git-auth"},
		{name: "runner disk full", file: "runner-disk-full.log", wantID: "runner-disk-full"},
		{name: "connection refused", file: "connection-refused.log", wantID: "connection-refused"},
		{name: "dns resolution", file: "dns-resolution.log", wantID: "dns-resolution"},
		{name: "disk full", file: "disk-full.log", wantID: "disk-full"},
		{name: "network context deadline", file: "network-context-deadline.log", wantID: "network-timeout"},
		{name: "network timeout", file: "network-timeout.log", wantID: "network-timeout"},
		{name: "npm ci lockfile", file: "npm-ci-lockfile.log", wantID: "npm-ci-lockfile"},
		{name: "oom killed", file: "oom-killed.log", wantID: "oom-killed"},
		{name: "parallelism conflict", file: "parallelism-conflict.log", wantID: "parallelism-conflict"},
		{name: "pipeline timeout", file: "pipeline-timeout.log", wantID: "pipeline-timeout"},
		{name: "artifact upload failure", file: "artifact-upload-failure.log", wantID: "artifact-upload-failure"},
		{name: "port conflict", file: "port-conflict.log", wantID: "port-conflict"},
		{name: "port in use", file: "port-in-use.log", wantID: "port-in-use"},
		{name: "python module missing", file: "python-module-missing.log", wantID: "python-module-missing"},
		{name: "install failure", file: "install-failure.log", wantID: "install-failure"},
		{name: "yarn lockfile", file: "yarn-lockfile.log", wantID: "yarn-lockfile"},
		{name: "working directory", file: "working-directory.log", wantID: "working-directory"},
		{name: "node version mismatch", file: "runtime-mismatch.log", wantID: "node-version-mismatch"},
		{name: "ssl cert error", file: "ssl-cert-error.log", wantID: "ssl-cert-error"},
		{name: "container crash", file: "container-crash.log", wantID: "container-crash"},
		{name: "health check failure", file: "health-check-failure.log", wantID: "health-check-failure"},
		{name: "config mismatch", file: "config-mismatch.log", wantID: "config-mismatch"},
		{name: "env var missing", file: "env-var-missing.log", wantID: "env-var-missing"},
		{name: "missing test fixture", file: "missing-test-fixture.log", wantID: "missing-test-fixture"},
		{name: "missing env", file: "missing-env.log", wantID: "missing-env"},
		{name: "secrets not available", file: "secrets-not-available.log", wantID: "secrets-not-available"},
		{name: "order dependency", file: "order-dependency.log", wantID: "order-dependency"},
		{name: "snapshot mismatch", file: "snapshot-mismatch.log", wantID: "snapshot-mismatch"},
		{name: "flaky test", file: "flaky-test.log", wantID: "flaky-test"},
		{name: "generic fixture path missing", file: "generic-fixture-path-missing.log", wantID: "missing-test-fixture"},
		{name: "generic working directory path missing", file: "generic-working-directory-path-missing.log", wantID: "working-directory"},
		{name: "test timeout", file: "test-timeout.log", wantID: "test-timeout"},
		{name: "test context deadline", file: "test-context-deadline.log", wantID: "test-timeout"},
		{name: "quality gate failure", file: "quality-gate-failure.log", wantID: "quality-gate-failure"},
		{name: "coverage gate failure", file: "coverage-gate-failure.log", wantID: "coverage-gate-failure"},
		{name: "dependency drift", file: "dependency-drift.log", wantID: "dependency-drift"},
		{name: "go compile error", file: "go-compile-error.log", wantID: "go-compile-error"},
		{name: "segfault", file: "segfault.log", wantID: "segfault"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			logPath := filepath.Join("testdata", "fixtures", tc.file)
			data, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("read fixture %s: %v", logPath, err)
			}

			lines, err := readLines(strings.NewReader(string(data)))
			if err != nil {
				t.Fatalf("read lines: %v", err)
			}
			ctx := ExtractContext(lines)
			results := matcher.Rank(pbs, lines, ctx)
			if len(results) == 0 {
				t.Fatalf("expected fixture %s to match at least one playbook", tc.file)
			}
			if got := results[0].Playbook.ID; got != tc.wantID {
				t.Fatalf("expected top match %s, got %s", tc.wantID, got)
			}
			if len(results[0].Evidence) == 0 {
				t.Fatalf("expected evidence for %s", tc.wantID)
			}

			again := matcher.Rank(pbs, lines, ctx)
			if len(again) != len(results) {
				t.Fatalf("expected deterministic result count for %s", tc.wantID)
			}
			for i := range results {
				if results[i].Playbook.ID != again[i].Playbook.ID ||
					results[i].Score != again[i].Score ||
					results[i].Confidence != again[i].Confidence {
					t.Fatalf("expected deterministic ranking for %s", tc.wantID)
				}
			}
		})
	}
}

func TestExtraPackFixtures(t *testing.T) {
	extraDir := requireExtraPack(t)
	pbs, err := playbooks.LoadPacks([]playbooks.Pack{
		{Name: playbooks.BundledPackName, Root: repoPlaybookDir(t)},
		{Name: "extra-local", Root: extraDir},
	})
	if err != nil {
		t.Fatalf("load playbook packs: %v", err)
	}

	tests := []struct {
		name   string
		file   string
		wantID string
	}{
		{name: "dotnet build", file: "dotnet-build.log", wantID: "dotnet-build"},
		{name: "rubocop failure", file: "rubocop-failure.log", wantID: "rubocop-failure"},
		{name: "rspec failure", file: "rspec-failure.log", wantID: "rspec-failure"},
		{name: "vitest failure", file: "vitest-failure.log", wantID: "vitest-failure"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			logPath := filepath.Join("testdata", "fixtures", tc.file)
			data, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("read fixture %s: %v", logPath, err)
			}

			lines, err := readLines(strings.NewReader(string(data)))
			if err != nil {
				t.Fatalf("read lines: %v", err)
			}
			ctx := ExtractContext(lines)
			results := matcher.Rank(pbs, lines, ctx)
			if len(results) == 0 {
				t.Fatalf("expected fixture %s to match at least one playbook", tc.file)
			}
			if got := results[0].Playbook.ID; got != tc.wantID {
				t.Fatalf("expected top match %s, got %s", tc.wantID, got)
			}
			if results[0].Playbook.Metadata.PackName != "extra-local" {
				t.Fatalf("expected extra-local pack metadata, got %#v", results[0].Playbook.Metadata)
			}
		})
	}
}
