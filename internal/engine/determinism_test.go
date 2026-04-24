package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/matcher"
	"faultline/internal/playbooks"
)

// TestDeterminismAcrossMultipleRuns verifies that the same input produces
// identical output across multiple analysis runs.
func TestDeterminismAcrossMultipleRuns(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	testCases := []struct {
		name string
		log  string
	}{
		{
			name: "git auth failure",
			log:  "fatal: could not read Username for 'https://github.com': terminal prompts disabled\n",
		},
		{
			name: "docker auth failure",
			log:  "Error response from daemon: Get https://registry-1.docker.io/v2/: unauthorized: authentication required\n",
		},
		{
			name: "network timeout",
			log:  "context deadline exceeded\nconnection timeout\n",
		},
		{
			name: "disk full",
			log:  "no space left on device\nwrite error\n",
		},
		{
			name: "permission denied",
			log:  "permission denied: /var/run/docker.sock\n",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Run analysis 5 times
			results := make([][]string, 5)
			for i := 0; i < 5; i++ {
				analysis, err := e.AnalyzeReader(strings.NewReader(tc.log))
				if err != nil && err != ErrNoMatch {
					t.Fatalf("run %d: analyze failed: %v", i+1, err)
				}

				// Extract result IDs for comparison
				if analysis != nil {
					for _, r := range analysis.Results {
						results[i] = append(results[i], r.Playbook.ID)
					}
				}
			}

			// Verify all runs produced identical results
			for i := 1; i < 5; i++ {
				if len(results[0]) != len(results[i]) {
					t.Fatalf("run 1 had %d results, run %d had %d results",
						len(results[0]), i+1, len(results[i]))
				}
				for j := range results[0] {
					if results[0][j] != results[i][j] {
						t.Fatalf("run 1 result %d was %s, run %d result %d was %s",
							j, results[0][j], i+1, j, results[i][j])
					}
				}
			}
		})
	}
}

// TestDeterminismWithFixtures verifies that fixture-based analysis is deterministic.
func TestDeterminismWithFixtures(t *testing.T) {
	pbs, err := playbooks.LoadDir(repoPlaybookDir(t))
	if err != nil {
		t.Fatalf("load playbooks: %v", err)
	}

	fixtureDir := filepath.Join("testdata", "fixtures")
	entries, err := os.ReadDir(fixtureDir)
	if err != nil {
		t.Fatalf("read fixtures dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}

		entry := entry
		t.Run(entry.Name(), func(t *testing.T) {
			logPath := filepath.Join(fixtureDir, entry.Name())
			data, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			// Run analysis 3 times
			var firstResults, secondResults, thirdResults []string

			for run := 0; run < 3; run++ {
				lines, err := readLines(strings.NewReader(string(data)))
				if err != nil {
					t.Fatalf("read lines: %v", err)
				}
				ctx := ExtractContext(lines)
				results := matcher.Rank(pbs, lines, ctx)

				for _, r := range results {
					switch run {
					case 0:
						firstResults = append(firstResults, r.Playbook.ID)
					case 1:
						secondResults = append(secondResults, r.Playbook.ID)
					case 2:
						thirdResults = append(thirdResults, r.Playbook.ID)
					}
				}
			}

			// Verify determinism
			if len(firstResults) != len(secondResults) || len(firstResults) != len(thirdResults) {
				t.Fatalf("inconsistent result counts: %d, %d, %d",
					len(firstResults), len(secondResults), len(thirdResults))
			}

			for i := range firstResults {
				if firstResults[i] != secondResults[i] || firstResults[i] != thirdResults[i] {
					t.Fatalf("result %d inconsistent: %s, %s, %s",
						i, firstResults[i], secondResults[i], thirdResults[i])
				}
			}
		})
	}
}

// TestDeterminismWithEdgeCases verifies determinism with edge case inputs.
func TestDeterminismWithEdgeCases(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	testCases := []struct {
		name string
		log  string
	}{
		{
			name: "empty lines",
			log:  "\n\n\n\nfatal: could not read Username for 'https://github.com': terminal prompts disabled\n\n\n",
		},
		{
			name: "trailing whitespace",
			log:  "fatal: could not read Username for 'https://github.com': terminal prompts disabled   \n",
		},
		{
			name: "leading whitespace",
			log:  "   fatal: could not read Username for 'https://github.com': terminal prompts disabled\n",
		},
		{
			name: "mixed case",
			log:  "FATAL: Could not read Username for 'https://github.com': Terminal Prompts Disabled\n",
		},
		{
			name: "repeated patterns",
			log:  strings.Repeat("fatal: could not read Username for 'https://github.com': terminal prompts disabled\n", 10),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Run twice
			a1, err1 := e.AnalyzeReader(strings.NewReader(tc.log))
			a2, err2 := e.AnalyzeReader(strings.NewReader(tc.log))

			// Errors should match
			if (err1 == nil) != (err2 == nil) {
				t.Fatalf("error mismatch: %v vs %v", err1, err2)
			}

			// Results should match
			if a1 == nil && a2 == nil {
				return
			}
			if a1 == nil || a2 == nil {
				t.Fatal("one analysis was nil, the other wasn't")
			}

			if len(a1.Results) != len(a2.Results) {
				t.Fatalf("result count mismatch: %d vs %d", len(a1.Results), len(a2.Results))
			}

			for i := range a1.Results {
				if a1.Results[i].Playbook.ID != a2.Results[i].Playbook.ID {
					t.Fatalf("result %d ID mismatch: %s vs %s",
						i, a1.Results[i].Playbook.ID, a2.Results[i].Playbook.ID)
				}
				if a1.Results[i].Score != a2.Results[i].Score {
					t.Fatalf("result %d score mismatch: %f vs %f",
						i, a1.Results[i].Score, a2.Results[i].Score)
				}
			}
		})
	}
}

