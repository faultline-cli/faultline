package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type commitSpec struct {
	Date    string
	Subject string
	Files   map[string]string
}

func TestParseSince(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "default", input: "", want: "--since=30 days ago"},
		{name: "days", input: "30d", want: "--since=30 days ago"},
		{name: "weeks", input: "2w", want: "--since=2 weeks ago"},
		{name: "months", input: "1m", want: "--since=1 month ago"},
		{name: "years", input: "1y", want: "--since=1 year ago"},
		{name: "bare integer", input: "7", want: "--since=7 days ago"},
		{name: "git native", input: "1 month ago", want: "--since=1 month ago"},
		{name: "invalid", input: "nonsense", want: "--since=30 days ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSince(tt.input)
			if err != nil {
				t.Fatalf("parseSince(%q): %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("parseSince(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseCommitBlock(t *testing.T) {
	t.Parallel()

	block := "abc123\x1e1712668800\x1efix: align deploy healthcheck\nDockerfile\nDockerfile\ninfra/deploy.yaml\n"
	commit, ok := parseCommitBlock(block)
	if !ok {
		t.Fatal("expected commit block to parse")
	}
	if commit.Hash != "abc123" {
		t.Fatalf("unexpected hash %q", commit.Hash)
	}
	if commit.Subject != "fix: align deploy healthcheck" {
		t.Fatalf("unexpected subject %q", commit.Subject)
	}
	if len(commit.Files) != 2 {
		t.Fatalf("expected duplicate files to be collapsed, got %v", commit.Files)
	}
	if commit.Files[0] != "Dockerfile" || commit.Files[1] != "infra/deploy.yaml" {
		t.Fatalf("unexpected files %v", commit.Files)
	}
}

func TestLoadHistoryAndSignals(t *testing.T) {
	repoDir := initTempRepo(t)
	writeCommits(t, repoDir, []commitSpec{
		{
			Date:    "2026-04-08T10:00:00Z",
			Subject: "feat: adjust health endpoint",
			Files: map[string]string{
				"cmd/server/main.go":      "package main\n",
				"deploy/healthcheck.yaml": "path: /healthz\n",
			},
		},
		{
			Date:    "2026-04-09T10:00:00Z",
			Subject: "hotfix: revert health timeout tweak",
			Files: map[string]string{
				"deploy/healthcheck.yaml": "path: /readyz\n",
				"Dockerfile":              "FROM golang:1.24\n",
			},
		},
		{
			Date:    "2026-04-10T10:00:00Z",
			Subject: "revert: restore startup probe",
			Files: map[string]string{
				"deploy/healthcheck.yaml": "path: /healthz\n",
			},
		},
	})

	scanner, err := NewScanner(repoDir)
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}

	commits, err := LoadHistory(scanner, "30d")
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(commits) != 3 {
		t.Fatalf("expected 3 commits, got %d", len(commits))
	}
	if commits[0].Subject != "revert: restore startup probe" {
		t.Fatalf("expected most recent commit first, got %q", commits[0].Subject)
	}
	if len(commits[1].Files) != 2 {
		t.Fatalf("expected files from --name-only parse, got %v", commits[1].Files)
	}

	signals := DeriveSignals(commits)
	if len(signals.HotfixCommits) != 1 || signals.HotfixCommits[0].Subject != "hotfix: revert health timeout tweak" {
		t.Fatalf("unexpected hotfix signals %#v", signals.HotfixCommits)
	}
	if len(signals.RevertCommits) != 2 {
		t.Fatalf("expected two revert-like commits, got %d", len(signals.RevertCommits))
	}
	if len(signals.RepeatedFiles) == 0 || signals.RepeatedFiles[0].File != "deploy/healthcheck.yaml" {
		t.Fatalf("expected repeated file signal for deploy/healthcheck.yaml, got %#v", signals.RepeatedFiles)
	}
}

func TestCorrelateHealthcheckFailure(t *testing.T) {
	t.Parallel()

	commits := []Commit{
		{
			Hash:    "1234567890",
			Subject: "hotfix: tune startup delay",
			Files:   []string{"deploy/healthcheck.yaml", "Dockerfile"},
		},
		{
			Hash:    "abcdef1234",
			Subject: "feat: add /healthz endpoint",
			Files:   []string{"cmd/server/main.go", "internal/http/health.go"},
		},
		{
			Hash:    "aaaaaa1111",
			Subject: "revert: restore previous probe config",
			Files:   []string{"deploy/healthcheck.yaml"},
		},
	}
	signals := DeriveSignals(commits)

	ctx := Correlate("/repo", "deploy", "health-check-failure", commits, signals)
	if ctx.RepoRoot != "/repo" {
		t.Fatalf("unexpected repo root %q", ctx.RepoRoot)
	}
	if len(ctx.RecentFiles) == 0 {
		t.Fatal("expected recent files")
	}
	if !contains(ctx.RecentFiles, "deploy/healthcheck.yaml") {
		t.Fatalf("expected healthcheck file in recent files, got %v", ctx.RecentFiles)
	}
	if len(ctx.RelatedCommits) == 0 || ctx.RelatedCommits[0].Hash != "1234567" {
		t.Fatalf("expected truncated related commit hash, got %#v", ctx.RelatedCommits)
	}
	if len(ctx.HotfixSignals) == 0 || !strings.Contains(ctx.HotfixSignals[0], "hotfix") {
		t.Fatalf("expected hotfix signal, got %v", ctx.HotfixSignals)
	}
	if len(ctx.DriftSignals) == 0 {
		t.Fatalf("expected drift signals, got %v", ctx.DriftSignals)
	}
}

func initTempRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	runGit(t, dir, nil, "init")
	runGit(t, dir, nil, "config", "user.name", "Faultline Test")
	runGit(t, dir, nil, "config", "user.email", "faultline@example.com")
	return dir
}

func writeCommits(t *testing.T, repoDir string, commits []commitSpec) {
	t.Helper()

	for _, commit := range commits {
		for path, content := range commit.Files {
			fullPath := filepath.Join(repoDir, path)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", filepath.Dir(fullPath), err)
			}
			if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
				t.Fatalf("write %s: %v", fullPath, err)
			}
		}
		runGit(t, repoDir, nil, "add", ".")
		env := []string{
			"GIT_AUTHOR_DATE=" + commit.Date,
			"GIT_COMMITTER_DATE=" + commit.Date,
		}
		runGit(t, repoDir, env, "commit", "--quiet", "-m", commit.Subject)
	}
}

