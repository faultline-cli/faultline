package workflow

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/model"
	"faultline/internal/workflow/bind"
	"faultline/internal/workflow/execute"
	"faultline/internal/workflow/plan"
	"faultline/internal/workflow/registry"
	"faultline/internal/workflow/schema"
)

// helpers

func writeTestWorkflow(t *testing.T, dir, workflowID string) {
	t.Helper()
	content := "schema_version: workflow.v1\nworkflow_id: " + workflowID + "\ntitle: Test\ndescription: Test workflow.\nsteps:\n  - id: step1\n    type: noop\n    args: {}\nverification:\n  - id: verify1\n    type: noop\n    args: {}\n    expect: {}\n"
	if err := os.WriteFile(filepath.Join(dir, workflowID+".yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write workflow file: %v", err)
	}
}

func testAnalysisWithArtifact(workflowRef string) *model.Analysis {
	return &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "test-fail",
					Title: "Test Failure",
					Remediation: model.RemediationSpec{
						Workflows: []model.RemediationWorkflowRef{
							{Ref: workflowRef},
						},
					},
				},
				Evidence: []string{"some error"},
			},
		},
	}
}

// --- MarshalPlanJSON ---

func TestMarshalPlanJSONProducesValidJSON(t *testing.T) {
	doc := PlanDocument{
		SchemaVersion: "workflow_plan.v1",
		WorkflowID:    "test.wf",
		Title:         "Test Workflow",
		Mode:          "explain",
	}
	out, err := MarshalPlanJSON(doc)
	if err != nil {
		t.Fatalf("MarshalPlanJSON: %v", err)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("expected trailing newline, got %q", out)
	}
	var round PlanDocument
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &round); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if round.WorkflowID != doc.WorkflowID {
		t.Errorf("workflow_id mismatch: got %q want %q", round.WorkflowID, doc.WorkflowID)
	}
}

func TestMarshalPlanJSONEmptyDocument(t *testing.T) {
	out, err := MarshalPlanJSON(PlanDocument{})
	if err != nil {
		t.Fatalf("MarshalPlanJSON empty: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty JSON for empty document")
	}
}

func TestMarshalPlanJSONIncludesSteps(t *testing.T) {
	doc := PlanDocument{
		WorkflowID: "wf1",
		Steps:      []PlanStep{{ID: "s1", Type: "noop"}},
	}
	out, err := MarshalPlanJSON(doc)
	if err != nil {
		t.Fatalf("MarshalPlanJSON: %v", err)
	}
	if !strings.Contains(out, `"s1"`) {
		t.Errorf("expected step id in JSON, got %q", out)
	}
}

// --- MarshalExecutionJSON ---

func TestMarshalExecutionJSONProducesValidJSON(t *testing.T) {
	record := &model.WorkflowExecutionRecord{
		SchemaVersion: "workflow_execution.v1",
		WorkflowID:    "test.wf",
		Status:        model.WorkflowExecutionStatusSucceeded,
	}
	out, err := MarshalExecutionJSON(record)
	if err != nil {
		t.Fatalf("MarshalExecutionJSON: %v", err)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("expected trailing newline, got %q", out)
	}
	var round model.WorkflowExecutionRecord
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &round); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if round.WorkflowID != record.WorkflowID {
		t.Errorf("workflow_id mismatch: got %q want %q", round.WorkflowID, record.WorkflowID)
	}
}

func TestMarshalExecutionJSONNilRecord(t *testing.T) {
	// A nil *model.WorkflowExecutionRecord is valid: json.Marshal returns "null".
	var record *model.WorkflowExecutionRecord

	out, err := MarshalExecutionJSON(record)
	if err != nil {
		t.Fatalf("unexpected error for nil record: %v", err)
	}
	if out != "null\n" {
		t.Fatalf("expected %q, got %q", "null\n", out)
	}
}

