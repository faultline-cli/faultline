package hypothesis

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"faultline/internal/model"
)

const (
	Version                    = "hypothesis.v1"
	defaultLimit               = 3
	defaultSupportWeight       = 0.4
	defaultContradictionWeight = -0.35
	defaultDiscriminatorWeight = 0.25
)

type Inputs struct {
	Results []model.Result
	Lines   []model.Line
	Context model.Context
	Delta   *model.Delta
	Limit   int
}

type evaluation struct {
	matched     bool
	description string
	evidence    []string
}

type signalDef struct {
	description string
	evaluate    func(*environment) evaluation
}

type environment struct {
	lines        []model.Line
	context      model.Context
	deltaSignals map[string]model.DeltaSignal
	cache        map[string]evaluation
}

func Build(inputs Inputs) ([]model.Result, *model.DifferentialDiagnosis) {
	if len(inputs.Results) == 0 {
		return nil, nil
	}

	env := newEnvironment(inputs.Lines, inputs.Context, inputs.Delta)
	results := append([]model.Result(nil), inputs.Results...)
	hasReasoning := false

	for i := range results {
		assessment := assess(results[i], env)
		results[i].Hypothesis = assessment
		if assessment != nil && (len(assessment.Supports) > 0 || len(assessment.Contradicts) > 0 || len(assessment.Discriminators) > 0 || len(assessment.Excludes) > 0) {
			hasReasoning = true
		}
	}

	sort.Slice(results, func(i, j int) bool {
		left := hypothesisScore(results[i].Hypothesis, results[i].Score)
		right := hypothesisScore(results[j].Hypothesis, results[j].Score)
		leftEliminated := hypothesisEliminated(results[i].Hypothesis)
		rightEliminated := hypothesisEliminated(results[j].Hypothesis)
		if leftEliminated != rightEliminated {
			return !leftEliminated
		}
		if left != right {
			return left > right
		}
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		if results[i].Confidence != results[j].Confidence {
			return results[i].Confidence > results[j].Confidence
		}
		return results[i].Playbook.ID < results[j].Playbook.ID
	})

	if !hasReasoning {
		return results, nil
	}
	return results, buildDifferential(results, inputs.Limit)
}

func ValidSignal(signal string) bool {
	signal = strings.TrimSpace(signal)
	if signal == "" {
		return false
	}
	if _, ok := signalRegistry[signal]; ok {
		return true
	}
	switch {
	case strings.HasPrefix(signal, "log.contains:"):
		return strings.TrimSpace(strings.TrimPrefix(signal, "log.contains:")) != ""
	case strings.HasPrefix(signal, "log.absent:"):
		return strings.TrimSpace(strings.TrimPrefix(signal, "log.absent:")) != ""
	case strings.HasPrefix(signal, "delta.signal:"):
		return strings.TrimSpace(strings.TrimPrefix(signal, "delta.signal:")) != ""
	case strings.HasPrefix(signal, "delta.absent:"):
		return strings.TrimSpace(strings.TrimPrefix(signal, "delta.absent:")) != ""
	case strings.HasPrefix(signal, "context.stage:"):
		return strings.TrimSpace(strings.TrimPrefix(signal, "context.stage:")) != ""
	case strings.HasPrefix(signal, "context.stage.absent:"):
		return strings.TrimSpace(strings.TrimPrefix(signal, "context.stage.absent:")) != ""
	default:
		return false
	}
}

func DescribeSignal(signal string) string {
	signal = strings.TrimSpace(signal)
	if def, ok := signalRegistry[signal]; ok {
		return def.description
	}
	switch {
	case strings.HasPrefix(signal, "log.contains:"):
		return fmt.Sprintf("log contains %q", strings.TrimSpace(strings.TrimPrefix(signal, "log.contains:")))
	case strings.HasPrefix(signal, "log.absent:"):
		return fmt.Sprintf("log does not contain %q", strings.TrimSpace(strings.TrimPrefix(signal, "log.absent:")))
	case strings.HasPrefix(signal, "delta.signal:"):
		return fmt.Sprintf("delta signal %q is present", strings.TrimSpace(strings.TrimPrefix(signal, "delta.signal:")))
	case strings.HasPrefix(signal, "delta.absent:"):
		return fmt.Sprintf("delta signal %q is absent", strings.TrimSpace(strings.TrimPrefix(signal, "delta.absent:")))
	case strings.HasPrefix(signal, "context.stage:"):
		return fmt.Sprintf("analysis stage is %q", strings.TrimSpace(strings.TrimPrefix(signal, "context.stage:")))
	case strings.HasPrefix(signal, "context.stage.absent:"):
		return fmt.Sprintf("analysis stage is not %q", strings.TrimSpace(strings.TrimPrefix(signal, "context.stage.absent:")))
	default:
		return signal
	}
}

