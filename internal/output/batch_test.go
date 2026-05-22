package output

import (
	"strings"
	"testing"

	"faultline/internal/model"
)

// ── FormatBatchText ───────────────────────────────────────────────────────────

func TestFormatBatchTextNoMatchesSingleFile(t *testing.T) {
	r := &model.BatchResult{
		Total:            1,
		Matched:          0,
		Unmatched:        1,
		UnmatchedSources: []string{"build.log"},
	}
	out := FormatBatchText(r)
	if !strings.Contains(out, "1 file") {
		t.Errorf("expected singular 'file', got: %s", out)
	}
	if !strings.Contains(out, "No playbook matched") {
		t.Errorf("expected no-match message, got: %s", out)
	}
	if !strings.Contains(out, "build.log") {
		t.Errorf("expected unmatched source listed, got: %s", out)
	}
}

func TestFormatBatchTextNoMatchesMultipleFiles(t *testing.T) {
	r := &model.BatchResult{
		Total:            3,
		Matched:          0,
		Unmatched:        3,
		UnmatchedSources: []string{"a.log", "b.log", "c.log"},
	}
	out := FormatBatchText(r)
	if !strings.Contains(out, "3 files") {
		t.Errorf("expected plural 'files', got: %s", out)
	}
	if !strings.Contains(out, "No playbook matched") {
		t.Errorf("expected no-match message, got: %s", out)
	}
}

func TestFormatBatchTextSinglePattern(t *testing.T) {
	r := &model.BatchResult{
		Total:   2,
		Matched: 2,
		Patterns: []model.BatchPattern{
			{FailureID: "docker-auth", Count: 2, Sources: []string{"a.log", "b.log"}},
		},
	}
	out := FormatBatchText(r)
	if !strings.Contains(out, "docker-auth") {
		t.Errorf("expected failure id in output, got: %s", out)
	}
	if !strings.Contains(out, "1 pattern") {
		t.Errorf("expected singular 'pattern', got: %s", out)
	}
	if !strings.Contains(out, "2/2 matched") {
		t.Errorf("expected match summary, got: %s", out)
	}
	// No 'distinct patterns' for single pattern
	if strings.Contains(out, "distinct patterns") {
		t.Errorf("unexpected 'distinct patterns' text for single pattern, got: %s", out)
	}
}

func TestFormatBatchTextMultiplePatterns(t *testing.T) {
	r := &model.BatchResult{
		Total:   3,
		Matched: 3,
		Patterns: []model.BatchPattern{
			{FailureID: "docker-auth", Count: 2, Sources: []string{"a.log", "b.log"}},
			{FailureID: "build-failure", Count: 1, Sources: []string{"c.log"}},
		},
	}
	out := FormatBatchText(r)
	if !strings.Contains(out, "2 distinct patterns") {
		t.Errorf("expected 'distinct patterns', got: %s", out)
	}
	if !strings.Contains(out, "3/3 matched") {
		t.Errorf("expected match summary, got: %s", out)
	}
}

func TestFormatBatchTextWithUnmatched(t *testing.T) {
	r := &model.BatchResult{
		Total:     3,
		Matched:   2,
		Unmatched: 1,
		Patterns: []model.BatchPattern{
			{FailureID: "docker-auth", Count: 2, Sources: []string{"a.log", "b.log"}},
		},
		UnmatchedSources: []string{"c.log"},
	}
	out := FormatBatchText(r)
	if !strings.Contains(out, "Unmatched") {
		t.Errorf("expected Unmatched section, got: %s", out)
	}
	if !strings.Contains(out, "c.log") {
		t.Errorf("expected unmatched file listed, got: %s", out)
	}
	if !strings.Contains(out, "2/3 matched") {
		t.Errorf("expected partial match summary, got: %s", out)
	}
	if !strings.Contains(out, "1 unmatched") {
		t.Errorf("expected unmatched count in summary, got: %s", out)
	}
}

func TestFormatBatchTextSingleUnmatched(t *testing.T) {
	r := &model.BatchResult{
		Total:     2,
		Matched:   1,
		Unmatched: 1,
		Patterns: []model.BatchPattern{
			{FailureID: "docker-auth", Count: 1, Sources: []string{"a.log"}},
		},
		UnmatchedSources: []string{"b.log"},
	}
	out := FormatBatchText(r)
	if !strings.Contains(out, "1 file") {
		t.Errorf("expected 'Unmatched (1 file)', got: %s", out)
	}
}

func TestFormatBatchTextSourcesTruncationAboveThree(t *testing.T) {
	r := &model.BatchResult{
		Total:   5,
		Matched: 5,
		Patterns: []model.BatchPattern{
			{
				FailureID: "docker-auth",
				Count:     5,
				Sources:   []string{"a.log", "b.log", "c.log", "d.log", "e.log"},
			},
		},
	}
	out := FormatBatchText(r)
	if !strings.Contains(out, "+2 more") {
		t.Errorf("expected '+2 more' truncation, got: %s", out)
	}
}

func TestFormatBatchTextSourcesThreeOrFewer(t *testing.T) {
	r := &model.BatchResult{
		Total:   3,
		Matched: 3,
		Patterns: []model.BatchPattern{
			{
				FailureID: "docker-auth",
				Count:     3,
				Sources:   []string{"a.log", "b.log", "c.log"},
			},
		},
	}
	out := FormatBatchText(r)
	// All three sources shown without truncation
	if strings.Contains(out, "more") {
		t.Errorf("unexpected truncation for 3 sources, got: %s", out)
	}
	if !strings.Contains(out, "a.log") || !strings.Contains(out, "b.log") || !strings.Contains(out, "c.log") {
		t.Errorf("expected all sources in output, got: %s", out)
	}
}

