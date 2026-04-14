package fixtures

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"faultline/internal/engine"
)

type SourceAdapter interface {
	Name() string
	Supports(rawURL string) bool
	Fetch(ctx context.Context, rawURL string, client *http.Client, now time.Time) ([]Fixture, error)
}

func adapterByName(name string) (SourceAdapter, error) {
	adapters := []SourceAdapter{
		GitHubIssueAdapter{},
		GitLabIssueAdapter{},
		StackExchangeQuestionAdapter{},
		DiscourseTopicAdapter{},
	}
	for _, adapter := range adapters {
		if adapter.Name() == name {
			return adapter, nil
		}
	}
	return nil, fmt.Errorf("unknown adapter %q", name)
}

func defaultHTTPClient(client *http.Client) *http.Client {
	if client != nil {
		return client
	}
	return &http.Client{Timeout: 20 * time.Second}
}

var (
	fencedBlockPattern = regexp.MustCompile("(?s)```[^\n`]*\n(.*?)```")
	errorSignals       = []string{
		"error", "failed", "exception", "traceback", "panic:", "timeout", "timed out",
		"permission denied", "connection refused", "unauthorized", "denied", "no such file or directory",
		"not found", "cannot find", "module not found", "context deadline exceeded", "crashloopbackoff",
		"oomkilled", "x509", "temporary failure in name resolution", "pull access denied", "npm err",
	}
)

func extractLogSnippets(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	var snippets []string
	seen := map[string]struct{}{}
	for _, match := range fencedBlockPattern.FindAllStringSubmatch(text, -1) {
		block := engine.CanonicalizeLog(match[1])
		if !looksLikeCILog(block) {
			continue
		}
		if _, ok := seen[block]; ok {
			continue
		}
		seen[block] = struct{}{}
		snippets = append(snippets, block)
	}
	if len(snippets) > 0 {
		return snippets
	}
	for _, chunk := range strings.Split(text, "\n\n") {
		block := engine.CanonicalizeLog(chunk)
		if !looksLikeCILog(block) {
			continue
		}
		if _, ok := seen[block]; ok {
			continue
		}
		seen[block] = struct{}{}
		snippets = append(snippets, block)
	}
	return snippets
}

func looksLikeCILog(block string) bool {
	lines := strings.Split(strings.TrimSpace(block), "\n")
	hits := 0
	lower := strings.ToLower(block)
	for _, signal := range errorSignals {
		if strings.Contains(lower, signal) {
			hits++
		}
	}
	return hits > 0 || len(lines) >= 6
}

func slugify(parts ...string) string {
	joined := strings.ToLower(strings.Join(parts, "-"))
	var b strings.Builder
	lastDash := false
	for _, r := range joined {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case !lastDash:
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "fixture"
	}
	return out
}

func buildFixtureID(provider, repository string, issueNumber, snippet int, logText string) string {
	repoSlug := slugify(strings.ReplaceAll(repository, "/", "-"))
	return fmt.Sprintf("%s-%s-%d-s%d-%s", provider, repoSlug, issueNumber, snippet, FingerprintForLog(logText))
}

func mergeFixtures(existing [][]Fixture) []Fixture {
	var out []Fixture
	for _, group := range existing {
		out = append(out, group...)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}
