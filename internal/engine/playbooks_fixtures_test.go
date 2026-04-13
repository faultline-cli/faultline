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
		{name: "image pull backoff", file: "image-pull-backoff.log", wantID: "image-pull-backoff"},
		{name: "git auth", file: "git-auth.log", wantID: "git-auth"},
		{name: "runner disk full", file: "runner-disk-full.log", wantID: "runner-disk-full"},
		{name: "connection refused", file: "connection-refused.log", wantID: "connection-refused"},
		{name: "disk full", file: "disk-full.log", wantID: "disk-full"},
		{name: "network timeout", file: "network-timeout.log", wantID: "network-timeout"},
		{name: "npm ci lockfile", file: "npm-ci-lockfile.log", wantID: "npm-ci-lockfile"},
		{name: "oom killed", file: "oom-killed.log", wantID: "oom-killed"},
		{name: "parallelism conflict", file: "parallelism-conflict.log", wantID: "parallelism-conflict"},
		{name: "pipeline timeout", file: "pipeline-timeout.log", wantID: "pipeline-timeout"},
		{name: "port conflict", file: "port-conflict.log", wantID: "port-conflict"},
		{name: "port in use", file: "port-in-use.log", wantID: "port-in-use"},
		{name: "python module missing", file: "python-module-missing.log", wantID: "python-module-missing"},
		{name: "yarn lockfile", file: "yarn-lockfile.log", wantID: "yarn-lockfile"},
		{name: "working directory", file: "working-directory.log", wantID: "working-directory"},
		{name: "runtime mismatch", file: "runtime-mismatch.log", wantID: "runtime-mismatch"},
		{name: "container crash", file: "container-crash.log", wantID: "container-crash"},
		{name: "config mismatch", file: "config-mismatch.log", wantID: "config-mismatch"},
		{name: "env var missing", file: "env-var-missing.log", wantID: "env-var-missing"},
		{name: "missing test fixture", file: "missing-test-fixture.log", wantID: "missing-test-fixture"},
		{name: "missing env", file: "missing-env.log", wantID: "missing-env"},
		{name: "snapshot mismatch", file: "snapshot-mismatch.log", wantID: "snapshot-mismatch"},
		{name: "flaky test", file: "flaky-test.log", wantID: "flaky-test"},
		{name: "test timeout", file: "test-timeout.log", wantID: "test-timeout"},
		{name: "quality gate failure", file: "quality-gate-failure.log", wantID: "quality-gate-failure"},
		{name: "coverage gate failure", file: "coverage-gate-failure.log", wantID: "coverage-gate-failure"},
		{name: "dependency drift", file: "dependency-drift.log", wantID: "dependency-drift"},
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
