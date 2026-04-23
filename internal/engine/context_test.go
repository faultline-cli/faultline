package engine

import (
	"testing"

	"faultline/internal/model"
)

func TestExtractContextBuildStage(t *testing.T) {
	lines := []model.Line{
		{Original: "go build ./...", Normalized: "go build ./..."},
	}
	ctx := ExtractContext(lines)
	if ctx.Stage != "build" {
		t.Errorf("expected stage=build, got %q", ctx.Stage)
	}
}

func TestExtractContextTestStage(t *testing.T) {
	lines := []model.Line{
		{Original: "go test ./...", Normalized: "go test ./..."},
	}
	ctx := ExtractContext(lines)
	if ctx.Stage != "test" {
		t.Errorf("expected stage=test, got %q", ctx.Stage)
	}
}

func TestExtractContextDeployStage(t *testing.T) {
	lines := []model.Line{
		{Original: "kubectl apply -f deploy.yaml", Normalized: "kubectl apply -f deploy.yaml"},
	}
	ctx := ExtractContext(lines)
	if ctx.Stage != "deploy" {
		t.Errorf("expected stage=deploy, got %q", ctx.Stage)
	}
}

func TestExtractContextCommandHint(t *testing.T) {
	lines := []model.Line{
		{Original: "$ docker push ghcr.io/example/app", Normalized: "$ docker push ghcr.io/example/app"},
	}
	ctx := ExtractContext(lines)
	if ctx.CommandHint == "" {
		t.Error("expected a command hint to be extracted")
	}
	if ctx.CommandHint != "docker push ghcr.io/example/app" {
		t.Errorf("unexpected command hint %q", ctx.CommandHint)
	}
}

func TestExtractContextStepFromGHActions(t *testing.T) {
	lines := []model.Line{
		{Original: "::group::Build Docker image", Normalized: "::group::build docker image"},
	}
	ctx := ExtractContext(lines)
	if ctx.Step == "" {
		t.Error("expected step to be extracted from ::group:: marker")
	}
	if ctx.Step != "Build Docker image" {
		t.Errorf("unexpected step %q", ctx.Step)
	}
}

func TestExtractContextEmptyLines(t *testing.T) {
	ctx := ExtractContext([]model.Line{})
	if ctx.Stage != "" || ctx.CommandHint != "" || ctx.Step != "" {
		t.Errorf("expected empty context for empty lines, got %+v", ctx)
	}
}

func TestExtractContextDeployBeatsTest(t *testing.T) {
	// "deploy" should win over "test" when both keywords are on the same line.
	lines := []model.Line{
		{Original: "deploy test release", Normalized: "deploy test release"},
	}
	ctx := ExtractContext(lines)
	if ctx.Stage != "deploy" {
		t.Errorf("expected deploy to win over test, got %q", ctx.Stage)
	}
}
