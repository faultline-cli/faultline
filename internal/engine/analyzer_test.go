package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"faultline/internal/detectors"
	"faultline/internal/model"
	"faultline/internal/repo"
	"faultline/internal/scoring"
)

func TestAnalyzeReaderEmptyInput(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	_, err := e.AnalyzeReader(strings.NewReader(""))
	if err != ErrNoInput {
		t.Fatalf("expected ErrNoInput, got %v", err)
	}
}

func TestAnalyzeReaderPartialMatch(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	// Log with both matching and non-matching content
	log := "some normal log output\nfatal: could not read Username for 'https://github.com': terminal prompts disabled\nmore normal output\n"
	a, err := e.AnalyzeReader(strings.NewReader(log))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(a.Results) == 0 {
		t.Fatal("expected at least one result for partial match")
	}
	if a.Results[0].Playbook.ID != "git-auth" {
		t.Fatalf("expected git-auth, got %s", a.Results[0].Playbook.ID)
	}
}

func TestAnalyzeReaderMultipleMatchesWithDifferentScores(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	// Log with multiple potential matches
	log := "pull access denied\nauthentication required\nfatal: could not read Username for 'https://github.com': terminal prompts disabled\n"
	a, err := e.AnalyzeReader(strings.NewReader(log))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(a.Results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(a.Results))
	}
	
	// Results should be sorted by score (highest first)
	if a.Results[0].Score < a.Results[1].Score {
		t.Errorf("results not sorted by score: %v < %v", a.Results[0].Score, a.Results[1].Score)
	}
}

func TestAnalyzeReaderWithVeryLongLine(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	// Create a very long line that might cause issues
	longLine := strings.Repeat("x", 10000) + "\nfatal: could not read Username for 'https://github.com': terminal prompts disabled\n"
	a, err := e.AnalyzeReader(strings.NewReader(longLine))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(a.Results) == 0 {
		t.Fatal("expected at least one result for long line input")
	}
}

func TestAnalyzeReaderWithUnicodeCharacters(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	// Log with unicode characters
	log := "normal log\nföö: authentication failed for repository 'https://github.com/repo'\nmore log\n"
	a, err := e.AnalyzeReader(strings.NewReader(log))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	// Should not crash, but may not match anything
	if a != nil {
		_ = a
	}
}

func TestAnalyzeReaderWithMixedLineEndings(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	// Log with mixed line endings (CRLF, LF, CR)
	log := "line one\r\nline two\nline three\rline four\nfatal: could not read Username for 'https://github.com': terminal prompts disabled\n"
	a, err := e.AnalyzeReader(strings.NewReader(log))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(a.Results) == 0 {
		t.Fatal("expected at least one result for mixed line endings")
	}
}

func TestAnalyzeReaderWithOnlyWhitespace(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	// Log with only whitespace
	log := "   \n\t\n  \n\n"
	_, err := e.AnalyzeReader(strings.NewReader(log))
	if err != ErrNoInput {
		t.Fatalf("expected ErrNoInput for whitespace-only input, got %v", err)
	}
}

func TestAnalyzeReaderWithBinaryData(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	// Log with some binary-like data
	log := "normal log\n\x00\x01\x02\x03fatal: could not read Username for 'https://github.com': terminal prompts disabled\n\xff\xfe\xfd\xfc\n"
	a, err := e.AnalyzeReader(strings.NewReader(log))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(a.Results) == 0 {
		t.Fatal("expected at least one result for input with binary data")
	}
}

func TestAnalyzeRepositoryWithEmptyChangeSet(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	dir := t.TempDir()
	
	// Create a simple Go file
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	
	a, err := e.AnalyzeRepository(dir, detectors.ChangeSet{})
	if err != nil && err != ErrNoMatch {
		t.Fatalf("unexpected error: %v", err)
	}
	// Either ErrNoMatch or a valid analysis is acceptable
	_ = a
}

func TestAnalyzeRepositoryWithNonExistentRoot(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	
	_, err := e.AnalyzeRepository("/nonexistent/path", detectors.ChangeSet{})
	if err == nil {
		t.Fatal("expected error for non-existent root")
	}
}

