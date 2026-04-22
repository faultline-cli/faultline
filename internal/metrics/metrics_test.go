package metrics

import (
	"encoding/json"
	"strings"
	"testing"

	"faultline/internal/model"
)

// ── TSS (FromLocalHistory) ────────────────────────────────────────────────────

func TestFromLocalHistory_TooFewEntries(t *testing.T) {
	got := FromLocalHistory("docker-auth", []LocalEntry{
		{FailureID: "docker-auth"},
	})
	if got != nil {
		t.Errorf("expected nil with < %d entries, got %+v", minTSSEntries, got)
	}
}

func TestFromLocalHistory_NilOnEmpty(t *testing.T) {
	got := FromLocalHistory("docker-auth", nil)
	if got != nil {
		t.Errorf("expected nil with no entries, got %+v", got)
	}
}

func TestFromLocalHistory_AllSame(t *testing.T) {
	entries := []LocalEntry{
		{FailureID: "docker-auth"},
		{FailureID: "docker-auth"},
		{FailureID: "docker-auth"},
	}
	got := FromLocalHistory("docker-auth", entries)
	if got == nil {
		t.Fatal("expected non-nil Metrics")
	}
	if got.TSS == nil {
		t.Fatal("expected TSS to be set")
	}
	if *got.TSS != 1.0 {
		t.Errorf("TSS = %.2f, want 1.00", *got.TSS)
	}
	if got.HistoryCount != 3 {
		t.Errorf("HistoryCount = %d, want 3", got.HistoryCount)
	}
}

func TestFromLocalHistory_Rounding(t *testing.T) {
	// 1 matching out of 3 → 0.33
	entries := []LocalEntry{
		{FailureID: "docker-auth"},
		{FailureID: "go-build"},
		{FailureID: "npm-install"},
	}
	got := FromLocalHistory("docker-auth", entries)
	if got == nil || got.TSS == nil {
		t.Fatal("expected non-nil Metrics and TSS")
	}
	if *got.TSS != 0.33 {
		t.Errorf("TSS = %.2f, want 0.33", *got.TSS)
	}
}

func TestFromLocalHistory_Zero(t *testing.T) {
	// current failure ID never seen → TSS = 0
	entries := []LocalEntry{
		{FailureID: "go-build"},
		{FailureID: "npm-install"},
		{FailureID: "go-build"},
	}
	got := FromLocalHistory("docker-auth", entries)
	if got == nil || got.TSS == nil {
		t.Fatal("expected non-nil Metrics and TSS")
	}
	if *got.TSS != 0.0 {
		t.Errorf("TSS = %.2f, want 0.00", *got.TSS)
	}
}

func TestFromLocalHistory_DriftPersistent(t *testing.T) {
	// TSS ≥ 0.5 → persistent drift component
	entries := []LocalEntry{
		{FailureID: "docker-auth"},
		{FailureID: "docker-auth"},
		{FailureID: "docker-auth"},
		{FailureID: "go-build"},
	}
	got := FromLocalHistory("docker-auth", entries)
	if got == nil {
		t.Fatal("expected non-nil Metrics")
	}
	found := false
	for _, d := range got.DriftComponents {
		if strings.Contains(d, "persistent failure") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected persistent-failure drift component, got %v", got.DriftComponents)
	}
}

// ── FPC / PHI (WithExplicitHistory) ──────────────────────────────────────────

func TestWithExplicitHistory_TooFewEntries(t *testing.T) {
	entries := []model.MetricsHistoryEntry{
		{Matched: true, FailureID: "docker-auth"},
		{Matched: false},
	}
	m := WithExplicitHistory(nil, entries)
	if m == nil {
		t.Fatal("expected non-nil Metrics")
	}
	if m.FPC != nil {
		t.Errorf("FPC should be nil with < %d entries", minFPCEntries)
	}
	if m.PHI != nil {
		t.Errorf("PHI should be nil with < %d entries", minFPCEntries)
	}
}

func TestWithExplicitHistory_FPCOnly(t *testing.T) {
	// 3 entries → FPC computed; < 5 → PHI absent
	entries := []model.MetricsHistoryEntry{
		{Matched: true, FailureID: "docker-auth"},
		{Matched: true, FailureID: "go-build"},
		{Matched: false},
	}
	m := WithExplicitHistory(nil, entries)
	if m.FPC == nil {
		t.Fatal("expected FPC to be set")
	}
	if *m.FPC != 0.67 {
		t.Errorf("FPC = %.2f, want 0.67", *m.FPC)
	}
	if m.PHI != nil {
		t.Errorf("PHI should be nil with < %d entries", minPHIEntries)
	}
}

