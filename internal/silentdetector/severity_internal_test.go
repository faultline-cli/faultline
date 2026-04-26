package silentdetector

import "testing"

// TestSeverityRankAllCases exercises all branches of severityRank, including
// the "low" and default (unknown) cases that the external test suite cannot
// reach because no built-in detector emits those severities.
func TestSeverityRankAllCases(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"high", 3},
		{"HIGH", 3},
		{"medium", 2},
		{"MEDIUM", 2},
		{"low", 1},
		{"LOW", 1},
		{"", 0},
		{"unknown", 0},
		{"critical", 0},
	}
	for _, tc := range cases {
		if got := severityRank(tc.input); got != tc.want {
			t.Errorf("severityRank(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

// TestConfidenceRankAllCases exercises all branches of confidenceRank, including
// the "low" and default (unknown) cases.
func TestConfidenceRankAllCases(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"high", 3},
		{"HIGH", 3},
		{"medium", 2},
		{"MEDIUM", 2},
		{"low", 1},
		{"LOW", 1},
		{"", 0},
		{"unknown", 0},
	}
	for _, tc := range cases {
		if got := confidenceRank(tc.input); got != tc.want {
			t.Errorf("confidenceRank(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}
