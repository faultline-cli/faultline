package playbooks

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"faultline/internal/model"
)

const HookCatalogFileName = "faultline-hooks.yaml"

type hookCatalog struct {
	Named     map[string]model.HookDefinition
	Playbooks map[string]model.PlaybookHooks
}

type rawHookCatalog struct {
	SchemaVersion string                          `yaml:"schema_version"`
	NamedHooks    map[string]model.HookDefinition `yaml:"named_hooks"`
	PlaybookHooks map[string]model.PlaybookHooks  `yaml:"playbook_hooks"`
}

func normalizePlaybookHooks(hooks model.PlaybookHooks) model.PlaybookHooks {
	return model.PlaybookHooks{
		Verify:    normalizeHookDefinitions(hooks.Verify),
		Collect:   normalizeHookDefinitions(hooks.Collect),
		Remediate: normalizeHookDefinitions(hooks.Remediate),
		Disable:   normalizeDisableList(hooks.Disable),
	}
}

func normalizeHookDefinitions(defs []model.HookDefinition) []model.HookDefinition {
	if len(defs) == 0 {
		return nil
	}
	out := make([]model.HookDefinition, 0, len(defs))
	for _, def := range defs {
		out = append(out, normalizeHookDefinition(def))
	}
	return out
}

func normalizeHookDefinition(def model.HookDefinition) model.HookDefinition {
	def.ID = strings.TrimSpace(def.ID)
	def.Use = strings.TrimSpace(def.Use)
	def.Extends = strings.TrimSpace(def.Extends)
	rawPath := strings.TrimSpace(def.Path)
	def.Path = filepath.Clean(rawPath)
	if rawPath == "" {
		def.Path = ""
	}
	def.EnvVar = strings.TrimSpace(def.EnvVar)
	def.Pattern = strings.TrimSpace(def.Pattern)
	if len(def.Command) > 0 {
		cmd := make([]string, 0, len(def.Command))
		for _, item := range def.Command {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			cmd = append(cmd, item)
		}
		def.Command = cmd
	}
	return def
}

