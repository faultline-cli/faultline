package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"faultline/internal/playbooks"
)

func main() {
	var (
		baselinePath   string
		updateBaseline bool
		verbose        bool
	)
	flag.StringVar(&baselinePath, "baseline", "playbooks/bundled/pattern-conflicts.baseline.txt", "checked-in pattern conflict baseline")
	flag.BoolVar(&updateBaseline, "update-baseline", false, "rewrite the pattern conflict baseline")
	flag.BoolVar(&verbose, "verbose", false, "print the full conflict report")
	flag.Parse()

	pbs, err := playbooks.NewCatalog("").Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	conflicts := playbooks.FindPatternConflicts(pbs)
	report := []byte(playbooks.FormatPatternConflicts(conflicts))
	if verbose {
		fmt.Print(string(report))
	}
	if updateBaseline {
		if err := os.MkdirAll(filepath.Dir(baselinePath), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(baselinePath, report, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("updated playbook review baseline: %s\n", baselinePath)
		return
	}
	baseline, err := os.ReadFile(baselinePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read playbook review baseline: %v\n", err)
		os.Exit(1)
	}
	if !bytes.Equal(report, baseline) {
		fmt.Fprintf(os.Stderr, "playbook pattern conflicts drifted from %s\n", baselinePath)
		fmt.Fprintln(os.Stderr, "Run `make review-update` after reviewing intentional conflict changes.")
		if !verbose {
			fmt.Fprintln(os.Stderr, "Use `make review-verbose` to print the full report.")
		}
		os.Exit(1)
	}
	fmt.Printf("playbook review passed (%d classified conflict patterns)\n", len(conflicts))
}
