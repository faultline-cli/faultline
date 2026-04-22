package hooks

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"faultline/internal/model"
)

const (
	maxHookConfidenceDelta = 0.30
	defaultExcerptBytes    = 256
	defaultExcerptLines    = 6
)

type HookContext struct {
	Playbook model.Playbook
	WorkDir  string
}

type HookPolicy struct {
	Mode model.HookMode
}

type HookRegistry struct {
	handlers map[model.HookKind]func(context.Context, Runner, model.HookDefinition, HookContext) model.HookResult
}

type HookExecutor struct {
	policy   HookPolicy
	registry HookRegistry
	runner   Runner
}

type Runner interface {
	Stat(path string) (os.FileInfo, error)
	ReadFile(path string) ([]byte, error)
	LookupEnv(key string) (string, bool)
	RunCommand(ctx context.Context, argv []string, dir string) (CommandResult, error)
}

type CommandResult struct {
	ExitCode int
	Output   string
}

type systemRunner struct{}

func NewRegistry() HookRegistry {
	return HookRegistry{
		handlers: map[model.HookKind]func(context.Context, Runner, model.HookDefinition, HookContext) model.HookResult{
			model.HookKindFileExists:           runFileExists,
			model.HookKindDirExists:            runDirExists,
			model.HookKindEnvVarPresent:        runEnvVarPresent,
			model.HookKindCommandExitZero:      runCommandExitZero,
			model.HookKindCommandOutputMatches: runCommandOutputMatches,
			model.HookKindCommandOutputCapture: runCommandOutputCapture,
			model.HookKindReadFileExcerpt:      runReadFileExcerpt,
		},
	}
}

func NewExecutor(policy HookPolicy) HookExecutor {
	return HookExecutor{
		policy:   policy,
		registry: NewRegistry(),
		runner:   systemRunner{},
	}
}

func NewExecutorWithRunner(policy HookPolicy, runner Runner) HookExecutor {
	if runner == nil {
		runner = systemRunner{}
	}
	return HookExecutor{
		policy:   policy,
		registry: NewRegistry(),
		runner:   runner,
	}
}

func (p HookPolicy) Validate() error {
	switch p.Mode {
	case "", model.HookModeOff, model.HookModeVerifyOnly, model.HookModeCollectOnly, model.HookModeSafe, model.HookModeFull:
		return nil
	default:
		return fmt.Errorf("unknown hook mode %q", p.Mode)
	}
}

func (e HookExecutor) Execute(ctx context.Context, playbook model.Playbook, baseConfidence float64, workDir string) *model.HookReport {
	if err := e.policy.Validate(); err != nil {
		return &model.HookReport{
			Mode:            e.policy.Mode,
			BaseConfidence:  round(baseConfidence),
			FinalConfidence: round(baseConfidence),
			Results: []model.HookResult{{
				ID:     "hooks-policy",
				Status: model.HookStatusFailed,
				Reason: err.Error(),
			}},
		}
	}

	if len(playbook.Hooks.Verify) == 0 && len(playbook.Hooks.Collect) == 0 && len(playbook.Hooks.Remediate) == 0 {
		return nil
	}

	hookCtx := HookContext{
		Playbook: playbook,
		WorkDir:  strings.TrimSpace(workDir),
	}
	results := make([]model.HookResult, 0, len(playbook.Hooks.Verify)+len(playbook.Hooks.Collect)+len(playbook.Hooks.Remediate))
	totalDelta := 0.0

	runCategory := func(category model.HookCategory, defs []model.HookDefinition) {
		for _, def := range defs {
			item := e.executeDefinition(ctx, category, def, hookCtx)
			results = append(results, item)
			if category == model.HookCategoryVerify && item.Status == model.HookStatusExecuted && item.Passed != nil {
				if *item.Passed {
					totalDelta += def.ConfidenceDelta
				} else {
					totalDelta -= def.ConfidenceDelta
				}
			}
		}
	}

	runCategory(model.HookCategoryVerify, playbook.Hooks.Verify)
	runCategory(model.HookCategoryCollect, playbook.Hooks.Collect)
	runCategory(model.HookCategoryRemediate, playbook.Hooks.Remediate)

	totalDelta = clamp(totalDelta, -maxHookConfidenceDelta, maxHookConfidenceDelta)
	finalConfidence := clamp(baseConfidence+totalDelta, 0, 1)
	return &model.HookReport{
		Mode:            e.policy.Mode,
		BaseConfidence:  round(baseConfidence),
		ConfidenceDelta: round(totalDelta),
		FinalConfidence: round(finalConfidence),
		Results:         results,
	}
}

