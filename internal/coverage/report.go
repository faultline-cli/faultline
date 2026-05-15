package coverage

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

	"faultline/internal/fixtures"
	"faultline/internal/model"
	"faultline/internal/playbooks"
)

// Options configures playbook coverage reporting.
type Options struct {
	PlaybookDir      string
	PlaybookPackDirs []string
	FixtureRoot      string
}

// Report describes behavioral fixture evidence for the resolved playbook set.
type Report struct {
	TotalPlaybooks          int          `json:"total_playbooks"`
	WithFixtures            int          `json:"with_fixtures"`
	FixtureRoot             string       `json:"fixture_root,omitempty"`
	FixtureMode             string       `json:"fixture_mode,omitempty"`
	FixtureCount            int          `json:"fixture_count"`
	PositiveFixtureCount    int          `json:"positive_fixture_count"`
	NegativeAssertionCount  int          `json:"negative_assertion_count"`
	StrictTop1FixtureCount  int          `json:"strict_top_1_fixture_count"`
	ByCategory              []Category   `json:"by_category"`
	ByDomain                []Domain     `json:"by_domain"`
	MissingFixtures         []string     `json:"missing_fixtures"`
	SourceDetectorPlaybooks []string     `json:"source_detector_playbooks"`
	DuplicateIDs            []string     `json:"duplicate_ids"`
	Robustness              []Robustness `json:"robustness"`
}

// Category groups playbooks and fixture evidence by playbook category.
type Category struct {
	Category           string   `json:"category"`
	Count              int      `json:"count"`
	WithFixtures       int      `json:"with_fixtures"`
	PositiveFixtures   int      `json:"positive_fixtures"`
	NegativeAssertions int      `json:"negative_assertions"`
	PlaybookIDs        []string `json:"playbook_ids"`
}

// Domain groups playbooks by product domain.
type Domain struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

// Robustness describes false-positive and false-negative risk for a playbook
// using only deterministic catalog metadata, fixture expectations, and pattern
// overlap information from the resolved playbook set.
type Robustness struct {
	PlaybookID             string   `json:"playbook_id"`
	Category               string   `json:"category"`
	Domain                 string   `json:"domain,omitempty"`
	Detector               string   `json:"detector,omitempty"`
	ConfidenceScore        int      `json:"confidence_score"`
	FalsePositiveRisk      string   `json:"false_positive_risk"`
	FalseNegativeRisk      string   `json:"false_negative_risk"`
	PositiveFixtures       int      `json:"positive_fixtures"`
	NegativeAssertions     int      `json:"negative_assertions"`
	StrictTop1Fixtures     int      `json:"strict_top_1_fixtures"`
	SharedPositivePatterns int      `json:"shared_positive_patterns"`
	ExplicitExclusions     int      `json:"explicit_exclusions"`
	RiskDrivers            []string `json:"risk_drivers"`
}

type fixtureEvidence struct {
	root                   string
	mode                   string
	fixtureCount           int
	positiveFixtureCount   int
	negativeAssertionCount int
	strictTop1FixtureCount int
	positiveByPlaybook     map[string]int
	negativeByPlaybook     map[string]int
	strictTop1ByPlaybook   map[string]int
}

