package repo

import (
	"testing"
)

// --------------------------------------------------------------------------
// isConfigFile
// --------------------------------------------------------------------------

func TestIsConfigFile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		file string
		want bool
	}{
		{"go.mod", "go.mod", true},
		{"go.sum", "go.sum", true},
		{"nested go.mod", "submodule/go.mod", true},
		{"package.json", "package.json", true},
		{"package-lock.json", "package-lock.json", true},
		{"yarn.lock", "yarn.lock", true},
		{"pnpm-lock.yaml", "pnpm-lock.yaml", true},
		{"Dockerfile root", "Dockerfile", true},
		{"Dockerfile nested", "services/api/Dockerfile", true},
		{"pyproject.toml", "pyproject.toml", true},
		{"Gemfile", "Gemfile", true},
		{"Gemfile.lock", "Gemfile.lock", true},
		{"Cargo.toml", "Cargo.toml", true},
		{"Cargo.lock", "Cargo.lock", true},
		{"pom.xml", "pom.xml", true},
		{"build.gradle", "build.gradle", true},
		{"build.gradle.kts", "build.gradle.kts", true},
		{"requirements.txt", "requirements.txt", true},
		{"requirements-dev.txt", "requirements-dev.txt", true},
		{"docker-compose.yml", "docker-compose.yml", true},
		{"docker-compose.yaml", "docker-compose.yaml", true},
		{"docker-compose.override.yml", "docker-compose.override.yml", true},
		{".env file", ".env", true},
		{".env.local", ".env.local", true},
		// Not config files.
		{"main.go", "main.go", false},
		{"handler.ts", "handler.ts", false},
		{"README.md", "README.md", false},
		{"schema.sql", "schema.sql", false},
		{"config.yaml", "config.yaml", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isConfigFile(tc.file)
			if got != tc.want {
				t.Errorf("isConfigFile(%q) = %v, want %v", tc.file, got, tc.want)
			}
		})
	}
}

// --------------------------------------------------------------------------
// isCIConfigFile
// --------------------------------------------------------------------------

func TestIsCIConfigFile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		file string
		want bool
	}{
		{"Makefile", "Makefile", true},
		{"Jenkinsfile", "Jenkinsfile", true},
		{"azure-pipelines.yml", "azure-pipelines.yml", true},
		{"bitbucket-pipelines.yml", "bitbucket-pipelines.yml", true},
		{".gitlab-ci.yml", ".gitlab-ci.yml", true},
		{".github workflow", ".github/workflows/ci.yml", true},
		{".github action", ".github/actions/setup/action.yml", true},
		{".circleci config", ".circleci/config.yml", true},
		{".gitlab dir", ".gitlab/ci/build.yml", true},
		// Not CI config files.
		{"main.go", "main.go", false},
		{"package.json", "package.json", false},
		{"deploy.yaml", "deploy/deploy.yaml", false},
		{"Dockerfile", "Dockerfile", false},
		{"README.md", "README.md", false},
		{"scripts/deploy.sh", "scripts/deploy.sh", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isCIConfigFile(tc.file)
			if got != tc.want {
				t.Errorf("isCIConfigFile(%q) = %v, want %v", tc.file, got, tc.want)
			}
		})
	}
}

// --------------------------------------------------------------------------
// DeriveSignals – large commits
// --------------------------------------------------------------------------

func TestDeriveSignals_largeCommit(t *testing.T) {
	t.Parallel()

	// Build a commit with LargeCommitFileThreshold files.
	largeFiles := make([]string, LargeCommitFileThreshold)
	for i := range largeFiles {
		largeFiles[i] = "src/file" + string(rune('a'+i)) + ".go"
	}
	commits := []Commit{
		{Hash: "aaa1111", Subject: "chore: big refactor", Files: largeFiles},
		{Hash: "bbb2222", Subject: "fix: small patch", Files: []string{"main.go"}},
	}
	sigs := DeriveSignals(commits)

	if len(sigs.LargeCommits) != 1 {
		t.Fatalf("want 1 large commit, got %d", len(sigs.LargeCommits))
	}
	if sigs.LargeCommits[0].Subject != "chore: big refactor" {
		t.Errorf("unexpected large commit: %q", sigs.LargeCommits[0].Subject)
	}
}

func TestDeriveSignals_noLargeCommit(t *testing.T) {
	t.Parallel()

	commits := []Commit{
		{Hash: "aaa1", Subject: "fix: trivial", Files: []string{"main.go", "go.sum"}},
	}
	sigs := DeriveSignals(commits)
	if len(sigs.LargeCommits) != 0 {
		t.Errorf("want no large commits, got %v", sigs.LargeCommits)
	}
}

// --------------------------------------------------------------------------
// DeriveSignals – config / CI file signals
// --------------------------------------------------------------------------

