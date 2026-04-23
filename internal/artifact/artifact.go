package artifact

import (
	"path/filepath"
	"strconv"
	"strings"

	"faultline/internal/model"
)

const SchemaVersion = "failure_artifact.v1"

func Build(a *model.Analysis) *model.FailureArtifact {
	if a == nil {
		return nil
	}
	artifact := &model.FailureArtifact{
		SchemaVersion: SchemaVersion,
		Status:        statusForAnalysis(a),
		Fingerprint:   strings.TrimSpace(a.Fingerprint),
		Environment: model.ArtifactEnvironment{
			Source:         a.Source,
			Context:        a.Context,
			PackProvenance: append([]model.PackProvenance(nil), a.PackProvenances...),
		},
		CandidateClusters: append([]model.CandidateCluster(nil), a.CandidateClusters...),
		DominantSignals:   append([]string(nil), a.DominantSignals...),
	}
	if a.SuggestedPlaybookSeed != nil {
		seed := *a.SuggestedPlaybookSeed
		artifact.SuggestedPlaybookSeed = &seed
	}
	if a.RepoContext != nil {
		artifact.Environment.RepoRoot = a.RepoContext.RepoRoot
		artifact.Environment.RecentFiles = append([]string(nil), a.RepoContext.RecentFiles...)
		artifact.Environment.RelatedCommits = append([]model.RepoCommit(nil), a.RepoContext.RelatedCommits...)
	}
	if a.Delta != nil {
		artifact.Environment.DeltaProvider = a.Delta.Provider
	}

	if len(a.Results) == 0 {
		artifact.Confidence = unknownConfidence(a)
		artifact.Evidence = append([]string(nil), artifact.DominantSignals...)
		artifact.FixSteps = unknownFixSteps(a)
		artifact.Remediation = buildUnknownRemediation(a)
		return artifact
	}

	top := a.Results[0]
	artifact.MatchedPlaybook = &model.ArtifactPlaybook{
		ID:       top.Playbook.ID,
		Title:    top.Playbook.Title,
		Category: top.Playbook.Category,
		Severity: top.Playbook.Severity,
		Detector: top.Detector,
		Pack:     displayPackName(top.Playbook.Metadata.PackName),
	}
	artifact.Evidence = append([]string(nil), top.Evidence...)
	artifact.Confidence = refinedConfidence(a, top)
	artifact.HistoryContext = &model.ArtifactHistoryContext{
		SeenCount:          top.SeenCount,
		SignatureHash:      top.SignatureHash,
		SeenBefore:         top.SeenBefore,
		OccurrenceCount:    top.OccurrenceCount,
		FirstSeenAt:        top.FirstSeenAt,
		LastSeenAt:         top.LastSeenAt,
		HookHistorySummary: cloneHookHistory(top.HookHistorySummary),
	}
	artifact.FixSteps = markdownListItems(top.Playbook.Fix)
	artifact.Remediation = buildMatchedRemediation(a, top)
	return artifact
}

func Sync(a *model.Analysis) *model.Analysis {
	if a == nil {
		return nil
	}
	clone := *a
	clone.Results = append([]model.Result(nil), a.Results...)
	clone.CandidateClusters = append([]model.CandidateCluster(nil), a.CandidateClusters...)
	clone.DominantSignals = append([]string(nil), a.DominantSignals...)
	if a.SuggestedPlaybookSeed != nil {
		seed := *a.SuggestedPlaybookSeed
		clone.SuggestedPlaybookSeed = &seed
	}
	clone.Status = statusForAnalysis(&clone)
	clone.Artifact = Build(&clone)
	return &clone
}

func statusForAnalysis(a *model.Analysis) model.ArtifactStatus {
	if a != nil && len(a.Results) > 0 {
		return model.ArtifactStatusMatched
	}
	return model.ArtifactStatusUnknown
}

func refinedConfidence(a *model.Analysis, result model.Result) float64 {
	confidence := result.Confidence
	if a == nil {
		return confidence
	}
	if a.RepoContext != nil && len(a.RepoContext.RelatedCommits) > 0 {
		confidence += 0.04
	}
	if a.Delta != nil && (len(a.Delta.Causes) > 0 || len(a.Delta.Signals) > 0) {
		confidence += 0.03
	}
	if result.Hooks != nil {
		confidence = result.Hooks.FinalConfidence
	}
	if confidence > 1 {
		return 1
	}
	return confidence
}

func unknownConfidence(a *model.Analysis) float64 {
	if a == nil || len(a.CandidateClusters) == 0 {
		return 0.2
	}
	return a.CandidateClusters[0].Confidence
}

func unknownFixSteps(a *model.Analysis) []string {
	steps := []string{
		"Capture the shortest stable failing excerpt and keep the surrounding noise out of the seed artifact.",
		"Use `faultline list` and `faultline explain <nearest-playbook>` to confirm whether an existing bundled diagnosis can absorb the dominant unmatched signals before authoring a new rule.",
	}
	if a != nil && a.SuggestedPlaybookSeed != nil && len(a.SuggestedPlaybookSeed.MatchAny) > 0 {
		steps = append(steps, "Start a new playbook seed from the dominant unmatched signals and confirm the nearest bundled neighbor cannot absorb them.")
	}
	if a != nil && len(a.DominantSignals) > 0 {
		steps = append(steps, "Inspect the dominant signals first and group them by one root cause before adding any new rule.")
	}
	steps = append(steps, "Re-run analysis after tightening the seed patterns or extending the catalog to confirm the unknown case now lands deterministically.")
	return steps
}

