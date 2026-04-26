package store

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"faultline/internal/model"
)

// --- DefaultPath ---

func TestDefaultPathReturnsPathUnderHome(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}
	// Should be under the user's home directory
	rel, err := filepath.Rel(home, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		t.Errorf("expected path under home dir %q, got %q", home, path)
	}
}

// --- OpenBestEffort strict mode ---

func TestOpenBestEffortStrictModeCorruptFileErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.db")
	if err := os.WriteFile(path, []byte("not-sqlite"), 0o600); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}
	_, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path, Strict: true})
	if err == nil {
		t.Fatal("expected error in strict mode for corrupt db")
	}
}

func TestOpenBestEffortNonStrictCorruptFileDegrades(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.db")
	if err := os.WriteFile(path, []byte("not-sqlite"), 0o600); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}
	st, info, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path, Strict: false})
	if err != nil {
		t.Fatalf("unexpected error in non-strict mode: %v", err)
	}
	defer st.Close()
	if !info.Degraded {
		t.Errorf("expected degraded=true, got %#v", info)
	}
	if info.Warning == "" {
		t.Error("expected non-empty warning in degraded info")
	}
}

func TestOpenBestEffortAutoModeWithValidPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "good.db")
	st, info, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()
	if info.Mode != ModeAuto {
		t.Errorf("expected auto mode, got %v", info.Mode)
	}
	if info.Backend != "sqlite" {
		t.Errorf("expected sqlite backend, got %q", info.Backend)
	}
	if info.Path != path {
		t.Errorf("expected path %q, got %q", path, info.Path)
	}
	if info.Degraded {
		t.Error("expected non-degraded for valid path")
	}
}

// --- ResolveConfig whitespace handling ---

func TestResolveConfigWhitespaceTrimmed(t *testing.T) {
	// Whitespace-only should be treated as empty → auto
	cfg, err := ResolveConfig("   ", false)
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Mode != ModeAuto {
		t.Errorf("expected auto for whitespace, got %v", cfg.Mode)
	}
	if cfg.Path != "" {
		t.Errorf("expected empty path for whitespace, got %q", cfg.Path)
	}
}

func TestResolveConfigDisableOverridesExplicitPath(t *testing.T) {
	cfg, err := ResolveConfig("/some/path.db", true)
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Mode != ModeOff {
		t.Errorf("expected off when disable=true, got %v", cfg.Mode)
	}
	if cfg.Path != "" {
		t.Errorf("expected empty path when disable=true, got %q", cfg.Path)
	}
}

// --- RecentTopFailures ---

func TestSQLiteStoreRecentTopFailuresReturnsIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC)
	for i, id := range []string{"docker-auth", "missing-exec", "network-timeout"} {
		sig := SignatureForResult(model.Result{
			Playbook: model.Playbook{ID: id},
			Evidence: []string{"evidence for " + id},
		}).Hash
		handle, err := st.BeginRun(ctx, BeginRunParams{
			Surface:    "analyze",
			SourceKind: "log",
			Source:     "stdin",
			InputHash:  id + "-input",
			StartedAt:  now.Add(time.Duration(i) * time.Minute),
		})
		if err != nil {
			t.Fatalf("BeginRun: %v", err)
		}
		if err := st.CompleteRun(ctx, handle, CompleteRunParams{
			CompletedAt: now.Add(time.Duration(i) * time.Minute),
			Analysis: &model.Analysis{
				Source:     "stdin",
				InputHash:  id + "-input",
				OutputHash: id + "-output",
				Results: []model.Result{{
					Playbook:      model.Playbook{ID: id, Title: id, Category: "test"},
					Detector:      "log",
					Score:         4.0,
					Confidence:    0.80,
					Evidence:      []string{"evidence for " + id},
					SignatureHash: sig,
				}},
			},
		}); err != nil {
			t.Fatalf("CompleteRun: %v", err)
		}
	}

	failures, err := st.RecentTopFailures(ctx, 10)
	if err != nil {
		t.Fatalf("RecentTopFailures: %v", err)
	}
	if len(failures) != 3 {
		t.Fatalf("expected 3 recent failures, got %d: %v", len(failures), failures)
	}
}

func TestSQLiteStoreRecentTopFailuresEmptyDB(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	failures, err := st.RecentTopFailures(context.Background(), 10)
	if err != nil {
		t.Fatalf("RecentTopFailures empty: %v", err)
	}
	if len(failures) != 0 {
		t.Errorf("expected 0 failures for empty db, got %d", len(failures))
	}
}

