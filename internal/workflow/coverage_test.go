package workflow

import (
	"strings"
	"testing"

	"faultline/internal/model"
)

// --- BuildWithOptions edge cases ---

func TestBuildWithOptionsNilAnalysisReturnsPrompt(t *testing.T) {
	plan := BuildWithOptions(nil, ModeLocal, BuildOptions{})
	if plan.SchemaVersion != schemaVersion {
		t.Errorf("expected schema version %q, got %q", schemaVersion, plan.SchemaVersion)
	}
	if len(plan.Steps) == 0 {
		t.Error("expected prompt step for nil analysis")
	}
	if !strings.Contains(plan.Steps[0], "faultline analyze") {
		t.Errorf("expected faultline analyze in step, got %q", plan.Steps[0])
	}
}

func TestBuildWithOptionsNoResultsReturnsHelpSteps(t *testing.T) {
	a := &model.Analysis{
		Source:  "build.log",
		Results: []model.Result{},
	}
	plan := BuildWithOptions(a, ModeLocal, BuildOptions{})
	if len(plan.Steps) < 2 {
		t.Errorf("expected multiple steps for no-match analysis, got %v", plan.Steps)
	}
	found := false
	for _, step := range plan.Steps {
		if strings.Contains(step, "faultline list") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'faultline list' guidance in steps, got %v", plan.Steps)
	}
}

func TestBuildWithOptionsDeltaHintsTruncatedToThree(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "npm-ci-lockfile",
					Title: "npm lockfile out of sync",
					Fix:   "1. Run `npm install` and commit the updated lock file.",
				},
			},
		},
		Delta: &model.Delta{
			Causes: []model.DeltaCause{
				{Kind: "dependency", Reasons: []string{"reason1", "reason2", "reason3", "reason4"}},
			},
		},
	}
	plan := BuildWithOptions(a, ModeLocal, BuildOptions{})
	if len(plan.DeltaHints) > 3 {
		t.Errorf("expected at most 3 delta hints, got %d: %v", len(plan.DeltaHints), plan.DeltaHints)
	}
}

func TestBuildWithOptionsDeltaHintsDeduped(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "npm-ci-lockfile",
					Title: "npm lockfile out of sync",
					Fix:   "1. Run `npm install`.",
				},
			},
		},
		Delta: &model.Delta{
			Causes: []model.DeltaCause{
				{Kind: "dependency", Reasons: []string{"package.json changed", "package.json changed"}},
			},
		},
	}
	plan := BuildWithOptions(a, ModeLocal, BuildOptions{})
	for i, hint := range plan.DeltaHints {
		for j, other := range plan.DeltaHints {
			if i != j && hint == other {
				t.Errorf("duplicate delta hint %q at indices %d and %d", hint, i, j)
			}
		}
	}
}

func TestBuildWithOptionsRankingHints(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "docker-auth",
					Title: "Docker auth failure",
					Fix:   "1. Run `docker login`.",
				},
				Ranking: &model.Ranking{
					StrongestPositive: []string{"docker login evidence matched"},
				},
			},
		},
	}
	plan := BuildWithOptions(a, ModeLocal, BuildOptions{})
	if len(plan.RankingHints) == 0 {
		t.Error("expected ranking hints to be included in plan")
	}
	found := false
	for _, step := range plan.Steps {
		if strings.Contains(step, "Ranking evidence") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected ranking evidence step, got %v", plan.Steps)
	}
}

func TestBuildWithOptionsCommandHintStep(t *testing.T) {
	a := &model.Analysis{
		Context: model.Context{CommandHint: "docker push ghcr.io/app"},
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "docker-auth",
					Title: "Docker auth failure",
					Fix:   "1. Run `docker login`.",
				},
			},
		},
	}
	plan := BuildWithOptions(a, ModeLocal, BuildOptions{})
	found := false
	for _, step := range plan.Steps {
		if strings.Contains(step, "docker push ghcr.io/app") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected command hint step with command, got %v", plan.Steps)
	}
}

func TestBuildWithOptionsCIStepIncluded(t *testing.T) {
	a := &model.Analysis{
		Context: model.Context{Step: "Build and push image"},
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "docker-auth",
					Title: "Docker auth failure",
					Fix:   "1. Run `docker login`.",
				},
			},
		},
	}
	plan := BuildWithOptions(a, ModeLocal, BuildOptions{})
	found := false
	for _, step := range plan.Steps {
		if strings.Contains(step, "Build and push image") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected CI step in plan steps, got %v", plan.Steps)
	}
}

