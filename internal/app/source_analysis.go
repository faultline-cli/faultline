package app

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"faultline/internal/detectors"
	"faultline/internal/engine"
	"faultline/internal/repo"
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

func sourceEngineOptions(opts AnalyzeOptions) engine.Options {
	return engine.Options{
		PlaybookDir:      opts.PlaybookDir,
		PlaybookPackDirs: opts.PlaybookPackDirs,
		GitSince:         opts.GitSince,
		RepoPath:         opts.RepoPath,
		BayesEnabled:     opts.BayesEnabled,
	}
}
