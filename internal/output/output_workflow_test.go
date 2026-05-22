package output

import (
	"encoding/json"
	"strings"
	"testing"

	"faultline/internal/model"
	"faultline/internal/workflow"
)

// makePlan constructs a workflow.Plan with the provided failure ID and steps.
func makePlan(failureID string, steps []string) workflow.Plan {
	return workflow.Plan{
		SchemaVersion: "workflow.v1",
		Mode:          workflow.ModeLocal,
		FailureID:     failureID,
		Title:         failureID + " title",
		Steps:         steps,
		Evidence:      []string{},
	}
}

// ── FormatWorkflowText – minimal (no failure ID) ──────────────────────────────

func TestFormatWorkflowTextNoFailureID(t *testing.T) {
	plan := workflow.Plan{
		Steps: []string{"step A", "step B"},
	}
	out := FormatWorkflowText(plan)
	if !strings.HasPrefix(out, "WORKFLOW\n") {
		t.Errorf("expected 'WORKFLOW' header for plan without failure ID, got:\n%s", out)
	}
	if !strings.Contains(out, "1. step A") {
		t.Errorf("expected numbered step '1. step A', got:\n%s", out)
	}
	if !strings.Contains(out, "2. step B") {
		t.Errorf("expected numbered step '2. step B', got:\n%s", out)
	}
}

// ── FormatWorkflowText – full plan ────────────────────────────────────────────

func TestFormatWorkflowTextWithFailureID(t *testing.T) {
	plan := makePlan("docker-auth", []string{"run docker login", "retry the build"})
	out := FormatWorkflowText(plan)
	if !strings.Contains(out, "WORKFLOW") {
		t.Errorf("expected WORKFLOW header, got:\n%s", out)
	}
	if !strings.Contains(out, "docker-auth") {
		t.Errorf("expected failure ID in workflow text, got:\n%s", out)
	}
	if !strings.Contains(out, "1. run docker login") {
		t.Errorf("expected '1. run docker login', got:\n%s", out)
	}
}

func TestFormatWorkflowTextSource(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step"})
	plan.Source = "stdin"
	out := FormatWorkflowText(plan)
	if !strings.Contains(out, "Source: stdin") {
		t.Errorf("expected 'Source: stdin', got:\n%s", out)
	}
}

func TestFormatWorkflowTextContextStage(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step"})
	plan.Context = model.Context{Stage: "build", CommandHint: "npm ci", Step: "install"}
	out := FormatWorkflowText(plan)
	if !strings.Contains(out, "Stage: build") {
		t.Errorf("expected 'Stage: build', got:\n%s", out)
	}
	if !strings.Contains(out, "Command: npm ci") {
		t.Errorf("expected 'Command: npm ci', got:\n%s", out)
	}
	if !strings.Contains(out, "Step: install") {
		t.Errorf("expected 'Step: install', got:\n%s", out)
	}
}

func TestFormatWorkflowTextEvidence(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step"})
	plan.Evidence = []string{"authentication required", "pull access denied"}
	out := FormatWorkflowText(plan)
	if !strings.Contains(out, "Evidence:") {
		t.Errorf("expected 'Evidence:' section, got:\n%s", out)
	}
	if !strings.Contains(out, "  - authentication required") {
		t.Errorf("expected evidence item, got:\n%s", out)
	}
}

func TestFormatWorkflowTextFocusFiles(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step"})
	plan.Files = []string{"Dockerfile", ".dockerignore"}
	out := FormatWorkflowText(plan)
	if !strings.Contains(out, "Focus files:") {
		t.Errorf("expected 'Focus files:' section, got:\n%s", out)
	}
	if !strings.Contains(out, "  - Dockerfile") {
		t.Errorf("expected Dockerfile in focus files, got:\n%s", out)
	}
}

func TestFormatWorkflowTextLocalRepro(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step"})
	plan.LocalRepro = []string{"docker pull registry.example.com/image"}
	out := FormatWorkflowText(plan)
	if !strings.Contains(out, "Local repro:") {
		t.Errorf("expected 'Local repro:' section, got:\n%s", out)
	}
}

func TestFormatWorkflowTextVerify(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step"})
	plan.Verify = []string{"go test ./..."}
	out := FormatWorkflowText(plan)
	if !strings.Contains(out, "Verify:") {
		t.Errorf("expected 'Verify:' section, got:\n%s", out)
	}
	if !strings.Contains(out, "  - go test ./...") {
		t.Errorf("expected verify command, got:\n%s", out)
	}
}