func TestDeriveSignals_configChangedFiles(t *testing.T) {
	t.Parallel()

	commits := []Commit{
		{Hash: "aaa", Subject: "chore: bump deps", Files: []string{"go.mod", "go.sum", "internal/app.go"}},
		{Hash: "bbb", Subject: "feat: add service", Files: []string{"go.mod", "cmd/main.go"}},
	}
	sigs := DeriveSignals(commits)

	if len(sigs.ConfigChangedFiles) == 0 {
		t.Fatal("want config changed files, got none")
	}
	// go.mod appears in 2 commits so it should be the top entry.
	if sigs.ConfigChangedFiles[0].File != "go.mod" {
		t.Errorf("want go.mod as top config file, got %q", sigs.ConfigChangedFiles[0].File)
	}
	if sigs.ConfigChangedFiles[0].Count != 2 {
		t.Errorf("want count=2 for go.mod, got %d", sigs.ConfigChangedFiles[0].Count)
	}
	// Verify go.sum is also present.
	found := false
	for _, fc := range sigs.ConfigChangedFiles {
		if fc.File == "go.sum" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected go.sum in ConfigChangedFiles, got %v", sigs.ConfigChangedFiles)
	}
}

func TestDeriveSignals_ciConfigChangedFiles(t *testing.T) {
	t.Parallel()

	commits := []Commit{
		{Hash: "ccc", Subject: "ci: add lint step", Files: []string{".github/workflows/ci.yml", "Makefile"}},
	}
	sigs := DeriveSignals(commits)

	if len(sigs.CIConfigChangedFiles) == 0 {
		t.Fatal("want CI config changed files, got none")
	}
	paths := make(map[string]bool)
	for _, fc := range sigs.CIConfigChangedFiles {
		paths[fc.File] = true
	}
	if !paths[".github/workflows/ci.yml"] {
		t.Errorf("expected .github/workflows/ci.yml in CIConfigChangedFiles, got %v", sigs.CIConfigChangedFiles)
	}
	if !paths["Makefile"] {
		t.Errorf("expected Makefile in CIConfigChangedFiles, got %v", sigs.CIConfigChangedFiles)
	}
}

func TestDeriveSignals_noConfigOrCIFiles(t *testing.T) {
	t.Parallel()

	commits := []Commit{
		{Hash: "ddd", Subject: "feat: add handler", Files: []string{"internal/handler.go", "internal/handler_test.go"}},
	}
	sigs := DeriveSignals(commits)

	if len(sigs.ConfigChangedFiles) != 0 {
		t.Errorf("want no config files, got %v", sigs.ConfigChangedFiles)
	}
	if len(sigs.CIConfigChangedFiles) != 0 {
		t.Errorf("want no CI config files, got %v", sigs.CIConfigChangedFiles)
	}
}

// --------------------------------------------------------------------------
// DeriveSignals – author diversity
// --------------------------------------------------------------------------

func TestDeriveSignals_authorDiversity(t *testing.T) {
	t.Parallel()

	commits := []Commit{
		{Hash: "e1", Subject: "feat: api", Author: "alice@example.com", Files: []string{"api.go"}},
		{Hash: "e2", Subject: "feat: db", Author: "bob@example.com", Files: []string{"db.go"}},
		{Hash: "e3", Subject: "fix: typo", Author: "alice@example.com", Files: []string{"api.go"}},
	}
	sigs := DeriveSignals(commits)

	if sigs.AuthorCount != 2 {
		t.Errorf("want AuthorCount=2, got %d", sigs.AuthorCount)
	}
	if len(sigs.TopAuthors) == 0 {
		t.Fatal("want TopAuthors populated")
	}
	// alice has 2 commits and should rank first.
	if sigs.TopAuthors[0].Author != "alice@example.com" {
		t.Errorf("want alice@example.com as top author, got %q", sigs.TopAuthors[0].Author)
	}
	if sigs.TopAuthors[0].Count != 2 {
		t.Errorf("want count=2 for alice, got %d", sigs.TopAuthors[0].Count)
	}
}

func TestDeriveSignals_authorCountEmpty(t *testing.T) {
	t.Parallel()

	// Commits with no Author field set.
	commits := []Commit{
		{Hash: "f1", Subject: "add file", Files: []string{"foo.go"}},
	}
	sigs := DeriveSignals(commits)

	if sigs.AuthorCount != 0 {
		t.Errorf("want AuthorCount=0 for commits with no author, got %d", sigs.AuthorCount)
	}
	if len(sigs.TopAuthors) != 0 {
		t.Errorf("want empty TopAuthors, got %v", sigs.TopAuthors)
	}
}

