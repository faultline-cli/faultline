package fixtures

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Tests for writeFixture.

func TestWriteFixtureCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-fixture.yaml")
	fixture := Fixture{
		ID:            "test-fixture",
		RawLog:        "npm ERR! lockfile mismatch",
		NormalizedLog: "npm ERR! lockfile mismatch",
		Fingerprint:   "abc123",
		FixtureClass:  ClassStaging,
	}
	if err := writeFixture(path, fixture); err != nil {
		t.Fatalf("writeFixture: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist at %s: %v", path, err)
	}
}

func TestWriteFixtureClearsFilePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-fixture.yaml")
	fixture := Fixture{
		ID:           "test-fixture",
		RawLog:       "some log",
		FilePath:     "/old/path/should/be/cleared.yaml",
		FixtureClass: ClassStaging,
	}
	if err := writeFixture(path, fixture); err != nil {
		t.Fatalf("writeFixture: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture file: %v", err)
	}
	// FilePath field should have been cleared before serialization — it must
	// not appear in the YAML output.
	if strings.Contains(string(data), "file_path") {
		t.Errorf("expected file_path to be absent from serialized YAML, got:\n%s", string(data))
	}
}

func TestWriteFixtureCreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "nested", "fixture.yaml")
	fixture := Fixture{
		ID:           "nested-fixture",
		RawLog:       "error log",
		FixtureClass: ClassStaging,
	}
	if err := writeFixture(path, fixture); err != nil {
		t.Fatalf("writeFixture with nested dirs: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file at %s: %v", path, err)
	}
}

// Tests for Promote.

func makeStagingLayout(t *testing.T) (Layout, string) {
	t.Helper()
	root := t.TempDir()
	staging := filepath.Join(root, "staging")
	real := filepath.Join(root, "real")
	if err := os.MkdirAll(staging, 0o755); err != nil {
		t.Fatalf("mkdir staging: %v", err)
	}
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatalf("mkdir real: %v", err)
	}
	layout := Layout{
		Root:       root,
		Fixtures:   root,
		StagingDir: staging,
		RealDir:    real,
	}
	return layout, staging
}

func writeStagingFixture(t *testing.T, stagingDir, id string) {
	t.Helper()
	path := filepath.Join(stagingDir, id+".yaml")
	fixture := Fixture{
		ID:            id,
		RawLog:        "npm ERR! code EINTEGRITY",
		NormalizedLog: "npm ERR! code EINTEGRITY",
		Fingerprint:   id + "-fp",
		FixtureClass:  ClassStaging,
		Review:        ReviewMetadata{Status: "ingested"},
	}
	if err := writeFixture(path, fixture); err != nil {
		t.Fatalf("writeStagingFixture %s: %v", id, err)
	}
	// Set the file path so Load finds it
	_ = path
}

