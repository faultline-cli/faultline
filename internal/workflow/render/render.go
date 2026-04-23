package render

import (
	"fmt"
	"strings"

	"faultline/internal/model"
	workflowpkg "faultline/internal/workflow"
)

func PlanText(doc workflowpkg.PlanDocument) string {
	var b strings.Builder
	fmt.Fprintf(&b, "WORKFLOW  %s  [%s · %s]\n", doc.WorkflowID, doc.Mode, doc.DefinitionVersion)
	if doc.Title != "" {
		fmt.Fprintf(&b, "Title: %s\n", doc.Title)
	}
	if doc.Description != "" {
		fmt.Fprintf(&b, "Description: %s\n", doc.Description)
	}
	if doc.SourceFailureID != "" {
		fmt.Fprintf(&b, "Source failure: %s\n", doc.SourceFailureID)
	}
	if doc.SourceFingerprint != "" {
		fmt.Fprintf(&b, "Artifact fingerprint: %s\n", doc.SourceFingerprint)
	}
	if len(doc.ResolvedInputs) > 0 {
		fmt.Fprintln(&b, "Resolved inputs:")
		for _, line := range orderedKeyValues(doc.ResolvedInputs) {
			fmt.Fprintf(&b, "  - %s\n", line)
		}
	}
	if len(doc.RequiredSafety) > 0 {
		fmt.Fprintln(&b, "Required policy:")
		for _, item := range doc.RequiredSafety {
			fmt.Fprintf(&b, "  - %s\n", item)
		}
	}
	if len(doc.PolicyNotes) > 0 {
		fmt.Fprintln(&b, "Policy notes:")
		for _, item := range doc.PolicyNotes {
			fmt.Fprintf(&b, "  - %s\n", item)
		}
	}
	if len(doc.Steps) > 0 {
		fmt.Fprintln(&b, "Steps:")
		writePlanSteps(&b, doc.Steps)
	}
	if len(doc.Verification) > 0 {
		fmt.Fprintln(&b, "Verification:")
		writePlanSteps(&b, doc.Verification)
	}
	return b.String()
}

func ExecutionText(record *model.WorkflowExecutionRecord) string {
	if record == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "WORKFLOW EXECUTION  %s  [%s]\n", record.ExecutionID, record.WorkflowID)
	fmt.Fprintf(&b, "Status: %s\n", record.Status)
	fmt.Fprintf(&b, "Verification: %s\n", record.VerificationStatus)
	if record.SourceFailureID != "" {
		fmt.Fprintf(&b, "Source failure: %s\n", record.SourceFailureID)
	}
	if record.SourceFingerprint != "" {
		fmt.Fprintf(&b, "Artifact fingerprint: %s\n", record.SourceFingerprint)
	}
	if len(record.ResolvedInputs) > 0 {
		fmt.Fprintln(&b, "Resolved inputs:")
		for _, line := range orderedKeyValues(record.ResolvedInputs) {
			fmt.Fprintf(&b, "  - %s\n", line)
		}
	}
	if len(record.StepResults) > 0 {
		fmt.Fprintln(&b, "Step results:")
		for i, item := range record.StepResults {
			fmt.Fprintf(&b, "  %d. [%s] %s (%s)\n", i+1, item.Status, item.StepID, item.StepType)
			if item.Message != "" {
				fmt.Fprintf(&b, "     message: %s\n", item.Message)
			}
			if item.Error != "" {
				fmt.Fprintf(&b, "     error: %s\n", item.Error)
			}
			if len(item.Outputs) > 0 {
				fmt.Fprintf(&b, "     outputs: %s\n", strings.Join(orderedKeyValues(item.Outputs), ", "))
			}
		}
	}
	return b.String()
}

func HistoryText(items []model.WorkflowExecutionSummary) string {
	var b strings.Builder
	if len(items) == 0 {
		return "No workflow executions recorded.\n"
	}
	fmt.Fprintln(&b, "WORKFLOW HISTORY")
	for _, item := range items {
		fmt.Fprintf(&b, "  - %s  %s  [%s · %s]\n", item.ExecutionID, item.WorkflowID, item.Status, item.VerificationStatus)
		if item.SourceFailureID != "" {
			fmt.Fprintf(&b, "    source: %s\n", item.SourceFailureID)
		}
		if item.FinishedAt != "" {
			fmt.Fprintf(&b, "    finished: %s\n", item.FinishedAt)
		}
	}
	return b.String()
}

func writePlanSteps(b *strings.Builder, steps []workflowpkg.PlanStep) {
	for i, item := range steps {
		fmt.Fprintf(b, "  %d. [%s] %s (%s)\n", i+1, item.SafetyClass, item.ID, item.Type)
		fmt.Fprintf(b, "     %s\n", item.Description)
		if len(item.Args) > 0 {
			fmt.Fprintf(b, "     args: %s\n", strings.Join(orderedAnyValues(item.Args), ", "))
		}
		if len(item.Expect) > 0 {
			fmt.Fprintf(b, "     expect: %s\n", strings.Join(orderedAnyValues(item.Expect), ", "))
		}
	}
}

func orderedKeyValues(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sortStrings(keys)
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, fmt.Sprintf("%s=%s", key, values[key]))
	}
	return out
}

func orderedAnyValues(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sortStrings(keys)
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, fmt.Sprintf("%s=%v", key, values[key]))
	}
	return out
}

func sortStrings(values []string) {
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[j] < values[i] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}
