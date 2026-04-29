// gen-failure-docs generates a crawlable failure catalog from the bundled
// playbooks. It writes one Markdown page per playbook and a catalog index.
//
// Usage:
//
//	go run ./tools/gen-failure-docs [--check] [--src <playbooks/bundled>] [--dst <docs/failures>]
//
// --check mode generates the content in memory and exits non-zero if any
// on-disk file differs from what would be generated.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------------------
// Minimal playbook schema – only the fields used for doc generation.
// ---------------------------------------------------------------------------

type Playbook struct {
	ID           string       `yaml:"id"`
	Title        string       `yaml:"title"`
	Category     string       `yaml:"category"`
	Severity     string       `yaml:"severity"`
	Tags         []string     `yaml:"tags"`
	StageHints   []string     `yaml:"stage_hints"`
	Domain       string       `yaml:"domain"`
	Class        string       `yaml:"class"`
	Summary      string       `yaml:"summary"`
	Diagnosis    string       `yaml:"diagnosis"`
	Fix          string       `yaml:"fix"`
	Validation   string       `yaml:"validation"`
	WhyItMatters string       `yaml:"why_it_matters"`
	Match        MatchSpec    `yaml:"match"`
	Workflow     WorkflowSpec `yaml:"workflow"`
	// Derived – set after loading.
	SourceRel string // relative path from repo root, e.g. playbooks/bundled/log/build/npm-ci-lockfile.yaml
	CatDir    string // normalized category for use as a directory slug, e.g. silent-failure
}

type MatchSpec struct {
	Any []string `yaml:"any"`
	All []string `yaml:"all"`
}

type WorkflowSpec struct {
	LikelyFiles []string `yaml:"likely_files"`
	LocalRepro  []string `yaml:"local_repro"`
	Verify      []string `yaml:"verify"`
}

// ---------------------------------------------------------------------------
// Templates
// ---------------------------------------------------------------------------

const pageTemplateText = `# {{ .Title }}

**Playbook ID:** ` + "`{{ .ID }}`" + `
{{ if .Category }}**Category:** {{ .Category }}{{ end }}
{{ if .Severity }}**Severity:** {{ .Severity }}{{ end }}
{{ if .Tags }}**Tags:** {{ joinTags .Tags }}{{ end }}

## What this failure means

{{ or .Summary "(No summary provided.)" }}

## Common log signals

{{ if .Match.Any -}}
` + "```text" + `
{{ range limitStrings .Match.Any 8 -}}
{{ . }}
{{ end -}}
` + "```" + `
{{ else -}}
*(This playbook uses source-code pattern matching rather than log signals.)*
{{ end }}
{{ if .Diagnosis -}}
{{ .Diagnosis }}

{{ end -}}
{{ if .Fix -}}
{{ .Fix }}

{{ end -}}
{{ if .Validation -}}
{{ .Validation }}

{{ end -}}
## Likely files to inspect

{{ if .Workflow.LikelyFiles -}}
{{ range .Workflow.LikelyFiles -}}
- ` + "`{{ . }}`" + `
{{ end -}}
{{ else -}}
*(Not specified.)*
{{ end }}

## Run Faultline

` + "```bash" + `
faultline analyze build.log
faultline explain {{ .ID }}
faultline workflow build.log --json --mode agent
` + "```" + `

## Search phrases this page answers

{{ range searchPhrases . -}}
- {{ . }}
{{ end }}

---

*Generated from [{{ .SourceRel }}]({{ relPlaybookLink .SourceRel .CatDir }}). Do not edit directly — run ` + "`make docs-generate`" + `.*
`

