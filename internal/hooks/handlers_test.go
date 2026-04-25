package hooks

import (
	"context"
	"errors"
	"os"
	"testing"

	"faultline/internal/model"
)

// --- HookPolicy.Validate ---

func TestHookPolicyValidateAcceptsAllKnownModes(t *testing.T) {
	for _, mode := range []model.HookMode{
		"",
		model.HookModeOff,
		model.HookModeVerifyOnly,
		model.HookModeCollectOnly,
		model.HookModeSafe,
		model.HookModeFull,
	} {
		p := HookPolicy{Mode: mode}
		if err := p.Validate(); err != nil {
			t.Fatalf("expected valid mode %q, got error: %v", mode, err)
		}
	}
}

func TestHookPolicyValidateRejectsUnknownMode(t *testing.T) {
	p := HookPolicy{Mode: "unknown_mode"}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for unknown hook mode")
	}
}

// --- Execute returns nil when no hooks ---

func TestExecuteReturnsNilWhenNoHooks(t *testing.T) {
	executor := NewExecutorWithRunner(HookPolicy{Mode: model.HookModeSafe}, fakeRunner{})
	playbook := model.Playbook{ID: "no-hooks"}
	report := executor.Execute(context.Background(), playbook, 0.80, ".")
	if report != nil {
		t.Fatalf("expected nil report for playbook with no hooks, got %#v", report)
	}
}

// --- Execute with invalid policy ---

func TestExecuteReturnsErrorResultForInvalidPolicy(t *testing.T) {
	executor := NewExecutorWithRunner(HookPolicy{Mode: "garbage"}, fakeRunner{})
	playbook := model.Playbook{
		ID: "bad-policy",
		Hooks: model.PlaybookHooks{
			Verify: []model.HookDefinition{{
				ID:   "probe",
				Kind: model.HookKindFileExists,
				Path: "/etc/hosts",
			}},
		},
	}
	report := executor.Execute(context.Background(), playbook, 0.80, ".")
	if report == nil {
		t.Fatal("expected non-nil report for invalid policy")
	}
	if len(report.Results) == 0 || report.Results[0].Status != model.HookStatusFailed {
		t.Fatalf("expected failed result for invalid policy, got %#v", report.Results)
	}
}

// --- blocked ---

func TestBlockedOffModeBlocksAllCategories(t *testing.T) {
	for _, mode := range []model.HookMode{"", model.HookModeOff} {
		p := HookPolicy{Mode: mode}
		for _, cat := range []model.HookCategory{model.HookCategoryVerify, model.HookCategoryCollect, model.HookCategoryRemediate} {
			blocked, _ := p.blocked(cat, model.HookKindFileExists)
			if !blocked {
				t.Fatalf("expected off mode to block category %s", cat)
			}
		}
	}
}

func TestBlockedVerifyOnlyBlocksCollect(t *testing.T) {
	p := HookPolicy{Mode: model.HookModeVerifyOnly}
	blocked, _ := p.blocked(model.HookCategoryCollect, model.HookKindFileExists)
	if !blocked {
		t.Fatal("expected verify-only mode to block collect")
	}
}

func TestBlockedVerifyOnlyBlocksCommandHooks(t *testing.T) {
	p := HookPolicy{Mode: model.HookModeVerifyOnly}
	blocked, _ := p.blocked(model.HookCategoryVerify, model.HookKindCommandExitZero)
	if !blocked {
		t.Fatal("expected verify-only mode to block command hooks in verify")
	}
}

func TestBlockedVerifyOnlyAllowsNonCommandVerifyHooks(t *testing.T) {
	p := HookPolicy{Mode: model.HookModeVerifyOnly}
	blocked, _ := p.blocked(model.HookCategoryVerify, model.HookKindFileExists)
	if blocked {
		t.Fatal("expected verify-only mode to allow file_exists in verify")
	}
}

func TestBlockedCollectOnlyBlocksVerify(t *testing.T) {
	p := HookPolicy{Mode: model.HookModeCollectOnly}
	blocked, _ := p.blocked(model.HookCategoryVerify, model.HookKindFileExists)
	if !blocked {
		t.Fatal("expected collect-only mode to block verify")
	}
}

