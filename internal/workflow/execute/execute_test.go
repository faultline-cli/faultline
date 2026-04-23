package execute

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"faultline/internal/model"
	"faultline/internal/workflow/bind"
	"faultline/internal/workflow/plan"
	"faultline/internal/workflow/registry"
	"faultline/internal/workflow/schema"
)

func TestApplyExecutesWorkflowAndVerifies(t *testing.T) {
	tempDir := t.TempDir()
	managerPath := filepath.Join(tempDir, "apt-get")
	managerScript := "#!/bin/sh\nset -eu\nif [ \"$1\" = \"install\" ]; then\n  target=\"" + filepath.Join(tempDir, "node") + "\"\n  printf '#!/bin/sh\\necho node installed\\n' > \"$target\"\n  /bin/chmod +x \"$target\"\nfi\n"
	if err := os.WriteFile(managerPath, []byte(managerScript), 0o755); err != nil {
		t.Fatalf("write fake manager: %v", err)
	}
	t.Setenv("PATH", tempDir)

	def, err := schema.LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: missing-executable.install
title: Install missing executable
description: Example.
inputs:
  missing_executable:
    type: string
    required: true
steps:
  - id: detect-manager
    type: detect_package_manager
    args: {}
  - id: install-tool
    type: install_package
    args:
      package: ${inputs.missing_executable}
      manager: ${steps.detect-manager.manager}
      check_command: ${inputs.missing_executable}
verification:
  - id: verify-tool
    type: which_command
    args:
      command: ${inputs.missing_executable}
    expect:
      found: true
`))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	compiled, err := plan.Build(context.Background(), def, plan.BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{
			Ref:    def.WorkflowID,
			Inputs: map[string]string{"missing_executable": "node"},
		},
		Runtime:  bind.RuntimeContext{WorkDir: tempDir, RepoRoot: tempDir},
		Registry: registry.Default(),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	record, err := Apply(context.Background(), compiled, Options{
		Runtime: bind.RuntimeContext{WorkDir: tempDir, RepoRoot: tempDir},
		Policy:  Policy{AllowEnvironmentMutation: true},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if record.Status != model.WorkflowExecutionStatusSucceeded || record.VerificationStatus != model.WorkflowVerificationStatusPassed {
		t.Fatalf("unexpected record status: %#v", record)
	}
	if len(record.StepResults) != 3 {
		t.Fatalf("expected three step results, got %#v", record.StepResults)
	}
}
