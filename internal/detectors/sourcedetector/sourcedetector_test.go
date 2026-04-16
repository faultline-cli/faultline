package sourcedetector

import (
	"testing"

	"faultline/internal/detectors"
	"faultline/internal/model"
)

func TestKind(t *testing.T) {
	d := Detector{}
	if d.Kind() != detectors.KindSource {
		t.Errorf("Kind() = %q, want %q", d.Kind(), detectors.KindSource)
	}
}

func TestDetectNoTriggerReturnsEmpty(t *testing.T) {
	pb := model.Playbook{
		ID:       "test.src-empty",
		Category: "build",
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "t1", Patterns: []string{"panic("}, Weight: 1.0},
			},
		},
	}
	target := detectors.Target{
		Files: []detectors.SourceFile{
			{Path: "main.go", Lines: []string{"func main() {", "fmt.Println(hello)", "}"}},
		},
	}
	results := Detector{}.Detect([]model.Playbook{pb}, target)
	if len(results) != 0 {
		t.Errorf("expected no results without trigger match, got %d", len(results))
	}
}

func TestDetectTriggerMatchReturnsResult(t *testing.T) {
	pb := model.Playbook{
		ID:        "test.src-panic",
		Category:  "runtime",
		BaseScore: 0.6,
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "t1", Patterns: []string{"panic("}, Weight: 1.0},
			},
		},
	}
	target := detectors.Target{
		Files: []detectors.SourceFile{
			{
				Path:  "server.go",
				Lines: []string{"func serve() {", "\tpanic(unreachable)", "}"},
			},
		},
	}
	results := Detector{}.Detect([]model.Playbook{pb}, target)
	if len(results) == 0 {
		t.Fatal("expected at least one result, got none")
	}
	if results[0].Playbook.ID != "test.src-panic" {
		t.Errorf("result ID = %q, want %q", results[0].Playbook.ID, "test.src-panic")
	}
	if results[0].Score == 0 {
		t.Error("expected non-zero score")
	}
}

func TestDetectResultsSortedByScoreDesc(t *testing.T) {
	pbLow := model.Playbook{
		ID:        "test.src-low",
		Category:  "build",
		BaseScore: 0.1,
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "low", Patterns: []string{"//TODO"}, Weight: 0.1},
			},
		},
	}
	pbHigh := model.Playbook{
		ID:        "test.src-high",
		Category:  "runtime",
		BaseScore: 1.0,
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "high", Patterns: []string{"panic("}, Weight: 1.0},
			},
		},
	}
	target := detectors.Target{
		Files: []detectors.SourceFile{
			{
				Path:  "main.go",
				Lines: []string{"//TODO fix me", "panic(oops)"},
			},
		},
	}
	results := Detector{}.Detect([]model.Playbook{pbLow, pbHigh}, target)
	if len(results) < 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Score < results[1].Score {
		t.Errorf("results not sorted desc: scores %v, %v", results[0].Score, results[1].Score)
	}
	if results[0].Playbook.ID != "test.src-high" {
		t.Errorf("top result = %q, want test.src-high", results[0].Playbook.ID)
	}
}

func TestDetectEmptyFileListReturnsEmpty(t *testing.T) {
	pb := model.Playbook{
		ID:       "test.src-nofiles",
		Category: "build",
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "t1", Patterns: []string{"panic("}, Weight: 1.0},
			},
		},
	}
	target := detectors.Target{Files: nil}
	results := Detector{}.Detect([]model.Playbook{pb}, target)
	if len(results) != 0 {
		t.Errorf("expected no results for empty file list, got %d", len(results))
	}
}

func TestDetectEmptyPlaybookListReturnsEmpty(t *testing.T) {
	target := detectors.Target{
		Files: []detectors.SourceFile{
			{Path: "main.go", Lines: []string{"panic(oops)"}},
		},
	}
	results := Detector{}.Detect(nil, target)
	if len(results) != 0 {
		t.Errorf("expected no results for empty playbook list, got %d", len(results))
	}
}

// ── utility helpers ──────────────────────────────────────────────────────────