func TestBlockedCollectOnlyBlocksCommandHooks(t *testing.T) {
	p := HookPolicy{Mode: model.HookModeCollectOnly}
	blocked, _ := p.blocked(model.HookCategoryCollect, model.HookKindCommandExitZero)
	if !blocked {
		t.Fatal("expected collect-only mode to block command hooks in collect")
	}
}

func TestBlockedCollectOnlyAllowsNonCommandCollectHooks(t *testing.T) {
	p := HookPolicy{Mode: model.HookModeCollectOnly}
	blocked, _ := p.blocked(model.HookCategoryCollect, model.HookKindFileExists)
	if blocked {
		t.Fatal("expected collect-only mode to allow file_exists in collect")
	}
}

func TestBlockedSafeModeBlocksRemediate(t *testing.T) {
	p := HookPolicy{Mode: model.HookModeSafe}
	blocked, _ := p.blocked(model.HookCategoryRemediate, model.HookKindFileExists)
	if !blocked {
		t.Fatal("expected safe mode to block remediate")
	}
}

func TestBlockedSafeModeBlocksCommandHooks(t *testing.T) {
	p := HookPolicy{Mode: model.HookModeSafe}
	blocked, _ := p.blocked(model.HookCategoryVerify, model.HookKindCommandExitZero)
	if !blocked {
		t.Fatal("expected safe mode to block command hooks")
	}
}

func TestBlockedSafeModeAllowsNonCommandVerify(t *testing.T) {
	p := HookPolicy{Mode: model.HookModeSafe}
	blocked, _ := p.blocked(model.HookCategoryVerify, model.HookKindFileExists)
	if blocked {
		t.Fatal("expected safe mode to allow file_exists in verify")
	}
}

func TestBlockedFullModeBlocksRemediate(t *testing.T) {
	p := HookPolicy{Mode: model.HookModeFull}
	blocked, _ := p.blocked(model.HookCategoryRemediate, model.HookKindCommandExitZero)
	if !blocked {
		t.Fatal("expected full mode to block remediate")
	}
}

func TestBlockedFullModeAllowsCommandHooksInVerify(t *testing.T) {
	p := HookPolicy{Mode: model.HookModeFull}
	blocked, _ := p.blocked(model.HookCategoryVerify, model.HookKindCommandExitZero)
	if blocked {
		t.Fatal("expected full mode to allow command hooks in verify")
	}
}

func TestBlockedFullModeAllowsCommandHooksInCollect(t *testing.T) {
	p := HookPolicy{Mode: model.HookModeFull}
	blocked, _ := p.blocked(model.HookCategoryCollect, model.HookKindCommandOutputCapture)
	if blocked {
		t.Fatal("expected full mode to allow command capture in collect")
	}
}

// --- runDirExists ---

func TestRunDirExistsFoundDirectory(t *testing.T) {
	tmp := t.TempDir()
	runner := fakeRunner{
		stats: map[string]os.FileInfo{
			tmp: fakeFileInfo{dir: true},
		},
	}
	def := model.HookDefinition{ID: "dir-check", Kind: model.HookKindDirExists, Path: tmp}
	result := runDirExists(context.Background(), runner, def, HookContext{})
	if result.Passed == nil || !*result.Passed {
		t.Fatalf("expected dir to pass, got %#v", result)
	}
	if result.Status != model.HookStatusExecuted {
		t.Fatalf("expected executed, got %v", result.Status)
	}
}

func TestRunDirExistsFoundFile(t *testing.T) {
	tmp := t.TempDir()
	runner := fakeRunner{
		stats: map[string]os.FileInfo{
			tmp: fakeFileInfo{dir: false},
		},
	}
	def := model.HookDefinition{ID: "dir-check", Kind: model.HookKindDirExists, Path: tmp}
	result := runDirExists(context.Background(), runner, def, HookContext{})
	if result.Passed == nil || *result.Passed {
		t.Fatalf("expected dir to fail when path is a file, got %#v", result)
	}
}

func TestRunDirExistsNotFound(t *testing.T) {
	runner := fakeRunner{stats: map[string]os.FileInfo{}}
	def := model.HookDefinition{ID: "dir-check", Kind: model.HookKindDirExists, Path: "/no/such/dir"}
	result := runDirExists(context.Background(), runner, def, HookContext{})
	if result.Passed == nil || *result.Passed {
		t.Fatalf("expected missing dir to fail, got %#v", result)
	}
}

