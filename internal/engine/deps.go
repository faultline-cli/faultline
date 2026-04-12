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

type historyRecorder interface {
	CountSeen(failureID string) int
	Record(result model.Result)
}

type repoEnricher interface {
	Enrich(result model.Result) *model.RepoContext
}

type sourceLoader interface {
	Load(root string) ([]detectors.SourceFile, error)
}

type defaultHistoryRecorder struct{}

func (defaultHistoryRecorder) CountSeen(failureID string) int {
	return countSeen(failureID)
}

func (defaultHistoryRecorder) Record(result model.Result) {
	recordResult(result)
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
