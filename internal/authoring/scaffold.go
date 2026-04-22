package authoring

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ScaffoldOptions configures playbook scaffold generation.
type ScaffoldOptions struct {
	// Category is the playbook category hint: auth, build, ci, deploy, network, runtime, or test.
	// Defaults to "build" when empty.
	Category string
	// ID overrides the auto-derived playbook ID. When empty, an ID is derived
	// from the first extracted candidate pattern and the category.
	ID string
	// PackDir is an optional destination directory. When set, the scaffold is
	// written to PackDir/<id>.yaml in addition to being returned in the result.
	PackDir string
	// MaxMatch caps the number of match.any patterns extracted from the log.
	// Defaults to 5 when zero.
	MaxMatch int
}

// ScaffoldResult contains the output of ScaffoldPlaybook.
type ScaffoldResult struct {
	// YAML is the generated playbook scaffold. All fields marked TODO require
	// human review before the scaffold can be committed.
	YAML string
	// SuggestedID is the derived or specified playbook ID.
	SuggestedID string
	// Candidates holds every extracted pattern before the MaxMatch cap, so the
	// caller can present the full extraction set for review.
	Candidates []string
	// OutputPath is the path where the scaffold was written. Empty when PackDir
	// was not set.
	OutputPath string
}

// slugRe replaces non-alphanumeric sequences with a single hyphen in IDs.
var slugRe = regexp.MustCompile(`[^a-z0-9]+`)
var validIDRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// validCategories is the closed set accepted by bundled playbooks.
var validCategories = map[string]bool{
	"auth":    true,
	"build":   true,
	"ci":      true,
	"deploy":  true,
	"network": true,
	"runtime": true,
	"test":    true,
}

// ScaffoldPlaybook generates a candidate playbook YAML scaffold from a
// sanitized log. The result is fully deterministic: the same sanitized log and
// options always produce the same scaffold. No network calls are made.
//
// The caller is responsible for applying fixtures.ApplySanitizeRules before
// passing the log to this function.
func ScaffoldPlaybook(sanitizedLog string, opts ScaffoldOptions) (ScaffoldResult, error) {
	maxMatch := opts.MaxMatch
	if maxMatch <= 0 {
		maxMatch = 5
	}

	category, err := normalizeCategory(opts.Category)
	if err != nil {
		return ScaffoldResult{}, err
	}

	// Extract candidate patterns. Over-ask so the caller sees a fuller list.
	allCandidates := ExtractCandidatePatterns(sanitizedLog, maxMatch*2)

	matchPatterns := allCandidates
	if len(matchPatterns) > maxMatch {
		matchPatterns = matchPatterns[:maxMatch]
	}

	// Derive a playbook ID.
	id := strings.TrimSpace(opts.ID)
	if id == "" {
		if len(matchPatterns) > 0 {
			base := slugRe.ReplaceAllString(strings.ToLower(matchPatterns[0]), "-")
			base = strings.Trim(base, "-")
			if len(base) > 40 {
				base = strings.TrimRight(base[:40], "-")
			}
			id = category + "-" + base
		} else {
			id = category + "-unknown"
		}
	} else if !validIDRe.MatchString(id) {
		return ScaffoldResult{}, fmt.Errorf("invalid playbook id %q: use lowercase letters, numbers, and single hyphens only", id)
	}

	yamlText := buildScaffoldYAML(id, category, matchPatterns)

	result := ScaffoldResult{
		YAML:        yamlText,
		SuggestedID: id,
		Candidates:  allCandidates,
	}

	if opts.PackDir != "" {
		if err := os.MkdirAll(opts.PackDir, 0o755); err != nil {
			return ScaffoldResult{}, fmt.Errorf("create pack directory %s: %w", opts.PackDir, err)
		}
		outPath := filepath.Join(opts.PackDir, id+".yaml")
		if err := os.WriteFile(outPath, []byte(yamlText), 0o644); err != nil {
			return ScaffoldResult{}, fmt.Errorf("write scaffold %s: %w", outPath, err)
		}
		result.OutputPath = outPath
	}

	return result, nil
}

