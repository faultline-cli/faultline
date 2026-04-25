package delta

import (
	"archive/zip"
	"bytes"
	"context"
	"testing"

	"faultline/internal/model"
)

// --- failingTestsFromLog ---

func TestFailingTestsFromLogGoPytest(t *testing.T) {
	log := "FAILED tests/test_api.py::test_login - AssertionError\nFAILED tests/test_db.py::test_connect\n"
	tests := failingTestsFromLog(log)
	if len(tests) != 2 {
		t.Fatalf("expected 2 failing pytest tests, got %v", tests)
	}
}

func TestFailingTestsFromLogGoTests(t *testing.T) {
	log := "--- FAIL: TestAuth (0.01s)\n--- FAIL: TestDB (0.02s)\n"
	tests := failingTestsFromLog(log)
	if len(tests) != 2 {
		t.Fatalf("expected 2 failing Go tests, got %v", tests)
	}
}

func TestFailingTestsFromLogJest(t *testing.T) {
	log := "  ● should authenticate user\n  ● should connect to database\n"
	tests := failingTestsFromLog(log)
	if len(tests) < 1 {
		t.Fatalf("expected jest test failures, got %v", tests)
	}
}

func TestFailingTestsFromLogDeduplicates(t *testing.T) {
	log := "--- FAIL: TestAuth (0.01s)\n--- FAIL: TestAuth (0.01s)\n"
	tests := failingTestsFromLog(log)
	if len(tests) != 1 {
		t.Fatalf("expected 1 deduped test, got %v", tests)
	}
}

func TestFailingTestsFromLogEmptyLog(t *testing.T) {
	tests := failingTestsFromLog("")
	if len(tests) != 0 {
		t.Fatalf("expected no tests from empty log, got %v", tests)
	}
}

func TestFailingTestsFromLogCRLF(t *testing.T) {
	log := "--- FAIL: TestCRLF (0.01s)\r\n"
	tests := failingTestsFromLog(log)
	if len(tests) != 1 {
		t.Fatalf("expected 1 test from CRLF log, got %v", tests)
	}
}

// --- subtractStrings ---

func TestSubtractStringsRemovesBaseline(t *testing.T) {
	current := []string{"a", "b", "c"}
	baseline := []string{"b", "c"}
	result := subtractStrings(current, baseline)
	if len(result) != 1 || result[0] != "a" {
		t.Errorf("expected [a], got %v", result)
	}
}

func TestSubtractStringsEmptyBaseline(t *testing.T) {
	current := []string{"a", "b"}
	result := subtractStrings(current, nil)
	if len(result) != 2 {
		t.Errorf("expected [a b], got %v", result)
	}
}

