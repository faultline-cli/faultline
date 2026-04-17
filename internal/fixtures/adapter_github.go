package fixtures

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type GitHubIssueAdapter struct {
	APIBase string
}

func (a GitHubIssueAdapter) Name() string {
	return "github-issue"
}

func (a GitHubIssueAdapter) Supports(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, "github.com") && strings.Contains(u.Path, "/issues/")
}

func (a GitHubIssueAdapter) Fetch(ctx context.Context, rawURL string, client *http.Client, now time.Time) ([]Fixture, error) {
	owner, repo, issueNumber, err := parseGitHubIssueURL(rawURL)
	if err != nil {
		return nil, err
	}
	apiBase := strings.TrimRight(a.APIBase, "/")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	client = defaultHTTPClient(client)
	requestOpts := jsonRequestOptions{
		AcceptHeader:       "application/vnd.github+json, application/json",
		OptionalStatusCodes: []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}

	var issue githubIssue
	if err := getJSON(ctx, client, fmt.Sprintf("%s/repos/%s/%s/issues/%d", apiBase, owner, repo, issueNumber), &issue, requestOpts); err != nil {
		return nil, err
	}
	var comments []githubIssueComment
	if err := getJSONOptional(ctx, client, fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments?per_page=100", apiBase, owner, repo, issueNumber), &comments, requestOpts); err != nil {
		return nil, err
	}

	repository := owner + "/" + repo
	var fixtures []Fixture
	for index, snippet := range extractLogSnippets(issue.Body) {
		fixtures = append(fixtures, githubFixture(repository, issueNumber, issue, "", index+1, snippet, now))
	}
	for _, comment := range comments {
		for index, snippet := range extractLogSnippets(comment.Body) {
			fixtures = append(fixtures, githubFixture(repository, issueNumber, issue, strconv.Itoa(comment.ID), index+1, snippet, now))
		}
	}
	return fixtures, nil
}

type githubIssue struct {
	Title  string             `json:"title"`
	Body   string             `json:"body"`
	User   githubIssueUser    `json:"user"`
	Labels []githubIssueLabel `json:"labels"`
}

type githubIssueUser struct {
	Login string `json:"login"`
}

type githubIssueLabel struct {
	Name string `json:"name"`
}

type githubIssueComment struct {
	ID   int             `json:"id"`
	Body string          `json:"body"`
	User githubIssueUser `json:"user"`
}

func parseGitHubIssueURL(rawURL string) (string, string, int, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", 0, fmt.Errorf("parse GitHub issue URL: %w", err)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != "issues" {
		return "", "", 0, fmt.Errorf("unsupported GitHub issue URL %q", rawURL)
	}
	issueNumber, err := strconv.Atoi(parts[3])
	if err != nil {
		return "", "", 0, fmt.Errorf("parse GitHub issue number from %q: %w", rawURL, err)
	}
	return parts[0], parts[1], issueNumber, nil
}

func githubFixture(repository string, issueNumber int, issue githubIssue, commentID string, snippetIndex int, snippet string, now time.Time) Fixture {
	labels := make([]string, 0, len(issue.Labels))
	for _, label := range issue.Labels {
		labels = append(labels, label.Name)
	}
	return Fixture{
		ID:            buildFixtureID("github", repository, issueNumber, snippetIndex, snippet),
		Title:         issue.Title,
		FixtureClass:  ClassStaging,
		NormalizedLog: snippet,
		Source: SourceMetadata{
			Adapter:     "github-issue",
			Provider:    "github",
			URL:         fmt.Sprintf("https://github.com/%s/issues/%d", repository, issueNumber),
			Repository:  repository,
			IssueNumber: issueNumber,
			CommentID:   commentID,
			Title:       issue.Title,
			Labels:      labels,
			Author:      issue.User.Login,
			Snippet:     snippetIndex,
			FetchedAt:   now.UTC().Format(time.RFC3339),
		},
		Review: ReviewMetadata{Status: "ingested"},
	}
}
