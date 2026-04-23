package plan

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"faultline/internal/model"
	"faultline/internal/workflow/bind"
	"faultline/internal/workflow/registry"
	"faultline/internal/workflow/schema"
)

type Step struct {
	Phase          string
	ID             string
	Type           string
	SafetyClass    registry.SafetyClass
	Idempotence    string
	Description    string
	ResolvedArgs   map[string]any
	ResolvedExpect map[string]any
	DecodedArgs    any
	DecodedExpect  any
	Handler        registry.StepType
}

type Plan struct {
	SchemaVersion     string                 `json:"schema_version"`
	DefinitionVersion string                 `json:"definition_schema_version"`
	WorkflowID        string                 `json:"workflow_id"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	SourceFingerprint string                 `json:"source_fingerprint,omitempty"`
	SourceFailureID   string                 `json:"source_failure_id,omitempty"`
	ResolvedInputs    map[string]string      `json:"resolved_inputs,omitempty"`
	Steps             []Step                 `json:"-"`
	Verification      []Step                 `json:"-"`
	RequiredSafety    []registry.SafetyClass `json:"required_safety,omitempty"`
}

type BuildOptions struct {
	Recommendation model.ArtifactWorkflowRecommendation
	Artifact       *model.FailureArtifact
	Runtime        bind.RuntimeContext
	Registry       *registry.Registry
	ProbeReadSteps bool
}

func Build(ctx context.Context, def schema.Definition, opts BuildOptions) (Plan, error) {
	reg := opts.Registry
	if reg == nil {
		reg = registry.Default()
	}
	inputs, err := bind.ResolveInputs(def, opts.Recommendation)
	if err != nil {
		return Plan{}, err
	}
	state := bind.State{
		Inputs:   inputs,
		Runtime:  opts.Runtime,
		Artifact: opts.Artifact,
		Steps:    map[string]map[string]string{},
	}
	compiled := Plan{
		SchemaVersion:     "workflow_plan.v1",
		DefinitionVersion: def.SchemaVersion,
		WorkflowID:        def.WorkflowID,
		Title:             def.Title,
		Description:       def.Description,
		ResolvedInputs:    inputs,
	}
	if opts.Artifact != nil {
		compiled.SourceFingerprint = strings.TrimSpace(opts.Artifact.Fingerprint)
		if opts.Artifact.MatchedPlaybook != nil {
			compiled.SourceFailureID = strings.TrimSpace(opts.Artifact.MatchedPlaybook.ID)
		}
	}
	knownOutputs := map[string][]string{}
	steps, err := compileSteps(ctx, "steps", def.Steps, reg, state, knownOutputs, opts.ProbeReadSteps)
	if err != nil {
		return Plan{}, err
	}
	compiled.Steps = steps
	verification, err := compileSteps(ctx, "verification", def.Verification, reg, state, knownOutputs, opts.ProbeReadSteps)
	if err != nil {
		return Plan{}, err
	}
	compiled.Verification = verification
	compiled.RequiredSafety = requiredSafety(append(append([]Step{}, compiled.Steps...), compiled.Verification...))
	return compiled, nil
}

func compileSteps(ctx context.Context, phase string, steps []schema.Step, reg *registry.Registry, state bind.State, knownOutputs map[string][]string, probeRead bool) ([]Step, error) {
	compiled := make([]Step, 0, len(steps))
	for _, raw := range steps {
		handler, err := reg.MustLookup(raw.Type)
		if err != nil {
			return nil, err
		}
		if err := validateReferences(raw.Args, knownOutputs, state); err != nil {
			return nil, fmt.Errorf("workflow step %s: %w", raw.ID, err)
		}
		if err := validateReferences(raw.Expect, knownOutputs, state); err != nil {
			return nil, fmt.Errorf("workflow step %s expectation: %w", raw.ID, err)
		}
		resolvedArgsAny, err := bind.ResolveValue(raw.Args, state, true)
		if err != nil {
			return nil, fmt.Errorf("resolve args for %s: %w", raw.ID, err)
		}
		resolvedArgs := mapFrom(resolvedArgsAny)
		decodedArgs, err := handler.DecodeArgs(resolvedArgs)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", raw.ID, err)
		}
		resolvedExpectAny, err := bind.ResolveValue(raw.Expect, state, true)
		if err != nil {
			return nil, fmt.Errorf("resolve expect for %s: %w", raw.ID, err)
		}
		resolvedExpect := mapFrom(resolvedExpectAny)
		decodedExpect, err := handler.DecodeExpect(resolvedExpect)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", raw.ID, err)
		}
		item := Step{
			Phase:          phase,
			ID:             raw.ID,
			Type:           raw.Type,
			SafetyClass:    handler.Safety(decodedArgs),
			Idempotence:    handler.Idempotence(),
			Description:    handler.DryRun(decodedArgs),
			ResolvedArgs:   resolvedArgs,
			ResolvedExpect: resolvedExpect,
			DecodedArgs:    decodedArgs,
			DecodedExpect:  decodedExpect,
			Handler:        handler,
		}
		compiled = append(compiled, item)
		knownOutputs[raw.ID] = handler.KnownOutputs()
		if probeRead && item.SafetyClass == registry.SafetyClassRead {
			result, err := handler.Execute(ctx, registry.Runtime{WorkDir: state.Runtime.WorkDir}, decodedArgs)
			if err != nil {
				return nil, fmt.Errorf("probe step %s: %w", raw.ID, err)
			}
			state.Steps[raw.ID] = result.Outputs
		}
	}
	return compiled, nil
}

func validateReferences(values map[string]any, knownOutputs map[string][]string, state bind.State) error {
	for _, ref := range bind.References(values) {
		switch {
		case strings.HasPrefix(ref, "inputs."):
			name := strings.TrimSpace(strings.TrimPrefix(ref, "inputs."))
			if _, ok := state.Inputs[name]; !ok {
				return fmt.Errorf("unknown input reference %q", ref)
			}
		case strings.HasPrefix(ref, "steps."):
			parts := strings.Split(ref, ".")
			if len(parts) != 3 {
				return fmt.Errorf("invalid step reference %q", ref)
			}
			outputs, ok := knownOutputs[parts[1]]
			if !ok {
				return fmt.Errorf("references future or unknown step %q", parts[1])
			}
			if !contains(outputs, parts[2]) {
				return fmt.Errorf("references unknown output %q on step %q", parts[2], parts[1])
			}
		case ref == "runtime.workdir", ref == "runtime.repo_root":
		case ref == "artifact.fingerprint", ref == "artifact.matched_playbook.id":
		case strings.HasPrefix(ref, "artifact.facts."):
		default:
			return fmt.Errorf("unsupported reference %q", ref)
		}
	}
	return nil
}

func requiredSafety(steps []Step) []registry.SafetyClass {
	seen := map[registry.SafetyClass]struct{}{}
	var out []registry.SafetyClass
	for _, step := range steps {
		if step.SafetyClass == registry.SafetyClassRead {
			continue
		}
		if _, ok := seen[step.SafetyClass]; ok {
			continue
		}
		seen[step.SafetyClass] = struct{}{}
		out = append(out, step.SafetyClass)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func mapFrom(value any) map[string]any {
	if value == nil {
		return nil
	}
	typed, _ := value.(map[string]any)
	return typed
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