func normalizeDisableList(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func validatePlaybookHooks(hooks model.PlaybookHooks, path string) error {
	if err := validateHookDefinitions(hooks.Verify, model.HookCategoryVerify, path, "hooks.verify", false); err != nil {
		return err
	}
	if err := validateHookDefinitions(hooks.Collect, model.HookCategoryCollect, path, "hooks.collect", false); err != nil {
		return err
	}
	if err := validateHookDefinitions(hooks.Remediate, model.HookCategoryRemediate, path, "hooks.remediate", false); err != nil {
		return err
	}
	for i, item := range hooks.Disable {
		if strings.TrimSpace(item) == "" {
			return fmt.Errorf("playbook %s: hooks.disable[%d] must not be empty", path, i)
		}
	}
	return nil
}

func validateHookDefinitions(defs []model.HookDefinition, category model.HookCategory, path, section string, allowExtends bool) error {
	seen := make(map[string]struct{}, len(defs))
	for i, def := range defs {
		if err := validateHookDefinition(def, category, path, fmt.Sprintf("%s[%d]", section, i), allowExtends); err != nil {
			return err
		}
		if def.Use != "" {
			continue
		}
		if def.ID == "" {
			continue
		}
		if _, ok := seen[def.ID]; ok {
			return fmt.Errorf("playbook %s: %s duplicates hook id %q", path, section, def.ID)
		}
		seen[def.ID] = struct{}{}
	}
	return nil
}

func validateHookDefinition(def model.HookDefinition, category model.HookCategory, path, section string, allowExtends bool) error {
	if def.Use != "" {
		if def.ID != "" || def.Extends != "" || def.Kind != "" || def.Path != "" || def.EnvVar != "" || len(def.Command) > 0 || def.Pattern != "" || def.Lines != 0 || def.MaxBytes != 0 || def.ConfidenceDelta != 0 {
			return fmt.Errorf("playbook %s: %s uses a named hook and must not set inline fields", path, section)
		}
		return nil
	}
	if def.ID == "" {
		return fmt.Errorf("playbook %s: %s must set hook id", path, section)
	}
	if def.Extends != "" && !allowExtends {
		return fmt.Errorf("playbook %s: %s cannot use extends", path, section)
	}
	if def.Kind == "" {
		if allowExtends && def.Extends != "" {
			return nil
		}
		return fmt.Errorf("playbook %s: %s must set kind", path, section)
	}
	if allowExtends {
		return nil
	}
	if math.Abs(def.ConfidenceDelta) > 0.20 {
		return fmt.Errorf("playbook %s: %s confidence_delta must stay within [-0.20, 0.20]", path, section)
	}
	if category != "" && category != model.HookCategoryVerify && def.ConfidenceDelta != 0 {
		return fmt.Errorf("playbook %s: %s confidence_delta is only supported for verify hooks", path, section)
	}
	if def.Lines < 0 {
		return fmt.Errorf("playbook %s: %s lines must be 0 or greater", path, section)
	}
	if def.MaxBytes < 0 {
		return fmt.Errorf("playbook %s: %s max_bytes must be 0 or greater", path, section)
	}
	switch def.Kind {
	case model.HookKindFileExists, model.HookKindDirExists, model.HookKindReadFileExcerpt:
		if def.Path == "" {
			return fmt.Errorf("playbook %s: %s kind %q requires path", path, section, def.Kind)
		}
	case model.HookKindEnvVarPresent:
		if def.EnvVar == "" {
			return fmt.Errorf("playbook %s: %s kind %q requires env_var", path, section, def.Kind)
		}
	case model.HookKindCommandExitZero, model.HookKindCommandOutputMatches, model.HookKindCommandOutputCapture:
		if len(def.Command) == 0 {
			return fmt.Errorf("playbook %s: %s kind %q requires command", path, section, def.Kind)
		}
		if def.Kind == model.HookKindCommandOutputMatches && def.Pattern == "" {
			return fmt.Errorf("playbook %s: %s kind %q requires pattern", path, section, def.Kind)
		}
	default:
		return fmt.Errorf("playbook %s: %s uses unknown hook kind %q", path, section, def.Kind)
	}
	return nil
}

func loadHookCatalog(root string, pack Pack) (hookCatalog, error) {
	path := filepath.Join(root, HookCatalogFileName)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return hookCatalog{}, nil
	}
	if err != nil {
		return hookCatalog{}, fmt.Errorf("read hook catalog %s: %w", path, err)
	}
	var raw rawHookCatalog
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return hookCatalog{}, fmt.Errorf("parse hook catalog %s: %w", path, err)
	}
	if raw.SchemaVersion != "" && raw.SchemaVersion != "hooks.v1" {
		return hookCatalog{}, fmt.Errorf("hook catalog %s: unsupported schema_version %q", path, raw.SchemaVersion)
	}

	catalog := hookCatalog{
		Named:     make(map[string]model.HookDefinition, len(raw.NamedHooks)),
		Playbooks: make(map[string]model.PlaybookHooks, len(raw.PlaybookHooks)),
	}

	namedKeys := make([]string, 0, len(raw.NamedHooks))
	for name := range raw.NamedHooks {
		namedKeys = append(namedKeys, name)
	}
	sort.Strings(namedKeys)
	for _, name := range namedKeys {
		def := normalizeHookDefinition(raw.NamedHooks[name])
		if def.ID == "" {
			def.ID = name
		}
		if def.Metadata.SourcePack == "" {
			def.Metadata.SourcePack = pack.Name
		}
		if def.Metadata.SourceFile == "" {
			def.Metadata.SourceFile = path
		}
		if err := validateHookDefinition(def, "", path, "named_hooks."+name, true); err != nil {
			return hookCatalog{}, err
		}
		catalog.Named[name] = def
	}

	playbookIDs := make([]string, 0, len(raw.PlaybookHooks))
	for id := range raw.PlaybookHooks {
		playbookIDs = append(playbookIDs, id)
	}
	sort.Strings(playbookIDs)
	for _, id := range playbookIDs {
		hooks := normalizePlaybookHooks(raw.PlaybookHooks[id])
		if err := validatePlaybookHooks(hooks, path); err != nil {
			return hookCatalog{}, err
		}
		catalog.Playbooks[id] = stampHookMetadata(hooks, pack, path)
	}

	return catalog, nil
}

