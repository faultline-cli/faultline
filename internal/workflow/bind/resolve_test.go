package bind

import (
	"testing"

	"faultline/internal/model"
	"faultline/internal/workflow/schema"
)

// --- ResolveInputs ---

func TestResolveInputsUsesDefault(t *testing.T) {
	def := schema.Definition{
		WorkflowID: "example",
		Inputs: map[string]schema.Input{
			"tool": {Type: "string", Default: "node"},
		},
	}
	inputs, err := ResolveInputs(def, model.ArtifactWorkflowRecommendation{
		Ref:    "example",
		Inputs: map[string]string{},
	})
	if err != nil {
		t.Fatalf("ResolveInputs: %v", err)
	}
	if inputs["tool"] != "node" {
		t.Fatalf("expected default 'node', got %q", inputs["tool"])
	}
}

func TestResolveInputsRequiredMissingErrors(t *testing.T) {
	def := schema.Definition{
		WorkflowID: "example",
		Inputs: map[string]schema.Input{
			"tool": {Type: "string", Required: true},
		},
	}
	if _, err := ResolveInputs(def, model.ArtifactWorkflowRecommendation{
		Ref:    "example",
		Inputs: map[string]string{},
	}); err == nil {
		t.Fatal("expected error for missing required input")
	}
}

func TestResolveInputsOptionalEmptyOmitted(t *testing.T) {
	def := schema.Definition{
		WorkflowID: "example",
		Inputs: map[string]schema.Input{
			"opt": {Type: "string"},
		},
	}
	inputs, err := ResolveInputs(def, model.ArtifactWorkflowRecommendation{
		Ref:    "example",
		Inputs: map[string]string{},
	})
	if err != nil {
		t.Fatalf("ResolveInputs: %v", err)
	}
	if _, ok := inputs["opt"]; ok {
		t.Fatal("expected optional empty input to be omitted")
	}
}

func TestResolveInputsRecommendationOverridesDefault(t *testing.T) {
	def := schema.Definition{
		WorkflowID: "example",
		Inputs: map[string]schema.Input{
			"tool": {Type: "string", Default: "node"},
		},
	}
	inputs, err := ResolveInputs(def, model.ArtifactWorkflowRecommendation{
		Ref:    "example",
		Inputs: map[string]string{"tool": "go"},
	})
	if err != nil {
		t.Fatalf("ResolveInputs: %v", err)
	}
	if inputs["tool"] != "go" {
		t.Fatalf("expected recommendation to override default, got %q", inputs["tool"])
	}
}

// --- References ---

func TestReferencesFromString(t *testing.T) {
	refs := References("${inputs.tool} and ${runtime.workdir}")
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %v", refs)
	}
}

func TestReferencesFromSlice(t *testing.T) {
	refs := References([]any{"${inputs.a}", "${inputs.b}"})
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %v", refs)
	}
}

func TestReferencesFromMap(t *testing.T) {
	refs := References(map[string]any{"key": "${steps.detect.manager}"})
	if len(refs) != 1 || refs[0] != "steps.detect.manager" {
		t.Fatalf("expected 1 ref, got %v", refs)
	}
}

func TestReferencesDeduped(t *testing.T) {
	refs := References([]any{"${inputs.tool}", "${inputs.tool}"})
	if len(refs) != 1 {
		t.Fatalf("expected deduped refs, got %v", refs)
	}
}

func TestReferencesNonString(t *testing.T) {
	refs := References(42)
	if len(refs) != 0 {
		t.Fatalf("expected no refs for non-string, got %v", refs)
	}
}

// --- ResolveValue ---

func TestResolveValuePassthroughNonString(t *testing.T) {
	got, err := ResolveValue(42, State{}, false)
	if err != nil {
		t.Fatalf("ResolveValue: %v", err)
	}
	if got != 42 {
		t.Fatalf("expected passthrough, got %v", got)
	}
}

func TestResolveValueSlice(t *testing.T) {
	state := State{Inputs: map[string]string{"tool": "node"}}
	got, err := ResolveValue([]any{"${inputs.tool}", "literal"}, state, false)
	if err != nil {
		t.Fatalf("ResolveValue: %v", err)
	}
	slice := got.([]any)
	if len(slice) != 2 || slice[0] != "node" || slice[1] != "literal" {
		t.Fatalf("unexpected slice: %v", slice)
	}
}

func TestResolveValueMap(t *testing.T) {
	state := State{Inputs: map[string]string{"tool": "node"}}
	got, err := ResolveValue(map[string]any{"cmd": "${inputs.tool}"}, state, false)
	if err != nil {
		t.Fatalf("ResolveValue: %v", err)
	}
	m := got.(map[string]any)
	if m["cmd"] != "node" {
		t.Fatalf("unexpected map: %v", m)
	}
}

func TestResolveValueUnresolvedRefErrors(t *testing.T) {
	state := State{Inputs: map[string]string{}}
	if _, err := ResolveValue("${inputs.missing}", state, false); err == nil {
		t.Fatal("expected error for unresolved reference")
	}
}

