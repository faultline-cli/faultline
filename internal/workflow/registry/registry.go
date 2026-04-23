package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type SafetyClass string

const (
	SafetyClassRead                SafetyClass = "read"
	SafetyClassLocalMutation       SafetyClass = "local_mutation"
	SafetyClassEnvironmentMutation SafetyClass = "environment_mutation"
	SafetyClassExternalSideEffect  SafetyClass = "external_side_effect"
)

type Runtime struct {
	WorkDir string
}

type Result struct {
	Outputs map[string]string
	Changed *bool
	Message string
}

type StepType interface {
	Name() string
	KnownOutputs() []string
	Idempotence() string
	DecodeArgs(map[string]any) (any, error)
	DecodeExpect(map[string]any) (any, error)
	Safety(any) SafetyClass
	DryRun(any) string
	Execute(context.Context, Runtime, any) (Result, error)
	Verify(any, Result) error
}

type Registry struct {
	types map[string]StepType
}

func Default() *Registry {
	reg := &Registry{types: map[string]StepType{}}
	for _, item := range []StepType{
		whichCommandStep{},
		fileExistsStep{},
		detectPackageManagerStep{},
		installPackageStep{},
		noopStep{},
		failStep{},
	} {
		reg.types[item.Name()] = item
	}
	return reg
}

func (r *Registry) Lookup(name string) (StepType, bool) {
	if r == nil {
		return nil, false
	}
	item, ok := r.types[strings.TrimSpace(name)]
	return item, ok
}

func (r *Registry) MustLookup(name string) (StepType, error) {
	item, ok := r.Lookup(name)
	if !ok {
		return nil, fmt.Errorf("unsupported workflow step type %q", name)
	}
	return item, nil
}

func decodeStrict(values map[string]any, target any) error {
	if values == nil {
		values = map[string]any{}
	}
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return err
	}
	return nil
}

func boolPtr(v bool) *bool {
	value := v
	return &value
}

type whichCommandArgs struct {
	Command string `json:"command"`
}

type whichCommandExpect struct {
	Found *bool `json:"found,omitempty"`
}

type whichCommandStep struct{}

func (whichCommandStep) Name() string           { return "which_command" }
func (whichCommandStep) KnownOutputs() []string { return []string{"command", "found", "path"} }
func (whichCommandStep) Idempotence() string    { return "idempotent" }
func (whichCommandStep) Safety(any) SafetyClass { return SafetyClassRead }
func (whichCommandStep) DryRun(args any) string {
	return fmt.Sprintf("resolve %q in PATH", args.(whichCommandArgs).Command)
}
func (whichCommandStep) DecodeArgs(values map[string]any) (any, error) {
	var args whichCommandArgs
	if err := decodeStrict(values, &args); err != nil {
		return nil, fmt.Errorf("decode which_command args: %w", err)
	}
	args.Command = strings.TrimSpace(args.Command)
	if args.Command == "" {
		return nil, fmt.Errorf("which_command requires command")
	}
	return args, nil
}
func (whichCommandStep) DecodeExpect(values map[string]any) (any, error) {
	if len(values) == 0 {
		return nil, nil
	}
	var expect whichCommandExpect
	if err := decodeStrict(values, &expect); err != nil {
		return nil, fmt.Errorf("decode which_command expect: %w", err)
	}
	return expect, nil
}
func (whichCommandStep) Execute(_ context.Context, _ Runtime, args any) (Result, error) {
	typed := args.(whichCommandArgs)
	path, err := exec.LookPath(typed.Command)
	if err != nil {
		return Result{
			Outputs: map[string]string{
				"command": typed.Command,
				"found":   "false",
			},
			Changed: boolPtr(false),
			Message: fmt.Sprintf("%q was not found in PATH", typed.Command),
		}, nil
	}
	return Result{
		Outputs: map[string]string{
			"command": typed.Command,
			"found":   "true",
			"path":    path,
		},
		Changed: boolPtr(false),
		Message: fmt.Sprintf("%q is available at %s", typed.Command, path),
	}, nil
}
func (whichCommandStep) Verify(expect any, result Result) error {
	if expect == nil {
		return nil
	}
	typed := expect.(whichCommandExpect)
	if typed.Found != nil {
		found := strings.EqualFold(result.Outputs["found"], "true")
		if found != *typed.Found {
			return fmt.Errorf("expected found=%t, got %t", *typed.Found, found)
		}
	}
	return nil
}

