package artifact

import (
	"testing"

	"faultline/internal/model"
)

// ── statusForAnalysis ─────────────────────────────────────────────────────────

func TestStatusForAnalysisNil(t *testing.T) {
	got := statusForAnalysis(nil)
	if got != model.ArtifactStatusUnknown {
		t.Errorf("expected ArtifactStatusUnknown for nil analysis, got %q", got)
	}
}

func TestStatusForAnalysisNoResults(t *testing.T) {
	a := &model.Analysis{Results: nil}
	got := statusForAnalysis(a)
	if got != model.ArtifactStatusUnknown {
		t.Errorf("expected ArtifactStatusUnknown for empty results, got %q", got)
	}
}

func TestStatusForAnalysisWithResults(t *testing.T) {
	a := &model.Analysis{
		Results: []model.Result{
			{Playbook: model.Playbook{ID: "docker-auth"}, Score: 0.9},
		},
	}
	got := statusForAnalysis(a)
	if got != model.ArtifactStatusMatched {
		t.Errorf("expected ArtifactStatusMatched, got %q", got)
	}
}

// ── refinedConfidence ─────────────────────────────────────────────────────────

func TestRefinedConfidenceBaseOnly(t *testing.T) {
	a := &model.Analysis{}
	result := model.Result{Confidence: 0.75}
	got := refinedConfidence(a, result)
	if got != 0.75 {
		t.Errorf("expected 0.75, got %f", got)
	}
}

func TestRefinedConfidenceWithRepoContext(t *testing.T) {
	a := &model.Analysis{
		RepoContext: &model.RepoContext{
			RelatedCommits: []model.RepoCommit{{Hash: "abc123", Subject: "fix thing"}},
		},
	}
	result := model.Result{Confidence: 0.75}
	got := refinedConfidence(a, result)
	want := 0.75 + 0.04
	if got != want {
		t.Errorf("expected %f with repo context boost, got %f", want, got)
	}
}

func TestRefinedConfidenceWithDelta(t *testing.T) {
	a := &model.Analysis{
		Delta: &model.Delta{
			Causes: []model.DeltaCause{{Kind: "dep-bump", Score: 1.0}},
		},
	}
	result := model.Result{Confidence: 0.75}
	got := refinedConfidence(a, result)
	want := 0.75 + 0.03
	if got != want {
		t.Errorf("expected %f with delta boost, got %f", want, got)
	}
}

func TestRefinedConfidenceWithBothBoosts(t *testing.T) {
	a := &model.Analysis{
		RepoContext: &model.RepoContext{
			RelatedCommits: []model.RepoCommit{{Hash: "abc"}},
		},
		Delta: &model.Delta{
			Signals: []model.DeltaSignal{{ID: "dep-change"}},
		},
	}
	result := model.Result{Confidence: 0.75}
	got := refinedConfidence(a, result)
	// Both boosts applied: +0.04 (repo) + 0.03 (delta) = 0.82.
	// Use epsilon comparison for floating-point safety.
	const epsilon = 1e-9
	want := 0.75 + 0.04 + 0.03
	if got < want-epsilon || got > want+epsilon {
		t.Errorf("expected ~%f with both boosts, got %f", want, got)
	}
}

func TestRefinedConfidenceCapsAt1(t *testing.T) {
	a := &model.Analysis{
		RepoContext: &model.RepoContext{
			RelatedCommits: []model.RepoCommit{{Hash: "abc"}},
		},
		Delta: &model.Delta{
			Causes: []model.DeltaCause{{Kind: "x", Score: 1.0}},
		},
	}
	result := model.Result{Confidence: 0.99}
	got := refinedConfidence(a, result)
	if got > 1.0 {
		t.Errorf("expected confidence capped at 1.0, got %f", got)
	}
	if got != 1.0 {
		t.Errorf("expected 1.0 after cap, got %f", got)
	}
}

func TestRefinedConfidenceNilAnalysis(t *testing.T) {
	result := model.Result{Confidence: 0.5}
	got := refinedConfidence(nil, result)
	if got != 0.5 {
		t.Errorf("expected 0.5 for nil analysis, got %f", got)
	}
}

// ── Build ─────────────────────────────────────────────────────────────────────

func TestBuildNilAnalysis(t *testing.T) {
	got := Build(nil)
	if got != nil {
		t.Errorf("expected nil for nil analysis, got %v", got)
	}
}