func TestResolveValueAllowsDeferredStepRef(t *testing.T) {
	state := State{Steps: map[string]map[string]string{}}
	got, err := ResolveValue("${steps.detect.manager}", state, true)
	if err != nil {
		t.Fatalf("ResolveValue with deferred step ref: %v", err)
	}
	// deferred step ref remains unresolved as a placeholder
	if got != "${steps.detect.manager}" {
		t.Fatalf("expected placeholder kept, got %q", got)
	}
}

// --- resolveReference variants ---

func TestResolveReferenceRuntimeWorkdir(t *testing.T) {
	state := State{Runtime: RuntimeContext{WorkDir: "/work"}}
	val, ok, err := resolveReference("runtime.workdir", state)
	if err != nil {
		t.Fatalf("resolveReference: %v", err)
	}
	if !ok || val != "/work" {
		t.Fatalf("expected /work, got %q (ok=%v)", val, ok)
	}
}

func TestResolveReferenceRuntimeWorkdirEmpty(t *testing.T) {
	state := State{}
	_, ok, err := resolveReference("runtime.workdir", state)
	if err != nil {
		t.Fatalf("resolveReference: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for empty workdir")
	}
}

func TestResolveReferenceRuntimeRepoRoot(t *testing.T) {
	state := State{Runtime: RuntimeContext{RepoRoot: "/repo"}}
	val, ok, err := resolveReference("runtime.repo_root", state)
	if err != nil {
		t.Fatalf("resolveReference: %v", err)
	}
	if !ok || val != "/repo" {
		t.Fatalf("expected /repo, got %q (ok=%v)", val, ok)
	}
}

func TestResolveReferenceArtifactFingerprint(t *testing.T) {
	state := State{Artifact: &model.FailureArtifact{Fingerprint: "fp-abc"}}
	val, ok, err := resolveReference("artifact.fingerprint", state)
	if err != nil {
		t.Fatalf("resolveReference: %v", err)
	}
	if !ok || val != "fp-abc" {
		t.Fatalf("expected fp-abc, got %q (ok=%v)", val, ok)
	}
}

func TestResolveReferenceArtifactFingerprintNilArtifact(t *testing.T) {
	state := State{}
	_, ok, err := resolveReference("artifact.fingerprint", state)
	if err != nil {
		t.Fatalf("resolveReference: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for nil artifact fingerprint")
	}
}

func TestResolveReferenceArtifactMatchedPlaybookID(t *testing.T) {
	state := State{Artifact: &model.FailureArtifact{
		MatchedPlaybook: &model.ArtifactPlaybook{ID: "docker-auth"},
	}}
	val, ok, err := resolveReference("artifact.matched_playbook.id", state)
	if err != nil {
		t.Fatalf("resolveReference: %v", err)
	}
	if !ok || val != "docker-auth" {
		t.Fatalf("expected docker-auth, got %q (ok=%v)", val, ok)
	}
}

func TestResolveReferenceArtifactMatchedPlaybookIDNilArtifact(t *testing.T) {
	state := State{}
	_, ok, err := resolveReference("artifact.matched_playbook.id", state)
	if err != nil {
		t.Fatalf("resolveReference: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for nil artifact")
	}
}

func TestResolveReferenceArtifactMatchedPlaybookIDNilPlaybook(t *testing.T) {
	state := State{Artifact: &model.FailureArtifact{}}
	_, ok, err := resolveReference("artifact.matched_playbook.id", state)
	if err != nil {
		t.Fatalf("resolveReference: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for nil matched playbook")
	}
}

func TestResolveReferenceArtifactFacts(t *testing.T) {
	state := State{Artifact: &model.FailureArtifact{
		Facts: map[string]string{"cmd": "node"},
	}}
	val, ok, err := resolveReference("artifact.facts.cmd", state)
	if err != nil {
		t.Fatalf("resolveReference: %v", err)
	}
	if !ok || val != "node" {
		t.Fatalf("expected node, got %q (ok=%v)", val, ok)
	}
}

func TestResolveReferenceArtifactFactsNilArtifact(t *testing.T) {
	state := State{}
	_, ok, err := resolveReference("artifact.facts.cmd", state)
	if err != nil {
		t.Fatalf("resolveReference: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for nil artifact facts")
	}
}

func TestResolveReferenceStepInvalidFormat(t *testing.T) {
	state := State{}
	if _, _, err := resolveReference("steps.only_two", state); err == nil {
		t.Fatal("expected error for invalid step reference format")
	}
}

func TestResolveReferenceStepMissingStep(t *testing.T) {
	state := State{Steps: map[string]map[string]string{}}
	_, ok, err := resolveReference("steps.missing.output", state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for missing step")
	}
}

func TestResolveReferenceUnsupported(t *testing.T) {
	state := State{}
	if _, _, err := resolveReference("unknown.reference", state); err == nil {
		t.Fatal("expected error for unsupported reference")
	}
}