type fileExistsArgs struct {
	Path string `json:"path"`
}

type fileExistsExpect struct {
	Exists *bool `json:"exists,omitempty"`
}

type fileExistsStep struct{}

func (fileExistsStep) Name() string           { return "file_exists" }
func (fileExistsStep) KnownOutputs() []string { return []string{"path", "exists"} }
func (fileExistsStep) Idempotence() string    { return "idempotent" }
func (fileExistsStep) Safety(any) SafetyClass { return SafetyClassRead }
func (fileExistsStep) DryRun(args any) string {
	return fmt.Sprintf("check whether %s exists", args.(fileExistsArgs).Path)
}
func (fileExistsStep) DecodeArgs(values map[string]any) (any, error) {
	var args fileExistsArgs
	if err := decodeStrict(values, &args); err != nil {
		return nil, fmt.Errorf("decode file_exists args: %w", err)
	}
	args.Path = strings.TrimSpace(args.Path)
	if args.Path == "" {
		return nil, fmt.Errorf("file_exists requires path")
	}
	return args, nil
}
func (fileExistsStep) DecodeExpect(values map[string]any) (any, error) {
	if len(values) == 0 {
		return nil, nil
	}
	var expect fileExistsExpect
	if err := decodeStrict(values, &expect); err != nil {
		return nil, fmt.Errorf("decode file_exists expect: %w", err)
	}
	return expect, nil
}
func (fileExistsStep) Execute(_ context.Context, runtime Runtime, args any) (Result, error) {
	typed := args.(fileExistsArgs)
	target := typed.Path
	if !filepath.IsAbs(target) && strings.TrimSpace(runtime.WorkDir) != "" {
		target = filepath.Join(runtime.WorkDir, target)
	}
	_, err := os.Stat(target)
	exists := err == nil
	return Result{
		Outputs: map[string]string{
			"path":   typed.Path,
			"exists": strings.ToLower(fmt.Sprintf("%t", exists)),
		},
		Changed: boolPtr(false),
		Message: fmt.Sprintf("exists=%t for %s", exists, typed.Path),
	}, nil
}
func (fileExistsStep) Verify(expect any, result Result) error {
	if expect == nil {
		return nil
	}
	typed := expect.(fileExistsExpect)
	if typed.Exists != nil {
		exists := strings.EqualFold(result.Outputs["exists"], "true")
		if exists != *typed.Exists {
			return fmt.Errorf("expected exists=%t, got %t", *typed.Exists, exists)
		}
	}
	return nil
}

type detectPackageManagerStep struct{}

