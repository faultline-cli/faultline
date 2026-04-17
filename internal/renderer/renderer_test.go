package renderer

import (
	"strings"
	"testing"

	"faultline/internal/model"
)

func samplePlaybook() model.Playbook {
	return model.Playbook{
		ID:           "go-sum-missing",
		Title:        "Missing go.sum entry",
		Category:     "build",
		Severity:     "medium",
		Summary:      "The build needs a checksum that is missing from `go.sum`.",
		Diagnosis:    "## Diagnosis\n\nThe dependency graph references a module without a checksum entry.",
		Fix:          "## Fix steps\n\n1. Run `go mod tidy`\n2. Commit `go.mod` and `go.sum`",
		Validation:   "## Validation\n\n- Run `go test ./...`",
		WhyItMatters: "## Why it matters\n\nChecksum drift breaks reproducible builds.",
		Match:        model.MatchSpec{Any: []string{"missing go.sum entry"}},
		Workflow:     model.WorkflowSpec{Verify: []string{"go test ./..."}},
	}
}

func TestRenderExplainPlain(t *testing.T) {
	out := New(Options{Plain: true, Width: 88}).RenderExplain(samplePlaybook())
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("plain output should not contain ANSI: %q", out)
	}
	for _, want := range []string{"Summary", "Diagnosis", "Fix", "Validation", "match.any"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in plain output, got:\n%s", want, out)
		}
	}
}

func TestRenderExplainStyled(t *testing.T) {
	out := New(Options{Plain: false, Width: 88, DarkBackground: true}).RenderExplain(samplePlaybook())
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("styled output should contain ANSI escape sequences, got:\n%s", out)
	}
	if !strings.Contains(out, "Diagnosis") {
		t.Fatalf("expected markdown section in styled output, got:\n%s", out)
	}
	for _, unwanted := range []string{"## Diagnosis", "## Why it matters", "## Fix steps", "## Validation"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected duplicate markdown heading %q to be stripped, got:\n%s", unwanted, out)
		}
	}
}

func TestRenderFixUsesSectionStyling(t *testing.T) {
	a := &model.Analysis{Results: []model.Result{{
		Playbook: model.Playbook{
			Title:    "Missing go.sum entry",
			Category: "build",
			Fix:      "## Fix steps\n\n1. Run `go mod tidy`\n2. Commit `go.sum`",
		},
		Confidence: 0.82,
	}}}
	out := New(Options{Plain: true, Width: 88}).RenderFix(a)
	if !strings.Contains(out, "Fix Steps") {
		t.Fatalf("expected fix section header, got:\n%s", out)
	}
	if strings.Contains(out, "## Fix steps") {
		t.Fatalf("expected duplicate markdown heading to be removed, got:\n%s", out)
	}
}

func TestRenderNoMatch(t *testing.T) {
	out := New(Options{Plain: true, Width: 88}).RenderNoMatch()
	if !strings.Contains(out, "No known playbook matched") {
		t.Errorf("RenderNoMatch missing main message, got %q", out)
	}
	if !strings.Contains(out, "faultline list") {
		t.Errorf("RenderNoMatch missing list hint, got %q", out)
	}
}

func TestRenderAnalyzeNilAnalysisCallsRenderNoMatch(t *testing.T) {
	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(nil, 1, false)
	if !strings.Contains(out, "No known playbook matched") {
		t.Errorf("RenderAnalyze(nil) should return no-match message, got %q", out)
	}
}

func TestRenderAnalyzeEmptyResultsCallsRenderNoMatch(t *testing.T) {
	a := &model.Analysis{Results: []model.Result{}}
	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, false)
	if !strings.Contains(out, "No known playbook matched") {
		t.Errorf("RenderAnalyze with empty results should return no-match message, got %q", out)
	}
}

func TestRenderAnalyzeQuickUsesFocusedSections(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{{
			Playbook: model.Playbook{
				ID:       "docker-auth",
				Title:    "Docker registry authentication failure",
				Category: "auth",
				Severity: "high",
				Summary:  "Registry auth failed.",
				Fix:      "1. Verify the registry credential.\n2. Ensure docker login runs first.",
			},
			Confidence: 0.84,
			Score:      2.0,
			Detector:   "log",
			Evidence:   []string{"pull access denied"},
		}},
	}

	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, false)
	for _, want := range []string{"Most Likely Diagnosis", "Matched Evidence", "Recommended Action", "More"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in quick analyze output, got:\n%s", want, out)
		}
	}
}

