package sourcedetector

import (
	"testing"

	"faultline/internal/detectors"
	"faultline/internal/model"
)

func TestLineChangedReturnsTrueForKnownLine(t *testing.T) {
	lines := map[int]struct{}{1: {}, 5: {}, 10: {}}
	if !lineChanged(5, lines) {
		t.Error("expected lineChanged(5) = true")
	}
}

func TestLineChangedReturnsFalseForUnknownLine(t *testing.T) {
	lines := map[int]struct{}{1: {}, 5: {}}
	if lineChanged(3, lines) {
		t.Error("expected lineChanged(3) = false")
	}
}

func TestCollectConsistencyNoRulesReturnsNil(t *testing.T) {
	pb := model.Playbook{ID: "no-consistency"}
	result := collectConsistency(pb, nil, nil)
	if result != nil {
		t.Fatalf("expected nil for playbook without local_consistency, got %#v", result)
	}
}

func TestCollectConsistencyReturnsAmplifierWhenPeersLackMitigation(t *testing.T) {
	pb := model.Playbook{
		ID: "consistency-test",
		Source: model.SourceSpec{
			LocalConsistency: []model.ConsistencyRule{
				{ID: "c1", Label: "missing safeguard", MinimumPeers: 2, Threshold: 0.6, Amplifier: 1.5},
			},
		},
	}

	// Three trigger scopes in the same module.
	// func1 and func2 have mitigations. func3 does not.
	// ratio = 2/3 ≈ 0.67 >= 0.6 threshold; func3 has no mitigation so it qualifies.
	triggers := []occurrence{
		{moduleKey: "modA", scopeKey: "modA|func1", evidence: model.Evidence{File: "a.go", Line: 1}},
		{moduleKey: "modA", scopeKey: "modA|func2", evidence: model.Evidence{File: "a.go", Line: 10}},
		{moduleKey: "modA", scopeKey: "modA|func3", evidence: model.Evidence{File: "a.go", Line: 20}},
	}
	// func1 and func2 have an expected mitigation; func3 does not
	mitigations := []occurrence{
		{moduleKey: "modA", scopeKey: "modA|func1"},
		{moduleKey: "modA", scopeKey: "modA|func2"},
	}

	results := collectConsistency(pb, triggers, mitigations)
	if len(results) == 0 {
		t.Fatal("expected at least one consistency amplifier")
	}
	if results[0].evidence.Kind != model.EvidenceAmplifier {
		t.Fatalf("expected EvidenceAmplifier kind, got %q", results[0].evidence.Kind)
	}
}

func TestRequiredCompoundWithoutEvidenceSuppressesResult(t *testing.T) {
	pb := model.Playbook{
		ID:        "required-compound",
		BaseScore: 1,
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "trigger", Patterns: []string{"danger()"}, Weight: 1},
			},
			Amplifiers: []model.SignalMatcher{
				{ID: "context", Patterns: []string{"critical"}, Weight: 1},
			},
			Mitigations: []model.SignalMatcher{
				{ID: "safe", Patterns: []string{"recover()"}, Weight: 3},
			},
			CompoundSignals: []model.CompoundSignal{
				{ID: "required", Require: []string{"trigger", "context"}, Scope: "file", Bonus: 2, Required: true},
			},
		},
	}
	target := detectors.Target{
		Files: []detectors.SourceFile{
			{
				Path: "handler.go",
				Lines: []string{
					"func handler() {",
					"  // critical path",
					"  recover()",
					"  danger()",
					"}",
				},
			},
		},
	}

	results := Detector{}.Detect([]model.Playbook{pb}, target)
	if len(results) != 0 {
		t.Fatalf("expected required compound to suppress mitigated result, got %#v", results)
	}
}

func TestCompoundRequiresDistinctSignalsInScope(t *testing.T) {
	pb := model.Playbook{
		ID:        "distinct-compound-signals",
		BaseScore: 1,
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "trigger", Patterns: []string{"#!/bin/sh"}, Weight: 1},
			},
			CompoundSignals: []model.CompoundSignal{
				{ID: "compound", Require: []string{"trigger", "missing"}, Scope: "file", Bonus: 2, Required: true},
			},
		},
	}
	target := detectors.Target{
		Files: []detectors.SourceFile{
			{
				Path: "install.sh",
				Lines: []string{
					"#!/bin/sh",
					"cat <<EOF",
					"#!/bin/sh",
					"EOF",
				},
			},
		},
	}

	results := Detector{}.Detect([]model.Playbook{pb}, target)
	if len(results) != 0 {
		t.Fatalf("expected duplicate trigger hits not to satisfy missing compound signal, got %#v", results)
	}
}

