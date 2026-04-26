package gaps_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/tools/eval-corpus/gaps"
	"faultline/tools/eval-corpus/model"
)

// --- GenerateStub ---

func TestGenerateStubContainsClusterID(t *testing.T) {
	c := gaps.Cluster{
		ClusterID:               "cluster-001",
		Count:                   5,
		SuspectedFailureClass:   "auth-failure",
		Confidence:              0.85,
		RepresentativeErrorLine: "unauthorized: authentication required",
		SampleFixtureIDs:        []string{"fix-a", "fix-b"},
	}
	stub := gaps.GenerateStub(c)
	if stub.ClusterID != "cluster-001" {
		t.Errorf("ClusterID = %q, want %q", stub.ClusterID, "cluster-001")
	}
	if !strings.Contains(stub.YAML, "cluster-001") {
		t.Errorf("stub YAML does not mention cluster-001:\n%s", stub.YAML)
	}
}

func TestGenerateStubYAMLContainsFailureClass(t *testing.T) {
	c := gaps.Cluster{
		ClusterID:             "cluster-002",
		Count:                 3,
		SuspectedFailureClass: "missing-executable",
		Confidence:            0.75,
	}
	stub := gaps.GenerateStub(c)
	if !strings.Contains(stub.YAML, "missing-executable") {
		t.Errorf("stub YAML does not mention failure class:\n%s", stub.YAML)
	}
	if !strings.Contains(stub.YAML, "stub-missing-executable-cluster-002") {
		t.Errorf("stub YAML does not contain expected ID:\n%s", stub.YAML)
	}
}

func TestGenerateStubYAMLContainsTODOMarkers(t *testing.T) {
	c := gaps.Cluster{
		ClusterID:               "cluster-003",
		SuspectedFailureClass:   "network-timeout",
		RepresentativeErrorLine: "context deadline exceeded",
	}
	stub := gaps.GenerateStub(c)
	if !strings.Contains(stub.YAML, "TODO") {
		t.Errorf("stub YAML should contain TODO markers:\n%s", stub.YAML)
	}
}

func TestGenerateStubHumanNameContainsClass(t *testing.T) {
	c := gaps.Cluster{
		ClusterID:             "cluster-004",
		SuspectedFailureClass: "docker-auth",
	}
	stub := gaps.GenerateStub(c)
	// humanName replaces "-" with " " and adds "stub <ID>"
	if !strings.Contains(stub.YAML, "docker auth") {
		t.Errorf("expected 'docker auth' (humanName) in YAML:\n%s", stub.YAML)
	}
}

func TestGenerateStubCountInComment(t *testing.T) {
	c := gaps.Cluster{
		ClusterID:             "cluster-005",
		Count:                 42,
		SuspectedFailureClass: "compile-error",
	}
	stub := gaps.GenerateStub(c)
	if !strings.Contains(stub.YAML, "42 unmatched fixtures") {
		t.Errorf("stub YAML should mention fixture count:\n%s", stub.YAML)
	}
}

func TestGenerateStubNotesIncluded(t *testing.T) {
	c := gaps.Cluster{
		ClusterID:             "cluster-006",
		SuspectedFailureClass: "test-failure",
		Notes:                 "This pattern appears in Python test logs only.",
	}
	stub := gaps.GenerateStub(c)
	if !strings.Contains(stub.YAML, "Python test logs only") {
		t.Errorf("stub YAML should include Notes:\n%s", stub.YAML)
	}
}

// --- WriteSamples ---

func TestWriteSamplesCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	c := gaps.Cluster{
		ClusterID:        "cluster-001",
		SampleFixtureIDs: []string{"fix-a", "fix-b"},
	}
	index := map[string]model.Fixture{
		"fix-a": {ID: "fix-a", Raw: "log line one"},
		"fix-b": {ID: "fix-b", Raw: "log line two"},
	}
	if err := gaps.WriteSamples(dir, c, index); err != nil {
		t.Fatalf("WriteSamples: %v", err)
	}
	for _, name := range []string{"sample-001.log", "sample-002.log"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected file %s to exist: %v", name, err)
		}
	}
}

