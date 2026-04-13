package fixtures

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGitHubIssueAdapterFetch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/widgets/issues/12", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{\"title\":\"CI fails on npm ci\",\"body\":\"Observed in CI\\n\\n```text\\nnpm ERR! code EUSAGE\\nnpm ERR! npm ci can only install packages when your package.json and package-lock.json are in sync.\\nError: Process completed with exit code 1.\\n```\",\"user\":{\"login\":\"alice\"},\"labels\":[{\"name\":\"ci\"}]}")
	})
	mux.HandleFunc("/repos/acme/widgets/issues/12/comments", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "[{\"id\":91,\"body\":\"A second failing block\\n\\n```text\\nError: Cannot find module 'yaml'\\nRequire stack:\\n- /home/runner/work/index.js\\n```\",\"user\":{\"login\":\"bob\"}}]")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	adapter := GitHubIssueAdapter{APIBase: server.URL}
	fixtures, err := adapter.Fetch(context.Background(), "https://github.com/acme/widgets/issues/12", server.Client(), time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("fetch GitHub issue fixtures: %v", err)
	}
	if len(fixtures) != 2 {
		t.Fatalf("expected 2 extracted fixtures, got %d", len(fixtures))
	}
	if fixtures[0].Source.Provider != "github" {
		t.Fatalf("expected github provider metadata, got %#v", fixtures[0].Source)
	}
	if fixtures[0].Source.Repository != "acme/widgets" {
		t.Fatalf("expected repository metadata, got %#v", fixtures[0].Source)
	}
	if fixtures[0].NormalizedLog == "" {
		t.Fatal("expected normalized log content")
	}
}

func TestGitLabIssueAdapterFetch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/group%2Fwidgets/issues/34", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{\"title\":\"Deploy fails with CrashLoopBackOff\",\"description\":\"From the incident\\n\\n```text\\nBack-off restarting failed container\\nCrashLoopBackOff\\nReadiness probe failed: HTTP probe failed with statuscode: 500\\n```\",\"author\":{\"username\":\"carol\"},\"labels\":[\"ci\",\"kubernetes\"]}")
	})
	mux.HandleFunc("/api/v4/projects/group%2Fwidgets/issues/34/notes", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "[{\"id\":7,\"body\":\"```text\\nlookup registry-1.docker.io: temporary failure in name resolution\\nError response from daemon\\npull access denied\\n```\",\"author\":{\"username\":\"dave\"}}]")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	adapter := GitLabIssueAdapter{APIBase: server.URL + "/api/v4"}
	fixtures, err := adapter.Fetch(context.Background(), "https://gitlab.com/group/widgets/-/issues/34", server.Client(), time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("fetch GitLab issue fixtures: %v", err)
	}
	if len(fixtures) != 2 {
		t.Fatalf("expected 2 extracted fixtures, got %d", len(fixtures))
	}
	if fixtures[0].Source.Provider != "gitlab" {
		t.Fatalf("expected gitlab provider metadata, got %#v", fixtures[0].Source)
	}
	if fixtures[0].Source.Repository != "group/widgets" {
		t.Fatalf("expected project metadata, got %#v", fixtures[0].Source)
	}
	if fixtures[0].NormalizedLog == "" {
		t.Fatal("expected normalized log content")
	}
}
