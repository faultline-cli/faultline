package sourcedetector

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"

	"faultline/internal/detectors"
	"faultline/internal/model"
)

type Detector struct{}

type occurrence struct {
	evidence  model.Evidence
	scopeKey  string
	moduleKey string
}

type consistencyCandidate struct {
	moduleKey string
	scopeKey  string
	hasBase   bool
	hasExpect bool
}

func (Detector) Kind() detectors.Kind {
	return detectors.KindSource
}

func (Detector) Detect(playbooks []model.Playbook, target detectors.Target) []model.Result {
	files := prepareFiles(target.Files)
	results := make([]model.Result, 0, len(playbooks))
	for _, pb := range playbooks {
		result := detectPlaybook(pb, files, target.ChangeSet)
		if result.Score == 0 {
			continue
		}
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		if results[i].Confidence != results[j].Confidence {
			return results[i].Confidence > results[j].Confidence
		}
		return results[i].Playbook.ID < results[j].Playbook.ID
	})
	return results
}

type preparedFile struct {
	path       string
	lines      []preparedLine
	moduleKey  string
	pathClass  string
	hotPath    bool
	critical   bool
	classBonus float64
}

type preparedLine struct {
	original   string
	normalized string
	number     int
	function   string
	depth      int
}

func prepareFiles(files []detectors.SourceFile) []preparedFile {
	out := make([]preparedFile, 0, len(files))
	for _, file := range files {
		lines := make([]preparedLine, 0, len(file.Lines))
		currentFunc := ""
		funcDepth := 0
		depth := 0
		for i, line := range file.Lines {
			trimmed := strings.TrimSpace(line)
			if fn := inferFunctionName(trimmed); fn != "" {
				currentFunc = fn
				funcDepth = depth
			}
			lines = append(lines, preparedLine{
				original:   line,
				normalized: normalize(line),
				number:     i + 1,
				function:   currentFunc,
				depth:      depth,
			})
			depth += strings.Count(line, "{") - strings.Count(line, "}")
			if currentFunc != "" && depth <= funcDepth {
				currentFunc = ""
			}
		}
		pathClass, hotPath, critical, classBonus := classifyPath(file.Path)
		out = append(out, preparedFile{
			path:       file.Path,
			lines:      lines,
			moduleKey:  filepath.Dir(file.Path),
			pathClass:  pathClass,
			hotPath:    hotPath,
			critical:   critical,
			classBonus: classBonus,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].path < out[j].path
	})
	return out
}

