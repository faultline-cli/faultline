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
}
