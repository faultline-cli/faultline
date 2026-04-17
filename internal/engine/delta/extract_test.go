package delta

import (
	"archive/zip"
	"bytes"
	"testing"

	"faultline/internal/model"
)

func TestBuildSnapshotExtractsMinimalDelta(t *testing.T) {
	snapshot := buildSnapshot(
		"github-actions",
		"--- FAIL: TestLockfile\nnpm ERR! package-lock.json is not in sync\n",
		"ok   example.com/app\n",
		[]string{"package.json", "package-lock.json"},
		map[string]model.DeltaEnvChange{
			"head_sha": {Baseline: "abc", Current: "def"},
		},
	)
	if snapshot.Provider != "github-actions" {
		t.Fatalf("expected provider, got %#v", snapshot)
	}
	if len(snapshot.FilesChanged) != 2 {
		t.Fatalf("expected changed files, got %#v", snapshot.FilesChanged)
	}
	if len(snapshot.TestsNewlyFailing) != 1 || snapshot.TestsNewlyFailing[0] != "TestLockfile" {
		t.Fatalf("expected newly failing test, got %#v", snapshot.TestsNewlyFailing)
	}
	if len(snapshot.ErrorsAdded) == 0 {
		t.Fatalf("expected added error lines, got %#v", snapshot.ErrorsAdded)
	}
	if snapshot.EnvDiff["head_sha"].Current != "def" {
		t.Fatalf("expected env diff to be preserved, got %#v", snapshot.EnvDiff)
	}
}

func TestUnzipLogsConcatenatesEntriesDeterministically(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	first, _ := zw.Create("b.log")
	_, _ = first.Write([]byte("second\n"))
	second, _ := zw.Create("a.log")
	_, _ = second.Write([]byte("first\n"))
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	logs, err := unzipLogs(buf.Bytes())
	if err != nil {
		t.Fatalf("unzipLogs: %v", err)
	}
	if got, want := logs, "first\n\nsecond\n"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
