package fixtures

import (
	"path/filepath"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}

func bundledPlaybookDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "playbooks", "bundled")
}

func TestMinimalFixtureCorpusIsStrict(t *testing.T) {
	layout, err := ResolveLayout(repoRoot(t))
	if err != nil {
		t.Fatalf("resolve layout: %v", err)
	}
	report, err := Evaluate(layout, ClassMinimal, EvaluateOptions{
		PlaybookDir: bundledPlaybookDir(t),
		NoHistory:   true,
	})
	if err != nil {
		t.Fatalf("evaluate minimal corpus: %v", err)
	}
	if report.FixtureCount < 40 {
		t.Fatalf("expected broad minimal corpus, got %d fixtures", report.FixtureCount)
	}
	if report.Top1Count != report.FixtureCount {
		t.Fatalf("expected exact minimal top-1 hits, got %d/%d", report.Top1Count, report.FixtureCount)
	}
	if report.UnmatchedCount != 0 {
		t.Fatalf("expected no unmatched minimal fixtures, got %v", report.UnmatchedFixtureIDs)
	}
	if report.FalsePositiveCount != 0 {
		t.Fatalf("expected no minimal false positives, got %d", report.FalsePositiveCount)
	}
}

func TestRealFixtureCorpusBaseline(t *testing.T) {
	layout, err := ResolveLayout(repoRoot(t))
	if err != nil {
		t.Fatalf("resolve layout: %v", err)
	}
	report, err := Evaluate(layout, ClassReal, EvaluateOptions{
		PlaybookDir: bundledPlaybookDir(t),
		NoHistory:   true,
	})
	if err != nil {
		t.Fatalf("evaluate real corpus: %v", err)
	}
	if report.FixtureCount < 25 {
		t.Fatalf("expected at least 25 curated real fixtures, got %d", report.FixtureCount)
	}
	baseline, err := LoadBaseline(filepath.Join(layout.RealDir, "baseline.json"))
	if err != nil {
		t.Fatalf("load real corpus baseline: %v", err)
	}
	if err := CheckBaseline(&report, baseline); err != nil {
		t.Fatalf("real corpus regression: %v", err)
	}
}
