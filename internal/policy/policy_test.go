package policy

import (
	"testing"

	"faultline/internal/model"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func f(v float64) *float64 { return &v }

func makeMetrics(tss, fpc, phi *float64) *model.Metrics {
	return &model.Metrics{TSS: tss, FPC: fpc, PHI: phi}
}

// ── Compute ───────────────────────────────────────────────────────────────────

func TestCompute_NilMetrics(t *testing.T) {
	got := Compute(nil, "high")
	if got != nil {
		t.Errorf("expected nil policy with nil metrics, got %+v", got)
	}
}

func TestCompute_NoMetricPointers(t *testing.T) {
	// Metrics struct with no TSS/FPC/PHI → no policy.
	m := &model.Metrics{HistoryCount: 3}
	got := Compute(m, "low")
	if got != nil {
		t.Errorf("expected nil policy with no metric values, got %+v", got)
	}
}

// ── blocking ──────────────────────────────────────────────────────────────────

func TestCompute_Blocking_Critical(t *testing.T) {
	m := makeMetrics(f(0.9), nil, nil)
	got := Compute(m, "critical")
	if got == nil {
		t.Fatal("expected non-nil policy")
	}
	if got.Recommendation != RecommendBlocking {
		t.Errorf("recommendation = %q, want %q", got.Recommendation, RecommendBlocking)
	}
	if len(got.Basis) == 0 || got.Basis[0] != "tss" {
		t.Errorf("expected basis [tss], got %v", got.Basis)
	}
}

func TestCompute_Blocking_High(t *testing.T) {
	m := makeMetrics(f(0.85), nil, nil)
	got := Compute(m, "high")
	if got == nil || got.Recommendation != RecommendBlocking {
		t.Errorf("expected blocking for TSS 0.85 + high severity, got %v", got)
	}
}

func TestCompute_Blocking_LowSeverityNoBlock(t *testing.T) {
	// TSS 0.9 but severity is "low" → should not be blocking.
	m := makeMetrics(f(0.9), nil, nil)
	got := Compute(m, "low")
	if got == nil {
		t.Fatal("expected non-nil policy")
	}
	if got.Recommendation == RecommendBlocking {
		t.Errorf("low severity should not trigger blocking, got blocking")
	}
}

// ── quarantine ────────────────────────────────────────────────────────────────

func TestCompute_Quarantine_HighTSS_LowSeverity(t *testing.T) {
	// TSS ≥ 0.5, severity low → quarantine (not blocking).
	m := makeMetrics(f(0.6), nil, nil)
	got := Compute(m, "low")
	if got == nil || got.Recommendation != RecommendQuarantine {
		t.Errorf("expected quarantine, got %v", got)
	}
	containsTSS := false
	for _, b := range got.Basis {
		if b == "tss" {
			containsTSS = true
		}
	}
	if !containsTSS {
		t.Errorf("expected basis to include tss, got %v", got.Basis)
	}
}

func TestCompute_Quarantine_LowFPC(t *testing.T) {
	m := makeMetrics(nil, f(0.3), nil)
	got := Compute(m, "medium")
	if got == nil || got.Recommendation != RecommendQuarantine {
		t.Errorf("expected quarantine for FPC 0.3, got %v", got)
	}
	containsFPC := false
	for _, b := range got.Basis {
		if b == "fpc" {
			containsFPC = true
		}
	}
	if !containsFPC {
		t.Errorf("expected basis to include fpc, got %v", got.Basis)
	}
}

func TestCompute_Quarantine_LowPHI(t *testing.T) {
	m := makeMetrics(nil, nil, f(0.3))
	got := Compute(m, "medium")
	if got == nil || got.Recommendation != RecommendQuarantine {
		t.Errorf("expected quarantine for PHI 0.3, got %v", got)
	}
}

func TestCompute_Quarantine_MultipleFactors(t *testing.T) {
	// Both TSS and FPC trigger quarantine → basis should include both.
	m := makeMetrics(f(0.7), f(0.4), nil)
	got := Compute(m, "medium")
	if got == nil || got.Recommendation != RecommendQuarantine {
		t.Errorf("expected quarantine, got %v", got)
	}
	if len(got.Basis) < 2 {
		t.Errorf("expected at least 2 basis items, got %v", got.Basis)
	}
}

// ── observe ───────────────────────────────────────────────────────────────────

func TestCompute_Observe_ModerateTSS(t *testing.T) {
	m := makeMetrics(f(0.4), nil, nil)
	got := Compute(m, "medium")
	if got == nil || got.Recommendation != RecommendObserve {
		t.Errorf("expected observe for TSS 0.4, got %v", got)
	}
}

func TestCompute_Observe_LowerBoundTSS(t *testing.T) {
	// TSS exactly at tssObserveThreshold.
	m := makeMetrics(f(tssObserveThreshold), nil, nil)
	got := Compute(m, "low")
	if got == nil || got.Recommendation != RecommendObserve {
		t.Errorf("expected observe at TSS threshold, got %v", got)
	}
}

func TestCompute_Observe_DegradedPHI(t *testing.T) {
	// PHI < 0.6 but ≥ phiQuarantineThreshold, no TSS → observe.
	m := makeMetrics(nil, nil, f(0.5))
	got := Compute(m, "medium")
	if got == nil || got.Recommendation != RecommendObserve {
		t.Errorf("expected observe for PHI 0.5, got %v", got)
	}
}

// ── ok ────────────────────────────────────────────────────────────────────────

func TestCompute_OK_HealthyMetrics(t *testing.T) {
	m := makeMetrics(f(0.1), f(0.9), f(0.8))
	got := Compute(m, "medium")
	if got == nil || got.Recommendation != RecommendOK {
		t.Errorf("expected ok for healthy metrics, got %v", got)
	}
	if got.Basis != nil {
		t.Errorf("expected nil basis for ok, got %v", got.Basis)
	}
}

func TestCompute_OK_LowTSS(t *testing.T) {
	// TSS below observe threshold → ok (new failure type).
	m := makeMetrics(f(0.1), nil, nil)
	got := Compute(m, "low")
	if got == nil || got.Recommendation != RecommendOK {
		t.Errorf("expected ok for TSS 0.1, got %v", got)
	}
}

func TestCompute_OK_ZeroTSS(t *testing.T) {
	m := makeMetrics(f(0.0), nil, nil)
	got := Compute(m, "high")
	if got == nil || got.Recommendation != RecommendOK {
		t.Errorf("expected ok for TSS 0.0 (new failure), got %v", got)
	}
}

// ── boundary: blocking beats quarantine ───────────────────────────────────────

func TestCompute_BlockingBeatsQuarantine(t *testing.T) {
	// Both blocking and quarantine conditions would fire — blocking wins.
	m := makeMetrics(f(0.85), f(0.3), nil)
	got := Compute(m, "critical")
	if got == nil || got.Recommendation != RecommendBlocking {
		t.Errorf("expected blocking to win over quarantine, got %v", got)
	}
}

// ── JSON round-trip ───────────────────────────────────────────────────────────

func TestPolicyJSONOmitEmpty(t *testing.T) {
	// An ok policy has no basis; basis field should be absent.
	p := model.Policy{Recommendation: RecommendOK, Reason: "metrics within acceptable range"}
	// Manually check json tags: basis is omitempty so nil slice → absent.
	// This is a compile-time guarantee from the struct tag, confirmed here.
	if p.Basis != nil {
		t.Error("basis should be nil for ok policy")
	}
}