func TestWriteSamplesFileContents(t *testing.T) {
	dir := t.TempDir()
	c := gaps.Cluster{
		SampleFixtureIDs: []string{"fix-x"},
	}
	index := map[string]model.Fixture{
		"fix-x": {ID: "fix-x", Raw: "hello world"},
	}
	if err := gaps.WriteSamples(dir, c, index); err != nil {
		t.Fatalf("WriteSamples: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "sample-001.log"))
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}
	if !strings.Contains(string(data), "hello world") {
		t.Errorf("sample file content = %q, want 'hello world'", string(data))
	}
}

func TestWriteSamplesSkipsMissingFixtures(t *testing.T) {
	dir := t.TempDir()
	c := gaps.Cluster{
		SampleFixtureIDs: []string{"present", "missing"},
	}
	index := map[string]model.Fixture{
		"present": {ID: "present", Raw: "log line"},
	}
	if err := gaps.WriteSamples(dir, c, index); err != nil {
		t.Fatalf("WriteSamples: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 file (missing fixture skipped), got %d", len(entries))
	}
}

func TestWriteSamplesEmptyCluster(t *testing.T) {
	dir := t.TempDir()
	c := gaps.Cluster{SampleFixtureIDs: nil}
	if err := gaps.WriteSamples(dir, c, nil); err != nil {
		t.Fatalf("WriteSamples with empty cluster: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 files for empty cluster, got %d", len(entries))
	}
}

// --- PrintClusterSummaryMarkdown ---

func TestPrintClusterSummaryMarkdownEmpty(t *testing.T) {
	var buf bytes.Buffer
	gaps.PrintClusterSummaryMarkdown(&buf, nil)
	out := buf.String()
	if !strings.Contains(out, "No unmatched clusters found") {
		t.Errorf("expected empty message for nil clusters:\n%s", out)
	}
}

func TestPrintClusterSummaryMarkdownContainsHeader(t *testing.T) {
	clusters := []gaps.Cluster{
		{
			ClusterID:               "cluster-001",
			Count:                   7,
			SuspectedFailureClass:   "auth-failure",
			Confidence:              0.90,
			RepresentativeErrorLine: "unauthorized: authentication required",
		},
	}
	var buf bytes.Buffer
	gaps.PrintClusterSummaryMarkdown(&buf, clusters)
	out := buf.String()
	if !strings.Contains(out, "# Unmatched Cluster Summary") {
		t.Errorf("expected heading:\n%s", out)
	}
	if !strings.Contains(out, "cluster-001") {
		t.Errorf("expected cluster ID in output:\n%s", out)
	}
	if !strings.Contains(out, "auth-failure") {
		t.Errorf("expected failure class in output:\n%s", out)
	}
}

func TestPrintClusterSummaryMarkdownTruncatesLongLine(t *testing.T) {
	longLine := strings.Repeat("x", 120)
	clusters := []gaps.Cluster{
		{
			ClusterID:               "cluster-001",
			Count:                   2,
			SuspectedFailureClass:   "unknown",
			Confidence:              0.5,
			RepresentativeErrorLine: longLine,
		},
	}
	var buf bytes.Buffer
	gaps.PrintClusterSummaryMarkdown(&buf, clusters)
	out := buf.String()
	// Only the table row is truncated; the details code fence renders the raw value.
	// Extract the table section (before "## Cluster Details").
	tableSection := out
	if idx := strings.Index(out, "## Cluster Details"); idx >= 0 {
		tableSection = out[:idx]
	}
	if strings.Contains(tableSection, longLine) {
		t.Errorf("expected long line to be truncated in markdown table")
	}
	if !strings.Contains(tableSection, "...") {
		t.Errorf("expected '...' truncation marker in table row")
	}
}

func TestPrintClusterSummaryMarkdownEscapesPipe(t *testing.T) {
	clusters := []gaps.Cluster{
		{
			ClusterID:               "cluster-001",
			Count:                   1,
			SuspectedFailureClass:   "unknown",
			Confidence:              0.5,
			RepresentativeErrorLine: "error: key|value pair",
		},
	}
	var buf bytes.Buffer
	gaps.PrintClusterSummaryMarkdown(&buf, clusters)
	out := buf.String()
	// Only the table row has pipe-escaping; the details code fence renders the raw value.
	// Check only the table section.
	tableSection := out
	if idx := strings.Index(out, "## Cluster Details"); idx >= 0 {
		tableSection = out[:idx]
	}
	if strings.Contains(tableSection, "key|value") {
		t.Errorf("pipe character should be escaped in markdown table:\n%s", tableSection)
	}
	if !strings.Contains(tableSection, `key\|value`) {
		t.Errorf("expected escaped pipe in table row:\n%s", tableSection)
	}
}

func TestPrintClusterSummaryMarkdownDetailsSection(t *testing.T) {
	clusters := []gaps.Cluster{
		{
			ClusterID:               "cluster-001",
			Count:                   3,
			SuspectedFailureClass:   "compile-error",
			Confidence:              0.80,
			Notes:                   "Seen in C++ builds",
			RepresentativeErrorLine: "error: undefined symbol",
			SampleFixtureIDs:        []string{"fix-a", "fix-b"},
		},
	}
	var buf bytes.Buffer
	gaps.PrintClusterSummaryMarkdown(&buf, clusters)
	out := buf.String()
	if !strings.Contains(out, "## Cluster Details") {
		t.Errorf("expected Cluster Details section:\n%s", out)
	}
	if !strings.Contains(out, "Seen in C++ builds") {
		t.Errorf("expected Notes in details:\n%s", out)
	}
	if !strings.Contains(out, "fix-a") {
		t.Errorf("expected sample fixture IDs in details:\n%s", out)
	}
}
