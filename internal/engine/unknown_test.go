package engine

import (
	"testing"

	"faultline/internal/model"
)

// ── inferUnknownCategory ──────────────────────────────────────────────────────

func TestInferUnknownCategoryKeywords(t *testing.T) {
	cases := []struct {
		signal string
		ctx    model.Context
		want   string
	}{
		// deploy keywords
		{"docker image pull failed", model.Context{}, "deploy"},
		{"kubectl apply error", model.Context{}, "deploy"},
		{"terraform plan failed", model.Context{}, "deploy"},
		{"helm upgrade error", model.Context{}, "deploy"},
		{"deployment failed", model.Context{}, "deploy"},
		// network keywords
		{"dial tcp: connection refused", model.Context{}, "network"},
		{"timeout waiting for response", model.Context{}, "network"},
		{"x509 certificate error", model.Context{}, "network"},
		{"lookup host failed", model.Context{}, "network"},
		// runtime keywords
		{"command not found: node", model.Context{}, "runtime"},
		{"no such file or directory", model.Context{}, "runtime"},
		{"node version mismatch", model.Context{}, "runtime"},
		{"python not found", model.Context{}, "runtime"},
		// test keywords
		{"fixture assertion failed", model.Context{}, "test"},
		{"expected value but got", model.Context{}, "test"},
		{"snapshot mismatch", model.Context{}, "test"},
		{"assert failed in test suite", model.Context{}, "test"},
		// auth keywords
		{"permission denied", model.Context{}, "auth"},
		{"unauthorized access", model.Context{}, "auth"},
		{"invalid token", model.Context{}, "auth"},
		{"forbidden login", model.Context{}, "auth"},
		// ci keywords
		{"workflow dispatch failed", model.Context{}, "ci"},
		{"yaml parse error in pipeline", model.Context{}, "ci"},
		{"runner not found for job", model.Context{}, "ci"},
		{"artifact upload failed", model.Context{}, "ci"},
		// context stage override: deploy
		{"some unrelated error", model.Context{Stage: "deploy"}, "deploy"},
		// context stage override: test
		{"some unrelated error", model.Context{Stage: "test"}, "test"},
		// default: build
		{"make: error in Makefile", model.Context{}, "build"},
		{"linker error", model.Context{}, "build"},
	}
	for _, tc := range cases {
		t.Run(tc.signal, func(t *testing.T) {
			got := inferUnknownCategory(tc.signal, tc.ctx)
			if got != tc.want {
				t.Errorf("inferUnknownCategory(%q, stage=%q) = %q, want %q", tc.signal, tc.ctx.Stage, got, tc.want)
			}
		})
	}
}

// ── likelyFilesForUnknownCategory ─────────────────────────────────────────────

func TestLikelyFilesForUnknownCategoryAllBranches(t *testing.T) {
	cases := []struct {
		category string
		contains string // at least one of the returned files should contain this
	}{
		{"auth", ".npmrc"},
		{"ci", "Makefile"},
		{"deploy", "Dockerfile"},
		{"network", "infra/"},
		{"runtime", ".tool-versions"},
		{"test", "fixtures/"},
		{"unknown-xyz", "Makefile"}, // default branch
	}
	for _, tc := range cases {
		t.Run(tc.category, func(t *testing.T) {
			got := likelyFilesForUnknownCategory(tc.category)
			if len(got) == 0 {
				t.Fatalf("likelyFilesForUnknownCategory(%q) returned empty slice", tc.category)
			}
			found := false
			for _, f := range got {
				if f == tc.contains {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("likelyFilesForUnknownCategory(%q) = %v, want it to contain %q", tc.category, got, tc.contains)
			}
		})
	}
}
