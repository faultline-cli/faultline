package engine

import (
	"faultline/internal/detectors"
	"faultline/internal/model"
	"faultline/internal/playbooks"
)

type playbookCatalog interface {
	Load() ([]model.Playbook, error)
	List() ([]model.Playbook, error)
	Explain(id string) (model.Playbook, error)
}

type repoSnapshotLoader interface {
	Load(repoPath string, changeSet detectors.ChangeSet) *repoSnapshot
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
