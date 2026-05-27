package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"faultline/internal/teams"
)

func TestSyncEnvCredentialsSyncsArtifact(t *testing.T) {
	var got teams.SyncRequest
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	t.Setenv("FAULTLINE_API_URL", srv.URL)
	t.Setenv("FAULTLINE_TOKEN", "ft_env")
	t.Setenv("FAULTLINE_PROJECT", "proj-env")
	input := strings.NewReader(`{"matched":true,"artifact":{"fingerprint":"fp-1","confidence":0.75}}`)
	var out bytes.Buffer

	err := NewService().Sync(context.Background(), input, SyncOptions{
		Source:    "local",
		Branch:    "main",
		CommitSHA: "abc123",
	}, &out)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if !strings.Contains(out.String(), "synced  new        sync-1") {
		t.Fatalf("unexpected output: %q", out.String())
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

func TestExtractSyncArtifactValidation(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{
			name:  "valid analyze JSON",
			input: `{"matched":true,"artifact":{"fingerprint":"fp-1"}}`,
			want:  `{"fingerprint":"fp-1"}`,
		},
		{
			name:    "unmatched analyze JSON",
			input:   `{"matched":false,"artifact":null}`,
			wantErr: "not matched",
		},
		{
			name:    "missing artifact field",
			input:   `{}`,
			wantErr: "no artifact field",
		},
		{
			name:    "artifact missing fingerprint",
			input:   `{"matched":true,"artifact":{"confidence":0.9}}`,
			wantErr: "missing fingerprint",
		},
		{
			name:    "artifact is scalar",
			input:   `{"matched":true,"artifact":42}`,
			wantErr: "JSON object",
		},
		{
			name:    "invalid JSON",
			input:   `{`,
			wantErr: "parse artifact JSON",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractSyncArtifact([]byte(tc.input))
			if tc.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %q, want substring %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("extractSyncArtifact: %v", err)
			}
			if string(got) != tc.want {
				t.Fatalf("artifact = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestReadSyncArtifactInputRejectsOversizedInput(t *testing.T) {
	_, err := readSyncArtifactInput(strings.NewReader(strings.Repeat("x", maxSyncArtifactBytes+1)))
	if err == nil || !strings.Contains(err.Error(), "maximum size") {
		t.Fatalf("expected maximum size error, got %v", err)
	}
}