const catalogTemplateText = `# Faultline Failure Catalog

This index links to every bundled playbook. Each page explains the failure, shows the log signals Faultline matches, and lists diagnosis and fix steps.

**{{ .Total }} playbooks** across {{ .CategoryCount }} categories.

Run ` + "`faultline list`" + ` to browse the catalog inside the terminal.

{{ range .Groups -}}
## {{ catTitle .Name }} ({{ len .Playbooks }})

{{ range .Playbooks -}}
- [{{ .Title }}](../{{ .CatDir }}/{{ .ID }}.md){{ if .Summary }} — {{ firstSentence .Summary }}{{ end }}
{{ end }}
{{ end }}

---

*Generated from ` + "`playbooks/bundled/`" + `. Do not edit directly — run ` + "`make docs-generate`" + `.*
`

// ---------------------------------------------------------------------------
// Template helpers
// ---------------------------------------------------------------------------

var tmplFuncs = template.FuncMap{
	"joinTags": func(tags []string) string {
		parts := make([]string, len(tags))
		for i, t := range tags {
			parts[i] = "`" + t + "`"
		}
		return strings.Join(parts, ", ")
	},
	"limitStrings": func(ss []string, n int) []string {
		if len(ss) <= n {
			return ss
		}
		return ss[:n]
	},
	"searchPhrases": searchPhrases,
	"firstSentence": firstSentence,
	"catTitle":      catTitle,
	"relPlaybookLink": func(sourceRel, catDir string) string {
		// From docs/failures/<catDir>/<id>.md we need ../../playbooks/...
		return "../../" + sourceRel
	},
}

var pageTmpl = template.Must(template.New("page").Funcs(tmplFuncs).Parse(pageTemplateText))
var catalogTmpl = template.Must(template.New("catalog").Funcs(tmplFuncs).Parse(catalogTemplateText))

// searchPhrases generates 3–6 natural search phrases from the playbook fields.
func searchPhrases(p Playbook) []string {
	phrases := make([]string, 0, 6)

	// 1. Exact title as-is.
	if p.Title != "" {
		phrases = append(phrases, p.Title)
	}

	// 2. Category + title compound.
	if p.Category != "" && p.Title != "" {
		phrases = append(phrases, catTitle(p.Category)+": "+strings.ToLower(p.Title))
	}

	// 3. Strongest match signal (longest match.any string, max 80 chars).
	if bestSignal := longestSignal(p.Match.Any); bestSignal != "" {
		phrases = append(phrases, bestSignal)
	}

	// 4. "GitHub Actions <title>" if the playbook is CI/build related.
	if containsAny(p.Tags, "github-actions", "ci", "actions") ||
		p.Category == "ci" || p.Category == "build" {
		if p.Title != "" {
			phrases = append(phrases, "GitHub Actions "+strings.ToLower(p.Title))
		}
	}

	// 5. faultline explain <id>
	if p.ID != "" {
		phrases = append(phrases, "faultline explain "+p.ID)
	}

	// 6. Tag-specific platform phrase e.g. "Docker <title>" or "npm <title>".
	for _, tag := range p.Tags {
		platform := platformLabel(tag)
		if platform != "" && p.Title != "" {
			candidate := platform + " " + strings.ToLower(p.Title)
			if !containsPhrase(phrases, candidate) {
				phrases = append(phrases, candidate)
				break
			}
		}
	}

	// Remove duplicates and cap to 6.
	phrases = dedupe(phrases)
	if len(phrases) > 6 {
		phrases = phrases[:6]
	}
	return phrases
}

func longestSignal(signals []string) string {
	best := ""
	for _, s := range signals {
		s = strings.TrimSpace(s)
		if len(s) > len(best) && len(s) <= 80 {
			best = s
		}
	}
	return best
}

func platformLabel(tag string) string {
	m := map[string]string{
		"docker":         "Docker",
		"npm":            "npm",
		"node":           "Node.js",
		"python":         "Python",
		"go":             "Go",
		"golang":         "Go",
		"rust":           "Rust",
		"cargo":          "Cargo",
		"java":           "Java",
		"maven":          "Maven",
		"gradle":         "Gradle",
		"kubernetes":     "Kubernetes",
		"terraform":      "Terraform",
		"github-actions": "GitHub Actions",
		"gitlab-ci":      "GitLab CI",
		"pnpm":           "pnpm",
		"yarn":           "Yarn",
	}
	return m[strings.ToLower(tag)]
}

