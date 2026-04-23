package bind

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"faultline/internal/model"
	"faultline/internal/workflow/schema"
)

var placeholderPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

type RuntimeContext struct {
	WorkDir  string
	RepoRoot string
}

type State struct {
	Inputs   map[string]string
	Runtime  RuntimeContext
	Artifact *model.FailureArtifact
	Steps    map[string]map[string]string
}

func ResolveInputs(def schema.Definition, recommendation model.ArtifactWorkflowRecommendation) (map[string]string, error) {
	values := map[string]string{}
	inputNames := make([]string, 0, len(def.Inputs))
	for name := range def.Inputs {
		inputNames = append(inputNames, name)
	}
	sort.Strings(inputNames)
	for _, name := range inputNames {
		spec := def.Inputs[name]
		value := strings.TrimSpace(recommendation.Inputs[name])
		if value == "" {
			value = strings.TrimSpace(spec.Default)
		}
		if value == "" && spec.Required {
			return nil, fmt.Errorf("workflow %s requires input %s", def.WorkflowID, name)
		}
		if value != "" {
			values[name] = value
		}
	}
	return values, nil
}

func ResolveValue(value any, state State, allowDeferredStepRefs bool) (any, error) {
	switch typed := value.(type) {
	case string:
		return resolveString(typed, state, allowDeferredStepRefs)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			resolved, err := ResolveValue(item, state, allowDeferredStepRefs)
			if err != nil {
				return nil, err
			}
			out = append(out, resolved)
		}
		return out, nil
	case map[string]any:
		out := make(map[string]any, len(typed))
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			resolved, err := ResolveValue(typed[key], state, allowDeferredStepRefs)
			if err != nil {
				return nil, err
			}
			out[key] = resolved
		}
		return out, nil
	default:
		return value, nil
	}
}

func References(value any) []string {
	set := map[string]struct{}{}
	collectReferences(value, set)
	out := make([]string, 0, len(set))
	for ref := range set {
		out = append(out, ref)
	}
	sort.Strings(out)
	return out
}

func collectReferences(value any, set map[string]struct{}) {
	switch typed := value.(type) {
	case string:
		for _, match := range placeholderPattern.FindAllStringSubmatch(typed, -1) {
			if len(match) == 2 {
				set[strings.TrimSpace(match[1])] = struct{}{}
			}
		}
	case []any:
		for _, item := range typed {
			collectReferences(item, set)
		}
	case map[string]any:
		for _, item := range typed {
			collectReferences(item, set)
		}
	}
}

func resolveString(value string, state State, allowDeferredStepRefs bool) (string, error) {
	matches := placeholderPattern.FindAllStringSubmatch(value, -1)
	if len(matches) == 0 {
		return value, nil
	}
	resolved := value
	for _, match := range matches {
		ref := strings.TrimSpace(match[1])
		found, ok, err := resolveReference(ref, state)
		if err != nil {
			return "", err
		}
		if !ok {
			if strings.HasPrefix(ref, "steps.") && allowDeferredStepRefs {
				continue
			}
			return "", fmt.Errorf("unresolved reference %q", ref)
		}
		resolved = strings.ReplaceAll(resolved, match[0], found)
	}
	return resolved, nil
}

func resolveReference(ref string, state State) (string, bool, error) {
	switch {
	case strings.HasPrefix(ref, "inputs."):
		key := strings.TrimSpace(strings.TrimPrefix(ref, "inputs."))
		value, ok := state.Inputs[key]
		return value, ok, nil
	case strings.HasPrefix(ref, "steps."):
		parts := strings.Split(ref, ".")
		if len(parts) != 3 {
			return "", false, fmt.Errorf("invalid step reference %q", ref)
		}
		outputs, ok := state.Steps[parts[1]]
		if !ok {
			return "", false, nil
		}
		value, ok := outputs[parts[2]]
		return value, ok, nil
	case ref == "runtime.workdir":
		return strings.TrimSpace(state.Runtime.WorkDir), strings.TrimSpace(state.Runtime.WorkDir) != "", nil
	case ref == "runtime.repo_root":
		return strings.TrimSpace(state.Runtime.RepoRoot), strings.TrimSpace(state.Runtime.RepoRoot) != "", nil
	case ref == "artifact.fingerprint":
		if state.Artifact == nil {
			return "", false, nil
		}
		value := strings.TrimSpace(state.Artifact.Fingerprint)
		return value, value != "", nil
	case ref == "artifact.matched_playbook.id":
		if state.Artifact == nil || state.Artifact.MatchedPlaybook == nil {
			return "", false, nil
		}
		value := strings.TrimSpace(state.Artifact.MatchedPlaybook.ID)
		return value, value != "", nil
	case strings.HasPrefix(ref, "artifact.facts."):
		if state.Artifact == nil {
			return "", false, nil
		}
		key := strings.TrimSpace(strings.TrimPrefix(ref, "artifact.facts."))
		value, ok := state.Artifact.Facts[key]
		return strings.TrimSpace(value), ok && strings.TrimSpace(value) != "", nil
	default:
		return "", false, fmt.Errorf("unsupported reference %q", ref)
	}
}