func TestPromoteMovesFixtureToRealDir(t *testing.T) {
	layout, staging := makeStagingLayout(t)
	writeStagingFixture(t, staging, "npm-lockfile-mismatch")

	opts := PromoteOptions{
		ExpectedPlaybook: "npm-lockfile",
		TopN:             1,
		PromotedAt:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	promoted, err := Promote(layout, []string{"npm-lockfile-mismatch"}, opts)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if len(promoted) != 1 {
		t.Fatalf("expected 1 promoted fixture, got %d", len(promoted))
	}
	if promoted[0].ID != "npm-lockfile-mismatch" {
		t.Errorf("expected promoted ID npm-lockfile-mismatch, got %q", promoted[0].ID)
	}
	if promoted[0].FixtureClass != ClassReal {
		t.Errorf("expected ClassReal after promotion, got %q", promoted[0].FixtureClass)
	}
	if promoted[0].Review.Status != "promoted" {
		t.Errorf("expected promoted review status, got %q", promoted[0].Review.Status)
	}
	// Real dir should have the file
	realPath := filepath.Join(layout.RealDir, "npm-lockfile-mismatch.yaml")
	if _, err := os.Stat(realPath); err != nil {
		t.Errorf("expected promoted fixture in real dir: %v", err)
	}
}

func TestPromoteRemovesStagingByDefault(t *testing.T) {
	layout, staging := makeStagingLayout(t)
	writeStagingFixture(t, staging, "docker-auth-fail")

	opts := PromoteOptions{
		ExpectedPlaybook: "docker-auth",
		KeepStaging:      false,
		PromotedAt:       time.Now().UTC(),
	}
	_, err := Promote(layout, []string{"docker-auth-fail"}, opts)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	stagingPath := filepath.Join(staging, "docker-auth-fail.yaml")
	if _, err := os.Stat(stagingPath); err == nil {
		t.Error("expected staging fixture to be removed after promotion")
	}
}

func TestPromoteKeepsStagingWhenOptionSet(t *testing.T) {
	layout, staging := makeStagingLayout(t)
	writeStagingFixture(t, staging, "keep-staging-test")

	opts := PromoteOptions{
		ExpectedPlaybook: "some-playbook",
		KeepStaging:      true,
		PromotedAt:       time.Now().UTC(),
	}
	_, err := Promote(layout, []string{"keep-staging-test"}, opts)
	if err != nil {
		t.Fatalf("Promote with KeepStaging: %v", err)
	}
	stagingPath := filepath.Join(staging, "keep-staging-test.yaml")
	if _, err := os.Stat(stagingPath); err != nil {
		t.Errorf("expected staging fixture to remain when KeepStaging=true: %v", err)
	}
}

func TestPromoteReturnsErrorForMissingID(t *testing.T) {
	layout, _ := makeStagingLayout(t)
	opts := PromoteOptions{PromotedAt: time.Now().UTC()}
	_, err := Promote(layout, []string{"nonexistent-fixture"}, opts)
	if err == nil {
		t.Fatal("expected error for nonexistent staging fixture")
	}
}

func TestPromoteSetsSortedOrder(t *testing.T) {
	layout, staging := makeStagingLayout(t)
	writeStagingFixture(t, staging, "zzz-fixture")
	writeStagingFixture(t, staging, "aaa-fixture")

	opts := PromoteOptions{
		ExpectedPlaybook: "test",
		PromotedAt:       time.Now().UTC(),
	}
	promoted, err := Promote(layout, []string{"zzz-fixture", "aaa-fixture"}, opts)
	if err != nil {
		t.Fatalf("Promote multiple: %v", err)
	}
	if len(promoted) != 2 {
		t.Fatalf("expected 2 promoted, got %d", len(promoted))
	}
	if promoted[0].ID != "aaa-fixture" || promoted[1].ID != "zzz-fixture" {
		t.Errorf("expected sorted order: got %q, %q", promoted[0].ID, promoted[1].ID)
	}
}

func TestPromoteStampsExpectationFields(t *testing.T) {
	layout, staging := makeStagingLayout(t)
	writeStagingFixture(t, staging, "npm-check-fixture")

	fixedTime := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	opts := PromoteOptions{
		ExpectedPlaybook:    "npm-lockfile",
		TopN:                3,
		ExpectedStage:       "build",
		StrictTop1:          true,
		DisallowedPlaybooks: []string{"docker-auth"},
		MinConfidence:       0.7,
		PromotedAt:          fixedTime,
	}
	promoted, err := Promote(layout, []string{"npm-check-fixture"}, opts)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if len(promoted) != 1 {
		t.Fatalf("expected 1 promoted, got %d", len(promoted))
	}
	p := promoted[0]
	if p.Expectation.ExpectedPlaybook != "npm-lockfile" {
		t.Errorf("expected ExpectedPlaybook=npm-lockfile, got %q", p.Expectation.ExpectedPlaybook)
	}
	if p.Expectation.TopN != 3 {
		t.Errorf("expected TopN=3, got %d", p.Expectation.TopN)
	}
	if p.Expectation.ExpectedStage != "build" {
		t.Errorf("expected ExpectedStage=build, got %q", p.Expectation.ExpectedStage)
	}
	if !p.Expectation.StrictTop1 {
		t.Error("expected StrictTop1=true")
	}
	if len(p.Expectation.DisallowedPlaybooks) != 1 || p.Expectation.DisallowedPlaybooks[0] != "docker-auth" {
		t.Errorf("expected DisallowedPlaybooks=[docker-auth], got %v", p.Expectation.DisallowedPlaybooks)
	}
	if p.Expectation.MinConfidence != 0.7 {
		t.Errorf("expected MinConfidence=0.7, got %f", p.Expectation.MinConfidence)
	}
	if p.Review.PromotedAt != fixedTime.Format("2006-01-02T15:04:05Z07:00") {
		t.Errorf("expected PromotedAt timestamp, got %q", p.Review.PromotedAt)
	}
}

func TestIngestDuplicateHandling(t *testing.T) {
	layout, _ := makeStagingLayout(t)
	now := time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC)

	issueSnippet := "npm ERR! code EUSAGE\nnpm ERR! npm ci can only install packages when your package.json and package-lock.json are in sync.\n"
	client := newHandlerClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/acme/widgets/issues/12":
			fmt.Fprint(w, "{\"title\":\"CI fails on npm ci\",\"body\":\"Observed in CI\\n\\n```text\\nnpm ERR! code EUSAGE\\nnpm ERR! npm ci can only install packages when your package.json and package-lock.json are in sync.\\n```\",\"user\":{\"login\":\"alice\"},\"labels\":[{\"name\":\"ci\"}]}")
		case "/repos/acme/widgets/issues/12/comments":
			fmt.Fprint(w, "[{\"id\":91,\"body\":\"A second failing block\\n\\n```text\\nError: Cannot find module 'yaml'\\nRequire stack:\\n- /home/runner/work/index.js\\n```\",\"user\":{\"login\":\"bob\"}}]")
		default:
			http.NotFound(w, r)
		}
	}))
	prefetched, err := GitHubIssueAdapter{}.Fetch(
		context.Background(),
		"https://github.com/acme/widgets/issues/12",
		client,
		now,
	)
	if err != nil {
		t.Fatalf("prefetch fixtures: %v", err)
	}
	if len(prefetched) != 2 {
		t.Fatalf("prefetched fixture count = %d, want 2", len(prefetched))
	}
	if err := writeFixture(filepath.Join(layout.RealDir, "existing-real.yaml"), Fixture{
		ID:            "existing-real",
		RawLog:        issueSnippet,
		NormalizedLog: issueSnippet,
		Fingerprint:   prefetched[0].Fingerprint,
		FixtureClass:  ClassReal,
	}); err != nil {
		t.Fatalf("write existing real fixture: %v", err)
	}

	result, err := Ingest(context.Background(), layout, IngestOptions{
		Adapter: "github-issue",
		URLs: []string{
			"https://github.com/acme/widgets/issues/12",
			"https://gitlab.com/group/widgets/-/issues/99",
		},
		Client: client,
		Now:    now,
	})
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	if len(result.Written) != 1 {
		t.Fatalf("written fixture count = %d, want 1", len(result.Written))
	}
	if result.Written[0].Source.CommentID != "91" {
		t.Fatalf("written fixture source = %+v, want comment fixture", result.Written[0].Source)
	}
	if len(result.Skipped) != 2 {
		t.Fatalf("skipped count = %d, want 2", len(result.Skipped))
	}
	if !strings.HasSuffix(result.Skipped[0], ": duplicate of existing-real") {
		t.Fatalf("duplicate skip = %q", result.Skipped[0])
	}
	if result.Skipped[1] != "https://gitlab.com/group/widgets/-/issues/99: unsupported URL for github-issue" {
		t.Fatalf("unsupported skip = %q", result.Skipped[1])
	}
	if _, err := os.Stat(filepath.Join(layout.StagingDir, result.Written[0].ID+".yaml")); err != nil {
		t.Fatalf("expected ingested fixture file: %v", err)
	}
}

