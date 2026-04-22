package playbooks

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// PackMetaFileName is the per-pack manifest stored at the root of every
// installed pack. It records provenance so analysis output can report which
// pack version produced each result.
const PackMetaFileName = "faultline-pack.yaml"

// PackMeta describes the provenance of a locally installed pack.
// It is written at install time and read during pack loading to populate
// version and pinned reference information in analysis output.
type PackMeta struct {
	Name      string `yaml:"name"`
	Version   string `yaml:"version,omitempty"`
	SourceURL string `yaml:"source_url,omitempty"`
	PinnedRef string `yaml:"pinned_ref,omitempty"`
	FetchedAt string `yaml:"fetched_at,omitempty"`
}

// ReadPackMeta reads faultline-pack.yaml from root.
// If the file does not exist, ok is false and err is nil.
func ReadPackMeta(root string) (PackMeta, bool, error) {
	path := filepath.Join(root, PackMetaFileName)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return PackMeta{}, false, nil
	}
	if err != nil {
		return PackMeta{}, false, fmt.Errorf("read pack manifest %s: %w", path, err)
	}
	var meta PackMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return PackMeta{}, false, fmt.Errorf("parse pack manifest %s: %w", path, err)
	}
	return meta, true, nil
}

// WritePackMeta writes meta as faultline-pack.yaml under dest with mode 0600.
// FetchedAt is set to the current UTC time if it is empty.
func WritePackMeta(dest string, meta PackMeta) error {
	if meta.FetchedAt == "" {
		meta.FetchedAt = time.Now().UTC().Format(time.RFC3339)
	}
	data, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal pack manifest: %w", err)
	}
	path := filepath.Join(dest, PackMetaFileName)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write pack manifest %s: %w", path, err)
	}
	return nil
}