func TestFormatWorkflowTextRankingHints(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step"})
	plan.RankingHints = []string{"hint A"}
	out := FormatWorkflowText(plan)
	if !strings.Contains(out, "Ranking hints:") {
		t.Errorf("expected 'Ranking hints:' section, got:\n%s", out)
	}
}

func TestFormatWorkflowTextDeltaHints(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step"})
	plan.DeltaHints = []string{"changed: go.sum"}
	out := FormatWorkflowText(plan)
	if !strings.Contains(out, "Delta hints:") {
		t.Errorf("expected 'Delta hints:' section, got:\n%s", out)
	}
}

func TestFormatWorkflowTextRemediation(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step"})
	plan.Remediation = &model.RemediationPlan{
		Commands: []model.RemediationCommand{
			{Phase: "fix", Command: []string{"docker", "login"}},
		},
		PatchSuggestions: []model.PatchSuggestion{
			{TargetFile: ".github/workflows/ci.yml", Summary: "add login step"},
		},
		CIConfigDiffs: []model.CIConfigDiff{
			{TargetFile: ".github/workflows/ci.yml", Summary: "configure registry"},
		},
	}
	out := FormatWorkflowText(plan)
	if !strings.Contains(out, "Remediation commands:") {
		t.Errorf("expected 'Remediation commands:' section, got:\n%s", out)
	}
	if !strings.Contains(out, "[fix] docker login") {
		t.Errorf("expected remediation command with phase, got:\n%s", out)
	}
	if !strings.Contains(out, "Patch suggestions:") {
		t.Errorf("expected 'Patch suggestions:' section, got:\n%s", out)
	}
	if !strings.Contains(out, "CI config diffs:") {
		t.Errorf("expected 'CI config diffs:' section, got:\n%s", out)
	}
}

func TestFormatWorkflowTextAgentPrompt(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step"})
	plan.AgentPrompt = "Check the registry configuration."
	out := FormatWorkflowText(plan)
	if !strings.Contains(out, "Agent prompt:") {
		t.Errorf("expected 'Agent prompt:' section, got:\n%s", out)
	}
	if !strings.Contains(out, "Check the registry configuration.") {
		t.Errorf("expected agent prompt text, got:\n%s", out)
	}
}

func TestFormatWorkflowTextNextSteps(t *testing.T) {
	plan := makePlan("docker-auth", []string{"do this", "then that"})
	out := FormatWorkflowText(plan)
	if !strings.Contains(out, "Next steps:") {
		t.Errorf("expected 'Next steps:' header, got:\n%s", out)
	}
	if !strings.Contains(out, "1. do this") {
		t.Errorf("expected numbered step, got:\n%s", out)
	}
	if !strings.Contains(out, "2. then that") {
		t.Errorf("expected second numbered step, got:\n%s", out)
	}
}

// ── FormatWorkflowJSON ────────────────────────────────────────────────────────

func TestFormatWorkflowJSONIsValidJSON(t *testing.T) {
	plan := makePlan("docker-auth", []string{"run docker login"})
	out, err := FormatWorkflowJSON(plan)
	if err != nil {
		t.Fatalf("FormatWorkflowJSON: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func TestFormatWorkflowJSONEndsWithNewline(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step"})
	out, err := FormatWorkflowJSON(plan)
	if err != nil {
		t.Fatalf("FormatWorkflowJSON: %v", err)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("expected trailing newline, got:\n%q", out)
	}
}

func TestFormatWorkflowJSONContainsSchemaVersion(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step"})
	plan.SchemaVersion = "workflow.v1"
	out, err := FormatWorkflowJSON(plan)
	if err != nil {
		t.Fatalf("FormatWorkflowJSON: %v", err)
	}
	if !strings.Contains(out, "workflow.v1") {
		t.Errorf("expected schema_version in JSON, got:\n%s", out)
	}
}

func TestFormatWorkflowJSONContainsSteps(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step 1", "step 2"})
	out, err := FormatWorkflowJSON(plan)
	if err != nil {
		t.Fatalf("FormatWorkflowJSON: %v", err)
	}
	if !strings.Contains(out, "step 1") {
		t.Errorf("expected step 1 in JSON, got:\n%s", out)
	}
}

func TestFormatWorkflowJSONWithRemediation(t *testing.T) {
	plan := makePlan("docker-auth", []string{"step"})
	plan.Remediation = &model.RemediationPlan{
		Commands: []model.RemediationCommand{
			{Phase: "fix", Command: []string{"docker", "login"}},
		},
	}
	out, err := FormatWorkflowJSON(plan)
	if err != nil {
		t.Fatalf("FormatWorkflowJSON: %v", err)
	}
	if !strings.Contains(out, `"remediation"`) {
		t.Errorf("expected remediation in JSON, got:\n%s", out)
	}
}
