package engine

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"faultline/internal/matcher"
	"faultline/internal/model"
	"faultline/internal/playbooks"
)

func TestBundledPlaybooksRepresentativeRanking(t *testing.T) {
	pbs, err := playbooks.LoadDir(repoPlaybookDir(t))
	if err != nil {
		t.Fatalf("load playbooks: %v", err)
	}

	index := newPatternIndex(pbs)
	for _, pb := range pbs {
		pb := pb
		t.Run(pb.ID, func(t *testing.T) {
			if pb.Detector == "source" {
				t.Skip("source playbooks are covered by dedicated repository fixture tests")
			}
			log := representativeLogForPlaybook(pb, index, "")
			results := rankRepresentativeLog(t, pbs, log, false)
			assertRepresentativeResult(t, results, pb.ID, hasExclusivePositiveSignal(pb, index))
		})
	}
}

func TestBundledPlaybookConflictScenarios(t *testing.T) {
	pbs, err := playbooks.LoadDir(repoPlaybookDir(t))
	if err != nil {
		t.Fatalf("load playbooks: %v", err)
	}

	index := newPatternIndex(pbs)
	byID := make(map[string]model.Playbook, len(pbs))
	for _, pb := range pbs {
		byID[pb.ID] = pb
	}

	for _, conflict := range playbooks.FindPatternConflicts(pbs) {
		for _, ref := range conflict.Positive {
			pb := byID[ref.PlaybookID]
			name := fmt.Sprintf("positive/%s/%s", conflict.Pattern, ref.PlaybookID)
			t.Run(name, func(t *testing.T) {
				log := representativeLogForPlaybook(pb, index, conflict.Pattern)
				results := rankRepresentativeLog(t, pbs, log, false)
				assertRepresentativeResult(t, results, pb.ID, hasExclusivePositiveSignal(pb, index))
			})
		}

		for _, ref := range conflict.Negative {
			pb := byID[ref.PlaybookID]
			name := fmt.Sprintf("negative/%s/%s", conflict.Pattern, ref.PlaybookID)
			t.Run(name, func(t *testing.T) {
				log := representativeLogForPlaybook(pb, index, conflict.Pattern)
				results := rankRepresentativeLog(t, pbs, log, true)
				if containsPlaybook(results, pb.ID) {
					t.Fatalf("expected %s to be excluded by shared pattern %q, got %v", pb.ID, conflict.Pattern, resultIDs(results))
				}
			})
		}
	}
}

type patternIndex struct {
	positiveCounts map[string]int
	negativeCounts map[string]int
}

func newPatternIndex(pbs []model.Playbook) patternIndex {
	index := patternIndex{
		positiveCounts: make(map[string]int),
		negativeCounts: make(map[string]int),
	}
	for _, pb := range pbs {
		seenPositive := make(map[string]struct{})
		for _, pattern := range append(append([]string{}, pb.Match.Any...), pb.Match.All...) {
			norm := normalizeRepresentativePattern(pattern)
			if norm == "" {
				continue
			}
			if _, ok := seenPositive[norm]; ok {
				continue
			}
			seenPositive[norm] = struct{}{}
			index.positiveCounts[norm]++
		}
		seenNegative := make(map[string]struct{})
		for _, pattern := range pb.Match.None {
			norm := normalizeRepresentativePattern(pattern)
			if norm == "" {
				continue
			}
			if _, ok := seenNegative[norm]; ok {
				continue
			}
			seenNegative[norm] = struct{}{}
			index.negativeCounts[norm]++
		}
	}
	return index
}

func representativeLogForPlaybook(pb model.Playbook, index patternIndex, requiredPattern string) string {
	lines := make([]string, 0, 8)
	if stageLine := stageLineForPlaybook(pb); stageLine != "" {
		lines = append(lines, stageLine)
	}
	lines = append(lines, "## Step: deterministic playbook coverage")
	lines = append(lines, filterPatternsAgainstNone(selectRepresentativePatterns(pb.Match.All, index, 0), pb.Match.None)...)

	selectedAny := filterPatternsAgainstNone(selectRepresentativePatterns(pb.Match.Any, index, 0), pb.Match.None)
	if len(selectedAny) == 0 && len(pb.Match.Any) > 0 {
		selectedAny = filterPatternsAgainstNone(selectLongestPatterns(pb.Match.Any, 0), pb.Match.None)
	}
	lines = append(lines, selectedAny...)

	if strings.TrimSpace(requiredPattern) != "" && !containsNormalizedPattern(lines, requiredPattern) {
		lines = append(lines, requiredPattern)
	}

	return strings.Join(dedupeKeepOrderStrings(lines), "\n") + "\n"
}

