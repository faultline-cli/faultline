package teams

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer returns an httptest.Server that handles sign-in and token
// creation requests. Callers supply optional overrides via the opts map:
//   - "signInStatus"   — HTTP status for POST /api/auth/sign-in/email (default 200)
//   - "tokenStatus"    — HTTP status for POST /v1/auth/tokens (default 201)
//   - "verifyStatus"   — HTTP status for GET /v1/auth/me (default 200)
func newTestServer(t *testing.T, opts map[string]int) *httptest.Server {
	t.Helper()

	get := func(key string, dflt int) int {
		if v, ok := opts[key]; ok {
			return v
		}
		return dflt
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/auth/sign-in/email", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		code := get("signInStatus", http.StatusOK)
		w.WriteHeader(code)
		if code == http.StatusOK {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]string{
					"id":    "user-1",
					"email": "user@example.com",
				},
			})
		}
	})

	mux.HandleFunc("/v1/auth/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		code := get("tokenStatus", http.StatusCreated)
		w.WriteHeader(code)
		if code == http.StatusCreated {
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id":    "tok-1",
				"name":  "faultline-cli",
				"token": "ft_generated_token",
			})
		}
	})

	mux.HandleFunc("/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		code := get("verifyStatus", http.StatusOK)
		w.WriteHeader(code)
		if code == http.StatusOK {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]string{
					"id":    "user-1",
					"email": "user@example.com",
				},
			})
		}
	})

	return httptest.NewServer(mux)
}

// ── Login ─────────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	srv := newTestServer(t, nil)
	defer srv.Close()

	client := NewClient(srv.URL)
	token, email, err := client.Login(context.Background(), "user@example.com", "password", "my-team", "faultline-cli")
	if err != nil {
		t.Fatalf("Login: unexpected error: %v", err)
	}
	if token != "ft_generated_token" {
		t.Errorf("Login: token = %q, want %q", token, "ft_generated_token")
	}
	if email != "user@example.com" {
		t.Errorf("Login: email = %q, want %q", email, "user@example.com")
	}
}

func TestLogin_InvalidCredentials_401(t *testing.T) {
	srv := newTestServer(t, map[string]int{"signInStatus": http.StatusUnauthorized})
	defer srv.Close()

	_, _, err := NewClient(srv.URL).Login(context.Background(), "bad@example.com", "wrong", "team", "cli")
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
	if !strings.Contains(err.Error(), "invalid email or password") {
		t.Errorf("error should mention invalid credentials; got: %v", err)
	}
}

func TestLogin_InvalidCredentials_400(t *testing.T) {
	srv := newTestServer(t, map[string]int{"signInStatus": http.StatusBadRequest})
	defer srv.Close()

	_, _, err := NewClient(srv.URL).Login(context.Background(), "bad@example.com", "wrong", "team", "cli")
	if err == nil {
		t.Fatal("expected error for 400, got nil")
	}
	if !strings.Contains(err.Error(), "invalid email or password") {
		t.Errorf("error should mention invalid credentials; got: %v", err)
	}
}

func TestLogin_SignInServerError(t *testing.T) {
	srv := newTestServer(t, map[string]int{"signInStatus": http.StatusInternalServerError})
	defer srv.Close()

	_, _, err := NewClient(srv.URL).Login(context.Background(), "user@example.com", "pw", "team", "cli")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "sign in failed") {
		t.Errorf("error should mention sign in failed; got: %v", err)
	}
}

func TestLogin_TokenCreationFailure(t *testing.T) {
	srv := newTestServer(t, map[string]int{"tokenStatus": http.StatusForbidden})
	defer srv.Close()

	_, _, err := NewClient(srv.URL).Login(context.Background(), "user@example.com", "pw", "team", "cli")
	if err == nil {
		t.Fatal("expected error when token creation fails, got nil")
	}
	if !strings.Contains(err.Error(), "create token") {
		t.Errorf("error should mention create token; got: %v", err)
	}
}