func TestAbsPositiveAndNegative(t *testing.T) {
	if abs(5) != 5 {
		t.Error("abs(5) != 5")
	}
	if abs(-5) != 5 {
		t.Error("abs(-5) != 5")
	}
	if abs(0) != 0 {
		t.Error("abs(0) != 0")
	}
}

func TestDefaultFloat(t *testing.T) {
	if defaultFloat(0, 0, 3.14) != 3.14 {
		t.Error("defaultFloat should return first non-zero value")
	}
	if defaultFloat(1.5, 2.5) != 1.5 {
		t.Error("defaultFloat should return first non-zero value")
	}
	if defaultFloat(0, 0) != 0 {
		t.Error("defaultFloat with all zeros should return 0")
	}
}

func TestDefaultInt(t *testing.T) {
	if defaultInt(0, 0, 7) != 7 {
		t.Error("defaultInt should return first non-zero value")
	}
	if defaultInt(3, 5) != 3 {
		t.Error("defaultInt should return first non-zero value")
	}
	if defaultInt(0, 0) != 0 {
		t.Error("defaultInt with all zeros should return 0")
	}
}

func TestCoalesce(t *testing.T) {
	if coalesce("", "  ", "hello") != "hello" {
		t.Error("coalesce should return first non-blank string")
	}
	if coalesce("a", "b") != "a" {
		t.Error("coalesce should return first non-blank string")
	}
	if coalesce("", "  ") != "" {
		t.Error("coalesce with all empty should return empty")
	}
}

func TestExtractField(t *testing.T) {
	line := "Error: code=42 message=bad request"
	if v := extractField(line, "code="); v != "42" {
		t.Errorf("extractField(code=) = %q, want %q", v, "42")
	}
	if v := extractField(line, "missing="); v != "" {
		t.Errorf("extractField(missing=) = %q, want empty", v)
	}
}

func TestScopedKey(t *testing.T) {
	occ := occurrence{
		scopeKey:  "file.go|myFunc",
		moduleKey: "mymodule",
		evidence: model.Evidence{
			File: "file.go",
		},
	}
	if scopedKey("module", occ) != "mymodule" {
		t.Errorf("scopedKey(module) = %q, want mymodule", scopedKey("module", occ))
	}
	if scopedKey("file", occ) != "file.go" {
		t.Errorf("scopedKey(file) = %q, want file.go", scopedKey("file", occ))
	}
	if scopedKey("function", occ) != "file.go|myFunc" {
		t.Errorf("scopedKey(function) = %q, want file.go|myFunc", scopedKey("function", occ))
	}
}

func TestScopedLabel(t *testing.T) {
	tests := []struct {
		scope string
		want  string
	}{
		{"module", "same_module"},
		{"package", "same_module"},
		{"file", "same_file"},
		{"function", "same_function"},
		{"", "same_function"},
	}
	for _, tt := range tests {
		if got := scopedLabel(tt.scope); got != tt.want {
			t.Errorf("scopedLabel(%q) = %q, want %q", tt.scope, got, tt.want)
		}
	}
}

func TestClassifyPath(t *testing.T) {
	tests := []struct {
		path     string
		wantKind string
	}{
		{"internal/testdata/fixture.go", "fixture"},
		{"pkg/examples/demo.go", "example"},
		{"db/migrations/001_init.sql", "migration"},
		{"internal/scripts/deploy.sh", "script"},
		{"internal/admin/dashboard.go", "admin"},
		{"pkg/handler/http.go", "production"},
		{"app_test.go", "test"},
		{"service/tests/helper.go", "test"},
	}
	for _, tt := range tests {
		kind, _, _, _ := classifyPath(tt.path)
		if kind != tt.wantKind {
			t.Errorf("classifyPath(%q) kind = %q, want %q", tt.path, kind, tt.wantKind)
		}
	}
}

func TestContainsAnyPath(t *testing.T) {
	if !containsAnyPath("internal/testdata/fixture.go", []string{"testdata"}) {
		t.Error("expected match for testdata")
	}
	if containsAnyPath("internal/service/handler.go", []string{"testdata"}) {
		t.Error("unexpected match for non-matching path")
	}
}