func (detectPackageManagerStep) Name() string           { return "detect_package_manager" }
func (detectPackageManagerStep) KnownOutputs() []string { return []string{"manager", "command"} }
func (detectPackageManagerStep) Idempotence() string    { return "idempotent" }
func (detectPackageManagerStep) Safety(any) SafetyClass { return SafetyClassRead }
func (detectPackageManagerStep) DryRun(any) string      { return "detect an available package manager" }
func (detectPackageManagerStep) DecodeArgs(values map[string]any) (any, error) {
	var args struct{}
	if err := decodeStrict(values, &args); err != nil {
		return nil, fmt.Errorf("decode detect_package_manager args: %w", err)
	}
	return args, nil
}
func (detectPackageManagerStep) DecodeExpect(values map[string]any) (any, error) {
	if len(values) == 0 {
		return nil, nil
	}
	var args struct{}
	if err := decodeStrict(values, &args); err != nil {
		return nil, fmt.Errorf("decode detect_package_manager expect: %w", err)
	}
	return args, nil
}
func (detectPackageManagerStep) Execute(_ context.Context, _ Runtime, _ any) (Result, error) {
	for _, item := range []string{"apt-get", "apk", "dnf", "yum", "brew", "pacman"} {
		if path, err := exec.LookPath(item); err == nil {
			return Result{
				Outputs: map[string]string{
					"manager": item,
					"command": path,
				},
				Changed: boolPtr(false),
				Message: fmt.Sprintf("detected package manager %s", item),
			}, nil
		}
	}
	return Result{}, fmt.Errorf("no supported package manager found in PATH")
}
func (detectPackageManagerStep) Verify(any, Result) error { return nil }

type installPackageArgs struct {
	Package      string `json:"package"`
	Manager      string `json:"manager"`
	CheckCommand string `json:"check_command,omitempty"`
}

type installPackageStep struct{}