func newEnvironment(lines []model.Line, ctx model.Context, delta *model.Delta) *environment {
	deltaSignals := make(map[string]model.DeltaSignal)
	if delta != nil {
		for _, signal := range delta.Signals {
			deltaSignals[signal.ID] = signal
		}
	}
	return &environment{
		lines:        append([]model.Line(nil), lines...),
		context:      ctx,
		deltaSignals: deltaSignals,
		cache:        make(map[string]evaluation),
	}
}

func assess(result model.Result, env *environment) *model.HypothesisAssessment {
	spec := result.Playbook.Hypothesis
	assessment := &model.HypothesisAssessment{
		BaseScore:  round(result.Score),
		FinalScore: round(result.Score),
	}

	for _, item := range spec.Supports {
		match := env.evaluate(item.Signal)
		if !match.matched {
			continue
		}
		weight := item.Weight
		if weight == 0 {
			weight = defaultSupportWeight
		}
		assessment.FinalScore = round(assessment.FinalScore + weight)
		assessment.Supports = append(assessment.Supports, model.HypothesisMatch{
			Signal:      item.Signal,
			Description: match.description,
			Weight:      round(weight),
			Evidence:    append([]string(nil), match.evidence...),
		})
	}

	for _, item := range spec.Contradicts {
		match := env.evaluate(item.Signal)
		if !match.matched {
			continue
		}
		weight := item.Weight
		if weight == 0 {
			weight = defaultContradictionWeight
		}
		if weight > 0 {
			weight = -weight
		}
		assessment.FinalScore = round(assessment.FinalScore + weight)
		assessment.Contradicts = append(assessment.Contradicts, model.HypothesisMatch{
			Signal:      item.Signal,
			Description: match.description,
			Weight:      round(weight),
			Evidence:    append([]string(nil), match.evidence...),
		})
	}

	for _, item := range spec.Discriminators {
		match := env.evaluate(item.Signal)
		if !match.matched {
			continue
		}
		weight := item.Weight
		if weight == 0 {
			weight = defaultDiscriminatorWeight
		}
		assessment.FinalScore = round(assessment.FinalScore + weight)
		description := strings.TrimSpace(item.Description)
		if description == "" {
			description = match.description
		}
		assessment.Discriminators = append(assessment.Discriminators, model.HypothesisMatch{
			Signal:      item.Signal,
			Description: description,
			Weight:      round(weight),
			Evidence:    append([]string(nil), match.evidence...),
		})
	}

	for _, item := range spec.Excludes {
		match := env.evaluate(item.Signal)
		if !match.matched {
			continue
		}
		assessment.Eliminated = true
		assessment.Excludes = append(assessment.Excludes, model.HypothesisMatch{
			Signal:      item.Signal,
			Description: match.description,
			Evidence:    append([]string(nil), match.evidence...),
		})
	}

	assessment.Why = summarizeMatches(assessment.Supports, assessment.Discriminators, 3)
	assessment.WhyLessLikely = summarizeMatches(assessment.Contradicts, nil, 2)
	assessment.RuledOutBy = summarizeMatches(assessment.Excludes, nil, 2)
	assessment.DisproofChecks = disproofChecks(spec, 3)
	return assessment
}

func buildDifferential(results []model.Result, limit int) *model.DifferentialDiagnosis {
	if len(results) == 0 {
		return nil
	}
	if limit <= 0 {
		limit = defaultLimit
	}

	diff := &model.DifferentialDiagnosis{Version: Version}
	likelyIndex := -1
	for i, result := range results {
		if !hypothesisEliminated(result.Hypothesis) {
			likelyIndex = i
			summary := summarizeResult(result)
			diff.Likely = &summary
			break
		}
	}
	if likelyIndex == -1 {
		summary := summarizeResult(results[0])
		diff.Likely = &summary
		likelyIndex = 0
	}

	seen := map[string]struct{}{
		rivalryKey(results[likelyIndex]): {},
	}
	for i, result := range results {
		if i == likelyIndex {
			continue
		}
		key := rivalryKey(result)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		summary := summarizeResult(result)
		if hypothesisEliminated(result.Hypothesis) {
			diff.RuledOut = append(diff.RuledOut, summary)
			if len(diff.RuledOut) >= limit {
				continue
			}
			continue
		}
		if len(diff.Alternatives) < limit-1 {
			if len(summary.WhyLessLikely) == 0 && diff.Likely != nil {
				summary.WhyLessLikely = compareAgainstLikely(*diff.Likely, summary)
			}
			diff.Alternatives = append(diff.Alternatives, summary)
		}
	}

	if diff.Likely == nil && len(diff.Alternatives) == 0 && len(diff.RuledOut) == 0 {
		return nil
	}
	return diff
}