func TestIngestForceAllowsDuplicateFingerprints(t *testing.T) {
	layout, _ := makeStagingLayout(t)
	now := time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC)

	snippet := "npm ERR! code EUSAGE\nnpm ERR! npm ci can only install packages when your package.json and package-lock.json are in sync.\n"
	issue := githubIssue{
		Title: "CI fails on npm ci",
		User:  githubIssueUser{Login: "alice"},
	}
	duplicateFixture := githubFixture("acme/widgets", 12, issue, "", 1, snippet, now)
	if err := writeFixture(filepath.Join(layout.RealDir, "existing-real.yaml"), Fixture{
		ID:            "existing-real",
		RawLog:        snippet,
		NormalizedLog: snippet,
		Fingerprint:   duplicateFixture.Fingerprint,
		FixtureClass:  ClassReal,
	}); err != nil {
		t.Fatalf("write existing real fixture: %v", err)
	}

	client := newHandlerClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/acme/widgets/issues/12":
			fmt.Fprint(w, "{\"title\":\"CI fails on npm ci\",\"body\":\"Observed in CI\\n\\n```text\\nnpm ERR! code EUSAGE\\nnpm ERR! npm ci can only install packages when your package.json and package-lock.json are in sync.\\n```\",\"user\":{\"login\":\"alice\"},\"labels\":[]}")
		case "/repos/acme/widgets/issues/12/comments":
			fmt.Fprint(w, "[]")
		default:
			http.NotFound(w, r)
		}
	}))

	result, err := Ingest(context.Background(), layout, IngestOptions{
		Adapter: "github-issue",
		URLs:    []string{"https://github.com/acme/widgets/issues/12"},
		Client:  client,
		Now:     now,
		Force:   true,
	})
	if err != nil {
		t.Fatalf("Ingest(force): %v", err)
	}
	if len(result.Written) != 1 {
		t.Fatalf("written fixture count = %d, want 1", len(result.Written))
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("skipped = %v, want none", result.Skipped)
	}
}

