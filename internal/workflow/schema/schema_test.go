package schema

import "testing"

func TestLoadBytesRejectsUnknownFields(t *testing.T) {
	_, err := LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: example
title: Example
description: Example workflow.
steps:
  - id: probe
    type: noop
    args: {}
unknown_field: true
verification:
  - id: verify
    type: noop
    args: {}
`))
	if err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestLoadBytesRequiresVerification(t *testing.T) {
	_, err := LoadBytes([]byte(`
schema_version: workflow.v1
workflow_id: example
title: Example
description: Example workflow.
steps:
  - id: probe
    type: noop
    args: {}
`))
	if err == nil {
		t.Fatal("expected verification validation error")
	}
}