// --- runCommandOutputMatches ---

func TestRunCommandOutputMatchesMatch(t *testing.T) {
	runner := fakeRunner{
		commands: map[string]CommandResult{
			"echo hello": {ExitCode: 0, Output: "hello world"},
		},
	}
	def := model.HookDefinition{
		ID:      "match-check",
		Kind:    model.HookKindCommandOutputMatches,
		Command: []string{"echo", "hello"},
		Pattern: "hello",
	}
	result := runCommandOutputMatches(context.Background(), runner, def, HookContext{})
	if result.Passed == nil || !*result.Passed {
		t.Fatalf("expected match to pass, got %#v", result)
	}
}

func TestRunCommandOutputMatchesNoMatch(t *testing.T) {
	runner := fakeRunner{
		commands: map[string]CommandResult{
			"echo hello": {ExitCode: 0, Output: "goodbye"},
		},
	}
	def := model.HookDefinition{
		ID:      "match-check",
		Kind:    model.HookKindCommandOutputMatches,
		Command: []string{"echo", "hello"},
		Pattern: "hello",
	}
	result := runCommandOutputMatches(context.Background(), runner, def, HookContext{})
	if result.Passed == nil || *result.Passed {
		t.Fatalf("expected no-match to fail, got %#v", result)
	}
}

func TestRunCommandOutputMatchesInvalidPattern(t *testing.T) {
	runner := fakeRunner{
		commands: map[string]CommandResult{
			"echo hello": {ExitCode: 0, Output: "hello"},
		},
	}
	def := model.HookDefinition{
		ID:      "match-check",
		Kind:    model.HookKindCommandOutputMatches,
		Command: []string{"echo", "hello"},
		Pattern: "[invalid(regex",
	}
	result := runCommandOutputMatches(context.Background(), runner, def, HookContext{})
	if result.Status != model.HookStatusFailed {
		t.Fatalf("expected failed status for invalid pattern, got %v", result.Status)
	}
}

func TestRunCommandOutputMatchesCommandFailure(t *testing.T) {
	runner := fakeRunner{
		commands: map[string]CommandResult{
			"cmd arg": {ExitCode: 0, Output: ""},
		},
		errors: map[string]error{
			"cmd arg": errors.New("exec: not found"),
		},
	}
	def := model.HookDefinition{
		ID:      "match-check",
		Kind:    model.HookKindCommandOutputMatches,
		Command: []string{"cmd", "arg"},
		Pattern: ".*",
	}
	result := runCommandOutputMatches(context.Background(), runner, def, HookContext{})
	if result.Status != model.HookStatusFailed {
		t.Fatalf("expected failed status when command errors, got %v", result.Status)
	}
}

func TestRunCommandOutputMatchesNonZeroExitFails(t *testing.T) {
	runner := fakeRunner{
		commands: map[string]CommandResult{
			"cmd": {ExitCode: 1, Output: "hello"},
		},
	}
	def := model.HookDefinition{
		ID:      "match-check",
		Kind:    model.HookKindCommandOutputMatches,
		Command: []string{"cmd"},
		Pattern: "hello",
	}
	result := runCommandOutputMatches(context.Background(), runner, def, HookContext{})
	if result.Passed == nil || *result.Passed {
		t.Fatalf("expected non-zero exit to fail even with pattern match, got %#v", result)
	}
}

// --- runReadFileExcerpt ---

func TestRunReadFileExcerptSuccess(t *testing.T) {
	runner := fakeRunner{
		files: map[string][]byte{
			"/etc/hosts": []byte("127.0.0.1 localhost\n::1 localhost"),
		},
	}
	def := model.HookDefinition{
		ID:   "read-check",
		Kind: model.HookKindReadFileExcerpt,
		Path: "/etc/hosts",
	}
	result := runReadFileExcerpt(context.Background(), runner, def, HookContext{})
	if result.Status != model.HookStatusExecuted {
		t.Fatalf("expected executed, got %v", result.Status)
	}
	if len(result.Evidence) == 0 {
		t.Fatal("expected evidence for file excerpt")
	}
}

