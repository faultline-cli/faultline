package execute

import (
	"context"
	"testing"
	"time"

	"faultline/internal/model"
	"faultline/internal/workflow/bind"
	"faultline/internal/workflow/plan"
	"faultline/internal/workflow/registry"
	"faultline/internal/workflow/schema"
)

// --- Policy.Allows ---

func TestPolicyAllowsRead(t *testing.T) {
	p := Policy{}
	if !p.Allows(registry.SafetyClassRead) {
		t.Fatal("expected read to be allowed by default policy")
	}
}

func TestPolicyAllowsLocalMutation(t *testing.T) {
	p0 := Policy{}
	if p0.Allows(registry.SafetyClassLocalMutation) {
		t.Fatal("expected local mutation blocked by default")
	}
	p1 := Policy{AllowLocalMutation: true}
	if !p1.Allows(registry.SafetyClassLocalMutation) {
		t.Fatal("expected local mutation allowed when enabled")
	}
}

func TestPolicyAllowsEnvironmentMutation(t *testing.T) {
	p0 := Policy{}
	if p0.Allows(registry.SafetyClassEnvironmentMutation) {
		t.Fatal("expected env mutation blocked by default")
	}
	p1 := Policy{AllowEnvironmentMutation: true}
	if !p1.Allows(registry.SafetyClassEnvironmentMutation) {
		t.Fatal("expected env mutation allowed when enabled")
	}
}

func TestPolicyAllowsExternalSideEffect(t *testing.T) {
	p0 := Policy{}
	if p0.Allows(registry.SafetyClassExternalSideEffect) {
		t.Fatal("expected external side effect blocked by default")
	}
	p1 := Policy{AllowExternalSideEffect: true}
	if !p1.Allows(registry.SafetyClassExternalSideEffect) {
		t.Fatal("expected external side effect allowed when enabled")
	}
}

func TestPolicyAllowsUnknownClass(t *testing.T) {
	p := Policy{}
	if p.Allows(registry.SafetyClass("unknown_class")) {
		t.Fatal("expected unknown safety class to be blocked")
	}
}

// --- Apply with noop workflow (deterministic, no mutations) ---

func buildNoopPlan(t *testing.T) plan.Plan {
	t.Helper()
	def, err := schema.LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: test.noop
title: Noop workflow
description: Test-only workflow.
steps:
  - id: step1
    type: noop
    args:
      message: hello
    expect: {}
verification:
  - id: verify1
    type: noop
    args: {}
`))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	compiled, err := plan.Build(context.Background(), def, plan.BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{Ref: def.WorkflowID},
		Runtime:        bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry:       registry.Default(),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return compiled
}

func TestApplySucceedsWithNoopWorkflow(t *testing.T) {
	compiled := buildNoopPlan(t)
	record, err := Apply(context.Background(), compiled, Options{
		Runtime: bind.RuntimeContext{WorkDir: t.TempDir()},
		Policy:  Policy{},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if record.Status != model.WorkflowExecutionStatusSucceeded {
		t.Fatalf("expected succeeded, got %v", record.Status)
	}
	if record.VerificationStatus != model.WorkflowVerificationStatusPassed {
		t.Fatalf("expected passed verification, got %v", record.VerificationStatus)
	}
}

func TestApplyUsesProvidedNow(t *testing.T) {
	compiled := buildNoopPlan(t)
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	record, err := Apply(context.Background(), compiled, Options{
		Runtime: bind.RuntimeContext{WorkDir: t.TempDir()},
		Policy:  Policy{},
		Now:     func() time.Time { return fixed },
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if record.StartedAt != "2026-01-01T00:00:00Z" {
		t.Fatalf("expected fixed start time, got %q", record.StartedAt)
	}
}

func TestApplyBlockedByPolicyReturnsError(t *testing.T) {
	def, err := schema.LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: test.install
title: Install workflow
description: Requires env mutation.
steps:
  - id: detect
    type: detect_package_manager
    args: {}
verification:
  - id: verify
    type: noop
    args: {}
`))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	compiled, err := plan.Build(context.Background(), def, plan.BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{Ref: def.WorkflowID},
		Runtime:        bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry:       registry.Default(),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	// detect_package_manager is read-class, so make the step look like env mutation
	// by using install_package instead
	def2, err := schema.LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: test.install2
title: Install workflow
description: Requires env mutation.
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
`))
	if err != nil {
		t.Fatalf("LoadBytes2: %v", err)
	}
	compiled2, err := plan.Build(context.Background(), def2, plan.BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{
			Ref:    def2.WorkflowID,
			Inputs: map[string]string{"pkg": "curl"},
		},
		Runtime:  bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry: registry.Default(),
	})
	if err != nil {
		t.Fatalf("Build2: %v", err)
	}
	record, err := Apply(context.Background(), compiled2, Options{
		Runtime: bind.RuntimeContext{WorkDir: t.TempDir()},
		Policy:  Policy{AllowEnvironmentMutation: false},
	})
	if err == nil {
		t.Fatal("expected error when policy blocks env mutation")
	}
	if record.Status != model.WorkflowExecutionStatusBlocked {
		t.Fatalf("expected blocked status, got %v", record.Status)
	}
	_ = compiled
}

func TestApplyFailStepReturnsError(t *testing.T) {
	def, err := schema.LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: test.fail
title: Fail workflow
description: Intentionally fails.
steps:
  - id: step1
    type: fail
    args:
      message: intentional test failure
verification:
  - id: verify
    type: noop
    args: {}
`))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	compiled, err := plan.Build(context.Background(), def, plan.BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{Ref: def.WorkflowID},
		Runtime:        bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry:       registry.Default(),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	record, err := Apply(context.Background(), compiled, Options{
		Runtime: bind.RuntimeContext{WorkDir: t.TempDir()},
		Policy:  Policy{},
	})
	if err == nil {
		t.Fatal("expected error from fail step")
	}
	if record.Status != model.WorkflowExecutionStatusFailed {
		t.Fatalf("expected failed status, got %v", record.Status)
	}
}

