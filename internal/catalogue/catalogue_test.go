package catalogue

import (
	"encoding/json"
	"strings"
	"testing"

	"faultline/internal/model"
)

// ---------------------------------------------------------------------------
// ValidateSlug
// ---------------------------------------------------------------------------

func TestValidateSlugValid(t *testing.T) {
	cases := []string{
		"aws-credentials",
		"npm-ci-lockfile-drift",
		"a",
		"abc123",
		"docker-auth",
	}
	for _, slug := range cases {
		if err := ValidateSlug(slug); err != nil {
			t.Errorf("ValidateSlug(%q) returned unexpected error: %v", slug, err)
		}
	}
}

func TestValidateSlugInvalid(t *testing.T) {
	cases := []string{
		"",
		"-starts-with-hyphen",
		"ends-with-hyphen-",
		"Has_Uppercase",
		"has.dot",
		"has space",
		"ALLCAPS",
	}
	for _, slug := range cases {
		if err := ValidateSlug(slug); err == nil {
			t.Errorf("ValidateSlug(%q) expected error, got nil", slug)
		}
	}
}

// ---------------------------------------------------------------------------
// SlugFromID
// ---------------------------------------------------------------------------

func TestSlugFromIDHyphens(t *testing.T) {
	got := SlugFromID("aws-credentials")
	if got != "aws-credentials" {
		t.Errorf("SlugFromID(%q) = %q, want %q", "aws-credentials", got, "aws-credentials")
	}
}

func TestSlugFromIDDotSeparated(t *testing.T) {
	got := SlugFromID("some.failure.id")
	if got != "some-failure-id" {
		t.Errorf("SlugFromID(%q) = %q, want %q", "some.failure.id", got, "some-failure-id")
	}
}

func TestSlugFromIDUnderscore(t *testing.T) {
	got := SlugFromID("npm_lockfile_drift")
	if got != "npm-lockfile-drift" {
		t.Errorf("SlugFromID(%q) = %q, want %q", "npm_lockfile_drift", got, "npm-lockfile-drift")
	}
}

func TestSlugFromIDUppercase(t *testing.T) {
	got := SlugFromID("Docker-Auth")
	if got != "docker-auth" {
		t.Errorf("SlugFromID(%q) = %q, want %q", "Docker-Auth", got, "docker-auth")
	}
}

// ---------------------------------------------------------------------------
// EcosystemsFromTags
// ---------------------------------------------------------------------------

func TestEcosystemsFromTagsFiltersKnown(t *testing.T) {
	tags := []string{"aws", "iam", "ci", "docker", "unknown-tag"}
	got := EcosystemsFromTags(tags)
	// Only aws and docker are in the known set.
	if len(got) != 2 {
		t.Fatalf("EcosystemsFromTags() = %v, want 2 entries", got)
	}
	if got[0] != "aws" || got[1] != "docker" {
		t.Errorf("EcosystemsFromTags() = %v, want [aws docker]", got)
	}
}

func TestEcosystemsFromTagsEmpty(t *testing.T) {
	got := EcosystemsFromTags(nil)
	if len(got) != 0 {
		t.Errorf("EcosystemsFromTags(nil) = %v, want empty", got)
	}
}

func TestEcosystemsFromTagsDeduplicated(t *testing.T) {
	tags := []string{"docker", "Docker", "DOCKER"}
	got := EcosystemsFromTags(tags)
	if len(got) != 1 {
		t.Errorf("EcosystemsFromTags() = %v, expected dedup to 1 entry", got)
	}
}

func TestEcosystemsFromTagsSorted(t *testing.T) {
	tags := []string{"node", "aws", "docker", "github-actions"}
	got := EcosystemsFromTags(tags)
	for i := 1; i < len(got); i++ {
		if got[i] < got[i-1] {
			t.Errorf("EcosystemsFromTags() result not sorted: %v", got)
		}
	}
}

// ---------------------------------------------------------------------------
// ConfidenceFromSeverity
// ---------------------------------------------------------------------------

