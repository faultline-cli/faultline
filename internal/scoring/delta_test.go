package scoring

import (
	"testing"

	"faultline/internal/model"
)

// ── appendReason ──────────────────────────────────────────────────────────────

func TestAppendReasonAddsNewReason(t *testing.T) {
	b := &deltaBucket{}
	appendReason(b, "dependency file changed")
	if len(b.reasons) != 1 || b.reasons[0] != "dependency file changed" {
		t.Errorf("expected [dependency file changed], got %v", b.reasons)
	}
}

func TestAppendReasonSkipsEmptyString(t *testing.T) {
	b := &deltaBucket{}
	appendReason(b, "")
	if len(b.reasons) != 0 {
		t.Errorf("expected empty reasons, got %v", b.reasons)
	}
}

func TestAppendReasonSkipsWhitespaceOnlyString(t *testing.T) {
	b := &deltaBucket{}
	appendReason(b, "   ")
	if len(b.reasons) != 0 {
		t.Errorf("expected empty reasons after whitespace-only input, got %v", b.reasons)
	}
}

func TestAppendReasonDeduplicatesExisting(t *testing.T) {
	b := &deltaBucket{reasons: []string{"already here"}}
	appendReason(b, "already here")
	if len(b.reasons) != 1 {
		t.Errorf("expected 1 reason after dedup, got %v", b.reasons)
	}
}

func TestAppendReasonSortsReasons(t *testing.T) {
	b := &deltaBucket{}
	appendReason(b, "zebra")
	appendReason(b, "alpha")
	appendReason(b, "middle")
	if b.reasons[0] != "alpha" || b.reasons[1] != "middle" || b.reasons[2] != "zebra" {
		t.Errorf("expected sorted reasons, got %v", b.reasons)
	}
}

// ── cloneEnvDiff ──────────────────────────────────────────────────────────────

func TestCloneEnvDiffDeltaNilInputReturnsNil(t *testing.T) {
	if out := cloneEnvDiff(nil); out != nil {
		t.Errorf("cloneEnvDiff(nil) = %v, want nil", out)
	}
}

func TestCloneEnvDiffDeltaEmptyInputReturnsNil(t *testing.T) {
	if out := cloneEnvDiff(map[string]model.DeltaEnvChange{}); out != nil {
		t.Errorf("cloneEnvDiff({}) = %v, want nil", out)
	}
}

func TestCloneEnvDiffDeltaCopiesEntries(t *testing.T) {
	in := map[string]model.DeltaEnvChange{
		"branch": {Baseline: "main", Current: "feature"},
	}
	out := cloneEnvDiff(in)
	if out == nil {
		t.Fatal("cloneEnvDiff returned nil for non-empty input")
	}
	if len(out) != 1 {
		t.Errorf("expected 1 entry, got %d", len(out))
	}
	if out["branch"].Baseline != "main" || out["branch"].Current != "feature" {
		t.Errorf("unexpected entry: %#v", out["branch"])
	}
}

func TestCloneEnvDiffDeltaSkipsBlankKeys(t *testing.T) {
	in := map[string]model.DeltaEnvChange{
		"":       {Baseline: "x", Current: "y"},
		"branch": {Baseline: "main", Current: "feature"},
	}
	out := cloneEnvDiff(in)
	if out == nil {
		t.Fatal("cloneEnvDiff returned nil unexpectedly")
	}
	if _, ok := out[""]; ok {
		t.Error("cloneEnvDiff should skip blank keys")
	}
	if len(out) != 1 {
		t.Errorf("expected 1 entry, got %d: %v", len(out), out)
	}
}

func TestCloneEnvDiffDeltaAllBlankKeysReturnsNil(t *testing.T) {
	in := map[string]model.DeltaEnvChange{
		"   ": {Baseline: "x", Current: "y"},
	}
	out := cloneEnvDiff(in)
	if out != nil {
		t.Errorf("cloneEnvDiff with all blank keys = %v, want nil", out)
	}
}

func TestCloneEnvDiffDeltaDoesNotMutateOriginal(t *testing.T) {
	in := map[string]model.DeltaEnvChange{
		"branch": {Baseline: "main", Current: "feature"},
	}
	out := cloneEnvDiff(in)
	if out == nil {
		t.Fatal("unexpected nil")
	}
	out["branch"] = model.DeltaEnvChange{Baseline: "other", Current: "other"}
	if in["branch"].Baseline != "main" {
		t.Error("cloneEnvDiff mutated the original map")
	}
}
