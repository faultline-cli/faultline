package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	artifactpkg "faultline/internal/artifact"
	"faultline/internal/model"
	"faultline/internal/workflow/bind"
	"faultline/internal/workflow/execute"
	"faultline/internal/workflow/plan"
	"faultline/internal/workflow/registry"
	"faultline/internal/workflow/schema"
)

const envKey = "FAULTLINE_WORKFLOW_DIR"

type Options struct {
	WorkflowDir string
	WorkflowRef string
	RepoPath    string
}

type PlanStep struct {
	Phase       string         `json:"phase,omitempty"`
	ID          string         `json:"id,omitempty"`
	Type        string         `json:"type,omitempty"`
	SafetyClass string         `json:"safety_class,omitempty"`
	Idempotence string         `json:"idempotence,omitempty"`
	Description string         `json:"description,omitempty"`
	Args        map[string]any `json:"args,omitempty"`
	Expect      map[string]any `json:"expect,omitempty"`
}

type PlanDocument struct {
	SchemaVersion     string            `json:"schema_version"`
	DefinitionVersion string            `json:"definition_schema_version"`
	WorkflowID        string            `json:"workflow_id"`
	Title             string            `json:"title"`
	Description       string            `json:"description"`
	Mode              string            `json:"mode"`
	SourceFingerprint string            `json:"source_fingerprint,omitempty"`
	SourceFailureID   string            `json:"source_failure_id,omitempty"`
	ResolvedInputs    map[string]string `json:"resolved_inputs,omitempty"`
	RequiredSafety    []string          `json:"required_safety,omitempty"`
	PolicyNotes       []string          `json:"policy_notes,omitempty"`
	Steps             []PlanStep        `json:"steps,omitempty"`
	Verification      []PlanStep        `json:"verification,omitempty"`
}

type Catalog struct {
	definitions map[string]schema.Definition
}

func Explain(ctx context.Context, analysis *model.Analysis, opts Options) (PlanDocument, error) {
	return describe(ctx, analysis, opts, model.WorkflowExecutionModeExplain, false)
}

func DryRun(ctx context.Context, analysis *model.Analysis, opts Options) (PlanDocument, error) {
	return describe(ctx, analysis, opts, model.WorkflowExecutionModeDryRun, true)
}

func Apply(ctx context.Context, analysis *model.Analysis, opts Options, policy execute.Policy) (*model.WorkflowExecutionRecord, error) {
	compiled, definition, err := resolvePlan(ctx, analysis, opts, false)
	if err != nil {
		return nil, err
	}
	record, err := execute.Apply(ctx, compiled, execute.Options{
		Runtime:  runtimeContext(opts),
		Artifact: analysis.Artifact,
		Policy:   policy,
	})
	if record != nil {
		record.Title = definition.Title
	}
	return record, err
}

func MarshalPlanJSON(doc PlanDocument) (string, error) {
	data, err := json.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("marshal workflow plan json: %w", err)
	}
	return string(data) + "\n", nil
}

func MarshalExecutionJSON(record *model.WorkflowExecutionRecord) (string, error) {
	data, err := json.Marshal(record)
	if err != nil {
		return "", fmt.Errorf("marshal workflow execution json: %w", err)
	}
	return string(data) + "\n", nil
}

func LoadCatalog(dir string) (*Catalog, error) {
	reg := registry.Default()
	root, err := defaultDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk workflow catalog: %w", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no workflow definitions found in %s", root)
	}
	catalog := &Catalog{definitions: map[string]schema.Definition{}}
	for _, path := range files {
		def, err := schema.LoadFile(path)
		if err != nil {
			return nil, err
		}
		if _, ok := catalog.definitions[def.WorkflowID]; ok {
			return nil, fmt.Errorf("duplicate workflow id %q", def.WorkflowID)
		}
		if err := validateDefinition(def, reg); err != nil {
			return nil, fmt.Errorf("workflow %s: %w", def.WorkflowID, err)
		}
		catalog.definitions[def.WorkflowID] = def
	}
	return catalog, nil
}

func (c *Catalog) Lookup(id string) (schema.Definition, bool) {
	if c == nil {
		return schema.Definition{}, false
	}
	def, ok := c.definitions[strings.TrimSpace(id)]
	return def, ok
}

func describe(ctx context.Context, analysis *model.Analysis, opts Options, mode model.WorkflowExecutionMode, probeRead bool) (PlanDocument, error) {
	compiled, definition, err := resolvePlan(ctx, analysis, opts, probeRead)
	if err != nil {
		return PlanDocument{}, err
	}
	return toDocument(compiled, definition, mode), nil
}

