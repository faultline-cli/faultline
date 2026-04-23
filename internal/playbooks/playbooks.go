package playbooks

// Package playbooks loads and validates YAML failure playbooks.
// It replaces the former internal/loader package and adds support for
// recursive directory trees so playbooks can be organised into sub-directories
// by category.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"faultline/internal/engine/hypothesis"

	"gopkg.in/yaml.v3"

	"faultline/internal/model"
)

const envKey = "FAULTLINE_PLAYBOOK_DIR"

type rawSignalMatcher struct {
	ID           string   `yaml:"id"`
	Label        string   `yaml:"label"`
	Description  string   `yaml:"description"`
	Patterns     []string `yaml:"patterns"`
	PathIncludes []string `yaml:"path_includes"`
	PathExcludes []string `yaml:"path_excludes"`
	Scopes       []string `yaml:"scopes"`
	Weight       float64  `yaml:"weight"`
	Required     bool     `yaml:"required"`
}

type rawSuppressionRule struct {
	Style       string   `yaml:"style"`
	Pattern     string   `yaml:"pattern"`
	Paths       []string `yaml:"paths"`
	Playbooks   []string `yaml:"playbooks"`
	Reason      string   `yaml:"reason"`
	ExpiresOn   string   `yaml:"expires_on"`
	Discount    float64  `yaml:"discount"`
	SuppressAll bool     `yaml:"suppress_all"`
}

type rawCompoundSignal struct {
	ID             string   `yaml:"id"`
	Label          string   `yaml:"label"`
	Require        []string `yaml:"require"`
	Scope          string   `yaml:"scope"`
	Bonus          float64  `yaml:"bonus"`
	Required       bool     `yaml:"required"`
	AllowMitigated bool     `yaml:"allow_mitigated"`
}

type rawConsistencyRule struct {
	ID                string   `yaml:"id"`
	Label             string   `yaml:"label"`
	BaselineSignalIDs []string `yaml:"baseline_signal_ids"`
	ExpectedSignalID  string   `yaml:"expected_signal_id"`
	Scope             string   `yaml:"scope"`
	MinimumPeers      int      `yaml:"minimum_peers"`
	Threshold         float64  `yaml:"threshold"`
	Amplifier         float64  `yaml:"amplifier"`
}

type rawPathClassRule struct {
	Class    string   `yaml:"class"`
	Paths    []string `yaml:"paths"`
	Adjust   float64  `yaml:"adjust"`
	HotPath  bool     `yaml:"hot_path"`
	Critical bool     `yaml:"critical"`
}

type rawSafeContextRule struct {
	ID       string   `yaml:"id"`
	Label    string   `yaml:"label"`
	Paths    []string `yaml:"paths"`
	Patterns []string `yaml:"patterns"`
	Discount float64  `yaml:"discount"`
}

type rawHypothesisSignal struct {
	Signal string  `yaml:"signal"`
	Weight float64 `yaml:"weight"`
}

type rawHypothesisDiscriminator struct {
	Description string  `yaml:"description"`
	Signal      string  `yaml:"signal"`
	Weight      float64 `yaml:"weight"`
}

type rawPartialMatchGroup struct {
	ID       string   `yaml:"id"`
	Label    string   `yaml:"label"`
	Minimum  int      `yaml:"minimum"`
	Patterns []string `yaml:"patterns"`
}

