package renderer

import (
	"strings"
	"testing"

	"faultline/internal/model"
)

// ── extractFixCodeBlocks ──────────────────────────────────────────────────────

func TestExtractFixCodeBlocksEmpty(t *testing.T) {
	if blocks := extractFixCodeBlocks(""); len(blocks) != 0 {
		t.Fatalf("expected no blocks from empty input, got %v", blocks)
	}
}

func TestExtractFixCodeBlocksNoFences(t *testing.T) {
	fix := "1. Run go mod tidy\n2. Commit go.sum"
	if blocks := extractFixCodeBlocks(fix); len(blocks) != 0 {
		t.Fatalf("expected no blocks when no fences, got %v", blocks)
	}
}

func TestExtractFixCodeBlocksSingle(t *testing.T) {
	fix := "Some intro.\n\n```\ngo mod tidy\n```\n\nDone."
	blocks := extractFixCodeBlocks(fix)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d: %v", len(blocks), blocks)
	}
	if blocks[0] != "go mod tidy" {
		t.Errorf("unexpected block content: %q", blocks[0])
	}
}

func TestExtractFixCodeBlocksMultiple(t *testing.T) {
	fix := "Step 1:\n\n```sh\ngit fetch\n```\n\nStep 2:\n\n```sh\ngit pull\n```"
	blocks := extractFixCodeBlocks(fix)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d: %v", len(blocks), blocks)
	}
	if blocks[0] != "git fetch" {
		t.Errorf("block[0] = %q, want %q", blocks[0], "git fetch")
	}
	if blocks[1] != "git pull" {
		t.Errorf("block[1] = %q, want %q", blocks[1], "git pull")
	}
}

func TestExtractFixCodeBlocksMultiLineBlock(t *testing.T) {
	fix := "```bash\nexport FOO=1\ngo build ./...\n```"
	blocks := extractFixCodeBlocks(fix)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if !strings.Contains(blocks[0], "export FOO=1") || !strings.Contains(blocks[0], "go build") {
		t.Errorf("unexpected multi-line block content: %q", blocks[0])
	}
}

func TestExtractFixCodeBlocksEmptyFencedBlock(t *testing.T) {
	// A fenced block with no content should not produce an entry.
	fix := "```\n```"
	blocks := extractFixCodeBlocks(fix)
	if len(blocks) != 0 {
		t.Fatalf("expected 0 blocks for empty fence, got %d: %v", len(blocks), blocks)
	}
}

func TestExtractFixCodeBlocksUnclosedBlock(t *testing.T) {
	fix := "```\nonly this\n"
	// Unclosed block yields no output (never closed).
	blocks := extractFixCodeBlocks(fix)
	if len(blocks) != 0 {
		t.Fatalf("expected 0 blocks for unclosed fence, got %d: %v", len(blocks), blocks)
	}
}

// ── extractFixSection ─────────────────────────────────────────────────────────

func TestExtractFixSectionEmpty(t *testing.T) {
	if got := extractFixSection("", "Preconditions"); got != "" {
		t.Fatalf("expected empty for empty fix, got %q", got)
	}
}

func TestExtractFixSectionNotFound(t *testing.T) {
	fix := "## Fix Steps\n\nDo something."
	if got := extractFixSection(fix, "Preconditions"); got != "" {
		t.Fatalf("expected empty when section not present, got %q", got)
	}
}

func TestExtractFixSectionFound(t *testing.T) {
	fix := "## Fix Steps\n\n1. Run tidy.\n\n## Preconditions\n\nYou need admin access.\n\n## Risks\n\nData loss."
	got := extractFixSection(fix, "Preconditions")
	if !strings.Contains(got, "You need admin access.") {
		t.Errorf("expected section content, got %q", got)
	}
	// Should NOT include content from the next section.
	if strings.Contains(got, "Data loss") {
		t.Errorf("expected stop at next heading, got %q", got)
	}
}

func TestExtractFixSectionCaseInsensitive(t *testing.T) {
	fix := "## RISKS\n\nData loss possible.\n"
	got := extractFixSection(fix, "risks")
	if !strings.Contains(got, "Data loss possible.") {
		t.Errorf("expected case-insensitive match, got %q", got)
	}
}

func TestExtractFixSectionStopsAtSameLevelHeading(t *testing.T) {
	fix := "## Steps\n\nRun tidy.\n\n## Preconditions\n\nNeed go installed.\n\n## Validation\n\nRun tests."
	got := extractFixSection(fix, "Preconditions")
	if strings.Contains(got, "Run tests") {
		t.Errorf("expected stop at ## Validation, but got content from it: %q", got)
	}
}

func TestExtractFixSectionDeepHeading(t *testing.T) {
	// Section at level ### should stop at the next ### or higher.
	fix := "### Preconditions\n\nNeed sudo.\n\n### NextSection\n\nOther content."
	got := extractFixSection(fix, "Preconditions")
	if !strings.Contains(got, "Need sudo.") {
		t.Errorf("expected section content, got %q", got)
	}
	if strings.Contains(got, "Other content") {
		t.Errorf("expected stop at next ### heading, got %q", got)
	}
}

// ── renderFixCommandsOnly ─────────────────────────────────────────────────────

func TestRenderFixCommandsOnlyNoCommands(t *testing.T) {
	r := New(Options{Plain: true, Width: 88})
	a := &model.Analysis{Results: []model.Result{{
		Playbook: model.Playbook{
			ID:  "go-sum-missing",
			Fix: "Run `go mod tidy` and commit.",
		},
		Confidence: 0.8,
	}}}
	// Build with FixCommandsOnly opt
	r2 := New(Options{Plain: true, Width: 88, FixCommandsOnly: true})
	out := r2.RenderFix(a)
	if !strings.Contains(out, "No runnable commands found") {
		t.Errorf("expected no-commands message when fix has no fenced blocks, got %q", out)
	}
	_ = r
}