func TestApplyVerificationFailureReturnsError(t *testing.T) {
	def, err := schema.LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: test.verifyfail
title: Verify fail workflow
description: Verification that always fails.
steps:
  - id: step1
    type: noop
    args: {}
verification:
  - id: verify
    type: which_command
    args:
      command: surely_not_a_command_xyz_123
    expect:
      found: true
`))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	compiled, err := plan.Build(context.Background(), def, plan.BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{Ref: def.WorkflowID},
		Runtime:        bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry:       registry.Default(),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	record, err := Apply(context.Background(), compiled, Options{
		Runtime: bind.RuntimeContext{WorkDir: t.TempDir()},
		Policy:  Policy{},
	})
	if err == nil {
		t.Fatal("expected error when verification fails")
	}
	if record.VerificationStatus != model.WorkflowVerificationStatusFailed {
		t.Fatalf("expected failed verification, got %v", record.VerificationStatus)
	}
}

func TestApplySetsWorkflowMetadata(t *testing.T) {
	compiled := buildNoopPlan(t)
	record, err := Apply(context.Background(), compiled, Options{
		Runtime: bind.RuntimeContext{WorkDir: t.TempDir()},
		Policy:  Policy{},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if record.WorkflowID != "test.noop" {
		t.Fatalf("expected workflow id test.noop, got %q", record.WorkflowID)
	}
	if record.Mode != model.WorkflowExecutionModeApply {
		t.Fatalf("expected apply mode, got %v", record.Mode)
	}
	if record.SchemaVersion == "" {
		t.Fatal("expected non-empty schema version")
	}
}

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

func TestMapFromMap(t *testing.T) {
	m := map[string]any{"key": "val"}
	got := mapFrom(m)
	if got == nil || got["key"] != "val" {
		t.Fatalf("unexpected result: %v", got)
	}
}

// ── Apply error paths ─────────────────────────────────────────────────────────

// TestApplyResolveArgsErrorReturnsFailed tests that when a step's args contain
// an unresolvable reference, Apply returns a failed record.
func TestApplyResolveArgsErrorReturnsFailed(t *testing.T) {
	// Build a plan with a step whose ResolvedArgs contain an unresolvable reference.
	def, err := schema.LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: test.resolve-err
title: Resolve error test
description: Triggers a resolve error.
steps:
  - id: step1
    type: noop
    args:
      message: hello
    expect: {}
verification:
  - id: verify1
    type: noop
    args: {}
`))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	compiled, err := plan.Build(context.Background(), def, plan.BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{Ref: def.WorkflowID},
		Runtime:        bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry:       registry.Default(),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	// Inject an unresolvable step reference into the compiled args.
	for i := range compiled.Steps {
		compiled.Steps[i].ResolvedArgs = map[string]any{
			"message": "${inputs.does_not_exist}",
		}
	}
	record, err := Apply(context.Background(), compiled, Options{
		Runtime: bind.RuntimeContext{WorkDir: t.TempDir()},
		Policy:  Policy{},
	})
	if err == nil {
		t.Fatal("expected error when ResolveValue fails for args")
	}
	if record.Status != model.WorkflowExecutionStatusFailed {
		t.Fatalf("expected failed status, got %v", record.Status)
	}
}

// TestApplyResolveExpectErrorReturnsFailed tests that when a step's expect contains
// an unresolvable reference, Apply returns a failed record.
func TestApplyResolveExpectErrorReturnsFailed(t *testing.T) {
	def, err := schema.LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: test.expect-resolve-err
title: Expect resolve error test
description: Triggers a resolve error in expect.
steps:
  - id: step1
    type: noop
    args: {}
verification:
  - id: verify1
    type: noop
    args: {}
`))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	compiled, err := plan.Build(context.Background(), def, plan.BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{Ref: def.WorkflowID},
		Runtime:        bind.RuntimeContext{WorkDir: t.TempDir()},
		Registry:       registry.Default(),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	// Inject an unresolvable reference into the expect map of the first step.
	for i := range compiled.Steps {
		compiled.Steps[i].ResolvedExpect = map[string]any{
			"message": "${inputs.does_not_exist}",
		}
	}
	record, err := Apply(context.Background(), compiled, Options{
		Runtime: bind.RuntimeContext{WorkDir: t.TempDir()},
		Policy:  Policy{},
	})
	if err == nil {
		t.Fatal("expected error when ResolveValue fails for expect")
	}
	if record.Status != model.WorkflowExecutionStatusFailed {
		t.Fatalf("expected failed status, got %v", record.Status)
	}
}
