package engine

import (
	"regexp"
	"strings"

	"faultline/internal/model"
)

// Stage constants used by context extraction and stage hint matching.
const (
	stageBuild  = "build"
	stageTest   = "test"
	stageDeploy = "deploy"
)

// keyword sets ordered from most specific to least specific.
var (
	deployKeywords = []string{
		"kubectl", "helm", "terraform", "docker push", "docker run",
		"deploy", "release", "publish", "helm upgrade", "kubectl apply",
	}
	testKeywords = []string{
		"go test", "pytest", "jest", "mocha", "rspec", "vitest",
		"phpunit", "cargo test", "npm test", "yarn test", "test",
	}
	buildKeywords = []string{
		"go build", "npm run build", "yarn build", "cargo build",
		"mvn package", "gradle build", "make", "javac", "compile",
		"build",
	}

	// stepPattern matches common CI log step/group header formats.
	// Captures: GitHub Actions ::group::, Azure ##[group], plain "Step N:" lines.
	stepPattern = regexp.MustCompile(
		`(?i)(?:##\[group\]|::group::|step\s+\d+[:/]|##\s*step[:\s])\s*(.+)`,
	)

	// cmdPattern matches shell command lines commonly logged in CI output.
	// Captures: lines preceded by $, >, RUN, or a leading + (bash -x trace).
	cmdPattern = regexp.MustCompile(`^(?:\$\s+|>\s+|RUN\s+|\+\s+)(.+)`)
)

// ExtractContext infers lightweight context from a set of log lines.
// The first definitive match for each field wins; extraction stops early once
// all three fields are populated.
func ExtractContext(lines []model.Line) model.Context {
	var ctx model.Context
	for _, line := range lines {
		if ctx.Stage == "" {
			ctx.Stage = inferStage(line.Normalized)
		}
		if ctx.Step == "" {
			if m := stepPattern.FindStringSubmatch(line.Original); len(m) > 1 {
				ctx.Step = strings.TrimSpace(m[1])
			}
		}
		if ctx.CommandHint == "" {
			if m := cmdPattern.FindStringSubmatch(line.Original); len(m) > 1 {
				ctx.CommandHint = strings.TrimSpace(m[1])
			}
		}
		if ctx.Stage != "" && ctx.Step != "" && ctx.CommandHint != "" {
			break
		}
	}
	return ctx
}

// inferStage returns the most likely CI stage for a single normalised line.
// Deploy keywords take priority over test keywords, which take priority over
// build keywords, because deploy lines often also contain build-related words.
func inferStage(norm string) string {
	for _, kw := range deployKeywords {
		if strings.Contains(norm, kw) {
			return stageDeploy
		}
	}
	for _, kw := range testKeywords {
		if strings.Contains(norm, kw) {
			return stageTest
		}
	}
	for _, kw := range buildKeywords {
		if strings.Contains(norm, kw) {
			return stageBuild
		}
	}
	return ""
}
