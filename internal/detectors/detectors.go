package detectors

import "faultline/internal/model"

// Kind identifies a detector module.
type Kind string

const (
	KindLog    Kind = "log"
	KindSource Kind = "source"
)

// Target is the canonical analysis input passed to detector modules.
type Target struct {
	LogLines       []model.Line
	LogContext     model.Context
	RepositoryRoot string
	Files          []SourceFile
	ChangeSet      ChangeSet
}

// SourceFile is a deterministic snapshot of a repository file.
type SourceFile struct {
	Path      string
	Content   string
	Lines     []string
	PathClass string
}

// ChangeSet lets CI or tests mark changed scopes without coupling the engine
// to a specific VCS implementation.
type ChangeSet struct {
	ChangedFiles map[string]FileChange
}

type FileChange struct {
	Status string
	Lines  map[int]struct{}
}

// Detector runs playbooks of a single kind against a canonical target.
type Detector interface {
	Kind() Kind
	Detect(playbooks []model.Playbook, target Target) []model.Result
}

func FilterPlaybooks(playbooks []model.Playbook, kind Kind) []model.Playbook {
	out := make([]model.Playbook, 0, len(playbooks))
	for _, pb := range playbooks {
		detector := pb.Detector
		if detector == "" {
			detector = string(KindLog)
		}
		if detector == string(kind) {
			out = append(out, pb)
		}
	}
	return out
}
