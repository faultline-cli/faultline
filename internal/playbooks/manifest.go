package playbooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const packManifestFile = "faultline-pack.yaml"

// Manifest describes optional metadata for a playbook pack.
// Packs without a manifest are still supported for backward compatibility.
type Manifest struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description,omitempty"`
	Detectors   []string `yaml:"detectors,omitempty"`
	Includes    []string `yaml:"includes,omitempty"`
}

func loadManifest(root string) (Manifest, bool, error) {
	path := filepath.Join(root, packManifestFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Manifest{}, false, nil
	}
	if err != nil {
		return Manifest{}, false, fmt.Errorf("read pack manifest %s: %w", path, err)
	}
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, true, fmt.Errorf("parse pack manifest %s: %w", path, err)
	}
	if err := validateManifest(manifest, path); err != nil {
		return Manifest{}, true, err
	}
	return manifest, true, nil
}

func validateManifest(manifest Manifest, path string) error {
	if strings.TrimSpace(manifest.Name) == "" {
		return fmt.Errorf("pack manifest %s: missing required field 'name'", path)
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return fmt.Errorf("pack manifest %s: missing required field 'version'", path)
	}
	for i, detector := range manifest.Detectors {
		switch strings.TrimSpace(detector) {
		case "", "log", "source":
			if strings.TrimSpace(detector) == "" {
				return fmt.Errorf("pack manifest %s: detectors[%d] must not be empty", path, i)
			}
		default:
			return fmt.Errorf("pack manifest %s: unknown detector %q", path, detector)
		}
	}
	for i, include := range manifest.Includes {
		if strings.TrimSpace(include) == "" {
			return fmt.Errorf("pack manifest %s: includes[%d] must not be empty", path, i)
		}
	}
	return nil
}
