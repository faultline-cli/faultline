package delta

import "faultline/internal/model"

type Options struct {
	Provider string
	GitHub   GitHubOptions
}

type GitHubOptions struct {
	Repository string
	Branch     string
	RunID      int64
	Token      string
	APIBaseURL string
}

type Snapshot struct {
	Provider          string
	FilesChanged      []string
	TestsNewlyFailing []string
	ErrorsAdded       []string
	EnvDiff           map[string]model.DeltaEnvChange
}