func summarizeResult(result model.Result) model.DifferentialCandidate {
	assessment := result.Hypothesis
	summary := model.DifferentialCandidate{
		FailureID:       result.Playbook.ID,
		Title:           result.Playbook.Title,
		Category:        result.Playbook.Category,
		Confidence:      round(result.Confidence),
		ConfidenceText:  confidenceText(result.Confidence),
		HypothesisScore: round(hypothesisScore(assessment, result.Score)),
	}
	if assessment == nil {
		return summary
	}
	summary.Why = append([]string(nil), assessment.Why...)
	summary.WhyLessLikely = append([]string(nil), assessment.WhyLessLikely...)
	summary.RuledOutBy = append([]string(nil), assessment.RuledOutBy...)
	summary.DisproofChecks = append([]string(nil), assessment.DisproofChecks...)
	return summary
}

func compareAgainstLikely(likely, alternative model.DifferentialCandidate) []string {
	if alternative.HypothesisScore == 0 || likely.HypothesisScore <= alternative.HypothesisScore {
		return nil
	}
	if len(likely.Why) == 0 {
		return []string{fmt.Sprintf("it trailed the likely cause by %.2f points", round(likely.HypothesisScore-alternative.HypothesisScore))}
	}
	return []string{
		fmt.Sprintf("it lacked the stronger discriminator: %s", likely.Why[0]),
	}
}

func summarizeMatches(primary, secondary []model.HypothesisMatch, limit int) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, limit)
	appendMatches := func(items []model.HypothesisMatch) {
		for _, item := range items {
			description := strings.TrimSpace(item.Description)
			if description == "" {
				continue
			}
			if _, ok := seen[description]; ok {
				continue
			}
			seen[description] = struct{}{}
			out = append(out, description)
			if len(out) >= limit {
				return
			}
		}
	}
	appendMatches(primary)
	if len(out) < limit {
		appendMatches(secondary)
	}
	return out
}

func disproofChecks(spec model.HypothesisSpec, limit int) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, limit)
	appendDescription := func(signal string) {
		description := DescribeSignal(signal)
		if description == "" {
			return
		}
		if _, ok := seen[description]; ok {
			return
		}
		seen[description] = struct{}{}
		out = append(out, description)
	}
	for _, item := range spec.Excludes {
		appendDescription(item.Signal)
		if len(out) >= limit {
			return out
		}
	}
	for _, item := range spec.Contradicts {
		appendDescription(item.Signal)
		if len(out) >= limit {
			return out
		}
	}
	return out
}

