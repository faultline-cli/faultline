package schema

import (
	"os"
	"path/filepath"
	"testing"
)

// --- LoadFile ---

func TestLoadFileReadsValidDefinition(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.yaml")
	content := []byte(`
schema_version: workflow.v1
workflow_id: example.load
title: Load test
description: Tests LoadFile.
steps:
  - id: step1
    type: noop
    args: {}
verification:
  - id: verify1
    type: noop
    args: {}
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	def, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if def.WorkflowID != "example.load" {
		t.Fatalf("unexpected workflow id: %q", def.WorkflowID)
	}
}

func TestLoadFileMissingFileErrors(t *testing.T) {
	if _, err := LoadFile("/no/such/file/workflow.yaml"); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadFileInvalidYAMLErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte(":\tinvalid:\tyaml:"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := LoadFile(path); err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

// --- validateDefinition ---

func TestLoadBytesRequiresSchemaVersion(t *testing.T) {
	_, err := LoadBytes([]byte(`
workflow_id: example
title: Example
description: Example.
steps:
  - id: step1
    type: noop
    args: {}
`))
	if err == nil {
		t.Fatal("expected error for missing schema_version")
	}
}

func TestLoadBytesRequiresWorkflowID(t *testing.T) {
	_, err := LoadBytes([]byte(`
schema_version: workflow.v1
title: Example
description: Example.
steps:
  - id: step1
    type: noop
    args: {}
`))
	if err == nil {
		t.Fatal("expected error for missing workflow_id")
	}
}

func TestLoadBytesRequiresTitle(t *testing.T) {
	_, err := LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: example
description: Example.
steps:
  - id: step1
    type: noop
    args: {}
`))
	if err == nil {
		t.Fatal("expected error for missing title")
	}
}

func TestLoadBytesRequiresDescription(t *testing.T) {
	_, err := LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: example
title: Example
steps:
  - id: step1
    type: noop
    args: {}
`))
	if err == nil {
		t.Fatal("expected error for missing description")
	}
}

func TestLoadBytesRequiresAtLeastOneStep(t *testing.T) {
	_, err := LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: example
title: Example
description: Example.
steps: []
verification:
  - id: verify1
    type: noop
    args: {}
`))
	if err == nil {
		t.Fatal("expected error for empty steps")
	}
}

func TestLoadBytesRejectsDuplicateStepID(t *testing.T) {
	_, err := LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: example
title: Example
description: Example.
steps:
  - id: step1
    type: noop
    args: {}
verification:
  - id: step1
    type: noop
    args: {}
`))
	if err == nil {
		t.Fatal("expected error for duplicate step id")
	}
}

func TestLoadBytesRejectsStepWithoutID(t *testing.T) {
	_, err := LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: example
title: Example
description: Example.
steps:
  - type: noop
    args: {}
verification:
  - id: verify1
    type: noop
    args: {}
`))
	if err == nil {
		t.Fatal("expected error for step without id")
	}
}

func TestLoadBytesRejectsStepWithoutType(t *testing.T) {
	_, err := LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: example
title: Example
description: Example.
steps:
  - id: step1
    args: {}
verification:
  - id: verify1
    type: noop
    args: {}
`))
	if err == nil {
		t.Fatal("expected error for step without type")
	}
}

func TestLoadBytesRejectsUnsupportedInputType(t *testing.T) {
	_, err := LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: example
title: Example
description: Example.
inputs:
  tool:
    type: integer
steps:
  - id: step1
    type: noop
    args: {}
verification:
  - id: verify1
    type: noop
    args: {}
`))
	if err == nil {
		t.Fatal("expected error for unsupported input type")
	}
}

