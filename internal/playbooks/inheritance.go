package playbooks

import (
	"fmt"
	"strings"

	"faultline/internal/model"
)

func resolvePlaybookInheritance(pbs []model.Playbook) ([]model.Playbook, error) {
	if len(pbs) == 0 {
		return nil, nil
	}
	index := make(map[string]model.Playbook, len(pbs))
	for _, pb := range pbs {
		index[pb.ID] = pb
	}

	resolved := make(map[string]model.Playbook, len(pbs))
	visiting := make(map[string]bool, len(pbs))

	var resolve func(id string) (model.Playbook, error)
	resolve = func(id string) (model.Playbook, error) {
		if pb, ok := resolved[id]; ok {
			return pb, nil
		}
		pb, ok := index[id]
		if !ok {
			return model.Playbook{}, fmt.Errorf("unknown playbook %q", id)
		}
		parentID := strings.TrimSpace(pb.Extends)
		if parentID == "" {
			resolved[id] = pb
			return pb, nil
		}
		if visiting[id] {
			return model.Playbook{}, fmt.Errorf("playbook %q forms an inheritance cycle", id)
		}
		visiting[id] = true
		parent, err := resolve(parentID)
		if err != nil {
			return model.Playbook{}, fmt.Errorf("playbook %q extends %q: %w", id, parentID, err)
		}
		merged := mergePlaybooks(parent, pb)
		visiting[id] = false
		resolved[id] = merged
		return merged, nil
	}

	out := make([]model.Playbook, len(pbs))
	for i, pb := range pbs {
		merged, err := resolve(pb.ID)
		if err != nil {
			return nil, err
		}
		out[i] = merged
	}
	return out, nil
}

func mergePlaybooks(base, child model.Playbook) model.Playbook {
	merged := child
	merged.Title = firstNonEmptyInherited(child.Title, base.Title)
	merged.Category = firstNonEmptyInherited(child.Category, base.Category)
	merged.Severity = firstNonEmptyInherited(child.Severity, base.Severity)
	merged.Detector = firstNonEmptyInherited(child.Detector, base.Detector)
	if child.BaseScore == 0 {
		merged.BaseScore = base.BaseScore
	}
	merged.Tags = mergeUnique(base.Tags, child.Tags)
	merged.StageHints = mergeUnique(base.StageHints, child.StageHints)
	merged.Match = model.MatchSpec{
		Any:     mergeUnique(base.Match.Any, child.Match.Any),
		All:     mergeUnique(base.Match.All, child.Match.All),
		None:    mergeUnique(base.Match.None, child.Match.None),
		Use:     mergeUnique(base.Match.Use, child.Match.Use),
		Partial: mergePartialGroups(base.Match.Partial, child.Match.Partial),
	}
	merged.Source = mergeSourceSpec(base.Source, child.Source)
	merged.Summary = firstNonEmptyInherited(child.Summary, base.Summary)
	merged.Diagnosis = firstNonEmptyInherited(child.Diagnosis, base.Diagnosis)
	merged.Fix = firstNonEmptyInherited(child.Fix, base.Fix)
	merged.Validation = firstNonEmptyInherited(child.Validation, base.Validation)
	merged.WhyItMatters = firstNonEmptyInherited(child.WhyItMatters, base.WhyItMatters)
	merged.RequiresDelta = child.RequiresDelta || base.RequiresDelta
	merged.DeltaBoost = append(append([]model.DeltaBoost(nil), base.DeltaBoost...), child.DeltaBoost...)
	merged.RequiresTopology = child.RequiresTopology || base.RequiresTopology
	merged.TopologyBoost = append(append([]model.TopologyBoost(nil), base.TopologyBoost...), child.TopologyBoost...)
	merged.Workflow = model.WorkflowSpec{
		LikelyFiles: mergeUnique(base.Workflow.LikelyFiles, child.Workflow.LikelyFiles),
		LocalRepro:  mergeUnique(base.Workflow.LocalRepro, child.Workflow.LocalRepro),
		Verify:      mergeUnique(base.Workflow.Verify, child.Workflow.Verify),
	}
	merged.Hooks = mergePlaybookHooks(base.Hooks, child.Hooks)
	merged.Scoring = mergeScoringConfig(base.Scoring, child.Scoring)
	merged.Contextual = model.ContextPolicy{
		PathIncludes: mergeUnique(base.Contextual.PathIncludes, child.Contextual.PathIncludes),
		PathExcludes: mergeUnique(base.Contextual.PathExcludes, child.Contextual.PathExcludes),
	}
	merged.Hypothesis = model.HypothesisSpec{
		Supports:       append(append([]model.HypothesisSignal(nil), base.Hypothesis.Supports...), child.Hypothesis.Supports...),
		Contradicts:    append(append([]model.HypothesisSignal(nil), base.Hypothesis.Contradicts...), child.Hypothesis.Contradicts...),
		Discriminators: append(append([]model.HypothesisDiscriminator(nil), base.Hypothesis.Discriminators...), child.Hypothesis.Discriminators...),
		Excludes:       append(append([]model.HypothesisSignal(nil), base.Hypothesis.Excludes...), child.Hypothesis.Excludes...),
	}
	if strings.TrimSpace(merged.Metadata.SchemaVersion) == "" {
		merged.Metadata.SchemaVersion = base.Metadata.SchemaVersion
	}
	return merged
}