func stampHookMetadata(hooks model.PlaybookHooks, pack Pack, path string) model.PlaybookHooks {
	stamp := func(defs []model.HookDefinition) []model.HookDefinition {
		if len(defs) == 0 {
			return nil
		}
		out := make([]model.HookDefinition, 0, len(defs))
		for _, def := range defs {
			if def.Metadata.SourcePack == "" {
				def.Metadata.SourcePack = pack.Name
			}
			if def.Metadata.SourceFile == "" {
				def.Metadata.SourceFile = path
			}
			out = append(out, def)
		}
		return out
	}
	return model.PlaybookHooks{
		Verify:    stamp(hooks.Verify),
		Collect:   stamp(hooks.Collect),
		Remediate: stamp(hooks.Remediate),
		Disable:   append([]string(nil), hooks.Disable...),
	}
}

func applyHookCatalogs(pbs []model.Playbook, catalogs []hookCatalog) ([]model.Playbook, error) {
	if len(pbs) == 0 || len(catalogs) == 0 {
		return pbs, nil
	}

	named := make(map[string]model.HookDefinition)
	for _, catalog := range catalogs {
		keys := make([]string, 0, len(catalog.Named))
		for key := range catalog.Named {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			named[key] = catalog.Named[key]
		}
	}

	resolvedNamed, err := resolveNamedHookDefinitions(named)
	if err != nil {
		return nil, err
	}

	index := make(map[string]int, len(pbs))
	resolved := make([]model.Playbook, len(pbs))
	copy(resolved, pbs)
	for i := range resolved {
		resolved[i].Hooks = normalizePlaybookHooks(resolved[i].Hooks)
		index[resolved[i].ID] = i
	}

	for _, catalog := range catalogs {
		keys := make([]string, 0, len(catalog.Playbooks))
		for key := range catalog.Playbooks {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			idx, ok := index[key]
			if !ok {
				return nil, fmt.Errorf("hook catalog references unknown playbook %q", key)
			}
			current := resolved[idx].Hooks
			overlay := catalog.Playbooks[key]
			current = disableHookIDs(current, overlay.Disable)
			current.Verify = append(current.Verify, overlay.Verify...)
			current.Collect = append(current.Collect, overlay.Collect...)
			current.Remediate = append(current.Remediate, overlay.Remediate...)
			resolved[idx].Hooks = current
		}
	}

	for i := range resolved {
		hooks, err := resolvePlaybookHooks(resolved[i].Hooks, resolvedNamed, resolved[i].Metadata.SourceFile)
		if err != nil {
			return nil, fmt.Errorf("playbook %s: %w", resolved[i].ID, err)
		}
		resolved[i].Hooks = hooks
	}

	return resolved, nil
}

func resolveNamedHookDefinitions(named map[string]model.HookDefinition) (map[string]model.HookDefinition, error) {
	if len(named) == 0 {
		return nil, nil
	}
	resolved := make(map[string]model.HookDefinition, len(named))
	visiting := make(map[string]bool, len(named))

	var resolve func(name string) (model.HookDefinition, error)
	resolve = func(name string) (model.HookDefinition, error) {
		if def, ok := resolved[name]; ok {
			return def, nil
		}
		def, ok := named[name]
		if !ok {
			return model.HookDefinition{}, fmt.Errorf("unknown named hook %q", name)
		}
		if visiting[name] {
			return model.HookDefinition{}, fmt.Errorf("named hook %q forms an inheritance cycle", name)
		}
		visiting[name] = true
		if def.Extends != "" {
			base, err := resolve(def.Extends)
			if err != nil {
				return model.HookDefinition{}, err
			}
			def = mergeHookDefinitions(base, def)
		}
		def.Extends = ""
		visiting[name] = false
		resolved[name] = def
		return def, nil
	}

	keys := make([]string, 0, len(named))
	for key := range named {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, err := resolve(key); err != nil {
			return nil, err
		}
	}
	return resolved, nil
}