func TestBuildNoResults(t *testing.T) {
	a := &model.Analysis{
		Source:      "test.log",
		Fingerprint: "fp-xyz",
	}
	got := Build(a)
	if got == nil {
		t.Fatal("expected non-nil artifact")
	}
	if got.Status != model.ArtifactStatusUnknown {
		t.Errorf("expected ArtifactStatusUnknown, got %q", got.Status)
	}
	if got.SchemaVersion != SchemaVersion {
		t.Errorf("expected schema version %q, got %q", SchemaVersion, got.SchemaVersion)
	}
	if got.MatchedPlaybook != nil {
		t.Errorf("expected nil MatchedPlaybook for no-result analysis, got %v", got.MatchedPlaybook)
	}
}

func TestBuildWithResult(t *testing.T) {
	a := &model.Analysis{
		Source:      "ci.log",
		Fingerprint: "fp-123",
		Results: []model.Result{
			{
				Playbook:   model.Playbook{ID: "docker-auth", Title: "Docker auth", Category: "auth"},
				Score:      0.9,
				Confidence: 0.85,
				Detector:   "log",
				Evidence:   []string{"authentication required"},
			},
		},
	}
	got := Build(a)
	if got == nil {
		t.Fatal("expected non-nil artifact")
	}
	if got.Status != model.ArtifactStatusMatched {
		t.Errorf("expected ArtifactStatusMatched, got %q", got.Status)
	}
	if got.MatchedPlaybook == nil {
		t.Fatal("expected non-nil MatchedPlaybook")
	}
	if got.MatchedPlaybook.ID != "docker-auth" {
		t.Errorf("expected ID 'docker-auth', got %q", got.MatchedPlaybook.ID)
	}
	if got.Fingerprint != "fp-123" {
		t.Errorf("expected fingerprint 'fp-123', got %q", got.Fingerprint)
	}
	if len(got.Evidence) == 0 {
		t.Error("expected non-empty Evidence")
	}
	if got.Confidence != 0.85 {
		t.Errorf("expected confidence 0.85, got %f", got.Confidence)
	}
}

func TestBuildFingerprint(t *testing.T) {
	a := &model.Analysis{
		Fingerprint: "  padded-fp  ",
	}
	got := Build(a)
	if got.Fingerprint != "padded-fp" {
		t.Errorf("expected trimmed fingerprint 'padded-fp', got %q", got.Fingerprint)
	}
}

// ── Sync ──────────────────────────────────────────────────────────────────────

func TestSyncNilAnalysis(t *testing.T) {
	got := Sync(nil)
	if got != nil {
		t.Errorf("expected nil for nil analysis, got %v", got)
	}
}

func TestSyncSetsStatusAndArtifact(t *testing.T) {
	a := &model.Analysis{
		Source:      "ci.log",
		Fingerprint: "fp-sync",
		Results: []model.Result{
			{
				Playbook:   model.Playbook{ID: "timeout", Title: "Timeout"},
				Score:      0.8,
				Confidence: 0.75,
			},
		},
	}
	got := Sync(a)
	if got == nil {
		t.Fatal("expected non-nil result from Sync")
	}
	if got.Status != model.ArtifactStatusMatched {
		t.Errorf("expected ArtifactStatusMatched after Sync, got %q", got.Status)
	}
	if got.Artifact == nil {
		t.Error("expected non-nil Artifact after Sync")
	}
	// original must not be mutated
	if a.Status != "" {
		t.Errorf("Sync must not mutate original analysis, but Status was set to %q", a.Status)
	}
}

func TestSyncDoesNotMutateOriginal(t *testing.T) {
	a := &model.Analysis{
		Source:          "ci.log",
		DominantSignals: []string{"sig1"},
	}
	synced := Sync(a)
	// Mutate the synced clone's DominantSignals to verify it's a copy.
	synced.DominantSignals = append(synced.DominantSignals, "sig2")
	if len(a.DominantSignals) != 1 {
		t.Errorf("Sync mutated original DominantSignals slice: %v", a.DominantSignals)
	}
}

func TestSyncWithEmptyResults(t *testing.T) {
	a := &model.Analysis{
		Source:      "ci.log",
		Fingerprint: "fp-empty",
	}
	got := Sync(a)
	if got.Status != model.ArtifactStatusUnknown {
		t.Errorf("expected ArtifactStatusUnknown for no results, got %q", got.Status)
	}
}

// ── splitCommand ──────────────────────────────────────────────────────────────