// TestDeterminismWithLargeInputs verifies determinism with large log inputs.
func TestDeterminismWithLargeInputs(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	// Create a large log with many lines
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString("normal log line\n")
	}
	sb.WriteString("fatal: could not read Username for 'https://github.com': terminal prompts disabled\n")
	for i := 0; i < 1000; i++ {
		sb.WriteString("more normal log line\n")
	}

	log := sb.String()

	// Run analysis twice
	a1, err1 := e.AnalyzeReader(strings.NewReader(log))
	a2, err2 := e.AnalyzeReader(strings.NewReader(log))

	if (err1 == nil) != (err2 == nil) {
		t.Fatalf("error mismatch: %v vs %v", err1, err2)
	}

	if a1 == nil || a2 == nil {
		t.Fatal("analysis was nil")
	}

	if len(a1.Results) != len(a2.Results) {
		t.Fatalf("result count mismatch: %d vs %d", len(a1.Results), len(a2.Results))
	}

	for i := range a1.Results {
		if a1.Results[i].Playbook.ID != a2.Results[i].Playbook.ID {
			t.Fatalf("result %d ID mismatch: %s vs %s",
				i, a1.Results[i].Playbook.ID, a2.Results[i].Playbook.ID)
		}
	}
}

// TestDeterminismWithSpecialCharacters verifies determinism with special characters.
func TestDeterminismWithSpecialCharacters(t *testing.T) {
	e := New(Options{PlaybookDir: repoPlaybookDir(t), NoHistory: true})

	testCases := []struct {
		name string
		log  string
	}{
		{
			name: "tabs",
			log:  "fatal:\tcould\tnot\tread\tUsername\tfor\t'https://github.com':\tterminal\tprompts\tdisabled\n",
		},
		{
			name: "multiple spaces",
			log:  "fatal:    could    not    read    Username    for    'https://github.com':    terminal    prompts    disabled\n",
		},
		{
			name: "special chars in path",
			log:  "fatal: could not read Username for 'https://github.com/user@domain/repo-name_v1.0': terminal prompts disabled\n",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			a1, _ := e.AnalyzeReader(strings.NewReader(tc.log))
			a2, _ := e.AnalyzeReader(strings.NewReader(tc.log))

			if a1 == nil && a2 == nil {
				return
			}
			if a1 == nil || a2 == nil {
				t.Fatal("one analysis was nil")
			}

			if len(a1.Results) != len(a2.Results) {
				t.Fatalf("result count mismatch: %d vs %d", len(a1.Results), len(a2.Results))
			}

			for i := range a1.Results {
				if a1.Results[i].Playbook.ID != a2.Results[i].Playbook.ID {
					t.Fatalf("result %d mismatch: %s vs %s",
						i, a1.Results[i].Playbook.ID, a2.Results[i].Playbook.ID)
				}
			}
		})
	}
}