// Build loads playbooks and fixture metadata, then returns a deterministic
// evidence-oriented coverage report.
func Build(opts Options) (Report, error) {
	pbs, err := playbooks.NewCatalogWithOptions(playbooks.CatalogOptions{
		OverrideDir:   opts.PlaybookDir,
		ExtraPackDirs: opts.PlaybookPackDirs,
	}).Load()
	if err != nil {
		return Report{}, err
	}

	evidence, err := collectFixtureEvidence(opts.FixtureRoot)
	if err != nil {
		return Report{}, err
	}

	duplicates := duplicatePlaybookIDs(pbs)
	byCategory := map[string][]model.Playbook{}
	domainCounts := map[string]int{}
	for _, pb := range pbs {
		category := strings.TrimSpace(pb.Category)
		if category == "" {
			category = "uncategorized"
		}
		byCategory[category] = append(byCategory[category], pb)
		if strings.TrimSpace(pb.Domain) != "" {
			domainCounts[pb.Domain]++
		}
	}

	missing := make([]string, 0)
	sourceDetector := make([]string, 0)
	withFixtures := 0
	for _, pb := range pbs {
		if pb.Detector == "source" {
			sourceDetector = append(sourceDetector, pb.ID)
			continue
		}
		if evidence.positiveByPlaybook[pb.ID] > 0 {
			withFixtures++
			continue
		}
		missing = append(missing, pb.ID)
	}
	sort.Strings(missing)
	sort.Strings(sourceDetector)

	return Report{
		TotalPlaybooks:          len(pbs),
		WithFixtures:            withFixtures,
		FixtureRoot:             evidence.root,
		FixtureMode:             evidence.mode,
		FixtureCount:            evidence.fixtureCount,
		PositiveFixtureCount:    evidence.positiveFixtureCount,
		NegativeAssertionCount:  evidence.negativeAssertionCount,
		StrictTop1FixtureCount:  evidence.strictTop1FixtureCount,
		ByCategory:              buildCategoryReports(byCategory, evidence),
		ByDomain:                buildDomainReports(domainCounts),
		MissingFixtures:         missing,
		SourceDetectorPlaybooks: sourceDetector,
		DuplicateIDs:            duplicates,
		Robustness:              buildRobustness(pbs, evidence, playbooks.FindPatternConflicts(pbs)),
	}, nil
}

// WriteText renders the coverage report for humans.
func WriteText(w io.Writer, report Report) error {
	fmt.Fprintf(w, "Playbook coverage report\n\n")
	fmt.Fprintf(w, "  Total playbooks       : %d\n", report.TotalPlaybooks)
	fmt.Fprintf(w, "  With positive fixtures: %d / %d\n", report.WithFixtures, report.TotalPlaybooks)
	fmt.Fprintf(w, "  Fixture assertions    : %d positive, %d negative\n", report.PositiveFixtureCount, report.NegativeAssertionCount)
	if report.StrictTop1FixtureCount > 0 {
		fmt.Fprintf(w, "  Strict top-1 fixtures : %d\n", report.StrictTop1FixtureCount)
	}
	if report.FixtureRoot != "" {
		fmt.Fprintf(w, "  Fixture root          : %s\n", report.FixtureRoot)
	}
	if report.FixtureMode != "" {
		fmt.Fprintf(w, "  Fixture mode          : %s\n", report.FixtureMode)
	}
	fmt.Fprintln(w)

	fmt.Fprintf(w, "By category:\n")
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, category := range report.ByCategory {
		fmt.Fprintf(
			tw,
			"  %s\t%d playbooks\t%d with fixtures\t%d positive\t%d negative\n",
			category.Category,
			category.Count,
			category.WithFixtures,
			category.PositiveFixtures,
			category.NegativeAssertions,
		)
	}
	_ = tw.Flush()

	if len(report.ByDomain) > 0 {
		fmt.Fprintf(w, "\nBy domain:\n")
		dw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		for _, domain := range report.ByDomain {
			fmt.Fprintf(dw, "  %s\t%d playbooks\n", domain.Domain, domain.Count)
		}
		_ = dw.Flush()
	}

	if len(report.SourceDetectorPlaybooks) > 0 {
		fmt.Fprintf(w, "\nSource-detector playbooks (%d) — covered by source-fixture accounting:\n", len(report.SourceDetectorPlaybooks))
		for _, id := range report.SourceDetectorPlaybooks {
			fmt.Fprintf(w, "  - %s\n", id)
		}
	}

	if len(report.MissingFixtures) > 0 {
		fmt.Fprintf(w, "\nPlaybooks missing positive fixtures (%d):\n", len(report.MissingFixtures))
		for _, id := range report.MissingFixtures {
			fmt.Fprintf(w, "  - %s\n", id)
		}
	}

	if len(report.DuplicateIDs) > 0 {
		fmt.Fprintf(w, "\nDuplicate IDs (%d):\n", len(report.DuplicateIDs))
		for _, duplicate := range report.DuplicateIDs {
			fmt.Fprintf(w, "  - %s\n", duplicate)
		}
	} else {
		fmt.Fprintf(w, "\nNo duplicate IDs detected.\n")
	}

	if len(report.Robustness) > 0 {
		fmt.Fprintf(w, "\nLowest robustness scores:\n")
		rw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		limit := len(report.Robustness)
		if limit > 12 {
			limit = 12
		}
		for _, item := range report.Robustness[:limit] {
			fmt.Fprintf(
				rw,
				"  %s\t%d\tFP %s\tFN %s\t+%d\t-%d\tstrict %d\n",
				item.PlaybookID,
				item.ConfidenceScore,
				item.FalsePositiveRisk,
				item.FalseNegativeRisk,
				item.PositiveFixtures,
				item.NegativeAssertions,
				item.StrictTop1Fixtures,
			)
		}
		_ = rw.Flush()
	}
	return nil
}

