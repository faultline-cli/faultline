package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"faultline/internal/playbooks"
)

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSliceFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("pack path must not be empty")
	}
	*s = append(*s, value)
	return nil
}

func main() {
	var packs stringSliceFlag
	var review bool
	flag.Var(&packs, "pack", "external playbook pack directory to compose with the bundled starter catalog; repeatable")
	flag.BoolVar(&review, "review", false, "print deterministic overlap review for the composed catalog")
	flag.Parse()

	if len(packs) == 0 {
		fmt.Fprintln(os.Stderr, "at least one --pack path is required")
		os.Exit(1)
	}

	catalog := playbooks.NewCatalogWithOptions(playbooks.CatalogOptions{ExtraPackDirs: packs})
	resolved, err := catalog.Packs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	pbs, err := catalog.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	names := make([]string, 0, len(resolved))
	for _, pack := range resolved {
		names = append(names, pack.Name)
	}

	fmt.Printf("composed %d packs successfully: %s\n", len(resolved), strings.Join(names, ", "))
	fmt.Printf("loaded %d playbooks\n", len(pbs))
	if review {
		conflicts := playbooks.FindPatternConflicts(pbs)
		fmt.Printf("found %d pattern conflicts across composed packs\n", len(conflicts))
		fmt.Print(playbooks.FormatPatternConflicts(conflicts))
	}
}
