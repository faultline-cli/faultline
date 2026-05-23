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

// useAuthTempDir points XDG_CONFIG_HOME at a fresh temp directory so that
// every auth test starts with no credentials and never touches the real config.
func useAuthTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	return dir
}

// writeTestCredentials writes credentials into the temp config dir so that
// auth commands that call LoadCredentials find them during a test.
func writeTestCredentials(t *testing.T, creds *teams.Credentials) {
	t.Helper()
	if err := teams.SaveCredentials(creds); err != nil {
		t.Fatalf("writeTestCredentials: SaveCredentials: %v", err)
	}
}

// ── auth logout ───────────────────────────────────────────────────────────────

func TestAuthLogoutCommand_ClearsCredentials(t *testing.T) {
	useAuthTempDir(t)
	writeTestCredentials(t, &teams.Credentials{Token: "ft_abc", TeamSlug: "team"})

	cmd := newAuthLogoutCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth logout: %v", err)
	}
	if !strings.Contains(buf.String(), "Logged out") {
		t.Errorf("auth logout: expected 'Logged out' in output, got: %q", buf.String())
	}

	// Credentials file must be gone.
	creds, err := teams.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials after logout: %v", err)
	}
	if creds != nil {
		t.Errorf("credentials still present after logout: %+v", creds)
	}
}

func TestAuthLogoutCommand_NotLoggedIn(t *testing.T) {
	useAuthTempDir(t)

	cmd := newAuthLogoutCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Logout when not logged in should succeed (idempotent).
	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth logout when not logged in: %v", err)
	}
}

// ── auth status ───────────────────────────────────────────────────────────────

func TestAuthStatusCommand_NotLoggedIn(t *testing.T) {
	useAuthTempDir(t)

	cmd := newAuthStatusCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth status: %v", err)
	}
	if !strings.Contains(buf.String(), "Not logged in") {
		t.Errorf("auth status: expected 'Not logged in' in output, got: %q", buf.String())
	}
}

func TestAuthStatusCommand_ValidToken(t *testing.T) {
	useAuthTempDir(t)

	// Start a mock API server that responds to GET /v1/auth/me.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/me" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]string{
				"id":    "user-1",
				"email": "user@example.com",
			},
		})
	}))
	defer srv.Close()

	writeTestCredentials(t, &teams.Credentials{
		APIURL:   srv.URL,
		Token:    "ft_valid",
		TeamSlug: "my-team",
		Email:    "user@example.com",
	})

	cmd := newAuthStatusCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth status: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "user@example.com") {
		t.Errorf("auth status: expected email in output, got: %q", out)
	}
	if !strings.Contains(out, "my-team") {
		t.Errorf("auth status: expected team in output, got: %q", out)
	}
}

func TestAuthStatusCommand_InvalidToken(t *testing.T) {
	useAuthTempDir(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	writeTestCredentials(t, &teams.Credentials{
		APIURL:   srv.URL,
		Token:    "ft_expired",
		TeamSlug: "my-team",
	})

	cmd := newAuthStatusCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// The command should NOT return an error; instead it prints guidance.
	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth status with invalid token: unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "re-authenticate") {
		t.Errorf("auth status: expected re-authenticate guidance, got: %q", out)
	}
}