func detectPlaybook(pb model.Playbook, files []preparedFile, changeSet detectors.ChangeSet) model.Result {
	triggers := collectMatches(pb.Source.Triggers, model.EvidenceTrigger, pb, files)
	if len(triggers) == 0 {
		return model.Result{}
	}
	amplifiers := collectMatches(pb.Source.Amplifiers, model.EvidenceAmplifier, pb, files)
	mitigations := collectMatches(pb.Source.Mitigations, model.EvidenceMitigation, pb, files)
	contextEvidence := collectMatches(pb.Source.Context, model.EvidenceContext, pb, files)
	safeContext := collectSafeContext(pb, files)
	suppressions, fullySuppressed := collectSuppressions(pb, files)
	consistency := collectConsistency(pb, triggers, mitigations)
	amplifiers = append(amplifiers, consistency...)

	baseScore := pb.BaseScore + weightedSum(triggers, nil)
	ampBonus := weightedSum(amplifiers, triggers)
	mitigationDiscount := weightedSum(mitigations, triggers)
	safeDiscount := weightedSum(safeContext, triggers)
	exceptionDiscount := weightedSuppression(suppressions)
	compoundBonus, compoundEvidence := applyCompoundSignals(pb, triggers, amplifiers, mitigations)
	amplifiers = append(amplifiers, compoundEvidence...)
	blastBonus, hotBonus := contextBonuses(pb, triggers, contextEvidence)
	changeBonus, changeStatus := changeAdjustment(pb, triggers, changeSet)

	finalScore := baseScore + ampBonus + compoundBonus + blastBonus + hotBonus + changeBonus - mitigationDiscount - exceptionDiscount - safeDiscount
	if fullySuppressed {
		finalScore = 0
	}
	if finalScore <= 0 {
		return model.Result{}
	}

	allEvidence := mergeEvidence(triggers, amplifiers, mitigations, suppressions, append(contextEvidence, safeContext...))
	score := round(finalScore)
	confidence := confidenceFromScore(baseScore+ampBonus+compoundBonus, finalScore)
	return model.Result{
		Playbook:   pb,
		Detector:   string(detectors.KindSource),
		Score:      score,
		Confidence: confidence,
		Evidence:   compactEvidence(allEvidence),
		EvidenceBy: model.EvidenceBundle{
			Triggers:     evidenceList(triggers),
			Amplifiers:   evidenceList(amplifiers),
			Mitigations:  evidenceList(mitigations),
			Suppressions: evidenceList(suppressions),
			Context:      evidenceList(append(contextEvidence, safeContext...)),
		},
		Explanation: buildExplanation(triggers, amplifiers, mitigations, suppressions, append(contextEvidence, safeContext...), changeStatus),
		Breakdown: model.ScoreBreakdown{
			BaseSignalScore:            round(baseScore),
			CompoundSignalBonus:        round(compoundBonus),
			BlastRadiusMultiplier:      round(blastBonus),
			HotPathMultiplier:          round(hotBonus),
			ChangeIntroducedBonus:      round(changeBonus),
			MitigatingEvidenceDiscount: round(mitigationDiscount),
			ExplicitExceptionDiscount:  round(exceptionDiscount),
			SafeContextDiscount:        round(safeDiscount),
			FinalScore:                 score,
		},
		ChangeStatus: changeStatus,
	}
}

func collectMatches(matchers []model.SignalMatcher, kind model.EvidenceKind, pb model.Playbook, files []preparedFile) []occurrence {
	out := make([]occurrence, 0)
	for _, matcher := range matchers {
		label := matcher.Label
		if label == "" {
			label = matcher.ID
		}
		weight := matcher.Weight
		if weight == 0 {
			switch kind {
			case model.EvidenceAmplifier:
				weight = 1
			case model.EvidenceMitigation:
				weight = 1
			default:
				weight = 1
			}
		}
		for _, file := range files {
			if !pathAllowed(file.path, pb.Contextual, matcher.PathIncludes, matcher.PathExcludes) {
				continue
			}
			for _, line := range file.lines {
				if !containsPattern(line.normalized, matcher.Patterns) {
					continue
				}
				scopeName, scopeKey, proximity := inferScope(file, line)
				out = append(out, occurrence{
					evidence: model.Evidence{
						Kind:      kind,
						SignalID:  matcher.ID,
						Label:     label,
						Detail:    strings.TrimSpace(line.original),
						File:      file.path,
						Line:      line.number,
						PathClass: file.pathClass,
						Scope:     "function",
						ScopeName: scopeName,
						Proximity: proximity,
						Weight:    weight,
						Source:    "source",
					},
					scopeKey:  scopeKey,
					moduleKey: file.moduleKey,
				})
			}
		}
	}
	sortOccurrences(out)
	return out
}

func collectSafeContext(pb model.Playbook, files []preparedFile) []occurrence {
	out := make([]occurrence, 0)
	for _, rule := range pb.Source.SafeContextClasses {
		for _, file := range files {
			if !containsAnyPath(file.path, rule.Paths) {
				continue
			}
			out = append(out, occurrence{
				evidence: model.Evidence{
					Kind:      model.EvidenceContext,
					SignalID:  rule.ID,
					Label:     coalesce(rule.Label, "safe context"),
					Detail:    fmt.Sprintf("%s is classified as safe context", file.path),
					File:      file.path,
					PathClass: file.pathClass,
					Scope:     "file",
					ScopeName: file.path,
					Proximity: "same_file",
					Weight:    defaultFloat(rule.Discount, pb.Scoring.SafeContextDiscount, 1),
					Source:    "source",
				},
				scopeKey:  file.path,
				moduleKey: file.moduleKey,
			})
		}
	}
	sortOccurrences(out)
	return out
}

