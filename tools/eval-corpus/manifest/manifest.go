// Package manifest provides corpus versioning and content addressing.
// A Manifest is a stable, deterministic summary of an ingested corpus.
package manifest

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"faultline/tools/eval-corpus/model"
)

// Manifest describes the contents of an ingested corpus.
type Manifest struct {
	// CorpusID is a human-assigned name for the corpus (e.g. "ci-realworld-v1").
	CorpusID string `json:"corpus_id"`
	// CreatedAt is an RFC-3339 timestamp recorded at manifest generation time.
	CreatedAt string `json:"created_at"`
	// InputSources lists the source names present in the corpus.
	InputSources []string `json:"input_sources"`
	// FixtureCount is the total number of fixtures.
	FixtureCount int `json:"fixture_count"`
	// OverallCorpusHash is a SHA-256 of all sorted fixture IDs, providing a
	// stable content address for the corpus independent of file order.
	OverallCorpusHash string `json:"overall_corpus_hash"`
	// ConfigHash is the SHA-256 of the ingest config YAML, if provided.
	ConfigHash string `json:"config_hash,omitempty"`
	// ToolVersion is the value of the --tool-version flag, if provided.
	ToolVersion string `json:"tool_version,omitempty"`
}

// Build reads a corpus JSONL stream and builds a Manifest.
// corpusID, configHash, and toolVersion are caller-supplied metadata.
func Build(r io.Reader, corpusID, configHash, toolVersion string) (Manifest, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 4<<20), 4<<20)

	sourceSet := map[string]struct{}{}
	var ids []string

	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var fix model.Fixture
		if err := json.Unmarshal(line, &fix); err != nil {
			return Manifest{}, fmt.Errorf("decode fixture: %w", err)
		}
		ids = append(ids, fix.ID)
		if fix.Source != "" {
			sourceSet[fix.Source] = struct{}{}
		}
	}
	if err := sc.Err(); err != nil {
		return Manifest{}, err
	}

	// Sort IDs before hashing — order in the file must not affect the result.
	sort.Strings(ids)

	sources := make([]string, 0, len(sourceSet))
	for s := range sourceSet {
		sources = append(sources, s)
	}
	sort.Strings(sources)

	m := Manifest{
		CorpusID:          corpusID,
		CreatedAt:         time.Now().UTC().Format(time.RFC3339),
		InputSources:      sources,
		FixtureCount:      len(ids),
		OverallCorpusHash: hashIDs(ids),
		ConfigHash:        configHash,
		ToolVersion:       toolVersion,
	}
	return m, nil
}

// BuildFromFile opens path and delegates to Build.
func BuildFromFile(path, corpusID, configHash, toolVersion string) (Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return Manifest{}, err
	}
	defer f.Close()
	return Build(f, corpusID, configHash, toolVersion)
}

// Write serialises the manifest as indented JSON to w.
func Write(w io.Writer, m Manifest) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}

// hashIDs returns the SHA-256 hex digest of the sorted, newline-joined IDs.
func hashIDs(ids []string) string {
	h := sha256.New()
	for _, id := range ids {
		_, _ = io.WriteString(h, id)
		_, _ = io.WriteString(h, "\n")
	}
	return hex.EncodeToString(h.Sum(nil))
}

// HashFile returns the SHA-256 of the contents of the file at path.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
