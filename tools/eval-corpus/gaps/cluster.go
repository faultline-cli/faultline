// Package gaps identifies high-value missing playbooks by clustering unmatched
// evaluation results into stable, deterministic groups.
//
// Clustering is entirely token-hash based — no ML, no external dependencies.
// The same results always produce the same clusters.
package gaps

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"faultline/tools/eval-corpus/model"
)

// Cluster groups unmatched fixtures that share the same normalized error
// signature.
type Cluster struct {
	// ClusterID is a stable 8-character hex prefix derived from the signature.
	ClusterID string `json:"cluster_id"`
	// Count is the number of unmatched fixtures in this cluster.
	Count int `json:"count"`
	// RepresentativeErrorLine is the first normalized error line shared by
	// most members of the cluster.
	RepresentativeErrorLine string `json:"representative_error_line"`
	// SampleFixtureIDs holds up to maxSamples fixture IDs for representative
	// sample extraction.
	SampleFixtureIDs []string `json:"sample_fixture_ids"`
	// SuspectedFailureClass is a coarse class inferred from token patterns.
	SuspectedFailureClass string `json:"suspected_failure_class"`
	// Notes carries human-readable hints produced during classification.
	Notes string `json:"notes,omitempty"`
	// Confidence is a rough score (0–1) reflecting how strongly the error
	// line tokens match a known class. Always <1 since this is heuristic.
	Confidence float64 `json:"confidence"`
}

// BuildClusters groups unmatched results into Cluster values.
// Only results where Matched==false and Error=="" are considered.
// Results are sorted by FixtureID before processing so output is deterministic.
func BuildClusters(results []model.EvalResult, maxSamples int) []Cluster {
	if maxSamples <= 0 {
		maxSamples = 5
	}

	// Normalise and sort for determinism.
	unmatched := make([]model.EvalResult, 0, len(results))
	for _, r := range results {
		if !r.Matched && r.Error == "" {
			unmatched = append(unmatched, r)
		}
	}
	sort.Slice(unmatched, func(i, j int) bool {
		return unmatched[i].FixtureID < unmatched[j].FixtureID
	})

	// Group by signature (stable 8-char hash of the normalised first line).
	type entry struct {
		sig     string
		repLine string
		ids     []string
	}
	index := map[string]*entry{}
	order := []string{} // preserve first-seen order for deterministic output

	for _, r := range unmatched {
		sig, repLine := signature(r.FirstLineSnippet)
		if e, ok := index[sig]; ok {
			e.ids = append(e.ids, r.FixtureID)
		} else {
			index[sig] = &entry{sig: sig, repLine: repLine, ids: []string{r.FixtureID}}
			order = append(order, sig)
		}
	}

	clusters := make([]Cluster, 0, len(order))
	for _, sig := range order {
		e := index[sig]
		samples := e.ids
		if len(samples) > maxSamples {
			samples = samples[:maxSamples]
		}
		fc, conf, notes := classifyLine(e.repLine)
		clusters = append(clusters, Cluster{
			ClusterID:               sig,
			Count:                   len(e.ids),
			RepresentativeErrorLine: e.repLine,
			SampleFixtureIDs:        samples,
			SuspectedFailureClass:   fc,
			Confidence:              conf,
			Notes:                   notes,
		})
	}

	// Sort clusters by count descending, then by cluster_id for stability.
	sort.Slice(clusters, func(i, j int) bool {
		if clusters[i].Count != clusters[j].Count {
			return clusters[i].Count > clusters[j].Count
		}
		return clusters[i].ClusterID < clusters[j].ClusterID
	})

	// Re-number cluster IDs as cluster-001, cluster-002, … after sorting so
	// the human-visible IDs are stable within a single run.
	for i := range clusters {
		clusters[i].ClusterID = fmt.Sprintf("cluster-%03d", i+1)
	}

	return clusters
}

// signature returns a stable 8-char cluster key and the cleaned representative
// line for a raw snippet.
func signature(snippet string) (sig, repLine string) {
	repLine = cleanLine(snippet)
	h := sha256.Sum256([]byte(repLine))
	sig = hex.EncodeToString(h[:])[:8]
	return sig, repLine
}