type raw struct {
	ID         string   `yaml:"id"`
	Extends    string   `yaml:"extends"`
	Title      string   `yaml:"title"`
	Category   string   `yaml:"category"`
	Severity   string   `yaml:"severity"`
	Detector   string   `yaml:"detector"`
	BaseScore  float64  `yaml:"base_score"`
	Tags       []string `yaml:"tags"`
	StageHints []string `yaml:"stage_hints"`
	Match      struct {
		Any     []string               `yaml:"any"`
		All     []string               `yaml:"all"`
		None    []string               `yaml:"none"`
		Use     []string               `yaml:"use"`
		Partial []rawPartialMatchGroup `yaml:"partial"`
	} `yaml:"match"`
	Source struct {
		Triggers          []rawSignalMatcher   `yaml:"triggers"`
		Amplifiers        []rawSignalMatcher   `yaml:"amplifiers"`
		Mitigations       []rawSignalMatcher   `yaml:"mitigations"`
		Suppressions      []rawSuppressionRule `yaml:"suppressions"`
		Context           []rawSignalMatcher   `yaml:"context"`
		CompoundSignals   []rawCompoundSignal  `yaml:"compound_signals"`
		LocalConsistency  []rawConsistencyRule `yaml:"local_consistency"`
		PathClasses       []rawPathClassRule   `yaml:"path_classes"`
		ChangeSensitivity struct {
			NewFileBonus        float64 `yaml:"new_file_bonus"`
			ModifiedLineBonus   float64 `yaml:"modified_line_bonus"`
			LegacyDiscount      float64 `yaml:"legacy_discount"`
			PreferChangedScopes bool    `yaml:"prefer_changed_scopes"`
		} `yaml:"change_sensitivity"`
		SafeContext []rawSafeContextRule `yaml:"safe_context"`
	} `yaml:"source"`
	Summary       string `yaml:"summary"`
	Diagnosis     string `yaml:"diagnosis"`
	Fix           string `yaml:"fix"`
	Validation    string `yaml:"validation"`
	WhyItMatters  string `yaml:"why_it_matters"`
	RequiresDelta bool   `yaml:"requires_delta"`
	DeltaBoost    []struct {
		Signal string  `yaml:"signal"`
		Weight float64 `yaml:"weight"`
	} `yaml:"delta_boost"`
	Workflow struct {
		LikelyFiles []string `yaml:"likely_files"`
		LocalRepro  []string `yaml:"local_repro"`
		Verify      []string `yaml:"verify"`
	} `yaml:"workflow"`
	Remediation struct {
		Workflows []struct {
			Ref    string `yaml:"ref"`
			Inputs map[string]struct {
				From  string `yaml:"from"`
				Value string `yaml:"value"`
			} `yaml:"inputs"`
		} `yaml:"workflows"`
	} `yaml:"remediation"`
	Hooks    model.PlaybookHooks `yaml:"hooks"`
	Metadata struct {
		SchemaVersion string `yaml:"schema_version"`
	} `yaml:"metadata"`
	Scoring struct {
		BaseTriggerWeight          float64 `yaml:"base_trigger_weight"`
		DefaultAmplifierWeight     float64 `yaml:"default_amplifier_weight"`
		DefaultMitigationDiscount  float64 `yaml:"default_mitigation_discount"`
		DefaultSuppressionDiscount float64 `yaml:"default_suppression_discount"`
		HotPathBonus               float64 `yaml:"hot_path_bonus"`
		BlastRadiusBonus           float64 `yaml:"blast_radius_bonus"`
		SafeContextDiscount        float64 `yaml:"safe_context_discount"`
	} `yaml:"scoring"`
	ContextFilters struct {
		PathIncludes []string `yaml:"path_includes"`
		PathExcludes []string `yaml:"path_excludes"`
	} `yaml:"context_filters"`
	Hypothesis struct {
		Supports       []rawHypothesisSignal        `yaml:"supports"`
		Contradicts    []rawHypothesisSignal        `yaml:"contradicts"`
		Discriminators []rawHypothesisDiscriminator `yaml:"discriminators"`
		Excludes       []rawHypothesisSignal        `yaml:"excludes"`
	} `yaml:"hypothesis"`
}

// LoadDefault loads playbooks from the default directory resolved by
// DefaultDir.
func LoadDefault() ([]model.Playbook, error) {
	dir, err := DefaultDir()
	if err != nil {
		return nil, err
	}
	return LoadDir(dir)
}

