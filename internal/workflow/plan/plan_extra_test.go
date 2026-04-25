package plan

import (
	"context"
	"testing"

	"faultline/internal/model"
	"faultline/internal/workflow/bind"
	"faultline/internal/workflow/registry"
	"faultline/internal/workflow/schema"
)

// --- mapFrom ---

func TestMapFromNil(t *testing.T) {
	if got := mapFrom(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestMapFromNonMap(t *testing.T) {
	if got := mapFrom("not a map"); got != nil {
		t.Fatalf("expected nil for non-map, got %v", got)
	}
}

func TestMapFromValidMap(t *testing.T) {
	m := map[string]any{"key": "val", "num": 42}
	got := mapFrom(m)
	if got == nil {
		t.Fatal("expected non-nil map")
	}
	if got["key"] != "val" {
		t.Errorf("expected key=val, got %v", got["key"])
	}
}

// --- contains ---

func TestContainsFound(t *testing.T) {
	if !contains([]string{"a", "b", "c"}, "b") {
		t.Error("expected to find 'b'")
	}
}

func TestContainsNotFound(t *testing.T) {
	if contains([]string{"a", "b", "c"}, "d") {
		t.Error("expected not to find 'd'")
	}
}

func TestContainsEmptySlice(t *testing.T) {
	if contains(nil, "x") {
		t.Error("expected false for empty slice")
	}
}

// --- requiredSafety ---

func TestRequiredSafetyExcludesRead(t *testing.T) {
	steps := []Step{
		{SafetyClass: registry.SafetyClassRead},
		{SafetyClass: registry.SafetyClassRead},
	}
	got := requiredSafety(steps)
	if len(got) != 0 {
		t.Errorf("expected no required safety for read-only steps, got %v", got)
	}
}

func TestRequiredSafetyDedupesClasses(t *testing.T) {
	steps := []Step{
		{SafetyClass: registry.SafetyClassEnvironmentMutation},
		{SafetyClass: registry.SafetyClassEnvironmentMutation},
		{SafetyClass: registry.SafetyClassLocalMutation},
	}
	got := requiredSafety(steps)
	if len(got) != 2 {
		t.Errorf("expected 2 unique classes, got %v", got)
	}
}

func TestRequiredSafetyIsSorted(t *testing.T) {
	steps := []Step{
		{SafetyClass: registry.SafetyClassExternalSideEffect},
		{SafetyClass: registry.SafetyClassLocalMutation},
		{SafetyClass: registry.SafetyClassEnvironmentMutation},
	}
	got := requiredSafety(steps)
	if len(got) != 3 {
		t.Fatalf("expected 3 classes, got %v", got)
	}
	for i := 1; i < len(got); i++ {
		if got[i] < got[i-1] {
			t.Errorf("expected sorted safety classes, got %v", got)
		}
	}
}

func TestRequiredSafetyMixedWithRead(t *testing.T) {
	steps := []Step{
		{SafetyClass: registry.SafetyClassRead},
		{SafetyClass: registry.SafetyClassLocalMutation},
		{SafetyClass: registry.SafetyClassRead},
	}
	got := requiredSafety(steps)
	if len(got) != 1 {
		t.Fatalf("expected 1 required class (excluding read), got %v", got)
	}
	if got[0] != registry.SafetyClassLocalMutation {
		t.Errorf("expected local_mutation, got %v", got[0])
	}
}

// --- validateReferences error paths (via Build) ---

func buildDef(t *testing.T, yaml string) schema.Definition {
	t.Helper()
	def, err := schema.LoadBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	return def
}

func TestBuildRejectsUnknownInputReference(t *testing.T) {
	def := buildDef(t, `
schema_version: workflow.v1
workflow_id: test.bad-ref
title: Bad ref
description: Unknown input ref.
inputs:
  known:
    type: string
    required: true
steps:
  - id: s1
    type: noop
    args:
      message: ${inputs.unknown_input}
verification:
  - id: v1
    type: which_command
    args:
      command: echo
    expect:
      found: true
`)
	_, err := Build(context.Background(), def, BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{
			Ref:    def.WorkflowID,
			Inputs: map[string]string{"known": "val"},
		},
		Runtime:  bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry: registry.Default(),
	})
	if err == nil {
		t.Fatal("expected error for unknown input reference")
	}
}

func TestBuildRejectsReferenceToFutureStep(t *testing.T) {
	def := buildDef(t, `
schema_version: workflow.v1
workflow_id: test.future-ref
title: Future ref
description: References a future step.
steps:
  - id: s1
    type: noop
    args:
      message: ${steps.s2.message}
  - id: s2
    type: noop
    args: {}
verification:
  - id: v1
    type: which_command
    args:
      command: echo
    expect:
      found: true
`)
	_, err := Build(context.Background(), def, BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{Ref: def.WorkflowID},
		Runtime:        bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry:       registry.Default(),
	})
	if err == nil {
		t.Fatal("expected error for reference to future step")
	}
}