func containsAny(slice []string, vals ...string) bool {
	for _, s := range slice {
		for _, v := range vals {
			if strings.EqualFold(s, v) {
				return true
			}
		}
	}
	return false
}

func containsPhrase(phrases []string, target string) bool {
	for _, p := range phrases {
		if strings.EqualFold(p, target) {
			return true
		}
	}
	return false
}

func dedupe(ss []string) []string {
	seen := make(map[string]struct{}, len(ss))
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		key := strings.ToLower(strings.TrimSpace(s))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}
	return out
}

func firstSentence(s string) string {
	s = strings.TrimSpace(s)
	// Strip markdown heading prefix if summary starts with "## ..."
	if strings.HasPrefix(s, "#") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = strings.TrimSpace(s[idx+1:])
		}
	}
	// Strip leading backtick-quoted word and find end of first sentence.
	end := strings.IndexAny(s, ".!?")
	if end == -1 || end > 150 {
		// No sentence boundary or very long – truncate at word boundary.
		words := strings.Fields(s)
		if len(words) > 20 {
			words = words[:20]
		}
		return strings.Join(words, " ")
	}
	return strings.TrimSpace(s[:end+1])
}

func catTitle(name string) string {
	title := strings.ReplaceAll(name, "-", " ")
	title = strings.ReplaceAll(title, "_", " ")
	if title == "" {
		return ""
	}
	// Title-case the first letter of each word.
	words := strings.Fields(title)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// normCatDir converts a category string to a directory-safe slug (lowercase, hyphens).
func normCatDir(category string) string {
	return strings.ToLower(strings.ReplaceAll(category, "_", "-"))
}

// ---------------------------------------------------------------------------
// Loading
// ---------------------------------------------------------------------------

func loadPlaybooks(srcDir string) ([]Playbook, error) {
	var playbooks []Playbook

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}

		var pb Playbook
		if decodeErr := yaml.Unmarshal(data, &pb); decodeErr != nil {
			return fmt.Errorf("parse %s: %w", path, decodeErr)
		}
		if pb.ID == "" {
			return nil // skip malformed
		}

		// Derive relative source path for the "generated from" footer.
		// We want a path relative to the repo root. srcDir is typically
		// "playbooks/bundled" relative to cwd (the repo root).
		rel, relErr := filepath.Rel(".", path)
		if relErr != nil {
			rel = path
		}
		pb.SourceRel = filepath.ToSlash(rel)
		pb.CatDir = normCatDir(pb.Category)

		playbooks = append(playbooks, pb)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return playbooks, nil
}

// ---------------------------------------------------------------------------
// Generation
// ---------------------------------------------------------------------------

// generatedFile is the in-memory representation of a file to be written.
type generatedFile struct {
	RelPath string // relative to dstDir
	Content []byte
}

func generateAll(playbooks []Playbook) ([]generatedFile, error) {
	// Sort deterministically: category then id.
	sorted := make([]Playbook, len(playbooks))
	copy(sorted, playbooks)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].CatDir != sorted[j].CatDir {
			return sorted[i].CatDir < sorted[j].CatDir
		}
		return sorted[i].ID < sorted[j].ID
	})

	var files []generatedFile

	// Per-playbook pages.
	for _, pb := range sorted {
		if pb.CatDir == "" || pb.ID == "" {
			continue
		}
		var buf bytes.Buffer
		if err := pageTmpl.Execute(&buf, pb); err != nil {
			return nil, fmt.Errorf("render page %s: %w", pb.ID, err)
		}
		files = append(files, generatedFile{
			RelPath: pb.CatDir + "/" + pb.ID + ".md",
			Content: buf.Bytes(),
		})
	}

	// Catalog index.
	catContent, err := renderCatalog(sorted)
	if err != nil {
		return nil, err
	}
	files = append(files, generatedFile{
		RelPath: "catalog/README.md",
		Content: catContent,
	})

	return files, nil
}

