package normalize_test

import (
	"testing"

	"faultline/tools/eval-corpus/normalize"
)

func TestDeduplicatorMark(t *testing.T) {
	d := normalize.NewDeduplicator()
	if d.Seen("abc") {
		t.Error("Seen should be false before Mark")
	}
	d.Mark("abc")
	if !d.Seen("abc") {
		t.Error("Seen should be true after Mark")
	}
}

func TestDeduplicatorCount(t *testing.T) {
	d := normalize.NewDeduplicator()
	for _, id := range []string{"a", "b", "c"} {
		d.Mark(id)
	}
	if d.Count() != 3 {
		t.Errorf("Count = %d, want 3", d.Count())
	}
	// Marking the same ID again should not increase count.
	d.Mark("a")
	if d.Count() != 3 {
		t.Errorf("Count = %d after re-mark, want 3", d.Count())
	}
}

func TestDeduplicatorNewIsEmpty(t *testing.T) {
	d := normalize.NewDeduplicator()
	if d.Count() != 0 {
		t.Errorf("new Deduplicator Count = %d, want 0", d.Count())
	}
}
