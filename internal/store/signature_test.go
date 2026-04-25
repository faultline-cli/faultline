package store

import (
	"testing"

	"faultline/internal/model"
)

func TestSignatureForResultStableAcrossNoise(t *testing.T) {
	left := model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Evidence: []string{
			"2026-04-22T12:05:31Z Error response from daemon: pull access denied for mcr/microsoft.com/mssql/server, repository does not exist or may require 'docker login'",
			"/home/runner/work/faultline/faultline/.github/workflows/ci.yml:118: exec /__e/node20/bin/node: no such file or directory",
		},
	}
	right := model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Evidence: []string{
			"2026-04-23T07:15:44Z error response from daemon: pull access denied for mcr/microsoft.com/mssql/server, repository does not exist or may require 'docker login'",
			"/home/runner/work/other-repo/other-repo/.github/workflows/ci.yml:241: exec /__e/node20/bin/node: no such file or directory",
		},
	}

	leftSig := SignatureForResult(left)
	rightSig := SignatureForResult(right)
	if leftSig.Hash != rightSig.Hash {
		t.Fatalf("expected noisy variants to normalize to one signature:\nleft=%s\nright=%s", leftSig.Normalized, rightSig.Normalized)
	}
}

func TestSignatureForResultDifferentFailuresStayDistinct(t *testing.T) {
	a := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "docker-auth"},
		Evidence: []string{"pull access denied for registry.example.com"},
	})
	b := SignatureForResult(model.Result{
		Playbook: model.Playbook{ID: "network-timeout"},
		Evidence: []string{"context deadline exceeded while waiting for upstream"},
	})
	if a.Hash == b.Hash {
		t.Fatalf("expected distinct failures to have distinct signatures: %s", a.Hash)
	}
}

func TestNormalizeEvidenceLineRealFixtureNoise(t *testing.T) {
	line := `/home/runner/work/app/app/internal/service/user.go:43:19: cannot use db (variable of type *sql.DB) as *postgres.DB value in function call`
	got := NormalizeEvidenceLine(line)
	if got != "<workspace>/internal/service/user.go:<n> cannot use db (variable of type *sql.db) as *postgres.db value in function call" {
		t.Fatalf("unexpected normalized line: %q", got)
	}
}

func TestInputHashForLogIsDeterministic(t *testing.T) {
	text := "pull access denied for registry.example.com/app:latest"
	h1 := InputHashForLog(text)
	h2 := InputHashForLog(text)
	if h1 == "" {
		t.Fatal("expected non-empty hash")
	}
	if h1 != h2 {
		t.Fatalf("hash is not deterministic: %q != %q", h1, h2)
	}
}

func TestInputHashForLogDistinctInputsDistinctHashes(t *testing.T) {
	h1 := InputHashForLog("authentication required for https://registry.example.com")
	h2 := InputHashForLog("dial tcp: connection refused")
	if h1 == h2 {
		t.Fatalf("expected distinct hashes for distinct inputs, got %q", h1)
	}
}

func TestInputHashForLogEmptyInput(t *testing.T) {
	h := InputHashForLog("")
	if h == "" {
		t.Fatal("expected non-empty hash for empty input")
	}
}
