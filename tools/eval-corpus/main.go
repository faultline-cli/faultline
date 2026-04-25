// Command faultline-eval is a deterministic log ingestion and evaluation
// pipeline for measuring Faultline coverage over large heterogeneous datasets.
//
// Usage:
//
//	faultline-eval ingest --config config.yaml --out corpus.jsonl
//	faultline-eval run    --corpus corpus.jsonl --out results.jsonl
//	faultline-eval report --results results.jsonl
package main

import (
	"fmt"
	"os"

	"faultline/tools/eval-corpus/cmd"
)

func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
