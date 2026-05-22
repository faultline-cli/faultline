package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/model"
)

func TestRelPlaybookLinkFromCategoryPage(t *testing.T) {
	got := relPlaybookLink("playbooks/bundled/log/auth/aws-credentials.yaml", "auth")
	want := "../../../playbooks/bundled/log/auth/aws-credentials.yaml"
	if got != want {
		t.Fatalf("relPlaybookLink() = %q, want %q", got, want)
	}
}

func TestRelPlaybookLinkFromNestedCategoryPage(t *testing.T) {
	got := relPlaybookLink("playbooks/bundled/log/silent/artifact-missing.yaml", "silent/failure")
	want := "../../../../playbooks/bundled/log/silent/artifact-missing.yaml"
	if got != want {
		t.Fatalf("relPlaybookLink() = %q, want %q", got, want)
	}
}

func TestCheckFilesReportsOrphanedGeneratedDoc(t *testing.T) {
	dir := t.TempDir()
	files := []generatedFile{
		{RelPath: "build/missing-executable.md", Content: []byte("current\n")},
		{RelPath: "catalog/README.md", Content: []byte("catalog\n")},
	}
	writeDoc(t, dir, "build/missing-executable.md", "current\n")
	writeDoc(t, dir, "catalog/README.md", "catalog\n")
	writeDoc(t, dir, "build/removed-playbook.md", "stale\n")
	writeDoc(t, dir, "README.md", "manual\n")

	stale, err := checkFiles(dir, files)
	if err != nil {
		t.Fatalf("checkFiles: %v", err)
	}
	got := strings.Join(stale, "\n")
	if !strings.Contains(got, "build/removed-playbook.md (stale generated doc)") {
		t.Fatalf("expected stale generated doc, got %q", got)
	}
	if strings.Contains(got, "README.md (stale generated doc)") {
		t.Fatalf("manual README should not be stale, got %q", got)
	}
}