type catalogGroup struct {
	Name      string
	Playbooks []Playbook
}

type catalogData struct {
	Total         int
	CategoryCount int
	Groups        []catalogGroup
}

func renderCatalog(sorted []Playbook) ([]byte, error) {
	groupMap := make(map[string][]Playbook)
	for _, pb := range sorted {
		groupMap[pb.CatDir] = append(groupMap[pb.CatDir], pb)
	}

	cats := make([]string, 0, len(groupMap))
	for cat := range groupMap {
		cats = append(cats, cat)
	}
	sort.Strings(cats)

	groups := make([]catalogGroup, 0, len(cats))
	for _, cat := range cats {
		groups = append(groups, catalogGroup{
			Name:      cat,
			Playbooks: groupMap[cat],
		})
	}

	data := catalogData{
		Total:         len(sorted),
		CategoryCount: len(groups),
		Groups:        groups,
	}

	var buf bytes.Buffer
	if err := catalogTmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render catalog: %w", err)
	}
	return buf.Bytes(), nil
}

// ---------------------------------------------------------------------------
// Write / check
// ---------------------------------------------------------------------------

func writeFiles(dstDir string, files []generatedFile) error {
	for _, f := range files {
		fullPath := filepath.Join(dstDir, filepath.FromSlash(f.RelPath))
		if mkErr := os.MkdirAll(filepath.Dir(fullPath), 0o755); mkErr != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(fullPath), mkErr)
		}
		if writeErr := os.WriteFile(fullPath, f.Content, 0o644); writeErr != nil {
			return fmt.Errorf("write %s: %w", fullPath, writeErr)
		}
	}
	return nil
}

func checkFiles(dstDir string, files []generatedFile) (stale []string, err error) {
	for _, f := range files {
		fullPath := filepath.Join(dstDir, filepath.FromSlash(f.RelPath))
		existing, readErr := os.ReadFile(fullPath)
		if errors.Is(readErr, os.ErrNotExist) {
			stale = append(stale, f.RelPath+" (missing)")
			continue
		}
		if readErr != nil {
			return nil, fmt.Errorf("read %s: %w", fullPath, readErr)
		}
		if !bytes.Equal(existing, f.Content) {
			stale = append(stale, f.RelPath+" (out of date)")
		}
	}
	return stale, nil
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	srcFlag := flag.String("src", "playbooks/bundled", "root of bundled playbook tree")
	dstFlag := flag.String("dst", "docs/failures", "root of failure docs output tree")
	checkFlag := flag.Bool("check", false, "verify docs are up to date without writing")
	flag.Parse()

	playbooks, err := loadPlaybooks(*srcFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gen-failure-docs: load playbooks: %v\n", err)
		os.Exit(1)
	}

	files, err := generateAll(playbooks)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gen-failure-docs: generate: %v\n", err)
		os.Exit(1)
	}

	if *checkFlag {
		stale, checkErr := checkFiles(*dstFlag, files)
		if checkErr != nil {
			fmt.Fprintf(os.Stderr, "gen-failure-docs: check: %v\n", checkErr)
			os.Exit(1)
		}
		if len(stale) > 0 {
			fmt.Fprintln(os.Stderr, "gen-failure-docs: generated docs are out of date:")
			for _, s := range stale {
				fmt.Fprintf(os.Stderr, "  %s\n", s)
			}
			fmt.Fprintln(os.Stderr, "Run 'make docs-generate' and commit the changes.")
			os.Exit(1)
		}
		fmt.Printf("gen-failure-docs: %d files are up to date\n", len(files))
		return
	}

	if writeErr := writeFiles(*dstFlag, files); writeErr != nil {
		fmt.Fprintf(os.Stderr, "gen-failure-docs: write: %v\n", writeErr)
		os.Exit(1)
	}
	fmt.Printf("gen-failure-docs: wrote %d files to %s\n", len(files), *dstFlag)
}
