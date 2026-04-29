package engine

import (
	"context"

	"faultline/internal/detectors"
	enginedelta "faultline/internal/engine/delta"
	"faultline/internal/model"
	"faultline/internal/playbooks"
	"faultline/internal/scoring"
)

type playbookCatalog interface {
	Load() ([]model.Playbook, error)
	List() ([]model.Playbook, error)
	Explain(id string) (model.Playbook, error)
}

type repoEnricher interface {
	Enrich(result model.Result) *model.RepoContext
}

type repoSnapshotLoader interface {
	Load(repoPath string, changeSet detectors.ChangeSet) *repoSnapshot
}

type providerDeltaLoader interface {
	Load(currentLog string) *scoring.RepoState
}

type detectorRegistry interface {
	MustLookup(detectors.Kind) (detectors.Detector, error)
}

type sourceLoader interface {
	Load(root string) ([]detectors.SourceFile, error)
}

type defaultSourceLoader struct{}

func (defaultSourceLoader) Load(root string) ([]detectors.SourceFile, error) {
	return loadSourceFiles(root)
}

func newCatalog(dir string, extra []string) playbookCatalog {
	return playbooks.NewCatalogWithOptions(playbooks.CatalogOptions{
		OverrideDir:   dir,
		ExtraPackDirs: extra,
	})
}

type localRepoSnapshotLoader struct {
	engine *Engine
}

func (l localRepoSnapshotLoader) Load(repoPath string, changeSet detectors.ChangeSet) *repoSnapshot {
	return l.engine.loadRepoSnapshotFromPath(repoPath, changeSet)
}

type providerDeltaResolver struct {
	opts Options
}

func (p providerDeltaResolver) Load(currentLog string) *scoring.RepoState {
	if p.opts.DeltaProvider == "" {
		return nil
	}
	resolver := enginedelta.NewResolver(nil)
	snapshot, err := resolver.Resolve(context.Background(), enginedelta.Options{
		Provider: p.opts.DeltaProvider,
		GitHub: enginedelta.GitHubOptions{
			Repository: p.opts.GitHubRepository,
			Branch:     p.opts.GitHubBranch,
			RunID:      p.opts.GitHubRunID,
			Token:      p.opts.GitHubToken,
		},
		GitLab: enginedelta.GitLabOptions{
			Project:    p.opts.GitLabProject,
			Branch:     p.opts.GitLabBranch,
			PipelineID: p.opts.GitLabPipelineID,
			JobID:      p.opts.GitLabJobID,
			Token:      p.opts.GitLabToken,
			APIBaseURL: p.opts.GitLabAPIBaseURL,
		},
	}, currentLog)
	if err != nil || snapshot == nil {
		return nil
	}
	return &scoring.RepoState{
		Provider:          snapshot.Provider,
		ChangedFiles:      append([]string(nil), snapshot.FilesChanged...),
		TestsNewlyFailing: append([]string(nil), snapshot.TestsNewlyFailing...),
		ErrorsAdded:       append([]string(nil), snapshot.ErrorsAdded...),
		EnvDiff:           cloneDeltaEnvDiff(snapshot.EnvDiff),
	}
}
