package compare

import (
	"testing"

	"faultline/internal/model"
)

func makeAnalysis(id, title string, confidence float64, evidence []string) *model.Analysis {
	return &model.Analysis{
		Source: id + ".json",
		Results: []model.Result{
			{
				Playbook: model.Playbook{
					ID:    id,
					Title: title,
				},
				Confidence: confidence,
				Score:      confidence,
				Evidence:   evidence,
			},
		},
	}
}

// ── Build ─────────────────────────────────────────────────────────────────────

func TestBuildSameDiagnosis(t *testing.T) {
	left := makeAnalysis("docker-auth", "Docker auth", 0.9, []string{"authentication required"})
	right := makeAnalysis("docker-auth", "Docker auth", 0.85, []string{"authentication required"})
	report := Build(left, right)
	if report.DiagnosisChanged {
		t.Error("expected DiagnosisChanged=false for same failure ID")
	}
	if report.Changed {
		t.Error("expected Changed=false when nothing differs")
	}
	if report.Previous == nil || report.Previous.FailureID != "docker-auth" {
		t.Errorf("expected Previous.FailureID=docker-auth, got %v", report.Previous)
	}
	if report.Current == nil || report.Current.FailureID != "docker-auth" {
		t.Errorf("expected Current.FailureID=docker-auth, got %v", report.Current)
	}
}

func TestBuildDiagnosisChanged(t *testing.T) {
	left := makeAnalysis("docker-auth", "Docker auth", 0.9, []string{"authentication required"})
	right := makeAnalysis("permission-denied", "Permission denied", 0.75, []string{"permission denied"})
	report := Build(left, right)
	if !report.DiagnosisChanged {
		t.Error("expected DiagnosisChanged=true")
	}
	if !report.Changed {
		t.Error("expected Changed=true when diagnosis changed")
	}
	if len(report.Summary) == 0 {
		t.Error("expected non-empty Summary")
	}
	found := false
	for _, s := range report.Summary {
		if s == "top diagnosis changed from docker-auth to permission-denied" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected diagnosis change line in summary, got %v", report.Summary)
	}
}

func TestBuildNilLeft(t *testing.T) {
	right := makeAnalysis("docker-auth", "Docker auth", 0.9, []string{"auth error"})
	report := Build(nil, right)
	if !report.DiagnosisChanged {
		t.Error("expected DiagnosisChanged=true when left is nil")
	}
	if report.Previous != nil {
		t.Error("expected Previous=nil")
	}
	if report.Current == nil {
		t.Error("expected Current non-nil")
	}
	if len(report.Summary) == 0 {
		t.Error("expected non-empty Summary")
	}
}

func TestBuildNilRight(t *testing.T) {
	left := makeAnalysis("docker-auth", "Docker auth", 0.9, []string{"auth error"})
	report := Build(left, nil)
	if !report.DiagnosisChanged {
		t.Error("expected DiagnosisChanged=true when right is nil")
	}
	if report.Current != nil {
		t.Error("expected Current=nil")
	}
}

func TestBuildBothNil(t *testing.T) {
	report := Build(nil, nil)
	if report.DiagnosisChanged {
		t.Error("expected DiagnosisChanged=false when both nil")
	}
	if report.Changed {
		t.Error("expected Changed=false when both nil")
	}
}

func TestBuildEvidenceDelta(t *testing.T) {
	left := makeAnalysis("docker-auth", "Docker auth", 0.9, []string{"authentication required", "pull access denied"})
	right := makeAnalysis("docker-auth", "Docker auth", 0.9, []string{"authentication required", "unauthorized"})
	report := Build(left, right)
	if len(report.Evidence.Added) != 1 || report.Evidence.Added[0] != "unauthorized" {
		t.Errorf("expected Added=[unauthorized], got %v", report.Evidence.Added)
	}
	if len(report.Evidence.Removed) != 1 || report.Evidence.Removed[0] != "pull access denied" {
		t.Errorf("expected Removed=[pull access denied], got %v", report.Evidence.Removed)
	}
	if !report.Changed {
		t.Error("expected Changed=true with evidence delta")
	}
}

func TestBuildRepoContextDelta(t *testing.T) {
	left := makeAnalysis("docker-auth", "Docker auth", 0.9, nil)
	left.RepoContext = &model.RepoContext{RecentFiles: []string{"Dockerfile", "compose.yml"}}
	right := makeAnalysis("docker-auth", "Docker auth", 0.9, nil)
	right.RepoContext = &model.RepoContext{RecentFiles: []string{"Dockerfile", ".env"}}
	report := Build(left, right)
	if len(report.RepoFiles.Added) != 1 || report.RepoFiles.Added[0] != ".env" {
		t.Errorf("expected RepoFiles.Added=[.env], got %v", report.RepoFiles.Added)
	}
	if len(report.RepoFiles.Removed) != 1 || report.RepoFiles.Removed[0] != "compose.yml" {
		t.Errorf("expected RepoFiles.Removed=[compose.yml], got %v", report.RepoFiles.Removed)
	}
}

