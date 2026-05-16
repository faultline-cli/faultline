package scoring

import (
	"testing"

	"faultline/internal/model"
)

// ── deltaPlaybookFeatures ─────────────────────────────────────────────────────

func TestDeltaPlaybookFeaturesNotRequested(t *testing.T) {
	inputs := Inputs{DeltaRequested: false}
	result := model.Result{
		Playbook: model.Playbook{
			ID:            "dep-drift",
			RequiresDelta: true,
		},
	}
	got := deltaPlaybookFeatures(inputs, result, nil)
	if got != nil {
		t.Errorf("expected nil when DeltaRequested=false, got %v", got)
	}
}

func TestDeltaPlaybookFeaturesRequiresDeltaMissing(t *testing.T) {
	inputs := Inputs{DeltaRequested: true}
	result := model.Result{
		Playbook: model.Playbook{
			ID:            "dep-drift",
			RequiresDelta: true,
		},
	}
	got := deltaPlaybookFeatures(inputs, result, nil)
	if len(got) != 1 {
		t.Fatalf("expected 1 feature for missing delta, got %d: %v", len(got), got)
	}
	if got[0].Name != "delta_required_missing" {
		t.Errorf("expected delta_required_missing, got %q", got[0].Name)
	}
	if got[0].Weight != -1.5 {
		t.Errorf("expected weight -1.5, got %f", got[0].Weight)
	}
}

func TestDeltaPlaybookFeaturesRequiresDeltaUnmatched(t *testing.T) {
	inputs := Inputs{DeltaRequested: true}
	result := model.Result{
		Playbook: model.Playbook{
			ID:            "dep-drift",
			RequiresDelta: true,
			DeltaBoost: []model.DeltaBoost{
				{Signal: "dependency_change", Weight: 1.5},
			},
		},
	}
	// Provide a delta with a different signal that doesn't match.
	delta := &model.Delta{
		Signals: []model.DeltaSignal{
			{ID: "ci_config_change", Detail: "modified .github/workflows/ci.yml"},
		},
	}
	got := deltaPlaybookFeatures(inputs, result, delta)
	// Should include delta_required_unmatched because RequiresDelta=true with DeltaBoost
	// but no matching signal.
	foundUnmatched := false
	for _, f := range got {
		if f.Name == "delta_required_unmatched" {
			foundUnmatched = true
			if f.Weight != -0.75 {
				t.Errorf("expected delta_required_unmatched weight -0.75, got %f", f.Weight)
			}
		}
	}
	if !foundUnmatched {
		t.Errorf("expected delta_required_unmatched feature, got %v", got)
	}
}

func TestDeltaPlaybookFeaturesBoostSignalMatched(t *testing.T) {
	inputs := Inputs{DeltaRequested: true}
	result := model.Result{
		Playbook: model.Playbook{
			ID: "dep-drift",
			DeltaBoost: []model.DeltaBoost{
				{Signal: "dependency_change", Weight: 1.5},
			},
		},
	}
	delta := &model.Delta{
		Signals: []model.DeltaSignal{
			{ID: "dependency_change", Detail: "go.sum was modified"},
		},
	}
	got := deltaPlaybookFeatures(inputs, result, delta)
	if len(got) == 0 {
		t.Fatal("expected at least one delta boost feature")
	}
	found := false
	for _, f := range got {
		if f.Name == "delta_boost:dependency_change" {
			found = true
			if f.Value != 1 {
				t.Errorf("expected value 1, got %f", f.Value)
			}
			if len(f.EvidenceRefs) == 0 || f.EvidenceRefs[0] != "go.sum was modified" {
				t.Errorf("expected evidence ref 'go.sum was modified', got %v", f.EvidenceRefs)
			}
		}
	}
	if !found {
		t.Errorf("expected delta_boost:dependency_change feature, got %v", got)
	}
}

func TestDeltaPlaybookFeaturesBoostDefaultWeight(t *testing.T) {
	inputs := Inputs{DeltaRequested: true}
	result := model.Result{
		Playbook: model.Playbook{
			ID: "dep-drift",
			DeltaBoost: []model.DeltaBoost{
				{Signal: "runtime_toolchain_change", Weight: 0}, // zero weight → defaults to 1
			},
		},
	}
	delta := &model.Delta{
		Signals: []model.DeltaSignal{
			{ID: "runtime_toolchain_change", Detail: "go version changed"},
		},
	}
	got := deltaPlaybookFeatures(inputs, result, delta)
	found := false
	for _, f := range got {
		if f.Name == "delta_boost:runtime_toolchain_change" {
			found = true
			if f.Weight != 1 {
				t.Errorf("expected default weight 1 for zero-weight boost, got %f", f.Weight)
			}
		}
	}
	if !found {
		t.Errorf("expected delta_boost:runtime_toolchain_change, got %v", got)
	}
}

func TestDeltaPlaybookFeaturesEmptySignal(t *testing.T) {
	inputs := Inputs{DeltaRequested: true}
	result := model.Result{
		Playbook: model.Playbook{
			ID: "dep-drift",
			DeltaBoost: []model.DeltaBoost{
				{Signal: "   ", Weight: 1.0}, // blank signal should be skipped
			},
		},
	}
	delta := &model.Delta{}
	got := deltaPlaybookFeatures(inputs, result, delta)
	if len(got) != 0 {
		t.Errorf("expected no features for blank signal, got %v", got)
	}
}

func TestDeltaPlaybookFeaturesNoBoostSignalNoMatch(t *testing.T) {
	// DeltaBoost signal not present in delta signals → no boost feature emitted.
	inputs := Inputs{DeltaRequested: true}
	result := model.Result{
		Playbook: model.Playbook{
			ID: "dep-drift",
			DeltaBoost: []model.DeltaBoost{
				{Signal: "ci_config_change", Weight: 1.0},
			},
		},
	}
	delta := &model.Delta{
		Signals: []model.DeltaSignal{
			{ID: "dependency_change", Detail: "go.sum modified"},
		},
	}
	got := deltaPlaybookFeatures(inputs, result, delta)
	// No matching signal, no RequiresDelta → no features.
	if len(got) != 0 {
		t.Errorf("expected no features when boost signal not in delta, got %v", got)
	}
}
