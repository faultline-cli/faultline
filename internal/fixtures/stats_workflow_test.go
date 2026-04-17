package fixtures

import (
	"strings"
	"testing"
)

func TestDuplicateStatusExactFingerprintMatch(t *testing.T) {
	target := Fixture{
		ID:            "target",
		NormalizedLog: "npm ERR! lockfile mismatch",
		Fingerprint:   "abc123",
	}
	existing := []Fixture{
		{ID: "existing-1", NormalizedLog: "npm ERR! lockfile mismatch", Fingerprint: "abc123"},
	}
	dupID, near, score := duplicateStatus(target, existing)
	if dupID != "existing-1" {
		t.Fatalf("expected exact match with existing-1, got %q", dupID)
	}
	if near != nil {
		t.Fatalf("expected nil near-duplicates for exact match, got %v", near)
	}
	if score != 1.0 {
		t.Fatalf("expected score 1.0 for exact match, got %.2f", score)
	}
}

func TestDuplicateStatusNoMatchReturnsEmpty(t *testing.T) {
	target := Fixture{
		ID:            "target",
		NormalizedLog: "npm ERR! peer dependency conflict",
	}
	existing := []Fixture{
		{ID: "other", NormalizedLog: "docker: pull access denied", Fingerprint: "xyz"},
	}
	dupID, near, _ := duplicateStatus(target, existing)
	if dupID != "" {
		t.Fatalf("expected no exact duplicate, got %q", dupID)
	}
	if len(near) != 0 {
		t.Fatalf("expected no near-duplicates for unrelated content, got %v", near)
	}
}

func TestDuplicateStatusSkipsSelf(t *testing.T) {
	target := Fixture{
		ID:          "self",
		Fingerprint: "samefp",
		RawLog:      "some log",
	}
	existing := []Fixture{
		// Same ID, same fingerprint — should be skipped
		{ID: "self", Fingerprint: "samefp", RawLog: "some log"},
	}
	dupID, near, _ := duplicateStatus(target, existing)
	if dupID != "" {
		t.Fatalf("expected no match against self, got %q", dupID)
	}
	if len(near) != 0 {
		t.Fatalf("expected no near-duplicates from self, got %v", near)
	}
}

func TestBaselineFingerprintIsDeterministic(t *testing.T) {
	r := Report{
		Class:             ClassMinimal,
		FixtureCount:      10,
		Top1Count:         9,
		Top3Count:         10,
		UnmatchedCount:    0,
		FalsePositiveCount: 0,
		RecurringPatterns: map[string]int{"npm-peer": 3},
	}
	fp1 := r.BaselineFingerprint()
	fp2 := r.BaselineFingerprint()
	if fp1 != fp2 {
		t.Fatalf("BaselineFingerprint not deterministic: %q != %q", fp1, fp2)
	}
	if len(fp1) == 0 {
		t.Fatal("expected non-empty fingerprint")
	}
}

func TestBaselineMethodPopulatesAllFields(t *testing.T) {
	r := Report{
		Class:        ClassMinimal,
		FixtureCount: 5,
		Top1Count:    4,
		Top3Count:    5,
	}
	thresholds := Thresholds{MinTop1: 0.8, MinTop3: 0.9}
	baseline := r.Baseline(thresholds)
	if baseline.Class != ClassMinimal {
		t.Errorf("expected class minimal, got %q", baseline.Class)
	}
	if baseline.FixtureCount != 5 {
		t.Errorf("expected 5 fixtures, got %d", baseline.FixtureCount)
	}
	if baseline.Thresholds.MinTop1 != 0.8 {
		t.Errorf("expected min_top1 0.8, got %.2f", baseline.Thresholds.MinTop1)
	}
	if baseline.GeneratedAt == "" {
		t.Error("expected non-empty GeneratedAt")
	}
	if baseline.Fingerprint == "" {
		t.Error("expected non-empty Fingerprint")
	}
}

func TestWriteAndLoadBaselineRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/baseline.json"
	original := Baseline{
		Class:        ClassMinimal,
		FixtureCount: 42,
		Top1Rate:     0.95,
		Thresholds:   Thresholds{MinTop1: 0.9},
		Fingerprint:  "abc123",
		GeneratedAt:  "2026-01-01T00:00:00Z",
	}
	if err := WriteBaseline(path, original); err != nil {
		t.Fatalf("WriteBaseline: %v", err)
	}
	loaded, err := LoadBaseline(path)
	if err != nil {
		t.Fatalf("LoadBaseline: %v", err)
	}
	if loaded.FixtureCount != original.FixtureCount {
		t.Errorf("FixtureCount: got %d, want %d", loaded.FixtureCount, original.FixtureCount)
	}
	if loaded.Thresholds.MinTop1 != original.Thresholds.MinTop1 {
		t.Errorf("MinTop1: got %.2f, want %.2f", loaded.Thresholds.MinTop1, original.Thresholds.MinTop1)
	}
	if loaded.Fingerprint != original.Fingerprint {
		t.Errorf("Fingerprint: got %q, want %q", loaded.Fingerprint, original.Fingerprint)
	}
}

func TestLoadBaselineErrorOnMissing(t *testing.T) {
	if _, err := LoadBaseline("/tmp/nonexistent_baseline_abc.json"); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestHTTPErrorString(t *testing.T) {
	e := &httpStatusError{StatusCode: 404, Status: "404 Not Found", URL: "https://example.com/api/resource", Body: "not found"}
	msg := e.Error()
	if !strings.Contains(msg, "404") {
		t.Errorf("expected 404 in error message, got %q", msg)
	}
}

func TestEffectiveClassFromExplicitField(t *testing.T) {
	f := Fixture{FixtureClass: ClassReal}
	if f.effectiveClass() != ClassReal {
		t.Errorf("expected ClassReal from explicit field, got %q", f.effectiveClass())
	}
}

func TestEffectiveClassFromFilePath(t *testing.T) {
	cases := []struct {
		path string
		want Class
	}{
		{"/repo/fixtures/real/test.yaml", ClassReal},
		{"/repo/fixtures/staging/test.yaml", ClassStaging},
		{"/repo/fixtures/minimal/test.yaml", ClassMinimal},
		{"/some/other/path.yaml", ClassMinimal},
	}
	for _, tc := range cases {
		f := Fixture{FilePath: tc.path}
		if got := f.effectiveClass(); got != tc.want {
			t.Errorf("effectiveClass(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestAllowedRankUsesExpectationTopN(t *testing.T) {
	f := Fixture{Expectation: Expectation{TopN: 5}}
	if got := f.allowedRank(); got != 5 {
		t.Errorf("expected allowedRank = 5, got %d", got)
	}
}

func TestAllowedRankDefaults(t *testing.T) {
	minimal := Fixture{FixtureClass: ClassMinimal}
	if got := minimal.allowedRank(); got != 1 {
		t.Errorf("expected minimal allowedRank = 1, got %d", got)
	}
	real := Fixture{FixtureClass: ClassReal}
	if got := real.allowedRank(); got != 3 {
		t.Errorf("expected real allowedRank = 3, got %d", got)
	}
}

func TestIsStrictTop1(t *testing.T) {
	strictExplicit := Fixture{Expectation: Expectation{StrictTop1: true}, FixtureClass: ClassReal}
	if !strictExplicit.isStrictTop1() {
		t.Error("expected isStrictTop1 = true when StrictTop1 explicitly set")
	}
	minimalImplicit := Fixture{FixtureClass: ClassMinimal}
	if !minimalImplicit.isStrictTop1() {
		t.Error("expected isStrictTop1 = true for minimal class")
	}
	realNotStrict := Fixture{FixtureClass: ClassReal}
	if realNotStrict.isStrictTop1() {
		t.Error("expected isStrictTop1 = false for real class without explicit flag")
	}
}

func TestConfidenceFloor(t *testing.T) {
	withExplicit := Fixture{Expectation: Expectation{MinConfidence: 0.8}}
	if got := withExplicit.confidenceFloor(); got != 0.8 {
		t.Errorf("expected 0.8, got %.2f", got)
	}
	real := Fixture{FixtureClass: ClassReal}
	if got := real.confidenceFloor(); got != 0.55 {
		t.Errorf("expected 0.55 for real class, got %.2f", got)
	}
	minimal := Fixture{FixtureClass: ClassMinimal}
	if got := minimal.confidenceFloor(); got != 0 {
		t.Errorf("expected 0 for minimal class, got %.2f", got)
	}
}