func collectSuppressions(pb model.Playbook, files []preparedFile) ([]occurrence, bool) {
	out := make([]occurrence, 0)
	fullySuppressed := false
	for _, file := range files {
		for _, line := range file.lines {
			norm := line.normalized
			if strings.Contains(norm, "faultline:ignore "+pb.ID) || strings.Contains(norm, "faultline:disable "+pb.ID) {
				reason := extractField(line.original, "reason=")
				expires := extractField(line.original, "until=")
				e := occurrence{
					evidence: model.Evidence{
						Kind:       model.EvidenceSuppression,
						Label:      "inline suppression",
						Detail:     strings.TrimSpace(line.original),
						File:       file.path,
						Line:       line.number,
						PathClass:  file.pathClass,
						Scope:      "line",
						ScopeName:  file.path,
						Proximity:  "same_file",
						Weight:     defaultFloat(pb.Scoring.DefaultSuppressionDiscount, 2, 2),
						Suppressed: true,
						ExpiresOn:  expires,
						Reason:     reason,
						Source:     "inline",
					},
					scopeKey:  file.path,
					moduleKey: file.moduleKey,
				}
				out = append(out, e)
				fullySuppressed = true
			}
		}
	}
	for _, rule := range pb.Source.Suppressions {
		for _, file := range files {
			if len(rule.Paths) > 0 && !containsAnyPath(file.path, rule.Paths) {
				continue
			}
			if len(rule.Playbooks) > 0 && !containsPattern(pb.ID, rule.Playbooks) {
				continue
			}
			if rule.Pattern != "" {
				for _, line := range file.lines {
					if !strings.Contains(line.normalized, normalize(rule.Pattern)) {
						continue
					}
					out = append(out, occurrence{
						evidence: model.Evidence{
							Kind:       model.EvidenceSuppression,
							Label:      coalesce(rule.Style, "configured suppression"),
							Detail:     strings.TrimSpace(line.original),
							File:       file.path,
							Line:       line.number,
							PathClass:  file.pathClass,
							Scope:      "line",
							ScopeName:  file.path,
							Proximity:  "same_file",
							Weight:     defaultFloat(rule.Discount, pb.Scoring.DefaultSuppressionDiscount, 2),
							Suppressed: rule.SuppressAll,
							ExpiresOn:  rule.ExpiresOn,
							Reason:     rule.Reason,
							Source:     "playbook",
						},
						scopeKey:  file.path,
						moduleKey: file.moduleKey,
					})
					fullySuppressed = fullySuppressed || rule.SuppressAll
				}
				continue
			}
			out = append(out, occurrence{
				evidence: model.Evidence{
					Kind:       model.EvidenceSuppression,
					Label:      coalesce(rule.Style, "path suppression"),
					Detail:     fmt.Sprintf("%s suppressed by playbook rule", file.path),
					File:       file.path,
					PathClass:  file.pathClass,
					Scope:      "file",
					ScopeName:  file.path,
					Proximity:  "same_file",
					Weight:     defaultFloat(rule.Discount, pb.Scoring.DefaultSuppressionDiscount, 2),
					Suppressed: rule.SuppressAll,
					ExpiresOn:  rule.ExpiresOn,
					Reason:     rule.Reason,
					Source:     "playbook",
				},
				scopeKey:  file.path,
				moduleKey: file.moduleKey,
			})
			fullySuppressed = fullySuppressed || rule.SuppressAll
		}
	}
	sortOccurrences(out)
	return out, fullySuppressed
}