// WriteJSON renders the coverage report for automation.
func WriteJSON(w io.Writer, report Report) error {
	report = normalizeReport(report)
	data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("marshal coverage JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

func normalizeReport(report Report) Report {
	if report.ByCategory == nil {
		report.ByCategory = []Category{}
	}
	if report.ByDomain == nil {
		report.ByDomain = []Domain{}
	}
	if report.MissingFixtures == nil {
		report.MissingFixtures = []string{}
	}
	if report.SourceDetectorPlaybooks == nil {
		report.SourceDetectorPlaybooks = []string{}
	}
	if report.DuplicateIDs == nil {
		report.DuplicateIDs = []string{}
	}
	if report.Robustness == nil {
		report.Robustness = []Robustness{}
	}
	for i := range report.Robustness {
		if report.Robustness[i].RiskDrivers == nil {
			report.Robustness[i].RiskDrivers = []string{}
		}
	}
	return report
}

func buildRobustness(pbs []model.Playbook, evidence fixtureEvidence, conflicts []playbooks.PatternConflict) []Robustness {
	positiveConflicts := map[string]int{}
	explicitExclusions := map[string]int{}
	for _, conflict := range conflicts {
		for _, ref := range conflict.Positive {
			positiveConflicts[ref.PlaybookID]++
		}
		for _, ref := range conflict.Negative {
			explicitExclusions[ref.PlaybookID]++
		}
	}

	out := make([]Robustness, 0, len(pbs))
	for _, pb := range pbs {
		positive := evidence.positiveByPlaybook[pb.ID]
		negative := evidence.negativeByPlaybook[pb.ID]
		strict := evidence.strictTop1ByPlaybook[pb.ID]
		shared := positiveConflicts[pb.ID]
		exclusions := explicitExclusions[pb.ID]

		fpRisk := falsePositiveRisk(pb, shared)
		fnRisk := falseNegativeRisk(pb, positive)
		drivers := robustnessDrivers(pb, positive, negative, strict, shared, exclusions, fpRisk, fnRisk)
		score := robustnessScore(fpRisk, fnRisk, positive, negative, strict)

		category := strings.TrimSpace(pb.Category)
		if category == "" {
			category = "uncategorized"
		}
		out = append(out, Robustness{
			PlaybookID:             pb.ID,
			Category:               category,
			Domain:                 strings.TrimSpace(pb.Domain),
			Detector:               strings.TrimSpace(pb.Detector),
			ConfidenceScore:        score,
			FalsePositiveRisk:      fpRisk,
			FalseNegativeRisk:      fnRisk,
			PositiveFixtures:       positive,
			NegativeAssertions:     negative,
			StrictTop1Fixtures:     strict,
			SharedPositivePatterns: shared,
			ExplicitExclusions:     exclusions,
			RiskDrivers:            drivers,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].ConfidenceScore != out[j].ConfidenceScore {
			return out[i].ConfidenceScore < out[j].ConfidenceScore
		}
		return out[i].PlaybookID < out[j].PlaybookID
	})
	return out
}

func falsePositiveRisk(pb model.Playbook, sharedPositivePatterns int) string {
	switch {
	case sharedPositivePatterns >= 7:
		return "high"
	case sharedPositivePatterns >= 3:
		return "medium"
	case pb.Category == "silent_failure":
		return "medium"
	default:
		return "low"
	}
}

func falseNegativeRisk(pb model.Playbook, positiveFixtures int) string {
	if pb.Detector == "source" {
		return "medium"
	}
	switch {
	case positiveFixtures <= 1:
		return "high"
	case positiveFixtures == 2:
		return "medium"
	default:
		return "low"
	}
}

func robustnessScore(fpRisk, fnRisk string, positiveFixtures, negativeAssertions, strictTop1 int) int {
	score := 86
	score -= riskPenalty(fpRisk)
	score -= riskPenalty(fnRisk)
	switch {
	case positiveFixtures >= 5:
		score += 6
	case positiveFixtures >= 3:
		score += 3
	}
	if negativeAssertions >= 3 {
		score += 3
	}
	if strictTop1 > 0 {
		score += 2
	}
	if score > 95 {
		return 95
	}
	if score < 45 {
		return 45
	}
	return score
}

func riskPenalty(risk string) int {
	switch risk {
	case "high":
		return 16
	case "medium":
		return 8
	default:
		return 0
	}
}

func robustnessDrivers(pb model.Playbook, positive, negative, strict, shared, exclusions int, fpRisk, fnRisk string) []string {
	drivers := make([]string, 0, 6)
	if pb.Detector == "source" {
		drivers = append(drivers, "source detector coverage is validated outside log fixture accounting")
	}
	if positive == 0 && pb.Detector != "source" {
		drivers = append(drivers, "no positive fixture evidence")
	} else if positive == 1 {
		drivers = append(drivers, "only one positive fixture")
	} else if positive == 2 {
		drivers = append(drivers, "two positive fixtures")
	}
	if strict == 0 {
		drivers = append(drivers, "no strict top-1 fixture")
	}
	if negative == 0 {
		drivers = append(drivers, "no disallowed-playbook assertion")
	}
	if shared >= 3 {
		drivers = append(drivers, fmt.Sprintf("%d shared positive pattern conflicts", shared))
	}
	if fpRisk == "high" || fnRisk == "high" {
		drivers = append(drivers, "high-risk band needs near-miss or noisy fixture promotion")
	}
	if exclusions > 0 {
		drivers = append(drivers, fmt.Sprintf("%d explicit exclusions protect nearby playbooks", exclusions))
	}
	sort.Strings(drivers)
	return drivers
}

func buildCategoryReports(byCategory map[string][]model.Playbook, evidence fixtureEvidence) []Category {
	categories := make([]string, 0, len(byCategory))
	for category := range byCategory {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	out := make([]Category, 0, len(categories))
	for _, category := range categories {
		playbookIDs := make([]string, 0, len(byCategory[category]))
		withFixtures := 0
		positive := 0
		negative := 0
		for _, pb := range byCategory[category] {
			playbookIDs = append(playbookIDs, pb.ID)
			positive += evidence.positiveByPlaybook[pb.ID]
			negative += evidence.negativeByPlaybook[pb.ID]
			if evidence.positiveByPlaybook[pb.ID] > 0 {
				withFixtures++
			}
		}
		sort.Strings(playbookIDs)
		out = append(out, Category{
			Category:           category,
			Count:              len(playbookIDs),
			WithFixtures:       withFixtures,
			PositiveFixtures:   positive,
			NegativeAssertions: negative,
			PlaybookIDs:        playbookIDs,
		})
	}
	return out
}

func buildDomainReports(domainCounts map[string]int) []Domain {
	domains := make([]string, 0, len(domainCounts))
	for domain := range domainCounts {
		domains = append(domains, domain)
	}
	sort.Strings(domains)
	out := make([]Domain, 0, len(domains))
	for _, domain := range domains {
		out = append(out, Domain{Domain: domain, Count: domainCounts[domain]})
	}
	return out
}

func duplicatePlaybookIDs(pbs []model.Playbook) []string {
	counts := map[string]int{}
	for _, pb := range pbs {
		counts[pb.ID]++
	}
	var duplicates []string
	for id, count := range counts {
		if count > 1 {
			duplicates = append(duplicates, fmt.Sprintf("%s (x%d)", id, count))
		}
	}
	sort.Strings(duplicates)
	if duplicates == nil {
		return []string{}
	}
	return duplicates
}

func collectFixtureEvidence(root string) (fixtureEvidence, error) {
	resolved := strings.TrimSpace(root)
	if resolved == "" {
		resolved = defaultFixtureRoot()
	}
	evidence := fixtureEvidence{
		root:                 resolved,
		positiveByPlaybook:   map[string]int{},
		negativeByPlaybook:   map[string]int{},
		strictTop1ByPlaybook: map[string]int{},
	}
	if resolved == "" {
		return evidence, nil
	}

	if layout, ok := fixtureLayoutForRoot(resolved); ok {
		evidence.mode = "fixture-corpus"
		for _, class := range []fixtures.Class{fixtures.ClassMinimal, fixtures.ClassReal, fixtures.ClassNoisy, fixtures.ClassStaging} {
			loaded, err := fixtures.Load(layout, class)
			if err != nil {
				return fixtureEvidence{}, err
			}
			for _, fixture := range loaded {
				evidence.addFixture(fixture)
			}
		}
		return evidence, nil
	}

	evidence.mode = "log-stems"
	err := filepath.WalkDir(resolved, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		if filepath.Ext(name) != ".log" {
			return nil
		}
		id := strings.TrimSuffix(name, filepath.Ext(name))
		evidence.fixtureCount++
		evidence.positiveFixtureCount++
		evidence.positiveByPlaybook[id]++
		return nil
	})
	if err != nil {
		return fixtureEvidence{}, err
	}
	return evidence, nil
}

func (e *fixtureEvidence) addFixture(f fixtures.Fixture) {
	e.fixtureCount++
	if f.Expectation.StrictTop1 {
		e.strictTop1FixtureCount++
	}
	expected := strings.TrimSpace(f.Expectation.ExpectedPlaybook)
	if expected != "" {
		e.positiveFixtureCount++
		e.positiveByPlaybook[expected]++
		if f.Expectation.StrictTop1 {
			e.strictTop1ByPlaybook[expected]++
		}
	}
	for _, disallowed := range f.Expectation.DisallowedPlaybooks {
		disallowed = strings.TrimSpace(disallowed)
		if disallowed == "" {
			continue
		}
		e.negativeAssertionCount++
		e.negativeByPlaybook[disallowed]++
	}
}

func fixtureLayoutForRoot(root string) (fixtures.Layout, bool) {
	clean := filepath.Clean(root)
	if hasFixtureClassDirs(clean) {
		return fixtures.Layout{
			Root:       filepath.Dir(clean),
			Fixtures:   clean,
			MinimalDir: filepath.Join(clean, string(fixtures.ClassMinimal)),
			RealDir:    filepath.Join(clean, string(fixtures.ClassReal)),
			StagingDir: filepath.Join(clean, string(fixtures.ClassStaging)),
			NoisyDir:   filepath.Join(clean, string(fixtures.ClassNoisy)),
		}, true
	}
	child := filepath.Join(clean, "fixtures")
	if hasFixtureClassDirs(child) {
		layout, err := fixtures.ResolveLayout(clean)
		if err != nil {
			return fixtures.Layout{}, false
		}
		return layout, true
	}
	return fixtures.Layout{}, false
}

func hasFixtureClassDirs(root string) bool {
	for _, name := range []string{string(fixtures.ClassMinimal), string(fixtures.ClassReal), string(fixtures.ClassNoisy), string(fixtures.ClassStaging)} {
		if info, err := os.Stat(filepath.Join(root, name)); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

func defaultFixtureRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := cwd
	for {
		candidate := filepath.Join(dir, "fixtures")
		if hasFixtureClassDirs(candidate) {
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
