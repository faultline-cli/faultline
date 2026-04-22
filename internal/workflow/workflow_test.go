package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/model"
)

func TestBuildLocalPlanUsesTopDiagnosis(t *testing.T) {
	a := &model.Analysis{
		Source: "build.log",
		Context: model.Context{
			Stage:       "build",
			CommandHint: "npm run build",
			Step:        "Compile assets",
		},
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:           "typescript-compile",
					Title:        "TypeScript compile or type-check failure",
					Fix:          "1. Run `tsc --noEmit` locally.\n2. Update the affected type definitions.",
					WhyItMatters: "## Prevention\n\n- Run a dedicated type-check step on every pull request.",
					Workflow: model.WorkflowSpec{
						LikelyFiles: []string{"tsconfig.json", "web/**/*.ts"},
						LocalRepro:  []string{"npm run build"},
						Verify:      []string{"npm run build", "tsc --noEmit"},
					},
				},
				Evidence: []string{"error TS2322: Type 'string' is not assignable to type 'number'."},
			},
		},
		RepoContext: &model.RepoContext{
			RecentFiles: []string{"web/src/app.ts", "tsconfig.json"},
		},
	}

	plan := Build(a, ModeLocal)
	if plan.FailureID != "typescript-compile" {
		t.Fatalf("expected top failure id, got %q", plan.FailureID)
	}
	if len(plan.Steps) < 4 {
		t.Fatalf("expected workflow steps, got %v", plan.Steps)
	}
	if len(plan.Verify) != 2 {
		t.Fatalf("expected verify commands, got %v", plan.Verify)
	}
	if plan.AgentPrompt != "" {
		t.Fatalf("local mode should not include agent prompt, got %q", plan.AgentPrompt)
	}
}

func TestBuildAgentPlanIncludesPrompt(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "docker-build-context",
					Title: "Docker build context or Dockerfile path issue",
					Fix:   "1. Verify the exact `docker build` command.",
					Workflow: model.WorkflowSpec{
						LocalRepro: []string{"docker build -f Dockerfile ."},
						Verify:     []string{"docker build -f Dockerfile ."},
					},
				},
				Evidence: []string{"failed to read Dockerfile"},
			},
		},
	}

	plan := Build(a, ModeAgent)
	if !strings.Contains(plan.AgentPrompt, "docker-build-context") {
		t.Fatalf("expected failure id in agent prompt, got %q", plan.AgentPrompt)
	}
	if !strings.Contains(plan.AgentPrompt, "Evidence lines:") {
		t.Fatalf("expected evidence section in agent prompt, got %q", plan.AgentPrompt)
	}
	if !strings.Contains(plan.AgentPrompt, "Verification commands:") {
		t.Fatalf("expected verification section in agent prompt, got %q", plan.AgentPrompt)
	}
}

func TestBuildWithOptionsResolvesLikelyFilesFromRepo(t *testing.T) {
	repoDir := t.TempDir()
	mustWriteFile(t, filepath.Join(repoDir, "Dockerfile"), "FROM alpine\n")
	mustWriteFile(t, filepath.Join(repoDir, ".dockerignore"), "dist/\n")
	mustWriteFile(t, filepath.Join(repoDir, ".github", "workflows", "build.yml"), "name: build\n")

	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "docker-build-context",
					Title: "Docker build context or Dockerfile path issue",
					Workflow: model.WorkflowSpec{
						LikelyFiles: []string{"Dockerfile", ".dockerignore", ".github/workflows/*.yml"},
					},
				},
			},
		},
	}

	plan := BuildWithOptions(a, ModeLocal, BuildOptions{RepoPath: repoDir})
	if len(plan.Files) != 3 {
		t.Fatalf("expected resolved files, got %v", plan.Files)
	}
	if plan.Files[0] != "Dockerfile" {
		t.Fatalf("expected Dockerfile first, got %v", plan.Files)
	}
}

func TestBuildIncludesMetricsAndPolicyHints(t *testing.T) {
	tss := 0.42
	fpc := 0.55
	phi := 0.38
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "docker-auth",
					Title: "Docker registry authentication failure",
					Fix:   "1. Re-authenticate before retrying.",
				},
				Evidence: []string{"authentication required"},
			},
		},
		Metrics: &model.Metrics{
			TSS:             &tss,
			FPC:             &fpc,
			PHI:             &phi,
			HistoryCount:    6,
			DriftComponents: []string{"auth failures dominating recent runs"},
		},
		Policy: &model.Policy{
			Recommendation: "quarantine",
			Reason:         "the same failure pattern is recurring",
			Basis:          []string{"tss", "phi"},
		},
	}

	plan := Build(a, ModeLocal)
	if len(plan.MetricsHints) < 4 {
		t.Fatalf("expected metrics hints, got %v", plan.MetricsHints)
	}
	if !strings.Contains(strings.Join(plan.MetricsHints, " | "), "TSS 0.42 (6 runs)") {
		t.Fatalf("expected TSS hint, got %v", plan.MetricsHints)
	}
	if !strings.Contains(strings.Join(plan.PolicyHints, " | "), "policy: quarantine") {
		t.Fatalf("expected policy recommendation hint, got %v", plan.PolicyHints)
	}
	if strings.Contains(strings.Join(plan.Steps, " | "), "Recent change drift points at") {
		t.Fatalf("did not expect delta hint step when delta is absent, got %v", plan.Steps)
	}
}

func TestMetricsHintsNilReturnsNil(t *testing.T) {
	if got := metricsHints(nil); got != nil {
		t.Fatalf("expected nil metrics hints, got %v", got)
	}
}

func TestPolicyHintsNilReturnsNil(t *testing.T) {
	if got := policyHints(nil); got != nil {
		t.Fatalf("expected nil policy hints, got %v", got)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
