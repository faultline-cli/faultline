package gaps_test

import (
	"strings"
	"testing"

	"faultline/tools/eval-corpus/gaps"
	"faultline/tools/eval-corpus/model"
)

// makeUnmatched builds a slice of unmatched EvalResults with the given
// first-line snippets. IDs are derived from position for uniqueness.
func makeUnmatched(snippets ...string) []model.EvalResult {
	results := make([]model.EvalResult, len(snippets))
	for i, s := range snippets {
		results[i] = model.EvalResult{
			FixtureID:        strings.ReplaceAll(s, " ", "-") + "-" + string(rune('a'+i)),
			Matched:          false,
			FirstLineSnippet: s,
			FirstLineTag:     "tag" + string(rune('a'+i)),
		}
	}
	return results
}

func TestBuildClustersEmpty(t *testing.T) {
	clusters := gaps.BuildClusters(nil, 5)
	if len(clusters) != 0 {
		t.Errorf("expected 0 clusters for empty input, got %d", len(clusters))
	}
}

func TestBuildClustersIgnoresMatched(t *testing.T) {
	results := []model.EvalResult{
		{FixtureID: "m1", Matched: true, FirstLineSnippet: "auth failure"},
		{FixtureID: "m2", Matched: false, FirstLineSnippet: "auth failure"},
	}
	clusters := gaps.BuildClusters(results, 5)
	// only the unmatched one should appear
	if len(clusters) != 1 {
		t.Errorf("expected 1 cluster, got %d", len(clusters))
	}
}

func TestBuildClustersIgnoresErrors(t *testing.T) {
	results := []model.EvalResult{
		{FixtureID: "e1", Matched: false, Error: "engine error", FirstLineSnippet: "something"},
		{FixtureID: "u1", Matched: false, FirstLineSnippet: "something"},
	}
	clusters := gaps.BuildClusters(results, 5)
	if len(clusters) != 1 {
		t.Errorf("expected 1 cluster (error fixture excluded), got %d", len(clusters))
	}
}

func TestBuildClustersDeterministic(t *testing.T) {
	// Two identical fixture sets should produce identical cluster lists.
	snippets := []string{
		"Error response from daemon: unauthorized",
		"exec /usr/bin/node: no such file or directory",
		"Error response from daemon: unauthorized", // duplicate — same cluster
		"cannot find module 'webpack'",
	}
	results := makeUnmatched(snippets...)

	c1 := gaps.BuildClusters(results, 5)
	c2 := gaps.BuildClusters(results, 5)

	if len(c1) != len(c2) {
		t.Fatalf("non-deterministic cluster count: %d vs %d", len(c1), len(c2))
	}
	for i := range c1 {
		if c1[i].ClusterID != c2[i].ClusterID {
			t.Errorf("[%d] ClusterID differs: %q vs %q", i, c1[i].ClusterID, c2[i].ClusterID)
		}
		if c1[i].Count != c2[i].Count {
			t.Errorf("[%d] Count differs: %d vs %d", i, c1[i].Count, c2[i].Count)
		}
	}
}

func TestBuildClustersGroupsByNormalizedLine(t *testing.T) {
	// Lines that differ only in timestamps should fall into the same cluster.
	results := []model.EvalResult{
		{FixtureID: "a", Matched: false, FirstLineSnippet: "2024-01-01T00:00:00Z connection refused to 192.168.1.1:8080"},
		{FixtureID: "b", Matched: false, FirstLineSnippet: "2024-06-15T12:34:56Z connection refused to 10.0.0.1:9090"},
	}
	clusters := gaps.BuildClusters(results, 5)
	if len(clusters) != 1 {
		t.Errorf("expected 1 cluster (lines normalize to same signature), got %d", len(clusters))
	}
}

func TestBuildClustersClusterIDFormat(t *testing.T) {
	results := makeUnmatched("alpha error", "beta error", "gamma error")
	clusters := gaps.BuildClusters(results, 5)
	for i, c := range clusters {
		want := strings.ReplaceAll(c.ClusterID, " ", "")
		// IDs should be cluster-001, cluster-002, …
		if !strings.HasPrefix(c.ClusterID, "cluster-") {
			t.Errorf("[%d] ClusterID %q does not start with 'cluster-'", i, want)
		}
	}
}

func TestBuildClustersMaxSamples(t *testing.T) {
	// Build 10 distinct unmatched fixtures that normalise to the same signature.
	results := make([]model.EvalResult, 10)
	for i := range results {
		results[i] = model.EvalResult{
			FixtureID:        strings.Repeat("x", i+1), // unique IDs
			Matched:          false,
			FirstLineSnippet: "connection refused",
		}
	}
	clusters := gaps.BuildClusters(results, 3)
	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(clusters))
	}
	if len(clusters[0].SampleFixtureIDs) > 3 {
		t.Errorf("SampleFixtureIDs len %d exceeds maxSamples 3", len(clusters[0].SampleFixtureIDs))
	}
}

func TestBuildClustersSortedByCountDesc(t *testing.T) {
	// Create two clusters: one with 3 members, one with 1.
	results := []model.EvalResult{
		{FixtureID: "a1", Matched: false, FirstLineSnippet: "auth failure"},
		{FixtureID: "a2", Matched: false, FirstLineSnippet: "auth failure"},
		{FixtureID: "a3", Matched: false, FirstLineSnippet: "auth failure"},
		{FixtureID: "b1", Matched: false, FirstLineSnippet: "network timeout"},
	}
	clusters := gaps.BuildClusters(results, 5)
	if len(clusters) < 2 {
		t.Fatalf("expected at least 2 clusters, got %d", len(clusters))
	}
	if clusters[0].Count < clusters[1].Count {
		t.Errorf("clusters not sorted by count desc: %d < %d", clusters[0].Count, clusters[1].Count)
	}
}

func TestClassifyAuthFailure(t *testing.T) {
	results := []model.EvalResult{
		{FixtureID: "x", Matched: false, FirstLineSnippet: "unauthorized: authentication required"},
	}
	clusters := gaps.BuildClusters(results, 5)
	if len(clusters) == 0 {
		t.Fatal("expected at least one cluster")
	}
	c := clusters[0]
	if c.SuspectedFailureClass != "auth-failure" {
		t.Errorf("SuspectedFailureClass = %q, want %q", c.SuspectedFailureClass, "auth-failure")
	}
	if c.Confidence == 0 {
		t.Error("Confidence should be > 0 for auth-failure match")
	}
}