// DefaultDir resolves the playbook directory using the following priority:
//  1. FAULTLINE_PLAYBOOK_DIR environment variable
//  2. A "playbooks/bundled" directory found by walking upward from the working
//     directory or the executable directory
//  3. A legacy "playbooks" directory found by the same upward walk
//  4. /playbooks/bundled or /playbooks (Docker container conventions)
func DefaultDir() (string, error) {
	if envDir := strings.TrimSpace(os.Getenv(envKey)); envDir != "" {
		return validateDir(envDir)
	}
	var candidates []string
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, upwardDirs(cwd)...)
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, upwardDirs(filepath.Dir(exe))...)
	}
	candidates = append(candidates, "/playbooks/bundled")
	candidates = append(candidates, "/playbooks")
	seen := make(map[string]struct{})
	for _, c := range candidates {
		if c == "" {
			continue
		}
		c = filepath.Clean(c)
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		if dir, err := validateDir(c); err == nil {
			return dir, nil
		}
	}
	return "", fmt.Errorf(
		"playbook directory not found; set %s or add a playbooks/bundled directory",
		envKey,
	)
}

// LoadDir loads all .yaml/.yml files found recursively under dir.
// Files are processed in lexical order to ensure deterministic loading.
// Duplicate playbook IDs are treated as a hard error.
func LoadDir(dir string) ([]model.Playbook, error) {
	dir, err := validateDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if (strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")) && name != PackMetaFileName && name != HookCatalogFileName && name != MatchCatalogFileName {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk playbook directory: %w", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no playbook files found in %s", dir)
	}
	playbooks := make([]model.Playbook, 0, len(files))
	seenIDs := make(map[string]string, len(files))
	for _, path := range files {
		pb, err := loadFile(path)
		if err != nil {
			return nil, err
		}
		if prev, ok := seenIDs[pb.ID]; ok {
			return nil, fmt.Errorf(
				"duplicate playbook id %q in %s and %s",
				pb.ID, prev, path,
			)
		}
		seenIDs[pb.ID] = path
		playbooks = append(playbooks, pb)
	}
	// Secondary sort by ID for fully deterministic ordering.
	sort.Slice(playbooks, func(i, j int) bool {
		return playbooks[i].ID < playbooks[j].ID
	})
	return playbooks, nil
}

func loadFile(path string) (model.Playbook, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Playbook{}, fmt.Errorf("read playbook %s: %w", path, err)
	}
	var r raw
	if err := yaml.Unmarshal(data, &r); err != nil {
		return model.Playbook{}, fmt.Errorf("parse playbook %s: %w", path, err)
	}
	if err := validate(r, path); err != nil {
		return model.Playbook{}, err
	}
	return model.Playbook{
		ID:         r.ID,
		Extends:    strings.TrimSpace(r.Extends),
		Title:      r.Title,
		Category:   r.Category,
		Severity:   r.Severity,
		Detector:   normalizeDetector(r.Detector),
		BaseScore:  r.BaseScore,
		Tags:       r.Tags,
		StageHints: r.StageHints,
		Match: model.MatchSpec{
			Any:     r.Match.Any,
			All:     r.Match.All,
			None:    r.Match.None,
			Use:     trimStrings(r.Match.Use),
			Partial: convertPartialMatchGroups(r.Match.Partial),
		},
		Source:        convertSourceSpec(r),
		Summary:       normalizeMarkdownBlock(r.Summary),
		Diagnosis:     normalizeMarkdownBlock(r.Diagnosis),
		Fix:           normalizeMarkdownBlock(r.Fix),
		Validation:    normalizeMarkdownBlock(r.Validation),
		WhyItMatters:  normalizeMarkdownBlock(r.WhyItMatters),
		RequiresDelta: r.RequiresDelta,
		DeltaBoost:    convertDeltaBoosts(r.DeltaBoost),
		Workflow: model.WorkflowSpec{
			LikelyFiles: r.Workflow.LikelyFiles,
			LocalRepro:  r.Workflow.LocalRepro,
			Verify:      r.Workflow.Verify,
		},
		Remediation: convertRemediationSpec(r.Remediation.Workflows),
		Hooks:       normalizePlaybookHooks(r.Hooks),
		Metadata: model.PlaybookMeta{
			SchemaVersion: r.Metadata.SchemaVersion,
			SourceFile:    path,
		},
		Scoring: model.ScoringConfig{
			BaseTriggerWeight:          r.Scoring.BaseTriggerWeight,
			DefaultAmplifierWeight:     r.Scoring.DefaultAmplifierWeight,
			DefaultMitigationDiscount:  r.Scoring.DefaultMitigationDiscount,
			DefaultSuppressionDiscount: r.Scoring.DefaultSuppressionDiscount,
			HotPathBonus:               r.Scoring.HotPathBonus,
			BlastRadiusBonus:           r.Scoring.BlastRadiusBonus,
			SafeContextDiscount:        r.Scoring.SafeContextDiscount,
		},
		Contextual: model.ContextPolicy{
			PathIncludes: r.ContextFilters.PathIncludes,
			PathExcludes: r.ContextFilters.PathExcludes,
		},
		Hypothesis: model.HypothesisSpec{
			Supports:       convertHypothesisSignals(r.Hypothesis.Supports),
			Contradicts:    convertHypothesisSignals(r.Hypothesis.Contradicts),
			Discriminators: convertHypothesisDiscriminators(r.Hypothesis.Discriminators),
			Excludes:       convertHypothesisSignals(r.Hypothesis.Excludes),
		},
	}, nil
}