func runGit(t *testing.T, dir string, env []string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func TestMatchesAnyBasicGlob(t *testing.T) {
	patterns := []string{"Dockerfile", "*.yaml", ".github*"}
	tests := []struct {
		file string
		want bool
	}{
		{"Dockerfile", true},
		{"config.yaml", true},
		{"deploy.yml", false},
		{".github/workflows/ci.yml", true},
		{"main.go", false},
		{"deep/path/config.yaml", true},
	}
	for _, tt := range tests {
		got := matchesAny(tt.file, patterns)
		if got != tt.want {
			t.Errorf("matchesAny(%q, %v) = %v, want %v", tt.file, patterns, got, tt.want)
		}
	}
}

func TestMatchesAnyNoPatterns(t *testing.T) {
	if matchesAny("Dockerfile", nil) {
		t.Error("matchesAny with no patterns should return false")
	}
}

func TestFailurePatternsKnownPlaybookIDs(t *testing.T) {
	knownCases := []struct {
		playbookID string
		wantFile   string
	}{
		{"docker-auth", "Dockerfile"},
		{"git-auth", ".gitconfig"},
		{"go-sum-missing", "go.sum"},
		{"npm-ci-lockfile", "package.json"},
		{"github-actions-syntax", ".github/workflows/ci.yml"},
		{"terraform-state-lock", "main.tf"},
	}
	for _, tt := range knownCases {
		patterns := failurePatterns(tt.playbookID, tt.playbookID)
		if !matchesAny(tt.wantFile, patterns) {
			t.Errorf("failurePatterns(%q) should match %q, patterns: %v", tt.playbookID, tt.wantFile, patterns)
		}
	}
}

func TestFailurePatternsUnknownFallsBackToCategory(t *testing.T) {
	patterns := failurePatterns("auth", "some-unknown-auth-failure")
	if len(patterns) == 0 {
		t.Error("expected non-empty patterns for auth category fallback")
	}
}

func TestFailurePatternsUnknownCategoryAndID(t *testing.T) {
	patterns := failurePatterns("completely-unknown", "no-such-id")
	if len(patterns) == 0 {
		t.Error("expected default patterns for completely unknown category/id")
	}
}

func TestCorrelateReturnsEmptyContextNoCommits(t *testing.T) {
	t.Parallel()

	ctx := Correlate("/repo", "build", "go-sum-missing", nil, Signals{})
	if ctx.RepoRoot != "/repo" {
		t.Errorf("expected /repo, got %q", ctx.RepoRoot)
	}
	if len(ctx.RecentFiles) != 0 {
		t.Errorf("expected no recent files with no commits, got %v", ctx.RecentFiles)
	}
}

func TestCorrelateUsesHotspotFallback(t *testing.T) {
	t.Parallel()

	sigs := Signals{
		HotspotFiles: []FileChurn{
			{File: "go.sum", Count: 5},
			{File: "go.mod", Count: 3},
		},
	}
	ctx := Correlate("/repo", "build", "go-sum-missing", nil, sigs)
	if len(ctx.RecentFiles) == 0 {
		t.Error("expected fallback to hotspot files when no matching commits")
	}
}

func TestDriftSignalsReturnsRevertSubjects(t *testing.T) {
	t.Parallel()

	commits := []Commit{
		{Hash: "aaa", Subject: "revert: bad deploy", Files: []string{"deploy/config.yaml"}},
	}
	sigs := DeriveSignals(commits)
	spec := correlationSpec{Patterns: failurePatterns("deploy", "argocd-sync-failed")}
	hints := driftSignals(spec, commits, sigs)
	if len(hints) == 0 {
		t.Error("expected drift signal for matching revert commit")
	}
}

func TestHotfixSignalsNoMatchFallsBack(t *testing.T) {
	t.Parallel()

	commits := []Commit{
		{Hash: "bbb", Subject: "hotfix: reduce timeout", Files: []string{"unrelated_file.rb"}},
	}
	sigs := DeriveSignals(commits)
	spec := correlationSpec{Patterns: failurePatterns("deploy", "ecs-deployment-failed")}
	hints := hotfixSignals(spec, commits, sigs)
	// Even without file match, hotfix commits that exist should eventually
	// appear in the fallback.
	if len(hints) == 0 {
		t.Error("expected at least one hotfix signal in fallback")
	}
}
