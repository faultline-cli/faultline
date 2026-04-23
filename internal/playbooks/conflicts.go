package playbooks

import (
	"fmt"
	"sort"
	"strings"

	"faultline/internal/model"
)

// PatternRef identifies where a normalized pattern appears.
type PatternRef struct {
	PlaybookID string
	Section    string
	Pattern    string
}

// PatternConflict groups positive and negative references for the same
// normalized pattern across bundled playbooks.
type PatternConflict struct {
	Pattern  string
	Positive []PatternRef
	Negative []PatternRef
}

// FindPatternConflicts reports patterns that are shared across playbooks or
// explicitly excluded by another playbook's match.none selector.
func FindPatternConflicts(pbs []model.Playbook) []PatternConflict {
	type bucket struct {
		positive []PatternRef
		negative []PatternRef
	}

	byPattern := make(map[string]*bucket)
	seen := make(map[string]struct{})
	add := func(pattern, playbookID, section string, positive bool) {
		norm := normalizePattern(pattern)
		if norm == "" {
			return
		}
		seenKey := fmt.Sprintf("%t|%s|%s|%s", positive, playbookID, section, norm)
		if _, ok := seen[seenKey]; ok {
			return
		}
		seen[seenKey] = struct{}{}
		b := byPattern[norm]
		if b == nil {
			b = &bucket{}
			byPattern[norm] = b
		}
		ref := PatternRef{
			PlaybookID: playbookID,
			Section:    section,
			Pattern:    pattern,
		}
		if positive {
			b.positive = append(b.positive, ref)
			return
		}
		b.negative = append(b.negative, ref)
	}

	for _, pb := range pbs {
		for _, pattern := range pb.Match.Any {
			add(pattern, pb.ID, "match.any", true)
		}
		for _, pattern := range pb.Match.All {
			add(pattern, pb.ID, "match.all", true)
		}
		for _, pattern := range pb.Match.None {
			add(pattern, pb.ID, "match.none", false)
		}
		for i, group := range pb.Match.Partial {
			section := fmt.Sprintf("match.partial[%d]", i)
			for _, pattern := range group.Patterns {
				add(pattern, pb.ID, section, true)
			}
		}
	}

	conflicts := make([]PatternConflict, 0)
	for pattern, b := range byPattern {
		positiveIDs := uniquePlaybookIDs(b.positive)
		if len(positiveIDs) > 1 || (len(positiveIDs) > 0 && len(b.negative) > 0) {
			sortRefs(b.positive)
			sortRefs(b.negative)
			conflicts = append(conflicts, PatternConflict{
				Pattern:  pattern,
				Positive: append([]PatternRef(nil), b.positive...),
				Negative: append([]PatternRef(nil), b.negative...),
			})
		}
	}

	sort.Slice(conflicts, func(i, j int) bool {
		return conflicts[i].Pattern < conflicts[j].Pattern
	})
	return conflicts
}

// FormatPatternConflicts renders a deterministic, human-readable report.
func FormatPatternConflicts(conflicts []PatternConflict) string {
	if len(conflicts) == 0 {
		return "No pattern conflicts detected.\n"
	}
	var b strings.Builder
	for i, c := range conflicts {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "%s\n", c.Pattern)
		for _, ref := range c.Positive {
			fmt.Fprintf(&b, "  + %s (%s)\n", ref.PlaybookID, ref.Section)
		}
		for _, ref := range c.Negative {
			fmt.Fprintf(&b, "  - %s (%s)\n", ref.PlaybookID, ref.Section)
		}
	}
	return b.String()
}

func uniquePlaybookIDs(refs []PatternRef) []string {
	seen := make(map[string]struct{}, len(refs))
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		if _, ok := seen[ref.PlaybookID]; ok {
			continue
		}
		seen[ref.PlaybookID] = struct{}{}
		ids = append(ids, ref.PlaybookID)
	}
	sort.Strings(ids)
	return ids
}

func sortRefs(refs []PatternRef) {
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].PlaybookID != refs[j].PlaybookID {
			return refs[i].PlaybookID < refs[j].PlaybookID
		}
		if refs[i].Section != refs[j].Section {
			return refs[i].Section < refs[j].Section
		}
		return refs[i].Pattern < refs[j].Pattern
	})
}