func (e *environment) evaluate(signal string) evaluation {
	signal = strings.TrimSpace(signal)
	if signal == "" {
		return evaluation{}
	}
	if cached, ok := e.cache[signal]; ok {
		return cached
	}

	var out evaluation
	if def, ok := signalRegistry[signal]; ok {
		out = def.evaluate(e)
		if out.description == "" {
			out.description = def.description
		}
		e.cache[signal] = out
		return out
	}

	switch {
	case strings.HasPrefix(signal, "log.contains:"):
		pattern := strings.TrimSpace(strings.TrimPrefix(signal, "log.contains:"))
		lines := e.linesContaining(pattern)
		out = evaluation{
			matched:     len(lines) > 0,
			description: genericSignalDescription(signal),
			evidence:    lines,
		}
	case strings.HasPrefix(signal, "log.absent:"):
		pattern := strings.TrimSpace(strings.TrimPrefix(signal, "log.absent:"))
		lines := e.linesContaining(pattern)
		out = evaluation{
			matched:     len(lines) == 0,
			description: genericSignalDescription(signal),
		}
	case strings.HasPrefix(signal, "delta.signal:"):
		id := strings.TrimSpace(strings.TrimPrefix(signal, "delta.signal:"))
		delta, ok := e.deltaSignals[id]
		out = evaluation{
			matched:     ok,
			description: genericSignalDescription(signal),
		}
		if ok && strings.TrimSpace(delta.Detail) != "" {
			out.evidence = []string{strings.TrimSpace(delta.Detail)}
		}
	case strings.HasPrefix(signal, "delta.absent:"):
		id := strings.TrimSpace(strings.TrimPrefix(signal, "delta.absent:"))
		_, ok := e.deltaSignals[id]
		out = evaluation{
			matched:     !ok,
			description: genericSignalDescription(signal),
		}
	case strings.HasPrefix(signal, "context.stage:"):
		stage := strings.TrimSpace(strings.TrimPrefix(signal, "context.stage:"))
		out = evaluation{
			matched:     strings.EqualFold(strings.TrimSpace(e.context.Stage), stage),
			description: genericSignalDescription(signal),
			evidence:    contextEvidence(e.context.Stage),
		}
	case strings.HasPrefix(signal, "context.stage.absent:"):
		stage := strings.TrimSpace(strings.TrimPrefix(signal, "context.stage.absent:"))
		out = evaluation{
			matched:     !strings.EqualFold(strings.TrimSpace(e.context.Stage), stage),
			description: genericSignalDescription(signal),
			evidence:    contextEvidence(e.context.Stage),
		}
	}

	e.cache[signal] = out
	return out
}

func (e *environment) linesContaining(pattern string) []string {
	pattern = normalize(pattern)
	if pattern == "" {
		return nil
	}
	out := make([]string, 0, 3)
	seen := make(map[string]struct{})
	for _, line := range e.lines {
		if !strings.Contains(line.Normalized, pattern) {
			continue
		}
		original := strings.TrimSpace(line.Original)
		if original == "" {
			continue
		}
		if _, ok := seen[original]; ok {
			continue
		}
		seen[original] = struct{}{}
		out = append(out, original)
		if len(out) >= 3 {
			break
		}
	}
	return out
}

func logSignal(description string, patterns ...string) signalDef {
	return signalDef{
		description: description,
		evaluate: func(env *environment) evaluation {
			for _, pattern := range patterns {
				if lines := env.linesContaining(pattern); len(lines) > 0 {
					return evaluation{
						matched:     true,
						description: description,
						evidence:    lines,
					}
				}
			}
			return evaluation{description: description}
		},
	}
}

func unionSignal(description string, signals ...string) signalDef {
	return signalDef{
		description: description,
		evaluate: func(env *environment) evaluation {
			var evidence []string
			seen := make(map[string]struct{})
			for _, signal := range signals {
				match := env.evaluate(signal)
				if !match.matched {
					continue
				}
				for _, line := range match.evidence {
					if _, ok := seen[line]; ok {
						continue
					}
					seen[line] = struct{}{}
					evidence = append(evidence, line)
				}
			}
			return evaluation{
				matched:     len(evidence) > 0,
				description: description,
				evidence:    evidence,
			}
		},
	}
}

func absentSignal(description, base string) signalDef {
	return signalDef{
		description: description,
		evaluate: func(env *environment) evaluation {
			match := env.evaluate(base)
			return evaluation{
				matched:     !match.matched,
				description: description,
			}
		},
	}
}

func deltaSignal(alias, target string) signalDef {
	return signalDef{
		description: alias,
		evaluate: func(env *environment) evaluation {
			item, ok := env.deltaSignals[target]
			out := evaluation{
				matched:     ok,
				description: alias,
			}
			if ok && strings.TrimSpace(item.Detail) != "" {
				out.evidence = []string{strings.TrimSpace(item.Detail)}
			}
			return out
		},
	}
}

var signalRegistry map[string]signalDef