func convertRemediationSpec(rawRefs []struct {
	Ref    string `yaml:"ref"`
	Inputs map[string]struct {
		From  string `yaml:"from"`
		Value string `yaml:"value"`
	} `yaml:"inputs"`
}) model.RemediationSpec {
	if len(rawRefs) == 0 {
		return model.RemediationSpec{}
	}
	workflows := make([]model.RemediationWorkflowRef, 0, len(rawRefs))
	for _, item := range rawRefs {
		ref := strings.TrimSpace(item.Ref)
		if ref == "" {
			continue
		}
		inputs := make(map[string]model.RemediationInputBinding, len(item.Inputs))
		for key, binding := range item.Inputs {
			name := strings.TrimSpace(key)
			if name == "" {
				continue
			}
			inputs[name] = model.RemediationInputBinding{
				From:  strings.TrimSpace(binding.From),
				Value: strings.TrimSpace(binding.Value),
			}
		}
		workflows = append(workflows, model.RemediationWorkflowRef{
			Ref:    ref,
			Inputs: inputs,
		})
	}
	return model.RemediationSpec{Workflows: workflows}
}

func convertDeltaBoosts(rawBoosts []struct {
	Signal string  `yaml:"signal"`
	Weight float64 `yaml:"weight"`
}) []model.DeltaBoost {
	if len(rawBoosts) == 0 {
		return nil
	}
	out := make([]model.DeltaBoost, 0, len(rawBoosts))
	for _, item := range rawBoosts {
		signal := strings.TrimSpace(item.Signal)
		if signal == "" {
			continue
		}
		out = append(out, model.DeltaBoost{
			Signal: signal,
			Weight: item.Weight,
		})
	}
	return out
}

func normalizeMarkdownBlock(value string) string {
	return strings.TrimSpace(value)
}