func TestLogin_LargeErrorBodyTruncated(t *testing.T) {
	// Verify that a server returning a huge error body does not cause unbounded
	// memory allocation (the body is capped at maxErrorBodySize).
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/sign-in/email", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		// Return far more than maxErrorBodySize bytes.
		_, _ = w.Write([]byte(strings.Repeat("x", maxErrorBodySize*4)))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	_, _, err := NewClient(srv.URL).Login(context.Background(), "u@example.com", "pw", "team", "cli")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The error message body should be capped, not the full oversized payload.
	if len(err.Error()) > maxErrorBodySize+200 {
		t.Errorf("error message looks unbounded (%d bytes); body may not be capped", len(err.Error()))
	}
}

// ── VerifyToken ───────────────────────────────────────────────────────────────

func TestVerifyToken_Success(t *testing.T) {
	srv := newTestServer(t, nil)
	defer srv.Close()

	email, err := NewClient(srv.URL).VerifyToken(context.Background(), "ft_valid_token", "my-team")
	if err != nil {
		t.Fatalf("VerifyToken: unexpected error: %v", err)
	}
	if email != "user@example.com" {
		t.Errorf("VerifyToken: email = %q, want %q", email, "user@example.com")
	}
}

func TestVerifyToken_InvalidToken(t *testing.T) {
	srv := newTestServer(t, map[string]int{"verifyStatus": http.StatusUnauthorized})
	defer srv.Close()

	_, err := NewClient(srv.URL).VerifyToken(context.Background(), "ft_bad", "team")
	if err == nil {
		t.Fatal("expected error for unauthorized token, got nil")
	}
	if !strings.Contains(err.Error(), "invalid or expired") {
		t.Errorf("error should mention invalid or expired; got: %v", err)
	}
}

func TestVerifyToken_ServerError(t *testing.T) {
	srv := newTestServer(t, map[string]int{"verifyStatus": http.StatusInternalServerError})
	defer srv.Close()

	_, err := NewClient(srv.URL).VerifyToken(context.Background(), "ft_token", "team")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "verify token") {
		t.Errorf("error should mention verify token; got: %v", err)
	}
}

func TestVerifyToken_NoTeamSlug(t *testing.T) {
	// teamSlug is optional; verify the call succeeds without it.
	srv := newTestServer(t, nil)
	defer srv.Close()

	email, err := NewClient(srv.URL).VerifyToken(context.Background(), "ft_token", "")
	if err != nil {
		t.Fatalf("VerifyToken with empty teamSlug: unexpected error: %v", err)
	}
	if email == "" {
		t.Error("VerifyToken with empty teamSlug: expected non-empty email")
	}
}

// TestNewClient_HasTimeout confirms the HTTP client is configured with a
// timeout so slow or hung servers cannot block indefinitely.
func TestNewClient_HasTimeout(t *testing.T) {
	c := NewClient("https://api.example.com")
	if c.httpClient.Timeout == 0 {
		t.Error("NewClient: http.Client has no timeout set")
	}
}

// -- Sync ---------------------------------------------------------------------

