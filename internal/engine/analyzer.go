// Package engine orchestrates log analysis: it loads playbooks, normalises
// log lines, extracts context, delegates to the matcher for ranking, and
// persists results to the local history store.
package engine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"faultline/internal/detectors"
	"faultline/internal/detectors/logdetector"
	"faultline/internal/detectors/sourcedetector"
	enginedelta "faultline/internal/engine/delta"
	"faultline/internal/engine/hypothesis"
	"faultline/internal/metrics"
	"faultline/internal/model"
	"faultline/internal/playbooks"
	"faultline/internal/policy"
	"faultline/internal/repo"
	"faultline/internal/repo/topology"
	"faultline/internal/scoring"
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
	// PlaybookPackDirs adds external pack roots on top of the bundled starter pack.
	PlaybookPackDirs []string
	// NoHistory disables both reading and writing of the local history store.
	NoHistory bool
	// GitContextEnabled enables enrichment of analysis results with local git history.
	GitContextEnabled bool
	// GitSince limits git history ingestion to commits newer than this duration
	// string (e.g. "30d", "7d"). Defaults to "30d" when GitContext is true.
	GitSince string
	// RepoPath is the path to the local git repository root.  Defaults to ".".
	RepoPath string
	// BayesEnabled enables deterministic Bayesian-inspired reranking over matches.
	BayesEnabled bool
	// DeltaProvider enables provider-backed failure delta resolution.
	DeltaProvider string
	// GitHubRepository identifies the GitHub repository for delta resolution.
	GitHubRepository string
	// GitHubBranch identifies the branch for delta resolution.
	GitHubBranch string
	// GitHubRunID identifies the current GitHub Actions run.
	GitHubRunID int64
	// GitHubToken authenticates GitHub Actions delta API requests.
	GitHubToken string
	// MetricsHistoryFile is an optional path to a JSONL file of MetricsHistoryEntry
	// records used to compute FPC and PHI. When empty, only TSS is computed
	// from the local history store.
	MetricsHistoryFile string
}

// Engine orchestrates log analysis against loaded playbooks.
type Engine struct {
	opts            Options
	catalog         playbookCatalog
	registry        detectors.Registry
	history         historyRecorder
	repoEnricher    repoEnricher
	sourceFileStore sourceLoader
}

type repoSnapshot struct {
	root    string
	commits []repo.Commit
	signals repo.Signals
	state   *scoring.RepoState
}

// New returns a new Engine configured with opts.
func New(opts Options) *Engine {
	engine := &Engine{
		opts:            opts,
		catalog:         newCatalog(opts.PlaybookDir, opts.PlaybookPackDirs),
		registry:        detectors.NewRegistry(logdetector.Detector{}, sourcedetector.Detector{}),
		history:         defaultHistoryRecorder{},
		sourceFileStore: defaultSourceLoader{},
	}
	engine.repoEnricher = localRepoEnricher{opts: opts}
	return engine
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
	packProv := playbooks.ProvenanceFromPlaybooks(pbs)

	lines, err := readLines(r)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrNoInput
	}

	ctx := ExtractContext(lines)
	currentLog := joinOriginalLines(lines)
	logDetector, err := e.registry.MustLookup(detectors.KindLog)
	if err != nil {
		return nil, err
	}
	results := logDetector.Detect(detectors.FilterPlaybooks(pbs, detectors.KindLog), detectors.Target{
		LogLines:   lines,
		LogContext: ctx,
	})

	if len(results) == 0 {
		return &model.Analysis{
			Results:         []model.Result{},
			Context:         ctx,
			PackProvenances: packProv,
		}, ErrNoMatch
	}

	var (
		snapshot *repoSnapshot
		delta    *model.Delta
	)
	deltaState := e.loadProviderDelta(currentLog)
	if e.opts.GitContextEnabled {
		snapshot = e.loadRepoSnapshot()
	}
	repoState := mergeRepoStates(repoStateFromSnapshot(snapshot), deltaState)
	if !e.opts.BayesEnabled {
		delta = scoring.DiagnoseDelta(repoState)
	}
	if e.opts.BayesEnabled {
		reranked, scoredDelta, scoreErr := scoring.Score(scoring.Inputs{
			Context:        ctx,
			Lines:          lines,
			Results:        results,
			RepoState:      repoState,
			DeltaRequested: e.deltaRequested(),
		})
		if scoreErr != nil {
			return nil, scoreErr
		}
		results = reranked
		delta = scoredDelta
	}
	results, differential := hypothesis.Build(hypothesis.Inputs{
		Results: results,
		Lines:   lines,
		Context: ctx,
		Delta:   delta,
		Limit:   3,
	})

	// Enrich results with history seen-counts (best-effort; never blocks analysis).
	if !e.opts.NoHistory {
		for i := range results {
			results[i].SeenCount = e.history.CountSeen(results[i].Playbook.ID)
		}
	}

	fp := fingerprint(results[0])
	analysis := &model.Analysis{
		Results:         results,
		Context:         ctx,
		Fingerprint:     fp,
		Delta:           delta,
		Differential:    differential,
		PackProvenances: packProv,
	}

	// Persist the top result so future runs can report recurrence.
	if !e.opts.NoHistory {
		e.history.Record(results[0])
	}

	// Compute pipeline reliability metrics (best-effort; never blocks analysis).
	if !e.opts.NoHistory {
		analysis.Metrics = e.computeMetrics(results[0].Playbook.ID)
		analysis.Policy = policy.Compute(analysis.Metrics, results[0].Playbook.Severity)
	}

	// Enrich with git repo context when requested (best-effort; never blocks).
	if e.opts.GitContextEnabled && len(results) > 0 {
		if rc := correlateSnapshot(snapshot, results[0]); rc != nil {
			analysis.RepoContext = rc
		} else if rc := e.repoEnricher.Enrich(results[0]); rc != nil {
			analysis.RepoContext = rc
		}
	}

	return analysis, nil
}

