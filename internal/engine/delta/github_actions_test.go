package delta

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolverGitHubActionsBuildsSnapshot(t *testing.T) {
	logArchive := func(body string) []byte {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		file, _ := zw.Create("job.txt")
		_, _ = file.Write([]byte(body))
		_ = zw.Close()
		return buf.Bytes()
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/acme/faultline/actions/runs/200":
			_ = json.NewEncoder(w).Encode(githubRun{
				ID:         200,
				HeadBranch: "main",
				HeadSHA:    "head123",
				Event:      "push",
				RunAttempt: 1,
				WorkflowID: 99,
				Path:       ".github/workflows/ci.yml",
			})
		case "/repos/acme/faultline/actions/workflows/99/runs":
			_ = json.NewEncoder(w).Encode(githubWorkflowRuns{
				WorkflowRuns: []githubRun{{
					ID:         100,
					HeadBranch: "main",
					HeadSHA:    "base123",
					Conclusion: "success",
					Event:      "push",
					RunAttempt: 1,
					WorkflowID: 99,
					Path:       ".github/workflows/ci.yml",
				}},
			})
		case "/repos/acme/faultline/actions/runs/100/logs":
			_, _ = w.Write(logArchive("ok   example.com/app\n"))
		case "/repos/acme/faultline/compare/base123...head123":
			_ = json.NewEncoder(w).Encode(githubCompare{
				Files: []struct {
					Filename string `json:"filename"`
				}{
					{Filename: "package.json"},
					{Filename: "package-lock.json"},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	resolver := NewResolver(server.Client())
	snapshot, err := resolver.Resolve(context.Background(), Options{
		Provider: "github-actions",
		GitHub: GitHubOptions{
			Repository: "acme/faultline",
			RunID:      200,
			Token:      "test-token",
			APIBaseURL: server.URL,
		},
	}, "--- FAIL: TestLockfile\nnpm ERR! package-lock.json is not in sync\n")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected snapshot")
	}
	if len(snapshot.FilesChanged) != 2 {
		t.Fatalf("expected compare files, got %#v", snapshot)
	}
	if len(snapshot.TestsNewlyFailing) != 1 || snapshot.TestsNewlyFailing[0] != "TestLockfile" {
		t.Fatalf("expected failing test delta, got %#v", snapshot.TestsNewlyFailing)
	}
	if snapshot.EnvDiff["head_sha"].Baseline != "base123" {
		t.Fatalf("expected env diff, got %#v", snapshot.EnvDiff)
	}
}