func TestReviewClassification(t *testing.T) {
	root := t.TempDir()
	layout := Layout{
		Root:       root,
		Fixtures:   filepath.Join(root, "fixtures"),
		MinimalDir: filepath.Join(root, "fixtures", string(ClassMinimal)),
		RealDir:    filepath.Join(root, "fixtures", string(ClassReal)),
		StagingDir: filepath.Join(root, "fixtures", string(ClassStaging)),
	}
	for _, dir := range []string{layout.MinimalDir, layout.RealDir, layout.StagingDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	playbookDir := filepath.Join(root, "playbooks")
	if err := os.MkdirAll(playbookDir, 0o755); err != nil {
		t.Fatalf("mkdir playbooks: %v", err)
	}
	playbook := `id: npm-ci
title: npm ci lockfile mismatch
category: node
severity: high
summary: |
  npm ci failed.
diagnosis: |
  lockfile and package manifest diverged.
fix: |
  regenerate the lockfile.
validation: |
  rerun npm ci.
match:
  any:
    - npm ERR! code EUSAGE
`
	if err := os.WriteFile(filepath.Join(playbookDir, "npm-ci.yaml"), []byte(playbook), 0o644); err != nil {
		t.Fatalf("write playbook: %v", err)
	}

	baseLog := strings.Join([]string{
		"npm ERR! code EUSAGE",
		"line-01",
		"line-02",
		"line-03",
		"line-04",
		"line-05",
		"line-06",
		"line-07",
		"line-08",
		"line-09",
		"line-10",
	}, "\n")
	nearLog := strings.Join([]string{
		"npm ERR! code EUSAGE",
		"line-01",
		"line-02",
		"line-03",
		"line-04",
		"line-05",
		"line-06",
		"line-07",
		"line-08",
		"line-09",
		"line-10-adjusted",
	}, "\n")
	candidateLog := "fatal: could not read Username for 'https://github.com': terminal prompts disabled\n"

	writeFixtureTo := func(dir string, fixture Fixture) {
		t.Helper()
		if err := writeFixture(filepath.Join(dir, fixture.ID+".yaml"), fixture); err != nil {
			t.Fatalf("write fixture %s: %v", fixture.ID, err)
		}
	}

	writeFixtureTo(layout.RealDir, Fixture{
		ID:            "real-base",
		NormalizedLog: baseLog,
		Expectation: Expectation{
			ExpectedPlaybook: "npm-ci",
		},
		FixtureClass: ClassReal,
	})
	makeStagingFixture := func(id, logText string) Fixture {
		return Fixture{
			ID:            id,
			NormalizedLog: logText,
			FixtureClass:  ClassStaging,
			Source: SourceMetadata{
				Adapter:  "github-issue",
				Provider: "github",
				URL:      "https://github.com/acme/widgets/issues/12",
			},
		}
	}
	writeFixtureTo(layout.StagingDir, makeStagingFixture("candidate", candidateLog))
	writeFixtureTo(layout.StagingDir, makeStagingFixture("duplicate", baseLog))
	writeFixtureTo(layout.StagingDir, makeStagingFixture("near", nearLog))

	report, err := Review(layout, EvaluateOptions{PlaybookDir: playbookDir})
	if err != nil {
		t.Fatalf("Review: %v", err)
	}
	if len(report.Items) != 3 {
		t.Fatalf("review item count = %d, want 3", len(report.Items))
	}

	itemsByID := map[string]ReviewItem{}
	for _, item := range report.Items {
		itemsByID[item.Fixture.ID] = item
	}

	if got := itemsByID["duplicate"]; got.Status != "duplicate" || got.DuplicateOf != "real-base" {
		t.Fatalf("duplicate item = %+v", got)
	}
	if got := itemsByID["near"]; got.Status != "near-duplicate" || got.Similarity < 0.82 {
		t.Fatalf("near item = %+v", got)
	}
	if got := itemsByID["candidate"]; got.Status != "candidate" {
		t.Fatalf("candidate item = %+v", got)
	}
	if got := itemsByID["duplicate"].PredictedTopID; got != "npm-ci" {
		t.Fatalf("duplicate predicted top ID = %q, want npm-ci", got)
	}
}
