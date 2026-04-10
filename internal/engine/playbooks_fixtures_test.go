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
		{name: "ssh key auth", file: "ssh-key-auth.log", wantID: "ssh-key-auth"},
		{name: "npm ci lockfile", file: "npm-ci-lockfile.log", wantID: "npm-ci-lockfile"},
		{name: "yarn lockfile", file: "yarn-lockfile.log", wantID: "yarn-lockfile"},
		{name: "working directory", file: "working-directory.log", wantID: "working-directory"},
		{name: "runtime mismatch", file: "runtime-mismatch.log", wantID: "runtime-mismatch"},
		{name: "env var missing", file: "env-var-missing.log", wantID: "env-var-missing"},
		{name: "missing env", file: "missing-env.log", wantID: "missing-env"},
		{name: "snapshot mismatch", file: "snapshot-mismatch.log", wantID: "snapshot-mismatch"},
		{name: "flaky test", file: "flaky-test.log", wantID: "flaky-test"},
		{name: "test timeout", file: "test-timeout.log", wantID: "test-timeout"},
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