func mergeSourceSpec(base, child model.SourceSpec) model.SourceSpec {
	return model.SourceSpec{
		Triggers:           append(append([]model.SignalMatcher(nil), base.Triggers...), child.Triggers...),
		Amplifiers:         append(append([]model.SignalMatcher(nil), base.Amplifiers...), child.Amplifiers...),
		Mitigations:        append(append([]model.SignalMatcher(nil), base.Mitigations...), child.Mitigations...),
		Suppressions:       append(append([]model.SuppressionRule(nil), base.Suppressions...), child.Suppressions...),
		Context:            append(append([]model.SignalMatcher(nil), base.Context...), child.Context...),
		CompoundSignals:    append(append([]model.CompoundSignal(nil), base.CompoundSignals...), child.CompoundSignals...),
		LocalConsistency:   append(append([]model.ConsistencyRule(nil), base.LocalConsistency...), child.LocalConsistency...),
		PathClasses:        append(append([]model.PathClassRule(nil), base.PathClasses...), child.PathClasses...),
		ChangeSensitivity:  mergeChangeSensitivity(base.ChangeSensitivity, child.ChangeSensitivity),
		SafeContextClasses: append(append([]model.SafeContextRule(nil), base.SafeContextClasses...), child.SafeContextClasses...),
	}
}

func mergeChangeSensitivity(base, child model.ChangeSensitivity) model.ChangeSensitivity {
	merged := child
	if merged.NewFileBonus == 0 {
		merged.NewFileBonus = base.NewFileBonus
	}
	if merged.ModifiedLineBonus == 0 {
		merged.ModifiedLineBonus = base.ModifiedLineBonus
	}
	if merged.LegacyDiscount == 0 {
		merged.LegacyDiscount = base.LegacyDiscount
	}
	merged.PreferChangedScopes = child.PreferChangedScopes || base.PreferChangedScopes
	return merged
}

func mergePlaybookHooks(base, child model.PlaybookHooks) model.PlaybookHooks {
	trimmedBase := disableHookIDs(base, child.Disable)
	return model.PlaybookHooks{
		Verify:    append(append([]model.HookDefinition(nil), trimmedBase.Verify...), child.Verify...),
		Collect:   append(append([]model.HookDefinition(nil), trimmedBase.Collect...), child.Collect...),
		Remediate: append(append([]model.HookDefinition(nil), trimmedBase.Remediate...), child.Remediate...),
	}
}

func mergeScoringConfig(base, child model.ScoringConfig) model.ScoringConfig {
	merged := child
	if merged.BaseTriggerWeight == 0 {
		merged.BaseTriggerWeight = base.BaseTriggerWeight
	}
	if merged.DefaultAmplifierWeight == 0 {
		merged.DefaultAmplifierWeight = base.DefaultAmplifierWeight
	}
	if merged.DefaultMitigationDiscount == 0 {
		merged.DefaultMitigationDiscount = base.DefaultMitigationDiscount
	}
	if merged.DefaultSuppressionDiscount == 0 {
		merged.DefaultSuppressionDiscount = base.DefaultSuppressionDiscount
	}
	if merged.HotPathBonus == 0 {
		merged.HotPathBonus = base.HotPathBonus
	}
	if merged.BlastRadiusBonus == 0 {
		merged.BlastRadiusBonus = base.BlastRadiusBonus
	}
	if merged.SafeContextDiscount == 0 {
		merged.SafeContextDiscount = base.SafeContextDiscount
	}
	return merged
}

func mergeUnique(base, child []string) []string {
	if len(base) == 0 && len(child) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(base)+len(child))
	add := func(items []string) {
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
	}
	add(base)
	add(child)
	return out
}

func mergePartialGroups(base, child []model.PartialMatchGroup) []model.PartialMatchGroup {
	if len(base) == 0 && len(child) == 0 {
		return nil
	}
	out := make([]model.PartialMatchGroup, 0, len(base)+len(child))
	seen := make(map[string]struct{}, len(base)+len(child))
	add := func(items []model.PartialMatchGroup) {
		for _, item := range items {
			key := partialGroupKey(item)
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			copied := item
			copied.Patterns = append([]string(nil), item.Patterns...)
			out = append(out, copied)
		}
	}
	add(base)
	add(child)
	return out
}

func partialGroupKey(group model.PartialMatchGroup) string {
	if strings.TrimSpace(group.ID) != "" {
		return "id:" + strings.TrimSpace(group.ID)
	}
	parts := append([]string(nil), group.Patterns...)
	return fmt.Sprintf("%d|%s", group.Minimum, strings.Join(parts, "\x00"))
}

func firstNonEmptyInherited(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
