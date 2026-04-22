package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"faultline/internal/workflow"
)

// workflowJSON is the stable schema emitted by FormatWorkflowJSON.
type workflowJSON struct {
	SchemaVersion string   `json:"schema_version"`
	Mode          string   `json:"mode"`
	FailureID     string   `json:"failure_id,omitempty"`
	Title         string   `json:"title,omitempty"`
	Source        string   `json:"source,omitempty"`
	Context       ctxJSON  `json:"context"`
	Evidence      []string `json:"evidence"`
	Files         []string `json:"files,omitempty"`
	LocalRepro    []string `json:"local_repro,omitempty"`
	Verify        []string `json:"verify,omitempty"`
	RankingHints  []string `json:"ranking_hints,omitempty"`
	DeltaHints    []string `json:"delta_hints,omitempty"`
	MetricsHints  []string `json:"metrics_hints,omitempty"`
	PolicyHints   []string `json:"policy_hints,omitempty"`
	Steps         []string `json:"steps"`
	AgentPrompt   string   `json:"agent_prompt,omitempty"`
}

// FormatWorkflowText formats a deterministic workflow follow-up plan as text.
func FormatWorkflowText(plan workflow.Plan) string {
	var b strings.Builder
	if plan.FailureID == "" {
		fmt.Fprintln(&b, "WORKFLOW")
		for i, step := range plan.Steps {
			fmt.Fprintf(&b, "  %d. %s\n", i+1, step)
		}
		return b.String()
	}

	fmt.Fprintf(&b, "WORKFLOW  %s · %s  [%s · %s]\n", plan.FailureID, plan.Title, plan.Mode, plan.SchemaVersion)
	if plan.Source != "" {
		fmt.Fprintf(&b, "Source: %s\n", plan.Source)
	}
	if plan.Context.Stage != "" {
		fmt.Fprintf(&b, "Stage: %s\n", plan.Context.Stage)
	}
	if plan.Context.CommandHint != "" {
		fmt.Fprintf(&b, "Command: %s\n", plan.Context.CommandHint)
	}
	if plan.Context.Step != "" {
		fmt.Fprintf(&b, "Step: %s\n", plan.Context.Step)
	}
	if len(plan.Evidence) > 0 {
		fmt.Fprintln(&b, "Evidence:")
		for _, line := range plan.Evidence {
			fmt.Fprintf(&b, "  - %s\n", line)
		}
	}
	if len(plan.Files) > 0 {
		fmt.Fprintln(&b, "Focus files:")
		for _, file := range plan.Files {
			fmt.Fprintf(&b, "  - %s\n", file)
		}
	}
	if len(plan.LocalRepro) > 0 {
		fmt.Fprintln(&b, "Local repro:")
		for _, cmd := range plan.LocalRepro {
			fmt.Fprintf(&b, "  - %s\n", cmd)
		}
	}
	if len(plan.Verify) > 0 {
		fmt.Fprintln(&b, "Verify:")
		for _, cmd := range plan.Verify {
			fmt.Fprintf(&b, "  - %s\n", cmd)
		}
	}
	if len(plan.RankingHints) > 0 {
		fmt.Fprintln(&b, "Ranking hints:")
		for _, item := range plan.RankingHints {
			fmt.Fprintf(&b, "  - %s\n", item)
		}
	}
	if len(plan.DeltaHints) > 0 {
		fmt.Fprintln(&b, "Delta hints:")
		for _, item := range plan.DeltaHints {
			fmt.Fprintf(&b, "  - %s\n", item)
		}
	}
	if len(plan.MetricsHints) > 0 {
		fmt.Fprintln(&b, "Metrics:")
		for _, item := range plan.MetricsHints {
			fmt.Fprintf(&b, "  - %s\n", item)
		}
	}
	if len(plan.PolicyHints) > 0 {
		fmt.Fprintln(&b, "Policy:")
		for _, item := range plan.PolicyHints {
			fmt.Fprintf(&b, "  - %s\n", item)
		}
	}
	fmt.Fprintln(&b, "Next steps:")
	for i, step := range plan.Steps {
		fmt.Fprintf(&b, "  %d. %s\n", i+1, step)
	}
	if plan.AgentPrompt != "" {
		fmt.Fprintln(&b, "\nAgent prompt:")
		fmt.Fprintln(&b, plan.AgentPrompt)
	}
	return b.String()
}

// FormatWorkflowJSON serializes a workflow plan as stable JSON.
func FormatWorkflowJSON(plan workflow.Plan) (string, error) {
	payload := workflowJSON{
		SchemaVersion: plan.SchemaVersion,
		Mode:          string(plan.Mode),
		FailureID:     plan.FailureID,
		Title:         plan.Title,
		Source:        plan.Source,
		Context: ctxJSON{
			Stage:       plan.Context.Stage,
			CommandHint: plan.Context.CommandHint,
			Step:        plan.Context.Step,
		},
		Evidence:     plan.Evidence,
		Files:        plan.Files,
		LocalRepro:   plan.LocalRepro,
		Verify:       plan.Verify,
		RankingHints: plan.RankingHints,
		DeltaHints:   plan.DeltaHints,
		MetricsHints: plan.MetricsHints,
		PolicyHints:  plan.PolicyHints,
		Steps:        plan.Steps,
		AgentPrompt:  plan.AgentPrompt,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal workflow JSON: %w", err)
	}
	return string(data) + "\n", nil
}