func TestBuildDeltaFilesDelta(t *testing.T) {
	left := makeAnalysis("go-sum", "Missing go.sum", 0.8, nil)
	left.Delta = &model.Delta{FilesChanged: []string{"go.sum"}}
	right := makeAnalysis("go-sum", "Missing go.sum", 0.8, nil)
	right.Delta = &model.Delta{FilesChanged: []string{"go.sum", "go.mod"}}
	report := Build(left, right)
	if len(report.DeltaFiles.Added) != 1 || report.DeltaFiles.Added[0] != "go.mod" {
		t.Errorf("expected DeltaFiles.Added=[go.mod], got %v", report.DeltaFiles.Added)
	}
}

func TestBuildDeltaTestsDelta(t *testing.T) {
	left := makeAnalysis("test-fail", "Test failure", 0.8, nil)
	left.Delta = &model.Delta{TestsNewlyFailing: []string{"TestA"}}
	right := makeAnalysis("test-fail", "Test failure", 0.8, nil)
	right.Delta = &model.Delta{TestsNewlyFailing: []string{"TestA", "TestB"}}
	report := Build(left, right)
	if len(report.DeltaTests.Added) != 1 || report.DeltaTests.Added[0] != "TestB" {
		t.Errorf("expected DeltaTests.Added=[TestB], got %v", report.DeltaTests.Added)
	}
	if !report.Changed {
		t.Error("expected Changed=true with delta tests")
	}
}

func TestBuildDeltaErrorsDelta(t *testing.T) {
	left := makeAnalysis("test-fail", "Test failure", 0.8, nil)
	left.Delta = &model.Delta{ErrorsAdded: []string{"ERR1"}}
	right := makeAnalysis("test-fail", "Test failure", 0.8, nil)
	right.Delta = &model.Delta{ErrorsAdded: []string{"ERR1", "ERR2"}}
	report := Build(left, right)
	if len(report.DeltaErrors.Added) != 1 || report.DeltaErrors.Added[0] != "ERR2" {
		t.Errorf("expected DeltaErrors.Added=[ERR2], got %v", report.DeltaErrors.Added)
	}
}

func TestBuildSourcesPopulated(t *testing.T) {
	left := makeAnalysis("docker-auth", "Docker auth", 0.9, nil)
	right := makeAnalysis("docker-auth", "Docker auth", 0.85, nil)
	report := Build(left, right)
	if report.LeftSource != "docker-auth.json" {
		t.Errorf("expected LeftSource=docker-auth.json, got %q", report.LeftSource)
	}
	if report.RightSource != "docker-auth.json" {
		t.Errorf("expected RightSource=docker-auth.json, got %q", report.RightSource)
	}
}

func TestBuildSummaryNoMaterialDifference(t *testing.T) {
	left := makeAnalysis("docker-auth", "Docker auth", 0.9, []string{"same evidence"})
	right := makeAnalysis("docker-auth", "Docker auth", 0.85, []string{"same evidence"})
	report := Build(left, right)
	// Should contain "no material differences" line
	found := false
	for _, s := range report.Summary {
		if s == "no material differences were found in the compared artifacts" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected no-material-differences line in summary, got %v", report.Summary)
	}
}

func TestBuildSummaryNewContextAppeared(t *testing.T) {
	left := makeAnalysis("docker-auth", "Docker auth", 0.9, nil)
	right := makeAnalysis("docker-auth", "Docker auth", 0.85, nil)
	right.Delta = &model.Delta{FilesChanged: []string{"src/main.go"}}
	report := Build(left, right)
	found := false
	for _, s := range report.Summary {
		if s == "current delta context contains new changed files, failing tests, or added errors" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected delta context line in summary, got %v", report.Summary)
	}
}

// ── diffStrings ───────────────────────────────────────────────────────────────

func TestDiffStringsEmptyInputs(t *testing.T) {
	d := diffStrings(nil, nil)
	if len(d.Added) != 0 || len(d.Removed) != 0 {
		t.Errorf("expected empty delta, got %+v", d)
	}
}

func TestDiffStringsSkipsEmptyStrings(t *testing.T) {
	d := diffStrings([]string{""}, []string{""})
	if len(d.Added) != 0 || len(d.Removed) != 0 {
		t.Errorf("expected empty delta filtering blanks, got %+v", d)
	}
}

func TestDiffStringsSorted(t *testing.T) {
	d := diffStrings([]string{"b", "a"}, []string{"c", "b", "d"})
	if len(d.Added) != 2 || d.Added[0] != "c" || d.Added[1] != "d" {
		t.Errorf("expected sorted Added=[c,d], got %v", d.Added)
	}
	if len(d.Removed) != 1 || d.Removed[0] != "a" {
		t.Errorf("expected Removed=[a], got %v", d.Removed)
	}
}

// ── hasDelta ──────────────────────────────────────────────────────────────────

func TestHasDelta(t *testing.T) {
	if hasDelta(StringDelta{}) {
		t.Error("expected hasDelta=false for empty delta")
	}
	if !hasDelta(StringDelta{Added: []string{"x"}}) {
		t.Error("expected hasDelta=true for non-empty Added")
	}
	if !hasDelta(StringDelta{Removed: []string{"x"}}) {
		t.Error("expected hasDelta=true for non-empty Removed")
	}
}