func TestWithExplicitHistory_PHIComputed(t *testing.T) {
	// 5 entries: 3 matched same ID → dominant_share = 1.0, PHI = FPC*(1-1) = 0
	entries := []model.MetricsHistoryEntry{
		{Matched: true, FailureID: "docker-auth"},
		{Matched: true, FailureID: "docker-auth"},
		{Matched: true, FailureID: "docker-auth"},
		{Matched: false},
		{Matched: false},
	}
	m := WithExplicitHistory(nil, entries)
	if m.FPC == nil {
		t.Fatal("expected FPC to be set")
	}
	if *m.FPC != 0.6 {
		t.Errorf("FPC = %.2f, want 0.60", *m.FPC)
	}
	if m.PHI == nil {
		t.Fatal("expected PHI to be set")
	}
	if *m.PHI != 0.0 {
		t.Errorf("PHI = %.2f, want 0.00 (single dominant failure)", *m.PHI)
	}
}

func TestWithExplicitHistory_PHISpread(t *testing.T) {
	// 5 matched entries, all different IDs → dominant_share = 0.2,
	// PHI = FPC*(1-0.2) = 1.0*0.8 = 0.80
	entries := []model.MetricsHistoryEntry{
		{Matched: true, FailureID: "docker-auth"},
		{Matched: true, FailureID: "go-build"},
		{Matched: true, FailureID: "npm-install"},
		{Matched: true, FailureID: "tls-cert"},
		{Matched: true, FailureID: "python-deps"},
	}
	m := WithExplicitHistory(nil, entries)
	if m.FPC == nil || *m.FPC != 1.0 {
		t.Errorf("FPC = %v, want 1.0", m.FPC)
	}
	if m.PHI == nil {
		t.Fatal("expected PHI to be set")
	}
	if *m.PHI != 0.8 {
		t.Errorf("PHI = %.2f, want 0.80", *m.PHI)
	}
}

func TestWithExplicitHistory_PreservesExistingTSS(t *testing.T) {
	tss := 0.75
	base := &model.Metrics{TSS: &tss, HistoryCount: 4}
	entries := make([]model.MetricsHistoryEntry, 5)
	for i := range entries {
		entries[i] = model.MetricsHistoryEntry{Matched: true, FailureID: "go-build"}
	}
	m := WithExplicitHistory(base, entries)
	if m.TSS == nil || *m.TSS != 0.75 {
		t.Errorf("TSS should be preserved; got %v", m.TSS)
	}
	if m.HistoryCount != 4 {
		t.Errorf("HistoryCount should be preserved; got %d", m.HistoryCount)
	}
}

func TestWithExplicitHistory_MissingData(t *testing.T) {
	// Zero entries → no FPC, no PHI; Metrics not nil but fields absent
	m := WithExplicitHistory(nil, nil)
	if m == nil {
		t.Fatal("expected non-nil Metrics")
	}
	if m.FPC != nil || m.PHI != nil {
		t.Errorf("FPC and PHI should be nil when no data, got FPC=%v PHI=%v", m.FPC, m.PHI)
	}
}

// ── JSON serialisation ────────────────────────────────────────────────────────

func TestMetricsJSON_OmitAbsent(t *testing.T) {
	// When no explicit history is supplied, PHI and FPC must be absent from JSON.
	tss := 0.5
	m := model.Metrics{TSS: &tss, HistoryCount: 4}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if strings.Contains(got, "fpc") {
		t.Errorf("fpc should be absent from JSON; got %s", got)
	}
	if strings.Contains(got, "phi") {
		t.Errorf("phi should be absent from JSON; got %s", got)
	}
	if !strings.Contains(got, "tss") {
		t.Errorf("tss should be present in JSON; got %s", got)
	}
}

func TestMetricsJSON_AllFields(t *testing.T) {
	tss := 0.5
	fpc := 0.8
	phi := 0.64
	m := model.Metrics{
		TSS:             &tss,
		FPC:             &fpc,
		PHI:             &phi,
		HistoryCount:    10,
		DriftComponents: []string{"low coverage: 3 of 15 runs unmatched"},
	}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{"tss", "fpc", "phi", "history_count", "drift_components"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in JSON; got %s", want, got)
		}
	}
}

// ── LoadHistoryFile ───────────────────────────────────────────────────────────

func TestLoadHistoryFile_Missing(t *testing.T) {
	entries, err := LoadHistoryFile("/no/such/file.jsonl")
	if err != nil {
		t.Errorf("expected nil error for missing file, got %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty entries for missing file, got %d", len(entries))
	}
}

func TestLoadHistoryFile_SkipMalformedLines(t *testing.T) {
	input := strings.NewReader(`{"matched":true,"failure_id":"docker-auth"}
not-json
{"matched":false}
`)
	entries, err := decodeHistoryEntries(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 valid entries (malformed skipped), got %d", len(entries))
	}
}

// ── Round2 ────────────────────────────────────────────────────────────────────

func TestRound2(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{0.333333, 0.33},
		{0.666666, 0.67},
		{1.0, 1.0},
		{0.0, 0.0},
		{0.005, 0.01},
	}
	for _, tc := range cases {
		got := round2(tc.in)
		if got != tc.want {
			t.Errorf("round2(%.6f) = %.2f, want %.2f", tc.in, got, tc.want)
		}
	}
}
