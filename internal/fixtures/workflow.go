package fixtures

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type IngestOptions struct {
	Adapter string
	URLs    []string
	Force   bool
	Now     time.Time
	Client  *http.Client
}

type IngestResult struct {
	Written []Fixture
	Skipped []string
}

type ReviewItem struct {
	Fixture        Fixture
	PredictedTopID string
	PredictedTop3  []string
	DuplicateOf    string
	NearDuplicates []string
	Similarity     float64
	Status         string
}

type ReviewReport struct {
	Items []ReviewItem
}

type PromoteOptions struct {
	ExpectedPlaybook    string
	TopN                int
	ExpectedStage       string
	StrictTop1          bool
	DisallowedPlaybooks []string
	MinConfidence       float64
	KeepStaging         bool
	PromotedAt          time.Time
}

func Ingest(ctx context.Context, layout Layout, opts IngestOptions) (IngestResult, error) {
	adapter, err := adapterByName(opts.Adapter)
	if err != nil {
		return IngestResult{}, err
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}
	if err := os.MkdirAll(layout.StagingDir, 0o755); err != nil {
		return IngestResult{}, fmt.Errorf("create staging directory: %w", err)
	}
	existingReal, err := Load(layout, ClassReal)
	if err != nil {
		return IngestResult{}, err
	}
	existingStaging, err := Load(layout, ClassStaging)
	if err != nil {
		return IngestResult{}, err
	}
	existing := mergeFixtures([][]Fixture{existingReal, existingStaging})
	seenFingerprints := map[string]string{}
	for _, fixture := range existing {
		seenFingerprints[fixture.Fingerprint] = fixture.ID
	}

	result := IngestResult{}
	for _, rawURL := range opts.URLs {
		if !adapter.Supports(rawURL) {
			result.Skipped = append(result.Skipped, fmt.Sprintf("%s: unsupported URL for %s", rawURL, opts.Adapter))
			continue
		}
		fetched, err := adapter.Fetch(ctx, rawURL, opts.Client, opts.Now)
		if err != nil {
			return result, err
		}
		for _, fixture := range fetched {
			fixture.FixtureClass = ClassStaging
			fixture.Review.Status = "ingested"
			if fixture.Fingerprint == "" {
				fixture.Fingerprint = FingerprintForLog(fixture.NormalizedLog)
			}
			if duplicateID, ok := seenFingerprints[fixture.Fingerprint]; ok && !opts.Force {
				result.Skipped = append(result.Skipped, fmt.Sprintf("%s: duplicate of %s", fixture.ID, duplicateID))
				continue
			}
			if err := writeFixture(filepath.Join(layout.StagingDir, fixture.ID+".yaml"), fixture); err != nil {
				return result, err
			}
			seenFingerprints[fixture.Fingerprint] = fixture.ID
			result.Written = append(result.Written, fixture)
		}
	}
	sort.Strings(result.Skipped)
	sort.Slice(result.Written, func(i, j int) bool {
		return result.Written[i].ID < result.Written[j].ID
	})
	return result, nil
}

func Review(layout Layout, opts EvaluateOptions) (ReviewReport, error) {
	staging, err := Load(layout, ClassStaging)
	if err != nil {
		return ReviewReport{}, err
	}
	realFixtures, err := Load(layout, ClassReal)
	if err != nil {
		return ReviewReport{}, err
	}
	report, err := EvaluateFixtures(layout, ClassStaging, staging, opts)
	if err != nil {
		return ReviewReport{}, err
	}
	allExisting := mergeFixtures([][]Fixture{realFixtures, staging})
	items := make([]ReviewItem, 0, len(report.Fixtures))
	for _, evaluated := range report.Fixtures {
		item := ReviewItem{Fixture: evaluated.Fixture}
		if len(evaluated.PredictedTopIDs) > 0 {
			item.PredictedTopID = evaluated.PredictedTopIDs[0]
			item.PredictedTop3 = append(item.PredictedTop3, evaluated.PredictedTopIDs[:min(3, len(evaluated.PredictedTopIDs))]...)
		}
		duplicateOf, nearDuplicates, similarity := duplicateStatus(evaluated.Fixture, allExisting)
		item.DuplicateOf = duplicateOf
		item.NearDuplicates = nearDuplicates
		item.Similarity = similarity
		switch {
		case duplicateOf != "":
			item.Status = "duplicate"
		case len(nearDuplicates) > 0:
			item.Status = "near-duplicate"
		default:
			item.Status = "candidate"
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Fixture.ID < items[j].Fixture.ID
	})
	return ReviewReport{Items: items}, nil
}

