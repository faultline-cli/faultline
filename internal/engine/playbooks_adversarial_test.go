package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/matcher"
	"faultline/internal/playbooks"
)

func TestBundledPlaybookAdversarialFixtures(t *testing.T) {
	pbs, err := playbooks.LoadDir(repoPlaybookDir(t))
	if err != nil {
		t.Fatalf("load playbooks: %v", err)
	}

	tests := []struct {
		name      string
		file      string
		wantTopID string
		absentIDs []string
	}{
		{
			name:      "database isolation beats generic order wording on unique violations",
			file:      "database-test-isolation.log",
			wantTopID: "database-test-isolation",
			absentIDs: []string{"order-dependency"},
		},
		{
			name:      "order dependency beats database isolation when logs describe prior-test state",
			file:      "order-dependency.log",
			wantTopID: "order-dependency",
			absentIDs: []string{"database-test-isolation"},
		},
		{
			name:      "runner disk full beats generic disk full wording",
			file:      "runner-disk-full.log",
			wantTopID: "runner-disk-full",
			absentIDs: []string{"disk-full"},
		},
		{
			name:      "generic disk full stays separate from runner quota signals",
			file:      "disk-full.log",
			wantTopID: "disk-full",
			absentIDs: []string{"runner-disk-full"},
		},
		{
			name:      "parallel test conflict beats runtime and deploy port wording",
			file:      "parallelism-conflict.log",
			wantTopID: "parallelism-conflict",
			absentIDs: []string{"port-in-use", "port-conflict"},
		},
		{
			name:      "runtime port in use beats test and deploy port wording",
			file:      "port-in-use.log",
			wantTopID: "port-in-use",
			absentIDs: []string{"parallelism-conflict", "port-conflict"},
		},
		{
			name:      "deploy port conflict beats runtime and test port wording",
			file:      "port-conflict.log",
			wantTopID: "port-conflict",
			absentIDs: []string{"parallelism-conflict", "port-in-use"},
		},
		{
			name:      "secret-empty auth failure beats runtime and deploy env wording",
			file:      "missing-env.log",
			wantTopID: "missing-env",
			absentIDs: []string{"env-var-missing", "config-mismatch"},
		},
		{
			name:      "ci secret availability stays with ci rule instead of generic env rules",
			file:      "secrets-not-available.log",
			wantTopID: "secrets-not-available",
			absentIDs: []string{"missing-env", "env-var-missing"},
		},
		{
			name:      "runtime env lookup beats auth and deploy env wording",
			file:      "env-var-missing.log",
			wantTopID: "env-var-missing",
			absentIDs: []string{"missing-env", "config-mismatch"},
		},
		{
			name:      "deploy config mismatch beats auth and runtime env wording",
			file:      "config-mismatch.log",
			wantTopID: "config-mismatch",
			absentIDs: []string{"missing-env", "env-var-missing", "missing-test-fixture"},
		},
		{
			name:      "connection refused beats generic failed-to-connect timeout wording",
			file:      "connection-refused.log",
			wantTopID: "connection-refused",
			absentIDs: []string{"network-timeout"},
		},
		{
			name:      "dns resolution stays above generic timeout wording when lookup evidence is present",
			file:      "dns-resolution.log",
			wantTopID: "dns-resolution",
			absentIDs: []string{"network-timeout"},
		},
		{
			name:      "working directory beats generic missing file wording in build logs",
			file:      "working-directory.log",
			wantTopID: "working-directory",
			absentIDs: []string{"missing-test-fixture"},
		},
		{
			name:      "missing test fixture beats generic missing file wording in test logs",
			file:      "missing-test-fixture.log",
			wantTopID: "missing-test-fixture",
			absentIDs: []string{"working-directory"},
		},
		{
			name:      "generic missing testdata path stays with missing test fixture",
			file:      "generic-fixture-path-missing.log",
			wantTopID: "missing-test-fixture",
			absentIDs: []string{"working-directory"},
		},
		{
			name:      "generic repo path miss stays with working directory",
			file:      "generic-working-directory-path-missing.log",
			wantTopID: "working-directory",
			absentIDs: []string{"missing-test-fixture"},
		},
		{
			name:      "network timeout beats test timeout style noise",
			file:      "network-timeout.log",
			wantTopID: "network-timeout",
			absentIDs: []string{"connection-refused", "test-timeout", "pipeline-timeout"},
		},
		{
			name:      "test context deadline stays with test timeout when runner evidence is present",
			file:      "test-context-deadline.log",
			wantTopID: "test-timeout",
			absentIDs: []string{"network-timeout"},
		},
		{
			name:      "network context deadline stays with network timeout outside test runner context",
			file:      "network-context-deadline.log",
			wantTopID: "network-timeout",
			absentIDs: []string{"test-timeout"},
		},
		{
			name:      "oom kill beats container crash symptom log",
			file:      "oom-killed.log",
			wantTopID: "oom-killed",
			absentIDs: []string{"container-crash"},
		},
		{
			name:      "pipeline timeout beats lower-level timeout noise",
			file:      "pipeline-timeout.log",
			wantTopID: "pipeline-timeout",
			absentIDs: []string{"test-timeout"},
		},
		{
			name:      "container crash beats non-oom restart noise",
			file:      "container-crash.log",
			wantTopID: "container-crash",
			absentIDs: []string{"oom-killed"},
		},
		{
			name:      "health check failure stays above crash and config neighbors when probe evidence is present",
			file:      "health-check-failure.log",
			wantTopID: "health-check-failure",
			absentIDs: []string{"container-crash", "config-mismatch"},
		},
		{
			name:      "python dependency import failure ranks above neighboring rules",
			file:      "python-module-missing.log",
			wantTopID: "python-module-missing",
			absentIDs: []string{"path-case-mismatch", "typescript-compile"},
		},
		{
			name:      "test timeout beats generic network timeout wording",
			file:      "test-timeout.log",
			wantTopID: "test-timeout",
			absentIDs: []string{"network-timeout", "pipeline-timeout"},
		},
		{
			name:      "javascript module resolution noise does not trigger python rule",
			file:      "python-module-missing-negative.log",
			wantTopID: "path-case-mismatch",
			absentIDs: []string{"python-module-missing"},
		},
		{
			name:      "package auth failure stays with missing env instead of package not found",
			file:      "install-failure-negative.log",
			wantTopID: "missing-env",
			absentIDs: []string{"install-failure"},
		},
		{
			name:      "coverage gate stays separate from timeout family",
			file:      "coverage-gate-failure.log",
			wantTopID: "coverage-gate-failure",
			absentIDs: []string{"test-timeout", "network-timeout", "pipeline-timeout"},
		},
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
			if got := results[0].Playbook.ID; got != tc.wantTopID {
				t.Fatalf("expected top match %s, got %s", tc.wantTopID, got)
			}
			for _, absentID := range tc.absentIDs {
				if containsPlaybook(results, absentID) {
					t.Fatalf("expected %s to be absent for %s, got %v", absentID, tc.file, resultIDs(results))
				}
			}
		})
	}
}