// reNoise strips highly variable tokens (timestamps, hex IDs, UUIDs, numbers)
// so that two log lines that differ only in those tokens hash to the same bucket.
var (
	reTimestamp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?`)
	reUUID      = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	reHexID     = regexp.MustCompile(`\b[0-9a-fA-F]{12,}\b`)
	reNumber    = regexp.MustCompile(`\b\d+\b`)
	rePath      = regexp.MustCompile(`(?:/[\w.\-]+){2,}`)
	reURL       = regexp.MustCompile(`https?://\S+`)
	reWS        = regexp.MustCompile(`\s{2,}`)
)

// cleanLine normalizes a log line by stripping variable tokens.
// Lowercasing is deferred until after regex substitutions so that patterns
// like [T ] in the ISO-8601 timestamp regex match the original casing.
func cleanLine(s string) string {
	s = strings.TrimSpace(s)
	s = reTimestamp.ReplaceAllString(s, "<ts>")
	s = reURL.ReplaceAllString(s, "<url>")
	s = reUUID.ReplaceAllString(s, "<uuid>")
	s = reHexID.ReplaceAllString(s, "<hex>")
	s = rePath.ReplaceAllString(s, "<path>")
	s = reNumber.ReplaceAllString(s, "<n>")
	s = reWS.ReplaceAllString(s, " ")
	return strings.ToLower(strings.TrimSpace(s))
}

// failureClass is a heuristic mapping from token patterns to a failure class.
type failureClass struct {
	name       string
	patterns   []*regexp.Regexp
	confidence float64
	notes      string
}

var failureClasses = []failureClass{
	{
		name:       "auth-failure",
		confidence: 0.85,
		notes:      "Token patterns suggest an authentication or authorization failure.",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`\b(?:unauthorized|unauthenticated|forbidden|401|403)\b`),
			regexp.MustCompile(`\b(?:auth(?:entication|orization)?|token|credential|permission|access.denied)\b`),
		},
	},
	{
		name:       "missing-executable",
		confidence: 0.85,
		notes:      "Token patterns suggest a missing binary or file.",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`\b(?:no such file|not found|exec format error|enoent|eacces)\b`),
			regexp.MustCompile(`\b(?:bin|exec|node|python|ruby|java|dotnet|runner)\b`),
		},
	},
	{
		name:       "dependency-resolution",
		confidence: 0.80,
		notes:      "Token patterns suggest a package or dependency resolution failure.",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`\b(?:cannot find|failed to resolve|module not found|import error|unable to load|no matching)\b`),
			regexp.MustCompile(`\b(?:package|module|dependency|require|npm|pip|gradle|maven|cargo)\b`),
		},
	},
	{
		name:       "network-timeout",
		confidence: 0.80,
		notes:      "Token patterns suggest a network or connectivity failure.",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`\b(?:timeout|timed.out|connection refused|connection reset|dns|unreachable|eof)\b`),
			regexp.MustCompile(`\b(?:http|tcp|socket|connect|dial|request|response)\b`),
		},
	},
	{
		name:       "container-runtime",
		confidence: 0.80,
		notes:      "Token patterns suggest a container or OCI runtime failure.",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`\b(?:docker|containerd|podman|oci|container|image|pull|push|registry)\b`),
			regexp.MustCompile(`\b(?:runtime|daemon|manifest|layer|overlay|cgroup)\b`),
		},
	},
	{
		name:       "test-assertion",
		confidence: 0.75,
		notes:      "Token patterns suggest a test assertion or expectation failure.",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`\b(?:assert|expect|should|must|fail(?:ed|ure)|panic)\b`),
			regexp.MustCompile(`\b(?:test|spec|suite|case|scenario|given|when|then)\b`),
		},
	},
}

// classifyLine returns the suspected failure class, confidence, and notes for
// a cleaned log line. Returns "unknown" when no class matches.
func classifyLine(clean string) (class string, confidence float64, notes string) {
	for _, fc := range failureClasses {
		matches := 0
		for _, p := range fc.patterns {
			if p.MatchString(clean) {
				matches++
			}
		}
		if matches == len(fc.patterns) {
			return fc.name, fc.confidence, fc.notes
		}
		// Partial match at half confidence.
		if matches > 0 {
			return fc.name, fc.confidence * 0.5, fc.notes + " (partial match)"
		}
	}
	return "unknown", 0.3, "No known failure class matched. Manual review recommended."
}