func TestRunReadFileExcerptMissingFile(t *testing.T) {
	runner := fakeRunner{files: map[string][]byte{}}
	def := model.HookDefinition{
		ID:   "read-check",
		Kind: model.HookKindReadFileExcerpt,
		Path: "/no/such/file",
	}
	result := runReadFileExcerpt(context.Background(), runner, def, HookContext{})
	if result.Status != model.HookStatusFailed {
		t.Fatalf("expected failed for missing file, got %v", result.Status)
	}
}

func TestRunReadFileExcerptWithWorkDir(t *testing.T) {
	runner := fakeRunner{
		files: map[string][]byte{
			"/work/config.yaml": []byte("key: value\n"),
		},
	}
	def := model.HookDefinition{
		ID:   "read-check",
		Kind: model.HookKindReadFileExcerpt,
		Path: "config.yaml",
	}
	result := runReadFileExcerpt(context.Background(), runner, def, HookContext{WorkDir: "/work"})
	if result.Status != model.HookStatusExecuted {
		t.Fatalf("expected executed, got %v", result.Status)
	}
}

// --- runEnvVarPresent ---

func TestRunEnvVarPresentFound(t *testing.T) {
	runner := fakeRunner{env: map[string]string{"MY_VAR": "some-value"}}
	def := model.HookDefinition{ID: "env-check", Kind: model.HookKindEnvVarPresent, EnvVar: "MY_VAR"}
	result := runEnvVarPresent(context.Background(), runner, def, HookContext{})
	if result.Passed == nil || !*result.Passed {
		t.Fatalf("expected present env var to pass, got %#v", result)
	}
}

func TestRunEnvVarPresentNotFound(t *testing.T) {
	runner := fakeRunner{env: map[string]string{}}
	def := model.HookDefinition{ID: "env-check", Kind: model.HookKindEnvVarPresent, EnvVar: "MISSING_VAR"}
	result := runEnvVarPresent(context.Background(), runner, def, HookContext{})
	if result.Passed == nil || *result.Passed {
		t.Fatalf("expected missing env var to fail, got %#v", result)
	}
}

func TestRunEnvVarPresentEmptyValueFails(t *testing.T) {
	runner := fakeRunner{env: map[string]string{"EMPTY_VAR": "   "}}
	def := model.HookDefinition{ID: "env-check", Kind: model.HookKindEnvVarPresent, EnvVar: "EMPTY_VAR"}
	result := runEnvVarPresent(context.Background(), runner, def, HookContext{})
	if result.Passed == nil || *result.Passed {
		t.Fatalf("expected whitespace-only env var to fail, got %#v", result)
	}
}

// --- runCommandExitZero ---

func TestRunCommandExitZeroSuccess(t *testing.T) {
	runner := fakeRunner{
		commands: map[string]CommandResult{
			"echo hi": {ExitCode: 0, Output: "hi"},
		},
	}
	def := model.HookDefinition{
		ID:      "cmd-check",
		Kind:    model.HookKindCommandExitZero,
		Command: []string{"echo", "hi"},
	}
	result := runCommandExitZero(context.Background(), runner, def, HookContext{})
	if result.Passed == nil || !*result.Passed {
		t.Fatalf("expected exit zero to pass, got %#v", result)
	}
}

func TestRunCommandExitZeroNonZero(t *testing.T) {
	runner := fakeRunner{
		commands: map[string]CommandResult{
			"fail": {ExitCode: 1, Output: "error"},
		},
	}
	def := model.HookDefinition{
		ID:      "cmd-check",
		Kind:    model.HookKindCommandExitZero,
		Command: []string{"fail"},
	}
	result := runCommandExitZero(context.Background(), runner, def, HookContext{})
	if result.Passed == nil || *result.Passed {
		t.Fatalf("expected non-zero exit to fail, got %#v", result)
	}
}

// --- runCommandOutputCapture ---