func collectConsistency(pb model.Playbook, triggers, mitigations []occurrence) []occurrence {
	if len(pb.Source.LocalConsistency) == 0 {
		return nil
	}
	candidates := make(map[string]*consistencyCandidate)
	for _, trigger := range triggers {
		key := trigger.moduleKey + "|" + trigger.scopeKey
		entry := candidates[key]
		if entry == nil {
			entry = &consistencyCandidate{moduleKey: trigger.moduleKey, scopeKey: trigger.scopeKey}
			candidates[key] = entry
		}
		entry.hasBase = true
	}
	for _, mitigation := range mitigations {
		key := mitigation.moduleKey + "|" + mitigation.scopeKey
		entry := candidates[key]
		if entry == nil {
			entry = &consistencyCandidate{moduleKey: mitigation.moduleKey, scopeKey: mitigation.scopeKey}
			candidates[key] = entry
		}
		entry.hasExpect = true
	}
	moduleStats := make(map[string][2]int)
	for _, entry := range candidates {
		stats := moduleStats[entry.moduleKey]
		if entry.hasBase {
			stats[0]++
			if entry.hasExpect {
				stats[1]++
			}
		}
		moduleStats[entry.moduleKey] = stats
	}
	out := make([]occurrence, 0)
	for _, rule := range pb.Source.LocalConsistency {
		for _, trigger := range triggers {
			stats := moduleStats[trigger.moduleKey]
			if stats[0] == 0 {
				continue
			}
			if stats[0] < defaultInt(rule.MinimumPeers, 2) {
				continue
			}
			ratio := float64(stats[1]) / float64(stats[0])
			threshold := defaultFloat(rule.Threshold, 0.6, 0.6)
			if ratio < threshold || scopeHasMitigation(trigger.scopeKey, mitigations) {
				continue
			}
			out = append(out, occurrence{
				evidence: model.Evidence{
					Kind:      model.EvidenceAmplifier,
					SignalID:  rule.ID,
					Label:     coalesce(rule.Label, "violates local idiom"),
					Detail:    fmt.Sprintf("%s omits the local safeguard used in %.0f%% of peer scopes", trigger.scopeKey, ratio*100),
					File:      trigger.evidence.File,
					Line:      trigger.evidence.Line,
					PathClass: trigger.evidence.PathClass,
					Scope:     trigger.evidence.Scope,
					ScopeName: trigger.evidence.ScopeName,
					Proximity: "same_module",
					Weight:    defaultFloat(rule.Amplifier, 1.25, 1.25),
					Source:    "source",
				},
				scopeKey:  trigger.scopeKey,
				moduleKey: trigger.moduleKey,
			})
			break
		}
	}
	sortOccurrences(out)
	return out
}