func TestDeriveSignals_topAuthorsLimit(t *testing.T) {
	t.Parallel()

	// Create commits from 7 distinct authors (above maxTopAuthors=5).
	authors := []string{
		"a@x.com", "b@x.com", "c@x.com", "d@x.com", "e@x.com", "f@x.com", "g@x.com",
	}
	commits := make([]Commit, len(authors))
	for i, a := range authors {
		commits[i] = Commit{Hash: "h" + string(rune('0'+i)), Subject: "change", Author: a, Files: []string{"file.go"}}
	}
	sigs := DeriveSignals(commits)

	if sigs.AuthorCount != len(authors) {
		t.Errorf("want AuthorCount=%d, got %d", len(authors), sigs.AuthorCount)
	}
	if len(sigs.TopAuthors) != maxTopAuthors {
		t.Errorf("want TopAuthors capped at %d, got %d", maxTopAuthors, len(sigs.TopAuthors))
	}
}

// --------------------------------------------------------------------------
// DeriveSignals – stable ordering
// --------------------------------------------------------------------------

func TestDeriveSignals_configFileOrdering(t *testing.T) {
	t.Parallel()

	// Files with same edit count should be ordered alphabetically.
	commits := []Commit{
		{Hash: "g1", Subject: "ci: renovate", Files: []string{"package.json", "go.mod"}},
	}
	sigs := DeriveSignals(commits)

	if len(sigs.ConfigChangedFiles) < 2 {
		t.Skip("fewer than 2 config files; skipping order check")
	}
	// Both have count=1; go.mod < package.json alphabetically.
	if sigs.ConfigChangedFiles[0].File >= sigs.ConfigChangedFiles[1].File {
		t.Errorf("want alphabetical order; got %q before %q",
			sigs.ConfigChangedFiles[0].File, sigs.ConfigChangedFiles[1].File)
	}
}

// --------------------------------------------------------------------------
// correlate – new RepoContext fields
// --------------------------------------------------------------------------

func TestCorrelate_newSignalFields(t *testing.T) {
	t.Parallel()

	// Build commits that trigger all three new context fields.
	largeFiles := make([]string, LargeCommitFileThreshold)
	for i := range largeFiles {
		largeFiles[i] = "src/module" + string(rune('a'+i)) + ".go"
	}
	commits := []Commit{
		{
			Hash:    "aaa0001",
			Subject: "chore: mass replace",
			Author:  "dev@example.com",
			Files:   largeFiles,
		},
		{
			Hash:    "bbb0002",
			Subject: "build: update dependencies",
			Author:  "bot@example.com",
			Files:   []string{"go.mod", "go.sum", ".github/workflows/ci.yml"},
		},
	}

	sigs := DeriveSignals(commits)
	ctx := Correlate("/workspace", "build", "runtime-mismatch", commits, sigs)

	// Config drift signals should include go.mod and/or go.sum.
	foundConfig := false
	for _, s := range ctx.ConfigDriftSignals {
		if s == "go.mod" || s == "go.sum" {
			foundConfig = true
		}
	}
	if !foundConfig {
		t.Errorf("want go.mod or go.sum in ConfigDriftSignals, got %v", ctx.ConfigDriftSignals)
	}

	// CI change signals should include the workflow file.
	foundCI := false
	for _, s := range ctx.CIChangeSignals {
		if s == ".github/workflows/ci.yml" {
			foundCI = true
		}
	}
	if !foundCI {
		t.Errorf("want .github/workflows/ci.yml in CIChangeSignals, got %v", ctx.CIChangeSignals)
	}

	// Large commit signals should name the large commit.
	if len(ctx.LargeCommitSignals) == 0 {
		t.Errorf("want LargeCommitSignals populated, got none")
	}
	if ctx.LargeCommitSignals[0] == "" {
		t.Error("large commit signal should not be empty")
	}
}

func TestCorrelate_emptySignalsNoNewFields(t *testing.T) {
	t.Parallel()

	// Commits with only plain source files – no config, CI or large commits.
	commits := []Commit{
		{Hash: "zzz", Subject: "fix: handler", Files: []string{"internal/handler.go"}},
	}
	sigs := DeriveSignals(commits)
	ctx := Correlate("/workspace", "build", "unknown", commits, sigs)

	if len(ctx.ConfigDriftSignals) != 0 {
		t.Errorf("want no ConfigDriftSignals, got %v", ctx.ConfigDriftSignals)
	}
	if len(ctx.CIChangeSignals) != 0 {
		t.Errorf("want no CIChangeSignals, got %v", ctx.CIChangeSignals)
	}
	if len(ctx.LargeCommitSignals) != 0 {
		t.Errorf("want no LargeCommitSignals, got %v", ctx.LargeCommitSignals)
	}
}

