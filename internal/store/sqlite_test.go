package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"faultline/internal/model"
)

func TestOpenBestEffortGracefullyDegradesCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	if err := os.WriteFile(path, []byte("not-a-sqlite-database"), 0o600); err != nil {
		t.Fatalf("write corrupt db: %v", err)
	}
	st, info, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()
	if !info.Degraded {
		t.Fatalf("expected degraded info for corrupt store, got %#v", info)
	}
}

func TestSQLiteStorePersistsRecurrenceAndDeterminism(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, info, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	if info.Degraded {
		t.Fatalf("unexpected degraded store info: %#v", info)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)
	analysis := &model.Analysis{
		Source:      "stdin",
		InputHash:   "input-1",
		OutputHash:  "output-1",
		Fingerprint: "fp-1",
		Results: []model.Result{{
			Playbook: model.Playbook{
				ID:       "docker-auth",
				Title:    "Docker auth",
				Category: "auth",
			},
			Detector:   "log",
			Score:      4.5,
			Confidence: 0.92,
			Evidence:   []string{"pull access denied for registry.example.com"},
			SignatureHash: SignatureForResult(model.Result{
				Playbook: model.Playbook{ID: "docker-auth"},
				Evidence: []string{"pull access denied for registry.example.com"},
			}).Hash,
		}},
	}

	handle, err := st.BeginRun(ctx, BeginRunParams{
		Surface:    "analyze",
		SourceKind: "log",
		Source:     "stdin",
		InputHash:  analysis.InputHash,
		StartedAt:  now,
	})
	if err != nil {
		t.Fatalf("BeginRun: %v", err)
	}
	if err := st.CompleteRun(ctx, handle, CompleteRunParams{
		CompletedAt: now,
		Analysis:    analysis,
	}); err != nil {
		t.Fatalf("CompleteRun first: %v", err)
	}

	history, err := st.LookupSignatureHistory(ctx, analysis.Results[0].SignatureHash)
	if err != nil {
		t.Fatalf("LookupSignatureHistory first: %v", err)
	}
	if history.OccurrenceCount != 1 {
		t.Fatalf("expected first occurrence count to be 1, got %#v", history)
	}

	secondHandle, err := st.BeginRun(ctx, BeginRunParams{
		Surface:    "analyze",
		SourceKind: "log",
		Source:     "stdin",
		InputHash:  analysis.InputHash,
		StartedAt:  now.Add(2 * time.Minute),
	})
	if err != nil {
		t.Fatalf("BeginRun second: %v", err)
	}
	second := *analysis
	second.OutputHash = "output-1"
	if err := st.CompleteRun(ctx, secondHandle, CompleteRunParams{
		CompletedAt: now.Add(2 * time.Minute),
		Analysis:    &second,
	}); err != nil {
		t.Fatalf("CompleteRun second: %v", err)
	}

	history, err = st.LookupSignatureHistory(ctx, analysis.Results[0].SignatureHash)
	if err != nil {
		t.Fatalf("LookupSignatureHistory second: %v", err)
	}
	if history.OccurrenceCount != 2 {
		t.Fatalf("expected second occurrence count to be 2, got %#v", history)
	}

	seenCount, err := st.CountSeenFailure(ctx, "docker-auth")
	if err != nil {
		t.Fatalf("CountSeenFailure: %v", err)
	}
	if seenCount != 2 {
		t.Fatalf("expected seen count 2, got %d", seenCount)
	}

	determinism, err := st.VerifyDeterminismForInputHash(ctx, "input-1")
	if err != nil {
		t.Fatalf("VerifyDeterminismForInputHash: %v", err)
	}
	if !determinism.Stable || determinism.RunCount != 2 || determinism.DistinctOutputHashes != 1 {
		t.Fatalf("unexpected determinism summary: %#v", determinism)
	}
}
