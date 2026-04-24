package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"faultline/internal/model"
	"faultline/internal/playbooks"
)

// newCoverageCommand returns the hidden `faultline coverage` command.
//
// It loads the configured playbook catalog, counts playbooks by category,
// identifies which playbooks have corresponding fixture files in the bundled
// fixture directory, and reports duplicate IDs.
//
// This command is intentionally hidden until it has release-grade tests and docs.
func newCoverageCommand() *cobra.Command {
	var (
		playbookDir   string
		playbookPacks []string
		fixtureDir    string
		jsonOut       bool
	)

	cmd := &cobra.Command{
		Use:    "coverage",
		Short:  "Report playbook count, category breakdown, and fixture coverage",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load playbooks using the same resolution as analyze.
			pbs, err := loadPlaybooksForCoverage(playbookDir, playbookPacks)
			if err != nil {
				return fmt.Errorf("load playbooks: %w", err)
			}

			// Detect duplicate IDs (LoadDir rejects duplicates within a single dir,
			// but cross-pack duplicates can still surface when packs are merged).
			idSeen := map[string]int{}
			for _, pb := range pbs {
				idSeen[pb.ID]++
			}
			var duplicates []string
			for id, count := range idSeen {
				if count > 1 {
					duplicates = append(duplicates, fmt.Sprintf("%s (×%d)", id, count))
				}
			}
			sort.Strings(duplicates)

			// Group by category.
			byCategory := map[string][]string{}
			for _, pb := range pbs {
				cat := strings.TrimSpace(pb.Category)
				if cat == "" {
					cat = "uncategorized"
				}
				byCategory[cat] = append(byCategory[cat], pb.ID)
			}

			// Resolve fixture directory.
			fixtureRoot := fixtureDir
			if fixtureRoot == "" {
				fixtureRoot = resolveDefaultFixtureDir()
			}

			// Build set of discovered fixture stems (.log filename without extension).
			fixtureStems := map[string]struct{}{}
			if fixtureRoot != "" {
				_ = filepath.WalkDir(fixtureRoot, func(path string, d fs.DirEntry, werr error) error {
					if werr != nil || d.IsDir() {
						return nil
					}
					name := d.Name()
					if strings.HasSuffix(name, ".log") {
						fixtureStems[strings.TrimSuffix(name, ".log")] = struct{}{}
					}
					return nil
				})
			}

			// Identify playbooks with no matching fixture file.
			var missingFixtures []string
			for _, pb := range pbs {
				if _, ok := fixtureStems[pb.ID]; !ok {
					missingFixtures = append(missingFixtures, pb.ID)
				}
			}
			sort.Strings(missingFixtures)

			if jsonOut {
				return printCoverageJSON(cmd.OutOrStdout(), pbs, byCategory, missingFixtures, duplicates, fixtureRoot)
			}
			return printCoverageText(cmd.OutOrStdout(), pbs, byCategory, missingFixtures, duplicates, fixtureRoot)
		},
	}

	cmd.Flags().StringVar(&playbookDir, "playbooks", "", "override playbook directory")
	cmd.Flags().StringSliceVar(&playbookPacks, "playbook-pack", nil, "load one or more extra playbook pack directories")
	cmd.Flags().StringVar(&fixtureDir, "fixture-dir", "", "directory to scan for .log fixture files (defaults to bundled testdata)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	return cmd
}

// loadPlaybooksForCoverage mirrors the playbook loading logic used by analyze.
func loadPlaybooksForCoverage(playbookDir string, playbookPacks []string) ([]model.Playbook, error) {
	if playbookDir != "" && len(playbookPacks) == 0 {
		return playbooks.LoadDir(playbookDir)
	}
	if len(playbookPacks) > 0 {
		bundledDir := playbookDir
		if bundledDir == "" {
			var err error
			bundledDir, err = playbooks.DefaultDir()
			if err != nil {
				return nil, err
			}
		}
		packs := []playbooks.Pack{
			{Name: playbooks.BundledPackName, Root: bundledDir},
		}
		for _, packDir := range playbookPacks {
			packs = append(packs, playbooks.Pack{
				Name: filepath.Base(packDir),
				Root: packDir,
			})
		}
		return playbooks.LoadPacks(packs)
	}
	return playbooks.LoadDefault()
}

