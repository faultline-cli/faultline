package app

import (
	"errors"
	"fmt"
	"io"

	"faultline/internal/detectors"
	"faultline/internal/engine"
	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/renderer"
	"faultline/internal/workflow"
)

// Service owns app-level orchestration for CLI commands.
type Service struct{}

// NewService returns the default CLI application service.
func NewService() Service {
	return Service{}
}

// Analyze performs log analysis and writes formatted output to w.
func (Service) Analyze(r io.Reader, source string, opts AnalyzeOptions, w io.Writer) error {
	a, err := analyzeLog(r, source, opts)
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}
	return writeAnalysis(a, opts, w)
}

// Fix performs log analysis and writes only the ranked fix steps to w.
func (Service) Fix(r io.Reader, source string, opts AnalyzeOptions, w io.Writer) error {
	a, err := analyzeLog(r, source, opts)
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}
	_, werr := fmt.Fprint(w, output.FormatFix(a, renderer.DetectOptions(w)))
	return werr
}

// List loads all playbooks and writes a formatted list to w.
func (Service) List(category, playbookDir string, playbookPacks []string, w io.Writer) error {
	pbs, err := engine.New(engine.Options{
		PlaybookDir:      playbookDir,
		PlaybookPackDirs: playbookPacks,
		NoHistory:        true,
	}).List()
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, output.FormatPlaybookList(pbs, category, renderer.DetectOptions(w)))
	return err
}

// Explain fetches a single playbook by id and writes its details to w.
func (Service) Explain(id, playbookDir string, playbookPacks []string, w io.Writer) error {
	pb, err := engine.New(engine.Options{
		PlaybookDir:      playbookDir,
		PlaybookPackDirs: playbookPacks,
		NoHistory:        true,
	}).Explain(id)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, output.FormatPlaybookDetails(pb, renderer.DetectOptions(w)))
	return err
}

// Workflow analyzes the log and emits a deterministic follow-up workflow.
func (Service) Workflow(r io.Reader, source string, opts AnalyzeOptions, mode workflow.Mode, jsonOut bool, w io.Writer) error {
	a, err := analyzeLog(r, source, opts)
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}

	plan := workflow.BuildWithOptions(a, mode, workflow.BuildOptions{
		RepoPath: opts.RepoPath,
	})
	if jsonOut {
		data, err := output.FormatWorkflowJSON(plan)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, data)
		return err
	}

	_, err = fmt.Fprint(w, output.FormatWorkflowText(plan))
	return err
}

// Inspect scans a repository tree with source-detector playbooks.
func (Service) Inspect(root string, opts AnalyzeOptions, w io.Writer) error {
	a, err := engine.New(engine.Options{
		PlaybookDir:      opts.PlaybookDir,
		PlaybookPackDirs: opts.PlaybookPackDirs,
		NoHistory:        opts.NoHistory,
	}).AnalyzeRepository(root, detectors.ChangeSet{})
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}
	return writeAnalysis(a, opts, w)
}

func analyzeLog(r io.Reader, source string, opts AnalyzeOptions) (*model.Analysis, error) {
	a, err := engine.New(engine.Options{
		PlaybookDir:       opts.PlaybookDir,
		PlaybookPackDirs:  opts.PlaybookPackDirs,
		NoHistory:         opts.NoHistory,
		GitContextEnabled: opts.GitContextEnabled,
		GitSince:          opts.GitSince,
		RepoPath:          opts.RepoPath,
	}).AnalyzeReader(r)
	if a != nil {
		a.Source = source
	}
	return a, err
}