// ── FormatBatchMarkdown ───────────────────────────────────────────────────────

func TestFormatBatchMarkdownNoMatchesSingleFile(t *testing.T) {
	r := &model.BatchResult{
		Total:            1,
		Matched:          0,
		Unmatched:        1,
		UnmatchedSources: []string{"build.log"},
	}
	out := FormatBatchMarkdown(r)
	if !strings.Contains(out, "# Faultline Batch") {
		t.Errorf("expected markdown heading, got: %s", out)
	}
	if !strings.Contains(out, "1 file") {
		t.Errorf("expected singular 'file', got: %s", out)
	}
	if !strings.Contains(out, "No playbook matched") {
		t.Errorf("expected no-match message, got: %s", out)
	}
	if !strings.Contains(out, "`build.log`") {
		t.Errorf("expected unmatched source as code, got: %s", out)
	}
}

func TestFormatBatchMarkdownNoMatchesMultipleFiles(t *testing.T) {
	r := &model.BatchResult{
		Total:            2,
		Matched:          0,
		Unmatched:        2,
		UnmatchedSources: []string{"a.log", "b.log"},
	}
	out := FormatBatchMarkdown(r)
	if !strings.Contains(out, "2 files") {
		t.Errorf("expected plural 'files', got: %s", out)
	}
}

func TestFormatBatchMarkdownSinglePattern(t *testing.T) {
	r := &model.BatchResult{
		Total:   1,
		Matched: 1,
		Patterns: []model.BatchPattern{
			{FailureID: "docker-auth", Count: 1, Sources: []string{"build.log"}},
		},
	}
	out := FormatBatchMarkdown(r)
	if !strings.Contains(out, "## Patterns") {
		t.Errorf("expected Patterns section, got: %s", out)
	}
	if !strings.Contains(out, "`docker-auth`") {
		t.Errorf("expected failure id as code, got: %s", out)
	}
	if !strings.Contains(out, "1 distinct pattern") {
		t.Errorf("expected singular 'pattern', got: %s", out)
	}
}

func TestFormatBatchMarkdownMultiplePatterns(t *testing.T) {
	r := &model.BatchResult{
		Total:   3,
		Matched: 3,
		Patterns: []model.BatchPattern{
			{FailureID: "docker-auth", Count: 2, Sources: []string{"a.log", "b.log"}},
			{FailureID: "build-failure", Count: 1, Sources: []string{"c.log"}},
		},
	}
	out := FormatBatchMarkdown(r)
	if !strings.Contains(out, "2 distinct patterns") {
		t.Errorf("expected plural 'patterns', got: %s", out)
	}
	if !strings.Contains(out, "| Pattern | Files | Sources |") {
		t.Errorf("expected markdown table header, got: %s", out)
	}
}

func TestFormatBatchMarkdownWithUnmatchedSection(t *testing.T) {
	r := &model.BatchResult{
		Total:     2,
		Matched:   1,
		Unmatched: 1,
		Patterns: []model.BatchPattern{
			{FailureID: "docker-auth", Count: 1, Sources: []string{"a.log"}},
		},
		UnmatchedSources: []string{"b.log"},
	}
	out := FormatBatchMarkdown(r)
	if !strings.Contains(out, "## Unmatched") {
		t.Errorf("expected Unmatched section, got: %s", out)
	}
	if !strings.Contains(out, "`b.log`") {
		t.Errorf("expected unmatched source as code, got: %s", out)
	}
	if !strings.Contains(out, "- Unmatched: 1/2") {
		t.Errorf("expected unmatched count in summary, got: %s", out)
	}
}

func TestFormatBatchMarkdownSingleUnmatchedFile(t *testing.T) {
	r := &model.BatchResult{
		Total:     2,
		Matched:   1,
		Unmatched: 1,
		Patterns: []model.BatchPattern{
			{FailureID: "docker-auth", Count: 1, Sources: []string{"a.log"}},
		},
		UnmatchedSources: []string{"b.log"},
	}
	out := FormatBatchMarkdown(r)
	if !strings.Contains(out, "1 file") {
		t.Errorf("expected singular 'file' in unmatched heading, got: %s", out)
	}
}

func TestFormatBatchMarkdownSourcesTruncationAboveThree(t *testing.T) {
	r := &model.BatchResult{
		Total:   5,
		Matched: 5,
		Patterns: []model.BatchPattern{
			{
				FailureID: "docker-auth",
				Count:     5,
				Sources:   []string{"a.log", "b.log", "c.log", "d.log", "e.log"},
			},
		},
	}
	out := FormatBatchMarkdown(r)
	if !strings.Contains(out, "+2 more") {
		t.Errorf("expected '+2 more' truncation, got: %s", out)
	}
}

func TestFormatBatchMarkdownSourcesThreeOrFewer(t *testing.T) {
	r := &model.BatchResult{
		Total:   3,
		Matched: 3,
		Patterns: []model.BatchPattern{
			{
				FailureID: "docker-auth",
				Count:     3,
				Sources:   []string{"x.log", "y.log", "z.log"},
			},
		},
	}
	out := FormatBatchMarkdown(r)
	if strings.Contains(out, "+") && strings.Contains(out, "more") {
		t.Errorf("unexpected truncation for 3 sources, got: %s", out)
	}
	if !strings.Contains(out, "`x.log`") || !strings.Contains(out, "`y.log`") || !strings.Contains(out, "`z.log`") {
		t.Errorf("expected all sources in output, got: %s", out)
	}
}
