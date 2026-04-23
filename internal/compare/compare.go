package compare

import (
	"fmt"
	"sort"

	"faultline/internal/model"
)

type Candidate struct {
	FailureID  string  `json:"failure_id,omitempty"`
	Title      string  `json:"title,omitempty"`
	Category   string  `json:"category,omitempty"`
	Score      float64 `json:"score,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

type StringDelta struct {
	Added   []string `json:"added,omitempty"`
	Removed []string `json:"removed,omitempty"`
}

type Report struct {
	LeftSource       string      `json:"left_source,omitempty"`
	RightSource      string      `json:"right_source,omitempty"`
	Changed          bool        `json:"changed"`
	DiagnosisChanged bool        `json:"diagnosis_changed"`
	StatusChanged    bool        `json:"status_changed"`
	Previous         *Candidate  `json:"previous,omitempty"`
	Current          *Candidate  `json:"current,omitempty"`
	Summary          []string    `json:"summary,omitempty"`
	Evidence         StringDelta `json:"evidence,omitempty"`
	FixSteps         StringDelta `json:"fix_steps,omitempty"`
	DominantSignals  StringDelta `json:"dominant_signals,omitempty"`
	RepoFiles        StringDelta `json:"repo_files,omitempty"`
	DeltaFiles       StringDelta `json:"delta_files,omitempty"`
	DeltaTests       StringDelta `json:"delta_tests,omitempty"`
	DeltaErrors      StringDelta `json:"delta_errors,omitempty"`
}

func Build(left, right *model.Analysis) Report {
	report := Report{
		LeftSource:  analysisSource(left),
		RightSource: analysisSource(right),
	}

	leftTop := topCandidate(left)
	rightTop := topCandidate(right)
	report.Previous = leftTop
	report.Current = rightTop
	report.DiagnosisChanged = diagnosisChanged(leftTop, rightTop)
	report.StatusChanged = analysisStatus(left) != analysisStatus(right)
	report.Evidence = diffStrings(topEvidence(left), topEvidence(right))
	report.FixSteps = diffStrings(fixSteps(left), fixSteps(right))
	report.DominantSignals = diffStrings(dominantSignals(left), dominantSignals(right))
	report.RepoFiles = diffStrings(repoFiles(left), repoFiles(right))
	report.DeltaFiles = diffStrings(deltaFiles(left), deltaFiles(right))
	report.DeltaTests = diffStrings(deltaTests(left), deltaTests(right))
	report.DeltaErrors = diffStrings(deltaErrors(left), deltaErrors(right))
	report.Summary = buildSummary(report)
	report.Changed = report.DiagnosisChanged ||
		report.StatusChanged ||
		hasDelta(report.Evidence) ||
		hasDelta(report.FixSteps) ||
		hasDelta(report.DominantSignals) ||
		hasDelta(report.RepoFiles) ||
		hasDelta(report.DeltaFiles) ||
		hasDelta(report.DeltaTests) ||
		hasDelta(report.DeltaErrors)
	return report
}

func buildSummary(report Report) []string {
	var lines []string
	switch {
	case report.Previous == nil && report.Current == nil:
		lines = append(lines, "neither artifact contains a matched diagnosis")
	case report.Previous == nil && report.Current != nil:
		lines = append(lines, fmt.Sprintf("a diagnosis appeared: %s", report.Current.FailureID))
	case report.Previous != nil && report.Current == nil:
		lines = append(lines, fmt.Sprintf("the previous diagnosis disappeared: %s", report.Previous.FailureID))
	case report.DiagnosisChanged:
		lines = append(lines, fmt.Sprintf("top diagnosis changed from %s to %s", report.Previous.FailureID, report.Current.FailureID))
	default:
		lines = append(lines, fmt.Sprintf("top diagnosis stayed the same: %s", report.Current.FailureID))
	}
	if report.StatusChanged {
		lines = append(lines, "artifact status changed between matched and unknown")
	}

	if len(report.Evidence.Added) > 0 {
		lines = append(lines, fmt.Sprintf("%d new top-diagnosis evidence line(s) appeared", len(report.Evidence.Added)))
	}
	if len(report.Evidence.Removed) > 0 {
		lines = append(lines, fmt.Sprintf("%d prior top-diagnosis evidence line(s) disappeared", len(report.Evidence.Removed)))
	}
	if len(report.DominantSignals.Added) > 0 || len(report.DominantSignals.Removed) > 0 {
		lines = append(lines, "dominant unmatched signals changed")
	}
	if len(report.DeltaFiles.Added) > 0 || len(report.DeltaTests.Added) > 0 || len(report.DeltaErrors.Added) > 0 {
		lines = append(lines, "current delta context contains new changed files, failing tests, or added errors")
	}
	if len(lines) == 1 && !hasDelta(report.Evidence) && !hasDelta(report.FixSteps) && !hasDelta(report.DominantSignals) && !hasDelta(report.RepoFiles) && !hasDelta(report.DeltaFiles) && !hasDelta(report.DeltaTests) && !hasDelta(report.DeltaErrors) {
		lines = append(lines, "no material differences were found in the compared artifacts")
	}
	return lines
}

func topCandidate(a *model.Analysis) *Candidate {
	if a == nil || len(a.Results) == 0 {
		return nil
	}
	top := a.Results[0]
	return &Candidate{
		FailureID:  top.Playbook.ID,
		Title:      top.Playbook.Title,
		Category:   top.Playbook.Category,
		Score:      top.Score,
		Confidence: top.Confidence,
	}
}

func topEvidence(a *model.Analysis) []string {
	if a != nil && a.Artifact != nil && len(a.Artifact.Evidence) > 0 {
		return append([]string(nil), a.Artifact.Evidence...)
	}
	if a == nil || len(a.Results) == 0 {
		return nil
	}
	return append([]string(nil), a.Results[0].Evidence...)
}

func fixSteps(a *model.Analysis) []string {
	if a == nil || a.Artifact == nil {
		return nil
	}
	return append([]string(nil), a.Artifact.FixSteps...)
}

func dominantSignals(a *model.Analysis) []string {
	if a == nil || a.Artifact == nil {
		return nil
	}
	return append([]string(nil), a.Artifact.DominantSignals...)
}

func repoFiles(a *model.Analysis) []string {
	if a == nil || a.RepoContext == nil {
		return nil
	}
	return append([]string(nil), a.RepoContext.RecentFiles...)
}

func deltaFiles(a *model.Analysis) []string {
	if a == nil || a.Delta == nil {
		return nil
	}
	return append([]string(nil), a.Delta.FilesChanged...)
}

func deltaTests(a *model.Analysis) []string {
	if a == nil || a.Delta == nil {
		return nil
	}
	return append([]string(nil), a.Delta.TestsNewlyFailing...)
}

func deltaErrors(a *model.Analysis) []string {
	if a == nil || a.Delta == nil {
		return nil
	}
	return append([]string(nil), a.Delta.ErrorsAdded...)
}

func diffStrings(left, right []string) StringDelta {
	leftSet := make(map[string]struct{}, len(left))
	rightSet := make(map[string]struct{}, len(right))
	for _, item := range left {
		if item == "" {
			continue
		}
		leftSet[item] = struct{}{}
	}
	for _, item := range right {
		if item == "" {
			continue
		}
		rightSet[item] = struct{}{}
	}

	var out StringDelta
	for item := range rightSet {
		if _, ok := leftSet[item]; !ok {
			out.Added = append(out.Added, item)
		}
	}
	for item := range leftSet {
		if _, ok := rightSet[item]; !ok {
			out.Removed = append(out.Removed, item)
		}
	}
	sort.Strings(out.Added)
	sort.Strings(out.Removed)
	return out
}

func diagnosisChanged(left, right *Candidate) bool {
	switch {
	case left == nil && right == nil:
		return false
	case left == nil || right == nil:
		return true
	default:
		return left.FailureID != right.FailureID
	}
}

func analysisSource(a *model.Analysis) string {
	if a == nil {
		return ""
	}
	return a.Source
}

func analysisStatus(a *model.Analysis) model.ArtifactStatus {
	if a == nil {
		return ""
	}
	if a.Artifact != nil && a.Artifact.Status != "" {
		return a.Artifact.Status
	}
	return a.Status
}

func hasDelta(delta StringDelta) bool {
	return len(delta.Added) > 0 || len(delta.Removed) > 0
}