func TestSubtractStringsAllInBaseline(t *testing.T) {
	current := []string{"a", "b"}
	baseline := []string{"a", "b"}
	result := subtractStrings(current, baseline)
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestSubtractStringsNoCurrent(t *testing.T) {
	result := subtractStrings(nil, []string{"a"})
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

// --- dedupeStrings ---

func TestDedupeStringsRemovesDuplicates(t *testing.T) {
	got := dedupeStrings([]string{"b", "a", "b", "c", "a"})
	// sorted + deduped
	if len(got) != 3 {
		t.Errorf("expected 3 unique items, got %v", got)
	}
}

func TestDedupeStringsFiltersEmpty(t *testing.T) {
	got := dedupeStrings([]string{"", "  ", "a"})
	if len(got) != 1 || got[0] != "a" {
		t.Errorf("expected [a], got %v", got)
	}
}

func TestDedupeStringsNil(t *testing.T) {
	got := dedupeStrings(nil)
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

// --- cloneEnvDiff ---

func TestCloneEnvDiffNilReturnsNil(t *testing.T) {
	if got := cloneEnvDiff(nil); got != nil {
		t.Errorf("expected nil for nil input, got %v", got)
	}
}

func TestCloneEnvDiffEmptyReturnsNil(t *testing.T) {
	if got := cloneEnvDiff(map[string]model.DeltaEnvChange{}); got != nil {
		t.Errorf("expected nil for empty map, got %v", got)
	}
}

func TestCloneEnvDiffFiltersBlankKeys(t *testing.T) {
	in := map[string]model.DeltaEnvChange{
		"":        {Baseline: "x", Current: "y"},
		"  ":      {Baseline: "a", Current: "b"},
		"good_key": {Baseline: "1", Current: "2"},
	}
	out := cloneEnvDiff(in)
	if len(out) != 1 {
		t.Errorf("expected 1 entry (blank keys filtered), got %v", out)
	}
	if out["good_key"].Current != "2" {
		t.Errorf("expected current=2, got %v", out["good_key"])
	}
}

func TestCloneEnvDiffCopiesValues(t *testing.T) {
	in := map[string]model.DeltaEnvChange{
		"sha": {Baseline: "aaa", Current: "bbb"},
	}
	out := cloneEnvDiff(in)
	// Mutate original — clone should be independent
	in["sha"] = model.DeltaEnvChange{Baseline: "xxx", Current: "yyy"}
	if out["sha"].Current != "bbb" {
		t.Errorf("expected clone to be independent, got %v", out["sha"])
	}
}

// --- unzipLogs ---

func TestUnzipLogsInvalidDataErrors(t *testing.T) {
	_, err := unzipLogs([]byte("not a zip file"))
	if err == nil {
		t.Fatal("expected error for invalid zip data")
	}
}

func TestUnzipLogsSkipsDirectories(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	// Add a directory entry
	if _, err := zw.Create("mydir/"); err != nil {
		t.Fatalf("create dir entry: %v", err)
	}
	// Add a real file
	f, err := zw.Create("mydir/file.log")
	if err != nil {
		t.Fatalf("create file entry: %v", err)
	}
	if _, err := f.Write([]byte("content\n")); err != nil {
		t.Fatalf("write file entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	result, err := unzipLogs(buf.Bytes())
	if err != nil {
		t.Fatalf("unzipLogs: %v", err)
	}
	if result != "content\n" {
		t.Errorf("expected 'content\\n', got %q", result)
	}
}

func TestUnzipLogsAddsNewlineToNonTerminatedEntry(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, err := zw.Create("file.log")
	if err != nil {
		t.Fatalf("create file entry: %v", err)
	}
	if _, err := f.Write([]byte("no trailing newline")); err != nil {
		t.Fatalf("write file entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	result, err := unzipLogs(buf.Bytes())
	if err != nil {
		t.Fatalf("unzipLogs: %v", err)
	}
	if result != "no trailing newline\n" {
		t.Errorf("expected newline appended, got %q", result)
	}
}

// --- normalizeProvider ---

func TestNormalizeProviderAliases(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"github", "github-actions"},
		{"GitHub", "github-actions"},
		{"gha", "github-actions"},
		{"GITHUB-ACTIONS", "github-actions"},
		{"gitlab", "gitlab-ci"},
		{"gl", "gitlab-ci"},
		{"gitlab-ci", "gitlab-ci"},
		{"GITLAB-CICD", "gitlab-ci"},
		{"none", "none"},
		{"", ""},
		{"  github  ", "github-actions"},
		{"unknown-provider", "unknown-provider"},
	}
	for _, tc := range cases {
		got := normalizeProvider(tc.input)
		if got != tc.want {
			t.Errorf("normalizeProvider(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// --- Resolver.Resolve ---

func TestResolveUnsupportedProviderErrors(t *testing.T) {
	r := NewResolver(nil)
	_, err := r.Resolve(context.Background(), Options{Provider: "bitbucket"}, "some log")
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
	if got := err.Error(); got == "" {
		t.Error("expected non-empty error message")
	}
}

func TestResolveEmptyProviderReturnsNil(t *testing.T) {
	r := NewResolver(nil)
	snapshot, err := r.Resolve(context.Background(), Options{Provider: ""}, "some log")
	if err != nil {
		t.Fatalf("expected no error for empty provider, got %v", err)
	}
	if snapshot != nil {
		t.Fatalf("expected nil snapshot for empty provider, got %v", snapshot)
	}
}

func TestResolveNoneProviderReturnsNil(t *testing.T) {
	r := NewResolver(nil)
	snapshot, err := r.Resolve(context.Background(), Options{Provider: "none"}, "some log")
	if err != nil {
		t.Fatalf("expected no error for none provider, got %v", err)
	}
	if snapshot != nil {
		t.Fatalf("expected nil snapshot for none provider, got %v", snapshot)
	}
}

func TestNewResolverWithNilClientUsesDefault(t *testing.T) {
	r := NewResolver(nil)
	if r.client == nil {
		t.Fatal("expected non-nil client when nil passed to NewResolver")
	}
}
