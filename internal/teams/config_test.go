package teams

import (
	"os"
	"path/filepath"
	"testing"
)

// useTempConfigDir points XDG_CONFIG_HOME at a fresh temp directory for the
// duration of the test, ensuring credentials are never written to the real
// user config directory during testing.
func useTempConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	return dir
}

// TestSaveAndLoadCredentials verifies a round-trip through the credentials file.
func TestSaveAndLoadCredentials(t *testing.T) {
	useTempConfigDir(t)

	want := &Credentials{
		APIURL:   "https://api.example.com",
		Token:    "ft_abc123",
		TeamSlug: "my-team",
		Email:    "user@example.com",
	}
	if err := SaveCredentials(want); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	got, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if got == nil {
		t.Fatal("LoadCredentials returned nil, want credentials")
	}
	if got.APIURL != want.APIURL {
		t.Errorf("APIURL = %q, want %q", got.APIURL, want.APIURL)
	}
	if got.Token != want.Token {
		t.Errorf("Token = %q, want %q", got.Token, want.Token)
	}
	if got.TeamSlug != want.TeamSlug {
		t.Errorf("TeamSlug = %q, want %q", got.TeamSlug, want.TeamSlug)
	}
	if got.Email != want.Email {
		t.Errorf("Email = %q, want %q", got.Email, want.Email)
	}
}

// TestLoadCredentials_MissingFile confirms that a missing credentials file is
// treated as "not logged in" (returns nil, nil) rather than an error.
func TestLoadCredentials_MissingFile(t *testing.T) {
	useTempConfigDir(t)

	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials on missing file: unexpected error: %v", err)
	}
	if creds != nil {
		t.Fatalf("LoadCredentials on missing file: expected nil, got %+v", creds)
	}
}

// TestSaveCredentials_FilePermissions asserts the credentials file is written
// with 0600 (owner read/write only) so the token is not world-readable.
func TestSaveCredentials_FilePermissions(t *testing.T) {
	useTempConfigDir(t)

	creds := &Credentials{Token: "ft_secret", TeamSlug: "team"}
	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	path, err := CredentialsPath()
	if err != nil {
		t.Fatalf("CredentialsPath: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat credentials file: %v", err)
	}

	const wantPerm = 0600
	if perm := info.Mode().Perm(); perm != wantPerm {
		t.Errorf("credentials file permissions = %04o, want %04o", perm, wantPerm)
	}
}

// TestSaveCredentials_DirPermissions asserts the config directory is created
// with 0700 (owner only) so no other user can enumerate it.
func TestSaveCredentials_DirPermissions(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	creds := &Credentials{Token: "ft_secret", TeamSlug: "team"}
	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	dir := filepath.Join(base, "faultline")
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat config dir: %v", err)
	}

	const wantPerm = 0700
	if perm := info.Mode().Perm(); perm != wantPerm {
		t.Errorf("config dir permissions = %04o, want %04o", perm, wantPerm)
	}
}

// TestClearCredentials removes the credentials file and confirms it is gone.
func TestClearCredentials(t *testing.T) {
	useTempConfigDir(t)

	if err := SaveCredentials(&Credentials{Token: "ft_x", TeamSlug: "t"}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	path, _ := CredentialsPath()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should exist before ClearCredentials: %v", err)
	}

	if err := ClearCredentials(); err != nil {
		t.Fatalf("ClearCredentials: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("credentials file still exists after ClearCredentials")
	}
}

// TestClearCredentials_Idempotent confirms that calling ClearCredentials when
// no file exists returns nil rather than an error.
func TestClearCredentials_Idempotent(t *testing.T) {
	useTempConfigDir(t)

	if err := ClearCredentials(); err != nil {
		t.Fatalf("ClearCredentials on missing file: %v", err)
	}
}

// TestCredentialsPath_UsesXDGConfigHome confirms the path is rooted under
// XDG_CONFIG_HOME when set, following the XDG Base Directory Specification.
func TestCredentialsPath_UsesXDGConfigHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	path, err := CredentialsPath()
	if err != nil {
		t.Fatalf("CredentialsPath: %v", err)
	}

	want := filepath.Join(dir, "faultline", "credentials")
	if path != want {
		t.Errorf("CredentialsPath = %q, want %q", path, want)
	}
}

// TestLoadCredentials_CorruptJSON confirms a malformed credentials file
// returns an error rather than silently producing a zero-value struct.
func TestLoadCredentials_CorruptJSON(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	dir := filepath.Join(base, "faultline")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(dir, "credentials")
	if err := os.WriteFile(path, []byte("not valid json"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := LoadCredentials()
	if err == nil {
		t.Fatal("expected error for corrupt credentials file, got nil")
	}
}