func TestSQLiteStoreRecentTopFailuresZeroLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()
	// Zero limit should default to 500 — should not error
	_, err = st.RecentTopFailures(context.Background(), 0)
	if err != nil {
		t.Fatalf("RecentTopFailures with limit=0: %v", err)
	}
}

// --- GetRecentFindingsBySignature ---

func TestSQLiteStoreGetRecentFindingsBySignature(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 26, 11, 0, 0, 0, time.UTC)

	sig := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Evidence: []string{"pull access denied"},
	}).Hash

	// Insert two runs with the same signature
	for i := range 2 {
		handle, err := st.BeginRun(ctx, BeginRunParams{
			Surface:    "analyze",
			SourceKind: "log",
			Source:     "stdin",
			InputHash:  "docker-input-" + string(rune('a'+i)),
			StartedAt:  now.Add(time.Duration(i) * time.Minute),
		})
		if err != nil {
			t.Fatalf("BeginRun: %v", err)
		}
		if err := st.CompleteRun(ctx, handle, CompleteRunParams{
			CompletedAt: now.Add(time.Duration(i) * time.Minute),
			Analysis: &model.Analysis{
				Source:     "stdin",
				InputHash:  "docker-input-" + string(rune('a'+i)),
				OutputHash: "docker-output",
				Results: []model.Result{{
					Playbook:      model.Playbook{ID: "docker-auth", Title: "Docker auth", Category: "auth"},
					Detector:      "log",
					Score:         4.5,
					Confidence:    0.90,
					Evidence:      []string{"pull access denied"},
					SignatureHash: sig,
				}},
			},
		}); err != nil {
			t.Fatalf("CompleteRun: %v", err)
		}
	}

	findings, err := st.GetRecentFindingsBySignature(ctx, sig, 10)
	if err != nil {
		t.Fatalf("GetRecentFindingsBySignature: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d: %v", len(findings), findings)
	}
	for _, f := range findings {
		if f.FailureID != "docker-auth" {
			t.Errorf("FailureID = %q, want docker-auth", f.FailureID)
		}
		if f.SignatureHash != sig {
			t.Errorf("SignatureHash = %q, want %q", f.SignatureHash, sig)
		}
	}
}

func TestSQLiteStoreGetRecentFindingsBySignatureEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	findings, err := st.GetRecentFindingsBySignature(context.Background(), "nonexistent-hash", 10)
	if err != nil {
		t.Fatalf("GetRecentFindingsBySignature: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for unknown signature, got %d", len(findings))
	}
}

func TestSQLiteStoreGetRecentFindingsBySignatureBlankHash(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	// Blank hash should short-circuit and return nil, nil
	findings, err := st.GetRecentFindingsBySignature(context.Background(), "   ", 10)
	if err != nil {
		t.Fatalf("GetRecentFindingsBySignature blank: %v", err)
	}
	if findings != nil {
		t.Errorf("expected nil for blank hash, got %v", findings)
	}
}

// --- LookupHookHistory ---

func TestSQLiteStoreLookupHookHistoryReturnsNilForNoResults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	summary, err := st.LookupHookHistory(context.Background(), "nonexistent-hash", "docker-auth")
	if err != nil {
		t.Fatalf("LookupHookHistory: %v", err)
	}
	if summary != nil {
		t.Errorf("expected nil for no hook history, got %#v", summary)
	}
}

func TestSQLiteStoreLookupHookHistoryBlankInputs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	// Either blank should return nil, nil without error
	ctx := context.Background()
	summary, err := st.LookupHookHistory(ctx, "", "docker-auth")
	if err != nil || summary != nil {
		t.Errorf("blank signatureHash: err=%v summary=%v", err, summary)
	}
	summary, err = st.LookupHookHistory(ctx, "some-hash", "")
	if err != nil || summary != nil {
		t.Errorf("blank playbookID: err=%v summary=%v", err, summary)
	}
}

func TestSQLiteStoreLookupHookHistorySummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 26, 12, 0, 0, 0, time.UTC)

	passed := true
	sig := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Evidence: []string{"pull access denied for registry.example.com"},
	}).Hash

	handle, err := st.BeginRun(ctx, BeginRunParams{
		Surface:    "analyze",
		SourceKind: "log",
		Source:     "stdin",
		InputHash:  "hook-test-input",
		StartedAt:  now,
	})
	if err != nil {
		t.Fatalf("BeginRun: %v", err)
	}
	if err := st.CompleteRun(ctx, handle, CompleteRunParams{
		CompletedAt: now,
		Analysis: &model.Analysis{
			Source:     "stdin",
			InputHash:  "hook-test-input",
			OutputHash: "hook-test-output",
			Results: []model.Result{{
				Playbook:      model.Playbook{ID: "docker-auth", Title: "Docker auth", Category: "auth"},
				Detector:      "log",
				Score:         4.5,
				Confidence:    0.95,
				Evidence:      []string{"pull access denied for registry.example.com"},
				SignatureHash: sig,
				Hooks: &model.HookReport{
					Mode:            model.HookModeSafe,
					BaseConfidence:  0.90,
					ConfidenceDelta: 0.05,
					FinalConfidence: 0.95,
					Results: []model.HookResult{{
						ID:              "registry-config",
						Category:        model.HookCategoryVerify,
						Status:          model.HookStatusExecuted,
						Passed:          &passed,
						ConfidenceDelta: 0.05,
					}},
				},
			}},
		},
	}); err != nil {
		t.Fatalf("CompleteRun: %v", err)
	}

	summary, err := st.LookupHookHistory(ctx, sig, "docker-auth")
	if err != nil {
		t.Fatalf("LookupHookHistory: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary after running with hooks")
	}
	if summary.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", summary.TotalCount)
	}
	if summary.PassedCount != 1 {
		t.Errorf("PassedCount = %d, want 1", summary.PassedCount)
	}
	if summary.ExecutedCount != 1 {
		t.Errorf("ExecutedCount = %d, want 1", summary.ExecutedCount)
	}
}

// --- nullableBool / boolToInt (internal helpers via exported behaviour) ---

func TestSQLiteStoreNullableBoolViaHookResult(t *testing.T) {
	// Verify that hook results with nil Passed are handled correctly (nullableBool(nil) → NULL)
	// and hook results with false Passed are stored correctly (nullableBool(&false) → 0)
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 26, 13, 0, 0, 0, time.UTC)

	passedNil := (*bool)(nil)
	passedFalse := false

	sig := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "test-playbook"},
		Evidence: []string{"some error"},
	}).Hash

	handle, err := st.BeginRun(ctx, BeginRunParams{
		Surface:    "analyze",
		SourceKind: "log",
		Source:     "stdin",
		InputHash:  "nullable-test-input",
		StartedAt:  now,
	})
	if err != nil {
		t.Fatalf("BeginRun: %v", err)
	}
	if err := st.CompleteRun(ctx, handle, CompleteRunParams{
		CompletedAt: now,
		Analysis: &model.Analysis{
			Source:     "stdin",
			InputHash:  "nullable-test-input",
			OutputHash: "nullable-test-output",
			Results: []model.Result{{
				Playbook:      model.Playbook{ID: "test-playbook", Title: "Test", Category: "test"},
				Detector:      "log",
				Score:         3.0,
				Confidence:    0.70,
				Evidence:      []string{"some error"},
				SignatureHash: sig,
				Hooks: &model.HookReport{
					Mode: model.HookModeSafe,
					Results: []model.HookResult{
						{
							ID:       "hook-nil-passed",
							Category: model.HookCategoryVerify,
							Status:   model.HookStatusSkipped,
							Passed:   passedNil,
						},
						{
							ID:       "hook-false-passed",
							Category: model.HookCategoryVerify,
							Status:   model.HookStatusExecuted,
							Passed:   &passedFalse,
						},
					},
				},
			}},
		},
	}); err != nil {
		t.Fatalf("CompleteRun: %v", err)
	}

	// Verify hook stats were recorded correctly
	hooks, err := st.ListHookStats(ctx, 10)
	if err != nil {
		t.Fatalf("ListHookStats: %v", err)
	}
	if len(hooks) == 0 {
		t.Fatal("expected hook stats to be recorded")
	}

	// Find the failed hook
	var failedHook *HookStats
	for i := range hooks {
		if hooks[i].HookID == "hook-false-passed" {
			failedHook = &hooks[i]
		}
	}
	if failedHook == nil {
		t.Fatal("expected hook-false-passed to appear in stats")
	}
	if failedHook.FailedCount != 1 {
		t.Errorf("FailedCount = %d, want 1 for hook-false-passed", failedHook.FailedCount)
	}
}

// --- OpenBestEffort ModeOff ---