func TestRenderFixCommandsOnlyWithCommands(t *testing.T) {
	r := New(Options{Plain: true, Width: 88, FixCommandsOnly: true})
	a := &model.Analysis{Results: []model.Result{{
		Playbook: model.Playbook{
			ID:  "go-sum-missing",
			Fix: "## Fix\n\n```sh\ngo mod tidy\n```\n\n```sh\ngit add go.sum\n```",
		},
		Confidence: 0.82,
	}}}
	out := r.RenderFix(a)
	if !strings.Contains(out, "go mod tidy") {
		t.Errorf("expected first command in output, got %q", out)
	}
	if !strings.Contains(out, "git add go.sum") {
		t.Errorf("expected second command in output, got %q", out)
	}
	if strings.Contains(out, "No runnable commands") {
		t.Errorf("unexpected no-commands message when commands exist, got %q", out)
	}
}

func TestRenderFixCommandsOnlyIncludesHeader(t *testing.T) {
	r := New(Options{Plain: true, Width: 88, FixCommandsOnly: true})
	a := &model.Analysis{Results: []model.Result{{
		Playbook: model.Playbook{
			ID:  "docker-auth",
			Fix: "```sh\ndocker login\n```",
		},
		Confidence: 0.9,
	}}}
	out := r.RenderFix(a)
	if !strings.Contains(out, "docker-auth") {
		t.Errorf("expected playbook ID in commands-only header, got %q", out)
	}
}

// ── RenderFix with opt-in sections ───────────────────────────────────────────

func TestRenderFixWithPreconditions(t *testing.T) {
	r := New(Options{Plain: true, Width: 88, FixWithPreconditions: true})
	a := &model.Analysis{Results: []model.Result{{
		Playbook: model.Playbook{
			ID:  "go-sum-missing",
			Fix: "## Steps\n\nRun tidy.\n\n## Preconditions\n\nMust have go installed.\n\n## Risks\n\nData loss.",
		},
		Confidence: 0.8,
	}}}
	out := r.RenderFix(a)
	if !strings.Contains(out, "Must have go installed.") {
		t.Errorf("expected preconditions in output, got %q", out)
	}
}

func TestRenderFixWithRisks(t *testing.T) {
	r := New(Options{Plain: true, Width: 88, FixWithRisks: true})
	a := &model.Analysis{Results: []model.Result{{
		Playbook: model.Playbook{
			ID:  "go-sum-missing",
			Fix: "## Steps\n\nRun tidy.\n\n## Preconditions\n\nMust have go installed.\n\n## Risks\n\nData loss.",
		},
		Confidence: 0.8,
	}}}
	out := r.RenderFix(a)
	if !strings.Contains(out, "Data loss.") {
		t.Errorf("expected risks in output, got %q", out)
	}
}

func TestRenderFixExcludesPreconditionsByDefault(t *testing.T) {
	r := New(Options{Plain: true, Width: 88})
	a := &model.Analysis{Results: []model.Result{{
		Playbook: model.Playbook{
			ID:  "go-sum-missing",
			Fix: "## Steps\n\nRun tidy.\n\n## Preconditions\n\nMust have go installed.\n\n## Risks\n\nData loss.",
		},
		Confidence: 0.8,
	}}}
	out := r.RenderFix(a)
	if strings.Contains(out, "Must have go installed.") {
		t.Errorf("expected preconditions stripped by default, got %q", out)
	}
	if strings.Contains(out, "Data loss.") {
		t.Errorf("expected risks stripped by default, got %q", out)
	}
}

func TestRenderFixNoFixStepsMessage(t *testing.T) {
	r := New(Options{Plain: true, Width: 88})
	a := &model.Analysis{Results: []model.Result{{
		Playbook: model.Playbook{
			ID:    "go-sum-missing",
			Title: "Missing go.sum",
			Fix:   "",
		},
		Confidence: 0.8,
	}}}
	out := r.RenderFix(a)
	if !strings.Contains(out, "No fix steps defined") {
		t.Errorf("expected 'No fix steps defined' message for empty fix, got %q", out)
	}
}

// ── renderMatchSummary ────────────────────────────────────────────────────────

func TestRenderMatchSummaryAllFields(t *testing.T) {
	r := New(Options{Plain: true, Width: 88})
	pb := model.Playbook{
		Match: model.MatchSpec{
			Any:  []string{"error A"},
			All:  []string{"error B"},
			None: []string{"error C"},
		},
		Workflow: model.WorkflowSpec{Verify: []string{"check-cmd"}},
	}
	got := r.renderMatchSummary(pb)
	for _, want := range []string{"match.any", "error A", "match.all", "error B", "match.none", "error C", "workflow.verify", "check-cmd"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in match summary, got:\n%s", want, got)
		}
	}
}

func TestRenderMatchSummaryEmpty(t *testing.T) {
	r := New(Options{Plain: true, Width: 88})
	got := r.renderMatchSummary(model.Playbook{})
	if got != "" {
		t.Errorf("expected empty match summary for empty playbook, got %q", got)
	}
}

// ── renderListRow ─────────────────────────────────────────────────────────────

func TestRenderListRowStyled(t *testing.T) {
	r := New(Options{Plain: false, Width: 100, DarkBackground: true})
	pb := model.Playbook{
		ID:       "docker-auth",
		Category: "auth",
		Severity: "high",
		Title:    "Docker auth failure",
	}
	got := r.renderListRow(pb)
	if !strings.Contains(got, "docker-auth") {
		t.Errorf("expected ID in styled row, got %q", got)
	}
}
