package delta

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	"faultline/internal/model"
)

type gitlabJob struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Stage    string `json:"stage"`
	Status   string `json:"status"`
	Pipeline struct {
		ID     int64  `json:"id"`
		SHA    string `json:"sha"`
		Ref    string `json:"ref"`
		Source string `json:"source"`
	} `json:"pipeline"`
}

type gitlabPipeline struct {
	ID     int64  `json:"id"`
	SHA    string `json:"sha"`
	Ref    string `json:"ref"`
	Source string `json:"source"`
	Status string `json:"status"`
}

type gitlabCompare struct {
	Diffs []struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	} `json:"diffs"`
}

func (r Resolver) resolveGitLabCI(ctx context.Context, opts GitLabOptions, currentLog string) (*Snapshot, error) {
	token := strings.TrimSpace(opts.Token)
	if token == "" {
		return nil, nil
	}
	project := strings.TrimSpace(opts.Project)
	if project == "" || opts.PipelineID == 0 || opts.JobID == 0 {
		return nil, nil
	}
	currentJob, err := r.gitlabJob(ctx, opts, opts.JobID)
	if err != nil {
		return nil, err
	}
	if currentJob.ID == 0 || currentJob.Pipeline.ID == 0 || strings.TrimSpace(currentJob.Name) == "" {
		return nil, nil
	}
	branch := strings.TrimSpace(opts.Branch)
	if branch == "" {
		branch = strings.TrimSpace(currentJob.Pipeline.Ref)
	}
	if branch == "" || strings.TrimSpace(currentJob.Pipeline.SHA) == "" {
		return nil, nil
	}
	baseline, err := r.lastSuccessfulGitLabPipeline(ctx, opts, branch, currentJob.Pipeline.SHA, currentJob.Pipeline.ID)
	if err != nil {
		return nil, err
	}
	if baseline == nil || baseline.ID == 0 || strings.TrimSpace(baseline.SHA) == "" {
		return nil, nil
	}
	baselineJob, err := r.matchingGitLabJob(ctx, opts, baseline.ID, currentJob)
	if err != nil {
		return nil, err
	}
	if baselineJob == nil {
		return nil, nil
	}
	baselineLog, err := r.gitlabJobTrace(ctx, opts, baselineJob.ID)
	if err != nil {
		return nil, err
	}
	changedFiles, err := r.gitlabChangedFiles(ctx, opts, baseline.SHA, currentJob.Pipeline.SHA)
	if err != nil {
		return nil, err
	}
	envDiff := map[string]model.DeltaEnvChange{}
	addEnvDiff(envDiff, "branch", baseline.Ref, branch)
	addEnvDiff(envDiff, "source", baseline.Source, currentJob.Pipeline.Source)
	addEnvDiff(envDiff, "head_sha", baseline.SHA, currentJob.Pipeline.SHA)
	addEnvDiff(envDiff, "pipeline_id", strconv.FormatInt(baseline.ID, 10), strconv.FormatInt(currentJob.Pipeline.ID, 10))
	addEnvDiff(envDiff, "job_name", baselineJob.Name, currentJob.Name)
	addEnvDiff(envDiff, "stage", baselineJob.Stage, currentJob.Stage)
	snapshot := buildSnapshot("gitlab-ci", currentLog, baselineLog, changedFiles, envDiff)
	return &snapshot, nil
}

func (r Resolver) gitlabJob(ctx context.Context, opts GitLabOptions, jobID int64) (gitlabJob, error) {
	var out gitlabJob
	err := r.gitlabGetJSON(ctx, opts, fmt.Sprintf("projects/%s/jobs/%d", url.PathEscape(strings.TrimSpace(opts.Project)), jobID), nil, &out)
	return out, err
}

func (r Resolver) lastSuccessfulGitLabPipeline(ctx context.Context, opts GitLabOptions, branch, currentSHA string, currentPipelineID int64) (*gitlabPipeline, error) {
	var pipelines []gitlabPipeline
	err := r.gitlabGetJSON(ctx, opts, fmt.Sprintf("projects/%s/pipelines", url.PathEscape(strings.TrimSpace(opts.Project))), map[string]string{
		"ref":      branch,
		"status":   "success",
		"per_page": "20",
	}, &pipelines)
	if err != nil {
		return nil, err
	}
	for _, pipeline := range pipelines {
		if pipeline.ID == 0 || pipeline.ID == currentPipelineID {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(pipeline.SHA), strings.TrimSpace(currentSHA)) {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(pipeline.Status), "success") {
			continue
		}
		pipelineCopy := pipeline
		return &pipelineCopy, nil
	}
	return nil, nil
}