func normalizeCategory(category string) (string, error) {
	category = strings.TrimSpace(category)
	if category == "" {
		return "build", nil
	}
	if !validCategories[category] {
		return "", fmt.Errorf("invalid category %q: must be one of auth, build, ci, deploy, network, runtime, or test", category)
	}
	return category, nil
}

// buildScaffoldYAML constructs the scaffold YAML string. Fields that require
// human review are annotated with TODO markers.
func buildScaffoldYAML(id, category string, matchPatterns []string) string {
	var sb strings.Builder

	sb.WriteString("# SCAFFOLD - Review all TODO fields before committing.\n")
	sb.WriteString("# Validate with: make review && make test && make fixture-check\n")
	sb.WriteString("#\n")
	sb.WriteString("# match.any  — trim each pattern to the exact error substring from real logs.\n")
	sb.WriteString("# match.none — add at least one exclusion for the nearest confusable neighbor.\n")
	sb.WriteString("\n")

	fmt.Fprintf(&sb, "id: %s\n", id)
	fmt.Fprintf(&sb, "title: \"TODO: one-sentence title\"\n")
	fmt.Fprintf(&sb, "category: %s\n", category)
	sb.WriteString("severity: medium\n")
	sb.WriteString("base_score: 0.5\n")
	fmt.Fprintf(&sb, "tags: [%s]\n", category)
	sb.WriteString("stage_hints: [build, test]\n")
	sb.WriteString("\n")

	sb.WriteString("match:\n")
	sb.WriteString("  any:\n")
	if len(matchPatterns) == 0 {
		sb.WriteString("    # TODO: Add at least one distinctive error substring.\n")
		sb.WriteString("    - TODO\n")
	} else {
		for _, p := range matchPatterns {
			fmt.Fprintf(&sb, "    - %s\n", yamlScalar(p))
		}
	}
	sb.WriteString("  none:\n")
	sb.WriteString("    # TODO: Add exclusions for the nearest confusable neighbor.\n")
	sb.WriteString("\n")

	sb.WriteString("summary: |\n")
	sb.WriteString("  TODO: One or two sentences describing what this failure is.\n")
	sb.WriteString("\n")

	sb.WriteString("diagnosis: |\n")
	sb.WriteString("  ## Diagnosis\n")
	sb.WriteString("\n")
	sb.WriteString("  TODO: Explain the root cause in plain language.\n")
	sb.WriteString("\n")

	sb.WriteString("fix: |\n")
	sb.WriteString("  ## Fix steps\n")
	sb.WriteString("\n")
	sb.WriteString("  1. TODO: Add fix steps.\n")
	sb.WriteString("\n")

	sb.WriteString("validation: |\n")
	sb.WriteString("  ## Validation\n")
	sb.WriteString("\n")
	sb.WriteString("  - TODO: Describe how to confirm the fix worked.\n")
	sb.WriteString("\n")

	sb.WriteString("workflow:\n")
	sb.WriteString("  likely_files:\n")
	sb.WriteString("    - TODO\n")
	sb.WriteString("  local_repro:\n")
	sb.WriteString("    - TODO\n")
	sb.WriteString("  verify:\n")
	sb.WriteString("    - TODO\n")

	return sb.String()
}

// yamlScalar encodes a single pattern string as a safe YAML scalar.
// Uses yaml.Marshal under the hood to get correct quoting, then strips the
// trailing newline that yaml.Marshal appends.
func yamlScalar(s string) string {
	b, err := yaml.Marshal(s)
	if err != nil {
		// Only fails for types that yaml.Marshal cannot handle; strings never fail.
		return fmt.Sprintf("%q", s)
	}
	return strings.TrimRight(string(b), "\n")
}
