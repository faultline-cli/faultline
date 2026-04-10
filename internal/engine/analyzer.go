// Package engine orchestrates log analysis: it loads playbooks, normalises
// log lines, extracts context, delegates to the matcher for ranking, and
// persists results to the local history store.
package engine

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"faultline/internal/matcher"
	"faultline/internal/model"
	"faultline/internal/playbooks"
	"faultline/internal/repo"
)

var (
	// ErrNoInput is returned when the log reader contains no usable lines.
	ErrNoInput = errors.New("no log input provided; pass a file path or pipe stdin")
	// ErrNoMatch is returned when the log was analysed but no playbook matched.
	ErrNoMatch = errors.New("no known failure pattern matched")
)

// Options configures an Engine instance.
type Options struct {
	// PlaybookDir overrides the default playbook directory resolution.
	PlaybookDir string
	// NoHistory disables both reading and writing of the local history store.
	NoHistory bool
	// GitContextEnabled enables enrichment of analysis results with local git history.
	GitContextEnabled bool
	// GitSince limits git history ingestion to commits newer than this duration
	// string (e.g. "30d", "7d"). Defaults to "30d" when GitContext is true.
	GitSince string
	// RepoPath is the path to the local git repository root.  Defaults to ".".
	RepoPath string
}

// Engine orchestrates log analysis against loaded playbooks.
type Engine struct {
	opts Options
}

// New returns a new Engine configured with opts.
func New(opts Options) *Engine {
	return &Engine{opts: opts}
}

// AnalyzePath opens path and delegates to AnalyzeReader.
func (e *Engine) AnalyzePath(path string) (*model.Analysis, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()

	a, err := e.AnalyzeReader(f)
	if a != nil {
		a.Source = path
	}
	return a, err
}

// AnalyzeReader reads log content from r and returns a ranked Analysis.
//
// When no playbook matches, the analysis is still returned (with an empty
// Results slice) alongside ErrNoMatch so callers can include context in output.
func (e *Engine) AnalyzeReader(r io.Reader) (*model.Analysis, error) {
	pbs, err := e.loadPlaybooks()
	if err != nil {
		return nil, err
	}

	lines, err := readLines(r)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrNoInput
	}

	ctx := ExtractContext(lines)
	results := matcher.Rank(pbs, lines, ctx)

	if len(results) == 0 {
		return &model.Analysis{
			Results: []model.Result{},
			Context: ctx,
		}, ErrNoMatch
	}

	// Enrich results with history seen-counts (best-effort; never blocks analysis).
	if !e.opts.NoHistory {
		for i := range results {
			results[i].SeenCount = countSeen(results[i].Playbook.ID)
		}
	}

	fp := fingerprint(results[0])
	analysis := &model.Analysis{
		Results:     results,
		Context:     ctx,
		Fingerprint: fp,
	}

	// Persist the top result so future runs can report recurrence.
	if !e.opts.NoHistory {
		recordResult(results[0])
	}

	// Enrich with git repo context when requested (best-effort; never blocks).
	if e.opts.GitContextEnabled && len(results) > 0 {
		if rc := e.enrichRepoContext(results[0]); rc != nil {
			analysis.RepoContext = rc
		}
	}

	return analysis, nil
}

// enrichRepoContext scans the local git repository and correlates the failure
// result with recent commit history. Errors are silently swallowed so that git
// failures never interrupt analysis output.
func (e *Engine) enrichRepoContext(result model.Result) *model.RepoContext {
	repoPath := e.opts.RepoPath
	if repoPath == "" {
		repoPath = "."
	}
	sinceStr := e.opts.GitSince
	if sinceStr == "" {
		sinceStr = "30d"
	}

	scanner, err := repo.NewScanner(repoPath)
	if err != nil {
		return nil
	}

	commits, err := repo.LoadHistory(scanner, sinceStr)
	if err != nil || len(commits) == 0 {
		return nil
	}

	sigs := repo.DeriveSignals(commits)
	repoCtx := repo.Correlate(
		scanner.Root,
		result.Playbook.Category,
		result.Playbook.ID,
		commits,
		sigs,
	)

	// Convert repo.RepoContext → model.RepoContext
	out := &model.RepoContext{
		RepoRoot:           repoCtx.RepoRoot,
		RecentFiles:        repoCtx.RecentFiles,
		HotspotDirectories: repoCtx.HotspotDirectories,
		CoChangeHints:      repoCtx.CoChangeHints,
		HotfixSignals:      repoCtx.HotfixSignals,
		DriftSignals:       repoCtx.DriftSignals,
	}
	out.RelatedCommits = make([]model.RepoCommit, len(repoCtx.RelatedCommits))
	for i, c := range repoCtx.RelatedCommits {
		out.RelatedCommits[i] = model.RepoCommit{
			Hash:    c.Hash,
			Subject: c.Subject,
			Date:    c.Date,
		}
	}
	return out
}

// List returns all playbooks available in the configured directory.
func (e *Engine) List() ([]model.Playbook, error) {
	return e.loadPlaybooks()
}

// Explain returns the playbook identified by id, or an error if not found.
func (e *Engine) Explain(id string) (model.Playbook, error) {
	pbs, err := e.loadPlaybooks()
	if err != nil {
		return model.Playbook{}, err
	}
	for _, pb := range pbs {
		if pb.ID == id {
			return pb, nil
		}
	}
	return model.Playbook{}, fmt.Errorf("unknown playbook %q", id)
}

func (e *Engine) loadPlaybooks() ([]model.Playbook, error) {
	if e.opts.PlaybookDir != "" {
		return playbooks.LoadDir(e.opts.PlaybookDir)
	}
	return playbooks.LoadDefault()
}

// readLines reads all bytes from r and splits into normalised Line values.
// Blank lines and lines that become empty after trimming are discarded.
func readLines(r io.Reader) ([]model.Line, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read log input: %w", err)
	}

	// Normalise line endings.
	raw := strings.ReplaceAll(string(data), "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")

	parts := strings.Split(raw, "\n")
	lines := make([]model.Line, 0, len(parts))
	for n, part := range parts {
		orig := strings.TrimSpace(part)
		if orig == "" {
			continue
		}
		lines = append(lines, model.Line{
			Original:   orig,
			Normalized: normalizeLine(orig),
			Number:     n + 1,
		})
	}
	return lines, nil
}

func normalizeLine(line string) string {
	return strings.ToLower(strings.Join(strings.Fields(line), " "))
}
