package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type analysisResultPayload struct {
	FailureID       string `json:"failure_id"`
	SignatureHash   string `json:"signature_hash"`
	OccurrenceCount int    `json:"occurrence_count"`
	SeenBefore      bool   `json:"seen_before"`
	FirstSeenAt     string `json:"first_seen_at"`
	LastSeenAt      string `json:"last_seen_at"`
}

type analysisPayload struct {
	OutputHash string                  `json:"output_hash"`
	Results    []analysisResultPayload `json:"results"`
}

func TestAnalyzeJSONRecurrenceVariantHarness(t *testing.T) {
	svc := NewService()
	storePath := filepath.Join(t.TempDir(), "faultline.db")
	variantDir := filepath.Join("..", "signature", "testdata", "variants")
	opts := AnalyzeOptions{
		JSON:        true,
		Store:       storePath,
		PlaybookDir: repoPlaybookDir(),
	}

	cases := []struct {
		name         string
		file         string
		wantPlaybook string
		group        string
	}{
		{name: "missing executable linux", file: "missing-executable-1.log", wantPlaybook: "missing-executable", group: "missing-executable"},
		{name: "missing executable windows", file: "missing-executable-2.log", wantPlaybook: "missing-executable", group: "missing-executable"},
		{name: "missing executable hosted toolcache", file: "missing-executable-3.log", wantPlaybook: "missing-executable", group: "missing-executable"},
		{name: "node version mismatch linux", file: "node-version-mismatch-1.log", wantPlaybook: "node-version-mismatch", group: "node-version-mismatch"},
		{name: "node version mismatch toolcache", file: "node-version-mismatch-2.log", wantPlaybook: "node-version-mismatch", group: "node-version-mismatch"},
		{name: "env var missing api base url", file: "env-var-missing-1.log", wantPlaybook: "env-var-missing", group: "env-var-api-base-url"},
		{name: "env var missing api base url windows", file: "env-var-missing-2.log", wantPlaybook: "env-var-missing", group: "env-var-api-base-url"},
		{name: "env var missing database url", file: "env-var-missing-3.log", wantPlaybook: "env-var-missing", group: "env-var-database-url"},
		{name: "dependency drift react", file: "dependency-drift-react.log", wantPlaybook: "dependency-drift", group: "dependency-drift-react"},
		{name: "dependency drift grpc", file: "dependency-drift-grpc.log", wantPlaybook: "dependency-drift", group: "dependency-drift-grpc"},
	}

	groupHashes := map[string]string{}
	groupCounts := map[string]int{}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(variantDir, tc.file))
			if err != nil {
				t.Fatalf("read variant fixture: %v", err)
			}

			var out bytes.Buffer
			if err := svc.Analyze(bytes.NewReader(data), tc.file, opts, &out); err != nil {
				t.Fatalf("Analyze: %v", err)
			}

			payload := decodeAnalysisPayload(t, out.Bytes())
			if len(payload.Results) == 0 {
				t.Fatalf("expected ranked results for %s", tc.file)
			}
			result := payload.Results[0]
			if result.FailureID != tc.wantPlaybook {
				t.Fatalf("expected top playbook %s, got %s", tc.wantPlaybook, result.FailureID)
			}
			if result.SignatureHash == "" {
				t.Fatalf("expected signature_hash for %s", tc.file)
			}

			previousCount := groupCounts[tc.group]
			if existing, ok := groupHashes[tc.group]; ok {
				if result.SignatureHash != existing {
					t.Fatalf("expected recurring group %s to stay stable, got %s and %s", tc.group, existing, result.SignatureHash)
				}
				if !result.SeenBefore {
					t.Fatalf("expected seen_before for repeat group %s", tc.group)
				}
				if result.OccurrenceCount != previousCount+1 {
					t.Fatalf("expected occurrence_count=%d for group %s, got %d", previousCount+1, tc.group, result.OccurrenceCount)
				}
				if result.FirstSeenAt == "" || result.LastSeenAt == "" {
					t.Fatalf("expected first_seen_at and last_seen_at for repeat group %s, got %#v", tc.group, result)
				}
			} else {
				groupHashes[tc.group] = result.SignatureHash
				if result.OccurrenceCount != 1 {
					t.Fatalf("expected first occurrence_count=1 for group %s, got %d", tc.group, result.OccurrenceCount)
				}
				if result.SeenBefore {
					t.Fatalf("did not expect seen_before on first occurrence for %s", tc.group)
				}
			}
			groupCounts[tc.group] = previousCount + 1
		})
	}

	hashOwners := map[string]string{}
	for group, hash := range groupHashes {
		if other, ok := hashOwners[hash]; ok {
			t.Fatalf("expected groups %s and %s to stay distinct, but both normalized to %s", group, other, hash)
		}
		hashOwners[hash] = group
	}

	var historyOut bytes.Buffer
	if err := svc.History(groupHashes["env-var-api-base-url"], storePath, 10, true, &historyOut); err != nil {
		t.Fatalf("History: %v", err)
	}
	var historyPayload struct {
		Signature struct {
			SignatureHash   string `json:"signature_hash"`
			OccurrenceCount int    `json:"occurrence_count"`
		} `json:"signature"`
		Findings []struct {
			FailureID string `json:"failure_id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(historyOut.Bytes(), &historyPayload); err != nil {
		t.Fatalf("unmarshal history detail JSON: %v", err)
	}
	if historyPayload.Signature.SignatureHash != groupHashes["env-var-api-base-url"] {
		t.Fatalf("expected history detail for %s, got %s", groupHashes["env-var-api-base-url"], historyPayload.Signature.SignatureHash)
	}
	if historyPayload.Signature.OccurrenceCount != 2 {
		t.Fatalf("expected env var api-base-url occurrence_count=2, got %d", historyPayload.Signature.OccurrenceCount)
	}
	if len(historyPayload.Findings) != 2 {
		t.Fatalf("expected two recent findings for env-var api-base-url, got %#v", historyPayload.Findings)
	}
}

func TestVerifyDeterminismStaysStableAsHistoryAccumulates(t *testing.T) {
	svc := NewService()
	storePath := filepath.Join(t.TempDir(), "faultline.db")
	logPath := filepath.Join("..", "signature", "testdata", "variants", "missing-executable-1.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read determinism fixture: %v", err)
	}

	opts := AnalyzeOptions{
		JSON:        true,
		Store:       storePath,
		PlaybookDir: repoPlaybookDir(),
	}

	var firstOut bytes.Buffer
	if err := svc.Analyze(bytes.NewReader(data), logPath, opts, &firstOut); err != nil {
		t.Fatalf("Analyze first: %v", err)
	}
	first := decodeAnalysisPayload(t, firstOut.Bytes())

	var secondOut bytes.Buffer
	if err := svc.Analyze(bytes.NewReader(data), logPath, opts, &secondOut); err != nil {
		t.Fatalf("Analyze second: %v", err)
	}
	second := decodeAnalysisPayload(t, secondOut.Bytes())

	if len(first.Results) == 0 || len(second.Results) == 0 {
		t.Fatalf("expected ranked results in both runs: first=%#v second=%#v", first, second)
	}
	if first.Results[0].SignatureHash != second.Results[0].SignatureHash {
		t.Fatalf("expected repeated input to keep signature stable: %s vs %s", first.Results[0].SignatureHash, second.Results[0].SignatureHash)
	}
	if first.OutputHash == "" || second.OutputHash == "" {
		t.Fatalf("expected output hashes in both runs: first=%q second=%q", first.OutputHash, second.OutputHash)
	}
	if first.OutputHash != second.OutputHash {
		t.Fatalf("expected output_hash to stay stable across repeated runs, got %s and %s", first.OutputHash, second.OutputHash)
	}
	if first.Results[0].OccurrenceCount != 1 || second.Results[0].OccurrenceCount != 2 {
		t.Fatalf("expected occurrence counts 1 then 2, got %d then %d", first.Results[0].OccurrenceCount, second.Results[0].OccurrenceCount)
	}
	if second.Results[0].SeenBefore != true {
		t.Fatalf("expected second run to report seen_before, got %#v", second.Results[0])
	}

	var determinismOut bytes.Buffer
	if err := svc.VerifyDeterminism(bytes.NewReader(data), logPath, storePath, true, &determinismOut); err != nil {
		t.Fatalf("VerifyDeterminism: %v", err)
	}
	var determinismPayload struct {
		Determinism struct {
			RunCount             int  `json:"run_count"`
			DistinctOutputHashes int  `json:"distinct_output_hashes"`
			Stable               bool `json:"stable"`
		} `json:"determinism"`
	}
	if err := json.Unmarshal(determinismOut.Bytes(), &determinismPayload); err != nil {
		t.Fatalf("unmarshal determinism JSON: %v", err)
	}
	if !determinismPayload.Determinism.Stable || determinismPayload.Determinism.RunCount != 2 || determinismPayload.Determinism.DistinctOutputHashes != 1 {
		t.Fatalf("unexpected determinism summary: %#v", determinismPayload.Determinism)
	}
}

func decodeAnalysisPayload(t *testing.T, data []byte) analysisPayload {
	t.Helper()
	var payload analysisPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal analysis JSON: %v", err)
	}
	return payload
}
