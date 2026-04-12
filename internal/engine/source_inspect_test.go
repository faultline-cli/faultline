package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/detectors"
	"faultline/internal/model"
)

func TestAnalyzeRepositoryTriggerAmplifierMitigationScoring(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	triggerOnly := analyzeRepoFixture(t, playbookDir, map[string]string{
		"internal/api/handler.go": sourceTriggerOnlyFixture(),
	}, detectors.ChangeSet{})
	amp := analyzeRepoFixture(t, playbookDir, map[string]string{
		"internal/api/handler.go": sourceAmplifiedFixture(),
	}, detectors.ChangeSet{})
	mitigated := analyzeRepoFixture(t, playbookDir, map[string]string{
		"internal/api/handler.go": sourceMitigatedFixture(),
	}, detectors.ChangeSet{})

	triggerResult := requireFinding(t, triggerOnly, "network-fanout-without-guards")
	ampResult := requireFinding(t, amp, "network-fanout-without-guards")
	mitigatedResult := requireFinding(t, mitigated, "network-fanout-without-guards")

	if ampResult.Score <= triggerResult.Score {
		t.Fatalf("expected amplifier score %.2f to exceed trigger-only score %.2f", ampResult.Score, triggerResult.Score)
	}
	if mitigatedResult.Score >= ampResult.Score {
		t.Fatalf("expected mitigation score %.2f to lower amplified score %.2f", mitigatedResult.Score, ampResult.Score)
	}
	if len(triggerResult.Explanation.TriggeredBy) == 0 {
		t.Fatal("expected trigger explanation")
	}
	if len(ampResult.Explanation.AmplifiedBy) == 0 {
		t.Fatal("expected amplifier explanation")
	}
	if len(mitigatedResult.Explanation.MitigatedBy) == 0 {
		t.Fatal("expected mitigation explanation")
	}
}

func TestAnalyzeRepositoryMitigationProximityAffectsScore(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	near := analyzeRepoFixture(t, playbookDir, map[string]string{
		"internal/api/handler.go": sourceMitigatedFixture(),
	}, detectors.ChangeSet{})
	far := analyzeRepoFixture(t, playbookDir, map[string]string{
		"internal/api/handler.go": sourceAmplifiedFixture(),
		"internal/api/client.go":  "package api\n\nfunc buildClient() *http.Client {\n\treturn &http.Client{Timeout: 5 * time.Second}\n}\n",
	}, detectors.ChangeSet{})

	nearResult := requireFinding(t, near, "network-fanout-without-guards")
	farResult := requireFinding(t, far, "network-fanout-without-guards")
	if nearResult.Score >= farResult.Score {
		t.Fatalf("expected near mitigation score %.2f to be lower than far mitigation score %.2f", nearResult.Score, farResult.Score)
	}
}

func TestAnalyzeRepositoryProductionScoresHigherThanTestPath(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	prod := analyzeRepoFixture(t, playbookDir, map[string]string{
		"internal/api/handler.go": sourceAmplifiedFixture(),
	}, detectors.ChangeSet{})
	testOnly := analyzeRepoFixture(t, playbookDir, map[string]string{
		"internal/api/handler_test.go": sourceAmplifiedFixture(),
	}, detectors.ChangeSet{})

	prodResult := requireFinding(t, prod, "network-fanout-without-guards")
	testResult := requireFinding(t, testOnly, "network-fanout-without-guards")
	if prodResult.Score <= testResult.Score {
		t.Fatalf("expected production path %.2f to exceed test path %.2f", prodResult.Score, testResult.Score)
	}
}

func TestAnalyzeRepositoryChangeAwareness(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	files := map[string]string{
		"internal/api/handler.go": sourceAmplifiedFixture(),
	}
	introduced := analyzeRepoFixture(t, playbookDir, files, detectors.ChangeSet{
		ChangedFiles: map[string]detectors.FileChange{
			"internal/api/handler.go": {Status: "added"},
		},
	})
	legacy := analyzeRepoFixture(t, playbookDir, files, detectors.ChangeSet{})

	introducedResult := requireFinding(t, introduced, "network-fanout-without-guards")
	legacyResult := requireFinding(t, legacy, "network-fanout-without-guards")
	if introducedResult.Score <= legacyResult.Score {
		t.Fatalf("expected introduced score %.2f to exceed legacy score %.2f", introducedResult.Score, legacyResult.Score)
	}
	if introducedResult.ChangeStatus != "introduced" {
		t.Fatalf("expected introduced change status, got %q", introducedResult.ChangeStatus)
	}
	if legacyResult.ChangeStatus != "legacy" {
		t.Fatalf("expected legacy change status, got %q", legacyResult.ChangeStatus)
	}
}

