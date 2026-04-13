package sourcedetector

import (
	"testing"

	"faultline/internal/detectors"
	"faultline/internal/model"
)

func TestKind(t *testing.T) {
	d := Detector{}
	if d.Kind() != detectors.KindSource {
		t.Errorf("Kind() = %q, want %q", d.Kind(), detectors.KindSource)
	}
}

func TestDetectNoTriggerReturnsEmpty(t *testing.T) {
	pb := model.Playbook{
		ID:       "test.src-empty",
		Category: "build",
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "t1", Patterns: []string{"panic("}, Weight: 1.0},
			},
		},
	}
	target := detectors.Target{
		Files: []detectors.SourceFile{
			{Path: "main.go", Lines: []string{"func main() {", "fmt.Println(hello)", "}"}},
		},
	}
	results := Detector{}.Detect([]model.Playbook{pb}, target)
	if len(results) != 0 {
		t.Errorf("expected no results without trigger match, got %d", len(results))
	}
}

func TestDetectTriggerMatchReturnsResult(t *testing.T) {
	pb := model.Playbook{
		ID:        "test.src-panic",
		Category:  "runtime",
		BaseScore: 0.6,
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "t1", Patterns: []string{"panic("}, Weight: 1.0},
			},
		},
	}
	target := detectors.Target{
		Files: []detectors.SourceFile{
			{
				Path:  "server.go",
				Lines: []string{"func serve() {", "\tpanic(unreachable)", "}"},
			},
		},
	}
	results := Detector{}.Detect([]model.Playbook{pb}, target)
	if len(results) == 0 {
		t.Fatal("expected at least one result, got none")
	}
	if results[0].Playbook.ID != "test.src-panic" {
		t.Errorf("result ID = %q, want %q", results[0].Playbook.ID, "test.src-panic")
	}
	if results[0].Score == 0 {
		t.Error("expected non-zero score")
	}
}

func TestDetectResultsSortedByScoreDesc(t *testing.T) {
	pbLow := model.Playbook{
		ID:        "test.src-low",
		Category:  "build",
		BaseScore: 0.1,
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "low", Patterns: []string{"//TODO"}, Weight: 0.1},
			},
		},
	}
	pbHigh := model.Playbook{
		ID:        "test.src-high",
		Category:  "runtime",
		BaseScore: 1.0,
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "high", Patterns: []string{"panic("}, Weight: 1.0},
			},
		},
	}
	target := detectors.Target{
		Files: []detectors.SourceFile{
			{
				Path:  "main.go",
				Lines: []string{"//TODO fix me", "panic(oops)"},
			},
		},
	}
	results := Detector{}.Detect([]model.Playbook{pbLow, pbHigh}, target)
	if len(results) < 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Score < results[1].Score {
		t.Errorf("results not sorted desc: scores %v, %v", results[0].Score, results[1].Score)
	}
	if results[0].Playbook.ID != "test.src-high" {
		t.Errorf("top result = %q, want test.src-high", results[0].Playbook.ID)
	}
}

func TestDetectEmptyFileListReturnsEmpty(t *testing.T) {
	pb := model.Playbook{
		ID:       "test.src-nofiles",
		Category: "build",
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "t1", Patterns: []string{"panic("}, Weight: 1.0},
			},
		},
	}
	target := detectors.Target{Files: nil}
	results := Detector{}.Detect([]model.Playbook{pb}, target)
	if len(results) != 0 {
		t.Errorf("expected no results for empty file list, got %d", len(results))
	}
}

func TestDetectEmptyPlaybookListReturnsEmpty(t *testing.T) {
	target := detectors.Target{
		Files: []detectors.SourceFile{
			{Path: "main.go", Lines: []string{"panic(oops)"}},
		},
	}
	results := Detector{}.Detect(nil, target)
	if len(results) != 0 {
		t.Errorf("expected no results for empty playbook list, got %d", len(results))
	}
}
