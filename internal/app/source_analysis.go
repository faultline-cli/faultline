package app

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"faultline/internal/detectors"
	"faultline/internal/engine"
	"faultline/internal/model"
	"faultline/internal/output"
	"faultline/internal/repo"
)

var ErrGuardFindings = errors.New("guard findings emitted")

// guardMinConfidence and guardMinScore are the thresholds used by the guard
// command to filter source-detector results down to high-confidence findings
// only. Lower values increase noise; higher values reduce recall.
const (
	guardMinConfidence = 0.75
	guardMinScore      = 3.5
)

// Inspect scans a repository tree with source-detector playbooks.
func (Service) Inspect(ctx context.Context, root string, opts AnalyzeOptions, w io.Writer) error {
	changeSet := detectors.ChangeSet{}
	if scanner, err := repo.NewScanner(root); err == nil {
		if loaded, loadErr := repo.LoadWorktreeChangeSet(scanner); loadErr != nil {
			return loadErr
		} else {
			absRoot, absErr := filepath.Abs(root)
			if absErr != nil {
				return absErr
			}
			prefix, relErr := filepath.Rel(scanner.Root, absRoot)
			if relErr != nil {
				return relErr
			}
			changeSet = repo.ChangeSetRelativeTo(loaded, prefix)
		}
	}
	a, err := engine.New(sourceEngineOptions(opts)).AnalyzeRepository(root, changeSet)
	if errors.Is(err, engine.ErrNoInput) {
		return err
	}
	if err != nil && !errors.Is(err, engine.ErrNoMatch) {
		return err
	}
	a, prepErr := prepareAnalysisWithStore(ctx, a, "", "repository", "inspect", opts, true)
	if prepErr != nil {
		return prepErr
	}
	return writeAnalysis(a, opts, w)
}

// Guard inspects changed repository files and only emits high-confidence findings.
func (Service) Guard(root string, opts AnalyzeOptions, w io.Writer) error {
	scanner, err := repo.NewScanner(root)
	if err != nil {
		return writeGuardNoFindings(root, opts, w)
	}
	changeSet, err := repo.LoadWorktreeChangeSet(scanner)
	if err != nil {
		return err
	}
	if len(changeSet.ChangedFiles) == 0 {
		return writeGuardNoFindings(scanner.Root, opts, w)
	}

	guardOpts := opts
	guardOpts.RepoPath = scanner.Root
	guardOpts.BayesEnabled = true
	a, err := engine.New(sourceEngineOptions(guardOpts)).AnalyzeRepository(scanner.Root, changeSet)
	if errors.Is(err, engine.ErrNoInput) || errors.Is(err, engine.ErrNoMatch) {
		return writeGuardNoFindings(scanner.Root, opts, w)
	}
	if err != nil {
		return err
	}

	filtered := guardFindings(a, opts.Top)
	if len(filtered.Results) == 0 {
		return writeGuardNoFindings(scanner.Root, opts, w)
	}
	if err := writeAnalysis(filtered, opts, w); err != nil {
		return err
	}
	return ErrGuardFindings
}

func sourceEngineOptions(opts AnalyzeOptions) engine.Options {
	return engine.Options{
		PlaybookDir:      opts.PlaybookDir,
		PlaybookPackDirs: opts.PlaybookPackDirs,
		GitSince:         opts.GitSince,
		RepoPath:         opts.RepoPath,
		BayesEnabled:     opts.BayesEnabled,
	}
}

func guardFindings(a *model.Analysis, top int) *model.Analysis {
	if a == nil {
		return &model.Analysis{Results: []model.Result{}}
	}
	filtered := make([]model.Result, 0, len(a.Results))
	for _, result := range a.Results {
		if result.Confidence < guardMinConfidence {
			continue
		}
		if result.Score < guardMinScore {
			continue
		}
		filtered = append(filtered, result)
	}
	if top > 0 && len(filtered) > top {
		filtered = filtered[:top]
	}
	return &model.Analysis{
		Results:     filtered,
		Context:     a.Context,
		Fingerprint: a.Fingerprint,
		Source:      a.Source,
		RepoContext: a.RepoContext,
		Delta:       a.Delta,
	}
}

func writeGuardNoFindings(root string, opts AnalyzeOptions, w io.Writer) error {
	if opts.JSON || opts.Format == output.FormatJSON {
		return writeAnalysis(&model.Analysis{
			Results: []model.Result{},
			Source:  root,
		}, opts, w)
	}
	return nil
}