func TestRunCommandOutputCaptureNoExpectation(t *testing.T) {
	runner := fakeRunner{
		commands: map[string]CommandResult{
			"git log --oneline": {ExitCode: 0, Output: "abc123 initial commit\ndef456 second commit"},
		},
	}
	def := model.HookDefinition{
		ID:      "capture",
		Kind:    model.HookKindCommandOutputCapture,
		Command: []string{"git", "log", "--oneline"},
	}
	result := runCommandOutputCapture(context.Background(), runner, def, HookContext{})
	if result.Status != model.HookStatusExecuted {
		t.Fatalf("expected executed, got %v", result.Status)
	}
	if len(result.Evidence) == 0 {
		t.Fatal("expected captured output in evidence")
	}
	// capture does not set Passed
	if result.Passed != nil {
		t.Fatalf("expected Passed to be nil for capture, got %v", result.Passed)
	}
}

// --- executeDefinition with unsupported hook kind ---

func TestExecuteDefinitionUnsupportedKindFails(t *testing.T) {
	executor := NewExecutorWithRunner(HookPolicy{Mode: model.HookModeFull}, fakeRunner{})
	playbook := model.Playbook{
		ID: "unsupported",
		Hooks: model.PlaybookHooks{
			Verify: []model.HookDefinition{{
				ID:   "unknown-hook",
				Kind: model.HookKind("unsupported_kind"),
			}},
		},
	}
	report := executor.Execute(context.Background(), playbook, 0.70, ".")
	if report == nil || len(report.Results) != 1 {
		t.Fatalf("expected one result, got %#v", report)
	}
	if report.Results[0].Status != model.HookStatusFailed {
		t.Fatalf("expected failed for unsupported kind, got %v", report.Results[0].Status)
	}
}

// --- excerptLines ---

func TestExcerptLinesEmpty(t *testing.T) {
	if got := excerptLines("", 0, 0); got != nil {
		t.Fatalf("expected nil for empty text, got %v", got)
	}
}

func TestExcerptLinesWhitespaceOnly(t *testing.T) {
	if got := excerptLines("   \n  ", 0, 0); got != nil {
		t.Fatalf("expected nil for whitespace-only text, got %v", got)
	}
}

func TestExcerptLinesTruncatesBytes(t *testing.T) {
	text := "line1\nline2\nline3\nline4\nline5"
	out := excerptLines(text, 10, 100)
	// maxBytes=10 should truncate the text
	if len(out) > 2 {
		t.Fatalf("expected at most 2 lines after byte truncation, got %v", out)
	}
}

func TestExcerptLinesTruncatesLineCount(t *testing.T) {
	text := "a\nb\nc\nd\ne\nf\ng"
	out := excerptLines(text, 10000, 3)
	if len(out) > 3 {
		t.Fatalf("expected at most 3 lines, got %v", out)
	}
}

func TestExcerptLinesDefaultLimits(t *testing.T) {
	// With maxBytes=0 and maxLines=0, defaults should be applied
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = "line content here"
	}
	text := ""
	for _, l := range lines {
		text += l + "\n"
	}
	out := excerptLines(text, 0, 0)
	// defaultExcerptLines=6, so we get at most 6 non-empty lines
	if len(out) > defaultExcerptLines {
		t.Fatalf("expected at most %d lines with defaults, got %d", defaultExcerptLines, len(out))
	}
}

func TestExcerptLinesSkipsEmptyLines(t *testing.T) {
	text := "line1\n\n\nline2"
	out := excerptLines(text, 1000, 100)
	for _, line := range out {
		if line == "" {
			t.Fatal("expected empty lines to be skipped")
		}
	}
}

// --- resolvePath ---

func TestResolvePathAbsolute(t *testing.T) {
	got := resolvePath("/some/workdir", "/absolute/path")
	if got != "/absolute/path" {
		t.Fatalf("expected absolute path unchanged, got %q", got)
	}
}

func TestResolvePathRelativeWithWorkDir(t *testing.T) {
	got := resolvePath("/work", "relative/file.txt")
	if got != "/work/relative/file.txt" {
		t.Fatalf("expected joined path, got %q", got)
	}
}

func TestResolvePathRelativeNoWorkDir(t *testing.T) {
	got := resolvePath("", "relative/file.txt")
	if got != "relative/file.txt" {
		t.Fatalf("expected clean path, got %q", got)
	}
}

// --- clamp ---

func TestClampBelowMin(t *testing.T) {
	if got := clamp(-0.5, 0, 1); got != 0 {
		t.Fatalf("expected clamped to min, got %v", got)
	}
}

