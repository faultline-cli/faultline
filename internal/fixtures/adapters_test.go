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

func TestGitHubIssueAdapterIgnoresCommentAuthFailures(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/widgets/issues/13", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{\"title\":\"CI fails on missing binary\",\"body\":\"Observed in CI\\n\\n```text\\nexec /usr/bin/tool: no such file or directory\\n```\",\"user\":{\"login\":\"alice\"},\"labels\":[{\"name\":\"ci\"}]}")
	})
	mux.HandleFunc("/repos/acme/widgets/issues/13/comments", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "{\"message\":\"forbidden\"}")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	adapter := GitHubIssueAdapter{APIBase: server.URL}
	fixtures, err := adapter.Fetch(context.Background(), "https://github.com/acme/widgets/issues/13", server.Client(), time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("fetch GitHub issue fixtures: %v", err)
	}
	if len(fixtures) != 1 {
		t.Fatalf("expected 1 extracted fixture when comments are unavailable, got %d", len(fixtures))
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

func TestGitLabIssueAdapterIgnoresNoteAuthFailures(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/group%2Fwidgets/issues/35", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{\"title\":\"Deploy fails with context deadline exceeded\",\"description\":\"Observed in CI\\n\\n```text\\ncontext deadline exceeded\\n```\",\"author\":{\"username\":\"carol\"},\"labels\":[\"ci\",\"kubernetes\"]}")
	})
	mux.HandleFunc("/api/v4/projects/group%2Fwidgets/issues/35/notes", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "{\"message\":\"unauthorized\"}")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	adapter := GitLabIssueAdapter{APIBase: server.URL + "/api/v4"}
	fixtures, err := adapter.Fetch(context.Background(), "https://gitlab.com/group/widgets/-/issues/35", server.Client(), time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("fetch GitLab issue fixtures: %v", err)
	}
	if len(fixtures) != 1 {
		t.Fatalf("expected 1 extracted fixture when notes are unavailable, got %d", len(fixtures))
	}
}

func TestStackExchangeQuestionAdapterFetch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/2.3/questions/12345", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{\"items\":[{\"question_id\":12345,\"title\":\"Docker executable file not found\",\"body\":\"<p>Observed in CI</p><pre><code>exec: \\\"grunt\\\": executable file not found in $PATH</code></pre>\",\"owner\":{\"display_name\":\"alice\"}}]}")
	})
	mux.HandleFunc("/2.3/questions/12345/answers", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{\"items\":[{\"answer_id\":67890,\"body\":\"<p>Use an absolute path.</p><pre><code>permission denied while starting container</code></pre>\",\"owner\":{\"display_name\":\"bob\"}}]}")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	adapter := StackExchangeQuestionAdapter{APIBase: server.URL + "/2.3"}
	fixtures, err := adapter.Fetch(context.Background(), "https://stackoverflow.com/questions/12345/docker-executable-file-not-found", server.Client(), time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("fetch Stack Exchange question fixtures: %v", err)
	}
	if len(fixtures) != 2 {
		t.Fatalf("expected 2 extracted fixtures, got %d", len(fixtures))
	}
	if fixtures[0].Source.Provider != "stackexchange" {
		t.Fatalf("expected stackexchange provider metadata, got %#v", fixtures[0].Source)
	}
	if fixtures[0].Source.Repository != "stackoverflow" {
		t.Fatalf("expected site metadata, got %#v", fixtures[0].Source)
	}
	if fixtures[0].NormalizedLog == "" {
		t.Fatal("expected normalized log content")
	}
}

func TestDiscourseTopicAdapterFetch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/t/discourse-doesnt-deliver-webpages-fresh-install-on-linode-ubuntu-14-04/30640.json", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{\"title\":\"Discourse doesn't deliver webpages\",\"post_stream\":{\"posts\":[{\"id\":101,\"username\":\"alice\",\"cooked\":\"<p>connection refused</p><pre><code>curl: (7) Failed to connect to localhost port 8080: Connection refused</code></pre>\",\"raw\":\"\"},{\"id\":102,\"username\":\"bob\",\"cooked\":\"<p>permission denied</p><pre><code>docker: permission denied while trying to connect to the Docker daemon socket</code></pre>\",\"raw\":\"\"}]}}")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	adapter := DiscourseTopicAdapter{}
	fixtures, err := adapter.Fetch(context.Background(), server.URL+"/t/discourse-doesnt-deliver-webpages-fresh-install-on-linode-ubuntu-14-04/30640", server.Client(), time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("fetch Discourse topic fixtures: %v", err)
	}
	if len(fixtures) != 2 {
		t.Fatalf("expected 2 extracted fixtures, got %d", len(fixtures))
	}
	if fixtures[0].Source.Provider != "discourse" {
		t.Fatalf("expected discourse provider metadata, got %#v", fixtures[0].Source)
	}
	if fixtures[0].Source.Repository == "" {
		t.Fatalf("expected forum host metadata, got %#v", fixtures[0].Source)
	}
}

func TestRedditPostAdapterFetch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/r/docker/comments/1fbi7v2/ssh_docker_daemon.json", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "[{\"data\":{\"children\":[{\"kind\":\"t3\",\"data\":{\"title\":\"SSH Docker Daemon\",\"selftext\":\"I tried to build images on a remote vps.\\n\\n```text\\nConnection refused\\n```\",\"author\":\"Pandoks_\",\"id\":\"1fbi7v2\",\"permalink\":\"/r/docker/comments/1fbi7v2/ssh_docker_daemon/\"}}]}},{\"data\":{\"children\":[{\"kind\":\"t1\",\"data\":{\"body\":\"Thanks for the help\",\"author\":\"alice\",\"id\":\"comment-1\",\"replies\":{\"data\":{\"children\":[{\"kind\":\"t1\",\"data\":{\"body\":\"```text\\ndocker: permission denied while trying to connect to the Docker daemon socket\\n```\",\"author\":\"bob\",\"id\":\"comment-2\",\"replies\":\"\"}}]}}}}]}}]")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	adapter := RedditPostAdapter{APIBase: server.URL}
	fixtures, err := adapter.Fetch(context.Background(), "https://www.reddit.com/r/docker/comments/1fbi7v2/ssh_docker_daemon/", server.Client(), time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("fetch Reddit post fixtures: %v", err)
	}
	if len(fixtures) != 2 {
		t.Fatalf("expected 2 extracted fixtures, got %d", len(fixtures))
	}
	if fixtures[0].Source.Provider != "reddit" {
		t.Fatalf("expected reddit provider metadata, got %#v", fixtures[0].Source)
	}
	if fixtures[0].Source.Repository != "r/docker" {
		t.Fatalf("expected subreddit repository metadata, got %#v", fixtures[0].Source)
	}
}
