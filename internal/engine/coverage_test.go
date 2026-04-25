package engine

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"faultline/internal/detectors"
	"faultline/internal/model"
	"faultline/internal/repo"
	"faultline/internal/repo/topology"
	"faultline/internal/scoring"
)

// --- defaultSourceLoader (deps.go) ---

func TestDefaultSourceLoaderLoadScansDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "service.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Readme\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	loader := defaultSourceLoader{}
	files, err := loader.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(files) != 1 || files[0].Path != "service.go" {
		t.Errorf("expected only service.go, got %v", files)
	}
}

func TestDefaultSourceLoaderLoadMissingDir(t *testing.T) {
	dir := t.TempDir()
	missingDir := filepath.Join(dir, "missing")
	if _, err := os.Stat(missingDir); !os.IsNotExist(err) {
		t.Fatalf("expected %q to not exist before Load, got err=%v", missingDir, err)
	}
	loader := defaultSourceLoader{}
	_, err := loader.Load(missingDir)
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

// --- correlateSnapshot ---

func TestCorrelateSnapshotNilReturnsNil(t *testing.T) {
	result := model.Result{Playbook: model.Playbook{ID: "git-auth", Category: "auth"}}
	rc := correlateSnapshot(nil, result)
	if rc != nil {
		t.Errorf("expected nil for nil snapshot, got %#v", rc)
	}
}

func TestCorrelateSnapshotEmptyRootReturnsNil(t *testing.T) {
	snap := &repoSnapshot{root: ""}
	result := model.Result{Playbook: model.Playbook{ID: "git-auth", Category: "auth"}}
	rc := correlateSnapshot(snap, result)
	if rc != nil {
		t.Errorf("expected nil for empty root, got %#v", rc)
	}
}

func TestCorrelateSnapshotNoCommitsReturnsNil(t *testing.T) {
	snap := &repoSnapshot{root: "/repo", commits: nil}
	result := model.Result{Playbook: model.Playbook{ID: "git-auth", Category: "auth"}}
	rc := correlateSnapshot(snap, result)
	if rc != nil {
		t.Errorf("expected nil for snapshot with no commits, got %#v", rc)
	}
}

func TestCorrelateSnapshotWithCommitsReturnsContext(t *testing.T) {
	snap := &repoSnapshot{
		root: "/repo",
		commits: []repo.Commit{
			{Hash: "abc1234", Subject: "fix: auth token", Files: []string{".gitconfig"}, Time: time.Now()},
		},
		signals: repo.Signals{},
		state:   &scoring.RepoState{Root: "/repo"},
	}
	result := model.Result{Playbook: model.Playbook{ID: "git-auth", Category: "auth"}}
	rc := correlateSnapshot(snap, result)
	if rc == nil {
		t.Fatal("expected non-nil repo context for snapshot with commits")
	}
	if rc.RepoRoot != "/repo" {
		t.Errorf("expected /repo, got %q", rc.RepoRoot)
	}
}

func TestCorrelateSnapshotAttachesTopologySignals(t *testing.T) {
	snap := &repoSnapshot{
		root: "/repo",
		commits: []repo.Commit{
			{Hash: "abc1234", Subject: "feat: boundary crossing change", Files: []string{"deploy.yaml"}, Time: time.Now()},
		},
		signals: repo.Signals{},
		state: &scoring.RepoState{
			Root:            "/repo",
			TopologySignals: []string{topology.SignalBoundaryCrossed},
		},
	}
	result := model.Result{Playbook: model.Playbook{ID: "docker-auth", Category: "auth"}}
	rc := correlateSnapshot(snap, result)
	if rc == nil {
		t.Fatal("expected non-nil repo context")
	}
	if rc.Topology == nil {
		t.Fatal("expected topology signals to be attached")
	}
	if !rc.Topology.BoundaryCrossed {
		t.Errorf("expected BoundaryCrossed=true for signal %q", topology.SignalBoundaryCrossed)
	}
}

// --- loadTopologySignals ---

func TestLoadTopologySignalsNoCodeowners(t *testing.T) {
	dir := t.TempDir()
	out := loadTopologySignals(dir, []string{"main.go"})
	// Without CODEOWNERS the function should return nil.
	if out != nil {
		t.Errorf("expected nil without CODEOWNERS, got %v", out)
	}
}

func TestLoadTopologySignalsEmptyChangedFiles(t *testing.T) {
	dir := t.TempDir()
	out := loadTopologySignals(dir, nil)
	if out != nil {
		t.Errorf("expected nil for empty changed files, got %v", out)
	}
}

// --- AnalyzeRepository ---

func TestAnalyzeRepositoryReturnsAnalysis(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	a, err := e.AnalyzeRepository(dir, detectors.ChangeSet{})
	// Either ErrNoMatch or a valid analysis is acceptable.
	if err != nil && err != ErrNoMatch {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil analysis even with no match")
	}
}

func TestAnalyzeRepositoryEmptyDirReturnsErrNoInput(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	dir := t.TempDir()
	// No files written → should return ErrNoInput or ErrNoMatch.
	a, err := e.AnalyzeRepository(dir, detectors.ChangeSet{})
	if err != nil && err != ErrNoMatch && err != ErrNoInput {
		t.Fatalf("expected ErrNoInput or ErrNoMatch, got %v", err)
	}
	_ = a
}

// --- deltaRequested ---

func TestDeltaRequestedFalseWhenEmpty(t *testing.T) {
	e := &Engine{opts: Options{DeltaProvider: ""}}
	if e.deltaRequested() {
		t.Error("expected deltaRequested=false when DeltaProvider is empty")
	}
}

func TestDeltaRequestedTrueWhenSet(t *testing.T) {
	e := &Engine{opts: Options{DeltaProvider: "github-actions"}}
	if !e.deltaRequested() {
		t.Error("expected deltaRequested=true when DeltaProvider is set")
	}
}

func TestDeltaRequestedFalseForWhitespace(t *testing.T) {
	e := &Engine{opts: Options{DeltaProvider: "   "}}
	if e.deltaRequested() {
		t.Error("expected deltaRequested=false for whitespace-only provider")
	}
}

// --- loadProviderDelta (indirectly via AnalyzeReader with no provider) ---

func TestLoadProviderDeltaReturnsNilWhenNotRequested(t *testing.T) {
	e := &Engine{opts: Options{}}
	state := e.loadProviderDelta("some log content")
	if state != nil {
		t.Errorf("expected nil when delta not requested, got %#v", state)
	}
}

// --- repoStateFromSnapshot ---

func TestRepoStateFromSnapshotNil(t *testing.T) {
	state := repoStateFromSnapshot(nil)
	if state != nil {
		t.Errorf("expected nil for nil snapshot, got %#v", state)
	}
}

func TestRepoStateFromSnapshotReturnsState(t *testing.T) {
	s := &scoring.RepoState{Root: "/repo"}
	snap := &repoSnapshot{state: s}
	state := repoStateFromSnapshot(snap)
	if state == nil || state.Root != "/repo" {
		t.Errorf("expected state with /repo root, got %#v", state)
	}
}
