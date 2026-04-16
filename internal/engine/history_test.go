package engine

import (
	"os"
	"path/filepath"
	"testing"

	"faultline/internal/model"
)

func TestFingerprintDeterministic(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Evidence: []string{"pull access denied", "authentication required"},
	}
	a := fingerprint(result)
	b := fingerprint(result)
	if a != b {
		t.Errorf("fingerprint is not deterministic: %q != %q", a, b)
	}
	if len(a) != 8 {
		t.Errorf("expected 8-char hex fingerprint, got %q", a)
	}
}

func TestFingerprintDistinctResults(t *testing.T) {
	r1 := model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Evidence: []string{"pull access denied"},
	}
	r2 := model.Result{
		Playbook: model.Playbook{ID: "git-auth"},
		Evidence: []string{"terminal prompts disabled"},
	}
	if fingerprint(r1) == fingerprint(r2) {
		t.Error("expected different fingerprints for different results")
	}
}

func TestFingerprintNoEvidence(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{ID: "oom-killed"},
	}
	fp := fingerprint(result)
	if fp == "" {
		t.Error("expected non-empty fingerprint even with no evidence")
	}
}

func TestCountSeenMissingFileReturnsZero(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	count := countSeen("docker-auth")
	if count != 0 {
		t.Errorf("countSeen with missing history file = %d, want 0", count)
	}
}

func TestRecordAndCountSeen(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	result := model.Result{
		Playbook: model.Playbook{
			ID:       "git-auth",
			Title:    "Git auth failure",
			Category: "auth",
		},
	}
	recordResult(result)
	recordResult(result)

	count := countSeen("git-auth")
	if count != 2 {
		t.Errorf("countSeen after 2 records = %d, want 2", count)
	}
	if countSeen("docker-auth") != 0 {
		t.Error("expected 0 for unrecorded failure ID")
	}
}

func TestLoadHistoryMissingFileReturnsEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	store, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory with missing file: %v", err)
	}
	if len(store.Entries) != 0 {
		t.Errorf("expected empty history, got %d entries", len(store.Entries))
	}
}

func TestLoadHistoryCorruptedJSONReturnsEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, historySubdir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, historyFile), []byte("not-json{{{"), 0o600); err != nil {
		t.Fatalf("write corrupted history: %v", err)
	}
	store, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory with corrupted file: %v", err)
	}
	if len(store.Entries) != 0 {
		t.Errorf("expected empty history on corruption, got %d entries", len(store.Entries))
	}
}

func TestSaveAndLoadHistory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	store := historyStore{
		Entries: []historyEntry{
			{Timestamp: "2026-01-01T00:00:00Z", FailureID: "docker-auth", Title: "Docker auth", Category: "auth"},
			{Timestamp: "2026-01-02T00:00:00Z", FailureID: "git-auth", Title: "Git auth", Category: "auth"},
		},
	}
	if err := saveHistory(store); err != nil {
		t.Fatalf("saveHistory: %v", err)
	}

	loaded, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if len(loaded.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].FailureID != "docker-auth" {
		t.Errorf("expected docker-auth, got %q", loaded.Entries[0].FailureID)
	}
}

func TestRecordResultTrimsOldEntries(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	result := model.Result{
		Playbook: model.Playbook{
			ID:       "git-auth",
			Title:    "Git auth failure",
			Category: "auth",
		},
	}
	// Fill beyond the maxEntries limit
	entries := make([]historyEntry, maxEntries)
	for i := range entries {
		entries[i] = historyEntry{FailureID: "old-entry"}
	}
	_ = saveHistory(historyStore{Entries: entries})

	recordResult(result)

	store, err := loadHistory()
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if len(store.Entries) > maxEntries {
		t.Errorf("history has %d entries, should be capped at %d", len(store.Entries), maxEntries)
	}
}