func applyCompoundSignals(pb model.Playbook, triggers, amplifiers, mitigations []occurrence) (float64, []occurrence) {
	total := 0.0
	out := make([]occurrence, 0)
	signalMap := make(map[string][]occurrence)
	for _, item := range append(append([]occurrence{}, triggers...), amplifiers...) {
		signalMap[item.evidence.SignalID] = append(signalMap[item.evidence.SignalID], item)
	}
	mitigatedScopes := make(map[string]struct{})
	for _, mitigation := range mitigations {
		mitigatedScopes[mitigation.scopeKey] = struct{}{}
	}
	for _, compound := range pb.Source.CompoundSignals {
		if len(compound.Require) == 0 {
			continue
		}
		scopeCounts := make(map[string]int)
		reference := make(map[string]occurrence)
		for _, signalID := range compound.Require {
			for _, hit := range signalMap[signalID] {
				scopeKey := scopedKey(compound.Scope, hit)
				scopeCounts[scopeKey]++
				if _, ok := reference[scopeKey]; !ok {
					reference[scopeKey] = hit
				}
			}
		}
		// Collect and sort scope keys for deterministic iteration
		scopeKeys := make([]string, 0, len(scopeCounts))
		for scopeKey := range scopeCounts {
			scopeKeys = append(scopeKeys, scopeKey)
		}
		sort.Strings(scopeKeys)
		
		for _, scopeKey := range scopeKeys {
			count := scopeCounts[scopeKey]
			if count < len(compound.Require) {
				continue
			}
			if !compound.AllowMitigated {
				if _, ok := mitigatedScopes[scopeKey]; ok {
					continue
				}
			}
			ref := reference[scopeKey]
			bonus := defaultFloat(compound.Bonus, 2.5, 2.5)
			total += bonus
			out = append(out, occurrence{
				evidence: model.Evidence{
					Kind:      model.EvidenceAmplifier,
					SignalID:  compound.ID,
					Label:     coalesce(compound.Label, "compound signal"),
					Detail:    fmt.Sprintf("compound signals co-occur within %s", scopedLabel(compound.Scope)),
					File:      ref.evidence.File,
					Line:      ref.evidence.Line,
					PathClass: ref.evidence.PathClass,
					Scope:     ref.evidence.Scope,
					ScopeName: ref.evidence.ScopeName,
					Proximity: scopedLabel(compound.Scope),
					Weight:    bonus,
					Source:    "source",
				},
				scopeKey:  ref.scopeKey,
				moduleKey: ref.moduleKey,
			})
			break
		}
	}
	sortOccurrences(out)
	return total, out
}

func contextBonuses(pb model.Playbook, triggers, contextEvidence []occurrence) (float64, float64) {
	blast := 0.0
	hot := 0.0
	for _, trigger := range triggers {
		if trigger.evidence.PathClass == "production" {
			blast += defaultFloat(pb.Scoring.BlastRadiusBonus, 1.5, 1.5)
		}
		if trigger.evidence.PathClass == "test" || trigger.evidence.PathClass == "fixture" || trigger.evidence.PathClass == "example" {
			blast -= 1.0
		}
		if trigger.evidence.Proximity == "same_block" && strings.Contains(strings.ToLower(trigger.evidence.File), "handler") {
			hot += defaultFloat(pb.Scoring.HotPathBonus, 1.0, 1.0)
		}
	}
	for _, item := range contextEvidence {
		if strings.Contains(strings.ToLower(item.evidence.Label), "hot path") {
			hot += item.evidence.Weight
		}
	}
	if blast < 0 {
		blast = 0
	}
	return blast, hot
}

func changeAdjustment(pb model.Playbook, triggers []occurrence, changeSet detectors.ChangeSet) (float64, string) {
	if len(changeSet.ChangedFiles) == 0 {
		return 0, "legacy"
	}
	changeStatus := "legacy"
	bonus := 0.0
	for _, trigger := range triggers {
		change, ok := changeSet.ChangedFiles[trigger.evidence.File]
		if !ok {
			continue
		}
		changeStatus = "modified"
		if change.Status == "added" {
			changeStatus = "introduced"
			bonus += defaultFloat(pb.Source.ChangeSensitivity.NewFileBonus, 1.5, 1.5)
			continue
		}
		if len(change.Lines) == 0 || lineChanged(trigger.evidence.Line, change.Lines) {
			bonus += defaultFloat(pb.Source.ChangeSensitivity.ModifiedLineBonus, 1.0, 1.0)
		}
	}
	if changeStatus == "legacy" {
		return -defaultFloat(pb.Source.ChangeSensitivity.LegacyDiscount, 0.5, 0.5), changeStatus
	}
	return bonus, changeStatus
}

func weightedSum(items, triggers []occurrence) float64 {
	total := 0.0
	for _, item := range items {
		prox := 1.0
		if len(triggers) > 0 && item.evidence.Kind != model.EvidenceTrigger {
			prox = closestProximity(item, triggers)
		}
		total += item.evidence.Weight * prox
	}
	return total
}

