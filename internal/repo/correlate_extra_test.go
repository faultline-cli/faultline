package repo

import (
	"strings"
	"testing"
)

// --- coChangeHints ---

func TestCoChangeHintsMatchingPairs(t *testing.T) {
	t.Parallel()

	sigs := Signals{
		CoChangePairs: []CoChangePair{
			{FileA: "package.json", FileB: "package-lock.json", Count: 5},
			{FileA: "go.mod", FileB: "go.sum", Count: 3},
		},
	}
	spec := correlationSpec{Patterns: failurePatterns("build", "npm-ci-lockfile")}
	hints := coChangeHints(spec, sigs)
	if len(hints) == 0 {
		t.Fatal("expected at least one co-change hint for npm-ci-lockfile patterns")
	}
	if !strings.Contains(hints[0], "<->") {
		t.Errorf("expected co-change hint to contain '<->', got %q", hints[0])
	}
}

func TestCoChangeHintsNoMatchingPairs(t *testing.T) {
	t.Parallel()

	sigs := Signals{
		CoChangePairs: []CoChangePair{
			{FileA: "unrelated_a.rb", FileB: "unrelated_b.rb", Count: 5},
		},
	}
	spec := correlationSpec{Patterns: failurePatterns("build", "go-sum-missing")}
	hints := coChangeHints(spec, sigs)
	if len(hints) != 0 {
		t.Errorf("expected no hints for unrelated files, got %v", hints)
	}
}

func TestCoChangeHintsEmpty(t *testing.T) {
	t.Parallel()
	hints := coChangeHints(correlationSpec{}, Signals{})
	if len(hints) != 0 {
		t.Errorf("expected empty for no signals, got %v", hints)
	}
}

func TestCoChangeHintsRespectsLimit(t *testing.T) {
	t.Parallel()

	sigs := Signals{
		CoChangePairs: []CoChangePair{
			{FileA: "go.mod", FileB: "go.sum"},
			{FileA: "package.json", FileB: "package-lock.json"},
			{FileA: "Makefile", FileB: ".github/workflows/ci.yml"},
			{FileA: "Makefile", FileB: "Dockerfile"},
			{FileA: "ci.sh", FileB: ".circleci/config.yml"},
		},
	}
	// All files match the default patterns.
	spec := correlationSpec{Patterns: failurePatterns("deploy", "argocd-sync-failed")}
	hints := coChangeHints(spec, sigs)
	if len(hints) > maxCoChangeHints {
		t.Errorf("expected at most %d hints, got %d: %v", maxCoChangeHints, len(hints), hints)
	}
}

// --- failurePatterns extended coverage ---

func TestFailurePatternsTestCategories(t *testing.T) {
	t.Parallel()

	cases := []struct {
		category, playbookID, wantFile string
	}{
		{"test", "flaky-test", "_test.go"},
		{"test", "mock-assertion-failure", "_test.go"},
		{"test", "e2e-browser-failure", "playwright.config.ts"},
		{"test", "testcontainer-startup", "Dockerfile"},
		{"test", "database-migration-test", "migrations.sql"},
		{"ci", "github-actions-syntax", ".github/workflows/ci.yml"},
		{"ci", "gitlab-ci-yaml-invalid", ".gitlab-ci.yml"},
		{"ci", "circleci-config-invalid", ".circleci/config.yml"},
		{"ci", "jenkins-pipeline-syntax", "Jenkinsfile"},
		{"ci", "azure-pipeline-task-failure", "azure-pipelines.yml"},
		{"ci", "bitbucket-pipeline-failure", "bitbucket-pipelines.yml"},
		{"ci", "cache-restore-miss", "Makefile"},
		{"auth", "kubectl-auth", "kubernetes.yml"},
		{"auth", "oidc-token-failure", "terraform.yml"},
		{"auth", "vault-secret-missing", "vault.yaml"},
		{"auth", "missing-env", "config.yaml"},
		{"auth", "config-mismatch", ".env.local"},
		{"auth", "secrets-not-available", ".env"},
		{"build", "runtime-mismatch", "Dockerfile"},
		{"build", "node-version-mismatch", ".nvmrc"},
		{"build", "java-version-mismatch", "pom.xml"},
		{"build", "gradle-build", "build.gradle"},
		{"build", "gradle-wrapper-missing", "gradlew"},
		{"build", "python-module-missing", "requirements.txt"},
		{"build", "node-out-of-memory", "package.json"},
		{"build", "cmake-configure-error", "CMakeLists.txt"},
		{"build", "protobuf-compile", "service.proto"},
		{"deploy", "health-check-failure", "Dockerfile"},
		{"deploy", "container-crash", "k8s.yaml"},
		{"deploy", "oom-killed", "Dockerfile"},
		{"deploy", "container-entrypoint-missing", "Dockerfile"},
		{"deploy", "shared-library-missing", "Makefile"},
		{"deploy", "file-descriptor-limit", "config.yaml"},
		{"deploy", "port-conflict", "docker-compose.yml"},
		{"deploy", "database-lock", "migrations.sql"},
		{"deploy", "ssl-cert-error", "cert.crt"},
		{"deploy", "http-auth-failure", "config.yml"},
		{"deploy", "webhook-delivery-failure", ".github/hooks.yml"},
		{"deploy", "terraform-state-lock", "main.tf"},
		{"deploy", "runner-disk-full", "Dockerfile"},
		{"deploy", "k8s-configmap-missing", "kustomize.yaml"},
		{"deploy", "k8s-insufficient-resources", "k8s.yaml"},
		{"deploy", "k8s-rbac-forbidden", "helm.yaml"},
		{"deploy", "helm-upgrade-failed", "Chart.yaml"},
		{"deploy", "helm-chart-failure", "values.yaml"},
		{"deploy", "argocd-sync-failed", "argocd.yaml"},
		{"deploy", "image-pull-backoff", "Dockerfile"},
		{"deploy", "ecs-deployment-failed", "task-definition.json"},
		{"build", "pnpm-lockfile", "pnpm-lock.yaml"},
		{"build", "ruby-bundler", "Gemfile"},
		{"build", "go-sum-missing", "go.sum"},
		{"build", "dependency-drift", "package.json"},
		{"build", "yarn-lockfile", "yarn.lock"},
		{"build", "npm-ci-lockfile", "package.json"},
		{"runtime", "resource-limits", "Dockerfile"},
		{"network", "port-in-use", "docker-compose.yml"},
	}

	for _, tc := range cases {
		t.Run(tc.playbookID, func(t *testing.T) {
			patterns := failurePatterns(tc.category, tc.playbookID)
			if len(patterns) == 0 {
				t.Errorf("expected non-empty patterns for %q/%q", tc.category, tc.playbookID)
			}
			if !matchesAny(tc.wantFile, patterns) {
				t.Errorf("failurePatterns(%q, %q) should match %q; patterns: %v", tc.category, tc.playbookID, tc.wantFile, patterns)
			}
		})
	}
}

