package teams

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// maxErrorBodySize caps how many bytes we read from unexpected server error
// responses. A remote server returning an unbounded body could otherwise exhaust
// memory and cause a denial-of-service.
const maxErrorBodySize = 4096

// Client is a minimal HTTP client for the Faultline Teams API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient returns a Teams API client targeting baseURL.
func NewClient(baseURL string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}
}

// signInResponse is the minimal shape we need from Better Auth's sign-in response.
type signInResponse struct {
	User struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

// createTokenResponse is returned by POST /v1/auth/tokens.
type createTokenResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

// meResponse is returned by GET /v1/auth/me.
type meResponse struct {
	Data struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"data"`
}

// Login authenticates with Better Auth, then creates a persistent API token.
// Returns the raw ft_ token and the verified user email.
func (c *Client) Login(email, password, teamSlug, tokenName string) (token string, userEmail string, err error) {
	// Step 1: sign in via Better Auth's email/password handler.
	signInPayload, err := json.Marshal(map[string]any{
		"email":      email,
		"password":   password,
		"rememberMe": true,
	})
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/auth/sign-in/email", bytes.NewReader(signInPayload))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("sign in: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// success — fall through
	case http.StatusUnauthorized, http.StatusBadRequest:
		return "", "", fmt.Errorf("sign in failed: invalid email or password")
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		return "", "", fmt.Errorf("sign in failed: server returned %d: %s", resp.StatusCode, string(body))
	}

	var signIn signInResponse
	if err := json.NewDecoder(resp.Body).Decode(&signIn); err != nil {
		return "", "", fmt.Errorf("sign in: decode response: %w", err)
	}
	userEmail = signIn.User.Email

	// Step 2: create a persistent API token using the active session cookie.
	createPayload, err := json.Marshal(map[string]string{
		"name": tokenName,
	})
	if err != nil {
		return "", "", err
	}

	tokenReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/v1/auth/tokens", bytes.NewReader(createPayload))
	if err != nil {
		return "", "", err
	}
	tokenReq.Header.Set("Content-Type", "application/json")
	tokenReq.Header.Set("X-Team-Slug", teamSlug)

	tokenResp, err := c.httpClient.Do(tokenReq)
	if err != nil {
		return "", "", fmt.Errorf("create token: %w", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(tokenResp.Body, maxErrorBodySize))
		return "", "", fmt.Errorf("create token: server returned %d: %s", tokenResp.StatusCode, string(body))
	}

	var tokenData createTokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
		return "", "", fmt.Errorf("create token: decode response: %w", err)
	}
	return tokenData.Token, userEmail, nil
}

// VerifyToken calls GET /v1/auth/me with the stored token to confirm it is
// still valid. Returns the authenticated user's email on success.
func (c *Client) VerifyToken(token, teamSlug string) (email string, err error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/v1/auth/me", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if teamSlug != "" {
		req.Header.Set("X-Team-Slug", teamSlug)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf("token is invalid or expired")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("verify token: server returned %d", resp.StatusCode)
	}

	var me meResponse
	if err := json.NewDecoder(resp.Body).Decode(&me); err != nil {
		return "", fmt.Errorf("verify token: decode response: %w", err)
	}
	return me.Data.Email, nil
}