func TestAuthStatusCommand_EmptyToken(t *testing.T) {
	useAuthTempDir(t)

	// Write credentials with an empty token (as if the file was hand-edited).
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	credDir := filepath.Join(dir, "faultline")
	if err := os.MkdirAll(credDir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	data, _ := json.Marshal(teams.Credentials{APIURL: "https://example.com", Token: "", TeamSlug: "t"})
	if err := os.WriteFile(filepath.Join(credDir, "credentials"), data, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cmd := newAuthStatusCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth status with empty token: %v", err)
	}
	if !strings.Contains(buf.String(), "Not logged in") {
		t.Errorf("expected 'Not logged in' for empty token, got: %q", buf.String())
	}
}

// ── auth token set ────────────────────────────────────────────────────────────

func TestAuthTokenSetCommand_InvalidPrefix(t *testing.T) {
	useAuthTempDir(t)

	cmd := newAuthTokenSetCommand()
	cmd.SetArgs([]string{"sk_notvalid", "--team", "my-team"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for token without ft_ prefix, got nil")
	}
	if !strings.Contains(err.Error(), "must start with 'ft_'") {
		t.Errorf("error should mention ft_ prefix; got: %v", err)
	}
}

func TestAuthTokenSetCommand_StoresToken(t *testing.T) {
	useAuthTempDir(t)

	cmd := newAuthTokenSetCommand()
	cmd.SetArgs([]string{"ft_mytoken123", "--team", "my-team"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth token set: %v", err)
	}
	if !strings.Contains(buf.String(), "Token stored") {
		t.Errorf("auth token set: expected 'Token stored' in output, got: %q", buf.String())
	}

	creds, err := teams.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials after token set: %v", err)
	}
	if creds == nil {
		t.Fatal("LoadCredentials: expected credentials, got nil")
	}
	if creds.Token != "ft_mytoken123" {
		t.Errorf("Token = %q, want %q", creds.Token, "ft_mytoken123")
	}
	if creds.TeamSlug != "my-team" {
		t.Errorf("TeamSlug = %q, want %q", creds.TeamSlug, "my-team")
	}
}

func TestAuthTokenSetCommand_UsesAPIURLFlag(t *testing.T) {
	useAuthTempDir(t)

	cmd := newAuthTokenSetCommand()
	cmd.SetArgs([]string{"ft_token", "--team", "t", "--api-url", "https://custom.api.example.com"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth token set with --api-url: %v", err)
	}

	creds, _ := teams.LoadCredentials()
	if creds == nil {
		t.Fatal("expected credentials after token set")
	}
	if creds.APIURL != "https://custom.api.example.com" {
		t.Errorf("APIURL = %q, want custom URL", creds.APIURL)
	}
}

func TestAuthTokenSetCommand_FallsBackToEnvVar(t *testing.T) {
	useAuthTempDir(t)
	t.Setenv("FAULTLINE_API_URL", "https://env.api.example.com")

	cmd := newAuthTokenSetCommand()
	cmd.SetArgs([]string{"ft_token", "--team", "t"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth token set with env var: %v", err)
	}

	creds, _ := teams.LoadCredentials()
	if creds == nil {
		t.Fatal("expected credentials after token set")
	}
	if creds.APIURL != "https://env.api.example.com" {
		t.Errorf("APIURL = %q, want env var URL", creds.APIURL)
	}
}

func TestAuthTokenSetCommand_DefaultsToProductionURL(t *testing.T) {
	useAuthTempDir(t)
	t.Setenv("FAULTLINE_API_URL", "") // clear any env override

	cmd := newAuthTokenSetCommand()
	cmd.SetArgs([]string{"ft_token", "--team", "t"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth token set (default URL): %v", err)
	}

	creds, _ := teams.LoadCredentials()
	if creds == nil {
		t.Fatal("expected credentials after token set")
	}
	if creds.APIURL != teams.DefaultAPIURL {
		t.Errorf("APIURL = %q, want default %q", creds.APIURL, teams.DefaultAPIURL)
	}
}

// ── auth command structure ────────────────────────────────────────────────────

func TestAuthCommand_HasExpectedSubcommands(t *testing.T) {
	auth := newAuthCommand()
	want := map[string]bool{
		"login":  false,
		"logout": false,
		"status": false,
		"token":  false,
	}
	for _, sub := range auth.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("auth command missing subcommand %q", name)
		}
	}
}

func TestAuthTokenCommand_HasSetSubcommand(t *testing.T) {
	token := newAuthTokenCommand()
	for _, sub := range token.Commands() {
		if sub.Name() == "set" {
			return
		}
	}
	t.Error("auth token command missing 'set' subcommand")
}
