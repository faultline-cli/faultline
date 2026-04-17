package scoring

import (
	"math"
	"path/filepath"
	"sort"
	"strings"

	"faultline/internal/model"
)

func featureSet(weights weightsFile, inputs Inputs, result model.Result, baseline []model.Result, index int, delta *model.Delta) []feature {
	topScore := baseline[0].Score

	features := []feature{
		scalarFeature(weights, "detector_score", normalizeAgainst(result.Score, topScore), "baseline detector score remains the anchor", nil),
		scalarFeature(weights, "detector_confidence", clamp01(result.Confidence), "detector confidence supports the candidate", nil),
		scalarFeature(weights, "candidate_separation", baselineCandidateSeparation(baseline, index), "competitive separation supports only the baseline leader", nil),
		scalarFeature(weights, "log_match_coverage", logMatchCoverage(inputs, result), "broader explicit signal coverage supports the candidate", nil),
		scalarFeature(weights, "error_exact_match", errorExactMatch(inputs, result), "exact log-pattern overlap is present", result.Evidence),
		scalarFeature(weights, "error_fuzzy_overlap", errorFuzzyOverlap(inputs, result), "token overlap between evidence and playbook patterns supports the candidate", result.Evidence),
		scalarFeature(weights, "stage_hint_match", stageHintMatch(inputs.Context, result.Playbook), "stage hint matches the observed context", nil),
		scalarFeature(weights, "tool_or_stack_match", toolOrStackMatch(inputs, result), "tool or stack tokens align with the evidence", result.Evidence),
		scalarFeature(weights, "file_path_relevance", filePathRelevance(inputs, result), "likely files overlap with repo or source evidence", likelyFilesAndEvidence(result)),
		scalarFeature(weights, "changed_files_relevance", changedFilesRelevance(inputs, result), "recent changes overlap with the likely failure area", changedFileRefs(inputs)),
		scalarFeature(weights, "dependency_change_relevance", classRelevance(inputs, result, "dependency"), "dependency drift aligns with the candidate", changedFileRefs(inputs)),
		scalarFeature(weights, "runtime_toolchain_change_relevance", classRelevance(inputs, result, "runtime_toolchain"), "runtime or toolchain drift aligns with the candidate", changedFileRefs(inputs)),
		scalarFeature(weights, "ci_config_change_relevance", classRelevance(inputs, result, "ci_config"), "CI workflow drift aligns with the candidate", changedFileRefs(inputs)),
		scalarFeature(weights, "environment_drift_relevance", environmentDrift(inputs), "repo drift signals increase uncertainty around recent changes", driftRefs(inputs)),
		scalarFeature(weights, "repo_delta_agreement", repoDeltaAgreement(delta, result), "top-ranked delta cause agrees with the candidate", topDeltaRefs(delta)),
		scalarFeature(weights, "historical_fixture_support", historicalFixtureSupport(weights, result.Playbook), "accepted fixtures provide conservative support", nil),
		scalarFeature(weights, "mitigation_presence", mitigationPresence(result), "mitigations or suppressions weaken the match", mitigationRefs(result)),
	}
	features = append(features, deltaPlaybookFeatures(inputs, result, delta)...)

	out := make([]feature, 0, len(features))
	for _, item := range features {
		if item.Weight == 0 || item.Value == 0 {
			continue
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		ai := math.Abs(out[i].Weight * out[i].Value)
		aj := math.Abs(out[j].Weight * out[j].Value)
		if ai != aj {
			return ai > aj
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func deltaPlaybookFeatures(inputs Inputs, result model.Result, delta *model.Delta) []feature {
	if !inputs.DeltaRequested {
		return nil
	}
	var out []feature
	signals := deltaSignalMap(delta)
	if result.Playbook.RequiresDelta && delta == nil {
		out = append(out, feature{
			Name:   "delta_required_missing",
			Value:  1,
			Weight: -1.5,
			Reason: "candidate requires delta context but no change delta was available",
		})
		return out
	}
	if result.Playbook.RequiresDelta && len(result.Playbook.DeltaBoost) > 0 && !hasDeltaBoostSignal(signals, result.Playbook.DeltaBoost) {
		out = append(out, feature{
			Name:   "delta_required_unmatched",
			Value:  1,
			Weight: -0.75,
			Reason: "candidate requires delta support but the observed change does not match its delta signals",
		})
	}
	for _, boost := range result.Playbook.DeltaBoost {
		signal := strings.TrimSpace(boost.Signal)
		if signal == "" {
			continue
		}
		detail, ok := signals[signal]
		if !ok {
			continue
		}
		weight := boost.Weight
		if weight == 0 {
			weight = 1
		}
		out = append(out, feature{
			Name:         "delta_boost:" + signal,
			Value:        1,
			Weight:       round(weight),
			Reason:       "delta signal " + signal + " aligns with the candidate",
			EvidenceRefs: []string{detail},
		})
	}
	return out
}

func scalarFeature(weights weightsFile, name string, value float64, reason string, refs []string) feature {
	return feature{
		Name:         name,
		Value:        round(value),
		Weight:       weights.FeatureWeights[name],
		Reason:       reason,
		EvidenceRefs: dedupeStrings(refs),
	}
}

func normalizeAgainst(value, top float64) float64 {
	if top <= 0 {
		return 0
	}
	return clamp01(value / top)
}

func candidateSeparation(score, next float64) float64 {
	if score <= 0 {
		return 0
	}
	if next >= score {
		return 0
	}
	return clamp01((score - next) / score)
}

func baselineCandidateSeparation(baseline []model.Result, index int) float64 {
	if len(baseline) == 0 || index < 0 || index >= len(baseline) {
		return 0
	}
	score := baseline[index].Score
	if score <= 0 {
		return 0
	}
	// Separation is only trustworthy for the unique detector leader. Trailing
	// or tied candidates should not gain a reranking boost simply because they
	// appear later in the baseline order.
	if index != 0 {
		return 0
	}
	if len(baseline) == 1 {
		return 1
	}
	if baseline[1].Score >= score {
		return 0
	}
	return candidateSeparation(score, baseline[1].Score)
}

func logMatchCoverage(inputs Inputs, result model.Result) float64 {
	switch result.Detector {
	case "log":
		total := len(result.Playbook.Match.Any) + len(result.Playbook.Match.All)
		if total == 0 {
			return 0
		}
		hits := 0
		seen := map[string]struct{}{}
		for _, pat := range append(append([]string(nil), result.Playbook.Match.Any...), result.Playbook.Match.All...) {
			norm := normalizeText(pat)
			if norm == "" {
				continue
			}
			for _, line := range inputs.Lines {
				if strings.Contains(line.Normalized, norm) {
					if _, ok := seen[norm]; !ok {
						seen[norm] = struct{}{}
						hits++
					}
					break
				}
			}
		}
		return clamp01(float64(hits) / float64(total))
	case "source":
		total := len(result.Playbook.Source.Triggers)
		if total == 0 {
			return 0
		}
		return clamp01(float64(len(result.EvidenceBy.Triggers)) / float64(total))
	default:
		return 0
	}
}

func errorExactMatch(inputs Inputs, result model.Result) float64 {
	if result.Detector != "log" {
		return 0
	}
	patterns := append(append([]string(nil), result.Playbook.Match.Any...), result.Playbook.Match.All...)
	for _, evidence := range result.Evidence {
		normEvidence := normalizeText(evidence)
		for _, pat := range patterns {
			if normEvidence == normalizeText(pat) {
				return 1
			}
		}
	}
	return 0
}

func errorFuzzyOverlap(inputs Inputs, result model.Result) float64 {
	if result.Detector != "log" {
		return 0
	}
	patterns := append(append([]string(nil), result.Playbook.Match.Any...), result.Playbook.Match.All...)
	best := 0.0
	for _, evidence := range result.Evidence {
		eTokens := tokenSet(evidence)
		for _, pat := range patterns {
			score := jaccard(eTokens, tokenSet(pat))
			if score > best {
				best = score
			}
		}
	}
	return clamp01(best)
}

func stageHintMatch(ctx model.Context, pb model.Playbook) float64 {
	if strings.TrimSpace(ctx.Stage) == "" {
		return 0
	}
	for _, hint := range pb.StageHints {
		if strings.EqualFold(hint, ctx.Stage) {
			return 1
		}
	}
	return 0
}

func toolOrStackMatch(inputs Inputs, result model.Result) float64 {
	haystack := strings.ToLower(strings.Join(append(append([]string(nil), result.Evidence...), inputs.Context.CommandHint, inputs.Context.Step), " "))
	if strings.TrimSpace(haystack) == "" {
		return 0
	}
	matched := 0
	total := 0
	for _, token := range candidateTokens(result.Playbook) {
		total++
		if strings.Contains(haystack, token) {
			matched++
		}
	}
	if total == 0 {
		return 0
	}
	return clamp01(float64(matched) / float64(total))
}

func filePathRelevance(inputs Inputs, result model.Result) float64 {
	files := likelyFilesAndEvidence(result)
	if len(files) == 0 || inputs.RepoState == nil {
		return 0
	}
	matches := 0
	for _, file := range dedupeStrings(append(append([]string(nil), inputs.RepoState.RecentFiles...), inputs.RepoState.ChangedFiles...)) {
		if matchesFileHints(file, files) {
			matches++
		}
	}
	return ratio(matches, len(dedupeStrings(files)))
}

func changedFilesRelevance(inputs Inputs, result model.Result) float64 {
	if inputs.RepoState == nil || len(inputs.RepoState.ChangedFiles) == 0 {
		return 0
	}
	files := likelyFilesAndEvidence(result)
	if len(files) == 0 {
		return 0
	}
	matches := 0
	for _, file := range inputs.RepoState.ChangedFiles {
		if matchesFileHints(file, files) {
			matches++
		}
	}
	return ratio(matches, len(inputs.RepoState.ChangedFiles))
}

func classRelevance(inputs Inputs, result model.Result, class string) float64 {
	if inputs.RepoState == nil || len(inputs.RepoState.ChangedFiles) == 0 {
		return 0
	}
	if !playbookSupportsClass(result.Playbook, class) {
		return 0
	}
	matches := 0
	for _, file := range inputs.RepoState.ChangedFiles {
		k, _, _ := classifyDeltaFile(file)
		if k == class {
			matches++
		}
	}
	return ratio(matches, len(inputs.RepoState.ChangedFiles))
}

func environmentDrift(inputs Inputs) float64 {
	if inputs.RepoState == nil {
		return 0
	}
	score := float64(len(inputs.RepoState.DriftSignals))*0.2 + float64(len(inputs.RepoState.HotfixSignals))*0.1
	return clamp01(score)
}

func repoDeltaAgreement(delta *model.Delta, result model.Result) float64 {
	if delta == nil || len(delta.Causes) == 0 {
		return 0
	}
	top := delta.Causes[0].Kind
	for _, class := range playbookLikelyClasses(result.Playbook) {
		if class == top {
			return 1
		}
	}
	return 0
}

func historicalFixtureSupport(weights weightsFile, pb model.Playbook) float64 {
	count := weights.PlaybookCounts[pb.ID]
	if count <= 0 {
		return 0
	}
	maxCount := 0
	for _, item := range weights.PlaybookCounts {
		if item > maxCount {
			maxCount = item
		}
	}
	if maxCount == 0 {
		return 0
	}
	return clamp01(float64(count) / float64(maxCount))
}

func mitigationPresence(result model.Result) float64 {
	negatives := len(result.EvidenceBy.Mitigations) + len(result.EvidenceBy.Suppressions)
	if negatives == 0 {
		return 0
	}
	base := len(result.EvidenceBy.Triggers)
	if base == 0 {
		base = 1
	}
	return clamp01(float64(negatives) / float64(base))
}

func likelyFilesAndEvidence(result model.Result) []string {
	files := append([]string(nil), result.Playbook.Workflow.LikelyFiles...)
	for _, evidence := range append(append([]model.Evidence(nil), result.EvidenceBy.Triggers...), result.EvidenceBy.Context...) {
		if strings.TrimSpace(evidence.File) != "" {
			files = append(files, evidence.File)
		}
	}
	return dedupeStrings(files)
}

func changedFileRefs(inputs Inputs) []string {
	if inputs.RepoState == nil {
		return nil
	}
	return dedupeStrings(inputs.RepoState.ChangedFiles)
}

func driftRefs(inputs Inputs) []string {
	if inputs.RepoState == nil {
		return nil
	}
	return dedupeStrings(inputs.RepoState.DriftSignals)
}

func topDeltaRefs(delta *model.Delta) []string {
	if delta == nil || len(delta.Causes) == 0 {
		return nil
	}
	return append([]string(nil), delta.Causes[0].Reasons...)
}

func mitigationRefs(result model.Result) []string {
	var refs []string
	for _, item := range result.EvidenceBy.Mitigations {
		if strings.TrimSpace(item.Detail) != "" {
			refs = append(refs, item.Detail)
		}
	}
	for _, item := range result.EvidenceBy.Suppressions {
		if strings.TrimSpace(item.Detail) != "" {
			refs = append(refs, item.Detail)
		}
	}
	return dedupeStrings(refs)
}

func deltaSignalMap(delta *model.Delta) map[string]string {
	if delta == nil || len(delta.Signals) == 0 {
		return nil
	}
	out := make(map[string]string, len(delta.Signals))
	for _, signal := range delta.Signals {
		id := strings.TrimSpace(signal.ID)
		if id == "" {
			continue
		}
		out[id] = strings.TrimSpace(signal.Detail)
	}
	return out
}

func hasDeltaBoostSignal(signals map[string]string, boosts []model.DeltaBoost) bool {
	if len(signals) == 0 {
		return false
	}
	for _, boost := range boosts {
		if _, ok := signals[strings.TrimSpace(boost.Signal)]; ok {
			return true
		}
	}
	return false
}

func playbookSupportsClass(pb model.Playbook, class string) bool {
	for _, candidate := range playbookLikelyClasses(pb) {
		if candidate == class {
			return true
		}
	}
	return false
}

func playbookLikelyClasses(pb model.Playbook) []string {
	var classes []string
	id := strings.ToLower(pb.ID)
	category := strings.ToLower(pb.Category)
	switch {
	case strings.Contains(id, "lockfile"), strings.Contains(id, "dependency"), strings.Contains(id, "go-sum"), strings.Contains(id, "pip"), strings.Contains(id, "poetry"), strings.Contains(id, "npm"):
		classes = append(classes, "dependency")
	case strings.Contains(id, "runtime"), strings.Contains(id, "node-version"), strings.Contains(id, "missing-executable"):
		classes = append(classes, "runtime_toolchain")
	}
	switch category {
	case "ci":
		classes = append(classes, "ci_config")
	case "deploy":
		classes = append(classes, "deploy_infra")
	case "auth", "runtime", "network":
		classes = append(classes, "environment")
	case "test":
		classes = append(classes, "test_data")
	}
	if strings.EqualFold(pb.Detector, "source") {
		classes = append(classes, "source_code")
	}
	for _, file := range pb.Workflow.LikelyFiles {
		class, _, _ := classifyDeltaFile(file)
		if class != "" {
			classes = append(classes, class)
		}
	}
	return dedupeStrings(classes)
}

func matchesFileHints(file string, hints []string) bool {
	file = filepath.ToSlash(file)
	base := filepath.Base(file)
	for _, hint := range hints {
		hint = strings.TrimSpace(filepath.ToSlash(hint))
		if hint == "" {
			continue
		}
		if strings.Contains(hint, "*") {
			if ok, _ := filepath.Match(hint, file); ok {
				return true
			}
			if ok, _ := filepath.Match(hint, base); ok {
				return true
			}
			continue
		}
		if file == hint || base == hint || strings.HasSuffix(file, "/"+hint) || strings.Contains(file, hint) {
			return true
		}
	}
	return false
}

func candidateTokens(pb model.Playbook) []string {
	tokens := append([]string(nil), pb.Tags...)
	tokens = append(tokens, pb.ID, pb.Title)
	out := make([]string, 0, len(tokens))
	seen := map[string]struct{}{}
	for _, token := range tokens {
		for _, part := range tokenSet(token) {
			if len(part) < 3 {
				continue
			}
			if _, ok := seen[part]; ok {
				continue
			}
			seen[part] = struct{}{}
			out = append(out, part)
		}
	}
	sort.Strings(out)
	return out
}

func tokenSet(value string) []string {
	fields := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	return dedupeStrings(fields)
}

func jaccard(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	seen := map[string]struct{}{}
	for _, item := range a {
		seen[item] = struct{}{}
	}
	inter := 0
	union := len(seen)
	for _, item := range b {
		if _, ok := seen[item]; ok {
			inter++
			continue
		}
		union++
	}
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}
