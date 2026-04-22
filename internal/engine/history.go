package engine

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"time"

	"faultline/internal/model"
)

const (
	historySubdir = ".faultline"
	historyFile   = "history.json"
	maxEntries    = 500
)

type historyEntry struct {
	Timestamp   string `json:"timestamp"`
	FailureID   string `json:"failure_id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Fingerprint string `json:"fingerprint"`
}

type historyStore struct {
	Entries []historyEntry `json:"entries"`
}

// countSeen returns how many history entries share the given failure ID.
// On any error (missing file, corrupt JSON), it returns 0 to avoid blocking
// analysis.
func countSeen(failureID string) int {
	store, err := loadHistory()
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range store.Entries {
		if e.FailureID == failureID {
			count++
		}
	}
	return count
}

// allHistoryEntries returns all local history entries for metrics computation.
// On any error it returns an empty slice so callers can silently skip metrics.
func allHistoryEntries() []historyEntry {
	store, err := loadHistory()
	if err != nil {
		return nil
	}
	return store.Entries
}

// recordResult appends result to the history file. Errors are silently
// swallowed so that history failures never interrupt analysis output.
func recordResult(result model.Result) {
	store, _ := loadHistory()
	store.Entries = append(store.Entries, historyEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		FailureID:   result.Playbook.ID,
		Title:       result.Playbook.Title,
		Category:    result.Playbook.Category,
		Fingerprint: fingerprint(result),
	})
	// Trim oldest entries so the file stays bounded.
	if len(store.Entries) > maxEntries {
		store.Entries = store.Entries[len(store.Entries)-maxEntries:]
	}
	_ = saveHistory(store)
}

// fingerprint returns a short hex hash derived from the playbook ID and the
// first three evidence lines so that identical failures get the same token.
func fingerprint(result model.Result) string {
	h := fnv.New32a()
	h.Write([]byte(result.Playbook.ID))
	lim := 3
	if len(result.Evidence) < lim {
		lim = len(result.Evidence)
	}
	for _, e := range result.Evidence[:lim] {
		h.Write([]byte(e))
	}
	return fmt.Sprintf("%08x", h.Sum32())
}

func historyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, historySubdir, historyFile), nil
}

func loadHistory() (historyStore, error) {
	path, err := historyPath()
	if err != nil {
		return historyStore{}, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return historyStore{}, nil
	}
	if err != nil {
		return historyStore{}, fmt.Errorf("read history: %w", err)
	}
	var store historyStore
	if err := json.Unmarshal(data, &store); err != nil {
		// Corrupted history: return empty rather than failing analysis.
		return historyStore{}, nil
	}
	return store, nil
}

func saveHistory(store historyStore) error {
	path, err := historyPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create history directory: %w", err)
	}
	data, err := json.Marshal(store)
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}
	// Write to a sibling temp file then rename for an atomic replacement,
	// so a crash mid-write never leaves the history file partially written.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write history temp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("commit history: %w", err)
	}
	return nil
}
