package engine

import (
	"testing"

	"faultline/internal/detectors"
	"faultline/internal/model"
	"faultline/internal/repo"
	"faultline/internal/repo/topology"
	"faultline/internal/scoring"
)

func TestDedupeSortedStringsBasic(t *testing.T) {
	got := dedupeSortedStrings([]string{"b", "a", "b", "  ", "", "c"})
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("dedupeSortedStrings: got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("dedupeSortedStrings[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDedupeSortedStringsNilInput(t *testing.T) {
	got := dedupeSortedStrings(nil)
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestChangedFilesFromChangeSet(t *testing.T) {
	cs := detectors.ChangeSet{
		ChangedFiles: map[string]detectors.FileChange{
			"z.go": {Status: "modified"},
			"a.go": {Status: "added"},
		},
	}
	files := changedFilesFromChangeSet(cs)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(files), files)
	}
	if files[0] != "a.go" || files[1] != "z.go" {
		t.Fatalf("expected sorted [a.go z.go], got %v", files)
	}
}

func TestChangedFilesFromChangeSetEmpty(t *testing.T) {
	if got := changedFilesFromChangeSet(detectors.ChangeSet{}); got != nil {
		t.Fatalf("expected nil for empty changeSet, got %v", got)
	}
}

func TestStateFromChangeSetNilWhenBothEmpty(t *testing.T) {
	if got := stateFromChangeSet("", detectors.ChangeSet{}); got != nil {
		t.Fatalf("expected nil for empty root and empty changeSet, got %#v", got)
	}
}

func TestStateFromChangeSetPopulatesFiles(t *testing.T) {
	cs := detectors.ChangeSet{
		ChangedFiles: map[string]detectors.FileChange{
			"main.go": {Status: "modified"},
		},
	}
	got := stateFromChangeSet("", cs)
	if got == nil {
		t.Fatal("expected non-nil state")
	}
	if len(got.ChangedFiles) != 1 || got.ChangedFiles[0] != "main.go" {
		t.Fatalf("unexpected changed files: %v", got.ChangedFiles)
	}
}

func TestStateFromChangeSetSetsRoot(t *testing.T) {
	got := stateFromChangeSet("/some/repo", detectors.ChangeSet{})
	if got == nil {
		t.Fatal("expected non-nil state when root is set")
	}
	if got.Root != "/some/repo" {
		t.Fatalf("expected root /some/repo, got %q", got.Root)
	}
}

func TestRecentFilesFromCommitsDedupes(t *testing.T) {
	commits := []repo.Commit{
		{Files: []string{"a.go", "b.go"}},
		{Files: []string{"b.go", "c.go"}},
	}
	files := recentFilesFromCommits(commits, 10)
	if len(files) != 3 {
		t.Fatalf("expected 3 unique files, got %v", files)
	}
}

func TestRecentFilesFromCommitsRespectsMax(t *testing.T) {
	commits := []repo.Commit{
		{Files: []string{"a.go", "b.go", "c.go"}},
	}
	files := recentFilesFromCommits(commits, 2)
	if len(files) != 2 {
		t.Fatalf("expected max 2 files, got %v", files)
	}
}

func TestHotspotDirsFromSignals(t *testing.T) {
	sigs := repo.Signals{
		HotspotDirs: []repo.DirChurn{
			{Dir: "pkg/service"},
			{Dir: "internal/handler"},
			{Dir: ""},
		},
	}
	dirs := hotspotDirsFromSignals(sigs, 2)
	if len(dirs) != 2 {
		t.Fatalf("expected 2 dirs (skipping empty), got %v", dirs)
	}
	if dirs[0] != "pkg/service" {
		t.Errorf("expected first dir pkg/service, got %q", dirs[0])
	}
}

func TestHotfixSignalsFromSignals(t *testing.T) {
	sigs := repo.Signals{
		HotfixCommits: []repo.Commit{
			{Subject: "hotfix: fix critical login bug"},
			{Subject: "hotfix: revert auth change"},
		},
	}
	hints := hotfixSignalsFromSignals(sigs, 1)
	if len(hints) != 1 {
		t.Fatalf("expected 1 hint (respecting max), got %v", hints)
	}
	if hints[0] != "hotfix: fix critical login bug" {
		t.Errorf("expected first hotfix subject, got %q", hints[0])
	}
}

func TestDriftSignalsFromSignals(t *testing.T) {
	sigs := repo.Signals{
		RevertCommits: []repo.Commit{
			{Subject: "Revert auth changes"},
		},
		RepeatedDirs: []repo.DirChurn{
			{Dir: "pkg/config"},
		},
	}
	hints := driftSignalsFromSignals(sigs, 5)
	if len(hints) != 2 {
		t.Fatalf("expected 2 drift hints (1 revert + 1 repeated dir), got %v", hints)
	}
}

func TestMergeRepoStatesNilInputReturnsNil(t *testing.T) {
	if got := mergeRepoStates(nil, nil); got != nil {
		t.Fatalf("expected nil for all-nil inputs, got %#v", got)
	}
}

func TestMergeRepoStatesMergesFiles(t *testing.T) {
	a := &scoring.RepoState{ChangedFiles: []string{"a.go"}, Root: "/root"}
	b := &scoring.RepoState{ChangedFiles: []string{"b.go"}, Provider: "github-actions"}
	merged := mergeRepoStates(a, b)
	if merged == nil {
		t.Fatal("expected non-nil merged state")
	}
	if merged.Root != "/root" {
		t.Errorf("expected root /root, got %q", merged.Root)
	}
	if merged.Provider != "github-actions" {
		t.Errorf("expected provider github-actions, got %q", merged.Provider)
	}
	if len(merged.ChangedFiles) != 2 {
		t.Errorf("expected 2 changed files, got %v", merged.ChangedFiles)
	}
}

func TestMergeRepoStatesDedupesFiles(t *testing.T) {
	a := &scoring.RepoState{ChangedFiles: []string{"shared.go", "only-a.go"}}
	b := &scoring.RepoState{ChangedFiles: []string{"shared.go", "only-b.go"}}
	merged := mergeRepoStates(a, b)
	if len(merged.ChangedFiles) != 3 {
		t.Errorf("expected 3 unique changed files, got %v", merged.ChangedFiles)
	}
}

func TestCloneDeltaEnvDiffEngine(t *testing.T) {
	if got := cloneDeltaEnvDiff(nil); got != nil {
		t.Fatalf("expected nil for nil input, got %#v", got)
	}
	in := map[string]model.DeltaEnvChange{
		"GO_VERSION": {Baseline: "1.21", Current: "1.22"},
	}
	out := cloneDeltaEnvDiff(in)
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out["GO_VERSION"].Baseline != "1.21" {
		t.Errorf("expected Baseline 1.21, got %q", out["GO_VERSION"].Baseline)
	}
	// Mutating original should not affect clone
	in["GO_VERSION"] = model.DeltaEnvChange{Baseline: "mutated"}
	if out["GO_VERSION"].Baseline != "1.21" {
		t.Errorf("clone was mutated by original modification")
	}
}

func TestToModelTopologySignals(t *testing.T) {
	active := []string{
		topology.SignalBoundaryCrossed,
		topology.SignalUpstreamChanged,
		topology.SignalOwnershipMismatch,
		topology.SignalFailureClustered,
	}
	ts := toModelTopologySignals(active)
	if ts == nil {
		t.Fatal("expected non-nil topology signals")
	}
	if !ts.BoundaryCrossed {
		t.Error("expected BoundaryCrossed = true")
	}
	if !ts.UpstreamChanged {
		t.Error("expected UpstreamChanged = true")
	}
	if !ts.OwnershipMismatch {
		t.Error("expected OwnershipMismatch = true")
	}
	if !ts.FailureClustered {
		t.Error("expected FailureClustered = true")
	}
	if len(ts.ActiveSignals) != len(active) {
		t.Errorf("expected %d active signals, got %d", len(active), len(ts.ActiveSignals))
	}
}

func TestToModelTopologySignalsEmpty(t *testing.T) {
	ts := toModelTopologySignals(nil)
	if ts == nil {
		t.Fatal("expected non-nil topology signals for nil input")
	}
	if ts.BoundaryCrossed || ts.UpstreamChanged || ts.OwnershipMismatch || ts.FailureClustered {
		t.Error("expected all boolean signals false for nil input")
	}
}