func validate(r raw, path string) error {
	inherits := strings.TrimSpace(r.Extends) != ""
	if strings.TrimSpace(r.ID) == "" {
		return fmt.Errorf("playbook %s: missing required field 'id'", path)
	}
	if !inherits && strings.TrimSpace(r.Title) == "" {
		return fmt.Errorf("playbook %s: missing required field 'title'", path)
	}
	if !inherits && strings.TrimSpace(r.Summary) == "" {
		return fmt.Errorf("playbook %s: missing required field 'summary'", path)
	}
	if !inherits && strings.TrimSpace(r.Diagnosis) == "" {
		return fmt.Errorf("playbook %s: missing required field 'diagnosis'", path)
	}
	if !inherits && strings.TrimSpace(r.Fix) == "" {
		return fmt.Errorf("playbook %s: missing required field 'fix'", path)
	}
	if !inherits && strings.TrimSpace(r.Validation) == "" {
		return fmt.Errorf("playbook %s: missing required field 'validation'", path)
	}
	detector := normalizeDetector(r.Detector)
	if detector == "" && !inherits {
		detector = "log"
	}
	if detector == "log" && !inherits && len(r.Match.Any) == 0 && len(r.Match.All) == 0 && len(r.Match.Use) == 0 && len(r.Match.Partial) == 0 {
		return fmt.Errorf(
			"playbook %s: must define at least one matcher in match.any, match.all, match.use, or match.partial",
			path,
		)
	}
	switch detector {
	case "log":
		if len(r.Match.Any) > 0 {
			if err := validatePatterns(r.Match.Any, "match.any", path); err != nil {
				return err
			}
		}
		if len(r.Match.All) > 0 {
			if err := validatePatterns(r.Match.All, "match.all", path); err != nil {
				return err
			}
		}
		if len(r.Match.None) > 0 {
			if err := validatePatterns(r.Match.None, "match.none", path); err != nil {
				return err
			}
		}
		if err := validateMatchRefs(r.Match.Use, path); err != nil {
			return err
		}
		if err := validatePartialMatchGroups(r.Match.Partial, path); err != nil {
			return err
		}
		if err := validateExclusions(r.Match.Any, r.Match.All, r.Match.None, path); err != nil {
			return err
		}
	case "source":
		if err := validateSource(r, path); err != nil {
			return err
		}
	case "":
		if !inherits {
			return fmt.Errorf("playbook %s: unknown detector %q", path, detector)
		}
	default:
		return fmt.Errorf("playbook %s: unknown detector %q", path, detector)
	}
	if err := validateHypothesis(r, path); err != nil {
		return err
	}
	if err := validatePlaybookHooks(r.Hooks, path); err != nil {
		return err
	}
	return nil
}

func validateHypothesis(r raw, path string) error {
	validateSignal := func(signal, section string) error {
		signal = strings.TrimSpace(signal)
		if signal == "" {
			return fmt.Errorf("playbook %s: %s must not be empty", path, section)
		}
		if !hypothesis.ValidSignal(signal) {
			return fmt.Errorf("playbook %s: %s references unknown signal %q", path, section, signal)
		}
		return nil
	}

	for i, item := range r.Hypothesis.Supports {
		if err := validateSignal(item.Signal, fmt.Sprintf("hypothesis.supports[%d].signal", i)); err != nil {
			return err
		}
	}
	for i, item := range r.Hypothesis.Contradicts {
		if err := validateSignal(item.Signal, fmt.Sprintf("hypothesis.contradicts[%d].signal", i)); err != nil {
			return err
		}
	}
	for i, item := range r.Hypothesis.Discriminators {
		if err := validateSignal(item.Signal, fmt.Sprintf("hypothesis.discriminators[%d].signal", i)); err != nil {
			return err
		}
	}
	for i, item := range r.Hypothesis.Excludes {
		if err := validateSignal(item.Signal, fmt.Sprintf("hypothesis.excludes[%d].signal", i)); err != nil {
			return err
		}
	}
	return nil
}

