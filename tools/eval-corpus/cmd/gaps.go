package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"faultline/tools/eval-corpus/gaps"
	"faultline/tools/eval-corpus/model"
	"faultline/tools/eval-corpus/report"
)

func newGapsCommand() *cobra.Command {
	var (
		resultsPath  string
		fixturesPath string
		outDir       string
		maxSamples   int
	)

	cmd := &cobra.Command{
		Use:   "gaps",
		Short: "Identify high-value missing playbooks from unmatched evaluation results",
		Long: `gaps clusters unmatched fixtures by their normalised error signature and
generates reviewable playbook stubs for the top recurring patterns.

Outputs written to --out:

  clusters.jsonl        — one cluster record per line (JSONL)
  cluster-summary.md   — human-readable Markdown table
  samples/<cluster>/   — representative log files per cluster
  playbook-stubs/      — YAML stubs ready for manual review`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGaps(resultsPath, fixturesPath, outDir, maxSamples)
		},
	}

	cmd.Flags().StringVar(&resultsPath, "results", "", "path to results JSONL produced by `faultline-eval run`")
	cmd.Flags().StringVar(&fixturesPath, "fixtures", "", "path to corpus JSONL (used to extract sample log text)")
	cmd.Flags().StringVar(&outDir, "out", "gaps", "output directory")
	cmd.Flags().IntVar(&maxSamples, "max-samples", 5, "maximum sample fixtures to extract per cluster")

	_ = cmd.MarkFlagRequired("results")

	return cmd
}

func runGaps(resultsPath, fixturesPath, outDir string, maxSamples int) error {
	results, err := report.LoadResults(resultsPath)
	if err != nil {
		return fmt.Errorf("load results: %w", err)
	}

	// Build fixture index if a corpus path was provided.
	fixtureIndex := map[string]model.Fixture{}
	if fixturesPath != "" {
		fixtureIndex, err = loadFixtureIndex(fixturesPath)
		if err != nil {
			return fmt.Errorf("load fixtures: %w", err)
		}
	}

	clusters := gaps.BuildClusters(results, maxSamples)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	// clusters.jsonl
	if err := writeClustersJSONL(filepath.Join(outDir, "clusters.jsonl"), clusters); err != nil {
		return err
	}

	// cluster-summary.md
	mdPath := filepath.Join(outDir, "cluster-summary.md")
	mdFile, err := os.Create(mdPath)
	if err != nil {
		return fmt.Errorf("create cluster-summary.md: %w", err)
	}
	defer mdFile.Close()
	bw := bufio.NewWriter(mdFile)
	gaps.PrintClusterSummaryMarkdown(bw, clusters)
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush cluster-summary.md: %w", err)
	}

	// samples/<cluster>/ and playbook-stubs/
	samplesDir := filepath.Join(outDir, "samples")
	stubsDir := filepath.Join(outDir, "playbook-stubs")
	if err := os.MkdirAll(stubsDir, 0o755); err != nil {
		return fmt.Errorf("create stubs dir: %w", err)
	}

	for _, c := range clusters {
		if len(fixtureIndex) > 0 {
			sampleDir := filepath.Join(samplesDir, c.ClusterID)
			if err := gaps.WriteSamples(sampleDir, c, fixtureIndex); err != nil {
				return fmt.Errorf("write samples for %s: %w", c.ClusterID, err)
			}
		}

		stub := gaps.GenerateStub(c)
		stubPath := filepath.Join(stubsDir, c.ClusterID+".yaml")
		if err := os.WriteFile(stubPath, []byte(stub.YAML), 0o644); err != nil {
			return fmt.Errorf("write stub %s: %w", c.ClusterID, err)
		}
	}

	fmt.Fprintf(os.Stderr, "gaps: %d clusters written to %s\n", len(clusters), outDir)
	return nil
}

func writeClustersJSONL(path string, clusters []gaps.Cluster) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create clusters.jsonl: %w", err)
	}
	defer f.Close()

	bw := bufio.NewWriterSize(f, 1<<20)
	enc := json.NewEncoder(bw)
	for _, c := range clusters {
		if err := enc.Encode(c); err != nil {
			return fmt.Errorf("encode cluster: %w", err)
		}
	}
	return bw.Flush()
}

// loadFixtureIndex reads a corpus JSONL and returns a map of fixture ID → Fixture.
func loadFixtureIndex(path string) (map[string]model.Fixture, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	index := map[string]model.Fixture{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 4<<20), 4<<20)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var fix model.Fixture
		if err := json.Unmarshal(line, &fix); err != nil {
			continue // skip malformed lines
		}
		index[fix.ID] = fix
	}
	return index, sc.Err()
}
