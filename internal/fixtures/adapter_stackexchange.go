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

type StackExchangeQuestionAdapter struct {
	APIBase string
}

func (a StackExchangeQuestionAdapter) Name() string {
	return "stackexchange-question"
}

func (a StackExchangeQuestionAdapter) Supports(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(strings.TrimPrefix(u.Host, "www."))
	if host == "stackoverflow.com" || host == "superuser.com" || host == "serverfault.com" || host == "askubuntu.com" {
		return strings.Contains(u.Path, "/questions/")
	}
	return strings.HasSuffix(host, ".stackexchange.com") && strings.Contains(u.Path, "/questions/")
}

func (a StackExchangeQuestionAdapter) Fetch(ctx context.Context, rawURL string, client *http.Client, now time.Time) ([]Fixture, error) {
	site, host, questionID, err := parseStackExchangeQuestionURL(rawURL)
	if err != nil {
		return nil, err
	}
	apiBase := strings.TrimRight(a.APIBase, "/")
	if apiBase == "" {
		apiBase = "https://api.stackexchange.com/2.3"
	}
	client = defaultHTTPClient(client)
	requestOpts := jsonRequestOptions{
		AcceptHeader:       "application/json",
		OptionalStatusCodes: []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}

	var question stackExchangeQuestionResponse
	if err := getJSON(ctx, client, fmt.Sprintf("%s/questions/%d?site=%s&filter=withbody&order=desc&sort=activity", apiBase, questionID, url.QueryEscape(site)), &question, requestOpts); err != nil {
		return nil, err
	}
	var answers stackExchangeQuestionResponse
	if err := getJSONOptional(ctx, client, fmt.Sprintf("%s/questions/%d/answers?site=%s&filter=withbody&order=desc&sort=votes", apiBase, questionID, url.QueryEscape(site)), &answers, requestOpts); err != nil {
		return nil, err
	}

	var fixtures []Fixture
	if len(question.Items) > 0 {
		q := question.Items[0]
		fixtures = append(fixtures, stackExchangeFixtures(site, host, questionID, q.Title, q.Body, q.Owner.DisplayName, 0, now)...)
	}
	for _, answer := range answers.Items {
		fixtures = append(fixtures, stackExchangeFixtures(site, host, questionID, questionTitle(question), answer.Body, answer.Owner.DisplayName, answer.AnswerID, now)...)
	}
	return fixtures, nil
}

type stackExchangeQuestionResponse struct {
	Items []stackExchangeQuestion `json:"items"`
}

type stackExchangeQuestion struct {
	QuestionID int               `json:"question_id"`
	AnswerID   int               `json:"answer_id"`
	Title      string            `json:"title"`
	Body       string            `json:"body"`
	Owner      stackExchangeUser `json:"owner"`
}

type stackExchangeUser struct {
	DisplayName string `json:"display_name"`
}

func parseStackExchangeQuestionURL(rawURL string) (string, string, int, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", 0, fmt.Errorf("parse Stack Exchange question URL: %w", err)
	}
	host := strings.ToLower(strings.TrimPrefix(u.Host, "www."))
	site := siteFromStackExchangeHost(host)
	if site == "" {
		return "", "", 0, fmt.Errorf("unsupported Stack Exchange host %q", rawURL)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i := 0; i < len(parts); i++ {
		if parts[i] == "questions" && i+1 < len(parts) {
			id, err := strconv.Atoi(parts[i+1])
			if err != nil {
				return "", "", 0, fmt.Errorf("parse Stack Exchange question id from %q: %w", rawURL, err)
			}
			return site, host, id, nil
		}
	}
	return "", "", 0, fmt.Errorf("unsupported Stack Exchange question URL %q", rawURL)
}

func siteFromStackExchangeHost(host string) string {
	switch host {
	case "stackoverflow.com":
		return "stackoverflow"
	case "superuser.com":
		return "superuser"
	case "serverfault.com":
		return "serverfault"
	case "askubuntu.com":
		return "askubuntu"
	}
	if strings.HasSuffix(host, ".stackexchange.com") {
		return strings.TrimSuffix(host, ".com")
	}
	return ""
}

func questionTitle(response stackExchangeQuestionResponse) string {
	if len(response.Items) == 0 {
		return ""
	}
	return response.Items[0].Title
}

func stackExchangeFixtures(site, host string, questionID int, title, body, author string, postID int, now time.Time) []Fixture {
	var fixtures []Fixture
	for index, snippet := range extractLogSnippets(htmlToText(body)) {
		fixtures = append(fixtures, Fixture{
			ID:            buildFixtureID("stackexchange", site, questionID, index+1, snippet),
			Title:         title,
			FixtureClass:  ClassStaging,
			NormalizedLog: snippet,
			Source: SourceMetadata{
				Adapter:     "stackexchange-question",
				Provider:    "stackexchange",
				URL:         fmt.Sprintf("https://%s/questions/%d", host, questionID),
				Repository:  site,
				IssueNumber: questionID,
				CommentID:   strconv.Itoa(postID),
				Title:       title,
				Author:      author,
				Snippet:     index + 1,
				FetchedAt:   now.UTC().Format(time.RFC3339),
			},
			Review: ReviewMetadata{Status: "ingested"},
		})
	}
	return fixtures
}