func Promote(layout Layout, ids []string, opts PromoteOptions) ([]Fixture, error) {
	staging, err := Load(layout, ClassStaging)
	if err != nil {
		return nil, err
	}
	index := map[string]Fixture{}
	for _, fixture := range staging {
		index[fixture.ID] = fixture
	}
	if opts.PromotedAt.IsZero() {
		opts.PromotedAt = time.Now().UTC()
	}
	var promoted []Fixture
	for _, id := range ids {
		fixture, ok := index[id]
		if !ok {
			return nil, fmt.Errorf("staging fixture %s not found", id)
		}
		fixture.FixtureClass = ClassReal
		fixture.Expectation.ExpectedPlaybook = opts.ExpectedPlaybook
		if opts.TopN > 0 {
			fixture.Expectation.TopN = opts.TopN
		}
		fixture.Expectation.ExpectedStage = opts.ExpectedStage
		fixture.Expectation.StrictTop1 = opts.StrictTop1
		fixture.Expectation.DisallowedPlaybooks = append([]string(nil), opts.DisallowedPlaybooks...)
		fixture.Expectation.MinConfidence = opts.MinConfidence
		fixture.Review.Status = "promoted"
		fixture.Review.PromotedAt = opts.PromotedAt.Format(time.RFC3339)
		fixture.FilePath = filepath.Join(layout.RealDir, fixture.ID+".yaml")
		if err := writeFixture(fixture.FilePath, fixture); err != nil {
			return nil, err
		}
		if !opts.KeepStaging {
			_ = os.Remove(index[id].FilePath)
		}
		promoted = append(promoted, fixture)
	}
	sort.Slice(promoted, func(i, j int) bool {
		return promoted[i].ID < promoted[j].ID
	})
	return promoted, nil
}

func duplicateStatus(target Fixture, existing []Fixture) (string, []string, float64) {
	targetText := target.NormalizedLog
	if targetText == "" {
		targetText = target.RawLog
	}
	targetSig := lineSignature(targetText)
	bestScore := 0.0
	var near []string
	for _, candidate := range existing {
		if candidate.ID == target.ID {
			continue
		}
		if candidate.Fingerprint == target.Fingerprint {
			return candidate.ID, nil, 1
		}
		score := jaccardSimilarity(targetSig, lineSignature(firstNonEmpty(candidate.NormalizedLog, candidate.RawLog)))
		if score >= 0.82 {
			near = append(near, candidate.ID)
		}
		if score > bestScore {
			bestScore = score
		}
	}
	sort.Strings(near)
	return "", near, bestScore
}

func lineSignature(text string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, line := range strings.Split(strings.ToLower(text), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		set[line] = struct{}{}
	}
	return set
}

func jaccardSimilarity(left, right map[string]struct{}) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	intersection := 0
	union := len(left)
	for key := range right {
		if _, ok := left[key]; ok {
			intersection++
		} else {
			union++
		}
	}
	return float64(intersection) / float64(union)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func writeFixture(path string, fixture Fixture) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create fixture directory: %w", err)
	}
	fixture.FilePath = ""
	fixture.manifestRoot = ""
	data, err := yaml.Marshal(fixture)
	if err != nil {
		return fmt.Errorf("marshal fixture %s: %w", fixture.ID, err)
	}
	return os.WriteFile(path, data, 0o644)
}