func weightedSuppression(items []occurrence) float64 {
	total := 0.0
	for _, item := range items {
		total += item.evidence.Weight
	}
	return total
}

func closestProximity(item occurrence, triggers []occurrence) float64 {
	best := 0.1
	for _, trigger := range triggers {
		score := proximity(trigger, item)
		if score > best {
			best = score
		}
	}
	return best
}

func proximity(a, b occurrence) float64 {
	if a.evidence.File == b.evidence.File {
		if a.evidence.ScopeName != "" && a.evidence.ScopeName == b.evidence.ScopeName {
			if abs(a.evidence.Line-b.evidence.Line) <= 6 {
				return 1.0
			}
			return 0.85
		}
		return 0.55
	}
	if a.moduleKey == b.moduleKey {
		return 0.3
	}
	return 0.1
}

func inferScope(file preparedFile, line preparedLine) (string, string, string) {
	if line.function != "" {
		scopeName := line.function
		if line.depth > 0 {
			return scopeName, file.path + "|" + scopeName, "same_block"
		}
		return scopeName, file.path + "|" + scopeName, "same_function"
	}
	return file.path, file.path, "same_file"
}

func classifyPath(path string) (string, bool, bool, float64) {
	p := strings.ToLower(filepath.ToSlash(path))
	switch {
	case strings.Contains(p, "/testdata/"), strings.Contains(p, "/fixtures/"), strings.Contains(p, "/fixture/"):
		return "fixture", false, false, -2
	case strings.Contains(p, "/examples/"), strings.Contains(p, "/example/"):
		return "example", false, false, -1.5
	case strings.Contains(p, "/migrations/"), strings.Contains(p, "/migration/"):
		return "migration", false, false, -1
	case strings.Contains(p, "/scripts/"), strings.Contains(p, "/bin/"):
		return "script", false, false, -1
	case strings.Contains(p, "/admin/"), strings.Contains(p, "/internal/tools/"):
		return "admin", false, false, -0.75
	case strings.Contains(p, "_test."), strings.Contains(p, "/tests/"), strings.Contains(p, "/test/"), strings.Contains(p, ".spec."), strings.Contains(p, ".test."):
		return "test", false, false, -1.5
	default:
		hot := strings.Contains(p, "/handler") || strings.Contains(p, "/api/") || strings.Contains(p, "/consumer") || strings.Contains(p, "/service")
		return "production", hot, hot, 1
	}
}

func containsPattern(text string, patterns []string) bool {
	norm := normalize(text)
	for _, pattern := range patterns {
		if strings.Contains(norm, normalize(pattern)) {
			return true
		}
	}
	return false
}

func containsAnyPath(path string, patterns []string) bool {
	norm := strings.ToLower(filepath.ToSlash(path))
	for _, pattern := range patterns {
		p := strings.ToLower(filepath.ToSlash(pattern))
		if strings.Contains(norm, p) {
			return true
		}
	}
	return false
}

func pathAllowed(path string, policy model.ContextPolicy, includes, excludes []string) bool {
	if len(policy.PathIncludes) > 0 && !containsAnyPath(path, policy.PathIncludes) {
		return false
	}
	if len(policy.PathExcludes) > 0 && containsAnyPath(path, policy.PathExcludes) {
		return false
	}
	if len(includes) > 0 && !containsAnyPath(path, includes) {
		return false
	}
	if len(excludes) > 0 && containsAnyPath(path, excludes) {
		return false
	}
	return true
}

func normalize(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(s)), " "))
}

