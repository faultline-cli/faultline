package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type RedditPostAdapter struct {
	APIBase string
}

func (a RedditPostAdapter) Name() string {
	return "reddit-post"
}

func (a RedditPostAdapter) Supports(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(strings.TrimPrefix(u.Host, "www."))
	switch host {
	case "reddit.com", "old.reddit.com", "np.reddit.com":
		return strings.Contains(u.Path, "/comments/")
	default:
		return false
	}
}

func (a RedditPostAdapter) Fetch(ctx context.Context, rawURL string, client *http.Client, now time.Time) ([]Fixture, error) {
	subreddit, postID, postSlug, err := parseRedditPostURL(rawURL)
	if err != nil {
		return nil, err
	}
	apiBase := strings.TrimRight(a.APIBase, "/")
	if apiBase == "" {
		apiBase = "https://www.reddit.com"
	}
	client = defaultHTTPClient(client)
	requestOpts := jsonRequestOptions{AcceptHeader: "application/json"}

	var listing []redditListing
	postURL := fmt.Sprintf("%s/r/%s/comments/%s/%s.json?raw_json=1", apiBase, subreddit, postID, postSlug)
	if err := getJSON(ctx, client, postURL, &listing, requestOpts); err != nil {
		return nil, err
	}
	if len(listing) == 0 || len(listing[0].Data.Children) == 0 {
		return nil, fmt.Errorf("reddit post %q did not return a submission", rawURL)
	}

	submission := listing[0].Data.Children[0].Data
	postTitle := submission.Title
	repository := "r/" + subreddit

	var fixtures []Fixture
	for index, snippet := range extractLogSnippets(submission.Selftext) {
		fixtures = append(fixtures, redditFixture(repository, postID, subreddit, postTitle, submission.Author, "0", index+1, snippet, now))
	}
	if len(listing) > 1 {
		comments := collectRedditComments(listing[1].Data.Children)
		for _, comment := range comments {
			for index, snippet := range extractLogSnippets(htmlToText(comment.Body)) {
				fixtures = append(fixtures, redditFixture(repository, postID, subreddit, postTitle, comment.Author, comment.ID, index+1, snippet, now))
			}
		}
	}
	return fixtures, nil
}

type redditListing struct {
	Data redditListingData `json:"data"`
}

type redditListingData struct {
	Children []redditThing `json:"children"`
}

type redditThing struct {
	Kind string      `json:"kind"`
	Data redditEntry `json:"data"`
}

type redditEntry struct {
	Title     string          `json:"title"`
	Selftext  string          `json:"selftext"`
	Body      string          `json:"body"`
	Author    string          `json:"author"`
	ID        string          `json:"id"`
	Permalink string          `json:"permalink"`
	Replies   json.RawMessage `json:"replies"`
}

type redditReplyListing struct {
	Data redditListingData `json:"data"`
}

func parseRedditPostURL(rawURL string) (string, string, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", "", fmt.Errorf("parse Reddit post URL: %w", err)
	}
	host := strings.ToLower(strings.TrimPrefix(u.Host, "www."))
	if host != "reddit.com" && host != "old.reddit.com" && host != "np.reddit.com" {
		return "", "", "", fmt.Errorf("unsupported Reddit host %q", rawURL)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i := 0; i < len(parts); i++ {
		if parts[i] == "comments" && i+1 < len(parts) {
			subreddit := parts[1]
			postID := parts[i+1]
			topicSlug := postID
			if i+2 < len(parts) {
				topicSlug = parts[i+2]
			}
			if subreddit == "" || postID == "" {
				return "", "", "", fmt.Errorf("unsupported Reddit post URL %q", rawURL)
			}
			return subreddit, postID, topicSlug, nil
		}
	}
	return "", "", "", fmt.Errorf("unsupported Reddit post URL %q", rawURL)
}

func collectRedditComments(children []redditThing) []redditEntry {
	var out []redditEntry
	for _, child := range children {
		if child.Kind != "t1" {
			continue
		}
		out = append(out, child.Data)
		replies := strings.TrimSpace(string(child.Data.Replies))
		if replies == "" || replies == `""` {
			continue
		}
		var replyListing redditReplyListing
		if err := json.Unmarshal(child.Data.Replies, &replyListing); err == nil && len(replyListing.Data.Children) > 0 {
			out = append(out, collectRedditComments(replyListing.Data.Children)...)
		}
	}
	return out
}

func redditFixture(repository, postID, subreddit, title, author, commentID string, snippetIndex int, snippet string, now time.Time) Fixture {
	idRepository := repository + "/" + postID + "/" + commentID
	return Fixture{
		ID:            buildFixtureID("reddit", idRepository, 0, snippetIndex, snippet),
		Title:         title,
		FixtureClass:  ClassStaging,
		NormalizedLog: snippet,
		Source: SourceMetadata{
			Adapter:     "reddit-post",
			Provider:    "reddit",
			URL:         fmt.Sprintf("https://www.reddit.com/r/%s/comments/%s/", subreddit, postID),
			Repository:  repository,
			IssueNumber: 0,
			CommentID:   commentID,
			Title:       title,
			Author:      author,
			Snippet:     snippetIndex,
			FetchedAt:   now.UTC().Format(time.RFC3339),
		},
		Review: ReviewMetadata{Status: "ingested"},
	}
}