func (installPackageStep) Name() string { return "install_package" }
func (installPackageStep) KnownOutputs() []string {
	return []string{"manager", "package", "command"}
}
func (installPackageStep) Idempotence() string { return "best_effort" }
func (installPackageStep) Safety(any) SafetyClass {
	return SafetyClassEnvironmentMutation
}
func (installPackageStep) DryRun(args any) string {
	typed := args.(installPackageArgs)
	command, _ := installCommand(typed.Manager, typed.Package)
	return fmt.Sprintf("install package %q with %s", typed.Package, strings.Join(command, " "))
}
func (installPackageStep) DecodeArgs(values map[string]any) (any, error) {
	var args installPackageArgs
	if err := decodeStrict(values, &args); err != nil {
		return nil, fmt.Errorf("decode install_package args: %w", err)
	}
	args.Package = strings.TrimSpace(args.Package)
	args.Manager = strings.TrimSpace(args.Manager)
	args.CheckCommand = strings.TrimSpace(args.CheckCommand)
	if args.Package == "" {
		return nil, fmt.Errorf("install_package requires package")
	}
	if args.Manager == "" {
		return nil, fmt.Errorf("install_package requires manager")
	}
	if args.CheckCommand == "" {
		args.CheckCommand = args.Package
	}
	if !strings.Contains(args.Manager, "${") {
		if _, err := installCommand(args.Manager, args.Package); err != nil {
			return nil, err
		}
	}
	return args, nil
}
func (installPackageStep) DecodeExpect(values map[string]any) (any, error) {
	if len(values) == 0 {
		return nil, nil
	}
	var args struct{}
	if err := decodeStrict(values, &args); err != nil {
		return nil, fmt.Errorf("decode install_package expect: %w", err)
	}
	return args, nil
}
func (installPackageStep) Execute(ctx context.Context, runtime Runtime, args any) (Result, error) {
	typed := args.(installPackageArgs)
	if typed.CheckCommand != "" {
		if path, err := exec.LookPath(typed.CheckCommand); err == nil {
			return Result{
				Outputs: map[string]string{
					"manager": typed.Manager,
					"package": typed.Package,
					"command": path,
				},
				Changed: boolPtr(false),
				Message: fmt.Sprintf("%q is already available at %s", typed.CheckCommand, path),
			}, nil
		}
	}
	command, err := installCommand(typed.Manager, typed.Package)
	if err != nil {
		return Result{}, err
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	if strings.TrimSpace(runtime.WorkDir) != "" {
		cmd.Dir = runtime.WorkDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return Result{}, fmt.Errorf("%s failed: %s", strings.Join(command, " "), truncateOutput(string(output)))
	}
	return Result{
		Outputs: map[string]string{
			"manager": typed.Manager,
			"package": typed.Package,
			"command": strings.Join(command, " "),
		},
		Changed: boolPtr(true),
		Message: truncateOutput(string(output)),
	}, nil
}
func (installPackageStep) Verify(any, Result) error { return nil }

func installCommand(manager, pkg string) ([]string, error) {
	switch manager {
	case "apt-get":
		return []string{"apt-get", "install", "-y", pkg}, nil
	case "apk":
		return []string{"apk", "add", "--no-cache", pkg}, nil
	case "dnf":
		return []string{"dnf", "install", "-y", pkg}, nil
	case "yum":
		return []string{"yum", "install", "-y", pkg}, nil
	case "brew":
		return []string{"brew", "install", pkg}, nil
	case "pacman":
		return []string{"pacman", "-Sy", "--noconfirm", pkg}, nil
	default:
		return nil, fmt.Errorf("unsupported package manager %q", manager)
	}
}

type noopArgs struct {
	Message string `json:"message,omitempty"`
}

type noopStep struct{}

func (noopStep) Name() string           { return "noop" }
func (noopStep) KnownOutputs() []string { return []string{"message"} }
func (noopStep) Idempotence() string    { return "idempotent" }
func (noopStep) Safety(any) SafetyClass { return SafetyClassRead }
func (noopStep) DryRun(args any) string {
	message := strings.TrimSpace(args.(noopArgs).Message)
	if message == "" {
		return "no-op"
	}
	return message
}
func (noopStep) DecodeArgs(values map[string]any) (any, error) {
	var args noopArgs
	if err := decodeStrict(values, &args); err != nil {
		return nil, fmt.Errorf("decode noop args: %w", err)
	}
	args.Message = strings.TrimSpace(args.Message)
	return args, nil
}
func (noopStep) DecodeExpect(values map[string]any) (any, error) {
	if len(values) == 0 {
		return nil, nil
	}
	var args struct{}
	if err := decodeStrict(values, &args); err != nil {
		return nil, fmt.Errorf("decode noop expect: %w", err)
	}
	return args, nil
}
func (noopStep) Execute(_ context.Context, _ Runtime, args any) (Result, error) {
	message := strings.TrimSpace(args.(noopArgs).Message)
	return Result{
		Outputs: map[string]string{"message": message},
		Changed: boolPtr(false),
		Message: message,
	}, nil
}
func (noopStep) Verify(any, Result) error { return nil }

type failArgs struct {
	Message string `json:"message"`
}

type failStep struct{}

func (failStep) Name() string           { return "fail" }
func (failStep) KnownOutputs() []string { return []string{"message"} }
func (failStep) Idempotence() string    { return "idempotent" }
func (failStep) Safety(any) SafetyClass { return SafetyClassRead }
func (failStep) DryRun(args any) string { return fmt.Sprintf("fail with %q", args.(failArgs).Message) }
func (failStep) DecodeArgs(values map[string]any) (any, error) {
	var args failArgs
	if err := decodeStrict(values, &args); err != nil {
		return nil, fmt.Errorf("decode fail args: %w", err)
	}
	args.Message = strings.TrimSpace(args.Message)
	if args.Message == "" {
		return nil, fmt.Errorf("fail requires message")
	}
	return args, nil
}
func (failStep) DecodeExpect(values map[string]any) (any, error) {
	if len(values) == 0 {
		return nil, nil
	}
	var args struct{}
	if err := decodeStrict(values, &args); err != nil {
		return nil, fmt.Errorf("decode fail expect: %w", err)
	}
	return args, nil
}
func (failStep) Execute(_ context.Context, _ Runtime, args any) (Result, error) {
	typed := args.(failArgs)
	return Result{Outputs: map[string]string{"message": typed.Message}}, errors.New(typed.Message)
}
func (failStep) Verify(any, Result) error { return nil }

func truncateOutput(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) > 240 {
		return value[:240]
	}
	return value
}