func validateSource(r raw, path string) error {
	if len(r.Source.Triggers) == 0 {
		return fmt.Errorf("playbook %s: source detector requires at least one trigger", path)
	}
	for i, matcher := range r.Source.Triggers {
		if err := validateSignalPatterns(matcher.Patterns, fmt.Sprintf("source.triggers[%d].patterns", i), path); err != nil {
			return err
		}
	}
	for i, matcher := range r.Source.Amplifiers {
		if err := validateSignalPatterns(matcher.Patterns, fmt.Sprintf("source.amplifiers[%d].patterns", i), path); err != nil {
			return err
		}
	}
	for i, matcher := range r.Source.Mitigations {
		if err := validateSignalPatterns(matcher.Patterns, fmt.Sprintf("source.mitigations[%d].patterns", i), path); err != nil {
			return err
		}
	}
	for i, matcher := range r.Source.Context {
		if err := validateSignalPatterns(matcher.Patterns, fmt.Sprintf("source.context[%d].patterns", i), path); err != nil {
			return err
		}
	}
	return nil
}

func validateSignalPatterns(patterns []string, section, path string) error {
	if len(patterns) == 0 {
		return fmt.Errorf("playbook %s: %s must define at least one pattern", path, section)
	}
	return validatePatterns(patterns, section, path)
}

func validatePatterns(patterns []string, section, path string) error {
	for i, pattern := range patterns {
		norm := normalizePattern(pattern)
		if norm == "" {
			return fmt.Errorf("playbook %s: %s[%d] must not be empty", path, section, i)
		}
	}
	return nil
}

func validateMatchRefs(refs []string, path string) error {
	for i, ref := range refs {
		if strings.TrimSpace(ref) == "" {
			return fmt.Errorf("playbook %s: match.use[%d] must not be empty", path, i)
		}
	}
	return nil
}

func validatePartialMatchGroups(groups []rawPartialMatchGroup, path string) error {
	for i, group := range groups {
		section := fmt.Sprintf("match.partial[%d]", i)
		if len(group.Patterns) == 0 {
			return fmt.Errorf("playbook %s: %s must define at least one pattern", path, section)
		}
		if err := validatePatterns(group.Patterns, section+".patterns", path); err != nil {
			return err
		}
		if group.Minimum <= 0 {
			return fmt.Errorf("playbook %s: %s.minimum must be greater than zero", path, section)
		}
		if group.Minimum > len(group.Patterns) {
			return fmt.Errorf("playbook %s: %s.minimum exceeds pattern count", path, section)
		}
	}
	return nil
}

func validateExclusions(any, all, none []string, path string) error {
	positive := make(map[string]string, len(any)+len(all))
	for i, pattern := range any {
		positive[normalizePattern(pattern)] = fmt.Sprintf("match.any[%d]", i)
	}
	for i, pattern := range all {
		norm := normalizePattern(pattern)
		if prev, ok := positive[norm]; ok {
			return fmt.Errorf(
				"playbook %s: match.all[%d] %q overlaps with %s",
				path, i, pattern, prev,
			)
		}
		positive[norm] = fmt.Sprintf("match.all[%d]", i)
	}
	for i, pattern := range none {
		norm := normalizePattern(pattern)
		if prev, ok := positive[norm]; ok {
			return fmt.Errorf(
				"playbook %s: match.none[%d] %q overlaps with %s",
				path, i, pattern, prev,
			)
		}
	}
	return nil
}

func normalizePattern(pattern string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(pattern)), " "))
}

func normalizeDetector(detector string) string {
	return strings.ToLower(strings.TrimSpace(detector))
}

func convertSourceSpec(r raw) model.SourceSpec {
	return model.SourceSpec{
		Triggers:         convertSignalMatchers(r.Source.Triggers),
		Amplifiers:       convertSignalMatchers(r.Source.Amplifiers),
		Mitigations:      convertSignalMatchers(r.Source.Mitigations),
		Suppressions:     convertSuppressions(r.Source.Suppressions),
		Context:          convertSignalMatchers(r.Source.Context),
		CompoundSignals:  convertCompoundSignals(r.Source.CompoundSignals),
		LocalConsistency: convertConsistencyRules(r.Source.LocalConsistency),
		PathClasses:      convertPathClassRules(r.Source.PathClasses),
		ChangeSensitivity: model.ChangeSensitivity{
			NewFileBonus:        r.Source.ChangeSensitivity.NewFileBonus,
			ModifiedLineBonus:   r.Source.ChangeSensitivity.ModifiedLineBonus,
			LegacyDiscount:      r.Source.ChangeSensitivity.LegacyDiscount,
			PreferChangedScopes: r.Source.ChangeSensitivity.PreferChangedScopes,
		},
		SafeContextClasses: convertSafeContextRules(r.Source.SafeContext),
	}
}