func TestOpenBestEffortModeOffReturnsNoop(t *testing.T) {
	st, info, err := OpenBestEffort(Config{Mode: ModeOff})
	if err != nil {
		t.Fatalf("OpenBestEffort(ModeOff): %v", err)
	}
	if info.Mode != ModeOff {
		t.Errorf("expected mode=off, got %v", info.Mode)
	}
	if info.Degraded {
		t.Error("expected non-degraded for explicit off mode")
	}
	// Noop store must not error on any operation
	ctx := context.Background()
	handle, err := st.BeginRun(ctx, BeginRunParams{
		Surface:    "test",
		SourceKind: "log",
		Source:     "stdin",
		InputHash:  "test-hash",
		StartedAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("Noop BeginRun: %v", err)
	}
	if err := st.CompleteRun(ctx, handle, CompleteRunParams{
		CompletedAt: time.Now(),
	}); err != nil {
		t.Fatalf("Noop CompleteRun: %v", err)
	}
	st.Close()
}

// TestOpenBestEffortDefaultPathFallback verifies that when no explicit path is
// provided, OpenBestEffort calls DefaultPath() and opens an SQLite store there.
func TestOpenBestEffortDefaultPathFallback(t *testing.T) {
	// Point HOME to a temp dir so DefaultPath() resolves to a writable location.
	home := t.TempDir()
	t.Setenv("HOME", home)

	st, info, err := OpenBestEffort(Config{Mode: ModeAuto, Path: ""})
	if err != nil {
		t.Fatalf("OpenBestEffort default path fallback: %v", err)
	}
	defer st.Close()
	if info.Path == "" {
		t.Error("expected non-empty resolved path from DefaultPath fallback")
	}
	if info.Backend != "sqlite" {
		t.Errorf("expected sqlite backend, got %q", info.Backend)
	}
}

// --- ResolveConfig ModeOff explicit value ---

func TestResolveConfigExplicitOffMode(t *testing.T) {
	cfg, err := ResolveConfig("off", false)
	if err != nil {
		t.Fatalf("ResolveConfig('off'): %v", err)
	}
	if cfg.Mode != ModeOff {
		t.Errorf("expected off mode, got %v", cfg.Mode)
	}
	if cfg.Path != "" {
		t.Errorf("expected empty path, got %q", cfg.Path)
	}
}

// --- boolToInt internal helper exercised via analysis with/without results ---

// TestBoolToIntBothBranchesViaCompleteRun exercises boolToInt(false) by
// completing a run whose analysis has no results, and boolToInt(true) by
// completing one that has results.
func TestBoolToIntBothBranchesViaCompleteRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Date(2026, 4, 26, 14, 0, 0, 0, time.UTC)

	// Run 1: analysis has zero results → boolToInt(false)
	handle1, err := st.BeginRun(ctx, BeginRunParams{
		Surface: "analyze", SourceKind: "log", Source: "stdin",
		InputHash: "no-results-input", StartedAt: now,
	})
	if err != nil {
		t.Fatalf("BeginRun 1: %v", err)
	}
	if err := st.CompleteRun(ctx, handle1, CompleteRunParams{
		CompletedAt: now,
		Analysis: &model.Analysis{
			Source:     "stdin",
			InputHash:  "no-results-input",
			OutputHash: "no-results-output",
			Results:    []model.Result{},
		},
	}); err != nil {
		t.Fatalf("CompleteRun no-results: %v", err)
	}

	// Run 2: analysis has one result → boolToInt(true)
	sig := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Evidence: []string{"auth failure"},
	}).Hash
	handle2, err := st.BeginRun(ctx, BeginRunParams{
		Surface: "analyze", SourceKind: "log", Source: "stdin",
		InputHash: "with-results-input", StartedAt: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("BeginRun 2: %v", err)
	}
	if err := st.CompleteRun(ctx, handle2, CompleteRunParams{
		CompletedAt: now.Add(time.Minute),
		Analysis: &model.Analysis{
			Source:     "stdin",
			InputHash:  "with-results-input",
			OutputHash: "with-results-output",
			Results: []model.Result{{
				Playbook:      model.Playbook{ID: "docker-auth", Title: "Docker Auth", Category: "auth"},
				Detector:      "log",
				Score:         4.0,
				Confidence:    0.85,
				Evidence:      []string{"auth failure"},
				SignatureHash: sig,
			}},
		},
	}); err != nil {
		t.Fatalf("CompleteRun with-results: %v", err)
	}
}
