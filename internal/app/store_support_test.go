package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeJSONIncludesStoreHistoryFields(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "faultline.db")
	log := "Error response from daemon: pull access denied for mcr/microsoft.com/mssql/server, repository does not exist or may require 'docker login'\n"
	opts := AnalyzeOptions{
		JSON:         true,
		NoHistory:    false,
		Store:        storePath,
		PlaybookDir:  repoPlaybookDir(),
		BayesEnabled: false,
	}

	var first bytes.Buffer
	if err := NewService().Analyze(bytes.NewBufferString(log), "stdin", opts, &first); err != nil {
		t.Fatalf("Analyze first: %v", err)
	}
	var firstPayload map[string]any
	if err := json.Unmarshal(first.Bytes(), &firstPayload); err != nil {
		t.Fatalf("unmarshal first payload: %v", err)
	}
	results := firstPayload["results"].([]any)
	firstResult := results[0].(map[string]any)
	if _, ok := firstPayload["input_hash"].(string); !ok {
		t.Fatalf("expected input_hash in first payload, got %v", firstPayload["input_hash"])
	}
	if _, ok := firstPayload["output_hash"].(string); !ok {
		t.Fatalf("expected output_hash in first payload, got %v", firstPayload["output_hash"])
	}
	if firstResult["signature_hash"] == "" {
		t.Fatalf("expected signature_hash in first payload, got %#v", firstResult)
	}
	if firstResult["occurrence_count"].(float64) != 1 {
		t.Fatalf("expected first occurrence_count=1, got %#v", firstResult["occurrence_count"])
	}

	var second bytes.Buffer
	if err := NewService().Analyze(bytes.NewBufferString(log), "stdin", opts, &second); err != nil {
		t.Fatalf("Analyze second: %v", err)
	}
	var secondPayload map[string]any
	if err := json.Unmarshal(second.Bytes(), &secondPayload); err != nil {
		t.Fatalf("unmarshal second payload: %v", err)
	}
	secondResults := secondPayload["results"].([]any)
	secondResult := secondResults[0].(map[string]any)
	if secondResult["seen_before"] != true {
		t.Fatalf("expected seen_before=true on second run, got %#v", secondResult["seen_before"])
	}
	if secondResult["occurrence_count"].(float64) != 2 {
		t.Fatalf("expected second occurrence_count=2, got %#v", secondResult["occurrence_count"])
	}
}

func TestAnalyzeGracefullyDegradesWhenStoreIsCorrupt(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "faultline.db")
	if err := os.WriteFile(storePath, []byte("not-a-sqlite-database"), 0o600); err != nil {
		t.Fatalf("write corrupt store: %v", err)
	}
	log := "Error response from daemon: pull access denied for mcr/microsoft.com/mssql/server, repository does not exist or may require 'docker login'\n"
	opts := AnalyzeOptions{
		JSON:        true,
		Store:       storePath,
		PlaybookDir: repoPlaybookDir(),
	}
	var out bytes.Buffer
	if err := NewService().Analyze(bytes.NewBufferString(log), "stdin", opts, &out); err != nil {
		t.Fatalf("Analyze with corrupt store should degrade gracefully, got %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("expected JSON output even when the store is corrupt")
	}
}
