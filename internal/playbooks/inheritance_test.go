package playbooks

import (
	"strings"
	"testing"

	"faultline/internal/model"
)

// helper builds a minimal Playbook for use in inheritance tests.
func pbWith(id, extends string) model.Playbook {
	return model.Playbook{
		ID:      id,
		Extends: extends,
	}
}

// TestResolvePlaybookInheritanceNoCycles verifies that a cycle in the extends
// graph is rejected at load time.
func TestResolvePlaybookInheritanceCycleDetected(t *testing.T) {
	pbs := []model.Playbook{
		{ID: "a", Extends: "b"},
		{ID: "b", Extends: "a"},
	}
	_, err := resolvePlaybookInheritance(pbs)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("expected 'cycle' in error, got: %v", err)
	}
}

// TestResolvePlaybookInheritanceMissingParent verifies that referencing an
// unknown parent is rejected.
func TestResolvePlaybookInheritanceMissingParent(t *testing.T) {
	pbs := []model.Playbook{
		{ID: "child", Extends: "nonexistent-base"},
	}
	_, err := resolvePlaybookInheritance(pbs)
	if err == nil {
		t.Fatal("expected error for unknown parent, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent-base") {
		t.Errorf("expected parent id in error, got: %v", err)
	}
}

// TestResolvePlaybookInheritanceNoExtends verifies that playbooks with no
// extends field pass through unchanged.
func TestResolvePlaybookInheritanceNoExtends(t *testing.T) {
	pbs := []model.Playbook{
		{ID: "standalone", Title: "Standalone", Summary: "unchanged"},
	}
	got, err := resolvePlaybookInheritance(pbs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "standalone" {
		t.Errorf("unexpected result: %+v", got)
	}
	if got[0].Summary != "unchanged" {
		t.Errorf("summary should be unchanged, got %q", got[0].Summary)
	}
}

// TestResolvePlaybookInheritanceEmpty verifies that an empty slice returns nil
// without error.
func TestResolvePlaybookInheritanceEmpty(t *testing.T) {
	got, err := resolvePlaybookInheritance(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for empty input, got %v", got)
	}
}

// TestMergePlaybooksChildOverridesTitle tests that child title takes precedence
// over base when explicitly set.
func TestMergePlaybooksChildOverridesTitle(t *testing.T) {
	base := model.Playbook{ID: "base", Title: "Base title", Summary: "Base summary"}
	child := model.Playbook{ID: "child", Extends: "base", Title: "Child title"}

	merged := mergePlaybooks(base, child)

	if merged.Title != "Child title" {
		t.Errorf("expected child title, got %q", merged.Title)
	}
	// Summary not set in child; should inherit from base.
	if merged.Summary != "Base summary" {
		t.Errorf("expected base summary, got %q", merged.Summary)
	}
}

// TestMergePlaybooksInheritsSummaryWhenChildEmpty tests that an empty child
// summary falls back to the base summary.
func TestMergePlaybooksInheritsSummaryWhenChildEmpty(t *testing.T) {
	base := model.Playbook{ID: "base", Summary: "Base summary"}
	child := model.Playbook{ID: "child", Extends: "base"}

	merged := mergePlaybooks(base, child)

	if merged.Summary != "Base summary" {
		t.Errorf("expected inherited summary, got %q", merged.Summary)
	}
}

// TestMergePlaybooksMatchAnyUnioned tests that match.any patterns from both
// base and child are combined without duplicates.
func TestMergePlaybooksMatchAnyUnioned(t *testing.T) {
	base := model.Playbook{
		ID:    "base",
		Match: model.MatchSpec{Any: []string{"base pattern", "shared pattern"}},
	}
	child := model.Playbook{
		ID:      "child",
		Extends: "base",
		Match:   model.MatchSpec{Any: []string{"child pattern", "shared pattern"}},
	}

	merged := mergePlaybooks(base, child)

	wantPatterns := map[string]bool{
		"base pattern":   true,
		"shared pattern": true,
		"child pattern":  true,
	}
	if len(merged.Match.Any) != 3 {
		t.Errorf("expected 3 unique patterns, got %d: %v", len(merged.Match.Any), merged.Match.Any)
	}
	for _, p := range merged.Match.Any {
		if !wantPatterns[p] {
			t.Errorf("unexpected pattern %q in merged.match.any", p)
		}
	}
}

// TestMergePlaybooksMatchNoneCombined tests that none patterns from base are
// propagated to the merged playbook.
func TestMergePlaybooksMatchNoneCombined(t *testing.T) {
	base := model.Playbook{
		ID:    "base",
		Match: model.MatchSpec{None: []string{"false-positive-signal"}},
	}
	child := model.Playbook{
		ID:      "child",
		Extends: "base",
		Match:   model.MatchSpec{Any: []string{"child-specific signal"}},
	}

	merged := mergePlaybooks(base, child)

	if len(merged.Match.None) != 1 || merged.Match.None[0] != "false-positive-signal" {
		t.Errorf("expected base none patterns to be inherited, got %v", merged.Match.None)
	}
}

// TestMergePlaybooksTagsUnioned tests that tags from base and child are merged
// without duplicates.
func TestMergePlaybooksTagsUnioned(t *testing.T) {
	base := model.Playbook{
		ID:   "base",
		Tags: []string{"shared-tag", "base-tag"},
	}
	child := model.Playbook{
		ID:      "child",
		Extends: "base",
		Tags:    []string{"shared-tag", "child-tag"},
	}

	merged := mergePlaybooks(base, child)

	if len(merged.Tags) != 3 {
		t.Errorf("expected 3 unique tags, got %d: %v", len(merged.Tags), merged.Tags)
	}
}

// TestMergePlaybooksBaseScoreInheritedWhenChildZero tests that when child has
// no base_score, the base score is inherited.
func TestMergePlaybooksBaseScoreInheritedWhenChildZero(t *testing.T) {
	base := model.Playbook{ID: "base", BaseScore: 1.2}
	child := model.Playbook{ID: "child", Extends: "base"}

	merged := mergePlaybooks(base, child)

	if merged.BaseScore != 1.2 {
		t.Errorf("expected inherited base_score 1.2, got %f", merged.BaseScore)
	}
}

// TestMergePlaybooksBaseScoreNotOverriddenWhenChildSet tests that an explicit
// child base_score is not replaced by the base.
func TestMergePlaybooksBaseScoreNotOverriddenWhenChildSet(t *testing.T) {
	base := model.Playbook{ID: "base", BaseScore: 1.2}
	child := model.Playbook{ID: "child", Extends: "base", BaseScore: 0.8}

	merged := mergePlaybooks(base, child)

	if merged.BaseScore != 0.8 {
		t.Errorf("expected child base_score 0.8, got %f", merged.BaseScore)
	}
}

// TestResolvePlaybookInheritanceGoldenPath tests the documented golden path:
// a generic base playbook and a narrower child that extends it.
// This mirrors the missing-executable → node-missing-executable example
// described in docs/playbooks.md.
func TestResolvePlaybookInheritanceGoldenPath(t *testing.T) {
	base := model.Playbook{
		ID:        "missing-executable",
		Title:     "Required executable or runtime binary missing",
		Severity:  "high",
		Summary:   "The job tried to launch a required tool but the executable was missing.",
		Diagnosis: "The runner could not find the binary in $PATH.",
		Fix:       "Install the missing package in the CI image.",
		Match: model.MatchSpec{
			Any:  []string{"command not found", "exit status 127"},
			None: []string{"fixture", "testdata"},
		},
		Tags: []string{"toolchain", "executable"},
	}

	child := model.Playbook{
		ID:        "node-missing-executable",
		Extends:   "missing-executable",
		Title:     "Node.js runtime or tool missing in CI",
		Summary:   "The CI job failed because a Node.js binary (node, npm, npx) was not found.",
		Diagnosis: "Node.js is not installed in the runner image or the PATH was not updated after installation.",
		Fix:       "Use the actions/setup-node action or install Node.js from the package manager in the base image.",
		Match: model.MatchSpec{
			Any:  []string{"node: command not found", "npm: command not found"},
			None: []string{"node_modules"},
		},
		Tags: []string{"node", "npm"},
	}

	pbs := []model.Playbook{base, child}
	resolved, err := resolvePlaybookInheritance(pbs)
	if err != nil {
		t.Fatalf("resolvePlaybookInheritance: %v", err)
	}

	var resolvedChild model.Playbook
	for _, pb := range resolved {
		if pb.ID == "node-missing-executable" {
			resolvedChild = pb
			break
		}
	}

	// Child title overrides base.
	if resolvedChild.Title != "Node.js runtime or tool missing in CI" {
		t.Errorf("expected child title, got %q", resolvedChild.Title)
	}

	// Child should have node-specific match patterns.
	wantAny := map[string]bool{
		"node: command not found": true,
		"npm: command not found":  true,
		"command not found":       true,
		"exit status 127":         true,
	}
	if len(resolvedChild.Match.Any) != 4 {
		t.Errorf("expected 4 any patterns, got %d: %v", len(resolvedChild.Match.Any), resolvedChild.Match.Any)
	}
	for _, p := range resolvedChild.Match.Any {
		if !wantAny[p] {
			t.Errorf("unexpected pattern in merged any: %q", p)
		}
	}

	// None patterns from both base and child should be merged.
	wantNone := map[string]bool{
		"fixture":      true,
		"testdata":     true,
		"node_modules": true,
	}
	if len(resolvedChild.Match.None) != 3 {
		t.Errorf("expected 3 none patterns, got %d: %v", len(resolvedChild.Match.None), resolvedChild.Match.None)
	}
	for _, p := range resolvedChild.Match.None {
		if !wantNone[p] {
			t.Errorf("unexpected pattern in merged none: %q", p)
		}
	}

	// Tags from both should be merged: toolchain, executable, node, npm.
	if len(resolvedChild.Tags) != 4 {
		t.Errorf("expected 4 tags, got %d: %v", len(resolvedChild.Tags), resolvedChild.Tags)
	}

	// Base playbook is not modified by the resolution.
	for _, pb := range resolved {
		if pb.ID == "missing-executable" {
			if len(pb.Match.Any) != 2 {
				t.Errorf("base match.any should be unchanged, got %v", pb.Match.Any)
			}
			break
		}
	}
}

// TestResolvePlaybookInheritanceDeepChain tests three-level inheritance
// (grandparent → parent → child).
func TestResolvePlaybookInheritanceDeepChain(t *testing.T) {
	grandparent := model.Playbook{
		ID:    "gp",
		Title: "Grandparent",
		Match: model.MatchSpec{Any: []string{"gp-signal"}},
	}
	parent := model.Playbook{
		ID:      "p",
		Extends: "gp",
		Match:   model.MatchSpec{Any: []string{"p-signal"}},
	}
	child := model.Playbook{
		ID:      "c",
		Extends: "p",
		Match:   model.MatchSpec{Any: []string{"c-signal"}},
	}

	pbs := []model.Playbook{grandparent, parent, child}
	resolved, err := resolvePlaybookInheritance(pbs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resolvedChild model.Playbook
	for _, pb := range resolved {
		if pb.ID == "c" {
			resolvedChild = pb
			break
		}
	}

	wantAny := map[string]bool{"gp-signal": true, "p-signal": true, "c-signal": true}
	if len(resolvedChild.Match.Any) != 3 {
		t.Errorf("expected 3 patterns from full chain, got %d: %v", len(resolvedChild.Match.Any), resolvedChild.Match.Any)
	}
	for _, p := range resolvedChild.Match.Any {
		if !wantAny[p] {
			t.Errorf("unexpected pattern: %q", p)
		}
	}
}