func buildUnknownRemediation(a *model.Analysis) *model.RemediationPlan {
	plan := &model.RemediationPlan{}
	if a != nil && a.SuggestedPlaybookSeed != nil && len(a.SuggestedPlaybookSeed.MatchAny) > 0 {
		plan.PatchSuggestions = append(plan.PatchSuggestions, model.PatchSuggestion{
			TargetFile: playbookSeedPath(a.SuggestedPlaybookSeed),
			Summary:    "Draft a new playbook seed from the unmatched dominant signals.",
			Actions: []string{
				"Copy the suggested match_any signals into a new playbook scaffold.",
				"Add at least one exclusion for the nearest confusable bundled playbook.",
				"Attach workflow.likely_files, workflow.local_repro, and workflow.verify before promotion.",
			},
		})
	}
	return plan
}

func buildMatchedRemediation(a *model.Analysis, result model.Result) *model.RemediationPlan {
	plan := &model.RemediationPlan{}
	for i, cmd := range result.Playbook.Workflow.LocalRepro {
		plan.Commands = append(plan.Commands, buildCommand("local-repro", i, cmd, "Reproduce the failure before editing."))
	}
	for i, cmd := range result.Playbook.Workflow.Verify {
		plan.Commands = append(plan.Commands, buildCommand("verify", i, cmd, "Confirm the fix removes the failure signature."))
	}

	files := likelyFiles(a, result)
	fixSteps := markdownListItems(result.Playbook.Fix)
	for _, file := range files {
		suggestion := model.PatchSuggestion{
			TargetFile: file,
			Summary:    "Adjust the likely failure surface for " + result.Playbook.ID + ".",
			Actions:    append([]string(nil), fixSteps...),
		}
		if len(suggestion.Actions) == 0 {
			suggestion.Actions = []string{"Apply the smallest deterministic change that aligns the file with the diagnosed fix path."}
		}
		plan.PatchSuggestions = append(plan.PatchSuggestions, suggestion)
		if isCIConfigFile(file) {
			plan.CIConfigDiffs = append(plan.CIConfigDiffs, model.CIConfigDiff{
				TargetFile: file,
				Summary:    "Update CI configuration to reflect the diagnosed remediation path.",
				Before: []string{
					"Failing step runs without the setup or path assumptions required by the diagnosis.",
				},
				After: ciAfterHints(result, fixSteps),
			})
		}
	}
	if len(plan.Commands) == 0 && len(plan.PatchSuggestions) == 0 && len(plan.CIConfigDiffs) == 0 {
		return nil
	}
	return plan
}

func buildCommand(phase string, index int, cmd string, rationale string) model.RemediationCommand {
	return model.RemediationCommand{
		ID:        phase + "-" + strconv.Itoa(index+1),
		Phase:     phase,
		Command:   splitCommand(cmd),
		WorkDir:   ".",
		Rationale: rationale,
	}
}

func likelyFiles(a *model.Analysis, result model.Result) []string {
	seen := map[string]struct{}{}
	var files []string
	add := func(values []string) {
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			files = append(files, value)
		}
	}
	add(result.Playbook.Workflow.LikelyFiles)
	if a != nil && a.RepoContext != nil {
		add(a.RepoContext.RecentFiles)
	}
	if len(files) > 5 {
		files = files[:5]
	}
	return files
}

func ciAfterHints(result model.Result, fixSteps []string) []string {
	hints := []string{
		"Add or reorder setup so the failing step sees the expected runtime, dependency, or artifact state.",
		"Keep the CI change minimal and verify it with the shipped workflow.verify commands.",
	}
	for _, step := range fixSteps {
		if strings.TrimSpace(step) == "" {
			continue
		}
		hints = append(hints, step)
		if len(hints) >= 3 {
			break
		}
	}
	return hints
}

func markdownListItems(markdown string) []string {
	lines := strings.Split(markdown, "\n")
	items := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "- "):
			items = append(items, strings.TrimSpace(strings.TrimPrefix(line, "- ")))
		case len(line) > 3 && line[1] == '.' && line[2] == ' ' && line[0] >= '0' && line[0] <= '9':
			items = append(items, strings.TrimSpace(line[3:]))
		}
	}
	return items
}

func splitCommand(cmd string) []string {
	fields := strings.Fields(strings.TrimSpace(cmd))
	if len(fields) == 0 {
		return nil
	}
	return fields
}

func cloneHookHistory(summary *model.HookHistorySummary) *model.HookHistorySummary {
	if summary == nil {
		return nil
	}
	clone := *summary
	return &clone
}

func displayPackName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == "starter" || name == "custom" {
		return ""
	}
	return name
}

func isCIConfigFile(path string) bool {
	clean := filepath.ToSlash(strings.TrimSpace(path))
	switch {
	case strings.HasPrefix(clean, ".github/workflows/"):
		return true
	case clean == ".gitlab-ci.yml":
		return true
	case clean == "azure-pipelines.yml":
		return true
	case strings.HasPrefix(clean, ".circleci/"):
		return true
	case clean == "Jenkinsfile":
		return true
	default:
		return false
	}
}

func playbookSeedPath(seed *model.SuggestedPlaybookSeed) string {
	if seed == nil || strings.TrimSpace(seed.Category) == "" {
		return "playbooks/bundled/log/unknown/seed.yaml"
	}
	return filepath.ToSlash(filepath.Join("playbooks", "bundled", "log", seed.Category, "seed.yaml"))
}