func (e HookExecutor) executeDefinition(ctx context.Context, category model.HookCategory, def model.HookDefinition, hookCtx HookContext) model.HookResult {
	result := baseHookResult(category, def)
	blocked, reason := e.policy.blocked(category, def.Kind)
	if blocked {
		result.Status = model.HookStatusBlocked
		result.Reason = reason
		return result
	}
	handler, ok := e.registry.handlers[def.Kind]
	if !ok {
		result.Status = model.HookStatusFailed
		result.Reason = fmt.Sprintf("unsupported hook kind %q", def.Kind)
		return result
	}
	item := handler(ctx, e.runner, def, hookCtx)
	item.ID = result.ID
	item.Category = result.Category
	item.Kind = result.Kind
	item.SourcePack = result.SourcePack
	item.SourceFile = result.SourceFile
	return item
}

func (p HookPolicy) blocked(category model.HookCategory, kind model.HookKind) (bool, string) {
	commandHook := isCommandHook(kind)
	switch p.Mode {
	case "", model.HookModeOff:
		return true, "hook execution mode is off"
	case model.HookModeVerifyOnly:
		if category != model.HookCategoryVerify {
			return true, "hook category is not enabled in verify-only mode"
		}
		if commandHook {
			return true, "command hooks require full mode"
		}
	case model.HookModeCollectOnly:
		if category != model.HookCategoryCollect {
			return true, "hook category is not enabled in collect-only mode"
		}
		if commandHook {
			return true, "command hooks require full mode"
		}
	case model.HookModeSafe:
		if category == model.HookCategoryRemediate {
			return true, "remediation hooks are not executable in this release"
		}
		if commandHook {
			return true, "command hooks require full mode"
		}
	case model.HookModeFull:
		if category == model.HookCategoryRemediate {
			return true, "remediation hooks are not executable in this release"
		}
	}
	return false, ""
}

func baseHookResult(category model.HookCategory, def model.HookDefinition) model.HookResult {
	return model.HookResult{
		ID:         def.ID,
		Category:   category,
		Kind:       def.Kind,
		SourcePack: def.Metadata.SourcePack,
		SourceFile: def.Metadata.SourceFile,
	}
}

func isCommandHook(kind model.HookKind) bool {
	switch kind {
	case model.HookKindCommandExitZero, model.HookKindCommandOutputMatches, model.HookKindCommandOutputCapture:
		return true
	default:
		return false
	}
}

func runFileExists(_ context.Context, runner Runner, def model.HookDefinition, hookCtx HookContext) model.HookResult {
	path := resolvePath(hookCtx.WorkDir, def.Path)
	info, err := runner.Stat(path)
	if err != nil {
		return model.HookResult{
			Status:   model.HookStatusExecuted,
			Passed:   boolPtr(false),
			Facts:    []model.HookFact{{Key: "path", Value: path}, {Key: "exists", Value: "false"}},
			Evidence: []string{err.Error()},
		}
	}
	passed := !info.IsDir()
	return model.HookResult{
		Status: model.HookStatusExecuted,
		Passed: boolPtr(passed),
		Facts: []model.HookFact{
			{Key: "path", Value: path},
			{Key: "exists", Value: "true"},
			{Key: "type", Value: fileType(info.IsDir())},
		},
	}
}

func runDirExists(_ context.Context, runner Runner, def model.HookDefinition, hookCtx HookContext) model.HookResult {
	path := resolvePath(hookCtx.WorkDir, def.Path)
	info, err := runner.Stat(path)
	if err != nil {
		return model.HookResult{
			Status:   model.HookStatusExecuted,
			Passed:   boolPtr(false),
			Facts:    []model.HookFact{{Key: "path", Value: path}, {Key: "exists", Value: "false"}},
			Evidence: []string{err.Error()},
		}
	}
	passed := info.IsDir()
	return model.HookResult{
		Status: model.HookStatusExecuted,
		Passed: boolPtr(passed),
		Facts: []model.HookFact{
			{Key: "path", Value: path},
			{Key: "exists", Value: "true"},
			{Key: "type", Value: fileType(info.IsDir())},
		},
	}
}

func runEnvVarPresent(_ context.Context, runner Runner, def model.HookDefinition, _ HookContext) model.HookResult {
	value, ok := runner.LookupEnv(def.EnvVar)
	present := ok && strings.TrimSpace(value) != ""
	return model.HookResult{
		Status: model.HookStatusExecuted,
		Passed: boolPtr(present),
		Facts: []model.HookFact{
			{Key: "env_var", Value: def.EnvVar},
			{Key: "present", Value: strconv.FormatBool(present)},
		},
	}
}

func runCommandExitZero(ctx context.Context, runner Runner, def model.HookDefinition, hookCtx HookContext) model.HookResult {
	cmdResult, err := runner.RunCommand(ctx, def.Command, hookCtx.WorkDir)
	if err != nil && cmdResult.ExitCode == 0 {
		return commandFailureResult(cmdResult, err)
	}
	passed := cmdResult.ExitCode == 0
	return model.HookResult{
		Status:   model.HookStatusExecuted,
		Passed:   boolPtr(passed),
		Facts:    commandFacts(def.Command, cmdResult.ExitCode),
		Evidence: excerptLines(cmdResult.Output, def.MaxBytes, def.Lines),
	}
}