func TestMarshalExecutionJSONEmptyRecord(t *testing.T) {
	out, err := MarshalExecutionJSON(&model.WorkflowExecutionRecord{})
	if err != nil {
		t.Fatalf("unexpected error for empty record: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

// --- selectRecommendation ---

func TestSelectRecommendationNilArtifact(t *testing.T) {
	_, err := selectRecommendation(nil, "")
	if err == nil {
		t.Fatal("expected error for nil artifact")
	}
}

func TestSelectRecommendationNoRecommendations(t *testing.T) {
	artifact := &model.FailureArtifact{
		Fingerprint:             "fp1",
		WorkflowRecommendations: nil,
	}
	_, err := selectRecommendation(artifact, "")
	if err == nil {
		t.Fatal("expected error when no recommendations")
	}
	if !strings.Contains(err.Error(), "fp1") {
		t.Errorf("expected fingerprint in error, got %q", err.Error())
	}
}

func TestSelectRecommendationEmptyRefReturnsFirst(t *testing.T) {
	artifact := &model.FailureArtifact{
		Fingerprint: "fp1",
		WorkflowRecommendations: []model.ArtifactWorkflowRecommendation{
			{Ref: "wf.first"},
			{Ref: "wf.second"},
		},
	}
	rec, err := selectRecommendation(artifact, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Ref != "wf.first" {
		t.Errorf("expected first recommendation, got %q", rec.Ref)
	}
}

func TestSelectRecommendationWhitespaceRefReturnsFirst(t *testing.T) {
	artifact := &model.FailureArtifact{
		Fingerprint: "fp1",
		WorkflowRecommendations: []model.ArtifactWorkflowRecommendation{
			{Ref: "wf.first"},
		},
	}
	rec, err := selectRecommendation(artifact, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Ref != "wf.first" {
		t.Errorf("expected wf.first for whitespace ref, got %q", rec.Ref)
	}
}

func TestSelectRecommendationByRefFound(t *testing.T) {
	artifact := &model.FailureArtifact{
		Fingerprint: "fp1",
		WorkflowRecommendations: []model.ArtifactWorkflowRecommendation{
			{Ref: "wf.first"},
			{Ref: "wf.second"},
		},
	}
	rec, err := selectRecommendation(artifact, "wf.second")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Ref != "wf.second" {
		t.Errorf("expected wf.second, got %q", rec.Ref)
	}
}

func TestSelectRecommendationByRefNotFound(t *testing.T) {
	artifact := &model.FailureArtifact{
		Fingerprint: "fp1",
		WorkflowRecommendations: []model.ArtifactWorkflowRecommendation{
			{Ref: "wf.first"},
		},
	}
	_, err := selectRecommendation(artifact, "wf.missing")
	if err == nil {
		t.Fatal("expected error when ref not found")
	}
	if !strings.Contains(err.Error(), "wf.missing") {
		t.Errorf("expected ref name in error, got %q", err.Error())
	}
}

// --- runtimeContext ---

func TestRuntimeContextUsesRepoPath(t *testing.T) {
	ctx := runtimeContext(Options{RepoPath: "/some/repo"})
	if ctx.WorkDir != "/some/repo" {
		t.Errorf("expected /some/repo, got %q", ctx.WorkDir)
	}
	if ctx.RepoRoot != "/some/repo" {
		t.Errorf("expected RepoRoot /some/repo, got %q", ctx.RepoRoot)
	}
}

func TestRuntimeContextDefaultsToDotWhenEmpty(t *testing.T) {
	ctx := runtimeContext(Options{})
	if ctx.WorkDir != "." {
		t.Errorf("expected '.', got %q", ctx.WorkDir)
	}
}

func TestRuntimeContextTrimsWhitespace(t *testing.T) {
	ctx := runtimeContext(Options{RepoPath: "  /my/path  "})
	if ctx.WorkDir != "/my/path" {
		t.Errorf("expected trimmed path, got %q", ctx.WorkDir)
	}
}

// runtimeContext returns a bind.RuntimeContext — verify the type is assignable.
var _ bind.RuntimeContext = runtimeContext(Options{})

// --- validateDir ---

func TestValidateDirNonExistent(t *testing.T) {
	_, err := validateDir("/nonexistent/path/xyz_faultline_test_12345")
	if err == nil {
		t.Fatal("expected error for non-existent path")
	}
}

func TestValidateDirFileNotDir(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "not-a-dir")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	f.Close()
	_, err = validateDir(f.Name())
	if err == nil {
		t.Fatal("expected error when path is a file, not a directory")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("expected 'not a directory' in error, got %q", err.Error())
	}
}

func TestValidateDirValidDirectory(t *testing.T) {
	dir := t.TempDir()
	got, err := validateDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != dir {
		t.Errorf("expected %q, got %q", dir, got)
	}
}

// --- upwardDirs ---

func TestUpwardDirsIncludesBundledAndWorkflows(t *testing.T) {
	dirs := upwardDirs("/a/b/c")
	if len(dirs) == 0 {
		t.Fatal("expected at least one candidate dir")
	}
	foundBundled := false
	foundWorkflows := false
	for _, d := range dirs {
		slashD := filepath.ToSlash(d)
		if strings.HasSuffix(slashD, "/workflows/bundled") {
			foundBundled = true
		}
		if strings.HasSuffix(slashD, "/workflows") && !strings.HasSuffix(slashD, "/bundled") {
			foundWorkflows = true
		}
	}
	if !foundBundled {
		t.Errorf("expected workflows/bundled path in candidates, got %v", dirs)
	}
	if !foundWorkflows {
		t.Errorf("expected workflows path in candidates, got %v", dirs)
	}
}

func TestUpwardDirsReachesRoot(t *testing.T) {
	dirs := upwardDirs("/a/b/c")
	// /a/b/c, /a/b, /a, / → 4 levels × 2 = 8 entries
	if len(dirs) < 8 {
		t.Errorf("expected >= 8 dirs for /a/b/c, got %d: %v", len(dirs), dirs)
	}
}

func TestUpwardDirsDoesNotLoop(t *testing.T) {
	dirs := upwardDirs("/")
	if len(dirs) == 0 {
		t.Fatal("expected at least one dir for /")
	}
}

// --- defaultDir ---

func TestDefaultDirUsesExplicitPath(t *testing.T) {
	dir := t.TempDir()
	got, err := defaultDir(dir)
	if err != nil {
		t.Fatalf("defaultDir: %v", err)
	}
	if got != dir {
		t.Errorf("expected %q, got %q", dir, got)
	}
}

func TestDefaultDirUsesEnvVar(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(envKey, dir)
	got, err := defaultDir("")
	if err != nil {
		t.Fatalf("defaultDir with env var: %v", err)
	}
	if got != dir {
		t.Errorf("expected %q from env, got %q", dir, got)
	}
}

func TestDefaultDirWhitespaceExplicitFallsBackToEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(envKey, dir)
	got, err := defaultDir("   ")
	if err != nil {
		t.Fatalf("defaultDir: %v", err)
	}
	if got != dir {
		t.Errorf("expected env dir %q, got %q", dir, got)
	}
}

func TestDefaultDirNonExistentExplicitErrors(t *testing.T) {
	t.Setenv(envKey, "")
	_, err := defaultDir("/nonexistent/xyz_faultline_default_dir_test")
	if err == nil {
		t.Fatal("expected error for non-existent explicit path")
	}
}

func TestDefaultDirAutoDiscoveryReturnsErrorWhenNoneFound(t *testing.T) {
	base := t.TempDir()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(base); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	t.Setenv(envKey, "")
	_, err = defaultDir("")
	if err == nil {
		t.Fatal("expected error when auto-discovery finds no workflows/bundled directory")
	}
}

func TestDefaultDirAutoDiscoveryFindsExistingWorkflowDir(t *testing.T) {
	// Build a temp dir tree with a workflows/bundled subdirectory
	base := t.TempDir()
	bundled := filepath.Join(base, "workflows", "bundled")
	if err := os.MkdirAll(bundled, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Set cwd to inside the base so auto-discovery can find it
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(base); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	t.Setenv(envKey, "")
	got, err := defaultDir("")
	if err != nil {
		t.Fatalf("expected auto-discovery to find workflows/bundled, got error: %v", err)
	}
	if got != bundled {
		t.Errorf("expected %q, got %q", bundled, got)
	}
}

// --- LoadCatalog ---

func TestLoadCatalogSucceeds(t *testing.T) {
	dir := t.TempDir()
	writeTestWorkflow(t, dir, "test.wf")

	catalog, err := LoadCatalog(dir)
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	def, ok := catalog.Lookup("test.wf")
	if !ok {
		t.Fatal("expected test.wf in catalog")
	}
	if def.WorkflowID != "test.wf" {
		t.Errorf("unexpected workflow id %q", def.WorkflowID)
	}
}

func TestLoadCatalogEmptyDirErrors(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadCatalog(dir)
	if err == nil {
		t.Fatal("expected error for empty directory")
	}
	if !strings.Contains(err.Error(), "no workflow definitions") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadCatalogNonExistentDirErrors(t *testing.T) {
	_, err := LoadCatalog("/nonexistent/path/xyz_faultline_test_catalog")
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
}

func TestLoadCatalogDuplicateIDErrors(t *testing.T) {
	dir := t.TempDir()
	content := "schema_version: workflow.v1\nworkflow_id: test.dup\ntitle: Dup\ndescription: Dup.\nsteps:\n  - id: s\n    type: noop\n    args: {}\nverification:\n  - id: v\n    type: noop\n    args: {}\n    expect: {}\n"
	if err := os.WriteFile(filepath.Join(dir, "a.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write a.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write b.yaml: %v", err)
	}
	_, err := LoadCatalog(dir)
	if err == nil {
		t.Fatal("expected error for duplicate workflow id")
	}
	if !strings.Contains(err.Error(), "test.dup") {
		t.Errorf("expected duplicate id in error, got %q", err.Error())
	}
}

func TestLoadCatalogWithYmlExtension(t *testing.T) {
	dir := t.TempDir()
	content := "schema_version: workflow.v1\nworkflow_id: test.yml.wf\ntitle: YML\ndescription: Uses .yml.\nsteps:\n  - id: s\n    type: noop\n    args: {}\nverification:\n  - id: v\n    type: noop\n    args: {}\n    expect: {}\n"
	if err := os.WriteFile(filepath.Join(dir, "test.yml.wf.yml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write .yml: %v", err)
	}
	catalog, err := LoadCatalog(dir)
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	if _, ok := catalog.Lookup("test.yml.wf"); !ok {
		t.Fatal("expected test.yml.wf in catalog")
	}
}

func TestLoadCatalogInvalidStepTypeErrors(t *testing.T) {
	dir := t.TempDir()
	content := "schema_version: workflow.v1\nworkflow_id: test.bad\ntitle: Bad\ndescription: Bad.\nsteps:\n  - id: bad\n    type: unknown_type_xyz_99\n    args: {}\nverification:\n  - id: v\n    type: noop\n    args: {}\n    expect: {}\n"
	if err := os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := LoadCatalog(dir)
	if err == nil {
		t.Fatal("expected error for unknown step type")
	}
	if !strings.Contains(err.Error(), "unknown_type_xyz_99") && !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected step type name in error, got %q", err.Error())
	}
}

// --- Catalog.Lookup ---

func TestCatalogLookupNilReturnsNotFound(t *testing.T) {
	var c *Catalog
	_, ok := c.Lookup("anything")
	if ok {
		t.Fatal("expected nil catalog to return not found")
	}
}

func TestCatalogLookupUnknownID(t *testing.T) {
	dir := t.TempDir()
	writeTestWorkflow(t, dir, "test.real")
	catalog, err := LoadCatalog(dir)
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	_, ok := catalog.Lookup("test.nonexistent")
	if ok {
		t.Fatal("expected not found for unknown id")
	}
}

func TestCatalogLookupTrimsWhitespace(t *testing.T) {
	dir := t.TempDir()
	writeTestWorkflow(t, dir, "test.ws")
	catalog, err := LoadCatalog(dir)
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	_, ok := catalog.Lookup("  test.ws  ")
	if !ok {
		t.Fatal("expected whitespace-trimmed lookup to succeed")
	}
}

// --- viewSteps ---

func TestViewStepsEmptySlice(t *testing.T) {
	result := viewSteps(nil)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestViewStepsMapsFields(t *testing.T) {
	reg := registry.Default()
	handler, ok := reg.Lookup("noop")
	if !ok {
		t.Fatal("noop handler not found")
	}
	decodedArgs, err := handler.DecodeArgs(nil)
	if err != nil {
		t.Fatalf("DecodeArgs: %v", err)
	}
	items := []plan.Step{
		{
			Phase:          "steps",
			ID:             "my-step",
			Type:           "noop",
			SafetyClass:    registry.SafetyClassRead,
			Idempotence:    "always",
			Description:    "a noop",
			ResolvedArgs:   map[string]any{"message": "hello"},
			ResolvedExpect: nil,
			DecodedArgs:    decodedArgs,
			Handler:        handler,
		},
	}
	result := viewSteps(items)
	if len(result) != 1 {
		t.Fatalf("expected 1 step, got %d", len(result))
	}
	s := result[0]
	if s.ID != "my-step" {
		t.Errorf("expected id 'my-step', got %q", s.ID)
	}
	if s.Type != "noop" {
		t.Errorf("expected type 'noop', got %q", s.Type)
	}
	if s.Phase != "steps" {
		t.Errorf("expected phase 'steps', got %q", s.Phase)
	}
	if s.SafetyClass != string(registry.SafetyClassRead) {
		t.Errorf("expected safety class 'read', got %q", s.SafetyClass)
	}
}

// --- toDocument ---

func TestToDocumentSetsAllFields(t *testing.T) {
	def, err := schema.LoadBytes([]byte("schema_version: workflow.v1\nworkflow_id: doc.wf\ntitle: Doc\ndescription: Doc test.\npolicy:\n  notes:\n    - check before running\nsteps:\n  - id: s\n    type: noop\n    args: {}\nverification:\n  - id: v\n    type: noop\n    args: {}\n    expect: {}\n"))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}

	compiled := plan.Plan{
		SchemaVersion:     "workflow_plan.v1",
		DefinitionVersion: "workflow.v1",
		WorkflowID:        "doc.wf",
		Title:             "Doc",
		Description:       "Doc test.",
		SourceFingerprint: "fp123",
		SourceFailureID:   "some-failure",
		ResolvedInputs:    map[string]string{"key": "val"},
		RequiredSafety:    []registry.SafetyClass{registry.SafetyClassLocalMutation},
	}

	doc := toDocument(compiled, def, model.WorkflowExecutionModeExplain)
	if doc.WorkflowID != "doc.wf" {
		t.Errorf("expected doc.wf, got %q", doc.WorkflowID)
	}
	if doc.Mode != string(model.WorkflowExecutionModeExplain) {
		t.Errorf("expected explain mode, got %q", doc.Mode)
	}
	if doc.SourceFingerprint != "fp123" {
		t.Errorf("expected fp123, got %q", doc.SourceFingerprint)
	}
	if doc.SourceFailureID != "some-failure" {
		t.Errorf("expected some-failure, got %q", doc.SourceFailureID)
	}
	if doc.ResolvedInputs["key"] != "val" {
		t.Errorf("expected key=val in resolved inputs")
	}
	if len(doc.PolicyNotes) == 0 {
		t.Error("expected policy notes in document")
	}
	if doc.PolicyNotes[0] != "check before running" {
		t.Errorf("unexpected policy note %q", doc.PolicyNotes[0])
	}
	found := false
	for _, s := range doc.RequiredSafety {
		if s == string(registry.SafetyClassLocalMutation) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected local_mutation in required safety, got %v", doc.RequiredSafety)
	}
}

func TestToDocumentEmptyStepsAndVerification(t *testing.T) {
	def, err := schema.LoadBytes([]byte("schema_version: workflow.v1\nworkflow_id: empty.wf\ntitle: Empty\ndescription: Empty.\nsteps:\n  - id: s\n    type: noop\n    args: {}\nverification:\n  - id: v\n    type: noop\n    args: {}\n    expect: {}\n"))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	// compiled plan with no steps; document reflects the compiled plan, not the def
	compiled := plan.Plan{WorkflowID: "empty.wf"}
	doc := toDocument(compiled, def, model.WorkflowExecutionModeDryRun)
	if len(doc.Steps) != 0 {
		t.Errorf("expected no steps in document (plan had no steps), got %d", len(doc.Steps))
	}
	if len(doc.Verification) != 0 {
		t.Errorf("expected no verification, got %d", len(doc.Verification))
	}
}

// --- validateDefinition (internal, exercised via LoadCatalog) ---

func TestValidateDefinitionUnknownStepType(t *testing.T) {
	dir := t.TempDir()
	content := "schema_version: workflow.v1\nworkflow_id: bad.step\ntitle: Bad\ndescription: Bad.\nsteps:\n  - id: s\n    type: nonexistent_step_type\n    args: {}\nverification:\n  - id: v\n    type: noop\n    args: {}\n    expect: {}\n"
	if err := os.WriteFile(filepath.Join(dir, "bad.step.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := LoadCatalog(dir)
	if err == nil {
		t.Fatal("expected error for nonexistent step type")
	}
}

// --- Explain and DryRun (end-to-end) ---

func TestExplainProducesPlanDocument(t *testing.T) {
	dir := t.TempDir()
	writeTestWorkflow(t, dir, "test.noop.wf")
	analysis := testAnalysisWithArtifact("test.noop.wf")

	doc, err := Explain(context.Background(), analysis, Options{WorkflowDir: dir})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	if doc.WorkflowID != "test.noop.wf" {
		t.Errorf("expected test.noop.wf, got %q", doc.WorkflowID)
	}
	if doc.Mode != string(model.WorkflowExecutionModeExplain) {
		t.Errorf("expected explain mode, got %q", doc.Mode)
	}
}

func TestDryRunProducesPlanDocument(t *testing.T) {
	dir := t.TempDir()
	writeTestWorkflow(t, dir, "test.noop.wf")
	analysis := testAnalysisWithArtifact("test.noop.wf")

	doc, err := DryRun(context.Background(), analysis, Options{WorkflowDir: dir})
	if err != nil {
		t.Fatalf("DryRun: %v", err)
	}
	if doc.Mode != string(model.WorkflowExecutionModeDryRun) {
		t.Errorf("expected dry_run mode, got %q", doc.Mode)
	}
}

func TestExplainNilAnalysisErrors(t *testing.T) {
	dir := t.TempDir()
	writeTestWorkflow(t, dir, "test.noop.wf")
	_, err := Explain(context.Background(), nil, Options{WorkflowDir: dir})
	if err == nil {
		t.Fatal("expected error for nil analysis")
	}
}

func TestExplainWorkflowRefNotFoundErrors(t *testing.T) {
	dir := t.TempDir()
	writeTestWorkflow(t, dir, "test.noop.wf")
	// Analysis recommends "nonexistent.wf" which is not in the catalog
	analysis := &model.Analysis{
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    "test-fail",
					Title: "Test Failure",
					Remediation: model.RemediationSpec{
						Workflows: []model.RemediationWorkflowRef{
							{Ref: "nonexistent.wf"},
						},
					},
				},
				Evidence: []string{"some error"},
			},
		},
	}

	_, err := Explain(context.Background(), analysis, Options{WorkflowDir: dir})
	if err == nil {
		t.Fatal("expected error when workflow ref not found in catalog")
	}
	if !strings.Contains(err.Error(), "nonexistent.wf") {
		t.Errorf("expected ref in error, got %q", err.Error())
	}
}

func TestApplyWithNoopWorkflow(t *testing.T) {
	dir := t.TempDir()
	writeTestWorkflow(t, dir, "test.noop.wf")
	analysis := testAnalysisWithArtifact("test.noop.wf")

	record, err := Apply(context.Background(), analysis, Options{WorkflowDir: dir, RepoPath: dir}, execute.Policy{})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if record.Status != model.WorkflowExecutionStatusSucceeded {
		t.Errorf("expected succeeded, got %v", record.Status)
	}
	if record.Title == "" {
		t.Error("expected non-empty title in record")
	}
}
