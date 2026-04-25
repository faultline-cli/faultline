package model

// EvalResult is the output of running Faultline against a single fixture.
type EvalResult struct {
	FixtureID  string   `json:"fixture_id"`
	Source     string   `json:"source,omitempty"`
	Matched    bool     `json:"matched"`
	FailureID  string   `json:"failure_id,omitempty"`
	Confidence float64  `json:"confidence,omitempty"`
	Evidence   []string `json:"evidence,omitempty"`
	DurationMS int64    `json:"duration_ms"`
	// Error holds a non-empty string when the engine failed to analyse the fixture.
	Error string `json:"error,omitempty"`
	// FirstLineTag is a short hash of the first normalized log line, used for
	// clustering unmatched results in the report.
	FirstLineTag string `json:"first_line_tag,omitempty"`
	// FirstLineSnippet is the first 120 characters of the first normalized log
	// line, used as a human-readable cluster label in the report.
	FirstLineSnippet string `json:"first_line_snippet,omitempty"`

	// Label fields — optional. When present, accuracy metrics can be computed.
	// When absent the result contributes only to coverage metrics.

	// ExpectedFailureID is the ground-truth failure ID for the fixture.
	ExpectedFailureID string `json:"expected_failure_id,omitempty"`
	// ExpectedFailureClass is the broad category of the expected failure.
	ExpectedFailureClass string `json:"expected_failure_class,omitempty"`
	// ExpectedSeverity is the expected semantic severity (e.g. "error", "warning").
	ExpectedSeverity string `json:"expected_severity,omitempty"`
	// ExpectedRemediationCategory is a coarse remediation hint (e.g. "auth", "dependency").
	ExpectedRemediationCategory string `json:"expected_remediation_category,omitempty"`
}

// LabelOutcome classifies a labelled result as TP/FP/FN/unknown.
type LabelOutcome string

const (
	// OutcomeTP is a true positive: matched and matched the expected failure.
	OutcomeTP LabelOutcome = "tp"
	// OutcomeFP is a false positive: matched but did not match the expected failure.
	OutcomeFP LabelOutcome = "fp"
	// OutcomeFN is a false negative: not matched but a label was present.
	OutcomeFN LabelOutcome = "fn"
	// OutcomeUnlabelled means no expected_failure_id was provided.
	OutcomeUnlabelled LabelOutcome = "unlabelled"
)

// ScoreResult computes the LabelOutcome for a single result.
// Unlabelled results (no ExpectedFailureID) return OutcomeUnlabelled
// and must not be counted toward accuracy metrics.
func ScoreResult(r EvalResult) LabelOutcome {
	if r.ExpectedFailureID == "" {
		return OutcomeUnlabelled
	}
	if r.Matched && r.FailureID == r.ExpectedFailureID {
		return OutcomeTP
	}
	if r.Matched && r.FailureID != r.ExpectedFailureID {
		return OutcomeFP
	}
	return OutcomeFN
}
