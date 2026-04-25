package gaps

import (
	"fmt"
	"io"
	"strings"
)

// PrintClusterSummaryMarkdown writes a cluster-summary.md to w.
func PrintClusterSummaryMarkdown(w io.Writer, clusters []Cluster) {
	fmt.Fprintln(w, "# Unmatched Cluster Summary")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Total clusters: %d\n", len(clusters))
	fmt.Fprintln(w)

	if len(clusters) == 0 {
		fmt.Fprintln(w, "_No unmatched clusters found._")
		return
	}

	fmt.Fprintln(w, "| Rank | Cluster ID | Count | Suspected Class | Confidence | Representative Line |")
	fmt.Fprintln(w, "|------|------------|------:|-----------------|:----------:|---------------------|")
	for i, c := range clusters {
		line := c.RepresentativeErrorLine
		if len(line) > 80 {
			line = line[:77] + "..."
		}
		line = strings.ReplaceAll(line, "|", "\\|")
		fmt.Fprintf(w, "| %d | %s | %d | %s | %.2f | `%s` |\n",
			i+1, c.ClusterID, c.Count, c.SuspectedFailureClass, c.Confidence, line)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "## Cluster Details")
	fmt.Fprintln(w)

	for _, c := range clusters {
		fmt.Fprintf(w, "### %s\n\n", c.ClusterID)
		fmt.Fprintf(w, "- **Count**: %d unmatched fixtures\n", c.Count)
		fmt.Fprintf(w, "- **Suspected class**: %s\n", c.SuspectedFailureClass)
		fmt.Fprintf(w, "- **Confidence**: %.2f\n", c.Confidence)
		if c.Notes != "" {
			fmt.Fprintf(w, "- **Notes**: %s\n", c.Notes)
		}
		fmt.Fprintf(w, "- **Representative error line**:\n\n  ```\n  %s\n  ```\n", c.RepresentativeErrorLine)
		if len(c.SampleFixtureIDs) > 0 {
			fmt.Fprintf(w, "- **Sample fixture IDs**: %s\n", strings.Join(c.SampleFixtureIDs, ", "))
		}
		fmt.Fprintln(w)
	}
}