func (r Resolver) matchingGitLabJob(ctx context.Context, opts GitLabOptions, pipelineID int64, current gitlabJob) (*gitlabJob, error) {
	var jobs []gitlabJob
	err := r.gitlabGetJSON(ctx, opts, fmt.Sprintf("projects/%s/pipelines/%d/jobs", url.PathEscape(strings.TrimSpace(opts.Project)), pipelineID), map[string]string{
		"per_page": "100",
	}, &jobs)
	if err != nil {
		return nil, err
	}
	sort.Slice(jobs, func(i, j int) bool {
		if jobs[i].ID != jobs[j].ID {
			return jobs[i].ID > jobs[j].ID
		}
		return jobs[i].Name < jobs[j].Name
	})
	for _, job := range jobs {
		if !strings.EqualFold(strings.TrimSpace(job.Status), "success") {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(job.Name), strings.TrimSpace(current.Name)) {
			jobCopy := job
			return &jobCopy, nil
		}
	}
	for _, job := range jobs {
		if !strings.EqualFold(strings.TrimSpace(job.Status), "success") {
			continue
		}
		if strings.TrimSpace(current.Stage) != "" && strings.EqualFold(strings.TrimSpace(job.Stage), strings.TrimSpace(current.Stage)) {
			jobCopy := job
			return &jobCopy, nil
		}
	}
	return nil, nil
}

func (r Resolver) gitlabJobTrace(ctx context.Context, opts GitLabOptions, jobID int64) (string, error) {
	req, err := r.gitlabNewRequest(ctx, opts, http.MethodGet, fmt.Sprintf("projects/%s/jobs/%d/trace", url.PathEscape(strings.TrimSpace(opts.Project)), jobID), nil)
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
		return "", fmt.Errorf("gitlab ci trace: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(body), "\r\n", "\n"), nil
}

func (r Resolver) gitlabChangedFiles(ctx context.Context, opts GitLabOptions, baseSHA, headSHA string) ([]string, error) {
	if strings.TrimSpace(baseSHA) == "" || strings.TrimSpace(headSHA) == "" || strings.EqualFold(baseSHA, headSHA) {
		return nil, nil
	}
	var compare gitlabCompare
	err := r.gitlabGetJSON(ctx, opts, fmt.Sprintf("projects/%s/repository/compare", url.PathEscape(strings.TrimSpace(opts.Project))), map[string]string{
		"from": baseSHA,
		"to":   headSHA,
	}, &compare)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(compare.Diffs))
	for _, diff := range compare.Diffs {
		file := strings.TrimSpace(diff.NewPath)
		if file == "" {
			file = strings.TrimSpace(diff.OldPath)
		}
		files = append(files, file)
	}
	return dedupeStrings(files), nil
}

func (r Resolver) gitlabGetJSON(ctx context.Context, opts GitLabOptions, endpoint string, query map[string]string, target any) error {
	req, err := r.gitlabNewRequest(ctx, opts, http.MethodGet, endpoint, query)
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
		return fmt.Errorf("gitlab ci api: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func (r Resolver) gitlabNewRequest(ctx context.Context, opts GitLabOptions, method string, endpoint string, query map[string]string) (*http.Request, error) {
	baseURL := strings.TrimSpace(opts.APIBaseURL)
	if baseURL == "" {
		baseURL = "https://gitlab.com/api/v4"
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	parsed.Path = path.Join(parsed.Path, strings.TrimPrefix(endpoint, "/"))
	values := parsed.Query()
	for key, value := range query {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		values.Set(key, value)
	}
	parsed.RawQuery = values.Encode()
	req, err := http.NewRequestWithContext(ctx, method, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("PRIVATE-TOKEN", strings.TrimSpace(opts.Token))
	req.Header.Set("JOB-TOKEN", strings.TrimSpace(opts.Token))
	return req, nil
}