// AnalyzeRepository scans a repository tree using source-detector playbooks.
func (e *Engine) AnalyzeRepository(root string, changeSet detectors.ChangeSet) (*model.Analysis, error) {
	pbs, err := e.loadPlaybooks()
	if err != nil {
		return nil, err
	}
	packProv := playbooks.ProvenanceFromPlaybooks(pbs)
	sourcePlaybooks := detectors.FilterPlaybooks(pbs, detectors.KindSource)
	if len(sourcePlaybooks) == 0 {
		return &model.Analysis{Results: []model.Result{}, PackProvenances: packProv}, ErrNoMatch
	}
	files, err := e.sourceFileStore.Load(root)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, ErrNoInput
	}
	sourceDetector, err := e.registry.MustLookup(detectors.KindSource)
	if err != nil {
		return nil, err
	}
	results := sourceDetector.Detect(sourcePlaybooks, detectors.Target{
		RepositoryRoot: root,
		Files:          files,
		ChangeSet:      changeSet,
	})
	if len(results) == 0 {
		return &model.Analysis{Results: []model.Result{}, PackProvenances: packProv}, ErrNoMatch
	}
	var (
		snapshot *repoSnapshot
		delta    *model.Delta
	)
	if len(changeSet.ChangedFiles) > 0 || e.opts.GitContextEnabled {
		snapshot = e.loadRepoSnapshotFromPath(root, changeSet)
	}
	repoState := repoStateFromSnapshot(snapshot)
	if !e.opts.BayesEnabled {
		delta = scoring.DiagnoseDelta(repoState)
	}
	if e.opts.BayesEnabled {
		reranked, scoredDelta, scoreErr := scoring.Score(scoring.Inputs{
			Results:        results,
			RepoState:      repoState,
			ChangeSet:      changeSet,
			DeltaRequested: e.deltaRequested() || len(changeSet.ChangedFiles) > 0,
		})
		if scoreErr != nil {
			return nil, scoreErr
		}
		results = reranked
		delta = scoredDelta
	}
	results, differential := hypothesis.Build(hypothesis.Inputs{
		Results: results,
		Delta:   delta,
		Limit:   3,
	})
	if !e.opts.NoHistory {
		for i := range results {
			results[i].SeenCount = e.history.CountSeen(results[i].Playbook.ID)
		}
		e.history.Record(results[0])
	}
	return &model.Analysis{
		Results:         results,
		Fingerprint:     fingerprint(results[0]),
		Source:          root,
		Delta:           delta,
		Differential:    differential,
		PackProvenances: packProv,
	}, nil
}

