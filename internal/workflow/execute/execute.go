package execute

import (
	"context"
	"errors"
	"fmt"
	"time"

	"faultline/internal/model"
	"faultline/internal/workflow/bind"
	"faultline/internal/workflow/plan"
	"faultline/internal/workflow/registry"
)

type Policy struct {
	AllowLocalMutation       bool
	AllowEnvironmentMutation bool
	AllowExternalSideEffect  bool
}

type Options struct {
	Runtime  bind.RuntimeContext
	Artifact *model.FailureArtifact
	Policy   Policy
	Now      func() time.Time
}

func (p Policy) Allows(class registry.SafetyClass) bool {
	switch class {
	case registry.SafetyClassRead:
		return true
	case registry.SafetyClassLocalMutation:
		return p.AllowLocalMutation
	case registry.SafetyClassEnvironmentMutation:
		return p.AllowEnvironmentMutation
	case registry.SafetyClassExternalSideEffect:
		return p.AllowExternalSideEffect
	default:
		return false
	}
}

func Apply(ctx context.Context, compiled plan.Plan, opts Options) (*model.WorkflowExecutionRecord, error) {
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	started := now().UTC()
	record := &model.WorkflowExecutionRecord{
		SchemaVersion:      "workflow_execution.v1",
		WorkflowID:         compiled.WorkflowID,
		Title:              compiled.Title,
		Mode:               model.WorkflowExecutionModeApply,
		SourceFingerprint:  compiled.SourceFingerprint,
		SourceFailureID:    compiled.SourceFailureID,
		StartedAt:          started.Format(time.RFC3339),
		ResolvedInputs:     compiled.ResolvedInputs,
		Status:             model.WorkflowExecutionStatusSucceeded,
		VerificationStatus: model.WorkflowVerificationStatusPending,
	}
	state := bind.State{
		Inputs:   compiled.ResolvedInputs,
		Runtime:  opts.Runtime,
		Artifact: opts.Artifact,
		Steps:    map[string]map[string]string{},
	}
	for _, item := range append(append([]plan.Step{}, compiled.Steps...), compiled.Verification...) {
		stepResult := model.WorkflowStepResult{
			Phase:       item.Phase,
			StepID:      item.ID,
			StepType:    item.Type,
			SafetyClass: string(item.SafetyClass),
			StartedAt:   now().UTC().Format(time.RFC3339),
		}
		if !opts.Policy.Allows(item.SafetyClass) {
			stepResult.Status = model.WorkflowExecutionStatusBlocked
			stepResult.Error = fmt.Sprintf("policy blocks safety class %s", item.SafetyClass)
			stepResult.FinishedAt = now().UTC().Format(time.RFC3339)
			record.StepResults = append(record.StepResults, stepResult)
			record.Status = model.WorkflowExecutionStatusBlocked
			record.VerificationStatus = model.WorkflowVerificationStatusFailed
			record.FinishedAt = stepResult.FinishedAt
			return record, errors.New(stepResult.Error)
		}
		resolvedArgsAny, err := bind.ResolveValue(item.ResolvedArgs, state, false)
		if err != nil {
			stepResult.Status = model.WorkflowExecutionStatusFailed
			stepResult.Error = err.Error()
			stepResult.FinishedAt = now().UTC().Format(time.RFC3339)
			record.StepResults = append(record.StepResults, stepResult)
			record.Status = model.WorkflowExecutionStatusFailed
			record.VerificationStatus = model.WorkflowVerificationStatusFailed
			record.FinishedAt = stepResult.FinishedAt
			return record, err
		}
		decodedArgs, err := item.Handler.DecodeArgs(mapFrom(resolvedArgsAny))
		if err != nil {
			stepResult.Status = model.WorkflowExecutionStatusFailed
			stepResult.Error = err.Error()
			stepResult.FinishedAt = now().UTC().Format(time.RFC3339)
			record.StepResults = append(record.StepResults, stepResult)
			record.Status = model.WorkflowExecutionStatusFailed
			record.VerificationStatus = model.WorkflowVerificationStatusFailed
			record.FinishedAt = stepResult.FinishedAt
			return record, err
		}
		resolvedExpectAny, err := bind.ResolveValue(item.ResolvedExpect, state, false)
		if err != nil {
			stepResult.Status = model.WorkflowExecutionStatusFailed
			stepResult.Error = err.Error()
			stepResult.FinishedAt = now().UTC().Format(time.RFC3339)
			record.StepResults = append(record.StepResults, stepResult)
			record.Status = model.WorkflowExecutionStatusFailed
			record.VerificationStatus = model.WorkflowVerificationStatusFailed
			record.FinishedAt = stepResult.FinishedAt
			return record, err
		}
		decodedExpect, err := item.Handler.DecodeExpect(mapFrom(resolvedExpectAny))
		if err != nil {
			stepResult.Status = model.WorkflowExecutionStatusFailed
			stepResult.Error = err.Error()
			stepResult.FinishedAt = now().UTC().Format(time.RFC3339)
			record.StepResults = append(record.StepResults, stepResult)
			record.Status = model.WorkflowExecutionStatusFailed
			record.VerificationStatus = model.WorkflowVerificationStatusFailed
			record.FinishedAt = stepResult.FinishedAt
			return record, err
		}
		result, err := item.Handler.Execute(ctx, registry.Runtime{WorkDir: opts.Runtime.WorkDir}, decodedArgs)
		stepResult.FinishedAt = now().UTC().Format(time.RFC3339)
		stepResult.Outputs = result.Outputs
		stepResult.Message = result.Message
		stepResult.Changed = result.Changed
		if err != nil {
			stepResult.Status = model.WorkflowExecutionStatusFailed
			stepResult.Error = err.Error()
			record.StepResults = append(record.StepResults, stepResult)
			record.Status = model.WorkflowExecutionStatusFailed
			record.VerificationStatus = model.WorkflowVerificationStatusFailed
			record.FinishedAt = stepResult.FinishedAt
			return record, err
		}
		if verr := item.Handler.Verify(decodedExpect, result); verr != nil {
			stepResult.Status = model.WorkflowExecutionStatusFailed
			stepResult.VerificationStatus = model.WorkflowVerificationStatusFailed
			stepResult.Error = verr.Error()
			record.StepResults = append(record.StepResults, stepResult)
			record.Status = model.WorkflowExecutionStatusFailed
			record.VerificationStatus = model.WorkflowVerificationStatusFailed
			record.FinishedAt = stepResult.FinishedAt
			return record, verr
		}
		stepResult.Status = model.WorkflowExecutionStatusSucceeded
		if item.Phase == "verification" || item.DecodedExpect != nil {
			stepResult.VerificationStatus = model.WorkflowVerificationStatusPassed
		}
		record.StepResults = append(record.StepResults, stepResult)
		state.Steps[item.ID] = result.Outputs
	}
	record.FinishedAt = now().UTC().Format(time.RFC3339)
	record.VerificationStatus = model.WorkflowVerificationStatusPassed
	return record, nil
}

func mapFrom(value any) map[string]any {
	if value == nil {
		return nil
	}
	typed, _ := value.(map[string]any)
	return typed
}
