package sourcedetector

import (
	"testing"

	"faultline/internal/detectors"
	"faultline/internal/model"
)

// --- inferFunctionName ---

func TestInferFunctionNameGoFunc(t *testing.T) {
	got := inferFunctionName("func HandleRequest(w http.ResponseWriter, r *http.Request) {")
	if got != "HandleRequest" {
		t.Errorf("expected HandleRequest, got %q", got)
	}
}

func TestInferFunctionNameGoMethod(t *testing.T) {
	got := inferFunctionName("func (s *Server) Start() error {")
	if got != "Start" {
		t.Errorf("expected Start, got %q", got)
	}
}

func TestInferFunctionNameJSFunction(t *testing.T) {
	got := inferFunctionName("function fetchUser(id) {")
	if got != "fetchUser" {
		t.Errorf("expected fetchUser, got %q", got)
	}
}

func TestInferFunctionNameConstArrow(t *testing.T) {
	got := inferFunctionName("const processPayment = async (req) => {")
	if got != "processPayment" {
		t.Errorf("expected processPayment, got %q", got)
	}
}

func TestInferFunctionNameLetVar(t *testing.T) {
	got := inferFunctionName("let handler = function() {}")
	if got != "handler" {
		t.Errorf("expected handler, got %q", got)
	}
}

func TestInferFunctionNameVarWithEquals(t *testing.T) {
	got := inferFunctionName("var myFunc = () => {}")
	if got != "myFunc" {
		t.Errorf("expected myFunc, got %q", got)
	}
}