func mergeHookDefinitions(base, override model.HookDefinition) model.HookDefinition {
	merged := base
	merged.ID = override.ID
	merged.Extends = override.Extends
	merged.Metadata = override.Metadata
	if override.Kind != "" {
		merged.Kind = override.Kind
	}
	if override.Path != "" {
		merged.Path = override.Path
	}
	if override.EnvVar != "" {
		merged.EnvVar = override.EnvVar
	}
	if len(override.Command) > 0 {
		merged.Command = append([]string(nil), override.Command...)
	}
	if override.Pattern != "" {
		merged.Pattern = override.Pattern
	}
	if override.Lines != 0 {
		merged.Lines = override.Lines
	}
	if override.MaxBytes != 0 {
		merged.MaxBytes = override.MaxBytes
	}
	if override.ConfidenceDelta != 0 {
		merged.ConfidenceDelta = override.ConfidenceDelta
	}
	return merged
}

func disableHookIDs(hooks model.PlaybookHooks, ids []string) model.PlaybookHooks {
	if len(ids) == 0 {
		return hooks
	}
	blocked := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		blocked[id] = struct{}{}
	}
	filter := func(defs []model.HookDefinition) []model.HookDefinition {
		if len(defs) == 0 {
			return nil
		}
		out := make([]model.HookDefinition, 0, len(defs))
		for _, def := range defs {
			if _, ok := blocked[hookRefID(def)]; ok {
				continue
			}
			out = append(out, def)
		}
		return out
	}
	hooks.Verify = filter(hooks.Verify)
	hooks.Collect = filter(hooks.Collect)
	hooks.Remediate = filter(hooks.Remediate)
	return hooks
}

func resolvePlaybookHooks(hooks model.PlaybookHooks, named map[string]model.HookDefinition, source string) (model.PlaybookHooks, error) {
	resolveCategory := func(defs []model.HookDefinition, category model.HookCategory) ([]model.HookDefinition, error) {
		if len(defs) == 0 {
			return nil, nil
		}
		out := make([]model.HookDefinition, 0, len(defs))
		seen := make(map[string]struct{}, len(defs))
		for i, def := range defs {
			if def.Use != "" {
				namedDef, ok := named[def.Use]
				if !ok {
					return nil, fmt.Errorf("%s[%d] references unknown named hook %q", category, i, def.Use)
				}
				def = namedDef
			}
			if err := validateHookDefinition(def, category, source, fmt.Sprintf("%s[%d]", category, i), false); err != nil {
				return nil, err
			}
			if _, ok := seen[def.ID]; ok {
				return nil, fmt.Errorf("%s duplicates resolved hook id %q", category, def.ID)
			}
			seen[def.ID] = struct{}{}
			def.Use = ""
			def.Extends = ""
			out = append(out, def)
		}
		return out, nil
	}

	verify, err := resolveCategory(hooks.Verify, model.HookCategoryVerify)
	if err != nil {
		return model.PlaybookHooks{}, err
	}
	collect, err := resolveCategory(hooks.Collect, model.HookCategoryCollect)
	if err != nil {
		return model.PlaybookHooks{}, err
	}
	remediate, err := resolveCategory(hooks.Remediate, model.HookCategoryRemediate)
	if err != nil {
		return model.PlaybookHooks{}, err
	}
	return model.PlaybookHooks{
		Verify:    verify,
		Collect:   collect,
		Remediate: remediate,
	}, nil
}

func hookRefID(def model.HookDefinition) string {
	if def.Use != "" {
		return def.Use
	}
	return def.ID
}

func loadPackHookCatalogs(packs []Pack) ([]hookCatalog, error) {
	if len(packs) == 0 {
		return nil, nil
	}
	out := make([]hookCatalog, 0, len(packs))
	for _, pack := range packs {
		catalog, err := loadHookCatalog(pack.Root, pack)
		if err != nil {
			return nil, err
		}
		if len(catalog.Named) == 0 && len(catalog.Playbooks) == 0 {
			continue
		}
		out = append(out, catalog)
	}
	return out, nil
}

func validateHooksResolved(pbs []model.Playbook) error {
	var errs []error
	for _, pb := range pbs {
		if err := validatePlaybookHooks(pb.Hooks, pb.Metadata.SourceFile); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
