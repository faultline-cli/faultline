package bind

import (
	"testing"

	"faultline/internal/model"
	"faultline/internal/workflow/schema"
)

func TestResolveInputsAndReferences(t *testing.T) {
	def := schema.Definition{
		WorkflowID: "example",
		Inputs: map[string]schema.Input{
			"tool": {Type: "string", Required: true},
		},
	}
	inputs, err := ResolveInputs(def, model.ArtifactWorkflowRecommendation{
		Ref:    "example",
		Inputs: map[string]string{"tool": "node"},
	})
	if err != nil {
		t.Fatalf("ResolveInputs: %v", err)
	}
	state := State{
		Inputs: inputs,
		Runtime: RuntimeContext{
			WorkDir: "/workspace",
		},
		Artifact: &model.FailureArtifact{
			Fingerprint: "fp-1",
			Facts: map[string]string{
				"missing_executable": "node",
			},
		},
		Steps: map[string]map[string]string{
			"detect": {"manager": "apt-get"},
		},
	}
	value, err := ResolveValue(map[string]any{
		"command": "${inputs.tool}",
		"path":    "${artifact.facts.missing_executable}",
		"manager": "${steps.detect.manager}",
		"root":    "${runtime.workdir}",
	}, state, false)
	if err != nil {
		t.Fatalf("ResolveValue: %v", err)
	}
	resolved := value.(map[string]any)
	if resolved["command"] != "node" || resolved["manager"] != "apt-get" || resolved["root"] != "/workspace" {
		t.Fatalf("unexpected resolved values: %#v", resolved)
	}
}
