package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestJSONSchemaContractSmoke(t *testing.T) {
	playbookDir, err := filepath.Abs("../playbooks/bundled")
	if err != nil {
		t.Fatalf("abs playbook dir: %v", err)
	}
	t.Setenv("FAULTLINE_PLAYBOOK_DIR", playbookDir)
	repoRoot := filepath.Clean("..")

	analyze := runRootCommand(t, repoRoot, "analyze", "--json", "--no-history", "--git=false", "examples/missing-executable.log")
	var analysis struct {
		Results []struct {
			FailureID string `json:"failure_id"`
		} `json:"results"`
		Artifact map[string]any `json:"artifact"`
	}
	if err := json.Unmarshal([]byte(analyze), &analysis); err != nil {
		t.Fatalf("analysis json must remain parseable: %v", err)
	}
	if len(analysis.Results) == 0 || analysis.Results[0].FailureID != "missing-executable" {
		t.Fatalf("unexpected analysis top result: %#v", analysis.Results)
	}
	if len(analysis.Artifact) == 0 {
		t.Fatalf("analysis json must include first-class artifact")
	}

	workflowJSON := runRootCommand(t, repoRoot, "workflow", "--json", "--mode", "agent", "--no-history", "--git=false", "examples/missing-executable.log")
	var legacyWorkflow struct {
		SchemaVersion string   `json:"schema_version"`
		FailureID     string   `json:"failure_id"`
		Steps         []string `json:"steps"`
	}
	if err := json.Unmarshal([]byte(workflowJSON), &legacyWorkflow); err != nil {
		t.Fatalf("legacy workflow json must remain parseable: %v", err)
	}
	if legacyWorkflow.SchemaVersion != "workflow.v1" || legacyWorkflow.FailureID != "missing-executable" || len(legacyWorkflow.Steps) == 0 {
		t.Fatalf("unexpected legacy workflow contract: %#v", legacyWorkflow)
	}

	structured := runRootCommand(t, repoRoot, "workflow", "explain", "--json", "--no-history", "--git=false", "examples/missing-executable.log")
	var structuredWorkflow struct {
		SchemaVersion   string `json:"schema_version"`
		WorkflowID      string `json:"workflow_id"`
		SourceFailureID string `json:"source_failure_id"`
	}
	if err := json.Unmarshal([]byte(structured), &structuredWorkflow); err != nil {
		t.Fatalf("structured workflow json must remain parseable: %v", err)
	}
	if structuredWorkflow.WorkflowID == "" || structuredWorkflow.SourceFailureID != "missing-executable" {
		t.Fatalf("unexpected structured workflow contract: %#v", structuredWorkflow)
	}

	batch := runRootCommand(t, repoRoot, "batch", "--json", "--no-history", "examples/missing-executable.log", "examples/runtime-mismatch.log")
	var batchPayload struct {
		SchemaVersion string `json:"schema_version"`
		Total         int    `json:"total"`
		Matched       int    `json:"matched"`
	}
	if err := json.Unmarshal([]byte(batch), &batchPayload); err != nil {
		t.Fatalf("batch json must remain parseable: %v", err)
	}
	if batchPayload.SchemaVersion != "batch.v1" || batchPayload.Total != 2 || batchPayload.Matched != 2 {
		t.Fatalf("unexpected batch contract: %#v", batchPayload)
	}
}
