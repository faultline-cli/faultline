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

func TestSQLiteStorePersistsRecurrence(t *testing.T) {
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
}

func TestSQLiteStoreListsHistorySummaries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	base := time.Date(2026, 4, 23, 9, 0, 0, 0, time.UTC)

	writeRun := func(offset time.Duration, analysis *model.Analysis) {
		handle, err := st.BeginRun(ctx, BeginRunParams{
			Surface:    "analyze",
			SourceKind: "log",
			Source:     "stdin",
			InputHash:  analysis.InputHash,
			StartedAt:  base.Add(offset),
		})
		if err != nil {
			t.Fatalf("BeginRun: %v", err)
		}
		if err := st.CompleteRun(ctx, handle, CompleteRunParams{
			CompletedAt: base.Add(offset),
			Analysis:    analysis,
		}); err != nil {
			t.Fatalf("CompleteRun: %v", err)
		}
	}

	dockerSig := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Detector: "log",
		Evidence: []string{"authentication required for registry.example.com"},
	}).Hash
	dockerAnalysis := &model.Analysis{
		Source:      "stdin",
		InputHash:   "docker-input",
		OutputHash:  "docker-output",
		Fingerprint: "docker-fp",
		Results: []model.Result{{
			Playbook: model.Playbook{
				ID:       "docker-auth",
				Title:    "Docker auth",
				Category: "auth",
			},
			Detector:      "log",
			Score:         4.5,
			Confidence:    0.91,
			Evidence:      []string{"authentication required for registry.example.com"},
			SignatureHash: dockerSig,
		}},
	}
	writeRun(0, dockerAnalysis)
	secondDocker := *dockerAnalysis
	secondDocker.OutputHash = "docker-output"
	writeRun(2*time.Minute, &secondDocker)

	timeoutSig := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "network-timeout"},
		Detector: "log",
		Evidence: []string{"context deadline exceeded while waiting for upstream"},
	}).Hash
	timeoutAnalysis := &model.Analysis{
		Source:      "stdin",
		InputHash:   "timeout-input",
		OutputHash:  "timeout-output",
		Fingerprint: "timeout-fp",
		Results: []model.Result{{
			Playbook: model.Playbook{
				ID:       "network-timeout",
				Title:    "Network timeout",
				Category: "network",
			},
			Detector:      "log",
			Score:         3.9,
			Confidence:    0.74,
			Evidence:      []string{"context deadline exceeded while waiting for upstream"},
			SignatureHash: timeoutSig,
		}},
	}
	writeRun(4*time.Minute, timeoutAnalysis)

	signatures, err := st.ListSignatures(ctx, 10)
	if err != nil {
		t.Fatalf("ListSignatures: %v", err)
	}
	if len(signatures) < 2 {
		t.Fatalf("expected at least two signatures, got %#v", signatures)
	}
	if signatures[0].FailureID != "docker-auth" || signatures[0].OccurrenceCount != 2 {
		t.Fatalf("expected docker-auth signature first, got %#v", signatures[0])
	}

	playbooks, err := st.ListPlaybookStats(ctx, 10)
	if err != nil {
		t.Fatalf("ListPlaybookStats: %v", err)
	}
	if len(playbooks) < 2 {
		t.Fatalf("expected at least two playbook stats, got %#v", playbooks)
	}
	if playbooks[0].FailureID != "docker-auth" || playbooks[0].SelectedCount != 2 || playbooks[0].RecurringRunCount != 2 {
		t.Fatalf("unexpected docker-auth playbook stats: %#v", playbooks[0])
	}
}

func TestSQLiteStorePlaybookStatsIncludeMatchCountsAndAverageRank(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	base := time.Date(2026, 4, 23, 11, 0, 0, 0, time.UTC)

	writeRun := func(offset time.Duration, analysis *model.Analysis) {
		handle, err := st.BeginRun(ctx, BeginRunParams{
			Surface:    "analyze",
			SourceKind: "log",
			Source:     "stdin",
			InputHash:  analysis.InputHash,
			StartedAt:  base.Add(offset),
		})
		if err != nil {
			t.Fatalf("BeginRun: %v", err)
		}
		if err := st.CompleteRun(ctx, handle, CompleteRunParams{
			CompletedAt: base.Add(offset),
			Analysis:    analysis,
		}); err != nil {
			t.Fatalf("CompleteRun: %v", err)
		}
	}

	dockerSig := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Detector: "log",
		Evidence: []string{"authentication required for registry.example.com"},
	}).Hash
	timeoutSig := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "network-timeout"},
		Detector: "log",
		Evidence: []string{"context deadline exceeded while waiting for upstream"},
	}).Hash

	writeRun(0, &model.Analysis{
		Source:      "stdin",
		InputHash:   "analysis-1",
		OutputHash:  "output-1",
		Fingerprint: "fp-1",
		Results: []model.Result{
			{
				Playbook:      model.Playbook{ID: "docker-auth", Title: "Docker auth", Category: "auth"},
				Detector:      "log",
				Score:         4.4,
				Confidence:    0.91,
				Evidence:      []string{"authentication required for registry.example.com"},
				SignatureHash: dockerSig,
			},
			{
				Playbook:   model.Playbook{ID: "network-timeout", Title: "Network timeout", Category: "network"},
				Detector:   "log",
				Score:      3.3,
				Confidence: 0.67,
				Evidence:   []string{"context deadline exceeded while waiting for upstream"},
			},
		},
	})

	writeRun(2*time.Minute, &model.Analysis{
		Source:      "stdin",
		InputHash:   "analysis-2",
		OutputHash:  "output-2",
		Fingerprint: "fp-2",
		Results: []model.Result{
			{
				Playbook:      model.Playbook{ID: "network-timeout", Title: "Network timeout", Category: "network"},
				Detector:      "log",
				Score:         3.9,
				Confidence:    0.79,
				Evidence:      []string{"context deadline exceeded while waiting for upstream"},
				SignatureHash: timeoutSig,
			},
		},
	})

	playbooks, err := st.ListPlaybookStats(ctx, 10)
	if err != nil {
		t.Fatalf("ListPlaybookStats: %v", err)
	}

	var timeoutStats PlaybookStats
	found := false
	for _, item := range playbooks {
		if item.FailureID == "network-timeout" {
			timeoutStats = item
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected network-timeout playbook stats, got %#v", playbooks)
	}
	if timeoutStats.SelectedCount != 1 {
		t.Fatalf("expected selected_count=1, got %#v", timeoutStats)
	}
	if timeoutStats.MatchedCount != 2 {
		t.Fatalf("expected matched_count=2, got %#v", timeoutStats)
	}
	if timeoutStats.NonSelectedCount != 1 {
		t.Fatalf("expected non_selected_count=1, got %#v", timeoutStats)
	}
	if timeoutStats.AvgRank != 1.5 {
		t.Fatalf("expected avg_rank=1.5, got %#v", timeoutStats)
	}
}

