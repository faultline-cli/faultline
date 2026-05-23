package teams

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	// DefaultAPIURL is the production Faultline Teams API base URL.
	DefaultAPIURL = "https://api.faultline.dev"

	credentialsFileName = "credentials"
)

// Credentials holds the stored Teams authentication state written to disk.
// The file is created at ~/.config/faultline/credentials (0600) so that
// other users on the same machine cannot read the token.
type Credentials struct {
	APIURL   string `json:"api_url"`
	Token    string `json:"token"`
	TeamSlug string `json:"team_slug"`
	Email    string `json:"email,omitempty"`
}

// configDir returns the XDG-compliant faultline config directory.
func configDir() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "faultline"), nil
}

// CredentialsPath returns the absolute path to the credentials file.
func CredentialsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, credentialsFileName), nil
}

// LoadCredentials reads stored credentials from disk.
// Returns nil, nil when the file does not exist (not logged in).
func LoadCredentials() (*Credentials, error) {
	path, err := CredentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// SaveCredentials persists credentials to disk with owner-only permissions.
func SaveCredentials(creds *Credentials) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path, err := CredentialsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	// 0600 — owner read/write only; token must not be world-readable.
	return os.WriteFile(path, data, 0600)
}

// ClearCredentials removes the credentials file. Returns nil when the file
// does not exist (idempotent).
func ClearCredentials() error {
	path, err := CredentialsPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); os.IsNotExist(err) {
		return nil
	} else {
		return err
	}
}