func TestRenderListPlain(t *testing.T) {
	pbs := []model.Playbook{
		{ID: "docker-auth", Category: "auth", Severity: "high", Title: "Docker Auth"},
		{ID: "oom-killed", Category: "runtime", Severity: "critical", Title: "OOM Killed"},
	}
	out := New(Options{Plain: true, Width: 100}).RenderList(pbs, "")
	if !strings.Contains(out, "docker-auth") {
		t.Errorf("RenderList missing docker-auth, got %q", out)
	}
	if !strings.Contains(out, "oom-killed") {
		t.Errorf("RenderList missing oom-killed, got %q", out)
	}
}

func TestRenderListFiltersByCategory(t *testing.T) {
	pbs := []model.Playbook{
		{ID: "docker-auth", Category: "auth", Severity: "high", Title: "Docker Auth"},
		{ID: "oom-killed", Category: "runtime", Severity: "critical", Title: "OOM Killed"},
	}
	out := New(Options{Plain: true, Width: 100}).RenderList(pbs, "auth")
	if !strings.Contains(out, "docker-auth") {
		t.Errorf("filtered list should contain docker-auth, got %q", out)
	}
	if strings.Contains(out, "oom-killed") {
		t.Errorf("filtered list should not contain oom-killed, got %q", out)
	}
}

func TestRenderListPlainIncludesHeader(t *testing.T) {
	pbs := []model.Playbook{{ID: "a", Category: "test", Severity: "low", Title: "A"}}
	out := New(Options{Plain: true, Width: 100}).RenderList(pbs, "")
	if !strings.Contains(out, "ID") || !strings.Contains(out, "CATEGORY") {
		t.Errorf("RenderList header missing columns, got %q", out)
	}
}

func TestRenderFixNilAnalysis(t *testing.T) {
	out := New(Options{Plain: true, Width: 88}).RenderFix(nil)
	if !strings.Contains(out, "No known playbook matched") {
		t.Errorf("RenderFix(nil) should return no-match, got %q", out)
	}
}

func TestRenderFixEmptyResults(t *testing.T) {
	a := &model.Analysis{Results: []model.Result{}}
	out := New(Options{Plain: true, Width: 88}).RenderFix(a)
	if !strings.Contains(out, "No known playbook matched") {
		t.Errorf("RenderFix(empty) should return no-match, got %q", out)
	}
}

func TestRenderAnalyzeDetailedAddsSpacingUnderHeaders(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   samplePlaybook(),
				Confidence: 0.82,
				Score:      12.5,
				Detector:   "log",
				Evidence:   []string{"missing go.sum entry", "module checksum not found"},
				Explanation: model.ResultExplanation{
					TriggeredBy: []string{"missing go.sum entry"},
				},
				Ranking: &model.Ranking{
					BaselineScore: 10,
					FinalScore:    12.5,
					Contributions: []model.RankingContribution{
						{Feature: "detector_score", Contribution: 1.6, Reason: "baseline detector score remains the anchor"},
						{Feature: "tool_or_stack_match", Contribution: 0.4, Reason: "tool or stack tokens align with the evidence"},
					},
				},
				Breakdown: model.ScoreBreakdown{
					BaseSignalScore:     10,
					FinalScore:          12.5,
					CompoundSignalBonus: 2.5,
				},
			},
			{
				Playbook: model.Playbook{ID: "runner-up", Title: "Runner Up"},
				Score:    11.8,
				Evidence: []string{"missing go.sum entry"},
				Ranking: &model.Ranking{
					BaselineScore: 10,
					FinalScore:    11.8,
					Contributions: []model.RankingContribution{
						{Feature: "detector_score", Contribution: 1.6, Reason: "baseline detector score remains the anchor"},
					},
				},
			},
		},
		Context: model.Context{Stage: "build"},
	}

	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, true)

	for _, want := range []string{
		"Summary\n-------\n\n",
		"Evidence\n--------\n\n",
		"Differential Diagnosis\n----------------------\n\n",
		"Confidence Breakdown\n--------------------\n\n",
		"Triggered by\n------------\n\n",
		"Score Breakdown\n---------------\n\n",
		"Suggested Fix\n-------------\n\n",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected blank line under section header %q, got:\n%s", want, out)
		}
	}
}

// ── quick render helpers ──────────────────────────────────────────────────────

