package output

import (
	"fmt"
	"strings"

	"faultline/internal/model"
)

// FormatBatchText returns a plain-text summary of a BatchResult.
func FormatBatchText(r *model.BatchResult) string {
	var b strings.Builder
	fileWord := "files"
	if r.Total == 1 {
		fileWord = "file"
	}
	fmt.Fprintf(&b, "FAULTLINE  batch  %d %s\n\n", r.Total, fileWord)

	if r.Matched == 0 {
		fmt.Fprintf(&b, "No playbook matched any of the %d input %s.\n", r.Total, fileWord)
		for _, src := range r.UnmatchedSources {
			fmt.Fprintf(&b, "  %s\n", src)
		}
		return b.String()
	}

	patternWord := "patterns"
	if len(r.Patterns) == 1 {
		patternWord = "pattern"
	}
	fmt.Fprintf(&b, "Patterns  (%d distinct %s)\n", len(r.Patterns), patternWord)
	fmt.Fprintln(&b, strings.Repeat("-", 40))
	for _, pat := range r.Patterns {
		fileCount := "files"
		if pat.Count == 1 {
			fileCount = "file"
		}
		srcDisplay := ""
		if len(pat.Sources) <= 3 {
			srcDisplay = strings.Join(pat.Sources, "  ")
		} else {
			srcDisplay = strings.Join(pat.Sources[:3], "  ") + fmt.Sprintf("  +%d more", len(pat.Sources)-3)
		}
		fmt.Fprintf(&b, "  %-32s  %d %s    %s\n", pat.FailureID, pat.Count, fileCount, srcDisplay)
	}

	if r.Unmatched > 0 {
		fmt.Fprintln(&b)
		unmatchedWord := "files"
		if r.Unmatched == 1 {
			unmatchedWord = "file"
		}
		fmt.Fprintf(&b, "Unmatched  (%d %s)\n", r.Unmatched, unmatchedWord)
		fmt.Fprintln(&b, strings.Repeat("-", 40))
		for _, src := range r.UnmatchedSources {
			fmt.Fprintf(&b, "  %s\n", src)
		}
	}

	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "%d/%d matched", r.Matched, r.Total)
	if len(r.Patterns) > 1 {
		fmt.Fprintf(&b, "  ·  %d distinct patterns", len(r.Patterns))
	} else if len(r.Patterns) == 1 {
		fmt.Fprintf(&b, "  ·  1 pattern")
	}
	if r.Unmatched > 0 {
		fmt.Fprintf(&b, "  ·  %d unmatched", r.Unmatched)
	}
	fmt.Fprintln(&b)
	return b.String()
}

// FormatBatchMarkdown returns a Markdown summary of a BatchResult.
func FormatBatchMarkdown(r *model.BatchResult) string {
	var b strings.Builder
	fileWord := "files"
	if r.Total == 1 {
		fileWord = "file"
	}
	fmt.Fprintf(&b, "# Faultline Batch — %d %s\n\n", r.Total, fileWord)
	fmt.Fprintf(&b, "- Matched: %d/%d\n", r.Matched, r.Total)
	if r.Unmatched > 0 {
		fmt.Fprintf(&b, "- Unmatched: %d/%d\n", r.Unmatched, r.Total)
	}
	if len(r.Patterns) > 0 {
		patternWord := "patterns"
		if len(r.Patterns) == 1 {
			patternWord = "pattern"
		}
		fmt.Fprintf(&b, "- Patterns: %d distinct %s\n", len(r.Patterns), patternWord)
	}

	if len(r.Patterns) > 0 {
		fmt.Fprintf(&b, "\n## Patterns\n\n")
		fmt.Fprintf(&b, "| Pattern | Files | Sources |\n")
		fmt.Fprintf(&b, "|---------|------:|---------|\n")
		for _, pat := range r.Patterns {
			var srcDisplay string
			if len(pat.Sources) <= 3 {
				parts := make([]string, len(pat.Sources))
				for i, s := range pat.Sources {
					parts[i] = "`" + s + "`"
				}
				srcDisplay = strings.Join(parts, " ")
			} else {
				parts := make([]string, 3)
				for i, s := range pat.Sources[:3] {
					parts[i] = "`" + s + "`"
				}
				srcDisplay = strings.Join(parts, " ") + fmt.Sprintf(" +%d more", len(pat.Sources)-3)
			}
			fmt.Fprintf(&b, "| `%s` | %d | %s |\n", pat.FailureID, pat.Count, srcDisplay)
		}
	}

	if r.Unmatched > 0 {
		unmatchedWord := "files"
		if r.Unmatched == 1 {
			unmatchedWord = "file"
		}
		fmt.Fprintf(&b, "\n## Unmatched — %d %s\n\n", r.Unmatched, unmatchedWord)
		for _, src := range r.UnmatchedSources {
			fmt.Fprintf(&b, "- `%s`\n", src)
		}
	}

	if r.Matched == 0 {
		fmt.Fprintf(&b, "\nNo playbook matched any of the %d input %s.\n", r.Total, fileWord)
	}
	return b.String()
}