func TestConfidenceFromScore(t *testing.T) {
	if confidenceFromScore(0, 0) != 0 {
		t.Error("expected 0 confidence when finalScore=0")
	}
	if confidenceFromScore(0, 1) != 0.1 {
		t.Errorf("expected 0.1 confidence when patternScore=0, got %v", confidenceFromScore(0, 1))
	}
	got := confidenceFromScore(10, 10)
	if got != 1.0 {
		t.Errorf("expected 1.0 confidence when pattern==final, got %v", got)
	}
}

func TestProximity(t *testing.T) {
	// Same file, same scope, close lines => 1.0
	a := occurrence{evidence: model.Evidence{File: "x.go", ScopeName: "func1", Line: 5}}
	b := occurrence{evidence: model.Evidence{File: "x.go", ScopeName: "func1", Line: 8}}
	if p := proximity(a, b); p != 1.0 {
		t.Errorf("proximity same scope close lines = %v, want 1.0", p)
	}

	// Same file, same scope, far lines => 0.85
	c := occurrence{evidence: model.Evidence{File: "x.go", ScopeName: "func1", Line: 100}}
	if p := proximity(a, c); p != 0.85 {
		t.Errorf("proximity same scope far lines = %v, want 0.85", p)
	}

	// Same file, different scope => 0.55
	d := occurrence{evidence: model.Evidence{File: "x.go", ScopeName: "func2", Line: 5}}
	if p := proximity(a, d); p != 0.55 {
		t.Errorf("proximity same file diff scope = %v, want 0.55", p)
	}

	// Same module key, different file => 0.3
	e := occurrence{moduleKey: "mod", evidence: model.Evidence{File: "y.go"}}
	f := occurrence{moduleKey: "mod", evidence: model.Evidence{File: "z.go"}}
	if p := proximity(e, f); p != 0.3 {
		t.Errorf("proximity same module = %v, want 0.3", p)
	}

	// Completely different => 0.1
	g := occurrence{moduleKey: "mod1", evidence: model.Evidence{File: "a.go"}}
	h := occurrence{moduleKey: "mod2", evidence: model.Evidence{File: "b.go"}}
	if p := proximity(g, h); p != 0.1 {
		t.Errorf("proximity different module = %v, want 0.1", p)
	}
}

func TestClosestProximity(t *testing.T) {
	item := occurrence{evidence: model.Evidence{File: "a.go", ScopeName: "fn", Line: 5}}
	triggers := []occurrence{
		{evidence: model.Evidence{File: "a.go", ScopeName: "fn", Line: 7}},
	}
	p := closestProximity(item, triggers)
	if p < 0.85 {
		t.Errorf("closestProximity with matching scope = %v, want >= 0.85", p)
	}
}

func TestScopeHasMitigation(t *testing.T) {
	mitigations := []occurrence{
		{scopeKey: "file.go|myFunc"},
	}
	if !scopeHasMitigation("file.go|myFunc", mitigations) {
		t.Error("expected scope mitigation to be found")
	}
	if scopeHasMitigation("file.go|otherFunc", mitigations) {
		t.Error("expected no mitigation for different scope")
	}
}

func TestSortOccurrences(t *testing.T) {
	items := []occurrence{
		{evidence: model.Evidence{File: "b.go", Line: 2}},
		{evidence: model.Evidence{File: "a.go", Line: 10}},
		{evidence: model.Evidence{File: "a.go", Line: 3}},
	}
	sortOccurrences(items)
	if items[0].evidence.File != "a.go" || items[0].evidence.Line != 3 {
		t.Errorf("expected a.go:3 first, got %s:%d", items[0].evidence.File, items[0].evidence.Line)
	}
	if items[1].evidence.File != "a.go" || items[1].evidence.Line != 10 {
		t.Errorf("expected a.go:10 second, got %s:%d", items[1].evidence.File, items[1].evidence.Line)
	}
}
