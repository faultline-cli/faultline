package fixtures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunIngestionPipelinePromptCoversAdaptersAndSourceFollowUp(t *testing.T) {
	body := readRepoDoc(t, filepath.Join("..", "..", "prompts", "run-ingestion-pipeline.md"))
	assertIngestionContract(t, body)
}

func TestIngestionPipelineSkillMatchesPromptContract(t *testing.T) {
	body := readRepoDoc(t, filepath.Join("..", "..", "agents", "skills", "ingestion-pipeline", "SKILL.md"))
	assertIngestionContract(t, body)
}

func readRepoDoc(t *testing.T, rel string) string {
	t.Helper()

	data, err := os.ReadFile(rel)
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

func assertIngestionContract(t *testing.T, body string) {
	t.Helper()

	required := []string{
		"github-issue",
		"gitlab-issue",
		"stackexchange-question",
		"discourse-topic",
		"reddit-post",
		"./bin/faultline fixtures stats --class real --json",
		"./bin/faultline fixtures stats --class real --check-baseline",
		"underrepresented adapters",
		"internal/engine/testdata/source/",
		"source-playbook",
	}

	lowerBody := strings.ToLower(body)
	for _, fragment := range required {
		if !strings.Contains(lowerBody, strings.ToLower(fragment)) {
			t.Fatalf("expected ingestion contract to mention %q", fragment)
		}
	}
}
