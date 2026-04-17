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

type DiscourseTopicAdapter struct {
	APIBase string
}

func (a DiscourseTopicAdapter) Name() string {
	return "discourse-topic"
}

func (a DiscourseTopicAdapter) Supports(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.Contains(u.Path, "/t/")
}

func (a DiscourseTopicAdapter) Fetch(ctx context.Context, rawURL string, client *http.Client, now time.Time) ([]Fixture, error) {
	topicURL, topicHost, topicID, err := parseDiscourseTopicURL(rawURL)
	if err != nil {
		return nil, err
	}
	apiURL := topicURL + ".json"
	client = defaultHTTPClient(client)
	requestOpts := jsonRequestOptions{AcceptHeader: "application/json"}

	var topic discourseTopicResponse
	if err := getJSON(ctx, client, apiURL, &topic, requestOpts); err != nil {
		return nil, err
	}

	var fixtures []Fixture
	for _, post := range topic.PostStream.Posts {
		for index, snippet := range extractLogSnippets(htmlToText(firstNonEmpty(post.Raw, post.Cooked))) {
			fixtures = append(fixtures, Fixture{
				ID:            buildFixtureID("discourse", topicHost, topicID, index+1, snippet),
				Title:         topic.Title,
				FixtureClass:  ClassStaging,
				NormalizedLog: snippet,
				Source: SourceMetadata{
					Adapter:     "discourse-topic",
					Provider:    "discourse",
					URL:         topicURL,
					Repository:  topicHost,
					IssueNumber: topicID,
					CommentID:   strconv.Itoa(post.ID),
					Title:       topic.Title,
					Author:      post.Username,
					Snippet:     index + 1,
					FetchedAt:   now.UTC().Format(time.RFC3339),
				},
				Review: ReviewMetadata{Status: "ingested"},
			})
		}
	}
	return fixtures, nil
}

type discourseTopicResponse struct {
	Title      string              `json:"title"`
	PostStream discoursePostStream `json:"post_stream"`
	Slug       string              `json:"slug"`
}

type discoursePostStream struct {
	Posts []discoursePost `json:"posts"`
}

type discoursePost struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Raw      string `json:"raw"`
	Cooked   string `json:"cooked"`
}

func parseDiscourseTopicURL(rawURL string) (string, string, int, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", 0, fmt.Errorf("parse Discourse topic URL: %w", err)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i := 0; i < len(parts); i++ {
		if parts[i] == "t" && i+2 < len(parts) {
			id, err := strconv.Atoi(parts[i+2])
			if err != nil {
				return "", "", 0, fmt.Errorf("parse Discourse topic id from %q: %w", rawURL, err)
			}
			if strings.HasSuffix(u.Path, ".json") {
				return strings.TrimSuffix(rawURL, ".json"), u.Host, id, nil
			}
			topicURL := fmt.Sprintf("%s://%s/t/%s/%d", u.Scheme, u.Host, parts[i+1], id)
			return topicURL, u.Host, id, nil
		}
	}
	return "", "", 0, fmt.Errorf("unsupported Discourse topic URL %q", rawURL)
}