func TestBuildWithOptionsLocalReproSteps(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "npm-ci-lockfile",
					Title: "npm lockfile",
					Fix:   "1. Run `npm install`.",
					Workflow: model.WorkflowSpec{
						LocalRepro: []string{"npm ci --dry-run"},
						Verify:     []string{"npm test"},
					},
				},
			},
		},
	}
	plan := BuildWithOptions(a, ModeLocal, BuildOptions{})
	if len(plan.LocalRepro) == 0 {
		t.Error("expected local repro commands in plan")
	}
	if len(plan.Verify) == 0 {
		t.Error("expected verify commands in plan")
	}
}

func TestBuildWithOptionsWhyItMattersStep(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "docker-auth",
					Title: "Docker auth",
					Fix:   "1. Run `docker login`.",
					WhyItMatters: "## Prevention\n\n- Store Docker credentials as a CI secret.",
				},
			},
		},
	}
	plan := BuildWithOptions(a, ModeLocal, BuildOptions{})
	found := false
	for _, step := range plan.Steps {
		if strings.Contains(step, "Store Docker credentials") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected WhyItMatters in last step, got %v", plan.Steps)
	}
}

// --- buildAgentPrompt ---

func TestBuildAgentPromptIncludesSource(t *testing.T) {
	a := &model.Analysis{
		Source: "build.log",
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "docker-auth",
					Title: "Docker auth failure",
					Fix:   "1. Run `docker login`.",
				},
			},
		},
	}
	plan := BuildWithOptions(a, ModeAgent, BuildOptions{})
	if !strings.Contains(plan.AgentPrompt, "build.log") {
		t.Errorf("expected source file in agent prompt, got:\n%s", plan.AgentPrompt)
	}
}

func TestBuildAgentPromptIncludesStageAndStep(t *testing.T) {
	a := &model.Analysis{
		Context: model.Context{Stage: "deploy", Step: "Push image"},
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "docker-auth",
					Title: "Docker auth failure",
					Fix:   "1. Run `docker login`.",
				},
			},
		},
	}
	plan := BuildWithOptions(a, ModeAgent, BuildOptions{})
	if !strings.Contains(plan.AgentPrompt, "deploy") {
		t.Errorf("expected stage in agent prompt, got:\n%s", plan.AgentPrompt)
	}
	if !strings.Contains(plan.AgentPrompt, "Push image") {
		t.Errorf("expected CI step in agent prompt, got:\n%s", plan.AgentPrompt)
	}
}

func TestBuildAgentPromptIncludesFiles(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "docker-auth",
					Title: "Docker auth failure",
					Fix:   "1. Run `docker login`.",
					Workflow: model.WorkflowSpec{
						LikelyFiles: []string{"Dockerfile", ".dockerignore"},
					},
				},
			},
		},
	}
	plan := BuildWithOptions(a, ModeAgent, BuildOptions{})
	if !strings.Contains(plan.AgentPrompt, "Likely files to inspect:") {
		t.Errorf("expected files section in agent prompt, got:\n%s", plan.AgentPrompt)
	}
	if !strings.Contains(plan.AgentPrompt, "Dockerfile") {
		t.Errorf("expected Dockerfile in agent prompt files, got:\n%s", plan.AgentPrompt)
	}
}

func TestBuildAgentPromptIncludesLocalRepro(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "npm-ci-lockfile",
					Title: "npm lockfile",
					Fix:   "1. Run `npm install`.",
					Workflow: model.WorkflowSpec{
						LocalRepro: []string{"npm ci --dry-run"},
					},
				},
			},
		},
	}
	plan := BuildWithOptions(a, ModeAgent, BuildOptions{})
	if !strings.Contains(plan.AgentPrompt, "Local repro commands:") {
		t.Errorf("expected local repro section in agent prompt, got:\n%s", plan.AgentPrompt)
	}
}

func TestBuildAgentPromptIncludesDeltaHints(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "npm-ci-lockfile",
					Title: "npm lockfile",
					Fix:   "1. Run `npm install`.",
				},
			},
		},
		Delta: &model.Delta{
			Causes: []model.DeltaCause{
				{Kind: "dependency", Reasons: []string{"package.json changed"}},
			},
		},
	}
	plan := BuildWithOptions(a, ModeAgent, BuildOptions{})
	if !strings.Contains(plan.AgentPrompt, "Delta hints:") {
		t.Errorf("expected delta hints section in agent prompt, got:\n%s", plan.AgentPrompt)
	}
}

// --- repoRoot ---

func TestRepoRootFromOptions(t *testing.T) {
	a := &model.Analysis{}
	root := repoRoot(a, BuildOptions{RepoPath: "/custom/repo"})
	if root != "/custom/repo" {
		t.Errorf("expected /custom/repo, got %q", root)
	}
}

func TestRepoRootFromAnalysisContext(t *testing.T) {
	a := &model.Analysis{
		RepoContext: &model.RepoContext{RepoRoot: "/from/context"},
	}
	root := repoRoot(a, BuildOptions{})
	if root != "/from/context" {
		t.Errorf("expected /from/context, got %q", root)
	}
}

