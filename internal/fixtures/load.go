package fixtures

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"faultline/internal/engine"

	"gopkg.in/yaml.v3"
)

func Load(layout Layout, class Class) ([]Fixture, error) {
	switch class {
	case ClassAll:
		minimal, err := loadDir(layout, ClassMinimal)
		if err != nil {
			return nil, err
		}
		realFixtures, err := loadDir(layout, ClassReal)
		if err != nil {
			return nil, err
		}
		return append(minimal, realFixtures...), nil
	case ClassMinimal, ClassReal, ClassStaging:
		return loadDir(layout, class)
	default:
		return nil, fmt.Errorf("unsupported fixture class %q", class)
	}
}

func loadDir(layout Layout, class Class) ([]Fixture, error) {
	dir := dirForClass(layout, class)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Fixture{}, nil
		}
		return nil, fmt.Errorf("read %s fixtures: %w", class, err)
	}

	fixtures := make([]Fixture, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".yaml" && filepath.Ext(name) != ".yml" {
			continue
		}
		path := filepath.Join(dir, name)
		loaded, err := loadFile(path, class, layout.Root)
		if err != nil {
			return nil, err
		}
		fixtures = append(fixtures, loaded...)
	}
	sort.Slice(fixtures, func(i, j int) bool {
		return fixtures[i].ID < fixtures[j].ID
	})
	return fixtures, nil
}

func loadFile(path string, defaultClass Class, repoRoot string) ([]Fixture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read fixture file %s: %w", path, err)
	}

	var manifest manifestFile
	if err := yaml.Unmarshal(data, &manifest); err == nil && len(manifest.Fixtures) > 0 {
		fixtures := make([]Fixture, 0, len(manifest.Fixtures))
		for _, item := range manifest.Fixtures {
			fixture, err := finalizeFixture(item, path, defaultClass, repoRoot)
			if err != nil {
				return nil, err
			}
			fixtures = append(fixtures, fixture)
		}
		return fixtures, nil
	}

	var fixture Fixture
	if err := yaml.Unmarshal(data, &fixture); err != nil {
		return nil, fmt.Errorf("parse fixture file %s: %w", path, err)
	}
	loaded, err := finalizeFixture(fixture, path, defaultClass, repoRoot)
	if err != nil {
		return nil, err
	}
	return []Fixture{loaded}, nil
}

func finalizeFixture(f Fixture, path string, defaultClass Class, repoRoot string) (Fixture, error) {
	if strings.TrimSpace(f.ID) == "" {
		return Fixture{}, fmt.Errorf("fixture file %s is missing id", path)
	}
	f.FilePath = path
	f.manifestRoot = filepath.Dir(path)
	if f.FixtureClass == "" {
		f.FixtureClass = defaultClass
	}
	text, err := fixtureLog(f, repoRoot)
	if err != nil {
		return Fixture{}, err
	}
	if f.Fingerprint == "" {
		f.Fingerprint = FingerprintForLog(text)
	}
	return f, nil
}

func fixtureLog(f Fixture, repoRoot string) (string, error) {
	if strings.TrimSpace(f.NormalizedLog) != "" {
		return engine.CanonicalizeLog(f.NormalizedLog), nil
	}
	if strings.TrimSpace(f.RawLog) != "" {
		return engine.CanonicalizeLog(f.RawLog), nil
	}
	if strings.TrimSpace(f.Path) == "" {
		return "", fmt.Errorf("fixture %s has no log content", f.ID)
	}
	path := f.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(repoRoot, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read fixture log for %s: %w", f.ID, err)
	}
	return engine.CanonicalizeLog(string(data)), nil
}

func dirForClass(layout Layout, class Class) string {
	switch class {
	case ClassMinimal:
		return layout.MinimalDir
	case ClassReal:
		return layout.RealDir
	case ClassStaging:
		return layout.StagingDir
	default:
		return layout.Fixtures
	}
}
