package schema

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const Version = "workflow.v1"

type Definition struct {
	SchemaVersion string            `yaml:"schema_version" json:"schema_version"`
	WorkflowID    string            `yaml:"workflow_id" json:"workflow_id"`
	Title         string            `yaml:"title" json:"title"`
	Description   string            `yaml:"description" json:"description"`
	Inputs        map[string]Input  `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Steps         []Step            `yaml:"steps" json:"steps"`
	Verification  []Step            `yaml:"verification,omitempty" json:"verification,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Policy        PolicyHints       `yaml:"policy,omitempty" json:"policy,omitempty"`
}

type Input struct {
	Type        string `yaml:"type" json:"type"`
	Required    bool   `yaml:"required,omitempty" json:"required,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Default     string `yaml:"default,omitempty" json:"default,omitempty"`
}

type Step struct {
	ID     string         `yaml:"id" json:"id"`
	Type   string         `yaml:"type" json:"type"`
	Args   map[string]any `yaml:"args,omitempty" json:"args,omitempty"`
	Expect map[string]any `yaml:"expect,omitempty" json:"expect,omitempty"`
}

type PolicyHints struct {
	Requires []string `yaml:"requires,omitempty" json:"requires,omitempty"`
	Notes    []string `yaml:"notes,omitempty" json:"notes,omitempty"`
}

func LoadFile(path string) (Definition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Definition{}, fmt.Errorf("read workflow %s: %w", path, err)
	}
	def, err := LoadBytes(data)
	if err != nil {
		return Definition{}, fmt.Errorf("parse workflow %s: %w", path, err)
	}
	return def, nil
}

func LoadBytes(data []byte) (Definition, error) {
	var def Definition
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&def); err != nil {
		return Definition{}, err
	}
	normalizeDefinition(&def)
	if err := validateDefinition(def); err != nil {
		return Definition{}, err
	}
	return def, nil
}

func normalizeDefinition(def *Definition) {
	if def == nil {
		return
	}
	def.SchemaVersion = strings.TrimSpace(def.SchemaVersion)
	def.WorkflowID = strings.TrimSpace(def.WorkflowID)
	def.Title = strings.TrimSpace(def.Title)
	def.Description = strings.TrimSpace(def.Description)
	if len(def.Metadata) == 0 {
		def.Metadata = nil
	}
	if len(def.Policy.Requires) == 0 {
		def.Policy.Requires = nil
	}
	if len(def.Policy.Notes) == 0 {
		def.Policy.Notes = nil
	}

	inputNames := make([]string, 0, len(def.Inputs))
	for name := range def.Inputs {
		inputNames = append(inputNames, name)
	}
	sort.Strings(inputNames)
	for _, name := range inputNames {
		item := def.Inputs[name]
		item.Type = strings.TrimSpace(item.Type)
		item.Description = strings.TrimSpace(item.Description)
		item.Default = strings.TrimSpace(item.Default)
		def.Inputs[name] = item
	}
	for i := range def.Steps {
		normalizeStep(&def.Steps[i])
	}
	for i := range def.Verification {
		normalizeStep(&def.Verification[i])
	}
}

func normalizeStep(step *Step) {
	step.ID = strings.TrimSpace(step.ID)
	step.Type = strings.TrimSpace(step.Type)
	step.Args = normalizeMap(step.Args)
	step.Expect = normalizeMap(step.Expect)
}

func normalizeMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]any, len(values))
	for key, value := range values {
		k := strings.TrimSpace(key)
		if k == "" {
			continue
		}
		out[k] = normalizeValue(value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return normalizeMap(typed)
	case map[any]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[fmt.Sprint(key)] = normalizeValue(item)
		}
		return normalizeMap(out)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, normalizeValue(item))
		}
		return out
	case string:
		return strings.TrimSpace(typed)
	default:
		return value
	}
}

func validateDefinition(def Definition) error {
	if def.SchemaVersion != Version {
		return fmt.Errorf("schema_version must be %q", Version)
	}
	if def.WorkflowID == "" {
		return fmt.Errorf("missing required field workflow_id")
	}
	if def.Title == "" {
		return fmt.Errorf("missing required field title")
	}
	if def.Description == "" {
		return fmt.Errorf("missing required field description")
	}
	if len(def.Steps) == 0 {
		return fmt.Errorf("workflow %s must define at least one step", def.WorkflowID)
	}
	if len(def.Inputs) > 0 {
		for name, input := range def.Inputs {
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("workflow %s has an empty input name", def.WorkflowID)
			}
			switch input.Type {
			case "string":
			default:
				return fmt.Errorf("workflow %s input %s has unsupported type %q", def.WorkflowID, name, input.Type)
			}
		}
	}
	seen := map[string]string{}
	validateStepSet := func(phase string, steps []Step) error {
		for _, step := range steps {
			if step.ID == "" {
				return fmt.Errorf("workflow %s %s step is missing id", def.WorkflowID, phase)
			}
			if step.Type == "" {
				return fmt.Errorf("workflow %s step %s is missing type", def.WorkflowID, step.ID)
			}
			if existing, ok := seen[step.ID]; ok {
				return fmt.Errorf("workflow %s step id %s is duplicated in %s and %s", def.WorkflowID, step.ID, existing, phase)
			}
			seen[step.ID] = phase
		}
		return nil
	}
	if err := validateStepSet("steps", def.Steps); err != nil {
		return err
	}
	if err := validateStepSet("verification", def.Verification); err != nil {
		return err
	}
	if len(def.Verification) == 0 {
		hasExpectation := false
		for _, step := range def.Steps {
			if len(step.Expect) > 0 {
				hasExpectation = true
				break
			}
		}
		if !hasExpectation {
			return fmt.Errorf("workflow %s must define verification steps or at least one step expectation", def.WorkflowID)
		}
	}
	return nil
}
