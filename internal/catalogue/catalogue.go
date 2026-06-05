// Package catalogue generates a static-site-ready export of the Faultline
// failure catalogue from the bundled playbooks.
//
// The export produces:
//
//   - catalogue/failures/<slug>.md       — one Markdown file per playbook, with
//     Astro content-collection frontmatter
//   - catalogue/catalogue.json           — full index of all entries
//   - catalogue/catalogue.manifest.json  — provenance and generation metadata
package catalogue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"faultline/internal/model"

	"gopkg.in/yaml.v3"
)

// knownEcosystems is the set of tag values recognised as public-facing
// ecosystem identifiers for the failure catalogue.
var knownEcosystems = map[string]bool{
	"aws":            true,
	"azure":          true,
	"cargo":          true,
	"docker":         true,
	"gcp":            true,
	"github-actions": true,
	"gitlab-ci":      true,
	"go":             true,
	"gradle":         true,
	"java":           true,
	"kubernetes":     true,
	"maven":          true,
	"node":           true,
	"npm":            true,
	"pnpm":           true,
	"python":         true,
	"rust":           true,
	"terraform":      true,
	"yarn":           true,
}

// slugRE matches a valid URL-safe slug: lowercase letters, digits, and
// hyphens; must start and end with a letter or digit.
var slugRE = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// ValidateSlug returns a non-nil error when s is not a valid catalogue slug.
func ValidateSlug(s string) error {
	if s == "" {
		return fmt.Errorf("slug must not be empty")
	}
	if !slugRE.MatchString(s) {
		return fmt.Errorf("slug %q is not URL-safe: use only lowercase letters, digits, and hyphens", s)
	}
	return nil
}

