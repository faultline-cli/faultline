package engine

import (
	"reflect"
	"strings"
	"testing"

	"faultline/internal/matcher"
	"faultline/internal/model"
	"faultline/internal/playbooks"
)

func TestBundledPlaybooksSmoke(t *testing.T) {
	pbs, err := playbooks.LoadDir(repoPlaybookDir(t))
	if err != nil {
		t.Fatalf("load playbooks: %v", err)
	}
	if len(pbs) == 0 {
		t.Fatal("expected bundled playbooks to load")
	}

	for _, pb := range pbs {
		pb := pb
		t.Run(pb.ID, func(t *testing.T) {
			if pb.Detector == "source" {
				t.Skip("source playbooks are covered by repository inspection tests")
			}
			log := fixtureLogForPlaybook(pb)
			lines, err := readLines(strings.NewReader(log))
			if err != nil {
				t.Fatalf("read lines: %v", err)
			}
			if len(lines) == 0 {
				t.Fatalf("expected fixture log for %s to produce lines", pb.ID)
			}

			ctx := ExtractContext(lines)
			results := matcher.Rank(pbs, lines, ctx)
			if len(results) == 0 {
				t.Fatalf("expected %s fixture to match at least one playbook", pb.ID)
			}
			if !containsPlaybook(results, pb.ID) {
				t.Fatalf(
					"expected fixture for %s to match that playbook, got top result %s with results %+v",
					pb.ID,
					results[0].Playbook.ID,
					resultIDs(results),
				)
			}

			// The same input must rank identically on repeated runs.
			again := matcher.Rank(pbs, lines, ctx)
			if !reflect.DeepEqual(results, again) {
				t.Fatalf("expected deterministic ranking for %s", pb.ID)
			}
		})
	}
}

func fixtureLogForPlaybook(pb model.Playbook) string {
	switch {
	case len(pb.Match.Any) > 0:
		return strongestPattern(pb.Match.Any)
	case len(pb.Match.All) > 0:
		return strings.Join(pb.Match.All, "\n")
	default:
		return pb.Title
	}
}

func strongestPattern(patterns []string) string {
	if len(patterns) == 0 {
		return ""
	}
	best := strings.TrimSpace(patterns[0])
	for _, pattern := range patterns[1:] {
		candidate := strings.TrimSpace(pattern)
		if len(candidate) > len(best) {
			best = candidate
		}
	}
	return best
}

func containsPlaybook(results []model.Result, id string) bool {
	for _, r := range results {
		if r.Playbook.ID == id {
			return true
		}
	}
	return false
}

func resultIDs(results []model.Result) []string {
	ids := make([]string, 0, len(results))
	for _, r := range results {
		ids = append(ids, r.Playbook.ID)
	}
	return ids
}
