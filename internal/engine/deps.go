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

type repoEnricher interface {
	Enrich(result model.Result) *model.RepoContext
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
