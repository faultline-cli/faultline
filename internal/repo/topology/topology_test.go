package topology

import (
	"os"
	"testing"
	"testing/fstest"
)

// --------------------------------------------------------------------------
// CODEOWNERS parsing
// --------------------------------------------------------------------------

func TestParseCODEOWNERSFile_basic(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	content := "# comment\n" +
		"*.go @team-go\n" +
		"docs/ @team-docs\n" +
		"internal/auth/ @team-auth @team-security\n"

	writeFile(t, dir+"/CODEOWNERS", content)

	rules, err := ParseCODEOWNERS(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("want 3 rules, got %d", len(rules))
	}
	if rules[0].Pattern != "*.go" || len(rules[0].Owners) != 1 || rules[0].Owners[0] != "@team-go" {
		t.Errorf("rule 0 mismatch: %+v", rules[0])
	}
	if rules[2].Pattern != "internal/auth/" || len(rules[2].Owners) != 2 {
		t.Errorf("rule 2 mismatch: %+v", rules[2])
	}
}

func TestParseCODEOWNERS_missing(t *testing.T) {
	t.Parallel()
	rules, err := ParseCODEOWNERS(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error for missing CODEOWNERS: %v", err)
	}
	if rules != nil {
		t.Errorf("want nil rules, got %v", rules)
	}
}

func TestParseCODEOWNERS_githubDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mkdirAll(t, dir+"/.github")
	writeFile(t, dir+"/.github/CODEOWNERS", "* @everyone\n")

	rules, err := ParseCODEOWNERS(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 1 || rules[0].Pattern != "*" {
		t.Errorf("expected wildcard rule, got %v", rules)
	}
}

// --------------------------------------------------------------------------
// OwnersFor
// --------------------------------------------------------------------------

func TestOwnersFor_lastRuleWins(t *testing.T) {
	t.Parallel()
	rules := []OwnerRule{
		{Pattern: "*.go", Owners: []string{"@team-go"}},
		{Pattern: "internal/auth/", Owners: []string{"@team-auth"}},
		{Pattern: "internal/", Owners: []string{"@team-platform"}},
	}
	// The last matching rule for internal/auth/handler.go should be @team-platform
	// because "internal/" is listed after "internal/auth/".
	got := OwnersFor(rules, "internal/auth/handler.go")
	if len(got) != 1 || got[0] != "@team-platform" {
		t.Errorf("want @team-platform, got %v", got)
	}
}

func TestOwnersFor_noMatch(t *testing.T) {
	t.Parallel()
	rules := []OwnerRule{
		{Pattern: "docs/", Owners: []string{"@team-docs"}},
	}
	got := OwnersFor(rules, "internal/handler.go")
	if len(got) != 0 {
		t.Errorf("want no owners, got %v", got)
	}
}

func TestOwnersFor_wildcardExtension(t *testing.T) {
	t.Parallel()
	rules := []OwnerRule{
		{Pattern: "*.go", Owners: []string{"@gophers"}},
	}
	if owners := OwnersFor(rules, "cmd/main.go"); len(owners) != 1 || owners[0] != "@gophers" {
		t.Errorf("want @gophers, got %v", owners)
	}
	if owners := OwnersFor(rules, "main.ts"); len(owners) != 0 {
		t.Errorf("want no owners for .ts, got %v", owners)
	}
}

// --------------------------------------------------------------------------
// Graph
// --------------------------------------------------------------------------

func TestBuildGraph_simpleTree(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"cmd/main.go":           &fstest.MapFile{},
		"internal/auth/auth.go": &fstest.MapFile{},
		"docs/index.md":         &fstest.MapFile{},
	}

	rules := []OwnerRule{
		{Pattern: "cmd/", Owners: []string{"@team-cli"}},
		{Pattern: "internal/auth/", Owners: []string{"@team-auth"}},
		{Pattern: "docs/", Owners: []string{"@team-docs"}},
	}

	graph := BuildGraph(".", rules, fsys)

	if len(graph.Nodes) == 0 {
		t.Fatal("want at least one node")
	}

	// Check that directory nodes were created.
	if _, ok := graph.NodeForPath("cmd"); !ok {
		t.Error("expected node for cmd/")
	}
	if _, ok := graph.NodeForPath("internal"); !ok {
		t.Error("expected node for internal/")
	}
	if _, ok := graph.NodeForPath("docs"); !ok {
		t.Error("expected node for docs/")
	}

	// OwnersFor should walk up the hierarchy.
	owners := graph.OwnersFor("cmd/main.go")
	if len(owners) != 1 || owners[0] != "@team-cli" {
		t.Errorf("want @team-cli for cmd/main.go, got %v", owners)
	}
}

