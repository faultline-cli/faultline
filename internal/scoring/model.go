package scoring

import (
	"faultline/internal/detectors"
	"faultline/internal/model"
)

const (
	ModeBayes = "bayes"
)

type RepoState struct {
	Root              string
	ChangedFiles      []string
	RecentFiles       []string
	HotspotDirs       []string
	HotfixSignals     []string
	DriftSignals      []string
	Provider          string
	TestsNewlyFailing []string
	ErrorsAdded       []string
	EnvDiff           map[string]model.DeltaEnvChange
}

type Inputs struct {
	Context        model.Context
	Lines          []model.Line
	Results        []model.Result
	RepoState      *RepoState
	ChangeSet      detectors.ChangeSet
	DeltaRequested bool
}

type weightsFile struct {
	Version        string             `json:"version"`
	PriorSmoothing float64            `json:"prior_smoothing"`
	PlaybookCounts map[string]int     `json:"playbook_counts"`
	FeatureWeights map[string]float64 `json:"feature_weights"`
}

type feature struct {
	Name         string
	Value        float64
	Weight       float64
	Reason       string
	EvidenceRefs []string
}