func TestLoadBytesStepExpectationCountsAsVerification(t *testing.T) {
	// A workflow with no verification section but a step with expect should be valid
	_, err := LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: example
title: Example
description: Example.
steps:
  - id: step1
    type: which_command
    args:
      command: sh
    expect:
      found: true
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadBytesValidDefinitionAccepted(t *testing.T) {
	def, err := LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: example.valid
title: Valid workflow
description: A valid workflow definition.
inputs:
  tool:
    type: string
    required: true
    description: The tool to install
steps:
  - id: step1
    type: noop
    args:
      message: start
verification:
  - id: verify1
    type: noop
    args: {}
`))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if def.WorkflowID != "example.valid" {
		t.Fatalf("unexpected workflow id: %q", def.WorkflowID)
	}
	if len(def.Inputs) != 1 {
		t.Fatalf("expected 1 input, got %d", len(def.Inputs))
	}
}

// --- normalizeValue ---

func TestNormalizeValueString(t *testing.T) {
	out := normalizeValue("  hello  ")
	if out != "hello" {
		t.Fatalf("expected trimmed string, got %v", out)
	}
}

func TestNormalizeValueSlice(t *testing.T) {
	out := normalizeValue([]any{"  a  ", "  b  "})
	slice := out.([]any)
	if len(slice) != 2 || slice[0] != "a" || slice[1] != "b" {
		t.Fatalf("unexpected slice normalization: %v", out)
	}
}

func TestNormalizeValueMapStringAny(t *testing.T) {
	out := normalizeValue(map[string]any{"  key  ": "  val  "})
	m := out.(map[string]any)
	if m["key"] != "val" {
		t.Fatalf("unexpected map normalization: %v", out)
	}
}

func TestNormalizeValueMapAnyAny(t *testing.T) {
	out := normalizeValue(map[any]any{"  key  ": "  val  "})
	m := out.(map[string]any)
	if m["key"] != "val" {
		t.Fatalf("unexpected map[any]any normalization: %v", out)
	}
}

func TestNormalizeValueNonString(t *testing.T) {
	out := normalizeValue(42)
	if out != 42 {
		t.Fatalf("expected passthrough for non-string, got %v", out)
	}
}

// --- normalizeMap ---

func TestNormalizeMapNilReturnsNil(t *testing.T) {
	if got := normalizeMap(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestNormalizeMapEmptyKeySkipped(t *testing.T) {
	out := normalizeMap(map[string]any{"  ": "value"})
	if len(out) != 0 {
		t.Fatalf("expected empty-key entry to be skipped, got %v", out)
	}
}

func TestNormalizeMapAllEmptyKeysReturnsNil(t *testing.T) {
	out := normalizeMap(map[string]any{"  ": "x", "   ": "y"})
	if out != nil {
		t.Fatalf("expected nil when all keys are empty, got %v", out)
	}
}

// --- normalizeDefinition ---

func TestNormalizeDefinitionHandlesNil(t *testing.T) {
	// should not panic
	normalizeDefinition(nil)
}

func TestNormalizeDefinitionClearsEmptyPolicySlices(t *testing.T) {
	def := &Definition{
		SchemaVersion: "workflow.v1",
		WorkflowID:    "example",
		Title:         "Example",
		Description:   "desc",
		Policy: PolicyHints{
			Requires: []string{},
			Notes:    []string{},
		},
		Steps: []Step{{ID: "s1", Type: "noop"}},
	}
	normalizeDefinition(def)
	if def.Policy.Requires != nil {
		t.Fatalf("expected nil Requires after normalization, got %v", def.Policy.Requires)
	}
	if def.Policy.Notes != nil {
		t.Fatalf("expected nil Notes after normalization, got %v", def.Policy.Notes)
	}
}

func TestLoadBytesPreservesMetadata(t *testing.T) {
	def, err := LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: example.meta
title: Meta test
description: Tests metadata field.
metadata:
  author: test
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
	if def.Metadata["author"] != "test" {
		t.Fatalf("expected metadata author=test, got %v", def.Metadata)
	}
}

func TestLoadBytesPreservesPolicyNotes(t *testing.T) {
	def, err := LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: example.policy
title: Policy test
description: Tests policy notes.
policy:
  notes:
    - Requires sudo access
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
	if len(def.Policy.Notes) != 1 || def.Policy.Notes[0] != "Requires sudo access" {
		t.Fatalf("unexpected policy notes: %v", def.Policy.Notes)
	}
}