func TestWriteFilesPrunesOrphanedGeneratedDoc(t *testing.T) {
	dir := t.TempDir()
	files := []generatedFile{
		{RelPath: "build/missing-executable.md", Content: []byte("current\n")},
		{RelPath: "catalog/README.md", Content: []byte("catalog\n")},
	}
	writeDoc(t, dir, "build/removed-playbook.md", "stale\n")
	writeDoc(t, dir, "README.md", "manual\n")

	if err := writeFiles(dir, files); err != nil {
		t.Fatalf("writeFiles: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "build", "removed-playbook.md")); !os.IsNotExist(err) {
		t.Fatalf("expected stale doc to be removed, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "README.md")); err != nil {
		t.Fatalf("manual README should remain: %v", err)
	}
}

func writeDoc(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// ── catTitle ──────────────────────────────────────────────────────────────────

func TestCatTitleHyphenated(t *testing.T) {
	got := catTitle("silent-failure")
	if got != "Silent Failure" {
		t.Errorf("catTitle(%q) = %q, want %q", "silent-failure", got, "Silent Failure")
	}
}

func TestCatTitleUnderscore(t *testing.T) {
	got := catTitle("build_error")
	if got != "Build Error" {
		t.Errorf("catTitle(%q) = %q, want %q", "build_error", got, "Build Error")
	}
}

func TestCatTitleEmpty(t *testing.T) {
	got := catTitle("")
	if got != "" {
		t.Errorf("catTitle(%q) = %q, want %q", "", got, "")
	}
}

func TestCatTitleSingleWord(t *testing.T) {
	got := catTitle("auth")
	if got != "Auth" {
		t.Errorf("catTitle(%q) = %q, want %q", "auth", got, "Auth")
	}
}

// ── normCatDir ────────────────────────────────────────────────────────────────

func TestNormCatDirLowercasesAndReplaces(t *testing.T) {
	got := normCatDir("Build_Error")
	if got != "build-error" {
		t.Errorf("normCatDir(%q) = %q, want %q", "Build_Error", got, "build-error")
	}
}

func TestNormCatDirAlreadyNormalized(t *testing.T) {
	got := normCatDir("silent-failure")
	if got != "silent-failure" {
		t.Errorf("normCatDir(%q) = %q, want %q", "silent-failure", got, "silent-failure")
	}
}

// ── firstSentence ─────────────────────────────────────────────────────────────

func TestFirstSentenceNormalPeriod(t *testing.T) {
	got := firstSentence("The build failed. More details follow.")
	if got != "The build failed." {
		t.Errorf("firstSentence() = %q, want %q", got, "The build failed.")
	}
}

func TestFirstSentenceNoSentenceBoundary(t *testing.T) {
	got := firstSentence("No sentence boundary here")
	if got == "" {
		t.Error("expected non-empty result")
	}
	// Should return the text (truncated at 20 words if needed).
	if !strings.Contains(got, "No sentence") {
		t.Errorf("unexpected result: %q", got)
	}
}

func TestFirstSentenceMarkdownHeading(t *testing.T) {
	got := firstSentence("## Why it fails\nThe auth token expired. Other details.")
	if got != "The auth token expired." {
		t.Errorf("firstSentence() with heading = %q, want %q", got, "The auth token expired.")
	}
}

func TestFirstSentenceEmpty(t *testing.T) {
	got := firstSentence("")
	if got != "" {
		t.Errorf("expected empty for empty input, got %q", got)
	}
}

func TestFirstSentenceExclamation(t *testing.T) {
	got := firstSentence("Build failed! Check the logs.")
	if got != "Build failed!" {
		t.Errorf("firstSentence() = %q, want %q", got, "Build failed!")
	}
}

func TestFirstSentenceLongNoTerminator(t *testing.T) {
	words := strings.Repeat("word ", 30)
	got := firstSentence(strings.TrimSpace(words))
	fields := strings.Fields(got)
	if len(fields) > 20 {
		t.Errorf("expected at most 20 words, got %d", len(fields))
	}
}

// ── longestSignal ─────────────────────────────────────────────────────────────

func TestLongestSignalPrefersLonger(t *testing.T) {
	got := longestSignal([]string{"short", "medium length signal", "x"})
	if got != "medium length signal" {
		t.Errorf("longestSignal() = %q, want %q", got, "medium length signal")
	}
}

func TestLongestSignalEmpty(t *testing.T) {
	got := longestSignal(nil)
	if got != "" {
		t.Errorf("expected empty for nil input, got %q", got)
	}
}

func TestLongestSignalSkipsOver80Chars(t *testing.T) {
	long := strings.Repeat("x", 81)
	got := longestSignal([]string{long, "shorter"})
	if got != "shorter" {
		t.Errorf("expected %q, got %q", "shorter", got)
	}
}

func TestLongestSignalExactly80Chars(t *testing.T) {
	exactly80 := strings.Repeat("x", 80)
	got := longestSignal([]string{exactly80})
	if got != exactly80 {
		t.Errorf("expected 80-char string to be accepted, got %q", got)
	}
}

// ── platformLabel ─────────────────────────────────────────────────────────────

func TestPlatformLabelKnownTags(t *testing.T) {
	cases := map[string]string{
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
	for tag, want := range cases {
		got := platformLabel(tag)
		if got != want {
			t.Errorf("platformLabel(%q) = %q, want %q", tag, got, want)
		}
	}
}

func TestPlatformLabelUnknownTag(t *testing.T) {
	got := platformLabel("unknown-platform")
	if got != "" {
		t.Errorf("expected empty for unknown tag, got %q", got)
	}
}

func TestPlatformLabelCaseInsensitive(t *testing.T) {
	got := platformLabel("DOCKER")
	if got != "Docker" {
		t.Errorf("platformLabel(DOCKER) = %q, want %q", got, "Docker")
	}
}

// ── containsAny ───────────────────────────────────────────────────────────────

func TestContainsAnyFound(t *testing.T) {
	if !containsAny([]string{"ci", "build"}, "ci") {
		t.Error("expected containsAny to find 'ci'")
	}
}

func TestContainsAnyNotFound(t *testing.T) {
	if containsAny([]string{"auth", "network"}, "ci") {
		t.Error("expected containsAny to not find 'ci'")
	}
}

func TestContainsAnyCaseInsensitive(t *testing.T) {
	if !containsAny([]string{"CI"}, "ci") {
		t.Error("expected case-insensitive match")
	}
}

func TestContainsAnyMultipleVals(t *testing.T) {
	if !containsAny([]string{"build", "runtime"}, "auth", "build") {
		t.Error("expected match on second val")
	}
}

func TestContainsAnyEmptySlice(t *testing.T) {
	if containsAny(nil, "ci") {
		t.Error("expected false for nil slice")
	}
}

// ── containsPhrase ────────────────────────────────────────────────────────────

func TestContainsPhraseFound(t *testing.T) {
	if !containsPhrase([]string{"docker auth failure", "npm install"}, "Docker Auth Failure") {
		t.Error("expected case-insensitive match")
	}
}

func TestContainsPhraseNotFound(t *testing.T) {
	if containsPhrase([]string{"auth", "build"}, "network") {
		t.Error("expected not found")
	}
}

// ── dedupe ────────────────────────────────────────────────────────────────────

func TestDedupeRemovesDuplicates(t *testing.T) {
	got := dedupe([]string{"alpha", "Beta", "alpha", "BETA", "gamma"})
	if len(got) != 3 {
		t.Errorf("expected 3 unique items, got %d: %v", len(got), got)
	}
}

func TestDedupePreservesOrder(t *testing.T) {
	got := dedupe([]string{"c", "b", "a", "b", "c"})
	if len(got) != 3 {
		t.Fatalf("expected 3, got %d: %v", len(got), got)
	}
	if got[0] != "c" || got[1] != "b" || got[2] != "a" {
		t.Errorf("unexpected order: %v", got)
	}
}

func TestDedupeEmpty(t *testing.T) {
	got := dedupe(nil)
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

// ── searchPhrases ─────────────────────────────────────────────────────────────

func TestSearchPhrasesContainsTitleAndID(t *testing.T) {
	pb := docPlaybook{
		Playbook: model.Playbook{
			ID:       "docker-auth",
			Title:    "Docker Auth Failure",
			Category: "auth",
			Tags:     []string{"docker"},
			Match:    model.MatchSpec{Any: []string{"pull access denied"}},
		},
	}
	phrases := searchPhrases(pb)
	// Must include the title.
	found := false
	for _, p := range phrases {
		if p == "Docker Auth Failure" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected title in phrases, got %v", phrases)
	}
	// Must include faultline explain id.
	foundID := false
	for _, p := range phrases {
		if strings.Contains(p, "faultline explain docker-auth") {
			foundID = true
		}
	}
	if !foundID {
		t.Errorf("expected explain phrase, got %v", phrases)
	}
}

func TestSearchPhrasesCapAtSix(t *testing.T) {
	pb := docPlaybook{
		Playbook: model.Playbook{
			ID:       "test-id",
			Title:    "Test Title",
			Category: "ci",
			Tags:     []string{"github-actions", "ci"},
			Match:    model.MatchSpec{Any: []string{"error msg"}},
		},
	}
	phrases := searchPhrases(pb)
	if len(phrases) > 6 {
		t.Errorf("expected at most 6 phrases, got %d: %v", len(phrases), phrases)
	}
}

func TestSearchPhrasesNoDuplicates(t *testing.T) {
	pb := docPlaybook{
		Playbook: model.Playbook{
			ID:    "test-id",
			Title: "Test Title",
			Tags:  []string{"docker"},
			Match: model.MatchSpec{Any: []string{"error"}},
		},
	}
	phrases := searchPhrases(pb)
	seen := map[string]bool{}
	for _, p := range phrases {
		key := strings.ToLower(strings.TrimSpace(p))
		if seen[key] {
			t.Errorf("duplicate phrase: %q", p)
		}
		seen[key] = true
	}
}

func TestSearchPhrasesEmptyPlaybook(t *testing.T) {
	pb := docPlaybook{}
	phrases := searchPhrases(pb)
	// Should not panic; may return empty or minimal list.
	if phrases == nil {
		phrases = []string{}
	}
	_ = phrases
}

// ── loadPlaybooks ─────────────────────────────────────────────────────────────

func TestLoadPlaybooksFromTempDir(t *testing.T) {
	dir := t.TempDir()
	// Write a minimal valid playbook.
	content := `id: test-pb
title: Test Playbook
category: build
severity: high
summary: "A test failure."
match:
  any:
    - "error: test failure"
`
	if err := os.WriteFile(filepath.Join(dir, "test-pb.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write playbook: %v", err)
	}

	playbooks, err := loadPlaybooks(dir)
	if err != nil {
		t.Fatalf("loadPlaybooks: %v", err)
	}
	if len(playbooks) != 1 {
		t.Fatalf("expected 1 playbook, got %d", len(playbooks))
	}
	if playbooks[0].ID != "test-pb" {
		t.Errorf("expected id=test-pb, got %q", playbooks[0].ID)
	}
	if playbooks[0].CatDir != "build" {
		t.Errorf("expected catDir=build, got %q", playbooks[0].CatDir)
	}
}

func TestLoadPlaybooksSkipsNonYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("not a playbook"), 0o644); err != nil {
		t.Fatalf("write non-yaml: %v", err)
	}

	playbooks, err := loadPlaybooks(dir)
	if err != nil {
		t.Fatalf("loadPlaybooks: %v", err)
	}
	if len(playbooks) != 0 {
		t.Errorf("expected 0 playbooks, got %d", len(playbooks))
	}
}

func TestLoadPlaybooksSkipsMissingID(t *testing.T) {
	dir := t.TempDir()
	// Playbook without an id field.
	content := `title: No ID
category: build
`
	if err := os.WriteFile(filepath.Join(dir, "no-id.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	playbooks, err := loadPlaybooks(dir)
	if err != nil {
		t.Fatalf("loadPlaybooks: %v", err)
	}
	if len(playbooks) != 0 {
		t.Errorf("expected 0 playbooks (missing id skipped), got %d", len(playbooks))
	}
}

func TestLoadPlaybooksNonexistentDirErrors(t *testing.T) {
	_, err := loadPlaybooks("/nonexistent/path/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

// ── generateAll ───────────────────────────────────────────────────────────────

func minimalPlaybooks() []docPlaybook {
	return []docPlaybook{
		{
			Playbook: model.Playbook{
				ID:       "docker-auth",
				Title:    "Docker Auth Failure",
				Category: "auth",
				Severity: "high",
				Summary:  "Image pull failed.",
				Match:    model.MatchSpec{Any: []string{"pull access denied"}},
			},
			SourceRel: "playbooks/bundled/log/auth/docker-auth.yaml",
			CatDir:    "auth",
		},
		{
			Playbook: model.Playbook{
				ID:       "missing-exec",
				Title:    "Missing Executable",
				Category: "runtime",
				Severity: "medium",
				Summary:  "Command not found.",
				Match:    model.MatchSpec{Any: []string{"executable file not found"}},
			},
			SourceRel: "playbooks/bundled/log/runtime/missing-exec.yaml",
			CatDir:    "runtime",
		},
	}
}

func TestGenerateAllProducesPageCatalogAndLLMsTxt(t *testing.T) {
	playbooks := minimalPlaybooks()
	files, err := generateAll(playbooks)
	if err != nil {
		t.Fatalf("generateAll: %v", err)
	}
	// Expect one .md per playbook + catalog/README.md + llms.txt
	if len(files) < 4 {
		t.Fatalf("expected at least 4 files, got %d", len(files))
	}
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.RelPath
	}
	hasPage := func(relPath string) bool {
		for _, p := range paths {
			if p == relPath {
				return true
			}
		}
		return false
	}
	if !hasPage("auth/docker-auth.md") {
		t.Errorf("expected auth/docker-auth.md, got %v", paths)
	}
	if !hasPage("runtime/missing-exec.md") {
		t.Errorf("expected runtime/missing-exec.md, got %v", paths)
	}
	if !hasPage("catalog/README.md") {
		t.Errorf("expected catalog/README.md, got %v", paths)
	}
	if !hasPage("llms.txt") {
		t.Errorf("expected llms.txt, got %v", paths)
	}
}

func TestGenerateAllPageContentContainsTitle(t *testing.T) {
	playbooks := minimalPlaybooks()
	files, err := generateAll(playbooks)
	if err != nil {
		t.Fatalf("generateAll: %v", err)
	}
	for _, f := range files {
		if f.RelPath == "auth/docker-auth.md" {
			if !bytes.Contains(f.Content, []byte("Docker Auth Failure")) {
				t.Errorf("expected title in page content, got %q", string(f.Content[:min(200, len(f.Content))]))
			}
			return
		}
	}
	t.Error("did not find auth/docker-auth.md in generated files")
}

func TestGenerateAllSkipsPlaybooksWithEmptyCatDir(t *testing.T) {
	playbooks := []docPlaybook{
		{
			Playbook:  model.Playbook{ID: "no-cat", Title: "No Category"},
			SourceRel: "playbooks/no-cat.yaml",
			CatDir:    "",
		},
		{
			Playbook:  model.Playbook{ID: "valid", Title: "Valid", Match: model.MatchSpec{Any: []string{"error"}}},
			SourceRel: "playbooks/valid.yaml",
			CatDir:    "build",
		},
	}
	files, err := generateAll(playbooks)
	if err != nil {
		t.Fatalf("generateAll: %v", err)
	}
	for _, f := range files {
		if strings.Contains(f.RelPath, "no-cat") {
			t.Errorf("expected no-cat playbook to be skipped, got file: %s", f.RelPath)
		}
	}
}

// ── renderCatalog ─────────────────────────────────────────────────────────────

func TestRenderCatalogContainsCategoryAndTitle(t *testing.T) {
	playbooks := minimalPlaybooks()
	content, err := renderCatalog(playbooks)
	if err != nil {
		t.Fatalf("renderCatalog: %v", err)
	}
	if !bytes.Contains(content, []byte("Docker Auth Failure")) {
		t.Errorf("expected title in catalog, got %q", string(content[:min(200, len(content))]))
	}
	if !bytes.Contains(content, []byte("2 playbooks")) {
		t.Errorf("expected playbook count in catalog")
	}
}

// ── renderLLMsTxt ─────────────────────────────────────────────────────────────

func TestRenderLLMsTxtContainsTitles(t *testing.T) {
	playbooks := minimalPlaybooks()
	content, err := renderLLMsTxt(playbooks)
	if err != nil {
		t.Fatalf("renderLLMsTxt: %v", err)
	}
	if !bytes.Contains(content, []byte("Docker Auth Failure")) {
		t.Errorf("expected title in llms.txt, got %q", string(content[:min(200, len(content))]))
	}
}

// ── checkFiles ────────────────────────────────────────────────────────────────

func TestCheckFilesOutOfDate(t *testing.T) {
	dir := t.TempDir()
	files := []generatedFile{
		{RelPath: "build/test.md", Content: []byte("new content\n")},
	}
	writeDoc(t, dir, "build/test.md", "old content\n")

	stale, err := checkFiles(dir, files)
	if err != nil {
		t.Fatalf("checkFiles: %v", err)
	}
	if len(stale) == 0 {
		t.Error("expected stale files, got none")
	}
	if !strings.Contains(strings.Join(stale, "\n"), "out of date") {
		t.Errorf("expected 'out of date' in stale list, got %v", stale)
	}
}

func TestCheckFilesMissingFile(t *testing.T) {
	dir := t.TempDir()
	files := []generatedFile{
		{RelPath: "build/new-file.md", Content: []byte("content\n")},
	}

	stale, err := checkFiles(dir, files)
	if err != nil {
		t.Fatalf("checkFiles: %v", err)
	}
	if len(stale) == 0 {
		t.Error("expected missing file to be reported as stale")
	}
	if !strings.Contains(strings.Join(stale, "\n"), "missing") {
		t.Errorf("expected 'missing' in stale list, got %v", stale)
	}
}

func TestCheckFilesUpToDate(t *testing.T) {
	dir := t.TempDir()
	content := []byte("current content\n")
	files := []generatedFile{
		{RelPath: "build/up-to-date.md", Content: content},
	}
	writeDoc(t, dir, "build/up-to-date.md", string(content))

	stale, err := checkFiles(dir, files)
	if err != nil {
		t.Fatalf("checkFiles: %v", err)
	}
	if len(stale) != 0 {
		t.Errorf("expected no stale files, got %v", stale)
	}
}

// ── writeFiles ────────────────────────────────────────────────────────────────

func TestWriteFilesCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	files := []generatedFile{
		{RelPath: "auth/docker-auth.md", Content: []byte("page content\n")},
		{RelPath: "catalog/README.md", Content: []byte("catalog\n")},
	}

	if err := writeFiles(dir, files); err != nil {
		t.Fatalf("writeFiles: %v", err)
	}
	for _, f := range files {
		fullPath := filepath.Join(dir, filepath.FromSlash(f.RelPath))
		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("read %s: %v", f.RelPath, err)
		}
		if !bytes.Equal(data, f.Content) {
			t.Errorf("content mismatch for %s", f.RelPath)
		}
	}
}

// ── removeEmptyDirs ───────────────────────────────────────────────────────────

func TestRemoveEmptyDirsCleanup(t *testing.T) {
	dir := t.TempDir()
	emptyDir := filepath.Join(dir, "empty-subdir")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := removeEmptyDirs(dir); err != nil {
		t.Fatalf("removeEmptyDirs: %v", err)
	}
	if _, err := os.Stat(emptyDir); !os.IsNotExist(err) {
		t.Errorf("expected empty dir to be removed, stat err = %v", err)
	}
}

func TestRemoveEmptyDirsKeepsNonEmpty(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "non-empty")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file.md"), []byte("content"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := removeEmptyDirs(dir); err != nil {
		t.Fatalf("removeEmptyDirs: %v", err)
	}
	if _, err := os.Stat(subDir); err != nil {
		t.Errorf("expected non-empty dir to remain: %v", err)
	}
}

// ── isManualDoc ───────────────────────────────────────────────────────────────

func TestIsManualDocKnownFiles(t *testing.T) {
	manuals := []string{"README.md", "_template.md", "PLAYBOOK_AUTHORING_GUIDE.md"}
	for _, f := range manuals {
		if !isManualDoc(f) {
			t.Errorf("expected isManualDoc(%q) = true", f)
		}
	}
}

func TestIsManualDocGeneratedFile(t *testing.T) {
	if isManualDoc("auth/docker-auth.md") {
		t.Error("expected isManualDoc(generated) = false")
	}
}

// helper function for min (Go 1.21+, but define locally for compatibility)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
