package fixtures

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"faultline/internal/engine"
)

func TestLoadAllMergesMinimalAndRealOnly(t *testing.T) {
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

	writeFixtureFile := func(dir string, fixture Fixture) {
		t.Helper()
		if err := writeFixture(filepath.Join(dir, fixture.ID+".yaml"), fixture); err != nil {
			t.Fatalf("write fixture %s: %v", fixture.ID, err)
		}
	}

	writeFixtureFile(layout.MinimalDir, Fixture{ID: "b-minimal", RawLog: "minimal failure"})
	writeFixtureFile(layout.RealDir, Fixture{ID: "a-real", RawLog: "real failure"})
	writeFixtureFile(layout.StagingDir, Fixture{ID: "z-staging", RawLog: "staging failure"})

	fixtures, err := Load(layout, ClassAll)
	if err != nil {
		t.Fatalf("Load(ClassAll): %v", err)
	}

	gotIDs := []string{fixtures[0].ID, fixtures[1].ID}
	wantIDs := []string{"b-minimal", "a-real"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("loaded IDs = %v, want %v", gotIDs, wantIDs)
	}
}

func TestLoadManifestUsesRelativeLogPathAndComputesFingerprint(t *testing.T) {
	root := t.TempDir()
	layout := Layout{
		Root:       root,
		Fixtures:   filepath.Join(root, "fixtures"),
		MinimalDir: filepath.Join(root, "fixtures", string(ClassMinimal)),
	}
	if err := os.MkdirAll(layout.MinimalDir, 0o755); err != nil {
		t.Fatalf("mkdir minimal dir: %v", err)
	}

	logPath := filepath.Join(root, "logs", "sample.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("mkdir logs dir: %v", err)
	}
	rawLog := "first line\r\nsecond line\r\n"
	if err := os.WriteFile(logPath, []byte(rawLog), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	manifestPath := filepath.Join(layout.MinimalDir, "manifest.yaml")
	manifest := "fixtures:\n" +
		"  - id: fixture-from-path\n" +
		"    path: logs/sample.log\n"
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	fixtures, err := Load(layout, ClassMinimal)
	if err != nil {
		t.Fatalf("Load(ClassMinimal): %v", err)
	}
	if len(fixtures) != 1 {
		t.Fatalf("expected 1 fixture, got %d", len(fixtures))
	}
	fixture := fixtures[0]
	if fixture.ID != "fixture-from-path" {
		t.Fatalf("fixture ID = %q, want fixture-from-path", fixture.ID)
	}
	if fixture.FilePath != manifestPath {
		t.Fatalf("FilePath = %q, want %q", fixture.FilePath, manifestPath)
	}
	if fixture.FixtureClass != ClassMinimal {
		t.Fatalf("FixtureClass = %q, want %q", fixture.FixtureClass, ClassMinimal)
	}
	wantFingerprint := FingerprintForLog(engine.CanonicalizeLog(rawLog))
	if fixture.Fingerprint != wantFingerprint {
		t.Fatalf("Fingerprint = %q, want %q", fixture.Fingerprint, wantFingerprint)
	}
}

func TestFixtureLogRejectsMissingContent(t *testing.T) {
	_, err := fixtureLog(Fixture{ID: "missing-log"}, t.TempDir())
	if err == nil {
		t.Fatal("expected missing log content error")
	}
}