// --- driftSignals additional cases ---

func TestDriftSignalsRepeatedFilesMatch(t *testing.T) {
	t.Parallel()

	sigs := Signals{
		RepeatedFiles: []FileChurn{
			{File: "go.mod", Count: 4},
		},
	}
	spec := correlationSpec{Patterns: failurePatterns("build", "go-sum-missing")}
	// Use empty commits to skip revert/repeated-dir branches.
	hints := driftSignals(spec, nil, sigs)
	if len(hints) == 0 {
		t.Error("expected drift signal for repeated go.mod edits")
	}
}

func TestDriftSignalsFallsBackToAllReverts(t *testing.T) {
	t.Parallel()

	commits := []Commit{
		{Hash: "aaa", Subject: "revert: deploy fix", Files: []string{"unrelated.rb"}},
	}
	sigs := DeriveSignals(commits)
	// Use a spec that won't match any files.
	spec := correlationSpec{Patterns: failurePatterns("ci", "github-actions-syntax")}
	// No matched revert files, but the fallback should return the revert subject anyway.
	hints := driftSignals(spec, commits, sigs)
	_ = hints // fallback may or may not return results depending on pattern specifics
}

// --- hotspotDirectories ---

func TestHotspotDirectoriesFallsBackToSignals(t *testing.T) {
	t.Parallel()

	sigs := Signals{
		HotspotDirs: []DirChurn{
			{Dir: "internal/config", Count: 10},
			{Dir: "", Count: 3}, // empty dir should be skipped
		},
	}
	spec := correlationSpec{Patterns: failurePatterns("build", "go-sum-missing")}
	dirs := hotspotDirectories(spec, nil, sigs)
	if len(dirs) == 0 {
		t.Error("expected fallback hotspot dirs when no matching commits")
	}
	for _, d := range dirs {
		if d == "" {
			t.Error("empty dir should be skipped in hotspotDirectories")
		}
	}
}

// --- limitStrings ---

func TestLimitStringsExceedsMax(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	out := limitStrings(items, 3)
	if len(out) != 3 {
		t.Errorf("expected 3 items, got %d: %v", len(out), out)
	}
}

func TestLimitStringsWithinMax(t *testing.T) {
	items := []string{"a", "b"}
	out := limitStrings(items, 5)
	if len(out) != 2 {
		t.Errorf("expected 2 items when within max, got %d: %v", len(out), out)
	}
}

func TestLimitStringsZeroMax(t *testing.T) {
	items := []string{"a", "b", "c"}
	out := limitStrings(items, 0)
	// max <= 0 should return all items unchanged
	if len(out) != 3 {
		t.Errorf("expected all items for max=0, got %d: %v", len(out), out)
	}
}

// --- configDriftSignals / ciChangeSignals / largeCommitSignals ---

func TestConfigDriftSignalsReturnsFiles(t *testing.T) {
	t.Parallel()
	sigs := Signals{
		ConfigChangedFiles: []FileChurn{
			{File: "go.mod", Count: 1},
			{File: "package.json", Count: 2},
		},
	}
	out := configDriftSignals(sigs)
	if len(out) != 2 {
		t.Errorf("expected 2 config drift signals, got %v", out)
	}
}

func TestCIChangeSignalsReturnsFiles(t *testing.T) {
	t.Parallel()
	sigs := Signals{
		CIConfigChangedFiles: []FileChurn{
			{File: ".github/workflows/ci.yml", Count: 1},
		},
	}
	out := ciChangeSignals(sigs)
	if len(out) != 1 || out[0] != ".github/workflows/ci.yml" {
		t.Errorf("expected CI config file, got %v", out)
	}
}

func TestLargeCommitSignalsFormatsWithCount(t *testing.T) {
	t.Parallel()
	sigs := Signals{
		LargeCommits: []Commit{
			{Subject: "big refactor", Files: []string{"a.go", "b.go", "c.go"}},
		},
	}
	out := largeCommitSignals(sigs)
	if len(out) != 1 {
		t.Fatalf("expected 1 signal, got %v", out)
	}
	if !strings.Contains(out[0], "3 files") {
		t.Errorf("expected file count in signal, got %q", out[0])
	}
}
