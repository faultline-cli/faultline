package workflow

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"faultline/internal/model"
)

// Mode selects the output shape for workflow generation.
type Mode string

const (
	ModeLocal Mode = "local"
	ModeAgent Mode = "agent"

	schemaVersion = "workflow.v1"
	maxFiles      = 8
)

// BuildOptions configures repo-aware workflow planning.
type BuildOptions struct {
	RepoPath string
}

// Plan is a deterministic next-step plan derived from the top diagnosis.
type Plan struct {
	SchemaVersion string        `json:"schema_version"`
	Mode          Mode          `json:"mode"`
	FailureID     string        `json:"failure_id,omitempty"`
	Title         string        `json:"title,omitempty"`
	Source        string        `json:"source,omitempty"`
	Context       model.Context `json:"context"`
	Evidence      []string      `json:"evidence"`
	Files         []string      `json:"files,omitempty"`
	LocalRepro    []string      `json:"local_repro,omitempty"`
	Verify        []string      `json:"verify,omitempty"`
	Steps         []string      `json:"steps"`
	AgentPrompt   string        `json:"agent_prompt,omitempty"`
}

// Build returns a workflow plan for the top-ranked diagnosis.
func Build(a *model.Analysis, mode Mode) Plan {
	return BuildWithOptions(a, mode, BuildOptions{})
}

// BuildWithOptions returns a repo-aware workflow plan for the top-ranked diagnosis.
func BuildWithOptions(a *model.Analysis, mode Mode, opts BuildOptions) Plan {
	plan := Plan{
		SchemaVersion: schemaVersion,
		Mode:          mode,
		Context:       model.Context{},
		Evidence:      []string{},
		Files:         []string{},
		LocalRepro:    []string{},
		Verify:        []string{},
		Steps:         []string{},
	}
	if a == nil {
		plan.Steps = append(plan.Steps, "Run `faultline analyze <logfile>` first to produce a diagnosis.")
		return plan
	}

	plan.Source = a.Source
	plan.Context = a.Context
	if len(a.Results) == 0 {
		plan.Steps = append(plan.Steps,
			"No deterministic playbook matched this log yet.",
			"Run `faultline list` to inspect existing coverage and add a new playbook for the observed failure signature.",
			"Capture a short failing log excerpt and reduce it to stable evidence lines before extending the library.",
		)
		return plan
	}

	top := a.Results[0]
	plan.FailureID = top.Playbook.ID
	plan.Title = top.Playbook.Title
	plan.Evidence = append(plan.Evidence, top.Evidence...)
	plan.LocalRepro = append(plan.LocalRepro, top.Playbook.Workflow.LocalRepro...)
	plan.Verify = append(plan.Verify, top.Playbook.Workflow.Verify...)
	plan.Files = resolveFiles(a, top, opts)
	plan.Steps = append(plan.Steps, baseSteps(a, top, plan)...)
	if mode == ModeAgent {
		plan.AgentPrompt = buildAgentPrompt(a, top, plan)
	}
	return plan
}

func baseSteps(a *model.Analysis, top model.Result, plan Plan) []string {
	steps := []string{
		fmt.Sprintf("Confirm the top diagnosis `%s` by reproducing the failure from the same command or CI step if possible.", top.Playbook.ID),
	}

	if len(top.Evidence) > 0 {
		steps = append(steps,
			fmt.Sprintf("Use the matched evidence lines as the starting point for triage: %s.", strings.Join(top.Evidence, " | ")),
		)
	}

	if a.Context.CommandHint != "" {
		steps = append(steps,
			fmt.Sprintf("Re-run or inspect the failing command `%s` locally before editing code.", a.Context.CommandHint),
		)
	}

	if a.Context.Step != "" {
		steps = append(steps,
			fmt.Sprintf("Check the CI step named `%s` for missing setup, ordering, or environment assumptions.", a.Context.Step),
		)
	}

	if len(plan.Files) > 0 {
		steps = append(steps,
			fmt.Sprintf("Inspect the most relevant local files first: %s.", strings.Join(plan.Files, ", ")),
		)
	}

	for _, cmd := range plan.LocalRepro {
		steps = append(steps, fmt.Sprintf("Local repro: `%s`.", cmd))
	}

	for _, step := range markdownListItems(top.Playbook.Fix) {
		steps = append(steps, step)
	}

	for _, cmd := range plan.Verify {
		steps = append(steps, fmt.Sprintf("Verify with `%s` after the fix.", cmd))
	}

	if suggestions := markdownListItems(top.Playbook.WhyItMatters); len(suggestions) > 0 {
		steps = append(steps,
			fmt.Sprintf("After the immediate fix, harden the workflow with: %s.", trimTerminalPunctuation(suggestions[0])),
		)
	}

	return steps
}

