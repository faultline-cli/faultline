package hooks

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"faultline/internal/model"
)

type fakeFileInfo struct {
	dir bool
}

func (f fakeFileInfo) Name() string       { return "" }
func (f fakeFileInfo) Size() int64        { return 0 }
func (f fakeFileInfo) Mode() os.FileMode  { return 0 }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool        { return f.dir }
func (f fakeFileInfo) Sys() any           { return nil }

type fakeRunner struct {
	stats    map[string]os.FileInfo
	files    map[string][]byte
	env      map[string]string
	commands map[string]CommandResult
	errors   map[string]error
}

func (f fakeRunner) Stat(path string) (os.FileInfo, error) {
	info, ok := f.stats[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return info, nil
}

func (f fakeRunner) ReadFile(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (f fakeRunner) LookupEnv(key string) (string, bool) {
	value, ok := f.env[key]
	return value, ok
}

func (f fakeRunner) RunCommand(_ context.Context, argv []string, _ string) (CommandResult, error) {
	key := ""
	for i, item := range argv {
		if i > 0 {
			key += " "
		}
		key += item
	}
	result := f.commands[key]
	if err, ok := f.errors[key]; ok {
		return result, err
	}
	return result, nil
}

func TestExecuteBlocksCommandHooksOutsideFullMode(t *testing.T) {
	runner := fakeRunner{}
	executor := NewExecutorWithRunner(HookPolicy{Mode: model.HookModeSafe}, runner)
	playbook := model.Playbook{
		ID: "docker-auth",
		Hooks: model.PlaybookHooks{
			Verify: []model.HookDefinition{{
				ID:      "docker-info",
				Kind:    model.HookKindCommandExitZero,
				Command: []string{"docker", "info"},
			}},
		},
	}

	report := executor.Execute(context.Background(), playbook, 0.80, ".")
	if report == nil || len(report.Results) != 1 {
		t.Fatalf("expected one hook result, got %#v", report)
	}
	if report.Results[0].Status != model.HookStatusBlocked {
		t.Fatalf("expected blocked status, got %#v", report.Results[0])
	}
	if report.Results[0].Reason != "command hooks require full mode" {
		t.Fatalf("unexpected block reason: %#v", report.Results[0])
	}
}

func TestExecuteAppliesVerifyConfidenceDeltas(t *testing.T) {
	root := t.TempDir()
	existing := filepath.Join(root, "go.mod")
	runner := fakeRunner{
		stats: map[string]os.FileInfo{
			existing: fakeFileInfo{dir: false},
		},
		env: map[string]string{},
	}
	executor := NewExecutorWithRunner(HookPolicy{Mode: model.HookModeSafe}, runner)
	playbook := model.Playbook{
		ID: "missing-env",
		Hooks: model.PlaybookHooks{
			Verify: []model.HookDefinition{
				{
					ID:              "go-mod-present",
					Kind:            model.HookKindFileExists,
					Path:            "go.mod",
					ConfidenceDelta: 0.05,
				},
				{
					ID:              "token-present",
					Kind:            model.HookKindEnvVarPresent,
					EnvVar:          "API_TOKEN",
					ConfidenceDelta: 0.04,
				},
			},
		},
	}

	report := executor.Execute(context.Background(), playbook, 0.60, root)
	if report == nil {
		t.Fatal("expected hook report")
	}
	if report.ConfidenceDelta != 0.01 {
		t.Fatalf("expected confidence delta 0.01, got %#v", report)
	}
	if report.FinalConfidence != 0.61 {
		t.Fatalf("expected final confidence 0.61, got %#v", report)
	}
	if len(report.Results) != 2 {
		t.Fatalf("expected two hook results, got %#v", report.Results)
	}
	if report.Results[0].Passed == nil || !*report.Results[0].Passed {
		t.Fatalf("expected first verify hook to pass, got %#v", report.Results[0])
	}
	if report.Results[1].Passed == nil || *report.Results[1].Passed {
		t.Fatalf("expected second verify hook to fail, got %#v", report.Results[1])
	}
}

func TestExecuteCommandCaptureInFullMode(t *testing.T) {
	runner := fakeRunner{
		commands: map[string]CommandResult{
			"git status --short": {
				ExitCode: 0,
				Output:   " M README.md\n?? notes.txt\n",
			},
		},
	}
	executor := NewExecutorWithRunner(HookPolicy{Mode: model.HookModeFull}, runner)
	playbook := model.Playbook{
		ID: "repo-status",
		Hooks: model.PlaybookHooks{
			Collect: []model.HookDefinition{{
				ID:      "git-status",
				Kind:    model.HookKindCommandOutputCapture,
				Command: []string{"git", "status", "--short"},
			}},
		},
	}

	report := executor.Execute(context.Background(), playbook, 0.50, ".")
	if report == nil || len(report.Results) != 1 {
		t.Fatalf("expected one hook result, got %#v", report)
	}
	item := report.Results[0]
	if item.Status != model.HookStatusExecuted {
		t.Fatalf("expected executed hook, got %#v", item)
	}
	if len(item.Evidence) != 2 {
		t.Fatalf("expected captured output lines, got %#v", item)
	}
}

func TestExecuteMarksMissingCommandAsFailed(t *testing.T) {
	runner := fakeRunner{
		commands: map[string]CommandResult{
			"docker info": {},
		},
		errors: map[string]error{
			"docker info": errors.New("exec: \"docker\": executable file not found in $PATH"),
		},
	}
	executor := NewExecutorWithRunner(HookPolicy{Mode: model.HookModeFull}, runner)
	playbook := model.Playbook{
		ID: "docker-auth",
		Hooks: model.PlaybookHooks{
			Verify: []model.HookDefinition{{
				ID:              "docker-info",
				Kind:            model.HookKindCommandExitZero,
				Command:         []string{"docker", "info"},
				ConfidenceDelta: 0.05,
			}},
		},
	}

	report := executor.Execute(context.Background(), playbook, 0.70, ".")
	if report == nil || len(report.Results) != 1 {
		t.Fatalf("expected one hook result, got %#v", report)
	}
	if report.Results[0].Status != model.HookStatusFailed {
		t.Fatalf("expected failed hook status, got %#v", report.Results[0])
	}
	if report.FinalConfidence != 0.70 {
		t.Fatalf("failed hooks should not change confidence, got %#v", report)
	}
}
