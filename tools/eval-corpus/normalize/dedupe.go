package normalize

// Deduplicator tracks fixture IDs that have already been emitted and provides
// an O(1) membership test for streaming deduplication. It is not goroutine-safe;
// callers must synchronise access when used from multiple goroutines.
type Deduplicator struct {
	seen map[string]struct{}
}

// NewDeduplicator returns an initialised Deduplicator.
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{seen: make(map[string]struct{})}
}

// Seen reports whether id has been marked before.
func (d *Deduplicator) Seen(id string) bool {
	_, ok := d.seen[id]
	return ok
}

// Mark records id so that subsequent calls to Seen return true.
func (d *Deduplicator) Mark(id string) {
	d.seen[id] = struct{}{}
}

// Count returns the number of unique IDs marked so far.
func (d *Deduplicator) Count() int {
	return len(d.seen)
}