// --------------------------------------------------------------------------
// parseCommitBlock – author field
// --------------------------------------------------------------------------

func TestParseCommitBlock_withAuthor(t *testing.T) {
	t.Parallel()

	block := "deadbeef\x1e1712668800\x1echore: update deps\x1ebot@example.com\ngo.mod\ngo.sum\n"
	commit, ok := parseCommitBlock(block)
	if !ok {
		t.Fatal("expected block to parse")
	}
	if commit.Author != "bot@example.com" {
		t.Errorf("want author %q, got %q", "bot@example.com", commit.Author)
	}
	if commit.Subject != "chore: update deps" {
		t.Errorf("unexpected subject %q", commit.Subject)
	}
}

func TestParseCommitBlock_withoutAuthor(t *testing.T) {
	t.Parallel()

	// Old-style block without author field – should still parse successfully.
	block := "abc123\x1e1712668800\x1efix: align deploy healthcheck\nDockerfile\ninfra/deploy.yaml\n"
	commit, ok := parseCommitBlock(block)
	if !ok {
		t.Fatal("expected legacy block to parse")
	}
	if commit.Author != "" {
		t.Errorf("want empty author for legacy block, got %q", commit.Author)
	}
	if len(commit.Files) != 2 {
		t.Fatalf("want 2 files, got %v", commit.Files)
	}
}

// --------------------------------------------------------------------------
// Full integration: LoadHistory captures author from real git repo
// --------------------------------------------------------------------------

func TestLoadHistoryIncludesAuthor(t *testing.T) {
	repoDir := initTempRepo(t)

	writeCommits(t, repoDir, []commitSpec{
		{
			Date:    "2026-04-10T10:00:00Z",
			Subject: "feat: add go module",
			Files:   map[string]string{"go.mod": "module example.com\n"},
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
	if len(commits) == 0 {
		t.Fatal("expected at least one commit")
	}
	// initTempRepo configures user.email = faultline@example.com.
	if commits[0].Author != "faultline@example.com" {
		t.Errorf("want author faultline@example.com, got %q", commits[0].Author)
	}
}

// --------------------------------------------------------------------------
// Integration: DeriveSignals from real git repo
// --------------------------------------------------------------------------

func TestDeriveSignalsFromGitRepo(t *testing.T) {
	repoDir := initTempRepo(t)

	// Create a commit that touches CI + config files.
	writeCommits(t, repoDir, []commitSpec{
		{
			Date:    "2026-04-10T09:00:00Z",
			Subject: "build: update go.mod and CI",
			Files: map[string]string{
				"go.mod":                   "module example.com\ngo 1.22\n",
				".github/workflows/ci.yml": "on: push\n",
				"internal/server.go":       "package main\n",
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

	sigs := DeriveSignals(commits)

	if len(sigs.ConfigChangedFiles) == 0 {
		t.Error("want config changed files detected (go.mod)")
	}
	if len(sigs.CIConfigChangedFiles) == 0 {
		t.Error("want CI config changed files detected (.github/workflows/ci.yml)")
	}
	if sigs.AuthorCount != 1 {
		t.Errorf("want AuthorCount=1, got %d", sigs.AuthorCount)
	}
}

// Ensure DeriveSignals is stable across multiple calls with the same input.
func TestDeriveSignals_deterministicNewFields(t *testing.T) {
	t.Parallel()

	commits := []Commit{
		{Hash: "h1", Subject: "chore: deps", Author: "z@x.com", Files: []string{"go.mod", "go.sum", "package.json"}},
		{Hash: "h2", Subject: "ci: add workflow", Author: "a@x.com", Files: []string{".github/workflows/test.yml"}},
	}

	s1 := DeriveSignals(commits)
	s2 := DeriveSignals(commits)

	if len(s1.ConfigChangedFiles) != len(s2.ConfigChangedFiles) {
		t.Error("ConfigChangedFiles not stable")
	}
	if len(s1.CIConfigChangedFiles) != len(s2.CIConfigChangedFiles) {
		t.Error("CIConfigChangedFiles not stable")
	}
	if s1.AuthorCount != s2.AuthorCount {
		t.Error("AuthorCount not stable")
	}
	if len(s1.TopAuthors) != len(s2.TopAuthors) {
		t.Error("TopAuthors length not stable")
	}
	for i := range s1.TopAuthors {
		if s1.TopAuthors[i] != s2.TopAuthors[i] {
			t.Errorf("TopAuthors[%d] not stable: %+v vs %+v", i, s1.TopAuthors[i], s2.TopAuthors[i])
		}
	}
	for i := range s1.ConfigChangedFiles {
		if s1.ConfigChangedFiles[i] != s2.ConfigChangedFiles[i] {
			t.Errorf("ConfigChangedFiles[%d] not stable", i)
		}
	}
}