func inferFunctionName(line string) string {
	line = strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(line, "func "):
		name := strings.TrimPrefix(line, "func ")
		if strings.HasPrefix(name, "(") {
			if idx := strings.Index(name, ")"); idx >= 0 && idx+1 < len(name) {
				name = strings.TrimSpace(name[idx+1:])
			}
		}
		if idx := strings.Index(name, "("); idx >= 0 {
			return strings.TrimSpace(name[:idx])
		}
	case strings.HasPrefix(line, "function "):
		name := strings.TrimPrefix(line, "function ")
		if idx := strings.Index(name, "("); idx >= 0 {
			return strings.TrimSpace(name[:idx])
		}
	case strings.HasPrefix(line, "const "), strings.HasPrefix(line, "let "), strings.HasPrefix(line, "var "):
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			return strings.TrimSuffix(parts[1], "=")
		}
	}
	return ""
}

func lineChanged(line int, lines map[int]struct{}) bool {
	_, ok := lines[line]
	return ok
}

func scopeHasMitigation(scopeKey string, mitigations []occurrence) bool {
	for _, mitigation := range mitigations {
		if mitigation.scopeKey == scopeKey {
			return true
		}
	}
	return false
}

func compactEvidence(items []occurrence) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item.evidence.Line > 0 {
			out = append(out, fmt.Sprintf("%s:%d %s", item.evidence.File, item.evidence.Line, item.evidence.Label))
		} else {
			out = append(out, fmt.Sprintf("%s %s", item.evidence.File, item.evidence.Label))
		}
	}
	return out
}

func evidenceList(items []occurrence) []model.Evidence {
	out := make([]model.Evidence, 0, len(items))
	for _, item := range items {
		out = append(out, item.evidence)
	}
	return out
}

func buildExplanation(triggers, amplifiers, mitigations, suppressions, context []occurrence, changeStatus string) model.ResultExplanation {
	return model.ResultExplanation{
		TriggeredBy:    compactLabels(triggers),
		AmplifiedBy:    compactLabels(amplifiers),
		MitigatedBy:    compactLabels(mitigations),
		SuppressedBy:   compactLabels(suppressions),
		Contextualized: compactLabels(context),
		ChangeStatus:   changeStatus,
	}
}

func compactLabels(items []occurrence) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		label := item.evidence.Label
		if item.evidence.File != "" {
			label = fmt.Sprintf("%s (%s:%d)", label, item.evidence.File, item.evidence.Line)
		}
		out = append(out, label)
	}
	return out
}

func mergeEvidence(groups ...[]occurrence) []occurrence {
	out := make([]occurrence, 0)
	for _, group := range groups {
		out = append(out, group...)
	}
	sortOccurrences(out)
	return out
}

func sortOccurrences(items []occurrence) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].evidence.File != items[j].evidence.File {
			return items[i].evidence.File < items[j].evidence.File
		}
		if items[i].evidence.Line != items[j].evidence.Line {
			return items[i].evidence.Line < items[j].evidence.Line
		}
		if items[i].evidence.Kind != items[j].evidence.Kind {
			return items[i].evidence.Kind < items[j].evidence.Kind
		}
		return items[i].evidence.Label < items[j].evidence.Label
	})
}

func confidenceFromScore(patternScore, finalScore float64) float64 {
	if finalScore <= 0 {
		return 0
	}
	if patternScore <= 0 {
		return 0.1
	}
	return round(math.Min(1, patternScore/math.Max(patternScore, finalScore)))
}

func scopedKey(scope string, hit occurrence) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "module", "package":
		return hit.moduleKey
	case "file":
		return hit.evidence.File
	default:
		return hit.scopeKey
	}
}

func scopedLabel(scope string) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "module", "package":
		return "same_module"
	case "file":
		return "same_file"
	default:
		return "same_function"
	}
}

func extractField(line, prefix string) string {
	idx := strings.Index(line, prefix)
	if idx < 0 {
		return ""
	}
	value := line[idx+len(prefix):]
	if end := strings.IndexAny(value, " \t"); end >= 0 {
		value = value[:end]
	}
	return strings.TrimSpace(value)
}

func coalesce(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func defaultFloat(values ...float64) float64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func defaultInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func round(v float64) float64 {
	return math.Round(v*100) / 100
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
