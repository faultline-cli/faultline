package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	artifactpkg "faultline/internal/artifact"
	"faultline/internal/engine"
	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/store"
)

func loadWorkflowAnalysis(r io.Reader, source string, opts AnalyzeOptions) (*model.Analysis, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read workflow input: %w", err)
	}
	if analysis, parseErr := output.ParseAnalysisJSON(data); parseErr == nil && analysis != nil {
		return artifactpkg.Sync(analysis), nil
	}
	a, err := analyzeLog(bytes.NewReader(data), source, opts, "workflow", false)
	if err != nil && !errors.Is(err, engine.ErrNoMatch) && !errors.Is(err, engine.ErrNoInput) {
		return nil, err
	}
	if err != nil && errors.Is(err, engine.ErrNoInput) {
		return nil, err
	}
	return artifactpkg.Sync(a), nil
}

func persistWorkflowExecution(record *model.WorkflowExecutionRecord, opts AnalyzeOptions) (*model.WorkflowExecutionRecord, error) {
	if record == nil {
		return nil, nil
	}
	st, _, err := openWorkflowStore(opts)
	if err != nil {
		return nil, err
	}
	defer st.Close()
	return st.RecordWorkflowExecution(context.Background(), record)
}

func loadWorkflowExecution(executionID string, opts AnalyzeOptions) (*model.WorkflowExecutionRecord, error) {
	st, _, err := openWorkflowStore(opts)
	if err != nil {
		return nil, err
	}
	defer st.Close()
	return st.GetWorkflowExecution(context.Background(), executionID)
}

func openWorkflowStore(opts AnalyzeOptions) (store.Store, store.Info, error) {
	cfg, err := store.ResolveConfig(opts.Store, opts.NoHistory)
	if err != nil {
		return nil, store.Info{}, err
	}
	st, info, err := store.OpenBestEffort(cfg)
	if err != nil {
		return nil, store.Info{}, err
	}
	return st, info, nil
}
