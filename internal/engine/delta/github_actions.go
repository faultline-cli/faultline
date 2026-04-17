package delta

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"faultline/internal/model"
)

type Resolver struct {
	client *http.Client
}

func NewResolver(client *http.Client) Resolver {
	if client == nil {
		client = http.DefaultClient
	}
	return Resolver{client: client}
}

func (r Resolver) Resolve(ctx context.Context, opts Options, currentLog string) (*Snapshot, error) {
	switch strings.TrimSpace(strings.ToLower(opts.Provider)) {
	case "", "none":
		return nil, nil
	case "github-actions":
		return r.resolveGitHubActions(ctx, opts.GitHub, currentLog)
	default:
		return nil, fmt.Errorf("unsupported delta provider %q", opts.Provider)
	}
}

type githubRun struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	DisplayTitle string `json:"display_title"`
	Event        string `json:"event"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	HeadBranch   string `json:"head_branch"`
	HeadSHA      string `json:"head_sha"`
	Path         string `json:"path"`
	RunAttempt   int    `json:"run_attempt"`
	WorkflowID   int64  `json:"workflow_id"`
}

type githubWorkflowRuns struct {
	WorkflowRuns []githubRun `json:"workflow_runs"`
}

type githubCompare struct {
	Files []struct {
		Filename string `json:"filename"`
	} `json:"files"`
}

func (r Resolver) resolveGitHubActions(ctx context.Context, opts GitHubOptions, currentLog string) (*Snapshot, error) {
	token := strings.TrimSpace(opts.Token)
	if token == "" {
		return nil, nil
	}
	repo := strings.TrimSpace(opts.Repository)
	if repo == "" || opts.RunID == 0 {
		return nil, nil
	}
	run, err := r.githubRun(ctx, opts, opts.RunID)
	if err != nil {
		return nil, err
	}
	branch := strings.TrimSpace(opts.Branch)
	if branch == "" {
		branch = strings.TrimSpace(run.HeadBranch)
	}
	if branch == "" || run.WorkflowID == 0 {
		return nil, nil
	}
	baseline, err := r.lastSuccessfulRun(ctx, opts, branch, run)
	if err != nil {
		return nil, err
	}
	if baseline == nil {
		return nil, nil
	}
	baselineLog, err := r.runLog(ctx, opts, baseline.ID)
	if err != nil {
		return nil, err
	}
	changedFiles, err := r.changedFiles(ctx, opts, baseline.HeadSHA, run.HeadSHA)
	if err != nil {
		return nil, err
	}
	envDiff := map[string]model.DeltaEnvChange{}
	addEnvDiff(envDiff, "branch", baseline.HeadBranch, run.HeadBranch)
	addEnvDiff(envDiff, "event", baseline.Event, run.Event)
	addEnvDiff(envDiff, "head_sha", baseline.HeadSHA, run.HeadSHA)
	addEnvDiff(envDiff, "workflow", workflowRef(*baseline), workflowRef(run))
	addEnvDiff(envDiff, "run_attempt", strconv.Itoa(baseline.RunAttempt), strconv.Itoa(run.RunAttempt))
	snapshot := buildSnapshot("github-actions", currentLog, baselineLog, changedFiles, envDiff)
	return &snapshot, nil
}

func workflowRef(run githubRun) string {
	if strings.TrimSpace(run.Path) != "" {
		return run.Path
	}
	if strings.TrimSpace(run.Name) != "" {
		return run.Name
	}
	return strconv.FormatInt(run.WorkflowID, 10)
}

func addEnvDiff(out map[string]model.DeltaEnvChange, key, baseline, current string) {
	baseline = strings.TrimSpace(baseline)
	current = strings.TrimSpace(current)
	if baseline == current || (baseline == "" && current == "") {
		return
	}
	out[key] = model.DeltaEnvChange{Baseline: baseline, Current: current}
}

func (r Resolver) githubRun(ctx context.Context, opts GitHubOptions, runID int64) (githubRun, error) {
	var out githubRun
	err := r.getJSON(ctx, opts, []string{"repos", opts.Repository, "actions", "runs", strconv.FormatInt(runID, 10)}, nil, &out)
	return out, err
}

func (r Resolver) lastSuccessfulRun(ctx context.Context, opts GitHubOptions, branch string, current githubRun) (*githubRun, error) {
	query := map[string]string{
		"branch":   branch,
		"status":   "success",
		"per_page": "20",
	}
	var runs githubWorkflowRuns
	err := r.getJSON(ctx, opts, []string{"repos", opts.Repository, "actions", "workflows", strconv.FormatInt(current.WorkflowID, 10), "runs"}, query, &runs)
	if err != nil {
		return nil, err
	}
	for _, run := range runs.WorkflowRuns {
		if run.ID == current.ID {
			continue
		}
		if strings.EqualFold(run.HeadSHA, current.HeadSHA) {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(run.Conclusion), "success") {
			continue
		}
		runCopy := run
		return &runCopy, nil
	}
	return nil, nil
}

func (r Resolver) runLog(ctx context.Context, opts GitHubOptions, runID int64) (string, error) {
	req, err := r.newRequest(ctx, opts, http.MethodGet, []string{"repos", opts.Repository, "actions", "runs", strconv.FormatInt(runID, 10), "logs"}, nil)
	if err != nil {
		return "", err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("github actions logs: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return unzipLogs(body)
}

func (r Resolver) changedFiles(ctx context.Context, opts GitHubOptions, baseSHA, headSHA string) ([]string, error) {
	if strings.TrimSpace(baseSHA) == "" || strings.TrimSpace(headSHA) == "" || strings.EqualFold(baseSHA, headSHA) {
		return nil, nil
	}
	var compare githubCompare
	err := r.getJSON(ctx, opts, []string{"repos", opts.Repository, "compare", baseSHA + "..." + headSHA}, nil, &compare)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(compare.Files))
	for _, item := range compare.Files {
		files = append(files, item.Filename)
	}
	return dedupeStrings(files), nil
}

func (r Resolver) getJSON(ctx context.Context, opts GitHubOptions, segments []string, query map[string]string, target any) error {
	req, err := r.newRequest(ctx, opts, http.MethodGet, segments, query)
	if err != nil {
		return err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("github actions api: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func (r Resolver) newRequest(ctx context.Context, opts GitHubOptions, method string, segments []string, query map[string]string) (*http.Request, error) {
	baseURL := strings.TrimSpace(opts.APIBaseURL)
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	allSegments := make([]string, 0, len(segments))
	for _, segment := range segments {
		for _, part := range strings.Split(segment, "/") {
			part = strings.TrimSpace(part)
			if part != "" {
				allSegments = append(allSegments, part)
			}
		}
	}
	parsed.Path = path.Join(parsed.Path, path.Join(allSegments...))
	values := parsed.Query()
	for key, value := range query {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		values.Set(key, value)
	}
	parsed.RawQuery = values.Encode()
	req, err := http.NewRequestWithContext(ctx, method, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(opts.Token))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return req, nil
}