func TestAnalyzeRepositoryWithCircularSymlink(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	dir := t.TempDir()
	
	// Create a circular symlink
	if err := os.Symlink(dir, filepath.Join(dir, "circular")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	
	// This should not crash, but may return an error or empty result
	_, err := e.AnalyzeRepository(dir, detectors.ChangeSet{})
	if err != nil && err != ErrNoMatch && err != ErrNoInput {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeRepositoryWithPermissionDeniedFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}
	
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	dir := t.TempDir()
	
	// Create a file with restricted permissions
	protectedFile := filepath.Join(dir, "protected.go")
	if err := os.WriteFile(protectedFile, []byte("package main\n"), 0o000); err != nil {
		t.Fatalf("write protected file: %v", err)
	}
	defer os.Chmod(protectedFile, 0o644) // cleanup
	
	// This should not crash, but may return an error due to permission denied
	_, err := e.AnalyzeRepository(dir, detectors.ChangeSet{})
	if err != nil && err != ErrNoMatch && err != ErrNoInput {
		// Permission denied is acceptable
		t.Logf("got expected permission error: %v", err)
	}
}

func TestCorrelateSnapshotWithNilResult(t *testing.T) {
	snap := &repoSnapshot{
		root: "/repo",
		commits: []repo.Commit{
			{Hash: "abc1234", Subject: "fix: auth token", Files: []string{".gitconfig"}, Time: time.Now()},
		},
		signals: repo.Signals{},
		state:   &scoring.RepoState{Root: "/repo"},
	}
	rc := correlateSnapshot(snap, model.Result{})
	// correlateSnapshot returns a RepoContext even for empty results
	if rc == nil {
		t.Errorf("expected non-nil RepoContext")
	}
	if rc.RepoRoot != "/repo" {
		t.Errorf("expected RepoRoot /repo, got %q", rc.RepoRoot)
	}
}

func TestLoadTopologySignalsWithInvalidCODEOWNERS(t *testing.T) {
	dir := t.TempDir()
	// Create an invalid CODEOWNERS file
	if err := os.WriteFile(filepath.Join(dir, "CODEOWNERS"), []byte("invalid format"), 0o644); err != nil {
		t.Fatalf("write CODEOWNERS: %v", err)
	}
	
	out := loadTopologySignals(dir, []string{"main.go"})
	if out != nil {
		t.Errorf("expected nil with invalid CODEOWNERS, got %v", out)
	}
}

func TestLoadTopologySignalsWithChangedFilesOutsideRepo(t *testing.T) {
	dir := t.TempDir()
	// Create a valid CODEOWNERS file
	if err := os.WriteFile(filepath.Join(dir, "CODEOWNERS"), []byte("* @team\n"), 0o644); err != nil {
		t.Fatalf("write CODEOWNERS: %v", err)
	}
	
	// Include a file path outside the repo root
	out := loadTopologySignals(dir, []string{"../outside/main.go"})
	if out != nil {
		t.Errorf("expected nil with outside files, got %v", out)
	}
}

func TestRepoStateFromSnapshotWithNil(t *testing.T) {
	state := repoStateFromSnapshot(nil)
	if state != nil {
		t.Errorf("expected nil for nil snapshot, got %#v", state)
	}
}

func TestMergeRepoStatesWithNil(t *testing.T) {
	state1 := &scoring.RepoState{Root: "/repo1", ChangedFiles: []string{"file1.go"}}
	state2 := &scoring.RepoState{Root: "/repo2", ChangedFiles: []string{"file2.go"}}
	
	merged := mergeRepoStates(state1, state2, nil)
	if len(merged.ChangedFiles) != 2 {
		t.Errorf("expected 2 changed files, got %d", len(merged.ChangedFiles))
	}
}

func TestChangedFilesFromChangeSetWithEmpty(t *testing.T) {
	files := changedFilesFromChangeSet(detectors.ChangeSet{})
	if files != nil {
		t.Errorf("expected nil for empty change set, got %v", files)
	}
}

func TestRecentFilesFromCommitsWithEmpty(t *testing.T) {
	files := recentFilesFromCommits([]repo.Commit{}, 10)
	if files != nil {
		t.Errorf("expected nil for empty commits, got %v", files)
	}
}

func TestHotspotDirsFromSignalsWithEmpty(t *testing.T) {
	dirs := hotspotDirsFromSignals(repo.Signals{}, 5)
	if dirs != nil {
		t.Errorf("expected nil for empty signals, got %v", dirs)
	}
}

func TestHotfixSignalsFromSignalsWithEmpty(t *testing.T) {
	signals := hotfixSignalsFromSignals(repo.Signals{}, 5)
	if signals != nil {
		t.Errorf("expected nil for empty signals, got %v", signals)
	}
}

func TestDriftSignalsFromSignalsWithEmpty(t *testing.T) {
	signals := driftSignalsFromSignals(repo.Signals{}, 5)
	if signals != nil {
		t.Errorf("expected nil for empty signals, got %v", signals)
	}
}

func TestLooksLikeSourceFileWithUnknownExtension(t *testing.T) {
	if looksLikeSourceFile("unknown.xyz") {
		t.Error("expected false for unknown extension")
	}
}

func TestShouldSkipSourceDirWithUnknownDir(t *testing.T) {
	if shouldSkipSourceDir("unknown_dir") {
		t.Error("expected false for unknown directory")
	}
}

func TestReadLinesWithVeryLongLine(t *testing.T) {
	longLine := strings.Repeat("x", 10000) + "\n"
	lines, err := ReadLines(strings.NewReader(longLine))
	if err != nil {
		t.Fatalf("ReadLines: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Original != longLine[:len(longLine)-1] {
		t.Errorf("line content mismatch")
	}
}

func repoPlaybookDir(_ testing.TB) string {
	return "../../playbooks/bundled"
}

// ── AnalyzePath ───────────────────────────────────────────────────────────────

func TestAnalyzePathMatchingFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "ci.log")
	content := "fatal: could not read Username for 'https://github.com': terminal prompts disabled\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0o600); err != nil {
		t.Fatalf("write log file: %v", err)
	}
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	a, err := e.AnalyzePath(tmpFile)
	if err != nil {
		t.Fatalf("AnalyzePath: %v", err)
	}
	if a == nil || len(a.Results) == 0 {
		t.Fatal("expected at least one result")
	}
	if a.Source != tmpFile {
		t.Fatalf("expected source to be set to path, got %q", a.Source)
	}
}