func init() {
	signalRegistry = map[string]signalDef{
		"cache.restore.detected":              logSignal("a cache restore or cache hit step was detected", "cache hit", "restored cache", "restore cache", "cache restored", "using cache", "actions/cache", "cache key"),
		"cache.restore.absent":                absentSignal("no cache restore step was detected", "cache.restore.detected"),
		"dependency.cache.corrupt":            logSignal("cache integrity or archive corruption was reported", "cache is corrupted", "corrupt cache", "invalid cache", "cache hit but install failed", "checksum mismatch", "hash mismatch", "unexpected end of archive", "corrupted tar archive", "bad gzip"),
		"dependency.resolution.conflict":      logSignal("resolver conflict wording points to incompatible dependency requirements", "conflicting requirements", "dependency conflict", "version conflict", "incompatible versions", "requires a different version", "error: there is a conflict", "failed to resolve dependencies", "dependency resolution failed"),
		"dependency.lockfile.sync_error":      unionSignal("a lockfile or generated dependency metadata drift signal was present", "dependency.npm.lockfile.sync_error", "dependency.pnpm.lockfile.sync_error", "dependency.yarn.lockfile.sync_error", "dependency.poetry.lockfile.stale", "dependency.go.sum.missing"),
		"dependency.npm.lockfile.sync_error":  logSignal("npm reported package.json and package-lock.json drift", "npm ci can only install packages when your package.json and package-lock.json", "package.json and package-lock.json are in sync", "missing package-lock.json", "npm err! cipm can only install packages", "package-lock.json does not exist", "run `npm install` to generate a lockfile"),
		"dependency.pnpm.lockfile.sync_error": logSignal("pnpm reported a frozen lockfile mismatch", "err_pnpm_outdated_lockfile", "err_pnpm_frozen_lockfile", "err_pnpm_lockfile_version", "cannot install with `frozen-lockfile`", "pnpm-lock.yaml is not up to date", "lockfile is not up to date with", "run `pnpm install` to update the lockfile", "pnpm install --frozen-lockfile"),
		"dependency.yarn.lockfile.sync_error": logSignal("Yarn reported an out-of-date immutable or frozen lockfile", "your lockfile needs to be updated", "frozen-lockfile", "yarn.lock: no such file or directory", "error your lockfile needs to be updated", "yarn install --frozen-lockfile", "yn0028", "--immutable", "the lockfile would have been modified by this install"),
		"dependency.poetry.lockfile.stale":    logSignal("Poetry reported that pyproject.toml and poetry.lock drifted apart", "poetry.lock is not consistent with pyproject.toml", "run `poetry lock [--no-update]` to fix it.", "installing dependencies from lock file", "version solving failed"),
		"dependency.go.sum.missing":           logSignal("Go reported missing or stale module checksum metadata", "missing go.sum entry for module providing package", "go.sum file is not up to date", "verifying module", "no required module provides"),
		"dependency.hash.mismatch":            logSignal("a package integrity hash mismatch was reported", "these packages do not match the hashes from the requirements file", "expected sha256", "hash mismatch (got:", "does not match expected hash", "record mismatch", "hash of the downloaded file", "there are no versions that match the hash"),
		"dependency.package.not_found":        logSignal("the package manager could not find the requested package or version", "error: could not find a version that satisfies", "error: no matching distribution found", "could not find package", "package not found", "e: unable to locate package"),
		"dependency.pip.install.failed":       logSignal("pip-specific install or wheel-building signals were present", "pip install", "pip._internal", "resolutionimpossible", "could not build wheels for", "error: command 'gcc' failed", "error: pip's dependency resolver", "requirements were not successfully fulfilled", "preparing metadata (setup.py) did not run successfully", "could not fetch url", "connection error", "error: failed building wheel"),
		"dependency.lockfile.changed":         deltaSignal("dependency-related files changed in the delta", "delta.dependency.changed"),
		"dependency.scope.changed":            deltaSignal("the failure scope changed in the delta", "delta.scope.changed"),
		"runtime.node.version.mismatch":       logSignal("Node.js engine or version requirements were violated", "the engine \"node\" is incompatible with this module", "required engine node", "requires node version", "engine: unsatisfied", "unsupported engine", "error node@", "warning: you are using node.js", "this version of node is not supported", "expected node version", "engines: {\"node\":"),
		"runtime.language.version.mismatch":   logSignal("Python, Ruby, or Go runtime requirements were violated", "python version", "requires python", "requires ruby version", "go version", "unsupported python", "ruby version", "go.mod requires go"),
		"runtime.version.mismatch":            unionSignal("a runtime version mismatch signal was present", "runtime.node.version.mismatch", "runtime.language.version.mismatch"),
		"test.retry.detected":                 logSignal("the failure looked intermittent or retried", "failed on attempt", "intermittent failure", "test is flaky", "flaky test", "retry attempt", "test passed on retry", "sporadically fails"),
		"test.timeout.detected":               logSignal("the test runner reported a timeout or hang", "test timed out", "exceeded timeout", "timeout - async callback was not invoked within", "test timeout of", "jest timeout", "go test -timeout", "test exceeded", "async test timed out", "waiting for assertion"),
		"test.parallel.resource_conflict":     logSignal("parallel test execution collided on a shared resource", "resource is busy", "test is not parallelizable", "race detected", "too many connections", "database locked"),
		"test.fixture.missing":                logSignal("the failing test could not load required fixture or seed data", "fixture not found", "missing fixture", "fixture file missing", "filenotfounderror", "could not find test data", "test helper", "before(:each) failed", "beforeall", "setup failed", "testdata"),
		"test.database.duplicate_state":       logSignal("duplicate rows or unexpected existing data point to database state pollution", "already exists", "duplicate key value", "uniqueconstraintviolation", "integrityerror", "duplicate entry", "expected count to be", "pg::uniqueviolation", "sqlstate 23505", "violates unique constraint"),
		"test.order.dependency":               logSignal("the failure referenced leaked state or prior-test ordering", "depends on previous test", "test order", "test assumes", "leaked state", "global state", "left over from previous"),
		"test.failure.introduced":             deltaSignal("the delta introduced a newly failing test", "delta.test.failure.introduced"),
		"error.new":                           deltaSignal("the delta added a new error signature", "delta.error.new"),
	}
}

