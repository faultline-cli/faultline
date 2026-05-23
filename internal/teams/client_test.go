package teams

import (
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
	token, email, err := client.Login("user@example.com", "password", "my-team", "faultline-cli")
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

	_, _, err := NewClient(srv.URL).Login("bad@example.com", "wrong", "team", "cli")
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

	_, _, err := NewClient(srv.URL).Login("bad@example.com", "wrong", "team", "cli")
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

	_, _, err := NewClient(srv.URL).Login("user@example.com", "pw", "team", "cli")
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

	_, _, err := NewClient(srv.URL).Login("user@example.com", "pw", "team", "cli")
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

	_, _, err := NewClient(srv.URL).Login("u@example.com", "pw", "team", "cli")
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

	email, err := NewClient(srv.URL).VerifyToken("ft_valid_token", "my-team")
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

	_, err := NewClient(srv.URL).VerifyToken("ft_bad", "team")
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

	_, err := NewClient(srv.URL).VerifyToken("ft_token", "team")
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

	email, err := NewClient(srv.URL).VerifyToken("ft_token", "")
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
