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

func TestResolverGitLabCIBuildsSnapshot(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			response := &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Request:    r,
			}
			switch r.URL.Path {
			case "/api/v4/projects/group%2Fwidgets/jobs/902":
				body := `{"id":902,"name":"test","stage":"verify","status":"failed","pipeline":{"id":77,"sha":"head123","ref":"main","source":"push"}}`
				response.Body = ioNopCloser(bytes.NewReader([]byte(body)))
			case "/api/v4/projects/group%2Fwidgets/pipelines":
				body := `[{"id":76,"sha":"base123","ref":"main","source":"push","status":"success"}]`
				response.Body = ioNopCloser(bytes.NewReader([]byte(body)))
			case "/api/v4/projects/group%2Fwidgets/pipelines/76/jobs":
				body := `[{"id":701,"name":"test","stage":"verify","status":"success","pipeline":{"id":76,"sha":"base123","ref":"main","source":"push"}}]`
				response.Body = ioNopCloser(bytes.NewReader([]byte(body)))
			case "/api/v4/projects/group%2Fwidgets/jobs/701/trace":
				response.Body = ioNopCloser(bytes.NewReader([]byte("ok   example.com/app\n")))
			case "/api/v4/projects/group%2Fwidgets/repository/compare":
				body := `{"diffs":[{"old_path":"package.json","new_path":"package.json"},{"old_path":"package-lock.json","new_path":"package-lock.json"}]}`
				response.Body = ioNopCloser(bytes.NewReader([]byte(body)))
			default:
				t.Fatalf("unexpected path %s", r.URL.Path)
			}
			return response, nil
		}),
	}

	resolver := NewResolver(client)
	snapshot, err := resolver.Resolve(context.Background(), Options{
		Provider: "gitlab-ci",
		GitLab: GitLabOptions{
			Project:    "group/widgets",
			PipelineID: 77,
			JobID:      902,
			Token:      "test-token",
			APIBaseURL: "https://gitlab.test/api/v4",
		},
	}, "--- FAIL: TestLockfile\nnpm ERR! package-lock.json is not in sync\n")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected snapshot")
	}
	if snapshot.Provider != "gitlab-ci" {
		t.Fatalf("expected gitlab-ci provider, got %#v", snapshot)
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

func TestNormalizeProvider(t *testing.T) {
	cases := map[string]string{
		"":               "",
		"none":           "none",
		"github":         "github-actions",
		"github-actions": "github-actions",
		"GHA":            "github-actions",
		"gitlab":         "gitlab-ci",
		"gitlab-ci":      "gitlab-ci",
		"gitlab-cicd":    "gitlab-ci",
		"GL":             "gitlab-ci",
		"unknown":        "unknown",
	}
	for in, want := range cases {
		if got := normalizeProvider(in); got != want {
			t.Fatalf("normalizeProvider(%q): got %q, want %q", in, got, want)
		}
	}
}

type nopCloser struct {
	*bytes.Reader
}

func (nopCloser) Close() error { return nil }

func ioNopCloser(r *bytes.Reader) nopCloser {
	return nopCloser{Reader: r}
}
