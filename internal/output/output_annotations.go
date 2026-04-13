package output

import (
	"fmt"
	"strings"

	"faultline/internal/model"
)

// FormatCIAnnotations emits GitHub Actions-compatible ::warning:: annotations,
// one per matched result up to top.
func FormatCIAnnotations(a *model.Analysis, top int) string {
	if a == nil || len(a.Results) == 0 {
		return ""
	}
	var b strings.Builder
	for _, r := range topN(a.Results, top) {
		fix := ""
		if first := firstMarkdownListItem(r.Playbook.Fix); first != "" {
			fix = " Fix: " + first
		}
		fmt.Fprintf(&b, "::warning title=%s::%s.%s\n",
			r.Playbook.ID, r.Playbook.Title, fix)
	}
	return b.String()
}

// firstMarkdownListItem returns the text of the first bullet or numbered list
// item in a markdown string, or empty string when none is found.
func firstMarkdownListItem(markdown string) string {
	for _, line := range strings.Split(markdown, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "- "):
			return strings.TrimSpace(strings.TrimPrefix(line, "- "))
		case len(line) > 3 && line[1] == '.' && line[2] == ' ' && line[0] >= '0' && line[0] <= '9':
			return strings.TrimSpace(line[3:])
		}
	}
	return ""
}
