package cli

import (
	"bufio"
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

// ── auth login ────────────────────────────────────────────────────────────────

// newLoginServer returns an httptest.Server that simulates the Better Auth
// sign-in and token-creation endpoints. signInStatus and tokenStatus let
// callers override the HTTP response codes.
func newLoginServer(t *testing.T, signInStatus, tokenStatus int) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/auth/sign-in/email", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(signInStatus)
		if signInStatus == http.StatusOK {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]string{"id": "u1", "email": "user@example.com"},
			})
		}
	})

	mux.HandleFunc("/v1/auth/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(tokenStatus)
		if tokenStatus == http.StatusCreated {
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id": "t1", "name": "faultline-cli", "token": "ft_generated",
			})
		}
	})

	return httptest.NewServer(mux)
}

func TestAuthLoginCommand_Success_AllFlagsProvided(t *testing.T) {
	useAuthTempDir(t)
	srv := newLoginServer(t, http.StatusOK, http.StatusCreated)
	defer srv.Close()

	// Provide all inputs via flags so no prompt is needed — the only
	// interactive read is the password, which goes through cmd.InOrStdin().
	cmd := newAuthLoginCommand()
	cmd.SetArgs([]string{
		"--email", "user@example.com",
		"--team", "my-team",
		"--api-url", srv.URL,
	})

	var out bytes.Buffer
	// stdin carries the password line.
	cmd.SetIn(strings.NewReader("supersecret\n"))
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth login: %v", err)
	}
	if !strings.Contains(out.String(), "Logged in as") {
		t.Errorf("expected 'Logged in as' in output, got: %q", out.String())
	}
	if !strings.Contains(out.String(), "my-team") {
		t.Errorf("expected team name in output, got: %q", out.String())
	}

	// Credentials must be persisted.
	creds, err := teams.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if creds == nil {
		t.Fatal("expected credentials after login, got nil")
	}
	if creds.Token != "ft_generated" {
		t.Errorf("Token = %q, want %q", creds.Token, "ft_generated")
	}
	if creds.TeamSlug != "my-team" {
		t.Errorf("TeamSlug = %q, want %q", creds.TeamSlug, "my-team")
	}
}

func TestAuthLoginCommand_PromptsForEmailAndTeam(t *testing.T) {
	useAuthTempDir(t)
	srv := newLoginServer(t, http.StatusOK, http.StatusCreated)
	defer srv.Close()

	t.Setenv("FAULTLINE_API_URL", srv.URL)

	cmd := newAuthLoginCommand()
	// No --email or --team flags; they will be read from stdin.
	// stdin layout: email\npassword\nteam-slug\n
	cmd.SetIn(strings.NewReader("prompted@example.com\nsecretpassword\nprompted-team\n"))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth login (prompts): %v", err)
	}

	creds, _ := teams.LoadCredentials()
	if creds == nil {
		t.Fatal("expected credentials after login, got nil")
	}
	if creds.TeamSlug != "prompted-team" {
		t.Errorf("TeamSlug = %q, want %q", creds.TeamSlug, "prompted-team")
	}
}

func TestAuthLoginCommand_EmptyEmailAfterPrompt(t *testing.T) {
	useAuthTempDir(t)

	cmd := newAuthLoginCommand()
	// email prompt returns empty line; team prompt returns empty too.
	cmd.SetIn(strings.NewReader("\nsomepassword\n\n"))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty email and team slug, got nil")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("error should mention required fields; got: %v", err)
	}
}

func TestAuthLoginCommand_PasswordEOF(t *testing.T) {
	useAuthTempDir(t)

	cmd := newAuthLoginCommand()
	cmd.SetArgs([]string{"--email", "user@example.com", "--team", "t"})
	// stdin is empty — the password read will hit EOF.
	cmd.SetIn(strings.NewReader(""))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error on password EOF, got nil")
	}
	if !strings.Contains(err.Error(), "password") {
		t.Errorf("error should mention password; got: %v", err)
	}
}

func TestReadPasswordFallsBackForNonTerminalFile(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe: %v", err)
	}
	defer r.Close()
	if _, err := w.Write([]byte("pipe-secret\n")); err != nil {
		t.Fatalf("write pipe: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}

	got, err := readPassword(r, bufio.NewScanner(r))
	if err != nil {
		t.Fatalf("readPassword: %v", err)
	}
	if string(got) != "pipe-secret" {
		t.Fatalf("password = %q", got)
	}
}

func TestAuthLoginCommand_LoginFailure(t *testing.T) {
	useAuthTempDir(t)
	srv := newLoginServer(t, http.StatusUnauthorized, http.StatusCreated)
	defer srv.Close()

	cmd := newAuthLoginCommand()
	cmd.SetArgs([]string{"--email", "bad@example.com", "--team", "t", "--api-url", srv.URL})
	cmd.SetIn(strings.NewReader("wrongpassword\n"))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}
	if !strings.Contains(err.Error(), "invalid email or password") {
		t.Errorf("error = %v, want mention of invalid credentials", err)
	}
}

func TestAuthLoginCommand_TokenSaveFailure(t *testing.T) {
	// Make the config directory read-only so SaveCredentials fails.
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	srv := newLoginServer(t, http.StatusOK, http.StatusCreated)
	defer srv.Close()

	// Create the config dir as read-only so the file write fails.
	configDir := strings.Join([]string{base, "faultline"}, string(os.PathSeparator))
	if err := os.MkdirAll(configDir, 0500); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	cmd := newAuthLoginCommand()
	cmd.SetArgs([]string{"--email", "u@example.com", "--team", "t", "--api-url", srv.URL})
	cmd.SetIn(strings.NewReader("pw\n"))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected save-credentials error when config dir is read-only, got nil")
	}
	if !strings.Contains(err.Error(), "save credentials") {
		t.Errorf("error = %v, want mention of save credentials", err)
	}
}
