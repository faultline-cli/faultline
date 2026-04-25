package registry

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// --- Registry ---

func TestMustLookupReturnsErrorForUnknown(t *testing.T) {
	reg := Default()
	_, err := reg.MustLookup("does_not_exist")
	if err == nil {
		t.Fatal("expected error for unknown step type")
	}
}

func TestMustLookupReturnsStepForKnown(t *testing.T) {
	reg := Default()
	step, err := reg.MustLookup("noop")
	if err != nil {
		t.Fatalf("MustLookup: %v", err)
	}
	if step.Name() != "noop" {
		t.Fatalf("expected noop, got %q", step.Name())
	}
}

func TestLookupNilRegistryReturnsFalse(t *testing.T) {
	var reg *Registry
	_, ok := reg.Lookup("noop")
	if ok {
		t.Fatal("expected nil registry to return false")
	}
}

func TestLookupTrimsSpace(t *testing.T) {
	reg := Default()
	_, ok := reg.Lookup("  noop  ")
	if !ok {
		t.Fatal("expected lookup with surrounding spaces to succeed")
	}
}

// --- truncateOutput ---

func TestTruncateOutputEmpty(t *testing.T) {
	if got := truncateOutput(""); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestTruncateOutputWithinLimit(t *testing.T) {
	input := "short output"
	if got := truncateOutput(input); got != input {
		t.Fatalf("expected %q, got %q", input, got)
	}
}

func TestTruncateOutputExceedsLimit(t *testing.T) {
	long := make([]byte, 300)
	for i := range long {
		long[i] = 'a'
	}
	out := truncateOutput(string(long))
	if len(out) != 240 {
		t.Fatalf("expected truncation to 240, got %d", len(out))
	}
}

func TestTruncateOutputTrimsWhitespace(t *testing.T) {
	if got := truncateOutput("  \n  "); got != "" {
		t.Fatalf("expected whitespace-only to be empty, got %q", got)
	}
}

// --- boolPtr ---

func TestBoolPtrTrue(t *testing.T) {
	p := boolPtr(true)
	if p == nil || !*p {
		t.Fatal("expected *true")
	}
}

func TestBoolPtrFalse(t *testing.T) {
	p := boolPtr(false)
	if p == nil || *p {
		t.Fatal("expected *false")
	}
}

// --- decodeStrict ---

func TestDecodeStrictNilInput(t *testing.T) {
	var target struct{ Name string }
	if err := decodeStrict(nil, &target); err != nil {
		t.Fatalf("expected nil input to succeed, got %v", err)
	}
}

func TestDecodeStrictUnknownFieldErrors(t *testing.T) {
	var target struct{ Name string `json:"name"` }
	if err := decodeStrict(map[string]any{"name": "ok", "unknown": "x"}, &target); err == nil {
		t.Fatal("expected error for unknown field")
	}
}

// --- whichCommandStep ---

func TestWhichCommandStepName(t *testing.T) {
	s := whichCommandStep{}
	if s.Name() != "which_command" {
		t.Fatal("unexpected name")
	}
}

func TestWhichCommandStepMetadata(t *testing.T) {
	s := whichCommandStep{}
	if len(s.KnownOutputs()) == 0 {
		t.Fatal("expected known outputs")
	}
	if s.Idempotence() == "" {
		t.Fatal("expected idempotence value")
	}
	if s.Safety(nil) != SafetyClassRead {
		t.Fatalf("expected read safety, got %v", s.Safety(nil))
	}
}

func TestWhichCommandStepDecodeArgsRequiresCommand(t *testing.T) {
	s := whichCommandStep{}
	if _, err := s.DecodeArgs(map[string]any{}); err == nil {
		t.Fatal("expected error for missing command")
	}
}

func TestWhichCommandStepDecodeArgsUnknownField(t *testing.T) {
	s := whichCommandStep{}
	if _, err := s.DecodeArgs(map[string]any{"command": "ls", "extra": true}); err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestWhichCommandStepDecodeArgsValid(t *testing.T) {
	s := whichCommandStep{}
	args, err := s.DecodeArgs(map[string]any{"command": "ls"})
	if err != nil {
		t.Fatalf("DecodeArgs: %v", err)
	}
	if args.(whichCommandArgs).Command != "ls" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestWhichCommandStepDecodeExpectEmpty(t *testing.T) {
	s := whichCommandStep{}
	v, err := s.DecodeExpect(map[string]any{})
	if err != nil || v != nil {
		t.Fatalf("expected nil for empty expect, got (%v, %v)", v, err)
	}
}

func TestWhichCommandStepDecodeExpectFound(t *testing.T) {
	s := whichCommandStep{}
	v, err := s.DecodeExpect(map[string]any{"found": true})
	if err != nil || v == nil {
		t.Fatalf("expected non-nil expect, got (%v, %v)", v, err)
	}
}

func TestWhichCommandStepDryRun(t *testing.T) {
	s := whichCommandStep{}
	desc := s.DryRun(whichCommandArgs{Command: "git"})
	if desc == "" {
		t.Fatal("expected non-empty dry-run description")
	}
}

func TestWhichCommandStepExecuteFoundCommand(t *testing.T) {
	s := whichCommandStep{}
	result, err := s.Execute(context.Background(), Runtime{}, whichCommandArgs{Command: "sh"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Outputs["found"] != "true" {
		t.Fatalf("expected found=true for sh, got %v", result.Outputs)
	}
}

func TestWhichCommandStepExecuteNotFound(t *testing.T) {
	s := whichCommandStep{}
	result, err := s.Execute(context.Background(), Runtime{}, whichCommandArgs{Command: "surely_not_a_command_xyz_123"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Outputs["found"] != "false" {
		t.Fatalf("expected found=false, got %v", result.Outputs)
	}
}

func TestWhichCommandStepVerifyNilExpect(t *testing.T) {
	s := whichCommandStep{}
	if err := s.Verify(nil, Result{}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestWhichCommandStepVerifyFoundMismatch(t *testing.T) {
	s := whichCommandStep{}
	f := false
	err := s.Verify(whichCommandExpect{Found: &f}, Result{Outputs: map[string]string{"found": "true"}})
	if err == nil {
		t.Fatal("expected verify error when found mismatch")
	}
}

func TestWhichCommandStepVerifyFoundMatch(t *testing.T) {
	s := whichCommandStep{}
	tr := true
	err := s.Verify(whichCommandExpect{Found: &tr}, Result{Outputs: map[string]string{"found": "true"}})
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}
}

// --- fileExistsStep ---

func TestFileExistsStepName(t *testing.T) {
	s := fileExistsStep{}
	if s.Name() != "file_exists" {
		t.Fatal("unexpected name")
	}
}

func TestFileExistsStepMetadata(t *testing.T) {
	s := fileExistsStep{}
	if len(s.KnownOutputs()) == 0 {
		t.Fatal("expected known outputs")
	}
	if s.Idempotence() == "" {
		t.Fatal("expected idempotence value")
	}
	if s.Safety(nil) != SafetyClassRead {
		t.Fatalf("expected read safety, got %v", s.Safety(nil))
	}
}

func TestFileExistsStepDecodeArgsRequiresPath(t *testing.T) {
	s := fileExistsStep{}
	if _, err := s.DecodeArgs(map[string]any{}); err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestFileExistsStepDecodeArgsValid(t *testing.T) {
	s := fileExistsStep{}
	args, err := s.DecodeArgs(map[string]any{"path": "/etc/hosts"})
	if err != nil {
		t.Fatalf("DecodeArgs: %v", err)
	}
	if args.(fileExistsArgs).Path != "/etc/hosts" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestFileExistsStepDryRun(t *testing.T) {
	s := fileExistsStep{}
	desc := s.DryRun(fileExistsArgs{Path: "/tmp/test"})
	if desc == "" {
		t.Fatal("expected non-empty dry-run description")
	}
}

func TestFileExistsStepDecodeExpectEmpty(t *testing.T) {
	s := fileExistsStep{}
	v, err := s.DecodeExpect(map[string]any{})
	if err != nil || v != nil {
		t.Fatalf("expected nil for empty expect, got (%v, %v)", v, err)
	}
}

func TestFileExistsStepDecodeExpectExists(t *testing.T) {
	s := fileExistsStep{}
	v, err := s.DecodeExpect(map[string]any{"exists": true})
	if err != nil || v == nil {
		t.Fatalf("expected non-nil expect, got (%v, %v)", v, err)
	}
}

func TestFileExistsStepExecuteExistingFile(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "test.txt")
	if err := os.WriteFile(target, []byte("hi"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	s := fileExistsStep{}
	result, err := s.Execute(context.Background(), Runtime{}, fileExistsArgs{Path: target})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Outputs["exists"] != "true" {
		t.Fatalf("expected exists=true, got %v", result.Outputs)
	}
}

func TestFileExistsStepExecuteRelativePath(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "test.txt")
	if err := os.WriteFile(target, []byte("hi"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	s := fileExistsStep{}
	result, err := s.Execute(context.Background(), Runtime{WorkDir: tmp}, fileExistsArgs{Path: "test.txt"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Outputs["exists"] != "true" {
		t.Fatalf("expected exists=true for relative path, got %v", result.Outputs)
	}
}

func TestFileExistsStepExecuteMissingFile(t *testing.T) {
	s := fileExistsStep{}
	result, err := s.Execute(context.Background(), Runtime{}, fileExistsArgs{Path: "/no/such/file/xyz_abc_123"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Outputs["exists"] != "false" {
		t.Fatalf("expected exists=false, got %v", result.Outputs)
	}
}

func TestFileExistsStepVerifyNilExpect(t *testing.T) {
	s := fileExistsStep{}
	if err := s.Verify(nil, Result{}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestFileExistsStepVerifyExistsMismatch(t *testing.T) {
	s := fileExistsStep{}
	tr := true
	err := s.Verify(fileExistsExpect{Exists: &tr}, Result{Outputs: map[string]string{"exists": "false"}})
	if err == nil {
		t.Fatal("expected verify error on mismatch")
	}
}

func TestFileExistsStepVerifyExistsMatch(t *testing.T) {
	s := fileExistsStep{}
	f := false
	err := s.Verify(fileExistsExpect{Exists: &f}, Result{Outputs: map[string]string{"exists": "false"}})
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}
}

// --- detectPackageManagerStep ---

func TestDetectPackageManagerStepName(t *testing.T) {
	s := detectPackageManagerStep{}
	if s.Name() != "detect_package_manager" {
		t.Fatal("unexpected name")
	}
}

func TestDetectPackageManagerStepMetadata(t *testing.T) {
	s := detectPackageManagerStep{}
	if len(s.KnownOutputs()) == 0 {
		t.Fatal("expected known outputs")
	}
	if s.Idempotence() == "" {
		t.Fatal("expected idempotence value")
	}
	if s.Safety(nil) != SafetyClassRead {
		t.Fatalf("expected read safety, got %v", s.Safety(nil))
	}
	if s.DryRun(nil) == "" {
		t.Fatal("expected non-empty dry-run description")
	}
}

func TestDetectPackageManagerStepDecodeArgsNoError(t *testing.T) {
	s := detectPackageManagerStep{}
	if _, err := s.DecodeArgs(map[string]any{}); err != nil {
		t.Fatalf("DecodeArgs: %v", err)
	}
}

func TestDetectPackageManagerStepDecodeExpectEmpty(t *testing.T) {
	s := detectPackageManagerStep{}
	v, err := s.DecodeExpect(map[string]any{})
	if err != nil || v != nil {
		t.Fatalf("expected nil for empty expect, got (%v, %v)", v, err)
	}
}

func TestDetectPackageManagerStepVerifyAlwaysNil(t *testing.T) {
	s := detectPackageManagerStep{}
	if err := s.Verify(nil, Result{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- noopStep ---

func TestNoopStepName(t *testing.T) {
	s := noopStep{}
	if s.Name() != "noop" {
		t.Fatal("unexpected name")
	}
}

func TestNoopStepMetadata(t *testing.T) {
	s := noopStep{}
	if len(s.KnownOutputs()) == 0 {
		t.Fatal("expected known outputs")
	}
	if s.Idempotence() == "" {
		t.Fatal("expected idempotence value")
	}
	if s.Safety(nil) != SafetyClassRead {
		t.Fatalf("expected read safety, got %v", s.Safety(nil))
	}
}

func TestNoopStepDryRunWithMessage(t *testing.T) {
	s := noopStep{}
	desc := s.DryRun(noopArgs{Message: "doing nothing"})
	if desc != "doing nothing" {
		t.Fatalf("expected message, got %q", desc)
	}
}

func TestNoopStepDryRunEmptyMessage(t *testing.T) {
	s := noopStep{}
	desc := s.DryRun(noopArgs{})
	if desc != "no-op" {
		t.Fatalf("expected 'no-op', got %q", desc)
	}
}

func TestNoopStepDecodeArgs(t *testing.T) {
	s := noopStep{}
	args, err := s.DecodeArgs(map[string]any{"message": "hi"})
	if err != nil {
		t.Fatalf("DecodeArgs: %v", err)
	}
	if args.(noopArgs).Message != "hi" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestNoopStepDecodeExpectEmpty(t *testing.T) {
	s := noopStep{}
	v, err := s.DecodeExpect(map[string]any{})
	if err != nil || v != nil {
		t.Fatalf("expected nil, got (%v, %v)", v, err)
	}
}

func TestNoopStepDecodeExpectNonEmpty(t *testing.T) {
	s := noopStep{}
	// noop has no expect fields, so any key is an unknown field
	if _, err := s.DecodeExpect(map[string]any{"unknown": true}); err == nil {
		t.Fatal("expected error for unknown expect field in noop")
	}
}

func TestNoopStepExecute(t *testing.T) {
	s := noopStep{}
	result, err := s.Execute(context.Background(), Runtime{}, noopArgs{Message: "ok"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Outputs["message"] != "ok" {
		t.Fatalf("unexpected output: %#v", result)
	}
	if result.Changed == nil || *result.Changed {
		t.Fatalf("expected Changed=false, got %v", result.Changed)
	}
}

func TestNoopStepVerifyAlwaysNil(t *testing.T) {
	s := noopStep{}
	if err := s.Verify(nil, Result{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- failStep ---

func TestFailStepName(t *testing.T) {
	s := failStep{}
	if s.Name() != "fail" {
		t.Fatal("unexpected name")
	}
}

func TestFailStepMetadata(t *testing.T) {
	s := failStep{}
	if len(s.KnownOutputs()) == 0 {
		t.Fatal("expected known outputs")
	}
	if s.Idempotence() == "" {
		t.Fatal("expected idempotence value")
	}
	if s.Safety(nil) != SafetyClassRead {
		t.Fatalf("expected read safety, got %v", s.Safety(nil))
	}
}

func TestFailStepDryRun(t *testing.T) {
	s := failStep{}
	desc := s.DryRun(failArgs{Message: "oops"})
	if desc == "" {
		t.Fatal("expected non-empty dry-run description")
	}
}

func TestFailStepDecodeArgsRequiresMessage(t *testing.T) {
	s := failStep{}
	if _, err := s.DecodeArgs(map[string]any{}); err == nil {
		t.Fatal("expected error for missing message")
	}
}

func TestFailStepDecodeArgsValid(t *testing.T) {
	s := failStep{}
	args, err := s.DecodeArgs(map[string]any{"message": "test failure"})
	if err != nil {
		t.Fatalf("DecodeArgs: %v", err)
	}
	if args.(failArgs).Message != "test failure" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestFailStepDecodeExpectEmpty(t *testing.T) {
	s := failStep{}
	v, err := s.DecodeExpect(map[string]any{})
	if err != nil || v != nil {
		t.Fatalf("expected nil, got (%v, %v)", v, err)
	}
}

func TestFailStepExecuteReturnsError(t *testing.T) {
	s := failStep{}
	_, err := s.Execute(context.Background(), Runtime{}, failArgs{Message: "intentional failure"})
	if err == nil {
		t.Fatal("expected error from fail step")
	}
	if err.Error() != "intentional failure" {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}

func TestFailStepVerifyAlwaysNil(t *testing.T) {
	s := failStep{}
	if err := s.Verify(nil, Result{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- installPackageStep ---

func TestInstallPackageStepName(t *testing.T) {
	s := installPackageStep{}
	if s.Name() != "install_package" {
		t.Fatal("unexpected name")
	}
}

func TestInstallPackageStepMetadata(t *testing.T) {
	s := installPackageStep{}
	if len(s.KnownOutputs()) == 0 {
		t.Fatal("expected known outputs")
	}
	if s.Idempotence() == "" {
		t.Fatal("expected idempotence value")
	}
	if s.Safety(nil) != SafetyClassEnvironmentMutation {
		t.Fatalf("expected environment mutation safety, got %v", s.Safety(nil))
	}
}

func TestInstallPackageStepDecodeArgsRequiresPackage(t *testing.T) {
	s := installPackageStep{}
	if _, err := s.DecodeArgs(map[string]any{"manager": "apt-get"}); err == nil {
		t.Fatal("expected error for missing package")
	}
}

func TestInstallPackageStepDecodeArgsRequiresManager(t *testing.T) {
	s := installPackageStep{}
	if _, err := s.DecodeArgs(map[string]any{"package": "curl"}); err == nil {
		t.Fatal("expected error for missing manager")
	}
}

func TestInstallPackageStepDecodeArgsUnsupportedManager(t *testing.T) {
	s := installPackageStep{}
	if _, err := s.DecodeArgs(map[string]any{"package": "curl", "manager": "unknown_mgr"}); err == nil {
		t.Fatal("expected error for unsupported manager")
	}
}

func TestInstallPackageStepDecodeArgsValid(t *testing.T) {
	s := installPackageStep{}
	args, err := s.DecodeArgs(map[string]any{"package": "curl", "manager": "apt-get"})
	if err != nil {
		t.Fatalf("DecodeArgs: %v", err)
	}
	typed := args.(installPackageArgs)
	if typed.Package != "curl" || typed.Manager != "apt-get" {
		t.Fatalf("unexpected args: %#v", typed)
	}
	if typed.CheckCommand != "curl" {
		t.Fatalf("expected check_command to default to package, got %q", typed.CheckCommand)
	}
}

func TestInstallPackageStepDecodeArgsVariableManager(t *testing.T) {
	// manager with ${...} should be accepted without installCommand validation
	s := installPackageStep{}
	_, err := s.DecodeArgs(map[string]any{"package": "curl", "manager": "${steps.detect.manager}"})
	if err != nil {
		t.Fatalf("expected variable manager to succeed, got %v", err)
	}
}

func TestInstallPackageStepDecodeExpectEmpty(t *testing.T) {
	s := installPackageStep{}
	v, err := s.DecodeExpect(map[string]any{})
	if err != nil || v != nil {
		t.Fatalf("expected nil for empty expect, got (%v, %v)", v, err)
	}
}

func TestInstallPackageStepDryRun(t *testing.T) {
	s := installPackageStep{}
	desc := s.DryRun(installPackageArgs{Package: "curl", Manager: "apt-get"})
	if desc == "" {
		t.Fatal("expected non-empty dry-run description")
	}
}

func TestInstallPackageStepVerifyAlwaysNil(t *testing.T) {
	s := installPackageStep{}
	if err := s.Verify(nil, Result{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInstallPackageStepExecuteAlreadyInstalled(t *testing.T) {
	// sh is always available, so CheckCommand="sh" should short-circuit
	s := installPackageStep{}
	result, err := s.Execute(context.Background(), Runtime{}, installPackageArgs{
		Package:      "sh",
		Manager:      "apt-get",
		CheckCommand: "sh",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Changed == nil || *result.Changed {
		t.Fatalf("expected Changed=false when already installed, got %v", result.Changed)
	}
}

// --- installCommand helper ---

func TestInstallCommandKnownManagers(t *testing.T) {
	for _, mgr := range []string{"apt-get", "apk", "dnf", "yum", "brew", "pacman"} {
		cmd, err := installCommand(mgr, "curl")
		if err != nil {
			t.Fatalf("installCommand(%q): %v", mgr, err)
		}
		if len(cmd) == 0 {
			t.Fatalf("expected non-empty command for %q", mgr)
		}
	}
}

func TestInstallCommandUnknownManagerErrors(t *testing.T) {
	if _, err := installCommand("unknown", "curl"); err == nil {
		t.Fatal("expected error for unknown manager")
	}
}

// --- detectPackageManagerStep.Execute ---

func TestDetectPackageManagerStepExecuteFindsManager(t *testing.T) {
	// Create a fake package manager script in a temp dir and point PATH at it.
	tmp := t.TempDir()
	script := "#!/bin/sh\nexit 0\n"
	fakeMgr := filepath.Join(tmp, "apt-get")
	if err := os.WriteFile(fakeMgr, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake manager: %v", err)
	}
	t.Setenv("PATH", tmp)

	s := detectPackageManagerStep{}
	result, err := s.Execute(context.Background(), Runtime{}, struct{}{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Outputs["manager"] != "apt-get" {
		t.Fatalf("expected manager=apt-get, got %v", result.Outputs)
	}
	if result.Outputs["command"] == "" {
		t.Fatal("expected non-empty command path")
	}
}

func TestDetectPackageManagerStepExecuteNoManagerErrors(t *testing.T) {
	// Point PATH to an empty temp dir so no package manager is found.
	t.Setenv("PATH", t.TempDir())

	s := detectPackageManagerStep{}
	_, err := s.Execute(context.Background(), Runtime{}, struct{}{})
	if err == nil {
		t.Fatal("expected error when no package manager is found")
	}
}
