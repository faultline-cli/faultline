package playbooks

import (
	"fmt"
	"os"
	"strings"

	"faultline/internal/model"
)

const packsEnvKey = "FAULTLINE_PLAYBOOK_PACKS"

// CatalogOptions configures deterministic playbook pack resolution.
type CatalogOptions struct {
	OverrideDir   string
	ExtraPackDirs []string
}

// Catalog provides deterministic access to a resolved playbook set.
type Catalog struct {
	overrideDir string
	extraPacks  []string
}

// NewCatalog returns a catalog backed by dir. When dir is empty, the default
// bundled starter pack is resolved lazily.
func NewCatalog(dir string) Catalog {
	return NewCatalogWithOptions(CatalogOptions{OverrideDir: dir})
}

// NewCatalogWithOptions returns a catalog configured for override or
// bundled-plus-extra pack composition.
func NewCatalogWithOptions(opts CatalogOptions) Catalog {
	return Catalog{
		overrideDir: strings.TrimSpace(opts.OverrideDir),
		extraPacks:  append([]string(nil), opts.ExtraPackDirs...),
	}
}

// Packs returns the resolved pack set.
func (c Catalog) Packs() ([]Pack, error) {
	if c.overrideDir != "" {
		if len(cleanPackDirs(c.extraPacks)) > 0 {
			return nil, fmt.Errorf("playbook override directory cannot be combined with additional playbook packs")
		}
		dir, err := validateDir(c.overrideDir)
		if err != nil {
			return nil, err
		}
		return []Pack{{
			Name: CustomPackName,
			Root: dir,
		}}, nil
	}

	root, err := DefaultDir()
	if err != nil {
		return nil, err
	}
	packs := []Pack{{
		Name: BundledPackName,
		Root: root,
	}}
	for i, dir := range c.resolvedExtraPackDirs() {
		root, err := validateDir(dir)
		if err != nil {
			return nil, err
		}
		packs = append(packs, Pack{
			Name: fmt.Sprintf("extra-%d", i+1),
			Root: root,
		})
	}
	return packs, nil
}

// ExtraPackEnvKey returns the environment variable name used for additional
// pack composition on top of the bundled starter catalog.
func ExtraPackEnvKey() string {
	return packsEnvKey
}

// Dir returns the resolved playbook directory.
func (c Catalog) Dir() (string, error) {
	packs, err := c.Packs()
	if err != nil {
		return "", err
	}
	if len(packs) != 1 {
		return "", fmt.Errorf("catalog spans %d packs; no single directory applies", len(packs))
	}
	return packs[0].Root, nil
}

// Load returns the full ordered playbook set.
func (c Catalog) Load() ([]model.Playbook, error) {
	packs, err := c.Packs()
	if err != nil {
		return nil, err
	}
	return LoadPacks(packs)
}

// List returns the full ordered playbook set.
func (c Catalog) List() ([]model.Playbook, error) {
	return c.Load()
}

// Explain returns a single playbook by deterministic ID lookup.
func (c Catalog) Explain(id string) (model.Playbook, error) {
	pbs, err := c.Load()
	if err != nil {
		return model.Playbook{}, err
	}
	for _, pb := range pbs {
		if pb.ID == id {
			return pb, nil
		}
	}
	return model.Playbook{}, fmt.Errorf("unknown playbook %q", id)
}

func (c Catalog) resolvedExtraPackDirs() []string {
	if len(c.extraPacks) > 0 {
		return cleanPackDirs(c.extraPacks)
	}
	raw := strings.TrimSpace(os.Getenv(packsEnvKey))
	if raw == "" {
		return nil
	}
	return cleanPackDirs(strings.Split(raw, string(os.PathListSeparator)))
}

func cleanPackDirs(dirs []string) []string {
	out := make([]string, 0, len(dirs))
	seen := make(map[string]struct{}, len(dirs))
	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		if _, ok := seen[dir]; ok {
			continue
		}
		seen[dir] = struct{}{}
		out = append(out, dir)
	}
	return out
}
