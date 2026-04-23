package signature

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/engine"
)

func TestSignatureVariantMatrix(t *testing.T) {
	root := repoRoot(t)
	playbookDir := filepath.Join(root, "playbooks", "bundled")
	testdataDir := filepath.Join(root, "internal", "signature", "testdata", "variants")

	cases := []struct {
		name         string
		file         string
		wantPlaybook string
		group        string
	}{
		{name: "missing executable linux runner", file: "missing-executable-1.log", wantPlaybook: "missing-executable", group: "missing-executable"},
		{name: "missing executable windows runner", file: "missing-executable-2.log", wantPlaybook: "missing-executable", group: "missing-executable"},
		{name: "missing executable hosted toolcache", file: "missing-executable-3.log", wantPlaybook: "missing-executable", group: "missing-executable"},
		{name: "npm lockfile mismatch linux", file: "npm-ci-lockfile-1.log", wantPlaybook: "npm-ci-lockfile", group: "npm-ci-lockfile"},
		{name: "npm lockfile mismatch windows", file: "npm-ci-lockfile-2.log", wantPlaybook: "npm-ci-lockfile", group: "npm-ci-lockfile"},
		{name: "node version mismatch linux", file: "node-version-mismatch-1.log", wantPlaybook: "node-version-mismatch", group: "node-version-mismatch"},
		{name: "node version mismatch toolcache", file: "node-version-mismatch-2.log", wantPlaybook: "node-version-mismatch", group: "node-version-mismatch"},
		{name: "env var missing github actions", file: "env-var-missing-1.log", wantPlaybook: "env-var-missing", group: "env-var-missing-api-base-url"},
		{name: "env var missing windows runner", file: "env-var-missing-2.log", wantPlaybook: "env-var-missing", group: "env-var-missing-api-base-url"},
		{name: "env var missing distinct variable", file: "env-var-missing-3.log", wantPlaybook: "env-var-missing", group: "env-var-missing-database-url"},
		{name: "dependency drift react", file: "dependency-drift-react.log", wantPlaybook: "dependency-drift", group: "dependency-drift-react"},
		{name: "dependency drift grpc", file: "dependency-drift-grpc.log", wantPlaybook: "dependency-drift", group: "dependency-drift-grpc"},
	}

	eng := engine.New(engine.Options{PlaybookDir: playbookDir, NoHistory: true})
	groupHashes := map[string]string{}
	allGroups := map[string]string{}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(testdataDir, tc.file))
			if err != nil {
				t.Fatalf("read variant fixture: %v", err)
			}

			analysis, err := eng.AnalyzeReader(strings.NewReader(string(data)))
			if err != nil {
				t.Fatalf("AnalyzeReader: %v", err)
			}
			if analysis == nil || len(analysis.Results) == 0 {
				t.Fatalf("expected a matched result for %s", tc.file)
			}
			if got := analysis.Results[0].Playbook.ID; got != tc.wantPlaybook {
				t.Fatalf("expected top playbook %s, got %s", tc.wantPlaybook, got)
			}

			sig := ForResult(analysis.Results[0])
			if sig.Hash == "" {
				t.Fatalf("expected signature hash for %s", tc.file)
			}

			if existing, ok := groupHashes[tc.group]; ok && existing != sig.Hash {
				t.Fatalf("expected group %s to remain stable, got %s and %s", tc.group, existing, sig.Hash)
			}
			groupHashes[tc.group] = sig.Hash
		})
	}

	for group, hash := range groupHashes {
		if other, ok := allGroups[hash]; ok {
			t.Fatalf("expected groups %s and %s to stay distinct, but both normalized to %s", group, other, hash)
		}
		allGroups[hash] = group
	}
}