func TestRepoRootDotPath(t *testing.T) {
	root := repoRoot(nil, BuildOptions{RepoPath: "."})
	if root != "." {
		t.Errorf("expected '.', got %q", root)
	}
}

func TestRepoRootEmptyWhenNothingSet(t *testing.T) {
	root := repoRoot(nil, BuildOptions{})
	if root != "" {
		t.Errorf("expected empty string, got %q", root)
	}
}

func TestRepoRootOptionsOverridesContext(t *testing.T) {
	a := &model.Analysis{
		RepoContext: &model.RepoContext{RepoRoot: "/from/context"},
	}
	root := repoRoot(a, BuildOptions{RepoPath: "/override"})
	if root != "/override" {
		t.Errorf("expected /override, got %q", root)
	}
}

// --- resolveFiles ---

func TestResolveFilesUsesRepoContextRecentFiles(t *testing.T) {
	a := &model.Analysis{
		RepoContext: &model.RepoContext{
			RecentFiles: []string{"main.go", "config.go"},
		},
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID: "git-auth",
					Workflow: model.WorkflowSpec{
						LikelyFiles: []string{},
					},
				},
			},
		},
	}
	top := a.Results[0]
	files := resolveFiles(a, top, BuildOptions{})
	if len(files) < 2 {
		t.Errorf("expected recent files included in resolved files, got %v", files)
	}
}

func TestResolveFilesLimitedToMaxFiles(t *testing.T) {
	recentFiles := make([]string, 20)
	for i := range recentFiles {
		recentFiles[i] = "file" + strings.Repeat("0", i) + ".go"
	}
	a := &model.Analysis{
		RepoContext: &model.RepoContext{RecentFiles: recentFiles},
	}
	top := model.Result{Playbook: model.Playbook{ID: "git-auth"}}
	files := resolveFiles(a, top, BuildOptions{})
	if len(files) > maxFiles {
		t.Errorf("expected at most %d files, got %d: %v", maxFiles, len(files), files)
	}
}

// --- matchesPattern ---

func TestMatchesPatternExact(t *testing.T) {
	if !matchesPattern("Dockerfile", "Dockerfile") {
		t.Error("expected exact filename match")
	}
}

func TestMatchesPatternGlob(t *testing.T) {
	if !matchesPattern(".github/workflows/ci.yml", "*.yml") {
		t.Error("expected glob match by basename")
	}
}

func TestMatchesPatternContains(t *testing.T) {
	if !matchesPattern("src/docker/Dockerfile.prod", "Dockerfile") {
		t.Error("expected contains match")
	}
}

func TestMatchesPatternNoMatch(t *testing.T) {
	if matchesPattern("main.go", "Dockerfile") {
		t.Error("expected no match between main.go and Dockerfile")
	}
}

func TestMatchesPatternDoubleStarNormalized(t *testing.T) {
	if !matchesPattern(".github/workflows/ci.yml", "**/*.yml") {
		t.Error("expected ** normalized to match")
	}
}

// --- markdownListItems ---

func TestMarkdownListItemsDashPrefix(t *testing.T) {
	md := "- Run `docker login`\n- Check credentials"
	items := markdownListItems(md)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %v", items)
	}
	if items[0] != "Run `docker login`" {
		t.Errorf("unexpected item: %q", items[0])
	}
}

func TestMarkdownListItemsNumberedList(t *testing.T) {
	md := "1. First step\n2. Second step\n3. Third step"
	items := markdownListItems(md)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %v", items)
	}
	if items[0] != "First step" {
		t.Errorf("unexpected item: %q", items[0])
	}
}

func TestMarkdownListItemsIgnoresHeaders(t *testing.T) {
	md := "## Fix steps\n\n1. Run the command.\n\n## Validation\n\n- Check the result."
	items := markdownListItems(md)
	if len(items) != 2 {
		t.Fatalf("expected 2 items (1 numbered + 1 dash), got %v", items)
	}
}

// --- dedupeKeepOrder ---

func TestDedupeKeepOrderBasic(t *testing.T) {
	out := dedupeKeepOrder([]string{"a"}, []string{"b", "a", "c", ""})
	if len(out) != 3 {
		t.Errorf("expected 3 unique items, got %v", out)
	}
	if out[0] != "a" || out[1] != "b" || out[2] != "c" {
		t.Errorf("unexpected order: %v", out)
	}
}

func TestDedupeKeepOrderNilBase(t *testing.T) {
	out := dedupeKeepOrder(nil, []string{"x", "y", "x"})
	if len(out) != 2 {
		t.Errorf("expected 2 unique items, got %v", out)
	}
}