func TestBuildRejectsReferenceToUnknownStepOutput(t *testing.T) {
	def := buildDef(t, `
schema_version: workflow.v1
workflow_id: test.unknown-output
title: Unknown output
description: References unknown output.
steps:
  - id: s1
    type: noop
    args: {}
  - id: s2
    type: noop
    args:
      message: ${steps.s1.nonexistent_output}
verification:
  - id: v1
    type: which_command
    args:
      command: echo
    expect:
      found: true
`)
	_, err := Build(context.Background(), def, BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{Ref: def.WorkflowID},
		Runtime:        bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry:       registry.Default(),
	})
	if err == nil {
		t.Fatal("expected error for unknown step output reference")
	}
}

func TestBuildRejectsUnsupportedReference(t *testing.T) {
	def := buildDef(t, `
schema_version: workflow.v1
workflow_id: test.bad-ref-kind
title: Bad ref kind
description: Uses unsupported ref.
steps:
  - id: s1
    type: noop
    args:
      message: ${env.SOME_VAR}
verification:
  - id: v1
    type: which_command
    args:
      command: echo
    expect:
      found: true
`)
	_, err := Build(context.Background(), def, BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{Ref: def.WorkflowID},
		Runtime:        bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry:       registry.Default(),
	})
	if err == nil {
		t.Fatal("expected error for unsupported reference kind")
	}
}

func TestBuildRejectsInvalidStepRefFormat(t *testing.T) {
	def := buildDef(t, `
schema_version: workflow.v1
workflow_id: test.invalid-ref-format
title: Invalid ref format
description: Ref has wrong number of parts.
steps:
  - id: s1
    type: noop
    args:
      message: ${steps.onlytwoparts}
verification:
  - id: v1
    type: which_command
    args:
      command: echo
    expect:
      found: true
`)
	_, err := Build(context.Background(), def, BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{Ref: def.WorkflowID},
		Runtime:        bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry:       registry.Default(),
	})
	if err == nil {
		t.Fatal("expected error for invalid step ref format")
	}
}

func TestBuildRejectsUnknownStepType(t *testing.T) {
	def := buildDef(t, `
schema_version: workflow.v1
workflow_id: test.unknown-type
title: Unknown type
description: Uses unknown step type.
steps:
  - id: s1
    type: unknown_step_type_xyz
    args: {}
verification:
  - id: v1
    type: which_command
    args:
      command: echo
    expect:
      found: true
`)
	_, err := Build(context.Background(), def, BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{Ref: def.WorkflowID},
		Runtime:        bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry:       registry.Default(),
	})
	if err == nil {
		t.Fatal("expected error for unknown step type")
	}
}

func TestBuildWithArtifactSetsSourceFields(t *testing.T) {
	def := buildDef(t, `
schema_version: workflow.v1
workflow_id: test.with-artifact
title: With artifact
description: Sets source fields from artifact.
steps:
  - id: s1
    type: noop
    args: {}
verification:
  - id: v1
    type: which_command
    args:
      command: echo
    expect:
      found: true
`)
	artifact := &model.FailureArtifact{
		Fingerprint: "fp-abc",
		MatchedPlaybook: &model.ArtifactPlaybook{
			ID: "my-failure",
		},
	}
	compiled, err := Build(context.Background(), def, BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{Ref: def.WorkflowID},
		Artifact:       artifact,
		Runtime:        bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry:       registry.Default(),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if compiled.SourceFingerprint != "fp-abc" {
		t.Errorf("expected fp-abc, got %q", compiled.SourceFingerprint)
	}
	if compiled.SourceFailureID != "my-failure" {
		t.Errorf("expected my-failure, got %q", compiled.SourceFailureID)
	}
}

func TestBuildRequiredSafetyFromPlan(t *testing.T) {
	tempDir := t.TempDir()
	def := buildDef(t, `
schema_version: workflow.v1
workflow_id: test.safety
title: Safety
description: Has environment mutation.
inputs:
  pkg:
    type: string
    required: true
steps:
  - id: install
    type: install_package
    args:
      package: ${inputs.pkg}
      manager: apt-get
verification:
  - id: verify
    type: which_command
    args:
      command: ${inputs.pkg}
    expect:
      found: true
`)
	compiled, err := Build(context.Background(), def, BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{
			Ref:    def.WorkflowID,
			Inputs: map[string]string{"pkg": "curl"},
		},
		Runtime:  bind.RuntimeContext{WorkDir: tempDir},
		Registry: registry.Default(),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	found := false
	for _, sc := range compiled.RequiredSafety {
		if sc == registry.SafetyClassEnvironmentMutation {
			found = true
		}
	}
	if !found {
		t.Errorf("expected environment_mutation in required safety, got %v", compiled.RequiredSafety)
	}
}
