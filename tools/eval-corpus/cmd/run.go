package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"faultline/tools/eval-corpus/model"
	"faultline/tools/eval-corpus/runner"
)

func newRunCommand() *cobra.Command {
	var (
		corpusPath  string
		outPath     string
		workers     int
		playbookDir string
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Evaluate a fixture corpus against Faultline playbooks",
		Long: `run reads a JSONL corpus produced by 'faultline-eval ingest', passes each
fixture through the Faultline detection engine, and writes a JSONL results file
where each line is an EvalResult sorted by fixture_id.

The results file can be inspected with:
  faultline-eval report --results <out>`,
		Example: `  faultline-eval run --corpus corpus.jsonl --out results.jsonl
  faultline-eval run --corpus corpus.jsonl --out results.jsonl --workers 8
  faultline-eval run --corpus corpus.jsonl --out results.jsonl --playbook-dir ./playbooks`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEval(corpusPath, outPath, workers, playbookDir, cmd)
		},
	}

	cmd.Flags().StringVar(&corpusPath, "corpus", "", "corpus JSONL file produced by ingest (required)")
	cmd.Flags().StringVar(&outPath, "out", "", "output results JSONL path (required)")
	cmd.Flags().IntVar(&workers, "workers", 4, "number of parallel evaluation workers")
	cmd.Flags().StringVar(&playbookDir, "playbook-dir", "", "override bundled playbook directory")
	_ = cmd.MarkFlagRequired("corpus")
	_ = cmd.MarkFlagRequired("out")

	return cmd
}

func runEval(corpusPath, outPath string, workers int, playbookDir string, cmd *cobra.Command) error {
	if workers < 1 {
		workers = 1
	}

	// Open corpus.
	corpusFile, err := os.Open(corpusPath) // #nosec G304
	if err != nil {
		return fmt.Errorf("open corpus: %w", err)
	}
	defer corpusFile.Close()

	// Open output.
	outFile, err := os.Create(outPath) // #nosec G304
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer outFile.Close()

	// Set up worker pool channels.
	fixtures := make(chan model.Fixture, workers*4)
	results := make(chan model.EvalResult, workers*4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Collect goroutine: gathers results as workers emit them.
	var allResults []model.EvalResult
	collectDone := make(chan struct{})
	go func() {
		defer close(collectDone)
		for r := range results {
			allResults = append(allResults, r)
		}
	}()

	// Worker goroutine: starts all workers reading from fixtures channel.
	runnerDone := make(chan error, 1)
	go func() {
		runnerDone <- runner.Run(ctx, runner.Options{
			PlaybookDir: playbookDir,
			Workers:     workers,
		}, fixtures, results)
		close(results)
	}()

	// Feed fixtures from corpus JSONL (streaming, no full load into memory).
	sc := bufio.NewScanner(corpusFile)
	sc.Buffer(make([]byte, 4<<20), 4<<20) // 4 MiB per line to handle large logs
	var total int
	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var fix model.Fixture
		if err := json.Unmarshal(line, &fix); err != nil {
			cancel()
			return fmt.Errorf("corpus line %d: %w", lineNum, err)
		}
		select {
		case fixtures <- fix:
			total++
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if err := sc.Err(); err != nil {
		cancel()
		return fmt.Errorf("scan corpus: %w", err)
	}
	close(fixtures) // signal workers that input is exhausted

	// Wait for all workers to finish.
	if err := <-runnerDone; err != nil {
		return fmt.Errorf("runner: %w", err)
	}

	// Wait for collector.
	<-collectDone

	// Sort by FixtureID for deterministic output.
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].FixtureID < allResults[j].FixtureID
	})

	// Write results JSONL.
	bw := bufio.NewWriterSize(outFile, 1<<20)
	enc := json.NewEncoder(bw)
	matched := 0
	for _, r := range allResults {
		if err := enc.Encode(r); err != nil {
			return fmt.Errorf("write result: %w", err)
		}
		if r.Matched {
			matched++
		}
	}
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush output: %w", err)
	}

	fmt.Fprintf(cmd.ErrOrStderr(),
		"run: total=%d matched=%d unmatched=%d workers=%d\n",
		total, matched, total-matched, workers)

	return nil
}
