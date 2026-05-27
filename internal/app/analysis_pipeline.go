package app

import (
	"bytes"
	"context"
	"errors"
	"io"

	"faultline/internal/engine"
	"faultline/internal/model"
)

type loadedAnalysisInput struct {
	Analysis  *model.Analysis
	Lines     []model.Line
	Playbooks []model.Playbook
}

func analyzeLog(ctx context.Context, r io.Reader, source string, opts AnalyzeOptions, surface string, persist bool) (*model.Analysis, error) {
	return analyzeLogWithEngine(ctx, engine.New(logEngineOptions(opts)), r, source, opts, surface, persist)
}

func analyzeLogWithEngine(ctx context.Context, eng *engine.Engine, r io.Reader, source string, opts AnalyzeOptions, surface string, persist bool) (*model.Analysis, error) {
	data, err := engine.ReadLogInput(r)
	if err != nil {
		return nil, err
	}
	a, err := eng.AnalyzeReader(bytes.NewReader(data))
	if a != nil {
		a.Source = source
	}
	if a != nil || errors.Is(err, engine.ErrNoMatch) {
		var prepErr error
		a, prepErr = prepareAnalysisWithStore(ctx, a, string(data), "log", surface, opts, persist)
		if prepErr != nil {
			return nil, prepErr
		}
	}
	return a, err
}

func loadAnalysisInput(ctx context.Context, r io.Reader, source string, opts AnalyzeOptions) (loadedAnalysisInput, error) {
	data, err := engine.ReadLogInput(r)
	if err != nil {
		return loadedAnalysisInput{}, err
	}
	lines, err := engine.ReadLines(bytes.NewReader(data))
	if err != nil {
		return loadedAnalysisInput{}, err
	}
	eng := engine.New(logEngineOptions(opts))
	pbs, err := eng.Playbooks()
	if err != nil {
		return loadedAnalysisInput{}, err
	}
	analysis, err := analyzeLogWithEngine(ctx, eng, bytes.NewReader(data), source, opts, "trace", false)
	return loadedAnalysisInput{
		Analysis:  analysis,
		Lines:     lines,
		Playbooks: pbs,
	}, err
}

func logEngineOptions(opts AnalyzeOptions) engine.Options {
	return engine.Options{
		PlaybookDir:       opts.PlaybookDir,
		PlaybookPackDirs:  opts.PlaybookPackDirs,
		GitContextEnabled: opts.GitContextEnabled,
		GitSince:          opts.GitSince,
		RepoPath:          opts.RepoPath,
		BayesEnabled:      opts.BayesEnabled,
	}
}