func convertSignalMatchers(items []rawSignalMatcher) []model.SignalMatcher {
	out := make([]model.SignalMatcher, 0, len(items))
	for _, item := range items {
		out = append(out, model.SignalMatcher{
			ID:           item.ID,
			Label:        item.Label,
			Description:  item.Description,
			Patterns:     item.Patterns,
			PathIncludes: item.PathIncludes,
			PathExcludes: item.PathExcludes,
			Scopes:       item.Scopes,
			Weight:       item.Weight,
			Required:     item.Required,
		})
	}
	return out
}

func convertSuppressions(items []rawSuppressionRule) []model.SuppressionRule {
	out := make([]model.SuppressionRule, 0, len(items))
	for _, item := range items {
		out = append(out, model.SuppressionRule(item))
	}
	return out
}

func convertCompoundSignals(items []rawCompoundSignal) []model.CompoundSignal {
	out := make([]model.CompoundSignal, 0, len(items))
	for _, item := range items {
		out = append(out, model.CompoundSignal(item))
	}
	return out
}

func convertConsistencyRules(items []rawConsistencyRule) []model.ConsistencyRule {
	out := make([]model.ConsistencyRule, 0, len(items))
	for _, item := range items {
		out = append(out, model.ConsistencyRule(item))
	}
	return out
}

func convertHypothesisSignals(items []rawHypothesisSignal) []model.HypothesisSignal {
	out := make([]model.HypothesisSignal, 0, len(items))
	for _, item := range items {
		out = append(out, model.HypothesisSignal{
			Signal: item.Signal,
			Weight: item.Weight,
		})
	}
	return out
}

func convertHypothesisDiscriminators(items []rawHypothesisDiscriminator) []model.HypothesisDiscriminator {
	out := make([]model.HypothesisDiscriminator, 0, len(items))
	for _, item := range items {
		out = append(out, model.HypothesisDiscriminator{
			Description: item.Description,
			Signal:      item.Signal,
			Weight:      item.Weight,
		})
	}
	return out
}

func convertPathClassRules(items []rawPathClassRule) []model.PathClassRule {
	out := make([]model.PathClassRule, 0, len(items))
	for _, item := range items {
		out = append(out, model.PathClassRule(item))
	}
	return out
}

func convertSafeContextRules(items []rawSafeContextRule) []model.SafeContextRule {
	out := make([]model.SafeContextRule, 0, len(items))
	for _, item := range items {
		out = append(out, model.SafeContextRule(item))
	}
	return out
}

func convertPartialMatchGroups(items []rawPartialMatchGroup) []model.PartialMatchGroup {
	out := make([]model.PartialMatchGroup, 0, len(items))
	for _, item := range items {
		out = append(out, model.PartialMatchGroup{
			ID:       strings.TrimSpace(item.ID),
			Label:    strings.TrimSpace(item.Label),
			Minimum:  item.Minimum,
			Patterns: append([]string(nil), item.Patterns...),
		})
	}
	return out
}

func trimStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func validateDir(dir string) (string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("playbook directory %s: %w", dir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dir)
	}
	return dir, nil
}

// upwardDirs returns a list of playbook directory candidates by walking upward
// from dir toward the filesystem root. Bundled-pack locations are preferred,
// with the legacy single-root layout kept as a fallback for compatibility.
func upwardDirs(dir string) []string {
	var result []string
	for {
		result = append(result, filepath.Join(dir, "playbooks", "bundled"))
		result = append(result, filepath.Join(dir, "playbooks"))
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return result
}
