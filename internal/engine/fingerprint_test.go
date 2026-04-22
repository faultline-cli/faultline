package engine

import (
	"testing"

	"faultline/internal/model"
)

func TestFingerprintDeterministic(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Evidence: []string{"pull access denied", "authentication required"},
	}
	a := fingerprint(result)
	b := fingerprint(result)
	if a != b {
		t.Errorf("fingerprint is not deterministic: %q != %q", a, b)
	}
	if len(a) != 8 {
		t.Errorf("expected 8-char hex fingerprint, got %q", a)
	}
}

func TestFingerprintDistinctResults(t *testing.T) {
	r1 := model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Evidence: []string{"pull access denied"},
	}
	r2 := model.Result{
		Playbook: model.Playbook{ID: "git-auth"},
		Evidence: []string{"terminal prompts disabled"},
	}
	if fingerprint(r1) == fingerprint(r2) {
		t.Error("expected different fingerprints for different results")
	}
}

func TestFingerprintNoEvidence(t *testing.T) {
	result := model.Result{
		Playbook: model.Playbook{ID: "oom-killed"},
	}
	fp := fingerprint(result)
	if fp == "" {
		t.Error("expected non-empty fingerprint even with no evidence")
	}
}
