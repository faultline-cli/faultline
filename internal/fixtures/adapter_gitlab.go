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

type GitLabIssueAdapter struct {
	APIBase string
}

func (a GitLabIssueAdapter) Name() string {
	return "gitlab-issue"
}

func (a GitLabIssueAdapter) Supports(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, "gitlab.com") && strings.Contains(u.Path, "/-/issues/")
}

func (a GitLabIssueAdapter) Fetch(ctx context.Context, rawURL string, client *http.Client, now time.Time) ([]Fixture, error) {
	project, issueNumber, err := parseGitLabIssueURL(rawURL)
	if err != nil {
		return nil, err
	}
	apiBase := strings.TrimRight(a.APIBase, "/")
	if apiBase == "" {
		apiBase = "https://gitlab.com/api/v4"
	}
	client = defaultHTTPClient(client)
	projectRef := url.PathEscape(project)
	requestOpts := jsonRequestOptions{
		AcceptHeader:        "application/json",
		OptionalStatusCodes: []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}

	var issue gitlabIssue
	if err := getJSON(ctx, client, fmt.Sprintf("%s/projects/%s/issues/%d", apiBase, projectRef, issueNumber), &issue, requestOpts); err != nil {
		return nil, err
	}
	var notes []gitlabIssueNote
	if err := getJSONOptional(ctx, client, fmt.Sprintf("%s/projects/%s/issues/%d/notes?per_page=100", apiBase, projectRef, issueNumber), &notes, requestOpts); err != nil {
		return nil, err
	}

	var fixtures []Fixture
	for index, snippet := range extractLogSnippets(issue.Description) {
		fixtures = append(fixtures, gitlabFixture(project, issueNumber, issue, "", index+1, snippet, now))
	}
	for _, note := range notes {
		for index, snippet := range extractLogSnippets(note.Body) {
			fixtures = append(fixtures, gitlabFixture(project, issueNumber, issue, strconv.Itoa(note.ID), index+1, snippet, now))
		}
	}
	return fixtures, nil
}

type gitlabIssue struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Author      glAuthor `json:"author"`
	Labels      []string `json:"labels"`
}

type gitlabIssueNote struct {
	ID   int      `json:"id"`
	Body string   `json:"body"`
	User glAuthor `json:"author"`
}

type glAuthor struct {
	Username string `json:"username"`
}

func parseGitLabIssueURL(rawURL string) (string, int, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", 0, fmt.Errorf("parse GitLab issue URL: %w", err)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	marker := -1
	for i := range parts {
		if parts[i] == "-" && i+2 < len(parts) && parts[i+1] == "issues" {
			marker = i
			break
		}
	}
	if marker == -1 {
		return "", 0, fmt.Errorf("unsupported GitLab issue URL %q", rawURL)
	}
	issueNumber, err := strconv.Atoi(parts[marker+2])
	if err != nil {
		return "", 0, fmt.Errorf("parse GitLab issue number from %q: %w", rawURL, err)
	}
	return strings.Join(parts[:marker], "/"), issueNumber, nil
}

func gitlabFixture(project string, issueNumber int, issue gitlabIssue, commentID string, snippetIndex int, snippet string, now time.Time) Fixture {
	return Fixture{
		ID:            buildFixtureID("gitlab", project, issueNumber, snippetIndex, snippet),
		Title:         issue.Title,
		FixtureClass:  ClassStaging,
		NormalizedLog: snippet,
		Source: SourceMetadata{
			Adapter:     "gitlab-issue",
			Provider:    "gitlab",
			URL:         fmt.Sprintf("https://gitlab.com/%s/-/issues/%d", project, issueNumber),
			Repository:  project,
			IssueNumber: issueNumber,
			CommentID:   commentID,
			Title:       issue.Title,
			Labels:      append([]string(nil), issue.Labels...),
			Author:      issue.Author.Username,
			Snippet:     snippetIndex,
			FetchedAt:   now.UTC().Format(time.RFC3339),
		},
		Review: ReviewMetadata{Status: "ingested"},
	}
}
