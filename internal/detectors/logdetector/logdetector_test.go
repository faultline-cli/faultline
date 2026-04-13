package logdetector

import (
	"testing"

	"faultline/internal/detectors"
	"faultline/internal/model"
)

func TestKind(t *testing.T) {
	d := Detector{}
	if d.Kind() != detectors.KindLog {
		t.Errorf("Kind() = %q, want %q", d.Kind(), detectors.KindLog)
	}
}

func TestDetectMatchesAnyPattern(t *testing.T) {
	pb := model.Playbook{
		ID:        "test.npm-enoent",
		Title:     "npm missing package",
		Category:  "dependency",
		BaseScore: 0.9,
		Match: model.MatchSpec{
			Any: []string{"npm ERR code ENOENT"},
		},
	}
	lines := []model.Line{
		{Original: "npm ERR code ENOENT", Normalized: "npm err code enoent", Number: 1},
	}
	target := detectors.Target{LogLines: lines}
	results := Detector{}.Detect([]model.Playbook{pb}, target)
	if len(results) == 0 {
		t.Fatal("expected at least one result, got none")
	}
	if results[0].Playbook.ID != "test.npm-enoent" {
		t.Errorf("first result ID = %q, want %q", results[0].Playbook.ID, "test.npm-enoent")
	}
}

func TestDetectNoMatchReturnsEmpty(t *testing.T) {
	pb := model.Playbook{
		ID:       "test.docker-push",
		Category: "docker",
		Match: model.MatchSpec{
			Any: []string{"denied: access forbidden"},
		},
	}
	lines := []model.Line{
		{Original: "Build succeeded", Normalized: "build succeeded", Number: 1},
	}
	target := detectors.Target{LogLines: lines}
	results := Detector{}.Detect([]model.Playbook{pb}, target)
	if len(results) != 0 {
		t.Errorf("expected no results, got %d", len(results))
	}
}

func TestDetectNonePatternExcludesMatch(t *testing.T) {
	pb := model.Playbook{
		ID:        "test.go-build",
		Category:  "build",
		BaseScore: 0.8,
		Match: model.MatchSpec{
			Any:  []string{"go build failed"},
			None: []string{"go vet ok"},
		},
	}
	// Log has the Any pattern AND the None pattern — None should suppress the match.
	lines := []model.Line{
		{Original: "go build failed", Normalized: "go build failed", Number: 1},
		{Original: "go vet ok", Normalized: "go vet ok", Number: 2},
	}
	target := detectors.Target{LogLines: lines}
	results := Detector{}.Detect([]model.Playbook{pb}, target)
	if len(results) != 0 {
		t.Errorf("None pattern should suppress match; got %d result(s)", len(results))
	}
}

// TestDetectAllPatternCompoundBonus verifies that matching all All-patterns
// yields a higher score than matching only a subset (compound bonus applied).
func TestDetectAllPatternCompoundBonus(t *testing.T) {
	pb := model.Playbook{
		ID:        "test.two-errors",
		Category:  "build",
		BaseScore: 0.7,
		Match: model.MatchSpec{
			All: []string{"error: undefined symbol", "linker failed"},
		},
	}
	partial := []model.Line{
		{Original: "error: undefined symbol", Normalized: "error: undefined symbol", Number: 1},
	}
	partialTarget := detectors.Target{LogLines: partial}
	partialRes := Detector{}.Detect([]model.Playbook{pb}, partialTarget)

	full := []model.Line{
		{Original: "error: undefined symbol", Normalized: "error: undefined symbol", Number: 1},
		{Original: "linker failed", Normalized: "linker failed", Number: 2},
	}
	fullTarget := detectors.Target{LogLines: full}
	fullRes := Detector{}.Detect([]model.Playbook{pb}, fullTarget)

	if len(fullRes) == 0 {
		t.Fatal("All full: expected match, got none")
	}
	// The complete match should score strictly higher due to compound bonus.
	if len(partialRes) > 0 && partialRes[0].Score >= fullRes[0].Score {
		t.Errorf("partial score (%v) should be < full score (%v)", partialRes[0].Score, fullRes[0].Score)
	}
}

func TestDetectEmptyPlaybookListReturnsEmpty(t *testing.T) {
	lines := []model.Line{
		{Original: "error: something bad", Normalized: "error: something bad", Number: 1},
	}
	results := Detector{}.Detect(nil, detectors.Target{LogLines: lines})
	if len(results) != 0 {
		t.Errorf("expected no results for empty playbook list, got %d", len(results))
	}
}
