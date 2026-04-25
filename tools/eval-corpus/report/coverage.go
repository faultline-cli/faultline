// Package report generates human-readable and machine-readable evaluation
// summaries from a corpus of EvalResult records.
package report

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"faultline/tools/eval-corpus/model"
)

// Coverage holds aggregate match statistics for an evaluation run.
type Coverage struct {
	Source    string  `json:"source,omitempty"`
	Total     int     `json:"total"`
	Matched   int     `json:"matched"`
	Unmatched int     `json:"unmatched"`
	Errors    int     `json:"errors"`
	MatchRate float64 `json:"match_rate"`
}

// DistributionEntry records how many times a particular failure ID appeared.
type DistributionEntry struct {
	FailureID string  `json:"failure_id"`
	Count     int     `json:"count"`
	Pct       float64 `json:"pct"`
}

// ClusterEntry groups unmatched results by their first normalised log line.
type ClusterEntry struct {
	Tag     string `json:"tag"`
	Snippet string `json:"snippet"`
	Count   int    `json:"count"`
}

// Report is the full evaluation summary.
type Report struct {
	Coverage     Coverage            `json:"coverage"`
	Distribution []DistributionEntry `json:"distribution"`
	Clusters     []ClusterEntry      `json:"clusters"`
	Determinism  *DeterminismResult  `json:"determinism,omitempty"`
}

// Compute builds a Report from a slice of EvalResults.
// topN controls the maximum number of entries in Distribution and Clusters.
func Compute(results []model.EvalResult, topN int) Report {
	if topN <= 0 {
		topN = 20
	}

	cov := Coverage{}
	if len(results) > 0 {
		cov.Source = results[0].Source
	}

	failureCounts := map[string]int{}
	clusterCounts := map[string]int{}
	clusterSnippets := map[string]string{}

	for _, r := range results {
		cov.Total++
		if r.Error != "" {
			cov.Errors++
		}
		if r.Matched {
			cov.Matched++
			if r.FailureID != "" {
				failureCounts[r.FailureID]++
			}
		} else {
			cov.Unmatched++
			if r.FirstLineTag != "" {
				clusterCounts[r.FirstLineTag]++
				if _, exists := clusterSnippets[r.FirstLineTag]; !exists {
					clusterSnippets[r.FirstLineTag] = r.FirstLineSnippet
				}
			}
		}
	}

	if cov.Total > 0 {
		cov.MatchRate = float64(cov.Matched) / float64(cov.Total)
	}

	dist := buildDistribution(failureCounts, cov.Matched, topN)
	clusters := buildClusters(clusterCounts, clusterSnippets, topN)

	return Report{
		Coverage:     cov,
		Distribution: dist,
		Clusters:     clusters,
	}
}

func buildDistribution(counts map[string]int, total int, topN int) []DistributionEntry {
	entries := make([]DistributionEntry, 0, len(counts))
	for id, n := range counts {
		pct := 0.0
		if total > 0 {
			pct = float64(n) / float64(total) * 100
		}
		entries = append(entries, DistributionEntry{FailureID: id, Count: n, Pct: pct})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Count != entries[j].Count {
			return entries[i].Count > entries[j].Count
		}
		return entries[i].FailureID < entries[j].FailureID
	})
	if len(entries) > topN {
		entries = entries[:topN]
	}
	return entries
}

func buildClusters(counts map[string]int, snippets map[string]string, topN int) []ClusterEntry {
	entries := make([]ClusterEntry, 0, len(counts))
	for tag, n := range counts {
		entries = append(entries, ClusterEntry{
			Tag:     tag,
			Snippet: snippets[tag],
			Count:   n,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Count != entries[j].Count {
			return entries[i].Count > entries[j].Count
		}
		return entries[i].Tag < entries[j].Tag
	})
	if len(entries) > topN {
		entries = entries[:topN]
	}
	return entries
}

// PrintText writes a human-readable coverage report to w.
func PrintText(w io.Writer, rpt Report) {
	c := rpt.Coverage
	fmt.Fprintln(w, "Coverage Report")
	fmt.Fprintln(w, "===============")
	if c.Source != "" {
		fmt.Fprintf(w, "Dataset  : %s\n", c.Source)
	}
	fmt.Fprintf(w, "Total    : %7d\n", c.Total)
	fmt.Fprintf(w, "Matched  : %7d (%5.1f%%)\n", c.Matched, pct(c.Matched, c.Total))
	fmt.Fprintf(w, "Unmatched: %7d (%5.1f%%)\n", c.Unmatched, pct(c.Unmatched, c.Total))
	if c.Errors > 0 {
		fmt.Fprintf(w, "Errors   : %7d (%5.1f%%)\n", c.Errors, pct(c.Errors, c.Total))
	}

	if len(rpt.Distribution) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Match Distribution")
		fmt.Fprintln(w, "==================")
		fmt.Fprintf(w, "%-4s  %-40s  %7s  %6s\n", "Rank", "Failure ID", "Count", "Pct")
		for i, d := range rpt.Distribution {
			fmt.Fprintf(w, "%-4d  %-40s  %7d  %5.1f%%\n", i+1, d.FailureID, d.Count, d.Pct)
		}
	}

	if len(rpt.Clusters) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Unmatched Clusters")
		fmt.Fprintln(w, "==================")
		for _, cl := range rpt.Clusters {
			fmt.Fprintf(w, "[%s] %d logs\n  First line: %q\n", cl.Tag, cl.Count, cl.Snippet)
		}
	}

	if rpt.Determinism != nil {
		fmt.Fprintln(w)
		PrintDeterminismText(w, *rpt.Determinism)
	}
}

// PrintJSON writes a JSON-encoded Report to w.
func PrintJSON(w io.Writer, rpt Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rpt)
}

// LoadResults reads a JSONL results file and returns the decoded slice.
func LoadResults(path string) ([]model.EvalResult, error) {
	f, err := os.Open(path) // #nosec G304 -- operator-provided path
	if err != nil {
		return nil, fmt.Errorf("open results: %w", err)
	}
	defer f.Close()
	return DecodeResults(f)
}

// DecodeResults reads JSONL-encoded EvalResults from r.
func DecodeResults(r io.Reader) ([]model.EvalResult, error) {
	var results []model.EvalResult
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var res model.EvalResult
		if err := json.Unmarshal(line, &res); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
		results = append(results, res)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan results: %w", err)
	}
	return results, nil
}

func pct(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total) * 100
}