func TestAnalyzeRepositorySuppression(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	engine := New(Options{PlaybookDir: playbookDir, NoHistory: true})
	root := writeRepoFiles(t, map[string]string{
		"examples/sample.go": sourceAmplifiedFixture(),
	})
	_, err := engine.AnalyzeRepository(root, detectors.ChangeSet{})
	if !errors.Is(err, ErrNoMatch) {
		t.Fatalf("expected example suppression to remove finding, got %v", err)
	}

	inline := analyzeRepoFixture(t, playbookDir, map[string]string{
		"internal/api/handler.go": "// faultline:ignore network-fanout-without-guards reason=temporary until=2026-12-31\n" + sourceAmplifiedFixture(),
	}, detectors.ChangeSet{})
	if len(inline.Results) != 0 {
		t.Fatalf("expected inline suppression to remove finding, got %+v", inline.Results)
	}
}

func TestAnalyzeRepositoryCompoundSignal(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	analysis := analyzeRepoFixture(t, playbookDir, map[string]string{
		"internal/api/handler.go": sourceAmplifiedFixture(),
	}, detectors.ChangeSet{})
	result := requireFinding(t, analysis, "network-fanout-without-guards")
	if !containsText(result.Explanation.AmplifiedBy, "fragile fan-out") {
		t.Fatalf("expected compound amplifier in %+v", result.Explanation.AmplifiedBy)
	}
}

func TestAnalyzeRepositoryLocalConsistencyAmplifier(t *testing.T) {
	playbookDir := repoPlaybookDir(t)
	analysis := analyzeRepoFixture(t, playbookDir, map[string]string{
		"internal/payments/charges.go": `package payments

func createPrimary(w http.ResponseWriter, r *http.Request) {
	http.HandleFunc("/charges/primary", createPrimary)
	ensureIdempotency(r)
	gateway.Charge(r.Context())
}

func createSecondary(w http.ResponseWriter, r *http.Request) {
	http.HandleFunc("/charges/secondary", createSecondary)
	ensureIdempotency(r)
	gateway.Charge(r.Context())
}

func createUnsafe(w http.ResponseWriter, r *http.Request) {
	http.HandleFunc("/charges/unsafe", createUnsafe)
	gateway.Charge(r.Context())
}
`,
	}, detectors.ChangeSet{})
	result := requireFinding(t, analysis, "missing-idempotency-guard")
	if !containsText(result.Explanation.AmplifiedBy, "local idempotency idiom") {
		t.Fatalf("expected local consistency amplifier in %+v", result.Explanation.AmplifiedBy)
	}
}

func TestAnalyzeReaderLogDetectionsStillWorkWithSourcePlaybooksLoaded(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})
	a, err := e.AnalyzeReader(strings.NewReader(
		"Warning Failed pod/app-123 Failed to pull image \"ghcr.io/acme/app:missing\": manifest unknown\nBack-off pulling image\n",
	))
	if err != nil {
		t.Fatalf("analyze reader: %v", err)
	}
	if a.Results[0].Playbook.ID != "image-pull-backoff" {
		t.Fatalf("expected image-pull-backoff, got %s", a.Results[0].Playbook.ID)
	}
}

func analyzeRepoFixture(t *testing.T, playbookDir string, files map[string]string, changeSet detectors.ChangeSet) *model.Analysis {
	t.Helper()
	root := writeRepoFiles(t, files)
	e := New(Options{PlaybookDir: playbookDir, NoHistory: true})
	a, err := e.AnalyzeRepository(root, changeSet)
	if err != nil && !errors.Is(err, ErrNoMatch) {
		t.Fatalf("analyze repository: %v", err)
	}
	if a == nil {
		t.Fatal("expected analysis")
	}
	return a
}

func requireFinding(t *testing.T, analysis *model.Analysis, id string) model.Result {
	t.Helper()
	for _, result := range analysis.Results {
		if result.Playbook.ID == id {
			return result
		}
	}
	t.Fatalf("expected finding %s in %+v", id, analysis.Results)
	return model.Result{}
}

func writeRepoFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	return root
}

func containsText(items []string, want string) bool {
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), strings.ToLower(want)) {
			return true
		}
	}
	return false
}

func sourceTriggerOnlyFixture() string {
	return `package api

func sendAll(items []string) error {
	for _, item := range items {
		req, _ := http.NewRequest("GET", item, nil)
		_, err := client.Do(req)
		if err != nil {
			return err
		}
	}
	return nil
}
`
}

func sourceAmplifiedFixture() string {
	return `package api

func sendAll(items []string) error {
	client := http.DefaultClient
	for _, item := range items {
		go func(url string) {
			req, _ := http.NewRequest("GET", url, nil)
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			_ = resp
		}(item)
	}
	return nil
}
`
}

func sourceMitigatedFixture() string {
	return `package api

func sendAll(items []string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	sem := make(chan struct{}, 4)
	for _, item := range items {
		sem <- struct{}{}
		req, _ := http.NewRequest("GET", item, nil)
		err := retry.Do(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			req = req.WithContext(ctx)
			_, err := client.Do(req)
			return err
		})
		<-sem
		if err != nil {
			return err
		}
	}
	return nil
}
`
}