// --------------------------------------------------------------------------
// DeriveSignals
// --------------------------------------------------------------------------

func TestDeriveSignals_boundaryCrossed(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"frontend/app.ts":   &fstest.MapFile{},
		"backend/server.go": &fstest.MapFile{},
	}
	rules := []OwnerRule{
		{Pattern: "frontend/", Owners: []string{"@team-fe"}},
		{Pattern: "backend/", Owners: []string{"@team-be"}},
	}
	graph := BuildGraph(".", rules, fsys)

	sigs := DeriveSignals(graph, []string{"frontend/app.ts"}, []string{"backend/server.go"})
	if !sigs.BoundaryCrossed {
		t.Error("want BoundaryCrossed = true")
	}
	if !containsStr(sigs.ActiveSignals, SignalBoundaryCrossed) {
		t.Errorf("want %s in ActiveSignals, got %v", SignalBoundaryCrossed, sigs.ActiveSignals)
	}
}

func TestDeriveSignals_ownershipMismatch(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"pkg/a/a.go": &fstest.MapFile{},
		"pkg/b/b.go": &fstest.MapFile{},
	}
	rules := []OwnerRule{
		{Pattern: "pkg/a/", Owners: []string{"@team-a"}},
		{Pattern: "pkg/b/", Owners: []string{"@team-b"}},
	}
	graph := BuildGraph(".", rules, fsys)

	sigs := DeriveSignals(graph, []string{"pkg/a/a.go"}, []string{"pkg/b/b.go"})
	if !sigs.OwnershipMismatch {
		t.Error("want OwnershipMismatch = true")
	}
}

func TestDeriveSignals_upstreamChanged(t *testing.T) {
	t.Parallel()
	graph := Graph{index: map[string]*Node{}}

	// No ownership needed for this signal - only directory ancestry matters.
	sigs := DeriveSignals(graph,
		[]string{"lib/core/util.go"},        // changed: ancestor of failing file
		[]string{"lib/core/api/handler.go"}, // failed: inside lib/core/
	)
	if !sigs.UpstreamChanged {
		t.Error("want UpstreamChanged = true when changed dir is ancestor of failure dir")
	}
	if !containsStr(sigs.ActiveSignals, SignalUpstreamChanged) {
		t.Errorf("want %s in ActiveSignals, got %v", SignalUpstreamChanged, sigs.ActiveSignals)
	}
}

func TestDeriveSignals_failureClustered(t *testing.T) {
	t.Parallel()
	graph := Graph{index: map[string]*Node{}}

	sigs := DeriveSignals(graph, nil, []string{
		"pkg/auth/login.go",
		"pkg/auth/logout.go",
		"pkg/auth/middleware.go",
	})
	if !sigs.FailureClustered {
		t.Error("want FailureClustered = true for files in same directory")
	}
	if !containsStr(sigs.ActiveSignals, SignalFailureClustered) {
		t.Errorf("want %s in ActiveSignals, got %v", SignalFailureClustered, sigs.ActiveSignals)
	}
}

func TestDeriveSignals_noSignals(t *testing.T) {
	t.Parallel()
	graph := Graph{index: map[string]*Node{}}
	sigs := DeriveSignals(graph, nil, nil)
	if len(sigs.ActiveSignals) != 0 {
		t.Errorf("want no signals, got %v", sigs.ActiveSignals)
	}
}

func TestDeriveSignals_stable(t *testing.T) {
	t.Parallel()
	graph := Graph{index: map[string]*Node{}}
	in1 := DeriveSignals(graph, []string{"a/b.go"}, []string{"a/b/c.go"})
	in2 := DeriveSignals(graph, []string{"a/b.go"}, []string{"a/b/c.go"})
	if len(in1.ActiveSignals) != len(in2.ActiveSignals) {
		t.Error("DeriveSignals is not deterministic")
	}
	for i, s := range in1.ActiveSignals {
		if in2.ActiveSignals[i] != s {
			t.Errorf("signal order changed at index %d: %s vs %s", i, s, in2.ActiveSignals[i])
		}
	}
}

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdirAll %s: %v", path, err)
	}
}