func TestClampAboveMax(t *testing.T) {
	if got := clamp(1.5, 0, 1); got != 1 {
		t.Fatalf("expected clamped to max, got %v", got)
	}
}

func TestClampInRange(t *testing.T) {
	if got := clamp(0.5, 0, 1); got != 0.5 {
		t.Fatalf("expected in-range value unchanged, got %v", got)
	}
}

// --- round ---

func TestRoundTwoDecimals(t *testing.T) {
	if got := round(0.123456); got != 0.12 {
		t.Fatalf("expected 0.12, got %v", got)
	}
}

func TestRoundExact(t *testing.T) {
	if got := round(0.50); got != 0.50 {
		t.Fatalf("expected 0.50, got %v", got)
	}
}

// --- fileType ---

func TestFileTypeDir(t *testing.T) {
	if got := fileType(true); got != "dir" {
		t.Fatalf("expected 'dir', got %q", got)
	}
}

func TestFileTypeFile(t *testing.T) {
	if got := fileType(false); got != "file" {
		t.Fatalf("expected 'file', got %q", got)
	}
}

// --- confidenceDelta clamping ---

func TestExecuteConfidenceDeltaClampedAtMax(t *testing.T) {
	root := t.TempDir()
	runner := fakeRunner{
		stats: map[string]os.FileInfo{
			root + "/go.mod":      fakeFileInfo{dir: false},
			root + "/go.sum":      fakeFileInfo{dir: false},
			root + "/Dockerfile":  fakeFileInfo{dir: false},
			root + "/main.go":     fakeFileInfo{dir: false},
		},
		env: map[string]string{"MY_TOKEN": "value"},
	}
	// Build many verify hooks with large positive deltas
	executor := NewExecutorWithRunner(HookPolicy{Mode: model.HookModeSafe}, runner)
	playbook := model.Playbook{
		ID: "many-hooks",
		Hooks: model.PlaybookHooks{
			Verify: []model.HookDefinition{
				{ID: "h1", Kind: model.HookKindFileExists, Path: root + "/go.mod", ConfidenceDelta: 0.20},
				{ID: "h2", Kind: model.HookKindFileExists, Path: root + "/go.sum", ConfidenceDelta: 0.20},
				{ID: "h3", Kind: model.HookKindFileExists, Path: root + "/Dockerfile", ConfidenceDelta: 0.20},
				{ID: "h4", Kind: model.HookKindEnvVarPresent, EnvVar: "MY_TOKEN", ConfidenceDelta: 0.20},
			},
		},
	}
	report := executor.Execute(context.Background(), playbook, 0.50, root)
	if report == nil {
		t.Fatal("expected report")
	}
	if report.ConfidenceDelta > maxHookConfidenceDelta {
		t.Fatalf("confidence delta exceeded max: %v", report.ConfidenceDelta)
	}
	if report.FinalConfidence > 1.0 {
		t.Fatalf("final confidence exceeded 1.0: %v", report.FinalConfidence)
	}
}

func TestExecuteConfidenceDeltaClampedAtMin(t *testing.T) {
	runner := fakeRunner{
		stats: map[string]os.FileInfo{},
	}
	executor := NewExecutorWithRunner(HookPolicy{Mode: model.HookModeSafe}, runner)
	playbook := model.Playbook{
		ID: "negative-hooks",
		Hooks: model.PlaybookHooks{
			Verify: []model.HookDefinition{
				{ID: "h1", Kind: model.HookKindFileExists, Path: "/no/such/file", ConfidenceDelta: 0.20},
				{ID: "h2", Kind: model.HookKindFileExists, Path: "/no/such/file2", ConfidenceDelta: 0.20},
				{ID: "h3", Kind: model.HookKindEnvVarPresent, EnvVar: "NO_SUCH_ENV", ConfidenceDelta: 0.20},
			},
		},
	}
	report := executor.Execute(context.Background(), playbook, 0.50, ".")
	if report == nil {
		t.Fatal("expected report")
	}
	if report.ConfidenceDelta < -maxHookConfidenceDelta {
		t.Fatalf("confidence delta exceeded negative max: %v", report.ConfidenceDelta)
	}
}
