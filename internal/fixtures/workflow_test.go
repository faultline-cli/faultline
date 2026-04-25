package fixtures

import (
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
