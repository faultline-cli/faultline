package playbooks

import (
	"fmt"

	"faultline/internal/model"
)

const (
	// BundledPackName is the shipped starter pack that Faultline loads by default.
	BundledPackName = "starter"
	// CustomPackName is used when callers point Faultline at an explicit directory.
	CustomPackName = "custom"
)

// Pack describes one deterministic playbook pack root.
type Pack struct {
	Name string
	Root string
}

// LoadPacks loads packs in the provided order and merges them into a single
// deterministic playbook set. Duplicate playbook IDs across packs are rejected.
func LoadPacks(packs []Pack) ([]model.Playbook, error) {
	merged := make([]model.Playbook, 0)
	seen := make(map[string]string)
	for _, pack := range packs {
		if pack.Root == "" {
			return nil, fmt.Errorf("playbook pack %q has no root directory", pack.Name)
		}
		pbs, err := LoadDir(pack.Root)
		if err != nil {
			return nil, fmt.Errorf("load playbook pack %q: %w", pack.Name, err)
		}
		for _, pb := range pbs {
			if prev, ok := seen[pb.ID]; ok {
				return nil, fmt.Errorf("duplicate playbook id %q across packs %q and %q", pb.ID, prev, pack.Name)
			}
			seen[pb.ID] = pack.Name
		}
		merged = append(merged, pbs...)
	}
	return merged, nil
}
