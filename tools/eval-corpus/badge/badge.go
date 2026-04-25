// Package badge generates lightweight coverage summary artifacts suitable for
// README badges, release notes, and CI PR comments.
package badge

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"faultline/tools/eval-corpus/model"
)

// Summary is the structured badge artifact.
type Summary struct {
	// CorpusVersion is an optional human-assigned corpus version label.
	CorpusVersion string `json:"corpus_version,omitempty"`
	// CorpusHash is the overall corpus content hash (from the manifest), if
	// available.
	CorpusHash string `json:"corpus_hash,omitempty"`
	// GeneratedAt is an RFC-3339 timestamp.
	GeneratedAt string `json:"generated_at"`
	// CorpusSize is the total number of evaluated fixtures.
	CorpusSize int `json:"corpus_size"`
	// Matched is the number of fixtures matched by a playbook.
	Matched int `json:"matched"`
	// MatchRate is Matched / CorpusSize (0–1).
	MatchRate float64 `json:"match_rate"`
	// Deterministic indicates whether the evaluation was deterministic.
	// "pass", "fail", or "unknown" when no determinism check was run.
	Deterministic string `json:"deterministic"`
	// TopCovered lists the top failure IDs by match count.
	TopCovered []string `json:"top_covered,omitempty"`
	// TopGaps lists the top unmatched cluster snippets.
	TopGaps []string `json:"top_gaps,omitempty"`
}

// Options configures summary generation.
type Options struct {
	// CorpusVersion is an optional label for the corpus version.
	CorpusVersion string
	// CorpusHash from the manifest.
	CorpusHash string
	// TopN controls how many entries appear in TopCovered and TopGaps.
	TopN int
	// Deterministic is "pass", "fail", or "" (unknown).
	Deterministic string
}

// Compute builds a Summary from a slice of EvalResults.
func Compute(results []model.EvalResult, opts Options) Summary {
	topN := opts.TopN
	if topN <= 0 {
		topN = 5
	}

	total := len(results)
	matched := 0
	failureCounts := map[string]int{}
	gapCounts := map[string]int{}
	gapSnippets := map[string]string{}

	for _, r := range results {
		if r.Matched {
			matched++
			if r.FailureID != "" {
				failureCounts[r.FailureID]++
			}
		} else if r.Error == "" && r.FirstLineTag != "" {
			gapCounts[r.FirstLineTag]++
			if _, ok := gapSnippets[r.FirstLineTag]; !ok {
				gapSnippets[r.FirstLineTag] = r.FirstLineSnippet
			}
		}
	}

	rate := 0.0
	if total > 0 {
		rate = float64(matched) / float64(total)
	}

	det := opts.Deterministic
	if det == "" {
		det = "unknown"
	}

	s := Summary{
		CorpusVersion: opts.CorpusVersion,
		CorpusHash:    opts.CorpusHash,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		CorpusSize:    total,
		Matched:       matched,
		MatchRate:     rate,
		Deterministic: det,
		TopCovered:    topKeys(failureCounts, topN),
		TopGaps:       topSnippets(gapCounts, gapSnippets, topN),
	}
	return s
}

// WriteJSON serialises the summary as indented JSON to w.
func WriteJSON(w io.Writer, s Summary) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

// WriteMarkdown writes the human-readable markdown snippet to w.
func WriteMarkdown(w io.Writer, s Summary) {
	fmt.Fprintln(w, "## Faultline Corpus Coverage")
	fmt.Fprintln(w)
	if s.CorpusVersion != "" {
		fmt.Fprintf(w, "**Corpus**: %s\n\n", s.CorpusVersion)
	}
	fmt.Fprintf(w, "| Metric | Value |\n|--------|-------|\n")
	fmt.Fprintf(w, "| Fixtures | %d |\n", s.CorpusSize)
	fmt.Fprintf(w, "| Matched | %d |\n", s.Matched)
	fmt.Fprintf(w, "| Coverage | %.2f%% |\n", s.MatchRate*100)
	fmt.Fprintf(w, "| Determinism | %s |\n", s.Deterministic)
	if s.CorpusHash != "" {
		fmt.Fprintf(w, "| Corpus hash | `%s` |\n", s.CorpusHash[:12])
	}
	fmt.Fprintf(w, "\n_Generated: %s_\n", s.GeneratedAt)

	if len(s.TopCovered) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "**Top covered**: %s\n", joinN(s.TopCovered, topN(s.TopCovered, 5)))
	}
	if len(s.TopGaps) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "**Top gaps**: %s\n", joinN(s.TopGaps, topN(s.TopGaps, 5)))
	}
}

// --- helpers ---

type countEntry struct {
	key     string
	snippet string
	count   int
}

func topKeys(counts map[string]int, n int) []string {
	entries := make([]countEntry, 0, len(counts))
	for k, c := range counts {
		entries = append(entries, countEntry{key: k, count: c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].key < entries[j].key
	})
	if len(entries) > n {
		entries = entries[:n]
	}
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.key
	}
	return out
}

func topSnippets(counts map[string]int, snippets map[string]string, n int) []string {
	entries := make([]countEntry, 0, len(counts))
	for k, c := range counts {
		entries = append(entries, countEntry{key: k, snippet: snippets[k], count: c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].key < entries[j].key
	})
	if len(entries) > n {
		entries = entries[:n]
	}
	out := make([]string, len(entries))
	for i, e := range entries {
		snip := e.snippet
		if snip == "" {
			snip = e.key
		}
		out[i] = snip
	}
	return out
}

func joinN(ss []string, n int) string {
	if n > len(ss) {
		n = len(ss)
	}
	out := ""
	for i, s := range ss[:n] {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}

func topN(ss []string, n int) int {
	if n > len(ss) {
		return len(ss)
	}
	return n
}