// resolveDefaultFixtureDir finds the bundled engine testdata/fixtures directory
// by walking upward from the working directory.
func resolveDefaultFixtureDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := cwd
	for {
		candidate := filepath.Join(dir, "internal", "engine", "testdata", "fixtures")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func printCoverageText(w io.Writer, pbs []model.Playbook, byCategory map[string][]string, missingFixtures, duplicates []string, fixtureRoot string) error {
	totalWithFixtures := len(pbs) - len(missingFixtures)

	fmt.Fprintf(w, "Playbook coverage report\n\n")
	fmt.Fprintf(w, "  Total playbooks : %d\n", len(pbs))
	fmt.Fprintf(w, "  With fixtures   : %d / %d\n", totalWithFixtures, len(pbs))
	if fixtureRoot != "" {
		fmt.Fprintf(w, "  Fixture dir     : %s\n", fixtureRoot)
	}
	fmt.Fprintln(w)

	// Build a quick missing-set for per-category covered counts.
	missingSet := map[string]struct{}{}
	for _, id := range missingFixtures {
		missingSet[id] = struct{}{}
	}

	// Category breakdown.
	cats := make([]string, 0, len(byCategory))
	for cat := range byCategory {
		cats = append(cats, cat)
	}
	sort.Strings(cats)

	fmt.Fprintf(w, "By category:\n")
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, cat := range cats {
		ids := byCategory[cat]
		sort.Strings(ids)
		covered := 0
		for _, id := range ids {
			if _, missing := missingSet[id]; !missing {
				covered++
			}
		}
		fmt.Fprintf(tw, "  %s\t%d playbooks\t%d with fixtures\n", cat, len(ids), covered)
	}
	_ = tw.Flush()

	if len(missingFixtures) > 0 {
		fmt.Fprintf(w, "\nPlaybooks missing fixtures (%d):\n", len(missingFixtures))
		for _, id := range missingFixtures {
			fmt.Fprintf(w, "  - %s\n", id)
		}
	}

	if len(duplicates) > 0 {
		fmt.Fprintf(w, "\nDuplicate IDs (%d):\n", len(duplicates))
		for _, d := range duplicates {
			fmt.Fprintf(w, "  - %s\n", d)
		}
	} else {
		fmt.Fprintf(w, "\nNo duplicate IDs detected.\n")
	}

	return nil
}

type coverageReportJSON struct {
	TotalPlaybooks  int                    `json:"total_playbooks"`
	WithFixtures    int                    `json:"with_fixtures"`
	FixtureDir      string                 `json:"fixture_dir,omitempty"`
	ByCategory      []coverageCategoryJSON `json:"by_category"`
	MissingFixtures []string               `json:"missing_fixtures"`
	DuplicateIDs    []string               `json:"duplicate_ids"`
}

type coverageCategoryJSON struct {
	Category     string   `json:"category"`
	Count        int      `json:"count"`
	WithFixtures int      `json:"with_fixtures"`
	PlaybookIDs  []string `json:"playbook_ids"`
}

func printCoverageJSON(w io.Writer, pbs []model.Playbook, byCategory map[string][]string, missingFixtures, duplicates []string, fixtureRoot string) error {
	missingSet := map[string]struct{}{}
	for _, id := range missingFixtures {
		missingSet[id] = struct{}{}
	}

	cats := make([]string, 0, len(byCategory))
	for cat := range byCategory {
		cats = append(cats, cat)
	}
	sort.Strings(cats)

	items := make([]coverageCategoryJSON, 0, len(cats))
	for _, cat := range cats {
		ids := byCategory[cat]
		sort.Strings(ids)
		covered := 0
		for _, id := range ids {
			if _, missing := missingSet[id]; !missing {
				covered++
			}
		}
		items = append(items, coverageCategoryJSON{
			Category:     cat,
			Count:        len(ids),
			WithFixtures: covered,
			PlaybookIDs:  ids,
		})
	}

	mf := missingFixtures
	if mf == nil {
		mf = []string{}
	}
	dups := duplicates
	if dups == nil {
		dups = []string{}
	}

	payload := coverageReportJSON{
		TotalPlaybooks:  len(pbs),
		WithFixtures:    len(pbs) - len(missingFixtures),
		FixtureDir:      fixtureRoot,
		ByCategory:      items,
		MissingFixtures: mf,
		DuplicateIDs:    dups,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal coverage JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
