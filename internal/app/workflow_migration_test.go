package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"faultline/internal/workflow"
)

func TestWorkflowLegacyAndStructuredPathsUseSameTopDiagnosis(t *testing.T) {
	logData, err := os.ReadFile(filepath.Join(repoRoot(t), "examples", "missing-executable.log"))
	if err != nil {
		t.Fatalf("read example log: %v", err)
	}
	opts := AnalyzeOptions{
		OutputOptions: OutputOptions{Top: 1},
		ProviderOptions: ProviderOptions{
			GitContextEnabled: false,
			RepoPath:          repoRoot(t),
		},
		PlaybookDir: repoPlaybookDir(),
		Store:       "off",
	}

	var legacyOut bytes.Buffer
	if err := NewService().Workflow(bytes.NewReader(logData), "examples/missing-executable.log", opts, workflow.ModeAgent, true, &legacyOut); err != nil {
		t.Fatalf("legacy workflow: %v", err)
	}
	var legacy struct {
		FailureID string `json:"failure_id"`
	}
	if err := json.Unmarshal(legacyOut.Bytes(), &legacy); err != nil {
		t.Fatalf("unmarshal legacy workflow: %v", err)
	}

	var structuredOut bytes.Buffer
	if err := NewService().WorkflowExplain(bytes.NewReader(logData), "examples/missing-executable.log", opts, "", true, &structuredOut); err != nil {
		t.Fatalf("structured workflow explain: %v", err)
	}
	var structured struct {
		SourceFailureID string `json:"source_failure_id"`
	}
	if err := json.Unmarshal(structuredOut.Bytes(), &structured); err != nil {
		t.Fatalf("unmarshal structured workflow: %v", err)
	}
	if legacy.FailureID != structured.SourceFailureID {
		t.Fatalf("workflow migration drift: legacy failure %q structured failure %q", legacy.FailureID, structured.SourceFailureID)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}
