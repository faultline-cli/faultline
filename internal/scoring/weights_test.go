package scoring

import (
	"testing"
)

// ── defaultWeights ────────────────────────────────────────────────────────────

func TestDefaultWeightsLoadsSuccessfully(t *testing.T) {
	w, err := defaultWeights()
	if err != nil {
		t.Fatalf("defaultWeights() returned error: %v", err)
	}
	if w.PriorSmoothing <= 0 {
		t.Errorf("PriorSmoothing = %f, want > 0", w.PriorSmoothing)
	}
	if w.FeatureWeights == nil {
		t.Error("FeatureWeights is nil, want non-nil map")
	}
	if w.PlaybookCounts == nil {
		t.Error("PlaybookCounts is nil, want non-nil map")
	}
}

func TestDefaultWeightsIsDeterministic(t *testing.T) {
	w1, err1 := defaultWeights()
	w2, err2 := defaultWeights()
	if err1 != nil || err2 != nil {
		t.Fatalf("defaultWeights errors: %v, %v", err1, err2)
	}
	if w1.PriorSmoothing != w2.PriorSmoothing {
		t.Errorf("PriorSmoothing is non-deterministic: %f != %f", w1.PriorSmoothing, w2.PriorSmoothing)
	}
	if len(w1.FeatureWeights) != len(w2.FeatureWeights) {
		t.Errorf("FeatureWeights lengths differ: %d != %d", len(w1.FeatureWeights), len(w2.FeatureWeights))
	}
	// Spot-check known stable keys from the embedded bayes_v1.json.
	knownKeys := []string{"detector_score", "detector_confidence", "candidate_separation"}
	for _, key := range knownKeys {
		v1, ok1 := w1.FeatureWeights[key]
		v2, ok2 := w2.FeatureWeights[key]
		if !ok1 || !ok2 {
			t.Errorf("expected feature weight key %q to exist in both calls", key)
			continue
		}
		if v1 != v2 {
			t.Errorf("feature weight %q differs between calls: %f != %f", key, v1, v2)
		}
	}
}