func resolvePlan(ctx context.Context, analysis *model.Analysis, opts Options, probeRead bool) (plan.Plan, schema.Definition, error) {
	if analysis == nil {
		return plan.Plan{}, schema.Definition{}, fmt.Errorf("analysis is required")
	}
	analysis = artifactpkg.Sync(analysis)
	if analysis.Artifact == nil {
		return plan.Plan{}, schema.Definition{}, fmt.Errorf("analysis artifact is required")
	}
	recommendation, err := selectRecommendation(analysis.Artifact, opts.WorkflowRef)
	if err != nil {
		return plan.Plan{}, schema.Definition{}, err
	}
	catalog, err := LoadCatalog(opts.WorkflowDir)
	if err != nil {
		return plan.Plan{}, schema.Definition{}, err
	}
	definition, ok := catalog.Lookup(recommendation.Ref)
	if !ok {
		return plan.Plan{}, schema.Definition{}, fmt.Errorf("workflow definition %q not found", recommendation.Ref)
	}
	compiled, err := plan.Build(ctx, definition, plan.BuildOptions{
		Recommendation: recommendation,
		Artifact:       analysis.Artifact,
		Runtime:        runtimeContext(opts),
		Registry:       registry.Default(),
		ProbeReadSteps: probeRead,
	})
	if err != nil {
		return plan.Plan{}, schema.Definition{}, err
	}
	return compiled, definition, nil
}

func toDocument(compiled plan.Plan, definition schema.Definition, mode model.WorkflowExecutionMode) PlanDocument {
	document := PlanDocument{
		SchemaVersion:     compiled.SchemaVersion,
		DefinitionVersion: compiled.DefinitionVersion,
		WorkflowID:        compiled.WorkflowID,
		Title:             compiled.Title,
		Description:       compiled.Description,
		Mode:              string(mode),
		SourceFingerprint: compiled.SourceFingerprint,
		SourceFailureID:   compiled.SourceFailureID,
		ResolvedInputs:    compiled.ResolvedInputs,
		PolicyNotes:       append([]string{}, definition.Policy.Notes...),
	}
	for _, class := range compiled.RequiredSafety {
		document.RequiredSafety = append(document.RequiredSafety, string(class))
	}
	document.Steps = append(document.Steps, viewSteps(compiled.Steps)...)
	document.Verification = append(document.Verification, viewSteps(compiled.Verification)...)
	return document
}

func viewSteps(items []plan.Step) []PlanStep {
	steps := make([]PlanStep, 0, len(items))
	for _, item := range items {
		steps = append(steps, PlanStep{
			Phase:       item.Phase,
			ID:          item.ID,
			Type:        item.Type,
			SafetyClass: string(item.SafetyClass),
			Idempotence: item.Idempotence,
			Description: item.Description,
			Args:        item.ResolvedArgs,
			Expect:      item.ResolvedExpect,
		})
	}
	return steps
}

func validateDefinition(def schema.Definition, reg *registry.Registry) error {
	for _, step := range append(append([]schema.Step{}, def.Steps...), def.Verification...) {
		handler, err := reg.MustLookup(step.Type)
		if err != nil {
			return err
		}
		if _, err := handler.DecodeArgs(step.Args); err != nil {
			return err
		}
		if _, err := handler.DecodeExpect(step.Expect); err != nil {
			return err
		}
	}
	return nil
}

func selectRecommendation(artifact *model.FailureArtifact, workflowRef string) (model.ArtifactWorkflowRecommendation, error) {
	if artifact == nil {
		return model.ArtifactWorkflowRecommendation{}, fmt.Errorf("failure artifact is required")
	}
	if len(artifact.WorkflowRecommendations) == 0 {
		return model.ArtifactWorkflowRecommendation{}, fmt.Errorf("artifact %s does not recommend any remediation workflows", artifact.Fingerprint)
	}
	requested := strings.TrimSpace(workflowRef)
	if requested == "" {
		return artifact.WorkflowRecommendations[0], nil
	}
	for _, item := range artifact.WorkflowRecommendations {
		if item.Ref == requested {
			return item, nil
		}
	}
	return model.ArtifactWorkflowRecommendation{}, fmt.Errorf("artifact %s does not recommend workflow %q", artifact.Fingerprint, requested)
}

func runtimeContext(opts Options) bind.RuntimeContext {
	workdir := strings.TrimSpace(opts.RepoPath)
	if workdir == "" {
		workdir = "."
	}
	return bind.RuntimeContext{
		WorkDir:  workdir,
		RepoRoot: workdir,
	}
}

func defaultDir(explicit string) (string, error) {
	if value := strings.TrimSpace(explicit); value != "" {
		return validateDir(value)
	}
	if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
		return validateDir(value)
	}
	var candidates []string
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, upwardDirs(cwd)...)
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, upwardDirs(filepath.Dir(exe))...)
	}
	candidates = append(candidates, "/workflows/bundled", "/workflows")
	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		if dir, err := validateDir(candidate); err == nil {
			return dir, nil
		}
	}
	return "", fmt.Errorf("workflow directory not found; set %s or add a workflows/bundled directory", envKey)
}

func upwardDirs(start string) []string {
	start = filepath.Clean(start)
	if start == "" || start == "." {
		start = "."
	}
	var out []string
	current := start
	for {
		out = append(out, filepath.Join(current, "workflows", "bundled"))
		out = append(out, filepath.Join(current, "workflows"))
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return out
}

func validateDir(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}
	return path, nil
}