func TestSplitCommand(t *testing.T) {
	cases := []struct {
		name string
		cmd  string
		want []string
	}{
		{name: "empty string", cmd: "", want: nil},
		{name: "whitespace only", cmd: "   ", want: nil},
		{name: "single word", cmd: "go", want: []string{"go"}},
		{name: "multi word", cmd: "go test ./...", want: []string{"go", "test", "./..."}},
		{name: "extra whitespace trimmed", cmd: "  go  test  ", want: []string{"go", "test"}},
		{name: "command with flag", cmd: "npm install --save-dev", want: []string{"npm", "install", "--save-dev"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := splitCommand(tc.cmd)
			if len(got) != len(tc.want) {
				t.Fatalf("splitCommand(%q) = %v, want %v", tc.cmd, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("splitCommand(%q)[%d] = %q, want %q", tc.cmd, i, got[i], tc.want[i])
				}
			}
		})
	}
}

// ── buildCommand ──────────────────────────────────────────────────────────────

func TestBuildCommand(t *testing.T) {
	cases := []struct {
		name      string
		phase     string
		index     int
		cmd       string
		rationale string
		wantID    string
		wantParts []string
	}{
		{
			name:  "index 0 becomes -1",
			phase: "diagnose", index: 0, cmd: "go test ./...", rationale: "run tests",
			wantID: "diagnose-1", wantParts: []string{"go", "test", "./..."},
		},
		{
			name:  "index 2 becomes -3",
			phase: "remediate", index: 2, cmd: "make build", rationale: "rebuild",
			wantID: "remediate-3", wantParts: []string{"make", "build"},
		},
		{
			name:  "empty command yields nil Command slice",
			phase: "check", index: 0, cmd: "", rationale: "no-op",
			wantID: "check-1", wantParts: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildCommand(tc.phase, tc.index, tc.cmd, tc.rationale)
			if got.ID != tc.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tc.wantID)
			}
			if got.Phase != tc.phase {
				t.Errorf("Phase = %q, want %q", got.Phase, tc.phase)
			}
			if got.WorkDir != "." {
				t.Errorf("WorkDir = %q, want %q", got.WorkDir, ".")
			}
			if got.Rationale != tc.rationale {
				t.Errorf("Rationale = %q, want %q", got.Rationale, tc.rationale)
			}
			if len(got.Command) != len(tc.wantParts) {
				t.Fatalf("Command = %v, want %v", got.Command, tc.wantParts)
			}
			for i := range got.Command {
				if got.Command[i] != tc.wantParts[i] {
					t.Errorf("Command[%d] = %q, want %q", i, got.Command[i], tc.wantParts[i])
				}
			}
		})
	}
}

// ── ciAfterHints ──────────────────────────────────────────────────────────────

func TestCIAfterHints(t *testing.T) {
	const hint1 = "Add or reorder setup so the failing step sees the expected runtime, dependency, or artifact state."
	const hint2 = "Keep the CI change minimal and verify it with the shipped workflow.verify commands."

	cases := []struct {
		name     string
		steps    []string
		wantLen  int
		wantLast string
	}{
		{
			name:  "no fix steps returns 2 canned hints",
			steps: nil, wantLen: 2, wantLast: hint2,
		},
		{
			name:  "empty fix steps skipped",
			steps: []string{"", "  "}, wantLen: 2, wantLast: hint2,
		},
		{
			name:  "one fix step appended",
			steps: []string{"run npm install"}, wantLen: 3, wantLast: "run npm install",
		},
		{
			name:  "only first non-empty step used, max 3 total",
			steps: []string{"step-one", "step-two", "step-three"}, wantLen: 3, wantLast: "step-one",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ciAfterHints(model.Result{}, tc.steps)
			if len(got) != tc.wantLen {
				t.Fatalf("ciAfterHints returned %d hints, want %d: %v", len(got), tc.wantLen, got)
			}
			if got[0] != hint1 {
				t.Errorf("hint[0] = %q, want canned hint1", got[0])
			}
			if got[len(got)-1] != tc.wantLast {
				t.Errorf("last hint = %q, want %q", got[len(got)-1], tc.wantLast)
			}
		})
	}
}

// ── isCIConfigFile ────────────────────────────────────────────────────────────