func TestAnalyzePathMissingFileErrors(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	_, err := e.AnalyzePath(filepath.Join(t.TempDir(), "no_such_file.log"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// ── loadRepoSnapshotFromPath / loadRepoSnapshot ───────────────────────────────

func TestLoadRepoSnapshotFromPathNonGitDir(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	// A non-git temp dir: NewScanner will fail; should return a snapshot with nil state.
	snap := e.loadRepoSnapshotFromPath(t.TempDir(), detectors.ChangeSet{})
	if snap == nil {
		t.Fatal("expected non-nil snapshot even for non-git directory")
	}
}

func TestLoadRepoSnapshotNonGitDir(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true, RepoPath: t.TempDir()})
	snap := e.loadRepoSnapshot()
	if snap == nil {
		t.Fatal("expected non-nil snapshot from loadRepoSnapshot")
	}
}

// ── localRepoEnricher.Enrich ──────────────────────────────────────────────────

func TestLocalRepoEnricherEnrichNonGitReturnsNil(t *testing.T) {
	repoPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoPath, ".git"), []byte("not a git dir"), 0o644); err != nil {
		t.Fatalf("write invalid .git marker: %v", err)
	}

	enricher := localRepoEnricher{opts: Options{RepoPath: repoPath}}
	rc := enricher.Enrich(model.Result{
		Playbook: model.Playbook{ID: "docker-auth", Category: "auth"},
	})
	if rc != nil {
		t.Fatalf("expected nil RepoContext for non-git dir, got %#v", rc)
	}
}

// ── List and Explain ──────────────────────────────────────────────────────────

func TestListReturnsBundledPlaybooks(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	pbs, err := e.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(pbs) == 0 {
		t.Fatal("expected at least one playbook from bundled directory")
	}
}

func TestExplainReturnsPlaybook(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	pbs, err := e.List()
	if err != nil || len(pbs) == 0 {
		t.Fatalf("List: %v (len=%d)", err, len(pbs))
	}
	pb, err := e.Explain(pbs[0].ID)
	if err != nil {
		t.Fatalf("Explain(%q): %v", pbs[0].ID, err)
	}
	if pb.ID != pbs[0].ID {
		t.Fatalf("expected ID %q, got %q", pbs[0].ID, pb.ID)
	}
}

func TestExplainUnknownIDReturnsError(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	_, err := e.Explain("definitely-not-a-real-playbook-id-xyz")
	if err == nil {
		t.Fatal("expected error for unknown playbook ID")
	}
}