func buildAgentPrompt(a *model.Analysis, top model.Result, plan Plan) string {
	var lines []string
	lines = append(lines,
		"You are helping resolve a deterministic CI failure in the local repository.",
		fmt.Sprintf("Workflow schema: %s.", plan.SchemaVersion),
		fmt.Sprintf("Top diagnosis: %s - %s.", top.Playbook.ID, top.Playbook.Title),
	)

	if a.Source != "" {
		lines = append(lines, fmt.Sprintf("Log source: %s.", a.Source))
	}
	if a.Context.Stage != "" {
		lines = append(lines, fmt.Sprintf("Likely stage: %s.", a.Context.Stage))
	}
	if a.Context.CommandHint != "" {
		lines = append(lines, fmt.Sprintf("Command hint: %s.", a.Context.CommandHint))
	}
	if a.Context.Step != "" {
		lines = append(lines, fmt.Sprintf("CI step: %s.", a.Context.Step))
	}
	if len(top.Evidence) > 0 {
		lines = append(lines, "Evidence lines:")
		for _, evidence := range top.Evidence {
			lines = append(lines, fmt.Sprintf("- %s", evidence))
		}
	}
	if len(plan.Files) > 0 {
		lines = append(lines, "Likely files to inspect:")
		for _, file := range plan.Files {
			lines = append(lines, fmt.Sprintf("- %s", file))
		}
	}
	if len(plan.LocalRepro) > 0 {
		lines = append(lines, "Local repro commands:")
		for _, cmd := range plan.LocalRepro {
			lines = append(lines, fmt.Sprintf("- %s", cmd))
		}
	}
	lines = append(lines, "Recommended fix steps:")
	for _, step := range markdownListItems(top.Playbook.Fix) {
		lines = append(lines, fmt.Sprintf("- %s", step))
	}
	if len(plan.Verify) > 0 {
		lines = append(lines, "Verification commands:")
		for _, cmd := range plan.Verify {
			lines = append(lines, fmt.Sprintf("- %s", cmd))
		}
	}
	lines = append(lines,
		"Work deterministically: inspect the failing path, make the smallest complete change, and verify the exact failure is resolved.",
	)
	return strings.Join(lines, "\n")
}

func resolveFiles(a *model.Analysis, top model.Result, opts BuildOptions) []string {
	files := dedupeKeepOrder(nil, top.Playbook.Workflow.LikelyFiles)
	if root := repoRoot(a, opts); root != "" {
		if resolved := resolveLikelyFiles(root, top.Playbook.Workflow.LikelyFiles); len(resolved) > 0 {
			files = resolved
		}
	}
	if a != nil && a.RepoContext != nil {
		files = dedupeKeepOrder(files, a.RepoContext.RecentFiles)
	}
	if len(files) > maxFiles {
		files = files[:maxFiles]
	}
	return files
}

func markdownListItems(markdown string) []string {
	lines := strings.Split(markdown, "\n")
	items := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "- "):
			items = append(items, strings.TrimSpace(strings.TrimPrefix(line, "- ")))
		case len(line) > 3 && line[1] == '.' && line[2] == ' ' && line[0] >= '0' && line[0] <= '9':
			items = append(items, strings.TrimSpace(line[3:]))
		}
	}
	return items
}

func repoRoot(a *model.Analysis, opts BuildOptions) string {
	if strings.TrimSpace(opts.RepoPath) != "" && opts.RepoPath != "." {
		return opts.RepoPath
	}
	if a != nil && a.RepoContext != nil && strings.TrimSpace(a.RepoContext.RepoRoot) != "" {
		return a.RepoContext.RepoRoot
	}
	if strings.TrimSpace(opts.RepoPath) == "." {
		return "."
	}
	return ""
}

func resolveLikelyFiles(root string, patterns []string) []string {
	if len(patterns) == 0 {
		return nil
	}
	var repoFiles []string
	_ = filepath.WalkDir(root, func(fullPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "bin", "dist":
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, fullPath)
		if err != nil {
			return nil
		}
		repoFiles = append(repoFiles, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(repoFiles)

	var matched []string
	for _, pattern := range patterns {
		pattern = filepath.ToSlash(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		for _, file := range repoFiles {
			if matchesPattern(file, pattern) {
				matched = append(matched, file)
			}
		}
	}
	return dedupeKeepOrder(nil, matched)
}

func matchesPattern(file, pattern string) bool {
	pattern = normalizePattern(pattern)
	if ok, _ := path.Match(pattern, file); ok {
		return true
	}
	if !strings.Contains(pattern, "/") {
		if ok, _ := path.Match(pattern, path.Base(file)); ok {
			return true
		}
	}
	return strings.Contains(file, pattern)
}

func normalizePattern(pattern string) string {
	pattern = strings.ReplaceAll(pattern, "**/", "")
	pattern = strings.ReplaceAll(pattern, "**", "*")
	return pattern
}

func dedupeKeepOrder(base []string, values []string) []string {
	out := append([]string{}, base...)
	seen := make(map[string]struct{}, len(out))
	for _, item := range out {
		seen[item] = struct{}{}
	}
	for _, item := range values {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func trimTerminalPunctuation(s string) string {
	return strings.TrimRight(strings.TrimSpace(s), ".!?")
}
