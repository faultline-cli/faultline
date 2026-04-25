package manifest_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/tools/eval-corpus/manifest"
)

const sampleCorpus = `{"id":"abc123","raw":"some log","source":"test","metadata":{}}
{"id":"def456","raw":"another log","source":"test","metadata":{}}
{"id":"ghi789","raw":"third log","source":"dataset-b","metadata":{}}
`

func TestBuildFixtureCount(t *testing.T) {
	r := strings.NewReader(sampleCorpus)
	m, err := manifest.Build(r, "test-corpus", "", "")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if m.FixtureCount != 3 {
		t.Errorf("FixtureCount = %d, want 3", m.FixtureCount)
	}
}

func TestBuildSourcesCollected(t *testing.T) {
	r := strings.NewReader(sampleCorpus)
	m, err := manifest.Build(r, "test-corpus", "", "")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(m.InputSources) != 2 {
		t.Errorf("InputSources = %v, want [dataset-b test]", m.InputSources)
	}
}

func TestBuildCorpusHashStable(t *testing.T) {
	r1 := strings.NewReader(sampleCorpus)
	m1, err := manifest.Build(r1, "test-corpus", "", "")
	if err != nil {
		t.Fatalf("Build 1: %v", err)
	}

	// Same fixtures in different order → same hash.
	reordered := `{"id":"ghi789","raw":"third log","source":"dataset-b","metadata":{}}
{"id":"abc123","raw":"some log","source":"test","metadata":{}}
{"id":"def456","raw":"another log","source":"test","metadata":{}}
`
	r2 := strings.NewReader(reordered)
	m2, err := manifest.Build(r2, "test-corpus", "", "")
	if err != nil {
		t.Fatalf("Build 2: %v", err)
	}

	if m1.OverallCorpusHash != m2.OverallCorpusHash {
		t.Errorf("hash differs for same fixtures in different order:\n  %s\n  %s",
			m1.OverallCorpusHash, m2.OverallCorpusHash)
	}
}

func TestBuildCorpusHashChangesOnDifferentFixtures(t *testing.T) {
	r1 := strings.NewReader(sampleCorpus)
	m1, _ := manifest.Build(r1, "c", "", "")

	differentCorpus := `{"id":"zzz999","raw":"different","source":"test","metadata":{}}
`
	r2 := strings.NewReader(differentCorpus)
	m2, _ := manifest.Build(r2, "c", "", "")

	if m1.OverallCorpusHash == m2.OverallCorpusHash {
		t.Error("hash should differ for different fixture sets")
	}
}

func TestBuildMetadataPreserved(t *testing.T) {
	r := strings.NewReader(sampleCorpus)
	m, err := manifest.Build(r, "my-corpus", "cfg-hash", "v1.2.3")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if m.CorpusID != "my-corpus" {
		t.Errorf("CorpusID = %q, want %q", m.CorpusID, "my-corpus")
	}
	if m.ConfigHash != "cfg-hash" {
		t.Errorf("ConfigHash = %q, want %q", m.ConfigHash, "cfg-hash")
	}
	if m.ToolVersion != "v1.2.3" {
		t.Errorf("ToolVersion = %q, want %q", m.ToolVersion, "v1.2.3")
	}
	if m.CreatedAt == "" {
		t.Error("CreatedAt should be set")
	}
}

func TestHashFileMatchesContent(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(tmp, []byte("field: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h1, err := manifest.HashFile(tmp)
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}
	h2, err := manifest.HashFile(tmp)
	if err != nil {
		t.Fatalf("HashFile 2: %v", err)
	}
	if h1 != h2 {
		t.Error("HashFile is not stable")
	}
	if len(h1) != 64 {
		t.Errorf("HashFile len = %d, want 64 (SHA-256 hex)", len(h1))
	}
}
