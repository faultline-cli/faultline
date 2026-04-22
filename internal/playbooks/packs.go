package playbooks

import (
	"fmt"
	"strings"

	"faultline/internal/model"
)

const (
	// BundledPackName is the shipped starter pack that Faultline loads by default.
	BundledPackName = "starter"
	// CustomPackName is used when callers point Faultline at an explicit directory.
	CustomPackName = "custom"
)

// Pack describes one deterministic playbook pack root and its provenance.
type Pack struct {
	Name      string
	Root      string
	Version   string // from faultline-pack.yaml; empty for the bundled starter pack
	SourceURL string // install-time source directory or URL; empty if not recorded
	PinnedRef string // git commit hash or tag at install time; empty if not available
}

// ProvenanceFromPlaybooks builds the ordered pack provenance list from a loaded
// playbook set. Packs appear in the order they were first encountered and each
// entry carries the version and pinned reference recorded at install time.
func ProvenanceFromPlaybooks(pbs []model.Playbook) []model.PackProvenance {
	type provKey struct{ name, root string }
	order := make([]provKey, 0)
	counts := make(map[provKey]int)
	meta := make(map[provKey]model.PackProvenance)

	for _, pb := range pbs {
		k := provKey{pb.Metadata.PackName, pb.Metadata.PackRoot}
		if _, seen := counts[k]; !seen {
			order = append(order, k)
			meta[k] = model.PackProvenance{
				Name:      pb.Metadata.PackName,
				Version:   pb.Metadata.PackVersion,
				SourceURL: pb.Metadata.PackSourceURL,
				PinnedRef: pb.Metadata.PackPinnedRef,
			}
		}
		counts[k]++
	}

	result := make([]model.PackProvenance, 0, len(order))
	for _, k := range order {
		p := meta[k]
		p.PlaybookCount = counts[k]
		result = append(result, p)
	}
	return result
}

// LoadPacks loads packs in the provided order and merges them into a single
// deterministic playbook set. Duplicate playbook IDs across packs are rejected.
// Pack metadata (version, source URL, pinned ref) is read from faultline-pack.yaml
// when present and propagated into each playbook's Metadata.
func LoadPacks(packs []Pack) ([]model.Playbook, error) {
	merged := make([]model.Playbook, 0)
	seen := make(map[string]string)
	for _, pack := range packs {
		if pack.Root == "" {
			return nil, fmt.Errorf("playbook pack %q has no root directory", pack.Name)
		}
		meta, _, _ := ReadPackMeta(pack.Root) // best-effort; missing manifest is fine
		pbs, err := LoadDir(pack.Root)
		if err != nil {
			return nil, fmt.Errorf("load playbook pack %q: %w", pack.Name, err)
		}
		version := firstNonEmpty(pack.Version, meta.Version)
		sourceURL := firstNonEmpty(pack.SourceURL, meta.SourceURL)
		pinnedRef := firstNonEmpty(pack.PinnedRef, meta.PinnedRef)
		for _, pb := range pbs {
			pb.Metadata.PackName = pack.Name
			pb.Metadata.PackRoot = pack.Root
			pb.Metadata.PackVersion = version
			pb.Metadata.PackSourceURL = sourceURL
			pb.Metadata.PackPinnedRef = pinnedRef
			if prev, ok := seen[pb.ID]; ok {
				return nil, fmt.Errorf("duplicate playbook id %q across packs %q and %q", pb.ID, prev, pack.Name)
			}
			seen[pb.ID] = pack.Name
			merged = append(merged, pb)
		}
	}
	return merged, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