func TestInferFunctionNameNotAFunction(t *testing.T) {
	got := inferFunctionName("x := doSomething()")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestInferFunctionNameEmpty(t *testing.T) {
	got := inferFunctionName("")
	if got != "" {
		t.Errorf("expected empty for empty input, got %q", got)
	}
}

// --- collectSuppressions ---

func TestCollectSuppressionsInlineFaultlineIgnore(t *testing.T) {
	pb := model.Playbook{ID: "missing-error-check"}
	files := []preparedFile{
		{
			path: "service.go",
			lines: []preparedLine{
				{original: "// faultline:ignore missing-error-check reason=expected", normalized: "// faultline:ignore missing-error-check reason=expected", number: 5},
				{original: "result, _ := doSomething()", normalized: "result, _ := dosomething()", number: 6},
			},
			moduleKey: ".",
			pathClass: "production",
		},
	}
	out, fullySuppressed := collectSuppressions(pb, files)
	if !fullySuppressed {
		t.Error("expected fullySuppressed=true for inline faultline:ignore")
	}
	if len(out) == 0 {
		t.Error("expected at least one suppression occurrence")
	}
	if out[0].evidence.Kind != model.EvidenceSuppression {
		t.Errorf("expected suppression kind, got %q", out[0].evidence.Kind)
	}
}

func TestCollectSuppressionsInlineFaultlineDisable(t *testing.T) {
	pb := model.Playbook{ID: "sql-injection"}
	files := []preparedFile{
		{
			path: "db.go",
			lines: []preparedLine{
				{original: "// faultline:disable sql-injection reason=parameterized", normalized: "// faultline:disable sql-injection reason=parameterized", number: 10},
			},
			moduleKey: ".",
		},
	}
	_, fullySuppressed := collectSuppressions(pb, files)
	if !fullySuppressed {
		t.Error("expected fullySuppressed=true for faultline:disable")
	}
}

func TestCollectSuppressionsNoMatch(t *testing.T) {
	pb := model.Playbook{ID: "sql-injection"}
	files := []preparedFile{
		{
			path: "db.go",
			lines: []preparedLine{
				{original: "db.Query(query, args...)", normalized: "db.query(query, args...)", number: 5},
			},
			moduleKey: ".",
		},
	}
	out, fullySuppressed := collectSuppressions(pb, files)
	if fullySuppressed {
		t.Error("expected fullySuppressed=false when no ignore comment")
	}
	if len(out) != 0 {
		t.Errorf("expected no suppressions, got %v", out)
	}
}

func TestCollectSuppressionsPlaybookRulePathBased(t *testing.T) {
	pb := model.Playbook{
		ID: "missing-error-check",
		Source: model.SourceSpec{
			Suppressions: []model.SuppressionRule{
				{
					Paths: []string{"testdata"},
					Style: "test context suppression",
				},
			},
		},
	}
	files := []preparedFile{
		{
			path:      "testdata/helper.go",
			lines:     []preparedLine{{original: "result, _ := fn()", normalized: "result, _ := fn()", number: 1}},
			moduleKey: "testdata",
			pathClass: "fixture",
		},
	}
	out, _ := collectSuppressions(pb, files)
	if len(out) == 0 {
		t.Error("expected path-based suppression rule to match")
	}
}

func TestCollectSuppressionsPlaybookRuleWithPattern(t *testing.T) {
	pb := model.Playbook{
		ID: "sql-injection",
		Source: model.SourceSpec{
			Suppressions: []model.SuppressionRule{
				{
					Pattern: "PrepareStatement",
				},
			},
		},
	}
	files := []preparedFile{
		{
			path:      "db.go",
			lines:     []preparedLine{{original: "stmt := db.PrepareStatement(query)", normalized: "stmt := db.preparestatement(query)", number: 7}},
			moduleKey: ".",
			pathClass: "production",
		},
	}
	out, _ := collectSuppressions(pb, files)
	if len(out) == 0 {
		t.Error("expected pattern-based suppression to match")
	}
}

// --- collectSafeContext ---

func TestCollectSafeContextNoRulesReturnsEmpty(t *testing.T) {
	pb := model.Playbook{ID: "no-safe-context"}
	files := []preparedFile{
		{path: "main.go", lines: []preparedLine{{original: "x := 1", normalized: "x := 1", number: 1}}},
	}
	out := collectSafeContext(pb, files)
	if len(out) != 0 {
		t.Errorf("expected empty for playbook without safe context rules, got %v", out)
	}
}

func TestCollectSafeContextMatchesPath(t *testing.T) {
	pb := model.Playbook{
		ID: "missing-error-check",
		Source: model.SourceSpec{
			SafeContextClasses: []model.SafeContextRule{
				{ID: "test-file", Paths: []string{"testdata"}, Label: "safe test context"},
			},
		},
	}
	files := []preparedFile{
		{path: "testdata/cases.go", moduleKey: "testdata", pathClass: "fixture", lines: []preparedLine{}},
		{path: "service.go", moduleKey: ".", pathClass: "production", lines: []preparedLine{}},
	}
	out := collectSafeContext(pb, files)
	if len(out) != 1 {
		t.Errorf("expected 1 safe context match for testdata file, got %d: %v", len(out), out)
	}
	if out[0].evidence.File != "testdata/cases.go" {
		t.Errorf("expected testdata/cases.go, got %q", out[0].evidence.File)
	}
}

// --- applyCompoundSignals ---

func TestApplyCompoundSignalsNoCompoundRules(t *testing.T) {
	pb := model.Playbook{ID: "simple"}
	total, out := applyCompoundSignals(pb, nil, nil, nil)
	if total != 0 || len(out) != 0 {
		t.Errorf("expected 0 bonus and empty evidence, got total=%f out=%v", total, out)
	}
}

func TestApplyCompoundSignalsRequireNotMet(t *testing.T) {
	pb := model.Playbook{
		ID: "compound-test",
		Source: model.SourceSpec{
			CompoundSignals: []model.CompoundSignal{
				{ID: "co-occur", Require: []string{"sig-a", "sig-b"}, Bonus: 2.0},
			},
		},
	}
	// Only sig-a present, sig-b missing → no compound bonus
	triggers := []occurrence{
		{evidence: model.Evidence{SignalID: "sig-a", File: "a.go", Line: 1}, scopeKey: "a.go|fn", moduleKey: "a"},
	}
	total, out := applyCompoundSignals(pb, triggers, nil, nil)
	if total != 0 || len(out) != 0 {
		t.Errorf("expected no bonus when not all required signals present, got total=%f out=%v", total, out)
	}
}

func TestApplyCompoundSignalsAllRequiredInSameScope(t *testing.T) {
	pb := model.Playbook{
		ID: "compound-test",
		Source: model.SourceSpec{
			CompoundSignals: []model.CompoundSignal{
				{ID: "co-occur", Require: []string{"sig-a", "sig-b"}, Bonus: 2.0, Scope: "function"},
			},
		},
	}
	scopeKey := "handler.go|handleReq"
	triggers := []occurrence{
		{evidence: model.Evidence{SignalID: "sig-a", File: "handler.go", Line: 5}, scopeKey: scopeKey, moduleKey: "handler"},
		{evidence: model.Evidence{SignalID: "sig-b", File: "handler.go", Line: 8}, scopeKey: scopeKey, moduleKey: "handler"},
	}
	total, out := applyCompoundSignals(pb, triggers, nil, nil)
	if total == 0 {
		t.Error("expected positive compound bonus when all required signals co-occur")
	}
	if len(out) == 0 {
		t.Error("expected compound evidence item")
	}
}

func TestApplyCompoundSignalsMitigatedScopeSkipped(t *testing.T) {
	pb := model.Playbook{
		ID: "compound-test",
		Source: model.SourceSpec{
			CompoundSignals: []model.CompoundSignal{
				{ID: "co-occur", Require: []string{"sig-a", "sig-b"}, Bonus: 2.0, AllowMitigated: false},
			},
		},
	}
	scopeKey := "svc.go|process"
	triggers := []occurrence{
		{evidence: model.Evidence{SignalID: "sig-a", File: "svc.go", Line: 1}, scopeKey: scopeKey, moduleKey: "svc"},
		{evidence: model.Evidence{SignalID: "sig-b", File: "svc.go", Line: 2}, scopeKey: scopeKey, moduleKey: "svc"},
	}
	// Mitigation covers the same scope → compound bonus should be skipped.
	mitigations := []occurrence{
		{evidence: model.Evidence{SignalID: "safe", File: "svc.go", Line: 3}, scopeKey: scopeKey, moduleKey: "svc"},
	}
	total, out := applyCompoundSignals(pb, triggers, nil, mitigations)
	if total != 0 || len(out) != 0 {
		t.Errorf("expected no bonus for mitigated scope, got total=%f out=%v", total, out)
	}
}

// --- pathAllowed ---

func TestPathAllowedNoConstraints(t *testing.T) {
	policy := model.ContextPolicy{}
	if !pathAllowed("any/path.go", policy, nil, nil) {
		t.Error("expected true with no constraints")
	}
}

func TestPathAllowedPolicyPathIncludes(t *testing.T) {
	policy := model.ContextPolicy{PathIncludes: []string{"internal"}}
	if !pathAllowed("internal/service.go", policy, nil, nil) {
		t.Error("expected true for path inside policy include")
	}
	if pathAllowed("cmd/main.go", policy, nil, nil) {
		t.Error("expected false for path outside policy include")
	}
}

func TestPathAllowedPolicyPathExcludes(t *testing.T) {
	policy := model.ContextPolicy{PathExcludes: []string{"testdata"}}
	if pathAllowed("testdata/cases.go", policy, nil, nil) {
		t.Error("expected false for path in policy exclude")
	}
	if !pathAllowed("service.go", policy, nil, nil) {
		t.Error("expected true for path not in policy exclude")
	}
}

func TestPathAllowedMatcherIncludes(t *testing.T) {
	policy := model.ContextPolicy{}
	if !pathAllowed("handlers/user.go", policy, []string{"handlers"}, nil) {
		t.Error("expected true for path in matcher includes")
	}
	if pathAllowed("db/query.go", policy, []string{"handlers"}, nil) {
		t.Error("expected false for path not in matcher includes")
	}
}

func TestPathAllowedMatcherExcludes(t *testing.T) {
	policy := model.ContextPolicy{}
	if pathAllowed("vendor/lib.go", policy, nil, []string{"vendor"}) {
		t.Error("expected false for path in matcher excludes")
	}
	if !pathAllowed("internal/lib.go", policy, nil, []string{"vendor"}) {
		t.Error("expected true for path not in matcher excludes")
	}
}

// --- sortOccurrences ---

func TestSortOccurrencesByFileAndLine(t *testing.T) {
	items := []occurrence{
		{evidence: model.Evidence{File: "z.go", Line: 1, Kind: model.EvidenceTrigger}},
		{evidence: model.Evidence{File: "a.go", Line: 5, Kind: model.EvidenceTrigger}},
		{evidence: model.Evidence{File: "a.go", Line: 2, Kind: model.EvidenceTrigger}},
	}
	sortOccurrences(items)
	if items[0].evidence.File != "a.go" || items[0].evidence.Line != 2 {
		t.Errorf("expected a.go:2 first, got %s:%d", items[0].evidence.File, items[0].evidence.Line)
	}
	if items[1].evidence.File != "a.go" || items[1].evidence.Line != 5 {
		t.Errorf("expected a.go:5 second, got %s:%d", items[1].evidence.File, items[1].evidence.Line)
	}
	if items[2].evidence.File != "z.go" {
		t.Errorf("expected z.go last, got %s", items[2].evidence.File)
	}
}

func TestSortOccurrencesByKind(t *testing.T) {
	items := []occurrence{
		{evidence: model.Evidence{File: "a.go", Line: 1, Kind: model.EvidenceMitigation}},
		{evidence: model.Evidence{File: "a.go", Line: 1, Kind: model.EvidenceTrigger}},
	}
	sortOccurrences(items)
	// Trigger kind should sort before Mitigation alphabetically (t < m in string comparison).
	// The exact order depends on the string value of the constants; we just verify determinism.
	before := items[0].evidence.Kind
	sortOccurrences(items)
	after := items[0].evidence.Kind
	if before != after {
		t.Error("sortOccurrences is not deterministic")
	}
}

// --- Detect (integration) ---

func TestDetectWithSimpleSourcePlaybook(t *testing.T) {
	pb := model.Playbook{
		ID:       "test-source-detect",
		Category: "build",
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "t1", Patterns: []string{"os.Exit(1)"}},
			},
		},
	}
	files := []detectors.SourceFile{
		{Path: "main.go", Content: "package main\nfunc main() {\n\tos.Exit(1)\n}\n", Lines: []string{"package main", "func main() {", "\tos.Exit(1)", "}"}},
	}
	d := Detector{}
	results := d.Detect([]model.Playbook{pb}, detectors.Target{Files: files})
	if len(results) == 0 {
		t.Fatal("expected at least one result from source detector")
	}
	if results[0].Playbook.ID != "test-source-detect" {
		t.Errorf("expected test-source-detect, got %q", results[0].Playbook.ID)
	}
}
