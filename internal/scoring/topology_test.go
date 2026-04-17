package scoring

import (
	"testing"

	"faultline/internal/model"
)

func TestTopologySignalSetNilState(t *testing.T) {
	if got := topologySignalSet(nil); got != nil {
		t.Fatalf("expected nil for nil state, got %#v", got)
	}
}

func TestTopologySignalSetEmptySignals(t *testing.T) {
	state := &RepoState{TopologySignals: nil}
	if got := topologySignalSet(state); got != nil {
		t.Fatalf("expected nil for empty signals, got %#v", got)
	}
}

func TestTopologySignalSetPopulatesMap(t *testing.T) {
	state := &RepoState{
		TopologySignals: []string{"topo.boundary.crossed", " topo.owner.zone.frontend "},
	}
	got := topologySignalSet(state)
	if got == nil {
		t.Fatal("expected non-nil map")
	}
	if _, ok := got["topo.boundary.crossed"]; !ok {
		t.Error("expected topo.boundary.crossed in signal set")
	}
	if _, ok := got["topo.owner.zone.frontend"]; !ok {
		t.Error("expected trimmed topo.owner.zone.frontend in signal set")
	}
}

func TestTopologyPlaybookFeaturesNotAppliedWithoutTopologyConfig(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{
			ID:               "no-topology",
			RequiresTopology: false,
			TopologyBoost:    nil,
		},
	}
	feats := topologyPlaybookFeatures(Inputs{}, result)
	if len(feats) != 0 {
		t.Fatalf("expected no features for playbook without topology config, got %#v", feats)
	}
}

func TestTopologyPlaybookFeaturesPenalisesRequiredWhenMissing(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{
			ID:               "topology-required",
			RequiresTopology: true,
		},
	}
	feats := topologyPlaybookFeatures(Inputs{RepoState: nil}, result)
	if len(feats) == 0 {
		t.Fatal("expected penalty feature when topology is required but absent")
	}
	if feats[0].Weight >= 0 {
		t.Fatalf("expected negative weight for missing required topology, got %.2f", feats[0].Weight)
	}
}

func TestTopologyPlaybookFeaturesBoostsMatchingSignals(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{
			ID: "with-boost",
			TopologyBoost: []model.TopologyBoost{
				{Signal: "topo.boundary.crossed", Weight: 1.5},
				{Signal: "topo.missing.signal", Weight: 0.8},
			},
		},
	}
	inputs := Inputs{
		RepoState: &RepoState{
			TopologySignals: []string{"topo.boundary.crossed"},
		},
	}
	feats := topologyPlaybookFeatures(inputs, result)
	if len(feats) != 1 {
		t.Fatalf("expected 1 boost feature for matched signal, got %d: %#v", len(feats), feats)
	}
	if feats[0].Weight != 1.5 {
		t.Fatalf("expected weight 1.5, got %.2f", feats[0].Weight)
	}
}

func TestTopologyPlaybookFeaturesUsesDefaultWeightForZero(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{
			ID: "zero-weight",
			TopologyBoost: []model.TopologyBoost{
				{Signal: "topo.active", Weight: 0},
			},
		},
	}
	inputs := Inputs{
		RepoState: &RepoState{
			TopologySignals: []string{"topo.active"},
		},
	}
	feats := topologyPlaybookFeatures(inputs, result)
	if len(feats) != 1 {
		t.Fatalf("expected 1 feature, got %d", len(feats))
	}
	if feats[0].Weight != 1.0 {
		t.Fatalf("expected default weight 1.0 for zero weight, got %.2f", feats[0].Weight)
	}
}
