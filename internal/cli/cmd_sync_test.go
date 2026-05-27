package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"faultline/internal/teams"
)

func TestSyncCommand_EnvCredentialsSyncsAnalyzeArtifact(t *testing.T) {
	useAuthTempDir(t)

	var got teams.SyncRequest
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sync/artifacts" {
			http.NotFound(w, r)
			return
		}
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":             "sync-1",
			"is_new":         true,
			"schema_version": "failure-artifact.v1",
		})
	}))
	defer srv.Close()

	path := writeSyncArtifactFile(t, `{"matched":true,"artifact":{"fingerprint":"fp-1","confidence":0.75}}`)
	t.Setenv("FAULTLINE_API_URL", srv.URL)
	t.Setenv("FAULTLINE_TOKEN", "ft_env")
	t.Setenv("FAULTLINE_PROJECT", "proj-env")

	cmd := newSyncCommand()
	cmd.SetArgs([]string{"--source", "local", "--branch", "main", "--commit-sha", "abc123", path})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync command: %v\noutput: %s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "synced  new        sync-1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if gotAuth != "Bearer ft_env" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if got.ProjectID != "proj-env" || got.Source != "local" || got.Branch != "main" || got.CommitSHA != "abc123" {
		t.Errorf("request metadata = %+v", got)
	}
	if string(got.Artifact) != `{"fingerprint":"fp-1","confidence":0.75}` {
		t.Errorf("artifact = %s", got.Artifact)
	}
}

func TestSyncCommand_StoredCredentialsIncludeTeamSlug(t *testing.T) {
	useAuthTempDir(t)

	var gotTeam string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTeam = r.Header.Get("X-Team-Slug")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":             "sync-2",
			"is_new":         false,
			"schema_version": "failure-artifact.v1",
		})
	}))
	defer srv.Close()

	writeTestCredentials(t, &teams.Credentials{
		APIURL:   srv.URL,
		Token:    "ft_stored",
		TeamSlug: "team-a",
	})

	cmd := newSyncCommand()
	cmd.SetArgs([]string{"--project", "proj-1", writeSyncArtifactFile(t, `{"matched":true,"artifact":{"fingerprint":"fp-2"}}`)})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync command: %v\noutput: %s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "synced  duplicate  sync-2") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if gotTeam != "team-a" {
		t.Errorf("X-Team-Slug = %q, want team-a", gotTeam)
	}
}

func TestSyncCommandRequiresProject(t *testing.T) {
	useAuthTempDir(t)
	t.Setenv("FAULTLINE_TOKEN", "ft_env")

	cmd := newSyncCommand()
	cmd.SetArgs([]string{writeSyncArtifactFile(t, `{"matched":true,"artifact":{"fingerprint":"fp-1"}}`)})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected project error, got nil")
	}
	if !strings.Contains(err.Error(), "--project is required") {
		t.Fatalf("error = %q", err.Error())
	}
}

func writeSyncArtifactFile(t *testing.T, data string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "artifact.json")
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	return path
}
