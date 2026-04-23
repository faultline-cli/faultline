package playbooks

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"faultline/internal/model"

	"gopkg.in/yaml.v3"
)

const MatchCatalogFileName = "faultline-matchers.yaml"

type rawMatchCatalog struct {
	SchemaVersion string                     `yaml:"schema_version"`
	NamedMatches  map[string]rawNamedMatcher `yaml:"named_matches"`
}

type rawNamedMatcher struct {
	Use     []string               `yaml:"use"`
	Any     []string               `yaml:"any"`
	All     []string               `yaml:"all"`
	None    []string               `yaml:"none"`
	Partial []rawPartialMatchGroup `yaml:"partial"`
}

type matchCatalog struct {
	PackName string
	Root     string
	Named    map[string]model.MatchSpec
}

func loadPackMatchCatalogs(packs []Pack) ([]matchCatalog, error) {
	out := make([]matchCatalog, 0, len(packs))
	for _, pack := range packs {
		catalog, err := loadMatchCatalog(pack.Root, pack.Name)
		if err != nil {
			return nil, err
		}
		if len(catalog.Named) == 0 {
			continue
		}
		out = append(out, catalog)
	}
	return out, nil
}

func loadMatchCatalog(root, packName string) (matchCatalog, error) {
	path := filepath.Join(root, MatchCatalogFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return matchCatalog{}, nil
		}
		return matchCatalog{}, fmt.Errorf("read match catalog %s: %w", path, err)
	}

	var rawCatalog rawMatchCatalog
	if err := yaml.Unmarshal(data, &rawCatalog); err != nil {
		return matchCatalog{}, fmt.Errorf("parse match catalog %s: %w", path, err)
	}
	if version := strings.TrimSpace(rawCatalog.SchemaVersion); version != "" && version != "matchers.v1" {
		return matchCatalog{}, fmt.Errorf("match catalog %s: unsupported schema version %q", path, version)
	}

	catalog := matchCatalog{
		PackName: packName,
		Root:     root,
		Named:    make(map[string]model.MatchSpec, len(rawCatalog.NamedMatches)),
	}

	keys := make([]string, 0, len(rawCatalog.NamedMatches))
	for key := range rawCatalog.NamedMatches {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		id := strings.TrimSpace(key)
		if id == "" {
			return matchCatalog{}, fmt.Errorf("match catalog %s: named_matches must not use an empty id", path)
		}
		item := rawCatalog.NamedMatches[key]
		if err := validateMatchCatalogItem(item, path, id); err != nil {
			return matchCatalog{}, err
		}
		catalog.Named[id] = model.MatchSpec{
			Use:     trimStrings(item.Use),
			Any:     append([]string(nil), item.Any...),
			All:     append([]string(nil), item.All...),
			None:    append([]string(nil), item.None...),
			Partial: convertPartialMatchGroups(item.Partial),
		}
	}
	return catalog, nil
}

func validateMatchCatalogItem(item rawNamedMatcher, path, id string) error {
	if len(item.Any) == 0 && len(item.All) == 0 && len(item.None) == 0 && len(item.Use) == 0 && len(item.Partial) == 0 {
		return fmt.Errorf("match catalog %s: named_matches.%s must define at least one matcher", path, id)
	}
	if len(item.Any) > 0 {
		if err := validatePatterns(item.Any, "named_matches."+id+".any", path); err != nil {
			return err
		}
	}
	if len(item.All) > 0 {
		if err := validatePatterns(item.All, "named_matches."+id+".all", path); err != nil {
			return err
		}
	}
	if len(item.None) > 0 {
		if err := validatePatterns(item.None, "named_matches."+id+".none", path); err != nil {
			return err
		}
	}
	if err := validateMatchRefs(item.Use, path); err != nil {
		return err
	}
	if err := validatePartialMatchGroups(item.Partial, path); err != nil {
		return err
	}
	if err := validateExclusions(item.Any, item.All, item.None, path); err != nil {
		return err
	}
	return nil
}

