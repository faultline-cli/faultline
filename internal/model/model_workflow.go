package model

type WorkflowExecutionMode string

const (
	WorkflowExecutionModeExplain WorkflowExecutionMode = "explain"
	WorkflowExecutionModeDryRun  WorkflowExecutionMode = "dry-run"
	WorkflowExecutionModeApply   WorkflowExecutionMode = "apply"
)

type WorkflowExecutionStatus string

const (
	WorkflowExecutionStatusPlanned   WorkflowExecutionStatus = "planned"
	WorkflowExecutionStatusSucceeded WorkflowExecutionStatus = "succeeded"
	WorkflowExecutionStatusFailed    WorkflowExecutionStatus = "failed"
	WorkflowExecutionStatusBlocked   WorkflowExecutionStatus = "blocked"
	WorkflowExecutionStatusSkipped   WorkflowExecutionStatus = "skipped"
)

type WorkflowVerificationStatus string

const (
	WorkflowVerificationStatusPending WorkflowVerificationStatus = "pending"
	WorkflowVerificationStatusPassed  WorkflowVerificationStatus = "passed"
	WorkflowVerificationStatusFailed  WorkflowVerificationStatus = "failed"
)

type WorkflowStepResult struct {
	Phase              string                     `json:"phase,omitempty"`
	StepID             string                     `json:"step_id,omitempty"`
	StepType           string                     `json:"step_type,omitempty"`
	SafetyClass        string                     `json:"safety_class,omitempty"`
	Status             WorkflowExecutionStatus    `json:"status,omitempty"`
	VerificationStatus WorkflowVerificationStatus `json:"verification_status,omitempty"`
	StartedAt          string                     `json:"started_at,omitempty"`
	FinishedAt         string                     `json:"finished_at,omitempty"`
	Changed            *bool                      `json:"changed,omitempty"`
	Message            string                     `json:"message,omitempty"`
	Outputs            map[string]string          `json:"outputs,omitempty"`
	Error              string                     `json:"error,omitempty"`
}

type WorkflowExecutionRecord struct {
	SchemaVersion      string                     `json:"schema_version,omitempty"`
	ExecutionID        string                     `json:"execution_id,omitempty"`
	WorkflowID         string                     `json:"workflow_id,omitempty"`
	Title              string                     `json:"title,omitempty"`
	Mode               WorkflowExecutionMode      `json:"mode,omitempty"`
	SourceFingerprint  string                     `json:"source_fingerprint,omitempty"`
	SourceFailureID    string                     `json:"source_failure_id,omitempty"`
	StartedAt          string                     `json:"started_at,omitempty"`
	FinishedAt         string                     `json:"finished_at,omitempty"`
	ResolvedInputs     map[string]string          `json:"resolved_inputs,omitempty"`
	StepResults        []WorkflowStepResult       `json:"step_results,omitempty"`
	VerificationStatus WorkflowVerificationStatus `json:"verification_status,omitempty"`
	Status             WorkflowExecutionStatus    `json:"status,omitempty"`
}

type WorkflowExecutionSummary struct {
	ExecutionID        string                     `json:"execution_id,omitempty"`
	WorkflowID         string                     `json:"workflow_id,omitempty"`
	Title              string                     `json:"title,omitempty"`
	Mode               WorkflowExecutionMode      `json:"mode,omitempty"`
	SourceFingerprint  string                     `json:"source_fingerprint,omitempty"`
	SourceFailureID    string                     `json:"source_failure_id,omitempty"`
	StartedAt          string                     `json:"started_at,omitempty"`
	FinishedAt         string                     `json:"finished_at,omitempty"`
	VerificationStatus WorkflowVerificationStatus `json:"verification_status,omitempty"`
	Status             WorkflowExecutionStatus    `json:"status,omitempty"`
}
