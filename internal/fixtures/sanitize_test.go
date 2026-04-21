package fixtures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplySanitizeRulesGitHubToken(t *testing.T) {
	input := "Fetching token ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdef12 from env"
	got, reps := ApplySanitizeRules(input)
	if strings.Contains(got, "ghp_") {
		t.Fatalf("token not redacted: %q", got)
	}
	if !strings.Contains(got, "<redacted-github-token>") {
		t.Fatalf("expected placeholder, got: %q", got)
	}
	if len(reps) == 0 || reps[0].Pattern != "github-token" || reps[0].Count != 1 {
		t.Fatalf("unexpected replacements: %v", reps)
	}
}

func TestApplySanitizeRulesAWSKey(t *testing.T) {
	input := "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE"
	got, reps := ApplySanitizeRules(input)
	if strings.Contains(got, "AKIAIOSFODNN7EXAMPLE") {
		t.Fatalf("AWS key not redacted: %q", got)
	}
	if !strings.Contains(got, "<redacted-aws-key>") {
		t.Fatalf("expected placeholder, got: %q", got)
	}
	if len(reps) == 0 || reps[0].Pattern != "aws-key" {
		t.Fatalf("unexpected replacements: %v", reps)
	}
}

func TestApplySanitizeRulesAuthHeader(t *testing.T) {
	input := "Authorization: Bearer supersecrettoken123abc"
	got, reps := ApplySanitizeRules(input)
	if strings.Contains(got, "supersecrettoken123abc") {
		t.Fatalf("token not redacted: %q", got)
	}
	if !strings.Contains(got, "Authorization:") {
		t.Fatalf("header prefix stripped unexpectedly: %q", got)
	}
	if !strings.Contains(got, "<redacted>") {
		t.Fatalf("expected redaction placeholder: %q", got)
	}
	if len(reps) == 0 || reps[0].Pattern != "auth-header" {
		t.Fatalf("unexpected replacements: %v", reps)
	}
}

func TestApplySanitizeRulesURLCredentials(t *testing.T) {
	input := "Cloning from https://user:s3cr3tpass@github.com/org/repo"
	got, reps := ApplySanitizeRules(input)
	if strings.Contains(got, "s3cr3tpass") {
		t.Fatalf("URL password not redacted: %q", got)
	}
	if !strings.Contains(got, "https://") {
		t.Fatalf("URL scheme stripped unexpectedly: %q", got)
	}
	if len(reps) == 0 || reps[0].Pattern != "url-credentials" {
		t.Fatalf("unexpected replacements: %v", reps)
	}
}

func TestApplySanitizeRulesCredentialKV(t *testing.T) {
	cases := []struct {
		input string
		key   string
	}{
		{"password=hunter2", "password"},
		{"secret: mysupersecret", "secret"},
		{"api_key=abc123xyz", "api_key"},
	}
	for _, tc := range cases {
		got, reps := ApplySanitizeRules(tc.input)
		if strings.Contains(got, "hunter2") || strings.Contains(got, "mysupersecret") || strings.Contains(got, "abc123xyz") {
			t.Fatalf("[%s] value not redacted: %q", tc.key, got)
		}
		if !strings.Contains(got, "<redacted>") {
			t.Fatalf("[%s] expected redaction placeholder: %q", tc.key, got)
		}
		if len(reps) == 0 || reps[0].Pattern != "credential-kv" {
			t.Fatalf("[%s] unexpected replacements: %v", tc.key, reps)
		}
	}
}

func TestApplySanitizeRulesJWT(t *testing.T) {
	input := "token: eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ1c2VyIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	got, reps := ApplySanitizeRules(input)
	if strings.Contains(got, "eyJhbGci") {
		t.Fatalf("JWT not redacted: %q", got)
	}
	if !strings.Contains(got, "<redacted-jwt>") {
		t.Fatalf("expected JWT placeholder: %q", got)
	}
	if len(reps) == 0 || reps[0].Pattern != "jwt" {
		t.Fatalf("unexpected replacements: %v", reps)
	}
}

func TestApplySanitizeRulesEmail(t *testing.T) {
	input := "Error: user john.doe@example.com not authorized"
	got, reps := ApplySanitizeRules(input)
	if strings.Contains(got, "john.doe@example.com") {
		t.Fatalf("email not redacted: %q", got)
	}
	if !strings.Contains(got, "<redacted-email>") {
		t.Fatalf("expected email placeholder: %q", got)
	}
	if len(reps) == 0 || reps[0].Pattern != "email" {
		t.Fatalf("unexpected replacements: %v", reps)
	}
}

func TestApplySanitizeRulesNoMatch(t *testing.T) {
	input := "exec /usr/bin/node: no such file or directory"
	got, reps := ApplySanitizeRules(input)
	if got != input {
		t.Fatalf("clean log modified unexpectedly: %q", got)
	}
	if len(reps) != 0 {
		t.Fatalf("expected no replacements, got %v", reps)
	}
}

func TestApplySanitizeRulesDeterministic(t *testing.T) {
	input := "Authorization: Bearer tok123 and email user@test.com and key ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdef12"
	first, _ := ApplySanitizeRules(input)
	second, _ := ApplySanitizeRules(input)
	if first != second {
		t.Fatalf("non-deterministic output: %q vs %q", first, second)
	}
}

