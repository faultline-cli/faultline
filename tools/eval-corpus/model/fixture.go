package model

// Fixture is a normalized, deduplicated log ready for evaluation.
type Fixture struct {
	// ID is the SHA-256 hash of the normalized log content.
	ID       string          `json:"id"`
	Raw      string          `json:"raw"`
	Source   string          `json:"source"`
	Metadata FixtureMetadata `json:"metadata"`
	// Label carries optional ground-truth annotations for accuracy scoring.
	// When nil the fixture is unlabelled.
	Label *FixtureLabel `json:"label,omitempty"`
}

// FixtureMetadata captures origin context preserved for reproducibility.
type FixtureMetadata struct {
	Timestamp string            `json:"timestamp,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
	LineNum   int               `json:"line_num,omitempty"`
}

// FixtureLabel holds optional ground-truth annotations.
// Presence of any field enables accuracy scoring for that fixture.
type FixtureLabel struct {
	ExpectedFailureID           string `json:"expected_failure_id,omitempty"`
	ExpectedFailureClass        string `json:"expected_failure_class,omitempty"`
	ExpectedSeverity            string `json:"expected_severity,omitempty"`
	ExpectedRemediationCategory string `json:"expected_remediation_category,omitempty"`
}
