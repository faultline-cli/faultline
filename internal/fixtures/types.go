package fixtures

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type Class string

const (
	ClassMinimal Class = "minimal"
	ClassReal    Class = "real"
	ClassStaging Class = "staging"
	ClassAll     Class = "all"
)

func ParseClass(value string) (Class, error) {
	switch Class(strings.TrimSpace(strings.ToLower(value))) {
	case "", ClassAll:
		return ClassAll, nil
	case ClassMinimal:
		return ClassMinimal, nil
	case ClassReal:
		return ClassReal, nil
	case ClassStaging:
		return ClassStaging, nil
	default:
		return "", fmt.Errorf("invalid fixture class %q", value)
	}
}

type Layout struct {
	Root       string
	Fixtures   string
	MinimalDir string
	RealDir    string
	StagingDir string
}

func ResolveLayout(root string) (Layout, error) {
	if strings.TrimSpace(root) == "" {
		root = "."
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Layout{}, fmt.Errorf("resolve repository root: %w", err)
	}
	fixturesRoot := filepath.Join(absRoot, "fixtures")
	return Layout{
		Root:       absRoot,
		Fixtures:   fixturesRoot,
		MinimalDir: filepath.Join(fixturesRoot, string(ClassMinimal)),
		RealDir:    filepath.Join(fixturesRoot, string(ClassReal)),
		StagingDir: filepath.Join(fixturesRoot, string(ClassStaging)),
	}, nil
}

type Expectation struct {
	ExpectedPlaybook    string   `yaml:"expected_playbook,omitempty" json:"expected_playbook,omitempty"`
	TopN                int      `yaml:"top_n,omitempty" json:"top_n,omitempty"`
	ExpectedStage       string   `yaml:"expected_stage,omitempty" json:"expected_stage,omitempty"`
	StrictTop1          bool     `yaml:"strict_top1,omitempty" json:"strict_top1,omitempty"`
	DisallowedPlaybooks []string `yaml:"disallowed_playbooks,omitempty" json:"disallowed_playbooks,omitempty"`
	MaxUnexpectedTopN   int      `yaml:"max_unexpected_top_n,omitempty" json:"max_unexpected_top_n,omitempty"`
	MinConfidence       float64  `yaml:"min_confidence,omitempty" json:"min_confidence,omitempty"`
}

type SourceMetadata struct {
	Adapter     string   `yaml:"adapter,omitempty" json:"adapter,omitempty"`
	Provider    string   `yaml:"provider,omitempty" json:"provider,omitempty"`
	URL         string   `yaml:"url,omitempty" json:"url,omitempty"`
	Repository  string   `yaml:"repository,omitempty" json:"repository,omitempty"`
	IssueNumber int      `yaml:"issue_number,omitempty" json:"issue_number,omitempty"`
	CommentID   string   `yaml:"comment_id,omitempty" json:"comment_id,omitempty"`
	Title       string   `yaml:"title,omitempty" json:"title,omitempty"`
	Labels      []string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Author      string   `yaml:"author,omitempty" json:"author,omitempty"`
	Snippet     int      `yaml:"snippet,omitempty" json:"snippet,omitempty"`
	FetchedAt   string   `yaml:"fetched_at,omitempty" json:"fetched_at,omitempty"`
}

type ReviewMetadata struct {
	Status         string   `yaml:"status,omitempty" json:"status,omitempty"`
	Notes          string   `yaml:"notes,omitempty" json:"notes,omitempty"`
	DuplicateOf    string   `yaml:"duplicate_of,omitempty" json:"duplicate_of,omitempty"`
	NearDuplicates []string `yaml:"near_duplicates,omitempty" json:"near_duplicates,omitempty"`
	PromotedAt     string   `yaml:"promoted_at,omitempty" json:"promoted_at,omitempty"`
}

type Fixture struct {
	ID            string         `yaml:"id" json:"id"`
	Title         string         `yaml:"title,omitempty" json:"title,omitempty"`
	FixtureClass  Class          `yaml:"fixture_class,omitempty" json:"fixture_class,omitempty"`
	Path          string         `yaml:"path,omitempty" json:"path,omitempty"`
	RawLog        string         `yaml:"raw_log,omitempty" json:"raw_log,omitempty"`
	NormalizedLog string         `yaml:"normalized_log,omitempty" json:"normalized_log,omitempty"`
	Expectation   Expectation    `yaml:"expectation,omitempty" json:"expectation,omitempty"`
	Source        SourceMetadata `yaml:"source,omitempty" json:"source,omitempty"`
	Review        ReviewMetadata `yaml:"review,omitempty" json:"review,omitempty"`
	Tags          []string       `yaml:"tags,omitempty" json:"tags,omitempty"`
	Fingerprint   string         `yaml:"fingerprint,omitempty" json:"fingerprint,omitempty"`
	FilePath      string         `yaml:"-" json:"-"`
	manifestRoot  string
}

func (f Fixture) effectiveClass() Class {
	if f.FixtureClass != "" {
		return f.FixtureClass
	}
	if strings.Contains(filepath.ToSlash(f.FilePath), "/fixtures/real/") {
		return ClassReal
	}
	if strings.Contains(filepath.ToSlash(f.FilePath), "/fixtures/staging/") {
		return ClassStaging
	}
	return ClassMinimal
}

func (f Fixture) allowedRank() int {
	if f.Expectation.TopN > 0 {
		return f.Expectation.TopN
	}
	if f.effectiveClass() == ClassMinimal {
		return 1
	}
	return 3
}

func (f Fixture) isStrictTop1() bool {
	if f.Expectation.StrictTop1 {
		return true
	}
	return f.effectiveClass() == ClassMinimal
}

func (f Fixture) confidenceFloor() float64 {
	if f.Expectation.MinConfidence > 0 {
		return f.Expectation.MinConfidence
	}
	if f.effectiveClass() == ClassReal {
		return 0.55
	}
	return 0
}

func FingerprintForLog(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:8])
}

type manifestFile struct {
	Fixtures []Fixture `yaml:"fixtures"`
}

type Baseline struct {
	Class             Class      `json:"class"`
	FixtureCount      int        `json:"fixture_count"`
	Top1Rate          float64    `json:"top_1_rate"`
	Top3Rate          float64    `json:"top_3_rate"`
	UnmatchedRate     float64    `json:"unmatched_rate"`
	FalsePositiveRate float64    `json:"false_positive_rate"`
	WeakMatchRate     float64    `json:"weak_match_rate"`
	Thresholds        Thresholds `json:"thresholds"`
	GeneratedAt       string     `json:"generated_at"`
	Fingerprint       string     `json:"fingerprint"`
}

type Thresholds struct {
	MinTop1          float64 `json:"min_top_1"`
	MinTop3          float64 `json:"min_top_3"`
	MaxUnmatched     float64 `json:"max_unmatched"`
	MaxFalsePositive float64 `json:"max_false_positive"`
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