func TestApplySanitizeRulesPEMKey(t *testing.T) {
	input := `Log line before
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA2a2rwplBQLzHPZe5
-----END RSA PRIVATE KEY-----
Log line after`
	got, reps := ApplySanitizeRules(input)
	if strings.Contains(got, "BEGIN RSA PRIVATE KEY") {
		t.Fatalf("PEM key not redacted: %q", got)
	}
	if !strings.Contains(got, "<redacted-pem-key>") {
		t.Fatalf("expected PEM placeholder: %q", got)
	}
	if len(reps) == 0 || reps[0].Pattern != "pem-key" {
		t.Fatalf("unexpected replacements: %v", reps)
	}
}

func TestSanitizeFixtureInPlace(t *testing.T) {
	staging := t.TempDir()
	layout := Layout{
		Root:       t.TempDir(),
		StagingDir: staging,
	}

	fixture := Fixture{
		ID:            "test-fixture",
		FixtureClass:  ClassStaging,
		RawLog:        "auth header: Authorization: Bearer secrettoken999 build failed",
		NormalizedLog: "auth header: Authorization: Bearer secrettoken999 build failed",
	}
	path := filepath.Join(staging, "test-fixture.yaml")
	if err := writeFixture(path, fixture); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	results, err := Sanitize(layout, []string{"test-fixture"}, SanitizeOptions{DryRun: false})
	if err != nil {
		t.Fatalf("Sanitize: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.TotalReplacements() == 0 {
		t.Fatal("expected replacements, got none")
	}

	// Reload and verify content was written.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if strings.Contains(string(data), "secrettoken999") {
		t.Fatalf("secret still present after in-place sanitize: %s", data)
	}
}

func TestSanitizeDryRunDoesNotModify(t *testing.T) {
	staging := t.TempDir()
	layout := Layout{
		Root:       t.TempDir(),
		StagingDir: staging,
	}

	original := "password=supersecret and email admin@internal.corp"
	fixture := Fixture{
		ID:            "dry-run-fixture",
		FixtureClass:  ClassStaging,
		RawLog:        original,
		NormalizedLog: original,
	}
	path := filepath.Join(staging, "dry-run-fixture.yaml")
	if err := writeFixture(path, fixture); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	results, err := Sanitize(layout, []string{"dry-run-fixture"}, SanitizeOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Sanitize dry-run: %v", err)
	}
	if len(results) != 1 || results[0].TotalReplacements() == 0 {
		t.Fatalf("expected replacements reported, got %+v", results)
	}
	if !results[0].DryRun {
		t.Fatal("expected DryRun=true on result")
	}

	// File must be unchanged.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if !strings.Contains(string(data), "supersecret") {
		t.Fatalf("dry-run modified the file unexpectedly")
	}
}

func TestSanitizeMissingIDReturnsError(t *testing.T) {
	layout := Layout{
		Root:       t.TempDir(),
		StagingDir: t.TempDir(),
	}
	_, err := Sanitize(layout, []string{"does-not-exist"}, SanitizeOptions{})
	if err == nil {
		t.Fatal("expected error for missing fixture ID")
	}
}

func TestSanitizeEmptyIDsReturnsError(t *testing.T) {
	layout := Layout{Root: t.TempDir(), StagingDir: t.TempDir()}
	_, err := Sanitize(layout, nil, SanitizeOptions{})
	if err == nil {
		t.Fatal("expected error for empty ID list")
	}
}

func TestFormatSanitizeResultsText(t *testing.T) {
	results := []SanitizeResult{
		{
			FixtureID: "fixture-abc",
			Path:      "/tmp/fixture-abc.yaml",
			Replacements: []Replacement{
				{Pattern: "github-token", Field: "raw_log", Count: 2},
				{Pattern: "email", Field: "normalized_log", Count: 1},
			},
			DryRun: false,
		},
	}
	text, err := FormatSanitizeResults(results, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{
		"fixture-abc: 3 replacement(s)",
		"github-token [raw_log]: 2",
		"email [normalized_log]: 1",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in output, got:\n%s", want, text)
		}
	}
}

func TestFormatSanitizeResultsTextDryRun(t *testing.T) {
	results := []SanitizeResult{
		{
			FixtureID:    "fixture-xyz",
			Replacements: []Replacement{{Pattern: "email", Field: "raw_log", Count: 1}},
			DryRun:       true,
		},
	}
	text, err := FormatSanitizeResults(results, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text, "(dry-run)") {
		t.Fatalf("expected dry-run marker in output: %s", text)
	}
}

func TestFormatSanitizeResultsTextEmpty(t *testing.T) {
	text, err := FormatSanitizeResults(nil, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "No staging fixtures sanitized.\n" {
		t.Fatalf("unexpected empty output: %q", text)
	}
}

func TestFormatSanitizeResultsJSON(t *testing.T) {
	results := []SanitizeResult{
		{
			FixtureID: "fixture-json",
			Path:      "/tmp/fixture-json.yaml",
			Replacements: []Replacement{
				{Pattern: "jwt", Field: "raw_log", Count: 1},
			},
			DryRun: false,
		},
	}
	jsonText, err := FormatSanitizeResults(results, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{
		`"fixture_id"`,
		`"fixture-json"`,
		`"jwt"`,
		`"total_replacements"`,
	} {
		if !strings.Contains(jsonText, want) {
			t.Fatalf("expected %q in JSON output: %s", want, jsonText)
		}
	}
}