func rivalryKey(result model.Result) string {
	parts := make([]string, 0, 6)
	parts = append(parts, strings.TrimSpace(result.Playbook.Category))
	if result.Hypothesis != nil {
		parts = append(parts, matchIDs(result.Hypothesis.Supports)...)
		parts = append(parts, matchIDs(result.Hypothesis.Discriminators)...)
	}
	if len(parts) == 1 {
		parts = append(parts, trimEvidence(result.Evidence)...)
	}
	if len(parts) == 1 {
		parts = append(parts, result.Playbook.ID)
	}
	sort.Strings(parts[1:])
	return strings.Join(parts, "|")
}

func matchIDs(items []model.HypothesisMatch) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		signal := strings.TrimSpace(item.Signal)
		if signal == "" {
			continue
		}
		out = append(out, signal)
	}
	return out
}

func trimEvidence(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
		if len(out) >= 2 {
			break
		}
	}
	return out
}

func hypothesisScore(assessment *model.HypothesisAssessment, fallback float64) float64 {
	if assessment == nil {
		return round(fallback)
	}
	return round(assessment.FinalScore)
}

func hypothesisEliminated(assessment *model.HypothesisAssessment) bool {
	return assessment != nil && assessment.Eliminated
}

func confidenceText(confidence float64) string {
	switch {
	case confidence >= 0.8:
		return "High"
	case confidence >= 0.55:
		return "Medium"
	default:
		return "Low"
	}
}

func contextEvidence(stage string) []string {
	stage = strings.TrimSpace(stage)
	if stage == "" {
		return nil
	}
	return []string{"stage: " + stage}
}

func genericSignalDescription(signal string) string {
	switch {
	case strings.HasPrefix(signal, "log.contains:"):
		return fmt.Sprintf("log contains %q", strings.TrimSpace(strings.TrimPrefix(signal, "log.contains:")))
	case strings.HasPrefix(signal, "log.absent:"):
		return fmt.Sprintf("log does not contain %q", strings.TrimSpace(strings.TrimPrefix(signal, "log.absent:")))
	case strings.HasPrefix(signal, "delta.signal:"):
		return fmt.Sprintf("delta signal %q is present", strings.TrimSpace(strings.TrimPrefix(signal, "delta.signal:")))
	case strings.HasPrefix(signal, "delta.absent:"):
		return fmt.Sprintf("delta signal %q is absent", strings.TrimSpace(strings.TrimPrefix(signal, "delta.absent:")))
	case strings.HasPrefix(signal, "context.stage:"):
		return fmt.Sprintf("analysis stage is %q", strings.TrimSpace(strings.TrimPrefix(signal, "context.stage:")))
	case strings.HasPrefix(signal, "context.stage.absent:"):
		return fmt.Sprintf("analysis stage is not %q", strings.TrimSpace(strings.TrimPrefix(signal, "context.stage.absent:")))
	default:
		return signal
	}
}

func normalize(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}

func round(value float64) float64 {
	return math.Round(value*100) / 100
}