func TestConfidenceFromSeverity(t *testing.T) {
	cases := []struct {
		severity string
		want     string
	}{
		{"critical", "high"},
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
		{"", "medium"},
		{"UNKNOWN", "medium"},
	}
	for _, tc := range cases {
		got := ConfidenceFromSeverity(tc.severity)
		if got != tc.want {
			t.Errorf("ConfidenceFromSeverity(%q) = %q, want %q", tc.severity, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Frontmatter rendering
// ---------------------------------------------------------------------------

func TestRenderFrontmatterValid(t *testing.T) {
	e := Entry{
		Slug:        "aws-credentials",
		Title:       "AWS credentials missing or invalid",
		Description: "The job could not authenticate with AWS.",
		FailureID:   "aws-credentials",
		Category:    "auth",
		Ecosystems:  []string{"aws", "github-actions"},
		Signals:     []string{"NoCredentialProviders", "Unable to locate credentials"},
		Confidence:  "high",
	}
	data, err := RenderFrontmatter(e)
	if err != nil {
		t.Fatalf("RenderFrontmatter() error: %v", err)
	}
	s := string(data)
	// Must start and end with --- delimiters.
	if !strings.HasPrefix(s, "---\n") {
		t.Errorf("frontmatter should start with ---\\n, got: %q", s[:min(20, len(s))])
	}
	if !strings.Contains(s, "\n---\n") {
		t.Errorf("frontmatter should contain closing ---\\n")
	}
	// Must contain required fields.
	for _, field := range []string{"title:", "description:", "failure_id:", "category:", "ecosystems:", "signals:", "confidence:"} {
		if !strings.Contains(s, field) {
			t.Errorf("frontmatter missing field %q:\n%s", field, s)
		}
	}
}

func TestRenderFrontmatterQuotesSpecialChars(t *testing.T) {
	e := Entry{
		Slug:        "test-entry",
		Title:       `Title with "quotes" and \backslash`,
		Description: "A description.",
		FailureID:   "test-entry",
		Category:    "build",
		Ecosystems:  []string{},
		Signals:     []string{},
		Confidence:  "medium",
	}
	data, err := RenderFrontmatter(e)
	if err != nil {
		t.Fatalf("RenderFrontmatter() error: %v", err)
	}
	s := string(data)
	// Verify backslash and quote are escaped.
	if !strings.Contains(s, `\\`) {
		t.Errorf("expected escaped backslash in frontmatter, got:\n%s", s)
	}
	if !strings.Contains(s, `\"`) {
		t.Errorf("expected escaped double-quote in frontmatter, got:\n%s", s)
	}
}

func TestRenderFrontmatterEmptyEcosystems(t *testing.T) {
	e := Entry{
		Slug:        "simple",
		Title:       "Simple failure",
		Description: "A simple description.",
		FailureID:   "simple",
		Category:    "build",
		Ecosystems:  []string{},
		Signals:     []string{},
		Confidence:  "medium",
	}
	data, err := RenderFrontmatter(e)
	if err != nil {
		t.Fatalf("RenderFrontmatter() error: %v", err)
	}
	// Empty ecosystems should render as []
	if !strings.Contains(string(data), "ecosystems: []") {
		t.Errorf("expected ecosystems: [], got:\n%s", string(data))
	}
}

// ---------------------------------------------------------------------------
// BuildEntries
// ---------------------------------------------------------------------------

func TestBuildEntriesSortsAndConverts(t *testing.T) {
	pbs := []sourcePlaybook{
		{Playbook: model.Playbook{ID: "npm-lockfile", Title: "NPM Lockfile Drift", Category: "build", Severity: "high", Summary: "The lockfile is out of sync.", Tags: []string{"npm", "node"}, Match: model.MatchSpec{Any: []string{"package-lock.json"}}}},
		{Playbook: model.Playbook{ID: "aws-credentials", Title: "AWS credentials missing", Category: "auth", Severity: "high", Summary: "AWS credentials not found.", Tags: []string{"aws"}, Match: model.MatchSpec{Any: []string{"NoCredentialProviders"}}}},
	}
	entries := BuildEntries(pbs)
	if len(entries) != 2 {
		t.Fatalf("BuildEntries() returned %d entries, want 2", len(entries))
	}
	// auth comes before build alphabetically.
	if entries[0].Category != "auth" {
		t.Errorf("expected first entry category=auth, got %q", entries[0].Category)
	}
	if entries[0].Slug != "aws-credentials" {
		t.Errorf("expected first entry slug=aws-credentials, got %q", entries[0].Slug)
	}
	if entries[1].Ecosystems[0] != "node" || entries[1].Ecosystems[1] != "npm" {
		t.Errorf("expected ecosystems [node npm], got %v", entries[1].Ecosystems)
	}
}

func TestBuildEntriesSkipsInvalidSlug(t *testing.T) {
	pbs := []sourcePlaybook{
		// ID that can't form a valid slug (starts with hyphen after normalization).
		{Playbook: model.Playbook{ID: "-invalid", Title: "Bad", Category: "build", Summary: "Bad."}},
		{Playbook: model.Playbook{ID: "valid-id", Title: "Good", Category: "build", Summary: "Good."}},
	}
	entries := BuildEntries(pbs)
	if len(entries) != 1 {
		t.Errorf("BuildEntries() = %d entries, want 1 (invalid slug skipped)", len(entries))
	}
	if entries[0].Slug != "valid-id" {
		t.Errorf("expected slug valid-id, got %q", entries[0].Slug)
	}
}

// ---------------------------------------------------------------------------
// catalogue.json generation
// ---------------------------------------------------------------------------

func TestRenderCatalogueJSONValid(t *testing.T) {
	entries := []Entry{
		{Slug: "docker-auth", Title: "Docker auth failure", Description: "Docker login failed.", FailureID: "docker-auth", Category: "auth", Ecosystems: []string{"docker"}, Signals: []string{"unauthorized"}, Confidence: "high"},
		{Slug: "npm-lockfile", Title: "NPM lockfile drift", Description: "Lockfile is stale.", FailureID: "npm-lockfile", Category: "build", Ecosystems: []string{"npm"}, Signals: []string{"npm ci"}, Confidence: "medium"},
	}
	data, err := renderCatalogueJSON(entries)
	if err != nil {
		t.Fatalf("renderCatalogueJSON() error: %v", err)
	}
	if err := ValidateJSON(data); err != nil {
		t.Fatalf("generated catalogue.json is not valid JSON: %v", err)
	}
	// Unmarshal and verify count.
	var out []map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("expected 2 entries in JSON, got %d", len(out))
	}
}

func TestRenderCatalogueJSONDeterministic(t *testing.T) {
	entries := []Entry{
		{Slug: "z-last", Title: "Z last", Description: "Z.", FailureID: "z-last", Category: "build", Confidence: "low"},
		{Slug: "a-first", Title: "A first", Description: "A.", FailureID: "a-first", Category: "auth", Confidence: "high"},
	}
	data1, _ := renderCatalogueJSON(entries)
	data2, _ := renderCatalogueJSON(entries)
	if string(data1) != string(data2) {
		t.Errorf("renderCatalogueJSON() is not deterministic")
	}
	// Verify sorted order: auth (a-first) before build (z-last).
	var out []map[string]any
	if err := json.Unmarshal(data1, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out[0]["slug"] != "a-first" {
		t.Errorf("expected a-first first, got %v", out[0]["slug"])
	}
}

// ---------------------------------------------------------------------------
// Manifest generation
// ---------------------------------------------------------------------------

func TestRenderManifestJSONValid(t *testing.T) {
	m := Manifest{
		SourceRepo:       "org/faultline",
		SourceCommit:     "abc123def456",
		GeneratedAt:      "2025-01-01T00:00:00Z",
		FailureCount:     42,
		GeneratorVersion: "1.2.3",
	}
	data, err := renderManifestJSON(m)
	if err != nil {
		t.Fatalf("renderManifestJSON() error: %v", err)
	}
	if err := ValidateJSON(data); err != nil {
		t.Fatalf("generated manifest is not valid JSON: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["failure_count"].(float64) != 42 {
		t.Errorf("expected failure_count=42, got %v", out["failure_count"])
	}
	if out["source_repo"] != "org/faultline" {
		t.Errorf("expected source_repo=org/faultline, got %v", out["source_repo"])
	}
}

// ---------------------------------------------------------------------------
// ValidateEntry / ValidateEntries
// ---------------------------------------------------------------------------

func TestValidateEntryValid(t *testing.T) {
	e := Entry{
		Slug:        "aws-credentials",
		Title:       "AWS credentials",
		Description: "AWS creds missing.",
		FailureID:   "aws-credentials",
		Category:    "auth",
		Confidence:  "high",
	}
	if err := ValidateEntry(e); err != nil {
		t.Errorf("ValidateEntry() unexpected error: %v", err)
	}
}

func TestValidateEntryMissingFields(t *testing.T) {
	cases := []struct {
		name  string
		entry Entry
	}{
		{"missing slug", Entry{Title: "T", Description: "D.", FailureID: "id"}},
		{"invalid slug", Entry{Slug: "Invalid_Slug", Title: "T", Description: "D.", FailureID: "id"}},
		{"missing failure_id", Entry{Slug: "valid", Title: "T", Description: "D."}},
		{"missing title", Entry{Slug: "valid", Description: "D.", FailureID: "id"}},
		{"missing description", Entry{Slug: "valid", Title: "T", FailureID: "id"}},
	}
	for _, tc := range cases {
		if err := ValidateEntry(tc.entry); err == nil {
			t.Errorf("ValidateEntry(%s) expected error, got nil", tc.name)
		}
	}
}

func TestValidateEntriesDuplicateSlug(t *testing.T) {
	entries := []Entry{
		{Slug: "dup", Title: "A", Description: "Desc.", FailureID: "dup"},
		{Slug: "dup", Title: "B", Description: "Desc.", FailureID: "dup2"},
	}
	if err := ValidateEntries(entries); err == nil {
		t.Errorf("ValidateEntries() expected duplicate slug error, got nil")
	}
}

// ---------------------------------------------------------------------------
// ValidateJSON
// ---------------------------------------------------------------------------

func TestValidateJSONValid(t *testing.T) {
	if err := ValidateJSON([]byte(`{"key":"value"}`)); err != nil {
		t.Errorf("ValidateJSON() on valid JSON returned error: %v", err)
	}
}

func TestValidateJSONInvalid(t *testing.T) {
	if err := ValidateJSON([]byte(`{invalid`)); err == nil {
		t.Errorf("ValidateJSON() on invalid JSON should return error")
	}
}

// ---------------------------------------------------------------------------
// descriptionFromSummary
// ---------------------------------------------------------------------------

func TestDescriptionFromSummaryFirstSentence(t *testing.T) {
	got := descriptionFromSummary("The build failed. More details follow.")
	if got != "The build failed." {
		t.Errorf("descriptionFromSummary() = %q, want %q", got, "The build failed.")
	}
}

func TestDescriptionFromSummaryMarkdownHeading(t *testing.T) {
	got := descriptionFromSummary("## Why it fails\nThe auth token expired. Other details.")
	if got != "The auth token expired." {
		t.Errorf("descriptionFromSummary() = %q, want %q", got, "The auth token expired.")
	}
}

func TestDescriptionFromSummaryEmpty(t *testing.T) {
	got := descriptionFromSummary("")
	if got != "" {
		t.Errorf("descriptionFromSummary(\"\") = %q, want empty", got)
	}
}

func TestDescriptionFromSummaryPeriodInToken(t *testing.T) {
	// Period inside a filename should not be treated as a sentence boundary.
	got := descriptionFromSummary("`npm ci` found a missing or out-of-sync `package-lock.json`.")
	want := "`npm ci` found a missing or out-of-sync `package-lock.json`."
	if got != want {
		t.Errorf("descriptionFromSummary() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// topSignals
// ---------------------------------------------------------------------------

func TestTopSignalsLimitsToN(t *testing.T) {
	signals := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	got := topSignals(signals, 8)
	if len(got) != 8 {
		t.Errorf("topSignals(10, limit=8) = %d signals, want 8", len(got))
	}
}

func TestTopSignalsLessThanN(t *testing.T) {
	signals := []string{"a", "b"}
	got := topSignals(signals, 8)
	if len(got) != 2 {
		t.Errorf("topSignals(2, limit=8) = %d signals, want 2", len(got))
	}
}

func TestTopSignalsSkipsEmpty(t *testing.T) {
	signals := []string{"a", "", "b", "  ", "c"}
	got := topSignals(signals, 8)
	if len(got) != 3 {
		t.Errorf("topSignals() = %v, expected empty strings skipped", got)
	}
}
