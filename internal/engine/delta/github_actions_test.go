package delta

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestResolverGitHubActionsBuildsSnapshot(t *testing.T) {
	logArchive := func(body string) []byte {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		file, _ := zw.Create("job.txt")
		_, _ = file.Write([]byte(body))
		_ = zw.Close()
		return buf.Bytes()
	}

	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			response := &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Request:    r,
			}
			switch r.URL.Path {
			case "/repos/acme/faultline/actions/runs/200":
				body, _ := json.Marshal(githubRun{
					ID:         200,
					HeadBranch: "main",
					HeadSHA:    "head123",
					Event:      "push",
					RunAttempt: 1,
					WorkflowID: 99,
					Path:       ".github/workflows/ci.yml",
				})
				response.Body = ioNopCloser(bytes.NewReader(body))
			case "/repos/acme/faultline/actions/workflows/99/runs":
				body, _ := json.Marshal(githubWorkflowRuns{
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
				response.Body = ioNopCloser(bytes.NewReader(body))
			case "/repos/acme/faultline/actions/runs/100/logs":
				response.Body = ioNopCloser(bytes.NewReader(logArchive("ok   example.com/app\n")))
			case "/repos/acme/faultline/compare/base123...head123":
				body, _ := json.Marshal(githubCompare{
					Files: []struct {
						Filename string `json:"filename"`
					}{
						{Filename: "package.json"},
						{Filename: "package-lock.json"},
					},
				})
				response.Body = ioNopCloser(bytes.NewReader(body))
			default:
				t.Fatalf("unexpected path %s", r.URL.Path)
			}
			return response, nil
		}),
	}

	resolver := NewResolver(client)
	snapshot, err := resolver.Resolve(context.Background(), Options{
		Provider: "github-actions",
		GitHub: GitHubOptions{
			Repository: "acme/faultline",
			RunID:      200,
			Token:      "test-token",
			APIBaseURL: "https://github.test",
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

type nopCloser struct {
	*bytes.Reader
}

func (nopCloser) Close() error { return nil }

func ioNopCloser(r *bytes.Reader) nopCloser {
	return nopCloser{Reader: r}
}