func TestIsCIConfigFile(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{".github/workflows/ci.yml", true},
		{".github/workflows/release.yaml", true},
		{".gitlab-ci.yml", true},
		{"azure-pipelines.yml", true},
		{".circleci/config.yml", true},
		{".circleci/continue.yml", true},
		{"Jenkinsfile", true},
		{"Makefile", false},
		{"src/main.go", false},
		{"README.md", false},
		{"docker-compose.yml", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			got := isCIConfigFile(tc.path)
			if got != tc.want {
				t.Errorf("isCIConfigFile(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

// ── playbookSeedPath ──────────────────────────────────────────────────────────

func TestPlaybookSeedPath(t *testing.T) {
	cases := []struct {
		name string
		seed *model.SuggestedPlaybookSeed
		want string
	}{
		{
			name: "nil seed returns unknown path",
			seed: nil,
			want: "playbooks/bundled/log/unknown/seed.yaml",
		},
		{
			name: "empty category returns unknown path",
			seed: &model.SuggestedPlaybookSeed{Category: ""},
			want: "playbooks/bundled/log/unknown/seed.yaml",
		},
		{
			name: "whitespace-only category returns unknown path",
			seed: &model.SuggestedPlaybookSeed{Category: "   "},
			want: "playbooks/bundled/log/unknown/seed.yaml",
		},
		{
			name: "known category returns correct path",
			seed: &model.SuggestedPlaybookSeed{Category: "auth"},
			want: "playbooks/bundled/log/auth/seed.yaml",
		},
		{
			name: "docker category",
			seed: &model.SuggestedPlaybookSeed{Category: "docker"},
			want: "playbooks/bundled/log/docker/seed.yaml",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := playbookSeedPath(tc.seed)
			if got != tc.want {
				t.Errorf("playbookSeedPath(%v) = %q, want %q", tc.seed, got, tc.want)
			}
		})
	}
}

// ── pathBase ──────────────────────────────────────────────────────────────────

func TestPathBase(t *testing.T) {
	cases := []struct {
		value string
		want  string
	}{
		{"/usr/bin/node", "node"},
		{"foo/bar/baz", "baz"},
		{"baz", "baz"},
		{"", ""},
		{"a/b/c/d", "d"},
	}
	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			got := pathBase(tc.value)
			if got != tc.want {
				t.Errorf("pathBase(%q) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

// ── executableName ────────────────────────────────────────────────────────────

func TestExecutableName(t *testing.T) {
	cases := []struct {
		value string
		want  string
	}{
		{"/usr/bin/node", "node"},
		{"node", "node"},
		{`"node"`, "node"},
		{"'node'", "node"},
		{"/usr/local/bin/npm", "npm"},
		{"./bin/runner", "runner"},
		{"", ""},
	}
	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			got := executableName(tc.value)
			if got != tc.want {
				t.Errorf("executableName(%q) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

// ── extractMissingExecutable ──────────────────────────────────────────────────

func TestExtractMissingExecutable(t *testing.T) {
	cases := []struct {
		name     string
		evidence []string
		wantExec string
		wantPath string
	}{
		{
			name:     "empty evidence",
			evidence: nil,
			wantExec: "", wantPath: "",
		},
		{
			name:     "blank lines skipped",
			evidence: []string{"", "   "},
			wantExec: "", wantPath: "",
		},
		{
			name:     "exec path pattern no-such-file",
			evidence: []string{"exec /usr/bin/node: no such file or directory"},
			wantExec: "node", wantPath: "/usr/bin/node",
		},
		{
			name:     "exec bare name no-such-file",
			evidence: []string{"exec foo: no such file or directory"},
			wantExec: "foo", wantPath: "foo",
		},
		{
			name:     "exec quoted pattern",
			evidence: []string{"exec: 'node': not found"},
			wantExec: "node", wantPath: "node",
		},
		{
			name:     "command not found pattern",
			evidence: []string{"node: command not found"},
			wantExec: "node", wantPath: "node",
		},
		{
			name:     "no matching pattern",
			evidence: []string{"some random log line"},
			wantExec: "", wantPath: "",
		},
		{
			name:     "first matching line wins",
			evidence: []string{"unrelated", "exec /usr/bin/npm: no such file or directory", "node: command not found"},
			wantExec: "npm", wantPath: "/usr/bin/npm",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotExec, gotPath := extractMissingExecutable(tc.evidence)
			if gotExec != tc.wantExec {
				t.Errorf("executable = %q, want %q", gotExec, tc.wantExec)
			}
			if gotPath != tc.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tc.wantPath)
			}
		})
	}
}