func TestSync_SuccessSendsArtifact(t *testing.T) {
	var got SyncRequest
	var gotAuth string
	var gotTeam string

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/sync/artifacts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		gotAuth = r.Header.Get("Authorization")
		gotTeam = r.Header.Get("X-Team-Slug")
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
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
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	result, err := NewClient(srv.URL).Sync(context.Background(), "ft_token", "my-team", SyncRequest{
		ProjectID: "proj-1",
		Source:    "local",
		Branch:    "main",
		CommitSHA: "abc123",
		Artifact:  json.RawMessage(`{"fingerprint":"fp-1"}`),
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if result.ID != "sync-1" || !result.IsNew || result.SchemaVersion != "failure-artifact.v1" {
		t.Fatalf("response = %+v", result)
	}
	if gotAuth != "Bearer ft_token" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotTeam != "my-team" {
		t.Errorf("X-Team-Slug = %q", gotTeam)
	}
	if got.ProjectID != "proj-1" || got.Source != "local" || got.Branch != "main" || got.CommitSHA != "abc123" {
		t.Errorf("request metadata = %+v", got)
	}
	if string(got.Artifact) != `{"fingerprint":"fp-1"}` {
		t.Errorf("artifact = %s", got.Artifact)
	}
}

func TestSync_StatusErrors(t *testing.T) {
	cases := []struct {
		name   string
		status int
		want   string
	}{
		{name: "unauthorized", status: http.StatusUnauthorized, want: "unauthorized"},
		{name: "payment required", status: http.StatusPaymentRequired, want: "sync:write"},
		{name: "rate limit", status: http.StatusTooManyRequests, want: "rate limit"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			_, err := NewClient(srv.URL).Sync(context.Background(), "ft_token", "", SyncRequest{
				ProjectID: "proj-1",
				Artifact:  json.RawMessage(`{"fingerprint":"fp-1"}`),
			})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tc.want)
			}
		})
	}
}

func TestSync_LargeErrorBodyTruncated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(strings.Repeat("x", maxErrorBodySize*4)))
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL).Sync(context.Background(), "ft_token", "", SyncRequest{
		ProjectID: "proj-1",
		Artifact:  json.RawMessage(`{"fingerprint":"fp-1"}`),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(err.Error()) > maxErrorBodySize+200 {
		t.Errorf("error message looks unbounded (%d bytes); body may not be capped", len(err.Error()))
	}
}

// ── JSON decode error paths ───────────────────────────────────────────────────

func TestLogin_MalformedSignInResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/sign-in/email", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	_, _, err := NewClient(srv.URL).Login(context.Background(), "u@example.com", "pw", "team", "cli")
	if err == nil {
		t.Fatal("expected error for malformed sign-in response, got nil")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("error = %v, want mention of decode response", err)
	}
}

func TestLogin_MalformedTokenResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/sign-in/email", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user": map[string]string{"id": "u1", "email": "u@example.com"},
		})
	})
	mux.HandleFunc("/v1/auth/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("not json"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	_, _, err := NewClient(srv.URL).Login(context.Background(), "u@example.com", "pw", "team", "cli")
	if err == nil {
		t.Fatal("expected error for malformed token response, got nil")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("error = %v, want mention of decode response", err)
	}
}

func TestVerifyToken_MalformedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL).VerifyToken(context.Background(), "ft_token", "team")
	if err == nil {
		t.Fatal("expected error for malformed verify response, got nil")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("error = %v, want mention of decode response", err)
	}
}

func TestSync_MalformedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL).Sync(context.Background(), "ft_token", "team", SyncRequest{
		ProjectID: "p1",
		Artifact:  json.RawMessage(`{"fingerprint":"fp-1"}`),
	})
	if err == nil {
		t.Fatal("expected error for malformed sync response, got nil")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("error = %v, want mention of decode response", err)
	}
}

// ── Network error paths ───────────────────────────────────────────────────────

func TestLogin_NetworkError(t *testing.T) {
	// Use a server that is immediately closed so the HTTP call fails.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	_, _, err := NewClient(srv.URL).Login(context.Background(), "u@example.com", "pw", "team", "cli")
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
	if !strings.Contains(err.Error(), "sign in") {
		t.Errorf("error = %v, want mention of sign in", err)
	}
}

func TestVerifyToken_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	_, err := NewClient(srv.URL).VerifyToken(context.Background(), "ft_token", "team")
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
	if !strings.Contains(err.Error(), "verify token") {
		t.Errorf("error = %v, want mention of verify token", err)
	}
}

func TestSync_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	_, err := NewClient(srv.URL).Sync(context.Background(), "ft_token", "", SyncRequest{
		ProjectID: "p1",
		Artifact:  json.RawMessage(`{"fingerprint":"fp-1"}`),
	})
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
	if !strings.Contains(err.Error(), "sync") {
		t.Errorf("error = %v, want mention of sync", err)
	}
}
