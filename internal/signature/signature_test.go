package signature

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/engine"
	"faultline/internal/model"
)

func TestNormalizeEvidenceLineRemovesDynamicNoise(t *testing.T) {
	line := `2026-04-22T12:05:31Z /home/runner/work/app/app/internal/service/user.go:43:19: request 9fd46ec5-6c4f-4f0c-a11d-f3c96f172d63 failed for commit 71E944493FA59840 on 10.24.6.9`
	got := NormalizeEvidenceLine(line)
	want := `<timestamp> <workspace>/internal/service/user.go:<n> request <id> failed for commit <hex> on <ip>`
	if got != want {
		t.Fatalf("unexpected normalized line:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestNormalizeEvidenceLineNormalizesOSPathsAndPreservesExitCodes(t *testing.T) {
	left := NormalizeEvidenceLine(`/tmp/build/output.log:12: exec /__e/node20/bin/node: no such file or directory exit code 127`)
	right := NormalizeEvidenceLine(`C:\Users\runneradmin\AppData\Local\Temp\build\output.log:98: exec D:\a\_temp\node20\bin\node: no such file or directory exit code 127`)
	if left != right {
		t.Fatalf("expected path variants to normalize the same:\nleft:  %q\nright: %q", left, right)
	}
	if !strings.Contains(left, "exit code 127") {
		t.Fatalf("expected exit code to be preserved, got %q", left)
	}
}

func TestNormalizeEvidenceLinesSplitsMultilineDeterministically(t *testing.T) {
	lines := NormalizeEvidenceLines(" first line \n\nSecond line\twith  spaces\r\nfirst line\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 normalized lines, got %#v", lines)
	}
	if lines[0] != "first line" || lines[1] != "second line with spaces" || lines[2] != "first line" {
		t.Fatalf("unexpected multiline normalization: %#v", lines)
	}
}

func TestForResultUsesStructuredTriggerAttributes(t *testing.T) {
	sig := ForResult(model.Result{
		Playbook: model.Playbook{ID: "panic-in-http-handler"},
		Detector: "source",
		Evidence: []string{"handler panic near api/users"},
		EvidenceBy: model.EvidenceBundle{
			Triggers: []model.Evidence{{
				SignalID:  "panic.handler",
				File:      "internal/api/handler.go",
				ScopeName: "ServeUser",
				Detail:    "panic(\"boom\")",
			}},
		},
	})
	if sig.Hash == "" {
		t.Fatal("expected signature hash")
	}
	if sig.Payload.Attributes == nil {
		t.Fatal("expected structured trigger attributes")
	}
	if got := sig.Payload.Attributes.Files; len(got) != 1 || got[0] != "internal/api/handler.go" {
		t.Fatalf("expected structured file attribute, got %#v", got)
	}
	if got := sig.Payload.Attributes.ScopeNames; len(got) != 1 || got[0] != "serveuser" {
		t.Fatalf("expected normalized scope name, got %#v", got)
	}
	if got := sig.Payload.Attributes.SignalIDs; len(got) != 1 || got[0] != "panic.handler" {
		t.Fatalf("expected normalized signal id, got %#v", got)
	}
}

func TestForResultStableAcrossEquivalentLogVariants(t *testing.T) {
	left := ForResult(model.Result{
		Playbook: model.Playbook{ID: "missing-executable"},
		Detector: "log",
		Evidence: []string{
			"2026-04-22T12:05:31Z /home/runner/work/faultline/faultline/.github/workflows/ci.yml:118: exec /__e/node20/bin/node: no such file or directory",
		},
	})
	right := ForResult(model.Result{
		Playbook: model.Playbook{ID: "missing-executable"},
		Detector: "log",
		Evidence: []string{
			"2026-04-23T07:15:44Z D:\\a\\faultline\\faultline\\.github\\workflows\\ci.yml:241: exec D:\\a\\_temp\\node20\\bin\\node: no such file or directory",
		},
	})
	if left.Hash != right.Hash {
		t.Fatalf("expected equivalent noisy variants to share a hash:\nleft payload:  %s\nright payload: %s", left.Normalized, right.Normalized)
	}
}

func TestForResultDistinctFailuresStayDistinct(t *testing.T) {
	a := ForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Detector: "log",
		Evidence: []string{"pull access denied for registry.example.com"},
	})
	b := ForResult(model.Result{
		Playbook: model.Playbook{ID: "network-timeout"},
		Detector: "log",
		Evidence: []string{"context deadline exceeded while waiting for upstream"},
	})
	if a.Hash == b.Hash {
		t.Fatalf("expected distinct failures to have distinct signatures: %s", a.Hash)
	}
}

func TestForResultPayloadSnapshot(t *testing.T) {
	sig := ForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Detector: "log",
		Evidence: []string{
			"2026-04-22T12:05:31Z Error response from daemon: pull access denied for mcr/microsoft.com/mssql/server, repository does not exist or may require 'docker login'",
			"/home/runner/work/faultline/faultline/.github/workflows/ci.yml:118: exec /__e/node20/bin/node: no such file or directory",
		},
	})
	want := `{"version":"signature.v1","failure_id":"docker-auth","detector":"log","evidence":["<timestamp> error response from daemon: pull access denied for mcr/microsoft.com/mssql/server, repository does not exist or may require 'docker login'","<workspace>/.github/workflows/ci.yml:<n> exec <runner>/node20/bin/node no such file or directory"]}`
	if sig.Normalized != want {
		t.Fatalf("unexpected canonical payload:\nwant: %s\ngot:  %s", want, sig.Normalized)
	}
}

func TestFixtureDrivenSignatureStabilityAcrossRealLogs(t *testing.T) {
	root := repoRoot(t)
	playbookDir := filepath.Join(root, "playbooks", "bundled")
	cases := []struct {
		name string
		path string
	}{
		{name: "missing executable", path: filepath.Join(root, "internal", "engine", "testdata", "corpus", "missing-executable-noisy.log")},
		{name: "dependency drift", path: filepath.Join(root, "internal", "engine", "testdata", "fixtures", "dependency-drift.log")},
		{name: "lockfile drift", path: filepath.Join(root, "internal", "engine", "testdata", "fixtures", "npm-ci-lockfile.log")},
		{name: "flaky test", path: filepath.Join(root, "internal", "engine", "testdata", "fixtures", "flaky-test.log")},
	}
	eng := engine.New(engine.Options{PlaybookDir: playbookDir})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.path)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			analysis, err := eng.AnalyzeReader(strings.NewReader(string(data)))
			if err != nil && err != engine.ErrNoMatch {
				t.Fatalf("AnalyzeReader: %v", err)
			}
			if analysis == nil || len(analysis.Results) == 0 {
				t.Fatalf("expected a matched result for fixture %s", tc.path)
			}
			sig := ForResult(analysis.Results[0])
			if sig.Hash == "" {
				t.Fatalf("expected signature hash for fixture %s", tc.path)
			}
			if !strings.Contains(sig.Normalized, `"version":"signature.v1"`) {
				t.Fatalf("expected versioned canonical payload, got %s", sig.Normalized)
			}
		})
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repository root")
		}
		dir = parent
	}
}
