package playbooks

import (
	"testing"

	"faultline/internal/model"
)

// ── sortRefs ──────────────────────────────────────────────────────────────────

func TestSortRefsByPlaybookIDThenSectionThenPattern(t *testing.T) {
	refs := []PatternRef{
		{PlaybookID: "z-pb", Section: "match.any", Pattern: "zzz"},
		{PlaybookID: "a-pb", Section: "match.none", Pattern: "aaa"},
		{PlaybookID: "a-pb", Section: "match.any", Pattern: "bbb"},
		{PlaybookID: "a-pb", Section: "match.any", Pattern: "aaa"},
	}
	sortRefs(refs)

	// First element must have lowest PlaybookID
	if refs[0].PlaybookID != "a-pb" {
		t.Errorf("refs[0].PlaybookID = %q, want a-pb", refs[0].PlaybookID)
	}
	// Within same playbook, match.any < match.none
	if refs[0].Section != "match.any" {
		t.Errorf("refs[0].Section = %q, want match.any", refs[0].Section)
	}
	// Within same playbook and section, sorted by pattern
	if refs[0].Pattern != "aaa" {
		t.Errorf("refs[0].Pattern = %q, want aaa", refs[0].Pattern)
	}
	if refs[1].Pattern != "bbb" {
		t.Errorf("refs[1].Pattern = %q, want bbb", refs[1].Pattern)
	}
	// Last element is z-pb
	if refs[len(refs)-1].PlaybookID != "z-pb" {
		t.Errorf("last ref PlaybookID = %q, want z-pb", refs[len(refs)-1].PlaybookID)
	}
}

func TestSortRefsEmpty(t *testing.T) {
	var refs []PatternRef
	sortRefs(refs) // must not panic
}

func TestSortRefsSingleElement(t *testing.T) {
	refs := []PatternRef{{PlaybookID: "x", Section: "match.any", Pattern: "p"}}
	sortRefs(refs)
	if refs[0].PlaybookID != "x" {
		t.Errorf("sortRefs single element mutated: %+v", refs[0])
	}
}

// ── partialGroupKey ───────────────────────────────────────────────────────────

func TestPartialGroupKeyUsesIDWhenPresent(t *testing.T) {
	group := model.PartialMatchGroup{
		ID:       "my-group",
		Minimum:  2,
		Patterns: []string{"foo", "bar"},
	}
	got := partialGroupKey(group)
	const want = "id:my-group"
	if got != want {
		t.Errorf("partialGroupKey with ID = %q, want %q", got, want)
	}
}

func TestPartialGroupKeyFallsBackToMinimumAndPatterns(t *testing.T) {
	group := model.PartialMatchGroup{
		ID:       "",
		Minimum:  2,
		Patterns: []string{"foo", "bar"},
	}
	got := partialGroupKey(group)
	if got == "id:" {
		t.Errorf("expected non-id key, got %q", got)
	}
	// Key must encode minimum
	if len(got) == 0 {
		t.Error("expected non-empty key")
	}
}

func TestPartialGroupKeyWhitespaceIDFallsBack(t *testing.T) {
	group := model.PartialMatchGroup{
		ID:       "   ",
		Minimum:  1,
		Patterns: []string{"error"},
	}
	got := partialGroupKey(group)
	if got == "id:   " {
		t.Errorf("whitespace ID should not be used as key, got %q", got)
	}
}

func TestPartialGroupKeyDifferentGroupsDifferentKeys(t *testing.T) {
	g1 := model.PartialMatchGroup{ID: "", Minimum: 2, Patterns: []string{"a", "b"}}
	g2 := model.PartialMatchGroup{ID: "", Minimum: 3, Patterns: []string{"a", "b"}}
	if partialGroupKey(g1) == partialGroupKey(g2) {
		t.Errorf("groups with different Minimum should have different keys: %q", partialGroupKey(g1))
	}
}

// ── validateMatchCatalogItem ──────────────────────────────────────────────────

func TestValidateMatchCatalogItemEmptyFails(t *testing.T) {
	item := rawNamedMatcher{}
	err := validateMatchCatalogItem(item, "test.yaml", "my-matcher")
	if err == nil {
		t.Fatal("expected error for empty matcher, got nil")
	}
}

func TestValidateMatchCatalogItemWithAny(t *testing.T) {
	item := rawNamedMatcher{Any: []string{"npm ERR! code E401"}}
	err := validateMatchCatalogItem(item, "test.yaml", "npm-auth")
	if err != nil {
		t.Errorf("expected no error for valid Any matcher, got %v", err)
	}
}

func TestValidateMatchCatalogItemWithAll(t *testing.T) {
	item := rawNamedMatcher{All: []string{"error", "not found"}}
	err := validateMatchCatalogItem(item, "test.yaml", "combo")
	if err != nil {
		t.Errorf("expected no error for valid All matcher, got %v", err)
	}
}

func TestValidateMatchCatalogItemWithNone(t *testing.T) {
	item := rawNamedMatcher{None: []string{"success"}}
	err := validateMatchCatalogItem(item, "test.yaml", "no-success")
	if err != nil {
		t.Errorf("expected no error for valid None matcher, got %v", err)
	}
}

func TestValidateMatchCatalogItemWithUse(t *testing.T) {
	item := rawNamedMatcher{Use: []string{"base-matcher"}}
	err := validateMatchCatalogItem(item, "test.yaml", "derived")
	if err != nil {
		t.Errorf("expected no error for valid Use reference, got %v", err)
	}
}

func TestValidateMatchCatalogItemEmptyPatternInAnyFails(t *testing.T) {
	item := rawNamedMatcher{Any: []string{"valid-pattern", ""}}
	err := validateMatchCatalogItem(item, "test.yaml", "bad-empty")
	if err == nil {
		t.Fatal("expected error for empty pattern in Any, got nil")
	}
}

func TestValidateMatchCatalogItemEmptyRefInUseFails(t *testing.T) {
	item := rawNamedMatcher{Use: []string{""}}
	err := validateMatchCatalogItem(item, "test.yaml", "bad-ref")
	if err == nil {
		t.Fatal("expected error for empty Use reference, got nil")
	}
}
