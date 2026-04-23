package plan

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"faultline/internal/model"
	"faultline/internal/workflow/bind"
	"faultline/internal/workflow/registry"
	"faultline/internal/workflow/schema"
)

func TestBuildDryRunProbesReadSteps(t *testing.T) {
	tempDir := t.TempDir()
	managerPath := filepath.Join(tempDir, "apt-get")
	if err := os.WriteFile(managerPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
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
	compiled, err := Build(context.Background(), def, BuildOptions{
		Recommendation: model.ArtifactWorkflowRecommendation{
			Ref:    def.WorkflowID,
			Inputs: map[string]string{"missing_executable": "node"},
		},
		Runtime:        bind.RuntimeContext{WorkDir: tempDir, RepoRoot: tempDir},
		Registry:       registry.Default(),
		ProbeReadSteps: true,
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if got := compiled.Steps[1].ResolvedArgs["manager"]; got != "apt-get" {
		t.Fatalf("expected probed manager, got %#v", got)
	}
}