func TestRenderAnalyzeQuickConfidenceLevels(t *testing.T) {
	for _, tc := range []struct {
		confidence float64
		label      string
	}{
		{0.9, "high"},
		{0.5, "medium"},
		{0.1, "low"},
	} {
		a := &model.Analysis{
			Results: []model.Result{{
				Playbook:   samplePlaybook(),
				Confidence: tc.confidence,
				Score:      tc.confidence,
				Evidence:   []string{"build error"},
			}},
		}
		out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, false)
		if !strings.Contains(out, tc.label) {
			t.Errorf("confidence %.2f: expected label %q in output:\n%s", tc.confidence, tc.label, out)
		}
	}
}

func TestRenderAnalyzeQuickWithAlternatives(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{
				Playbook:   samplePlaybook(),
				Confidence: 0.8,
				Score:      0.8,
				Evidence:   []string{"build error"},
			},
			{
				Playbook:   model.Playbook{ID: "runner-up", Title: "Runner Up", Category: "build"},
				Confidence: 0.4,
				Score:      0.4,
				Evidence:   []string{"another error"},
			},
		},
	}
	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 2, false)
	if !strings.Contains(out, "Other Likely Matches") {
		t.Errorf("expected Other Likely Matches section with alternatives:\n%s", out)
	}
	if !strings.Contains(out, "runner-up") {
		t.Errorf("expected runner-up in alternatives:\n%s", out)
	}
}

func TestRenderAnalyzeQuickNoFixSteps(t *testing.T) {
	pb := samplePlaybook()
	pb.Fix = ""
	a := &model.Analysis{
		Results: []model.Result{{
			Playbook:   pb,
			Confidence: 0.75,
			Score:      0.75,
			Evidence:   []string{"build error"},
		}},
	}
	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, false)
	if !strings.Contains(out, "faultline fix") {
		t.Errorf("expected fallback fix hint when no fix steps, got:\n%s", out)
	}
}

func TestRenderAnalyzeQuickManyEvidence(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{{
			Playbook:   samplePlaybook(),
			Confidence: 0.85,
			Score:      0.85,
			Evidence:   []string{"line1", "line2", "line3", "line4", "line5"},
		}},
	}
	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, false)
	// Should be truncated to 3 evidence lines
	if strings.Contains(out, "line4") || strings.Contains(out, "line5") {
		t.Errorf("expected evidence truncated to 3 lines, but line4/line5 appear in:\n%s", out)
	}
}

func TestRenderAnalyzeQuickManyFixSteps(t *testing.T) {
	pb := samplePlaybook()
	pb.Fix = "- Step one\n- Step two\n- Step three\n- Step four"
	a := &model.Analysis{
		Results: []model.Result{{
			Playbook:   pb,
			Confidence: 0.85,
			Score:      0.85,
			Evidence:   []string{"build error"},
		}},
	}
	out := New(Options{Plain: true, Width: 88}).RenderAnalyze(a, 1, false)
	// Should be truncated to 2 fix steps
	if strings.Contains(out, "Step three") || strings.Contains(out, "Step four") {
		t.Errorf("expected fix steps truncated to 2, but step3/step4 appear in:\n%s", out)
	}
}

func TestConfidenceLabelBoundaries(t *testing.T) {
	for _, tc := range []struct {
		confidence float64
		want       string
	}{
		{1.0, "high"},
		{0.8, "high"},
		{0.79, "medium"},
		{0.5, "medium"},
		{0.49, "low"},
		{0.01, "low"},
		{0.0, "unknown"},
	} {
		got := confidenceLabel(tc.confidence)
		if got != tc.want {
			t.Errorf("confidenceLabel(%.2f) = %q, want %q", tc.confidence, got, tc.want)
		}
	}
}

func TestMarkdownListItems(t *testing.T) {
	markdown := "- First item\n- Second item\n1. Third item\n2. Fourth item"
	items := markdownListItems(markdown)
	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d: %v", len(items), items)
	}
	if items[0] != "First item" {
		t.Errorf("expected items[0]=First item, got %q", items[0])
	}
	if items[2] != "Third item" {
		t.Errorf("expected items[2]=Third item, got %q", items[2])
	}
}

func TestMarkdownListItemsEmpty(t *testing.T) {
	if items := markdownListItems(""); len(items) != 0 {
		t.Errorf("expected empty items for empty markdown, got %v", items)
	}
}

func TestTrimTerminalPunctuation(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"Fix it.", "Fix it"},
		{"Fix it", "Fix it"},
		{"", ""},
		{"  spaces  ", "spaces"},
	} {
		got := trimTerminalPunctuation(tc.in)
		if got != tc.want {
			t.Errorf("trimTerminalPunctuation(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