func applyMatchCatalogs(pbs []model.Playbook, catalogs []matchCatalog) ([]model.Playbook, error) {
	if len(pbs) == 0 {
		return nil, nil
	}

	named := make(map[string]model.MatchSpec)
	for _, catalog := range catalogs {
		for id, spec := range catalog.Named {
			named[id] = spec
		}
	}

	playbooksByID := make(map[string]model.Playbook, len(pbs))
	for _, pb := range pbs {
		playbooksByID[pb.ID] = pb
	}

	resolvedNamed := make(map[string]model.MatchSpec, len(named))
	visitingNamed := make(map[string]bool, len(named))
	var resolveNamed func(string) (model.MatchSpec, error)
	resolveNamed = func(id string) (model.MatchSpec, error) {
		if spec, ok := resolvedNamed[id]; ok {
			return spec, nil
		}
		spec, ok := named[id]
		if !ok {
			return model.MatchSpec{}, fmt.Errorf("unknown named match %q", id)
		}
		if visitingNamed[id] {
			return model.MatchSpec{}, fmt.Errorf("named match %q forms a composition cycle", id)
		}
		visitingNamed[id] = true
		expanded, err := expandMatchSpec(spec, func(ref string) (model.MatchSpec, error) {
			if strings.HasPrefix(ref, "playbook:") {
				return model.MatchSpec{}, fmt.Errorf("named match %q cannot reference playbook %q", id, strings.TrimPrefix(ref, "playbook:"))
			}
			return resolveNamed(ref)
		})
		visitingNamed[id] = false
		if err != nil {
			return model.MatchSpec{}, fmt.Errorf("named match %q: %w", id, err)
		}
		resolvedNamed[id] = expanded
		return expanded, nil
	}

	resolvedPlaybooks := make(map[string]model.MatchSpec, len(pbs))
	visitingPlaybooks := make(map[string]bool, len(pbs))
	var resolvePlaybook func(string) (model.MatchSpec, error)
	resolvePlaybook = func(id string) (model.MatchSpec, error) {
		if spec, ok := resolvedPlaybooks[id]; ok {
			return spec, nil
		}
		pb, ok := playbooksByID[id]
		if !ok {
			return model.MatchSpec{}, fmt.Errorf("unknown playbook %q", id)
		}
		if visitingPlaybooks[id] {
			return model.MatchSpec{}, fmt.Errorf("playbook %q forms a match composition cycle", id)
		}
		visitingPlaybooks[id] = true
		expanded, err := expandMatchSpec(pb.Match, func(ref string) (model.MatchSpec, error) {
			if strings.HasPrefix(ref, "playbook:") {
				return resolvePlaybook(strings.TrimSpace(strings.TrimPrefix(ref, "playbook:")))
			}
			return resolveNamed(ref)
		})
		visitingPlaybooks[id] = false
		if err != nil {
			return model.MatchSpec{}, fmt.Errorf("playbook %q: %w", id, err)
		}
		resolvedPlaybooks[id] = expanded
		return expanded, nil
	}

	out := make([]model.Playbook, len(pbs))
	for i, pb := range pbs {
		resolved, err := resolvePlaybook(pb.ID)
		if err != nil {
			return nil, err
		}
		resolved.Use = append([]string(nil), pb.Match.Use...)
		if err := validateExclusions(resolved.Any, resolved.All, resolved.None, pb.Metadata.SourceFile); err != nil {
			return nil, err
		}
		pb.Match = resolved
		out[i] = pb
	}
	return out, nil
}

func expandMatchSpec(spec model.MatchSpec, resolveRef func(string) (model.MatchSpec, error)) (model.MatchSpec, error) {
	expanded := model.MatchSpec{}
	for _, ref := range spec.Use {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		referenced, err := resolveRef(ref)
		if err != nil {
			return model.MatchSpec{}, err
		}
		expanded = mergeExpandedMatch(expanded, referenced)
	}
	expanded = mergeExpandedMatch(expanded, model.MatchSpec{
		Any:     spec.Any,
		All:     spec.All,
		None:    spec.None,
		Partial: spec.Partial,
	})
	return expanded, nil
}

func mergeExpandedMatch(base, child model.MatchSpec) model.MatchSpec {
	return model.MatchSpec{
		Any:     mergeUnique(base.Any, child.Any),
		All:     mergeUnique(base.All, child.All),
		None:    mergeUnique(base.None, child.None),
		Partial: mergePartialGroups(base.Partial, child.Partial),
	}
}
