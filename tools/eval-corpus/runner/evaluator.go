// Package runner evaluates a corpus of fixtures against Faultline's detection
// engine and produces structured EvalResult records.
package runner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	faultlineengine "faultline/internal/engine"
	"faultline/tools/eval-corpus/model"
)

// Options configures a corpus evaluation run.
type Options struct {
	// PlaybookDir overrides the default bundled playbook directory.
	// Leave empty to use the bundled starter pack.
	PlaybookDir string
	// Workers is the number of concurrent evaluation goroutines.
	// Must be >= 1. Defaults to 1 when omitted.
	Workers int
}

// Run evaluates every fixture read from in, writes EvalResults to out, and
// returns when in is closed or ctx is cancelled. Results are emitted in
// arrival order (non-deterministic); callers that need stable ordering must
// sort by FixtureID after collection.
//
// Run creates one engine.Engine per worker. Playbooks are re-parsed on each
// AnalyzeReader call (engine behaviour); workers share no mutable state.
func Run(ctx context.Context, opts Options, in <-chan model.Fixture, out chan<- model.EvalResult) error {
	workers := opts.Workers
	if workers < 1 {
		workers = 1
	}

	engineOpts := faultlineengine.Options{
		PlaybookDir:       opts.PlaybookDir,
		GitContextEnabled: false,
		BayesEnabled:      false,
		NoHistory:         true,
	}

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		eng := faultlineengine.New(engineOpts)
		go func(e *faultlineengine.Engine) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case fix, ok := <-in:
					if !ok {
						return
					}
					out <- evaluate(e, fix)
				}
			}
		}(eng)
	}

	wg.Wait()
	return ctx.Err()
}

// evaluate runs the Faultline engine against a single fixture and returns
// a structured result. It never panics; engine errors are recorded in
// EvalResult.Error.
func evaluate(eng *faultlineengine.Engine, fix model.Fixture) model.EvalResult {
	start := time.Now()

	firstLine, firstLineTag := extractFirstLine(fix.Raw)

	result := model.EvalResult{
		FixtureID:        fix.ID,
		Source:           fix.Source,
		FirstLineTag:     firstLineTag,
		FirstLineSnippet: firstLine,
	}

	r := strings.NewReader(fix.Raw)
	analysis, err := eng.AnalyzeReader(r)
	result.DurationMS = time.Since(start).Milliseconds()

	switch {
	case errors.Is(err, faultlineengine.ErrNoInput):
		result.Error = "empty fixture"
		return result
	case errors.Is(err, faultlineengine.ErrNoMatch):
		// Analysed successfully; no playbook matched.
		return result
	case err != nil:
		result.Error = err.Error()
		return result
	}

	if analysis != nil && len(analysis.Results) > 0 {
		top := analysis.Results[0]
		result.Matched = true
		result.FailureID = top.Playbook.ID
		result.Confidence = top.Confidence
		result.Evidence = top.Evidence
	}

	return result
}

// extractFirstLine returns the first normalised line of raw (lowercased, at
// most 120 bytes) together with its 8-character SHA-256 prefix for clustering.
func extractFirstLine(raw string) (snippet, tag string) {
	line := raw
	if idx := strings.IndexByte(raw, '\n'); idx >= 0 {
		line = raw[:idx]
	}
	line = strings.ToLower(strings.TrimSpace(line))
	if len(line) > 120 {
		line = line[:120]
	}
	if line == "" {
		return "", ""
	}
	sum := sha256.Sum256([]byte(line))
	return line, hex.EncodeToString(sum[:])[:8]
}
