package detectors

import (
	"fmt"
	"sort"
	"strings"
)

// Registry resolves detector implementations by kind.
type Registry struct {
	detectors map[Kind]Detector
}

// NewRegistry builds a registry from the provided detectors.
func NewRegistry(list ...Detector) Registry {
	byKind := make(map[Kind]Detector, len(list))
	for _, detector := range list {
		byKind[detector.Kind()] = detector
	}
	return Registry{detectors: byKind}
}

// Lookup returns the detector for kind.
func (r Registry) Lookup(kind Kind) (Detector, bool) {
	detector, ok := r.detectors[kind]
	return detector, ok
}

// MustLookup returns the detector for kind or a descriptive error.
func (r Registry) MustLookup(kind Kind) (Detector, error) {
	detector, ok := r.Lookup(kind)
	if ok {
		return detector, nil
	}
	return nil, fmt.Errorf("detector %q is not registered (available: %s)", kind, r.availableKinds())
}

func (r Registry) availableKinds() string {
	if len(r.detectors) == 0 {
		return "none"
	}
	kinds := make([]string, 0, len(r.detectors))
	for kind := range r.detectors {
		kinds = append(kinds, string(kind))
	}
	sort.Strings(kinds)
	return strings.Join(kinds, ", ")
}
