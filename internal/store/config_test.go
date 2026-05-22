package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"faultline/internal/model"
)

// --- ResolveConfig ---

func TestResolveConfigOffStringReturnsOffDirect(t *testing.T) {
	cfg, err := ResolveConfig("off")
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Mode != ModeOff {
		t.Fatalf("expected off mode, got %v", cfg.Mode)
	}
}

func TestResolveConfigEmptyStringReturnsAuto(t *testing.T) {
	cfg, err := ResolveConfig("")
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Mode != ModeAuto {
		t.Fatalf("expected auto mode, got %v", cfg.Mode)
	}
}

func TestResolveConfigAutoStringReturnsAuto(t *testing.T) {
	for _, raw := range []string{"auto", "AUTO", "Auto"} {
		cfg, err := ResolveConfig(raw)
		if err != nil {
			t.Fatalf("ResolveConfig(%q): %v", raw, err)
		}
		if cfg.Mode != ModeAuto {
			t.Fatalf("expected auto mode for %q, got %v", raw, cfg.Mode)
		}
	}
}

func TestResolveConfigOffStringReturnsOff(t *testing.T) {
	for _, raw := range []string{"off", "OFF", "Off"} {
		cfg, err := ResolveConfig(raw)
		if err != nil {
			t.Fatalf("ResolveConfig(%q): %v", raw, err)
		}
		if cfg.Mode != ModeOff {
			t.Fatalf("expected off mode for %q, got %v", raw, cfg.Mode)
		}
	}
}

func TestResolveConfigExplicitPathReturnsAutoWithPath(t *testing.T) {
	cfg, err := ResolveConfig("/tmp/custom.db")
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Mode != ModeAuto {
		t.Fatalf("expected auto mode, got %v", cfg.Mode)
	}
	if cfg.Path != "/tmp/custom.db" {
		t.Fatalf("expected explicit path, got %q", cfg.Path)
	}
}

// --- OpenBestEffort with ModeOff ---

func TestOpenBestEffortOffModeReturnsNoop(t *testing.T) {
	st, info, err := OpenBestEffort(Config{Mode: ModeOff})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()
	if info.Mode != ModeOff {
		t.Fatalf("expected off mode info, got %v", info.Mode)
	}
	if info.Degraded {
		t.Fatalf("expected non-degraded info for off mode, got %#v", info)
	}
}

func TestOpenBestEffortEmptyModeDefaultsToAuto(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	st, info, err := OpenBestEffort(Config{Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()
	if info.Mode != ModeAuto {
		t.Fatalf("expected auto mode for empty config mode, got %v", info.Mode)
	}
}

// --- GetRecentFindingsBySignature ---

func TestGetRecentFindingsBySignature(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)

	sig := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Detector: "log",
		Evidence: []string{"pull access denied"},
	}).Hash

	analysis := &model.Analysis{
		Source:      "stdin",
		InputHash:   "hash-1",
		OutputHash:  "out-1",
		Fingerprint: "fp-1",
		Results: []model.Result{{
			Playbook:      model.Playbook{ID: "docker-auth", Title: "Docker auth", Category: "auth"},
			Detector:      "log",
			Score:         4.5,
			Confidence:    0.90,
			Evidence:      []string{"pull access denied"},
			SignatureHash: sig,
		}},
	}

	handle, err := st.BeginRun(ctx, BeginRunParams{
		Surface:    "analyze",
		SourceKind: "log",
		Source:     "stdin",
		InputHash:  "hash-1",
		StartedAt:  now,
	})
	if err != nil {
		t.Fatalf("BeginRun: %v", err)
	}
	if err := st.CompleteRun(ctx, handle, CompleteRunParams{
		CompletedAt: now,
		Analysis:    analysis,
	}); err != nil {
		t.Fatalf("CompleteRun: %v", err)
	}

	findings, err := st.GetRecentFindingsBySignature(ctx, sig, 10)
	if err != nil {
		t.Fatalf("GetRecentFindingsBySignature: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].FailureID != "docker-auth" {
		t.Fatalf("unexpected finding: %#v", findings[0])
	}
}

func TestGetRecentFindingsBySignatureEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	findings, err := st.GetRecentFindingsBySignature(context.Background(), "nonexistent-sig", 10)
	if err != nil {
		t.Fatalf("GetRecentFindingsBySignature: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for unknown sig, got %d", len(findings))
	}
}

// --- RecentTopFailures ---

func TestRecentTopFailures(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)

	analysis := &model.Analysis{
		Source:      "stdin",
		InputHash:   "hash-topfail",
		OutputHash:  "out-topfail",
		Fingerprint: "fp-topfail",
		Results: []model.Result{
			{
				Playbook:   model.Playbook{ID: "docker-auth", Title: "Docker auth", Category: "auth"},
				Detector:   "log",
				Score:      4.5,
				Confidence: 0.90,
				Evidence:   []string{"pull access denied"},
			},
			{
				Playbook:   model.Playbook{ID: "network-timeout", Title: "Network timeout", Category: "network"},
				Detector:   "log",
				Score:      3.8,
				Confidence: 0.75,
				Evidence:   []string{"context deadline exceeded"},
			},
		},
	}

	handle, err := st.BeginRun(ctx, BeginRunParams{
		Surface:    "analyze",
		SourceKind: "log",
		Source:     "stdin",
		InputHash:  "hash-topfail",
		StartedAt:  now,
	})
	if err != nil {
		t.Fatalf("BeginRun: %v", err)
	}
	if err := st.CompleteRun(ctx, handle, CompleteRunParams{
		CompletedAt: now,
		Analysis:    analysis,
	}); err != nil {
		t.Fatalf("CompleteRun: %v", err)
	}

	failures, err := st.RecentTopFailures(ctx, 10)
	if err != nil {
		t.Fatalf("RecentTopFailures: %v", err)
	}
	if len(failures) == 0 {
		t.Fatal("expected at least one recent top failure")
	}
}

func TestRecentTopFailuresEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	failures, err := st.RecentTopFailures(context.Background(), 10)
	if err != nil {
		t.Fatalf("RecentTopFailures: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("expected 0 failures for empty store, got %d", len(failures))
	}
}

// --- Noop store methods ---

func TestNoopStoreMethodsReturnSafeDefaults(t *testing.T) {
	st, _, err := OpenBestEffort(Config{Mode: ModeOff})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()

	// BeginRun should return a valid handle
	handle, err := st.BeginRun(ctx, BeginRunParams{})
	if err != nil {
		t.Fatalf("noop BeginRun: %v", err)
	}

	// CompleteRun should succeed
	if err := st.CompleteRun(ctx, handle, CompleteRunParams{}); err != nil {
		t.Fatalf("noop CompleteRun: %v", err)
	}

	// All read methods should return empty without errors
	history, err := st.LookupSignatureHistory(ctx, "sig")
	if err != nil {
		t.Fatalf("noop LookupSignatureHistory: %v", err)
	}
	if history.SeenBefore {
		t.Fatal("expected noop to return not-seen-before")
	}

	count, err := st.CountSeenFailure(ctx, "id")
	if err != nil {
		t.Fatalf("noop CountSeenFailure: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}

	if _, err := st.RecentTopFailures(ctx, 10); err != nil {
		t.Fatalf("noop RecentTopFailures: %v", err)
	}
	if _, err := st.ListSignatures(ctx, 10); err != nil {
		t.Fatalf("noop ListSignatures: %v", err)
	}
	if _, err := st.GetRecentFindingsBySignature(ctx, "sig", 10); err != nil {
		t.Fatalf("noop GetRecentFindingsBySignature: %v", err)
	}
	if _, err := st.ListPlaybookStats(ctx, 10); err != nil {
		t.Fatalf("noop ListPlaybookStats: %v", err)
	}
}
