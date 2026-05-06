package fixtures

import (
	"testing"
)

// ── parseGitHubIssueURL ───────────────────────────────────────────────────────

func TestParseGitHubIssueURLValid(t *testing.T) {
	owner, repo, number, err := parseGitHubIssueURL("https://github.com/acme/widgets/issues/42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owner != "acme" {
		t.Errorf("owner = %q, want %q", owner, "acme")
	}
	if repo != "widgets" {
		t.Errorf("repo = %q, want %q", repo, "widgets")
	}
	if number != 42 {
		t.Errorf("number = %d, want 42", number)
	}
}

func TestParseGitHubIssueURLMissingIssuesSegment(t *testing.T) {
	if _, _, _, err := parseGitHubIssueURL("https://github.com/acme/widgets/pull/5"); err == nil {
		t.Fatal("expected error for URL without 'issues' segment, got nil")
	}
}

func TestParseGitHubIssueURLTooFewSegments(t *testing.T) {
	if _, _, _, err := parseGitHubIssueURL("https://github.com/acme"); err == nil {
		t.Fatal("expected error for URL with too few path segments, got nil")
	}
}

func TestParseGitHubIssueURLNonNumericIssueNumber(t *testing.T) {
	if _, _, _, err := parseGitHubIssueURL("https://github.com/acme/widgets/issues/notanumber"); err == nil {
		t.Fatal("expected error for non-numeric issue number, got nil")
	}
}

func TestParseGitHubIssueURLWithQueryAndFragment(t *testing.T) {
	owner, repo, number, err := parseGitHubIssueURL("https://github.com/acme/widgets/issues/42?utm_source=test#issuecomment-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owner != "acme" {
		t.Errorf("owner = %q, want %q", owner, "acme")
	}
	if repo != "widgets" {
		t.Errorf("repo = %q, want %q", repo, "widgets")
	}
	if number != 42 {
		t.Errorf("number = %d, want 42", number)
	}
}

// ── parseGitLabIssueURL ───────────────────────────────────────────────────────

func TestParseGitLabIssueURLValid(t *testing.T) {
	project, number, err := parseGitLabIssueURL("https://gitlab.com/group/widgets/-/issues/34")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project != "group/widgets" {
		t.Errorf("project = %q, want %q", project, "group/widgets")
	}
	if number != 34 {
		t.Errorf("number = %d, want 34", number)
	}
}

func TestParseGitLabIssueURLMissingMarker(t *testing.T) {
	if _, _, err := parseGitLabIssueURL("https://gitlab.com/group/widgets/issues/34"); err == nil {
		t.Fatal("expected error for URL without '-/issues' marker, got nil")
	}
}

func TestParseGitLabIssueURLNonNumericIssueNumber(t *testing.T) {
	if _, _, err := parseGitLabIssueURL("https://gitlab.com/group/widgets/-/issues/notanumber"); err == nil {
		t.Fatal("expected error for non-numeric issue number, got nil")
	}
}

// ── parseDiscourseTopicURL ────────────────────────────────────────────────────

func TestParseDiscourseTopicURLValid(t *testing.T) {
	topicURL, host, id, err := parseDiscourseTopicURL("https://discuss.example.org/t/some-topic/30640")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if host != "discuss.example.org" {
		t.Errorf("host = %q, want %q", host, "discuss.example.org")
	}
	if id != 30640 {
		t.Errorf("id = %d, want 30640", id)
	}
	if topicURL == "" {
		t.Error("expected non-empty topicURL")
	}
}

func TestParseDiscourseTopicURLWithPostNumber(t *testing.T) {
	// URL with a post number appended (e.g. /t/slug/123/4) should still parse correctly.
	topicURL, host, id, err := parseDiscourseTopicURL("https://discuss.example.org/t/some-topic/30640/3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if host != "discuss.example.org" {
		t.Errorf("host = %q, want %q", host, "discuss.example.org")
	}
	if id != 30640 {
		t.Errorf("id = %d, want 30640", id)
	}
	if topicURL == "" {
		t.Error("expected non-empty topicURL")
	}
}

func TestParseDiscourseTopicURLMissingTSegment(t *testing.T) {
	if _, _, _, err := parseDiscourseTopicURL("https://discuss.example.org/c/some-category/5"); err == nil {
		t.Fatal("expected error for URL without '/t/' segment, got nil")
	}
}

func TestParseDiscourseTopicURLNonNumericID(t *testing.T) {
	if _, _, _, err := parseDiscourseTopicURL("https://discuss.example.org/t/some-topic/notanumber"); err == nil {
		t.Fatal("expected error for non-numeric topic ID, got nil")
	}
}

// ── parseRedditPostURL ────────────────────────────────────────────────────────

func TestParseRedditPostURLValid(t *testing.T) {
	subreddit, postID, postURL, err := parseRedditPostURL("https://www.reddit.com/r/golang/comments/abc123/my_post/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subreddit != "golang" {
		t.Errorf("subreddit = %q, want %q", subreddit, "golang")
	}
	if postID != "abc123" {
		t.Errorf("postID = %q, want %q", postID, "abc123")
	}
	if postURL == "" {
		t.Error("expected non-empty postURL")
	}
}

func TestParseRedditPostURLOldReddit(t *testing.T) {
	subreddit, postID, postURL, err := parseRedditPostURL("https://old.reddit.com/r/docker/comments/xyz789/title/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subreddit != "docker" {
		t.Errorf("subreddit = %q, want %q", subreddit, "docker")
	}
	if postID != "xyz789" {
		t.Errorf("postID = %q, want %q", postID, "xyz789")
	}
	if postURL == "" {
		t.Error("expected non-empty postURL")
	}
}

func TestParseRedditPostURLUnsupportedHost(t *testing.T) {
	if _, _, _, err := parseRedditPostURL("https://www.example.com/r/test/comments/abc/title/"); err == nil {
		t.Fatal("expected error for unsupported host, got nil")
	}
}

func TestParseRedditPostURLMissingCommentsSegment(t *testing.T) {
	if _, _, _, err := parseRedditPostURL("https://www.reddit.com/r/golang/"); err == nil {
		t.Fatal("expected error when URL has no 'comments' segment, got nil")
	}
}

// ── siteFromStackExchangeHost ─────────────────────────────────────────────────

func TestSiteFromStackExchangeHostKnownHosts(t *testing.T) {
	cases := []struct {
		host string
		want string
	}{
		{"stackoverflow.com", "stackoverflow"},
		{"superuser.com", "superuser"},
		{"serverfault.com", "serverfault"},
		{"askubuntu.com", "askubuntu"},
		{"unix.stackexchange.com", "unix.stackexchange"},
		{"meta.stackexchange.com", "meta.stackexchange"},
		{"unknown.example.com", ""},
	}
	for _, tc := range cases {
		got := siteFromStackExchangeHost(tc.host)
		if got != tc.want {
			t.Errorf("siteFromStackExchangeHost(%q) = %q, want %q", tc.host, got, tc.want)
		}
	}
}

