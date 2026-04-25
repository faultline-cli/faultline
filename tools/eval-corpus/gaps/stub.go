package gaps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"faultline/tools/eval-corpus/model"
)

// PlaybookStub is a minimal YAML playbook candidate generated for review.
// It is not intended to be used as-is; it surfaces a reviewable starting point.
type PlaybookStub struct {
	// ClusterID ties the stub back to its source cluster.
	ClusterID string
	// YAML is the rendered stub content.
	YAML string
}

// GenerateStub produces a reviewable playbook stub YAML string from a Cluster.
func GenerateStub(c Cluster) PlaybookStub {
	id := stubID(c)
	yaml := fmt.Sprintf(`# Auto-generated stub for review — NOT production-ready.
# Cluster: %s  (%d unmatched fixtures)
# Suspected class: %s  confidence: %.2f
# Notes: %s
#
# TODO: Replace the pattern below with a real regex drawn from the sample logs.
# TODO: Fill in links, remediation steps, and severity.
# TODO: Remove this comment block when the playbook is production-ready.

id: %s
name: "%s"
description: "TODO: Describe the failure pattern."
severity: error
tags:
  - %s
  - unverified-stub
patterns:
  - type: regex
    value: "TODO: replace with a real regex"
    # Representative error line (normalized):
    # %s
evidence:
  - type: log_line
    label: "TODO: describe what to capture"
    pattern: "TODO: replace with real pattern"
links: []
remediation: "TODO: add remediation guidance."
`,
		c.ClusterID, c.Count,
		c.SuspectedFailureClass, c.Confidence,
		c.Notes,
		id, humanName(c),
		c.SuspectedFailureClass,
		c.RepresentativeErrorLine,
	)
	return PlaybookStub{ClusterID: c.ClusterID, YAML: yaml}
}

// stubID produces a stable playbook ID from the cluster ID and class.
func stubID(c Cluster) string {
	class := strings.ReplaceAll(c.SuspectedFailureClass, " ", "-")
	return fmt.Sprintf("stub-%s-%s", class, c.ClusterID)
}

// humanName returns a readable playbook name candidate.
func humanName(c Cluster) string {
	class := strings.ReplaceAll(c.SuspectedFailureClass, "-", " ")
	return fmt.Sprintf("%s (stub %s)", class, c.ClusterID)
}

// WriteSamples writes sample log files for a cluster into a directory.
// Each sample is named sample-NNN.log and contains the raw log text of the
// fixture identified by SampleFixtureIDs.
// fixtureIndex maps fixture ID → Fixture for fast lookup.
func WriteSamples(dir string, c Cluster, fixtureIndex map[string]model.Fixture) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create sample dir %q: %w", dir, err)
	}
	for i, id := range c.SampleFixtureIDs {
		f, ok := fixtureIndex[id]
		if !ok {
			continue
		}
		name := filepath.Join(dir, fmt.Sprintf("sample-%03d.log", i+1))
		if err := os.WriteFile(name, []byte(f.Raw+"\n"), 0o644); err != nil {
			return fmt.Errorf("write sample %q: %w", name, err)
		}
	}
	return nil
}
