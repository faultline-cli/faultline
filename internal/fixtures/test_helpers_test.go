package fixtures

import (
	"path/filepath"
	"testing"
)

func writeFixtureInDir(t *testing.T, dir string, fixture Fixture) {
	t.Helper()
	if err := writeFixture(filepath.Join(dir, fixture.ID+".yaml"), fixture); err != nil {
		t.Fatalf("write fixture %s: %v", fixture.ID, err)
	}
}