func runCommandOutputMatches(ctx context.Context, runner Runner, def model.HookDefinition, hookCtx HookContext) model.HookResult {
	cmdResult, err := runner.RunCommand(ctx, def.Command, hookCtx.WorkDir)
	if err != nil && cmdResult.ExitCode == 0 {
		return commandFailureResult(cmdResult, err)
	}
	re, compileErr := regexp.Compile(def.Pattern)
	if compileErr != nil {
		return model.HookResult{
			Status: model.HookStatusFailed,
			Reason: fmt.Sprintf("invalid pattern: %v", compileErr),
			Facts:  commandFacts(def.Command, cmdResult.ExitCode),
		}
	}
	matched := re.MatchString(cmdResult.Output)
	passed := cmdResult.ExitCode == 0 && matched
	facts := append(commandFacts(def.Command, cmdResult.ExitCode),
		model.HookFact{Key: "pattern", Value: def.Pattern},
		model.HookFact{Key: "matched", Value: strconv.FormatBool(matched)},
	)
	return model.HookResult{
		Status:   model.HookStatusExecuted,
		Passed:   boolPtr(passed),
		Facts:    facts,
		Evidence: excerptLines(cmdResult.Output, def.MaxBytes, def.Lines),
	}
}

func runCommandOutputCapture(ctx context.Context, runner Runner, def model.HookDefinition, hookCtx HookContext) model.HookResult {
	cmdResult, err := runner.RunCommand(ctx, def.Command, hookCtx.WorkDir)
	if err != nil && cmdResult.ExitCode == 0 {
		return commandFailureResult(cmdResult, err)
	}
	return model.HookResult{
		Status:   model.HookStatusExecuted,
		Facts:    commandFacts(def.Command, cmdResult.ExitCode),
		Evidence: excerptLines(cmdResult.Output, def.MaxBytes, def.Lines),
	}
}

func runReadFileExcerpt(_ context.Context, runner Runner, def model.HookDefinition, hookCtx HookContext) model.HookResult {
	path := resolvePath(hookCtx.WorkDir, def.Path)
	data, err := runner.ReadFile(path)
	if err != nil {
		return model.HookResult{
			Status: model.HookStatusFailed,
			Reason: err.Error(),
			Facts:  []model.HookFact{{Key: "path", Value: path}},
		}
	}
	evidence := excerptLines(string(data), def.MaxBytes, def.Lines)
	return model.HookResult{
		Status: model.HookStatusExecuted,
		Facts: []model.HookFact{
			{Key: "path", Value: path},
			{Key: "bytes", Value: strconv.Itoa(len(data))},
		},
		Evidence: evidence,
	}
}

func commandFailureResult(cmdResult CommandResult, err error) model.HookResult {
	return model.HookResult{
		Status:   model.HookStatusFailed,
		Reason:   err.Error(),
		Facts:    commandFacts(nil, cmdResult.ExitCode),
		Evidence: excerptLines(cmdResult.Output, 0, 0),
	}
}

func commandFacts(command []string, exitCode int) []model.HookFact {
	facts := []model.HookFact{{Key: "exit_code", Value: strconv.Itoa(exitCode)}}
	if len(command) > 0 {
		facts = append([]model.HookFact{{Key: "command", Value: strings.Join(command, " ")}}, facts...)
	}
	return facts
}

func resolvePath(workDir, path string) string {
	if filepath.IsAbs(path) || strings.TrimSpace(workDir) == "" {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(workDir, path))
}

func excerptLines(text string, maxBytes, maxLines int) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	if maxBytes <= 0 {
		maxBytes = defaultExcerptBytes
	}
	if maxLines <= 0 {
		maxLines = defaultExcerptLines
	}
	raw := []byte(text)
	if len(raw) > maxBytes {
		raw = raw[:maxBytes]
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}

func clamp(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func round(value float64) float64 {
	return math.Round(value*100) / 100
}

func fileType(isDir bool) string {
	if isDir {
		return "dir"
	}
	return "file"
}

func boolPtr(value bool) *bool {
	return &value
}

func (systemRunner) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (systemRunner) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (systemRunner) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func (systemRunner) RunCommand(ctx context.Context, argv []string, dir string) (CommandResult, error) {
	if len(argv) == 0 {
		return CommandResult{}, fmt.Errorf("missing command")
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	if strings.TrimSpace(dir) != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	result := CommandResult{
		Output: strings.TrimSpace(string(bytes.TrimSpace(out))),
	}
	if err == nil {
		return result, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
		return result, err
	}
	return result, err
}