func selectRepresentativePatterns(patterns []string, index patternIndex, limit int) []string {
	if len(patterns) == 0 {
		return nil
	}
	unique := make([]string, 0, len(patterns))
	shared := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		norm := normalizeRepresentativePattern(pattern)
		if norm == "" {
			continue
		}
		if index.positiveCounts[norm] == 1 {
			unique = append(unique, pattern)
			continue
		}
		shared = append(shared, pattern)
	}
	uniqueSelected := selectLongestPatterns(unique, limit)
	remaining := limit - len(uniqueSelected)
	if limit == 0 {
		remaining = 0
	}
	selected := append(uniqueSelected, selectLongestPatterns(shared, remaining)...)
	if limit == 0 || len(selected) <= limit {
		return dedupeKeepOrderStrings(selected)
	}
	return dedupeKeepOrderStrings(selected[:limit])
}

func selectLongestPatterns(patterns []string, limit int) []string {
	if len(patterns) == 0 {
		return nil
	}
	copyPatterns := append([]string(nil), patterns...)
	sort.SliceStable(copyPatterns, func(i, j int) bool {
		left := strings.TrimSpace(copyPatterns[i])
		right := strings.TrimSpace(copyPatterns[j])
		if len(left) != len(right) {
			return len(left) > len(right)
		}
		return left < right
	})
	if limit <= 0 || limit > len(copyPatterns) {
		limit = len(copyPatterns)
	}
	return copyPatterns[:limit]
}

func stageLineForPlaybook(pb model.Playbook) string {
	for _, hint := range pb.StageHints {
		switch strings.ToLower(strings.TrimSpace(hint)) {
		case stageBuild:
			return "$ go build ./..."
		case stageTest:
			return "$ go test ./..."
		case stageDeploy:
			return "$ kubectl apply -f deploy.yaml"
		}
	}
	return ""
}

func rankRepresentativeLog(t *testing.T, pbs []model.Playbook, log string, allowNoMatch bool) []model.Result {
	t.Helper()
	lines, err := readLines(strings.NewReader(log))
	if err != nil {
		t.Fatalf("read lines: %v", err)
	}
	ctx := ExtractContext(lines)
	results := matcher.Rank(pbs, lines, ctx)
	if len(results) == 0 && !allowNoMatch {
		t.Fatalf("expected representative log to match at least one playbook, got log:\n%s", log)
	}
	again := matcher.Rank(pbs, lines, ctx)
	if len(results) != len(again) {
		t.Fatalf("expected deterministic result count, got %d and %d", len(results), len(again))
	}
	for i := range results {
		if results[i].Playbook.ID != again[i].Playbook.ID ||
			results[i].Score != again[i].Score ||
			results[i].Confidence != again[i].Confidence {
			t.Fatalf("expected deterministic ranking for representative log, got %v and %v", resultIDs(results), resultIDs(again))
		}
	}
	return results
}

func assertRepresentativeResult(t *testing.T, results []model.Result, wantID string, requireTop bool) {
	t.Helper()
	idx := -1
	for i, result := range results {
		if result.Playbook.ID == wantID {
			idx = i
			break
		}
	}
	if idx == -1 {
		t.Fatalf("expected %s to appear in ranked results, got %v", wantID, resultIDs(results))
	}
	if requireTop && (results[idx].Score != results[0].Score || results[idx].Confidence != results[0].Confidence) {
		t.Fatalf("expected %s to be top-ranked or tied for top, got top=%s score=%.2f confidence=%.2f target-score=%.2f target-confidence=%.2f", wantID, results[0].Playbook.ID, results[0].Score, results[0].Confidence, results[idx].Score, results[idx].Confidence)
	}
	if len(results[idx].Evidence) == 0 {
		t.Fatalf("expected evidence for %s", wantID)
	}
}

func hasExclusivePositiveSignal(pb model.Playbook, index patternIndex) bool {
	for _, pattern := range append(append([]string{}, pb.Match.Any...), pb.Match.All...) {
		if index.positiveCounts[normalizeRepresentativePattern(pattern)] == 1 {
			return true
		}
	}
	return false
}

func filterPatternsAgainstNone(patterns, none []string) []string {
	if len(none) == 0 {
		return patterns
	}
	blocked := make([]string, 0, len(none))
	for _, pattern := range none {
		norm := normalizeRepresentativePattern(pattern)
		if norm != "" {
			blocked = append(blocked, norm)
		}
	}
	filtered := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		norm := normalizeRepresentativePattern(pattern)
		if norm == "" {
			continue
		}
		skip := false
		for _, blockedPattern := range blocked {
			if strings.Contains(norm, blockedPattern) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		filtered = append(filtered, pattern)
	}
	return filtered
}

func dedupeKeepOrderStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		norm := normalizeRepresentativePattern(value)
		if norm == "" {
			continue
		}
		if _, ok := seen[norm]; ok {
			continue
		}
		seen[norm] = struct{}{}
		result = append(result, strings.TrimSpace(value))
	}
	return result
}

func containsNormalizedPattern(lines []string, pattern string) bool {
	norm := normalizeRepresentativePattern(pattern)
	for _, line := range lines {
		if normalizeRepresentativePattern(line) == norm {
			return true
		}
	}
	return false
}

func normalizeRepresentativePattern(pattern string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(pattern)), " "))
}