func joinOriginalLines(lines []model.Line) string {
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(line.Original)
		if !strings.HasSuffix(line.Original, "\n") {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

type localRepoEnricher struct {
	opts Options
}

// Enrich scans the local git repository and correlates the failure result with
// recent commit history. Errors are silently swallowed so that git failures
// never interrupt analysis output.
func (e localRepoEnricher) Enrich(result model.Result) *model.RepoContext {
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
	return &repoCtx
}

func (e *Engine) loadRepoSnapshot() *repoSnapshot {
	repoPath := e.opts.RepoPath
	if repoPath == "" {
		repoPath = "."
	}
	return e.loadRepoSnapshotFromPath(repoPath, detectors.ChangeSet{})
}

func (e *Engine) loadRepoSnapshotFromPath(repoPath string, changeSet detectors.ChangeSet) *repoSnapshot {
	if repoPath == "" {
		repoPath = "."
	}
	sinceStr := e.opts.GitSince
	if sinceStr == "" {
		sinceStr = "30d"
	}
	scanner, err := repo.NewScanner(repoPath)
	if err != nil {
		return &repoSnapshot{state: stateFromChangeSet(repoPath, changeSet)}
	}
	commits, err := repo.LoadHistory(scanner, sinceStr)
	if err != nil {
		return &repoSnapshot{state: stateFromChangeSet(scanner.Root, changeSet)}
	}
	sigs := repo.DeriveSignals(commits)
	state := &scoring.RepoState{
		Root:          scanner.Root,
		RecentFiles:   recentFilesFromCommits(commits, 10),
		ChangedFiles:  changedFilesFromChangeSet(changeSet),
		HotspotDirs:   hotspotDirsFromSignals(sigs, 5),
		HotfixSignals: hotfixSignalsFromSignals(sigs, 5),
		DriftSignals:  driftSignalsFromSignals(sigs, 5),
	}
	if len(state.ChangedFiles) == 0 {
		state.ChangedFiles = append([]string(nil), state.RecentFiles...)
	}
	// Load CODEOWNERS and derive topology signals (best-effort).
	state.TopologySignals = loadTopologySignals(scanner.Root, state.ChangedFiles)
	return &repoSnapshot{
		root:    scanner.Root,
		commits: commits,
		signals: sigs,
		state:   state,
	}
}

func repoStateFromSnapshot(snapshot *repoSnapshot) *scoring.RepoState {
	if snapshot == nil {
		return nil
	}
	return snapshot.state
}

func correlateSnapshot(snapshot *repoSnapshot, result model.Result) *model.RepoContext {
	if snapshot == nil || snapshot.root == "" || len(snapshot.commits) == 0 {
		return nil
	}
	repoCtx := repo.Correlate(snapshot.root, result.Playbook.Category, result.Playbook.ID, snapshot.commits, snapshot.signals)
	// Attach topology signals when they were derived.
	if snapshot.state != nil && len(snapshot.state.TopologySignals) > 0 {
		repoCtx.Topology = toModelTopologySignals(snapshot.state.TopologySignals)
	}
	return &repoCtx
}

// loadTopologySignals parses CODEOWNERS from root, builds an ownership graph,
// and derives topology signals from the given changed files. The failure files
// are approximated from the hotspot files. Errors are silently ignored so
// topology never blocks analysis.
func loadTopologySignals(root string, changedFiles []string) []string {
	rules, err := topology.ParseCODEOWNERS(root)
	if err != nil || len(rules) == 0 {
		return nil
	}
	fsys := os.DirFS(root)
	graph := topology.BuildGraph(root, rules, fsys)
	// Use hotspot dirs as proxies for failure-file locations.
	sigs := topology.DeriveSignals(graph, changedFiles, nil)
	return sigs.ActiveSignals
}

// toModelTopologySignals converts active topology signal names into the model
// representation stored on RepoContext.
func toModelTopologySignals(active []string) *model.TopologySignals {
	ts := &model.TopologySignals{
		ActiveSignals: append([]string(nil), active...),
	}
	for _, s := range active {
		switch s {
		case topology.SignalBoundaryCrossed:
			ts.BoundaryCrossed = true
		case topology.SignalUpstreamChanged:
			ts.UpstreamChanged = true
		case topology.SignalOwnershipMismatch:
			ts.OwnershipMismatch = true
		case topology.SignalFailureClustered:
			ts.FailureClustered = true
		}
	}
	return ts
}

func stateFromChangeSet(root string, changeSet detectors.ChangeSet) *scoring.RepoState {
	files := changedFilesFromChangeSet(changeSet)
	if root == "" && len(files) == 0 {
		return nil
	}
	return &scoring.RepoState{
		Root:         root,
		ChangedFiles: files,
	}
}

func mergeRepoStates(states ...*scoring.RepoState) *scoring.RepoState {
	var merged *scoring.RepoState
	for _, state := range states {
		if state == nil {
			continue
		}
		if merged == nil {
			copyState := *state
			copyState.ChangedFiles = append([]string(nil), state.ChangedFiles...)
			copyState.RecentFiles = append([]string(nil), state.RecentFiles...)
			copyState.HotspotDirs = append([]string(nil), state.HotspotDirs...)
			copyState.HotfixSignals = append([]string(nil), state.HotfixSignals...)
			copyState.DriftSignals = append([]string(nil), state.DriftSignals...)
			copyState.TopologySignals = append([]string(nil), state.TopologySignals...)
			copyState.TestsNewlyFailing = append([]string(nil), state.TestsNewlyFailing...)
			copyState.ErrorsAdded = append([]string(nil), state.ErrorsAdded...)
			copyState.EnvDiff = cloneDeltaEnvDiff(state.EnvDiff)
			merged = &copyState
			continue
		}
		if merged.Root == "" {
			merged.Root = state.Root
		}
		if merged.Provider == "" {
			merged.Provider = state.Provider
		}
		merged.ChangedFiles = dedupeSortedStrings(append(merged.ChangedFiles, state.ChangedFiles...))
		merged.RecentFiles = dedupeSortedStrings(append(merged.RecentFiles, state.RecentFiles...))
		merged.HotspotDirs = dedupeSortedStrings(append(merged.HotspotDirs, state.HotspotDirs...))
		merged.HotfixSignals = dedupeSortedStrings(append(merged.HotfixSignals, state.HotfixSignals...))
		merged.DriftSignals = dedupeSortedStrings(append(merged.DriftSignals, state.DriftSignals...))
		merged.TopologySignals = dedupeSortedStrings(append(merged.TopologySignals, state.TopologySignals...))
		merged.TestsNewlyFailing = dedupeSortedStrings(append(merged.TestsNewlyFailing, state.TestsNewlyFailing...))
		merged.ErrorsAdded = dedupeSortedStrings(append(merged.ErrorsAdded, state.ErrorsAdded...))
		for key, value := range state.EnvDiff {
			if merged.EnvDiff == nil {
				merged.EnvDiff = map[string]model.DeltaEnvChange{}
			}
			merged.EnvDiff[key] = value
		}
	}
	return merged
}

func dedupeSortedStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func cloneDeltaEnvDiff(in map[string]model.DeltaEnvChange) map[string]model.DeltaEnvChange {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]model.DeltaEnvChange, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func (e *Engine) deltaRequested() bool {
	return strings.TrimSpace(e.opts.DeltaProvider) != ""
}

func (e *Engine) loadProviderDelta(currentLog string) *scoring.RepoState {
	if !e.deltaRequested() {
		return nil
	}
	resolver := enginedelta.NewResolver(nil)
	snapshot, err := resolver.Resolve(context.Background(), enginedelta.Options{
		Provider: e.opts.DeltaProvider,
		GitHub: enginedelta.GitHubOptions{
			Repository: e.opts.GitHubRepository,
			Branch:     e.opts.GitHubBranch,
			RunID:      e.opts.GitHubRunID,
			Token:      e.opts.GitHubToken,
		},
	}, currentLog)
	if err != nil || snapshot == nil {
		return nil
	}
	return &scoring.RepoState{
		Provider:          snapshot.Provider,
		ChangedFiles:      append([]string(nil), snapshot.FilesChanged...),
		TestsNewlyFailing: append([]string(nil), snapshot.TestsNewlyFailing...),
		ErrorsAdded:       append([]string(nil), snapshot.ErrorsAdded...),
		EnvDiff:           cloneDeltaEnvDiff(snapshot.EnvDiff),
	}
}

func changedFilesFromChangeSet(changeSet detectors.ChangeSet) []string {
	if len(changeSet.ChangedFiles) == 0 {
		return nil
	}
	files := make([]string, 0, len(changeSet.ChangedFiles))
	for file := range changeSet.ChangedFiles {
		files = append(files, filepath.ToSlash(file))
	}
	sort.Strings(files)
	return files
}

func recentFilesFromCommits(commits []repo.Commit, max int) []string {
	seen := map[string]struct{}{}
	var files []string
	for _, commit := range commits {
		for _, file := range commit.Files {
			if _, ok := seen[file]; ok {
				continue
			}
			seen[file] = struct{}{}
			files = append(files, file)
			if len(files) >= max {
				return files
			}
		}
	}
	return files
}

func hotspotDirsFromSignals(sigs repo.Signals, max int) []string {
	var dirs []string
	for _, item := range sigs.HotspotDirs {
		if item.Dir == "" {
			continue
		}
		dirs = append(dirs, item.Dir)
		if len(dirs) >= max {
			break
		}
	}
	return dirs
}

func hotfixSignalsFromSignals(sigs repo.Signals, max int) []string {
	var hints []string
	for _, commit := range sigs.HotfixCommits {
		hints = append(hints, commit.Subject)
		if len(hints) >= max {
			break
		}
	}
	return hints
}

func driftSignalsFromSignals(sigs repo.Signals, max int) []string {
	var hints []string
	for _, commit := range sigs.RevertCommits {
		hints = append(hints, commit.Subject)
		if len(hints) >= max {
			return hints
		}
	}
	for _, dir := range sigs.RepeatedDirs {
		if dir.Dir == "" {
			continue
		}
		hints = append(hints, "Repeated edits in "+dir.Dir)
		if len(hints) >= max {
			break
		}
	}
	return hints
}

// List returns all playbooks available in the configured directory.
func (e *Engine) List() ([]model.Playbook, error) {
	return e.catalog.List()
}

// Explain returns the playbook identified by id, or an error if not found.
func (e *Engine) Explain(id string) (model.Playbook, error) {
	return e.catalog.Explain(id)
}

// computeMetrics builds pipeline reliability metrics for the given failure ID.
// TSS is derived from the local history store. When MetricsHistoryFile is set,
// FPC and PHI are also computed from the supplied artifact file.
// Errors are silently discarded so metrics never block analysis output.
func (e *Engine) computeMetrics(failureID string) *model.Metrics {
	rawEntries := e.history.AllEntries()
	localEntries := make([]metrics.LocalEntry, len(rawEntries))
	for i, he := range rawEntries {
		localEntries[i] = metrics.LocalEntry{FailureID: he.FailureID}
	}
	m := metrics.FromLocalHistory(failureID, localEntries)

	if e.opts.MetricsHistoryFile != "" {
		explicit, err := metrics.LoadHistoryFile(e.opts.MetricsHistoryFile)
		if err == nil {
			m = metrics.WithExplicitHistory(m, explicit)
		}
	}
	return m
}

// ReadLines reads log input into canonicalized line values used by the matcher.
func ReadLines(r io.Reader) ([]model.Line, error) {
	return readLines(r)
}

func (e *Engine) loadPlaybooks() ([]model.Playbook, error) {
	return e.catalog.Load()
}

// readLines reads all bytes from r and splits into normalised Line values.
// Blank lines and lines that become empty after trimming are discarded.
func readLines(r io.Reader) ([]model.Line, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read log input: %w", err)
	}

	parts := strings.Split(CanonicalizeLog(string(data)), "\n")
	lines := make([]model.Line, 0, len(parts))
	for n, part := range parts {
		orig := strings.TrimSpace(part)
		if orig == "" {
			continue
		}
		lines = append(lines, model.Line{
			Original:   orig,
			Normalized: NormalizeLine(orig),
			Number:     n + 1,
		})
	}
	return lines, nil
}

func loadSourceFiles(root string) ([]detectors.SourceFile, error) {
	var files []detectors.SourceFile
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !looksLikeSourceFile(path) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		content := strings.ReplaceAll(string(data), "\r\n", "\n")
		content = strings.ReplaceAll(content, "\r", "\n")
		files = append(files, detectors.SourceFile{
			Path:    filepath.ToSlash(rel),
			Content: content,
			Lines:   strings.Split(content, "\n"),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk repository: %w", err)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	return files, nil
}

func looksLikeSourceFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go", ".js", ".jsx", ".ts", ".tsx", ".py", ".java", ".rb", ".php", ".cs", ".kt", ".yaml", ".yml":
		return true
	default:
		return false
	}
}