func TestSQLiteStoreListsFailureReports(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faultline.db")
	st, _, err := OpenBestEffort(Config{Mode: ModeAuto, Path: path})
	if err != nil {
		t.Fatalf("OpenBestEffort: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	base := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)

	writeRun := func(offset time.Duration, analysis *model.Analysis) {
		handle, err := st.BeginRun(ctx, BeginRunParams{
			Surface:    "analyze",
			SourceKind: "log",
			Source:     analysis.Source,
			InputHash:  analysis.InputHash,
			StartedAt:  base.Add(offset),
		})
		if err != nil {
			t.Fatalf("BeginRun: %v", err)
		}
		if err := st.CompleteRun(ctx, handle, CompleteRunParams{
			CompletedAt: base.Add(offset),
			Analysis:    analysis,
		}); err != nil {
			t.Fatalf("CompleteRun: %v", err)
		}
	}

	writeRun(0, &model.Analysis{
		Source:      "build-1.log",
		InputHash:   "input-1",
		OutputHash:  "output-1",
		Fingerprint: "fp-1",
		Results: []model.Result{{
			Playbook:      model.Playbook{ID: "docker-auth", Title: "Docker auth", Category: "auth"},
			Detector:      "log",
			Score:         4.5,
			Confidence:    0.91,
			Evidence:      []string{"authentication required\nsecond line is not retained"},
			SignatureHash: "sig-1",
		}},
	})
	writeRun(time.Minute, &model.Analysis{
		Source:      "build-2.log",
		InputHash:   "input-2",
		OutputHash:  "output-2",
		Fingerprint: "fp-2",
		Results: []model.Result{{
			Playbook:      model.Playbook{ID: "network-timeout", Title: "Network timeout", Category: "network"},
			Detector:      "log",
			Score:         3.5,
			Confidence:    0.8,
			Evidence:      []string{"context deadline exceeded"},
			SignatureHash: "sig-2",
		}},
	})
	writeRun(2*time.Minute, &model.Analysis{
		Source:      "build-3.log",
		InputHash:   "input-3",
		OutputHash:  "output-3",
		Fingerprint: "fp-3",
		Results: []model.Result{{
			Playbook:      model.Playbook{ID: "docker-auth", Title: "Docker auth", Category: "auth"},
			Detector:      "log",
			Score:         4.6,
			Confidence:    0.93,
			Evidence:      []string{"pull access denied for registry.example.com"},
			SignatureHash: "sig-3",
		}},
	})

	reports, err := st.ListFailureReports(ctx, 0)
	if err != nil {
		t.Fatalf("ListFailureReports: %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("expected two report rows, got %#v", reports)
	}
	if reports[0].FailureID != "docker-auth" || reports[0].Count != 2 {
		t.Fatalf("expected docker-auth count first, got %#v", reports[0])
	}
	if reports[0].LastSeenAt != base.Add(2*time.Minute).Format(time.RFC3339) {
		t.Fatalf("expected latest docker-auth timestamp, got %#v", reports[0])
	}
	if reports[0].ExampleEvidence != "pull access denied for registry.example.com" {
		t.Fatalf("expected latest first evidence line, got %#v", reports[0])
	}
}

func boolPtr(value bool) *bool {
	return &value
}

// ── nullableBool ──────────────────────────────────────────────────────────────

func TestNullableBoolNilPointer(t *testing.T) {
	got := nullableBool(nil)
	if got != nil {
		t.Errorf("expected nil for nil pointer, got %v", got)
	}
}

func TestNullableBoolTrue(t *testing.T) {
	v := true
	got := nullableBool(&v)
	if got != 1 {
		t.Errorf("expected 1 for true, got %v", got)
	}
}

func TestNullableBoolFalse(t *testing.T) {
	v := false
	got := nullableBool(&v)
	if got != 0 {
		t.Errorf("expected 0 for false, got %v", got)
	}
}
