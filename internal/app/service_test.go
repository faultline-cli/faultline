package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"faultline/internal/engine"
	"faultline/internal/output"
	"faultline/internal/workflow"
)

// repoPlaybookDir returns the canonical bundled playbook directory relative to
// the module root. Tests in the app package run from the package directory, so
// we walk up two levels.
func repoPlaybookDir() string {
	return "../../playbooks/bundled"
}

// baseOpts returns a minimal AnalyzeOptions that avoids file-system side
// effects and overrides the playbook directory.
func baseOpts() AnalyzeOptions {
	return AnalyzeOptions{
		Top:         1,
		Mode:        output.ModeQuick,
		Format:      output.FormatTerminal,
		NoHistory:   true,
		PlaybookDir: repoPlaybookDir(),
	}
}

// ── Analyze ──────────────────────────────────────────────────────────────────

func TestAnalyzeMatchedTextOutput(t *testing.T) {
	svc := NewService()
	log := "pull access denied\nError response from daemon: authentication required\n"
	var buf bytes.Buffer

	err := svc.Analyze(strings.NewReader(log), "test.log", baseOpts(), &buf)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if !strings.Contains(buf.String(), "Docker registry authentication failure") {
		t.Errorf("expected docker-auth result in output, got %q", buf.String())
	}
}

func TestAnalyzeMatchedJSONOutput(t *testing.T) {
	svc := NewService()
	log := "pull access denied\nError response from daemon: authentication required\n"
	opts := baseOpts()
	opts.JSON = true
	var buf bytes.Buffer

	err := svc.Analyze(strings.NewReader(log), "stdin", opts, &buf)
	if err != nil {
		t.Fatalf("Analyze JSON: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &payload); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if payload["matched"] != true {
		t.Errorf("expected matched=true, got %v", payload["matched"])
	}
}

func TestAnalyzeNoMatchReturnsNoError(t *testing.T) {
	svc := NewService()
	log := "everything is fine, all checks passed\n"
	var buf bytes.Buffer

	// ErrNoMatch is swallowed - output should still be written without error.
	err := svc.Analyze(strings.NewReader(log), "", baseOpts(), &buf)
	if err != nil {
		t.Fatalf("expected no error on no-match, got %v", err)
	}
}

func TestAnalyzeEmptyInputReturnsErrNoInput(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer

	err := svc.Analyze(strings.NewReader(""), "", baseOpts(), &buf)
	if !errors.Is(err, engine.ErrNoInput) {
		t.Fatalf("expected ErrNoInput, got %v", err)
	}
}

// ── Fix ──────────────────────────────────────────────────────────────────────

func TestFixOutputContainsFixSteps(t *testing.T) {
	svc := NewService()
	log := "fatal: could not read Username for 'https://github.com': terminal prompts disabled\n"
	var buf bytes.Buffer

	err := svc.Fix(strings.NewReader(log), "", baseOpts(), &buf)
	if err != nil {
		t.Fatalf("Fix: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty fix output")
	}
}

func TestFixMarkdownOutput(t *testing.T) {
	svc := NewService()
	log := "fatal: could not read Username for 'https://github.com': terminal prompts disabled\n"
	opts := baseOpts()
	opts.Format = output.FormatMarkdown
	var buf bytes.Buffer

	err := svc.Fix(strings.NewReader(log), "", opts, &buf)
	if err != nil {
		t.Fatalf("Fix markdown: %v", err)
	}
	if !strings.HasPrefix(buf.String(), "#") {
		t.Errorf("expected markdown heading, got %q", buf.String()[:min(60, buf.Len())])
	}
}

// ── List ─────────────────────────────────────────────────────────────────────

func TestListWritesPlaybooks(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer

	err := svc.List("", repoPlaybookDir(), nil, &buf)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected list output to be non-empty")
	}
}

func TestListFiltersByCategory(t *testing.T) {
	svc := NewService()
	var all, filtered bytes.Buffer

	if err := svc.List("", repoPlaybookDir(), nil, &all); err != nil {
		t.Fatalf("List (all): %v", err)
	}
	if err := svc.List("auth", repoPlaybookDir(), nil, &filtered); err != nil {
		t.Fatalf("List (auth): %v", err)
	}
	if filtered.Len() >= all.Len() {
		t.Error("filtered list should be smaller than the full list")
	}
}

// ── Explain ──────────────────────────────────────────────────────────────────

func TestExplainKnownPlaybook(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer

	err := svc.Explain("git-auth", repoPlaybookDir(), nil, output.FormatTerminal, &buf)
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	if !strings.Contains(buf.String(), "git-auth") {
		t.Errorf("expected playbook ID in output, got %q", buf.String()[:min(80, buf.Len())])
	}
}

func TestExplainUnknownPlaybookReturnsError(t *testing.T) {
	svc := NewService()
	var buf bytes.Buffer

	err := svc.Explain("does-not-exist-abc123", repoPlaybookDir(), nil, output.FormatTerminal, &buf)
	if err == nil {
		t.Error("expected error for unknown playbook ID, got nil")
	}
}

// ── Workflow ──────────────────────────────────────────────────────────────────

func TestWorkflowLocalMode(t *testing.T) {
	svc := NewService()
	log := "pull access denied\nError response from daemon: authentication required\n"
	var buf bytes.Buffer

	err := svc.Workflow(strings.NewReader(log), "", baseOpts(), workflow.ModeLocal, false, &buf)
	if err != nil {
		t.Fatalf("Workflow: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty workflow output")
	}
}

func TestWorkflowJSONOutput(t *testing.T) {
	svc := NewService()
	log := "pull access denied\nError response from daemon: authentication required\n"
	var buf bytes.Buffer

	err := svc.Workflow(strings.NewReader(log), "", baseOpts(), workflow.ModeLocal, true, &buf)
	if err != nil {
		t.Fatalf("Workflow JSON: %v", err)
	}
	var payload map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &payload); jsonErr != nil {
		t.Fatalf("unmarshal workflow JSON: %v", jsonErr)
	}
}

// min is needed for Go versions before 1.21 built-in min.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