func TestSafeContextPatternsApplyToMatchingLines(t *testing.T) {
	pb := model.Playbook{
		ID:        "safe-pattern",
		BaseScore: 1,
		Source: model.SourceSpec{
			Triggers: []model.SignalMatcher{
				{ID: "trigger", Patterns: []string{"danger()"}, Weight: 1},
			},
			SafeContextClasses: []model.SafeContextRule{
				{ID: "safe", Label: "documented safe context", Patterns: []string{"optional"}, Discount: 1.5},
			},
		},
	}
	target := detectors.Target{
		Files: []detectors.SourceFile{
			{Path: "script.sh", Lines: []string{"danger() # optional probe"}},
		},
	}

	results := Detector{}.Detect([]model.Playbook{pb}, target)
	if len(results) == 0 {
		t.Fatal("expected result with safe context evidence")
	}
	if len(results[0].EvidenceBy.Context) == 0 {
		t.Fatalf("expected safe context evidence, got %#v", results[0].EvidenceBy)
	}
}

func TestCollectConsistencySkipsWhenBelowMinimumPeers(t *testing.T) {
	pb := model.Playbook{
		ID: "consistency-min-peers",
		Source: model.SourceSpec{
			LocalConsistency: []model.ConsistencyRule{
				{ID: "c1", MinimumPeers: 3, Threshold: 0.6},
			},
		},
	}
	// Only two scopes — below the required minimum of 3
	triggers := []occurrence{
		{moduleKey: "modB", scopeKey: "modB|func1", evidence: model.Evidence{File: "b.go", Line: 1}},
		{moduleKey: "modB", scopeKey: "modB|func2", evidence: model.Evidence{File: "b.go", Line: 5}},
	}
	results := collectConsistency(pb, triggers, nil)
	if len(results) != 0 {
		t.Fatalf("expected no results below minimum peers, got %#v", results)
	}
}

func TestChangeAdjustmentLegacyWhenNoChangedFiles(t *testing.T) {
	pb := model.Playbook{ID: "change-adj"}
	triggers := []occurrence{
		{evidence: model.Evidence{File: "main.go", Line: 5}},
	}
	bonus, status := changeAdjustment(pb, triggers, detectors.ChangeSet{})
	if status != "legacy" {
		t.Fatalf("expected legacy status, got %q", status)
	}
	// When changeSet is empty the function returns early with (0, "legacy")
	if bonus != 0 {
		t.Fatalf("expected 0 bonus for early-return legacy path, got %.2f", bonus)
	}
}

func TestChangeAdjustmentLegacyDiscountForUnchangedFile(t *testing.T) {
	pb := model.Playbook{ID: "change-adj-discount"}
	triggers := []occurrence{
		{evidence: model.Evidence{File: "old.go", Line: 1}},
	}
	// changeSet has a file, but it's not the trigger file → legacy discount
	changeSet := detectors.ChangeSet{
		ChangedFiles: map[string]detectors.FileChange{
			"new.go": {Status: "modified"},
		},
	}
	bonus, status := changeAdjustment(pb, triggers, changeSet)
	if status != "legacy" {
		t.Fatalf("expected legacy status, got %q", status)
	}
	if bonus >= 0 {
		t.Fatalf("expected negative legacy discount, got %.2f", bonus)
	}
}

func TestChangeAdjustmentModifiedFileGetBonus(t *testing.T) {
	pb := model.Playbook{ID: "change-adj-mod"}
	triggers := []occurrence{
		{evidence: model.Evidence{File: "service.go", Line: 10}},
	}
	changeSet := detectors.ChangeSet{
		ChangedFiles: map[string]detectors.FileChange{
			"service.go": {Status: "modified", Lines: map[int]struct{}{10: {}}},
		},
	}
	bonus, status := changeAdjustment(pb, triggers, changeSet)
	if status != "modified" {
		t.Fatalf("expected modified status, got %q", status)
	}
	if bonus <= 0 {
		t.Fatalf("expected positive bonus for modified+changed line, got %.2f", bonus)
	}
}

func TestChangeAdjustmentAddedFileGetHigherBonus(t *testing.T) {
	pb := model.Playbook{ID: "change-adj-add"}
	triggers := []occurrence{
		{evidence: model.Evidence{File: "new_service.go", Line: 1}},
	}
	changeSet := detectors.ChangeSet{
		ChangedFiles: map[string]detectors.FileChange{
			"new_service.go": {Status: "added"},
		},
	}
	bonus, status := changeAdjustment(pb, triggers, changeSet)
	if status != "introduced" {
		t.Fatalf("expected introduced status, got %q", status)
	}
	if bonus <= 0 {
		t.Fatalf("expected positive bonus for added file, got %.2f", bonus)
	}
}
