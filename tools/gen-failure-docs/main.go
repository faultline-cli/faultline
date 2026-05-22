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

	"faultline/internal/model"

	"gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------------------
// Doc-generation playbook type – embeds the canonical model.Playbook and
// adds two derived fields used only during doc generation.
// ---------------------------------------------------------------------------

// docPlaybook wraps model.Playbook with the two derived fields that are
// specific to the doc-generation tool and have no place in the core model.
type docPlaybook struct {
	model.Playbook
	SourceRel string // relative path from repo root, e.g. playbooks/bundled/log/build/npm-ci-lockfile.yaml
	CatDir    string // normalized category for use as a directory slug, e.g. silent-failure
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

const llmsTxtTemplateText = `# Faultline Failure Catalog

> {{ .Total }} bundled CI failure playbooks across {{ .CategoryCount }} categories. Each entry covers cause, log signals, diagnosis, fix steps, and the Faultline playbook ID used to detect it.

Generated from ` + "`playbooks/bundled/`" + `. Run ` + "`make docs-generate`" + ` to rebuild after adding or modifying playbooks.

## Index

- [All {{ .Total }} playbooks by category](catalog/README.md): Full generated catalog index with one-line descriptions

{{ range .Groups -}}
## {{ catTitle .Name }} ({{ len .Playbooks }})

{{ range .Playbooks -}}
- [{{ .Title }}]({{ .CatDir }}/{{ .ID }}.md){{ if .Summary }}: {{ firstSentence .Summary }}{{ end }}
{{ end }}
{{ end -}}
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
	"searchPhrases":   searchPhrases,
	"firstSentence":   firstSentence,
	"catTitle":        catTitle,
	"relPlaybookLink": relPlaybookLink,
}

var pageTmpl = template.Must(template.New("page").Funcs(tmplFuncs).Parse(pageTemplateText))
var catalogTmpl = template.Must(template.New("catalog").Funcs(tmplFuncs).Parse(catalogTemplateText))
var llmsTxt = template.Must(template.New("llms-txt").Funcs(tmplFuncs).Parse(llmsTxtTemplateText))

// searchPhrases generates 3–6 natural search phrases from the playbook fields.
func searchPhrases(p docPlaybook) []string {
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

func relPlaybookLink(sourceRel, catDir string) string {
	pageRelPath := filepath.ToSlash(filepath.Join(catDir, "playbook.md"))
	pageDir := filepath.Dir(pageRelPath)
	if pageDir == "." {
		return "../../" + sourceRel
	}
	depth := len(strings.Split(pageDir, "/"))
	return strings.Repeat("../", depth+2) + sourceRel
}

// normCatDir converts a category string to a directory-safe slug (lowercase, hyphens).
func normCatDir(category string) string {
	return strings.ToLower(strings.ReplaceAll(category, "_", "-"))
}

// ---------------------------------------------------------------------------
// Loading
// ---------------------------------------------------------------------------

func loadPlaybooks(srcDir string) ([]docPlaybook, error) {
	var playbooks []docPlaybook

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

		var pb model.Playbook
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
		playbooks = append(playbooks, docPlaybook{
			Playbook:  pb,
			SourceRel: filepath.ToSlash(rel),
			CatDir:    normCatDir(pb.Category),
		})
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

func generateAll(playbooks []docPlaybook) ([]generatedFile, error) {
	// Sort deterministically: category then id.
	sorted := make([]docPlaybook, len(playbooks))
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

	// llms.txt for the failure catalog subtree.
	llmsContent, err := renderLLMsTxt(sorted)
	if err != nil {
		return nil, err
	}
	files = append(files, generatedFile{
		RelPath: "llms.txt",
		Content: llmsContent,
	})

	return files, nil
}

type catalogGroup struct {
	Name      string
	Playbooks []docPlaybook
}

type catalogData struct {
	Total         int
	CategoryCount int
	Groups        []catalogGroup
}

func renderCatalog(sorted []docPlaybook) ([]byte, error) {
	groupMap := make(map[string][]docPlaybook)
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

func renderLLMsTxt(sorted []docPlaybook) ([]byte, error) {
	groupMap := make(map[string][]docPlaybook)
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
	if err := llmsTxt.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render llms.txt: %w", err)
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
	if pruneErr := pruneOrphanedDocs(dstDir, files); pruneErr != nil {
		return pruneErr
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
	orphans, orphanErr := findOrphanedDocs(dstDir, files)
	if orphanErr != nil {
		return nil, orphanErr
	}
	for _, rel := range orphans {
		stale = append(stale, rel+" (stale generated doc)")
	}
	return stale, nil
}

func pruneOrphanedDocs(dstDir string, files []generatedFile) error {
	orphans, err := findOrphanedDocs(dstDir, files)
	if err != nil {
		return err
	}
	for _, rel := range orphans {
		fullPath := filepath.Join(dstDir, filepath.FromSlash(rel))
		if removeErr := os.Remove(fullPath); removeErr != nil {
			return fmt.Errorf("remove stale doc %s: %w", fullPath, removeErr)
		}
	}
	return removeEmptyDirs(dstDir)
}

func findOrphanedDocs(dstDir string, files []generatedFile) ([]string, error) {
	if _, err := os.Stat(dstDir); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("stat docs dir %s: %w", dstDir, err)
	}

	generated := generatedFileSet(files)
	var orphans []string

	err := filepath.WalkDir(dstDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, relErr := filepath.Rel(dstDir, path)
		if relErr != nil {
			return fmt.Errorf("rel %s: %w", path, relErr)
		}
		rel = filepath.ToSlash(rel)
		if _, ok := generated[rel]; ok || isManualDoc(rel) {
			return nil
		}
		orphans = append(orphans, rel)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk generated docs: %w", err)
	}
	sort.Strings(orphans)
	return orphans, nil
}

func generatedFileSet(files []generatedFile) map[string]struct{} {
	generated := make(map[string]struct{}, len(files))
	for _, f := range files {
		rel := filepath.ToSlash(filepath.Clean(filepath.FromSlash(f.RelPath)))
		generated[rel] = struct{}{}
	}
	return generated
}

func isManualDoc(rel string) bool {
	switch rel {
	case "README.md", "_template.md", "PLAYBOOK_AUTHORING_GUIDE.md":
		return true
	default:
		return false
	}
}

func removeEmptyDirs(dstDir string) error {
	var dirs []string
	err := filepath.WalkDir(dstDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != dstDir {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk docs dirs: %w", err)
	}
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})
	for _, dir := range dirs {
		if removeErr := os.Remove(dir); removeErr != nil && !errors.Is(removeErr, fs.ErrNotExist) {
			if entries, readErr := os.ReadDir(dir); readErr != nil || len(entries) > 0 {
				continue
			}
			return fmt.Errorf("remove empty dir %s: %w", dir, removeErr)
		}
	}
	return nil
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