// SlugFromID derives a URL-safe catalogue slug from a playbook ID.
// Dots and underscores are converted to hyphens; the result is lowercased.
func SlugFromID(id string) string {
	s := strings.ToLower(id)
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

// EcosystemsFromTags filters the supplied playbook tags, returning only those
// that are in the recognised ecosystem set, deduplicated and sorted.
func EcosystemsFromTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := []string{}
	for _, t := range tags {
		key := strings.ToLower(strings.TrimSpace(t))
		if !knownEcosystems[key] {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

// ConfidenceFromSeverity maps a playbook severity value to a catalogue
// confidence level.  Unknown values map to "medium".
func ConfidenceFromSeverity(severity string) string {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical", "high":
		return "high"
	case "medium":
		return "medium"
	case "low":
		return "low"
	default:
		return "medium"
	}
}

// Entry is one failure entry in the public catalogue.
type Entry struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	FailureID   string   `json:"failure_id"`
	Category    string   `json:"category"`
	Ecosystems  []string `json:"ecosystems"`
	Signals     []string `json:"signals"`
	Confidence  string   `json:"confidence"`
	SourcePath  string   `json:"source_path,omitempty"`
}

// Manifest describes the provenance of a generated catalogue export.
type Manifest struct {
	SourceRepo       string `json:"source_repo"`
	SourceCommit     string `json:"source_commit"`
	GeneratedAt      string `json:"generated_at"`
	FailureCount     int    `json:"failure_count"`
	GeneratorVersion string `json:"generator_version,omitempty"`
}

// ExportOptions configures a catalogue export run.
type ExportOptions struct {
	// SrcDir is the root of the playbook tree to scan.
	// Defaults to "playbooks/bundled" relative to the working directory.
	SrcDir string

	// OutDir is the destination directory for the generated catalogue files.
	OutDir string

	// SourceRepo is stamped into the manifest (e.g. "org/repo").
	SourceRepo string

	// SourceCommit is stamped into the manifest (e.g. a 40-char git SHA).
	SourceCommit string

	// GeneratorVersion is the CLI version string, stamped into the manifest.
	GeneratorVersion string

	// GeneratedAt is the timestamp stamped into the manifest. When empty, the
	// exporter derives a deterministic timestamp from SOURCE_DATE_EPOCH, then
	// the source commit time, then the Unix epoch.
	GeneratedAt string
}

// Export generates the full catalogue export into opts.OutDir.
// It is safe to call multiple times; existing files are overwritten.
func Export(opts ExportOptions) error {
	if opts.SrcDir == "" {
		opts.SrcDir = "playbooks/bundled"
	}
	pbs, err := loadPlaybooks(opts.SrcDir)
	if err != nil {
		return fmt.Errorf("load playbooks: %w", err)
	}
	files, err := generateAll(pbs, opts)
	if err != nil {
		return fmt.Errorf("generate catalogue: %w", err)
	}
	for _, f := range files {
		fullPath := filepath.Join(opts.OutDir, filepath.FromSlash(f.relPath))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, f.content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", fullPath, err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Internal generation types
// ---------------------------------------------------------------------------

type catalogueFile struct {
	relPath string // relative to OutDir
	content []byte
}

// sourcePlaybook pairs a parsed playbook with its source path.
type sourcePlaybook struct {
	model.Playbook
	// SourceRel is the path of the YAML file relative to the repo root.
	SourceRel string
}

// ---------------------------------------------------------------------------
// Loading
// ---------------------------------------------------------------------------

func loadPlaybooks(srcDir string) ([]sourcePlaybook, error) {
	var out []sourcePlaybook
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
			return nil // skip malformed/meta files
		}
		rel, relErr := filepath.Rel(".", path)
		if relErr != nil {
			rel = path
		}
		out = append(out, sourcePlaybook{
			Playbook:  pb,
			SourceRel: filepath.ToSlash(rel),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Entry building
// ---------------------------------------------------------------------------

// BuildEntries converts the loaded playbooks into catalogue Entry values.
// The result is sorted deterministically by category then slug.
func BuildEntries(pbs []sourcePlaybook) []Entry {
	entries := make([]Entry, 0, len(pbs))
	for _, pb := range pbs {
		if pb.ID == "" {
			continue
		}
		slug := SlugFromID(pb.ID)
		if err := ValidateSlug(slug); err != nil {
			continue // skip any ID that cannot produce a valid slug
		}
		entries = append(entries, Entry{
			Slug:        slug,
			Title:       pb.Title,
			Description: descriptionFromSummary(pb.Summary),
			FailureID:   pb.ID,
			Category:    pb.Category,
			Ecosystems:  EcosystemsFromTags(pb.Tags),
			Signals:     topSignals(pb.Match.Any, 8),
			Confidence:  ConfidenceFromSeverity(pb.Severity),
			SourcePath:  pb.SourceRel,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Category != entries[j].Category {
			return entries[i].Category < entries[j].Category
		}
		return entries[i].Slug < entries[j].Slug
	})
	return entries
}

// descriptionFromSummary returns the first sentence of a summary string,
// stripped of Markdown heading prefixes and backtick spans.  Returns an
// empty string when the summary is empty.
func descriptionFromSummary(summary string) string {
	s := strings.TrimSpace(summary)
	if s == "" {
		return ""
	}
	// Strip leading Markdown heading.
	if strings.HasPrefix(s, "#") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = strings.TrimSpace(s[idx+1:])
		}
	}
	// Find the first sentence boundary: ., !, or ? followed by whitespace,
	// end-of-string, or a closing punctuation character.  This avoids
	// splitting on periods inside tokens like `package-lock.json`.
	for i, ch := range s {
		if ch != '.' && ch != '!' && ch != '?' {
			continue
		}
		after := i + 1
		// Sentence ends at the very end of the string.
		if after >= len(s) {
			// Collapse any embedded newlines from word-wrapped YAML source.
			return strings.Join(strings.Fields(s[:after]), " ")
		}
		// Sentence ends when followed by whitespace or closing punctuation.
		next := rune(s[after])
		if next == ' ' || next == '\t' || next == '\n' || next == '"' || next == ')' || next == ']' {
			// Collapse any embedded newlines from word-wrapped YAML source.
			candidate := strings.Join(strings.Fields(s[:after]), " ")
			if len(candidate) <= 200 {
				return candidate
			}
		}
	}
	// No clear sentence boundary — truncate at word boundary.
	words := strings.Fields(s)
	if len(words) > 25 {
		words = words[:25]
	}
	return strings.Join(words, " ")
}

// topSignals returns up to n signals, trimmed of whitespace.
func topSignals(signals []string, n int) []string {
	out := make([]string, 0, min(n, len(signals)))
	for _, s := range signals {
		if len(out) >= n {
			break
		}
		s = escapeControlChars(strings.TrimSpace(s))
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func escapeControlChars(s string) string {
	var buf strings.Builder
	for _, r := range s {
		switch r {
		case '\a':
			buf.WriteString(`\a`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		case '\v':
			buf.WriteString(`\v`)
		default:
			if r < 0x20 || r == 0x7f {
				fmt.Fprintf(&buf, `\u%04x`, r)
				continue
			}
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// File generation
// ---------------------------------------------------------------------------

func generateAll(pbs []sourcePlaybook, opts ExportOptions) ([]catalogueFile, error) {
	sorted := make([]sourcePlaybook, len(pbs))
	copy(sorted, pbs)
	sort.Slice(sorted, func(i, j int) bool {
		ci, cj := sorted[i].Category, sorted[j].Category
		if ci != cj {
			return ci < cj
		}
		return SlugFromID(sorted[i].ID) < SlugFromID(sorted[j].ID)
	})

	entries := BuildEntries(sorted)

	var files []catalogueFile

	// Per-failure Markdown pages.
	for i, pb := range sorted {
		slug := SlugFromID(pb.ID)
		if err := ValidateSlug(slug); err != nil {
			continue
		}
		content, err := renderFailureMarkdown(pb, entries[i])
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", pb.ID, err)
		}
		files = append(files, catalogueFile{
			relPath: "failures/" + slug + ".md",
			content: content,
		})
	}

	// catalogue.json
	indexContent, err := renderCatalogueJSON(entries)
	if err != nil {
		return nil, err
	}
	files = append(files, catalogueFile{
		relPath: "catalogue.json",
		content: indexContent,
	})

	// catalogue.manifest.json
	manifest := Manifest{
		SourceRepo:       opts.SourceRepo,
		SourceCommit:     opts.SourceCommit,
		GeneratedAt:      generatedAt(opts),
		FailureCount:     len(entries),
		GeneratorVersion: opts.GeneratorVersion,
	}
	manifestContent, err := renderManifestJSON(manifest)
	if err != nil {
		return nil, err
	}
	files = append(files, catalogueFile{
		relPath: "catalogue.manifest.json",
		content: manifestContent,
	})

	return files, nil
}

// ---------------------------------------------------------------------------
// Frontmatter and Markdown rendering
// ---------------------------------------------------------------------------

// RenderFrontmatter returns the YAML frontmatter block (including the ---
// delimiters) for a catalogue entry, formatted for Astro content collections.
func RenderFrontmatter(e Entry) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("---\n")
	fmt.Fprintf(&buf, "title: %s\n", yamlQuote(e.Title))
	fmt.Fprintf(&buf, "description: %s\n", yamlQuote(e.Description))
	fmt.Fprintf(&buf, "failure_id: %s\n", yamlQuote(e.FailureID))
	fmt.Fprintf(&buf, "category: %s\n", yamlQuote(e.Category))
	// ecosystems as inline JSON array.
	ecoBytes, err := json.Marshal(e.Ecosystems)
	if err != nil {
		return nil, fmt.Errorf("marshal ecosystems: %w", err)
	}
	fmt.Fprintf(&buf, "ecosystems: %s\n", ecoBytes)
	if len(e.Signals) > 0 {
		buf.WriteString("signals:\n")
		for _, s := range e.Signals {
			fmt.Fprintf(&buf, "  - %s\n", yamlQuote(s))
		}
	} else {
		buf.WriteString("signals: []\n")
	}
	fmt.Fprintf(&buf, "confidence: %s\n", yamlQuote(e.Confidence))
	buf.WriteString("---\n")
	return buf.Bytes(), nil
}

// yamlQuote returns a JSON-escaped string literal, which is valid YAML inside
// frontmatter and covers control characters as well as quotes and backslashes.
func yamlQuote(s string) string {
	quoted, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return string(quoted)
}

func renderFailureMarkdown(pb sourcePlaybook, e Entry) ([]byte, error) {
	fm, err := RenderFrontmatter(e)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.Write(fm)
	buf.WriteByte('\n')

	// Title heading.
	fmt.Fprintf(&buf, "# %s\n\n", pb.Title)

	// What this failure means.
	if pb.Summary != "" {
		buf.WriteString("## What this failure means\n\n")
		buf.WriteString(strings.TrimSpace(pb.Summary))
		buf.WriteString("\n\n")
	}

	// Symptoms / common log signals.
	if len(pb.Match.Any) > 0 {
		buf.WriteString("## Symptoms\n\n")
		buf.WriteString("Faultline looks for one or more of these log fragments:\n\n")
		buf.WriteString("```text\n")
		for _, sig := range topSignals(pb.Match.Any, 8) {
			buf.WriteString(sig)
			buf.WriteByte('\n')
		}
		buf.WriteString("```\n\n")
	}

	// Diagnosis / common causes.
	if pb.Diagnosis != "" {
		buf.WriteString(strings.TrimSpace(pb.Diagnosis))
		buf.WriteString("\n\n")
	}

	// Fix.
	if pb.Fix != "" {
		buf.WriteString(strings.TrimSpace(pb.Fix))
		buf.WriteString("\n\n")
	}

	// Validation.
	if pb.Validation != "" {
		buf.WriteString(strings.TrimSpace(pb.Validation))
		buf.WriteString("\n\n")
	}

	// Why it matters.
	if pb.WhyItMatters != "" {
		buf.WriteString(strings.TrimSpace(pb.WhyItMatters))
		buf.WriteString("\n\n")
	}

	// Try it locally.
	if len(pb.Workflow.LocalRepro) > 0 || len(pb.Workflow.Verify) > 0 {
		buf.WriteString("## Try it locally\n\n")
		cmds := append(append([]string(nil), pb.Workflow.LocalRepro...), pb.Workflow.Verify...)
		buf.WriteString("```bash\n")
		for _, cmd := range cmds {
			buf.WriteString(cmd)
			buf.WriteByte('\n')
		}
		buf.WriteString("```\n\n")
	}

	// How Faultline detects it.
	buf.WriteString("## How Faultline detects it\n\n")
	fmt.Fprintf(&buf, "Use [`faultline explain %s`](https://faultline.dev/failures/%s) to see the full playbook.\n\n", pb.ID, e.Slug)
	buf.WriteString("```bash\n")
	buf.WriteString("faultline analyze build.log\n")
	fmt.Fprintf(&buf, "faultline explain %s\n", pb.ID)
	buf.WriteString("```\n\n")

	// Footer.
	if pb.SourceRel != "" {
		fmt.Fprintf(&buf, "---\n\n*Generated from [`%s`](%s). Do not edit directly.*\n", pb.SourceRel, "../"+pb.SourceRel)
	}

	return buf.Bytes(), nil
}

// ---------------------------------------------------------------------------
// JSON rendering
// ---------------------------------------------------------------------------

// BuildIndex returns the full catalogue index as a list of Entry values,
// sorted by category then slug.
func BuildIndex(entries []Entry) []Entry {
	out := make([]Entry, len(entries))
	copy(out, entries)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return out[i].Slug < out[j].Slug
	})
	return out
}

func renderCatalogueJSON(entries []Entry) ([]byte, error) {
	sorted := BuildIndex(entries)
	data, err := json.MarshalIndent(sorted, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal catalogue.json: %w", err)
	}
	return append(data, '\n'), nil
}

// BuildManifest constructs the manifest value for the export.
func BuildManifest(opts ExportOptions, failureCount int) Manifest {
	return Manifest{
		SourceRepo:       opts.SourceRepo,
		SourceCommit:     opts.SourceCommit,
		GeneratedAt:      generatedAt(opts),
		FailureCount:     failureCount,
		GeneratorVersion: opts.GeneratorVersion,
	}
}

func generatedAt(opts ExportOptions) string {
	if opts.GeneratedAt != "" {
		return opts.GeneratedAt
	}
	if epoch := strings.TrimSpace(os.Getenv("SOURCE_DATE_EPOCH")); epoch != "" {
		if seconds, err := strconv.ParseInt(epoch, 10, 64); err == nil {
			return time.Unix(seconds, 0).UTC().Format(time.RFC3339)
		}
	}
	if opts.SourceCommit != "" {
		if out, err := exec.Command("git", "show", "-s", "--format=%cI", opts.SourceCommit).Output(); err == nil {
			if ts := strings.TrimSpace(string(out)); ts != "" {
				if parsed, parseErr := time.Parse(time.RFC3339, ts); parseErr == nil {
					return parsed.UTC().Format(time.RFC3339)
				}
			}
		}
	}
	return time.Unix(0, 0).UTC().Format(time.RFC3339)
}

func renderManifestJSON(m Manifest) ([]byte, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal catalogue.manifest.json: %w", err)
	}
	return append(data, '\n'), nil
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

// ValidateEntry returns a non-nil error when an entry is missing required
// fields or has an invalid slug.
func ValidateEntry(e Entry) error {
	if err := ValidateSlug(e.Slug); err != nil {
		return err
	}
	if e.FailureID == "" {
		return fmt.Errorf("entry %q: failure_id must not be empty", e.Slug)
	}
	if e.Title == "" {
		return fmt.Errorf("entry %q: title must not be empty", e.Slug)
	}
	if e.Description == "" {
		return fmt.Errorf("entry %q: description must not be empty", e.Slug)
	}
	return nil
}

// ValidateEntries runs ValidateEntry on every entry and returns the first
// error encountered.
func ValidateEntries(entries []Entry) error {
	slugs := make(map[string]int, len(entries))
	for i, e := range entries {
		if err := ValidateEntry(e); err != nil {
			return err
		}
		slugs[e.Slug]++
		if slugs[e.Slug] > 1 {
			return fmt.Errorf("duplicate slug %q at index %d", e.Slug, i)
		}
	}
	return nil
}

// ValidateJSON returns nil when data is well-formed JSON.
func ValidateJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}
