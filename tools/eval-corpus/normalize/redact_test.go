package normalize_test

import (
	"strings"
	"testing"

	"faultline/tools/eval-corpus/normalize"
)

func TestRedactEmails(t *testing.T) {
	opts := normalize.RedactOptions{Emails: true}
	in := "user alice@example.com logged in"
	out := normalize.Redact(opts, in)
	if strings.Contains(out, "alice@example.com") {
		t.Errorf("email not redacted: %q", out)
	}
	if !strings.Contains(out, "<email>") {
		t.Errorf("redacted marker not present: %q", out)
	}
}

func TestRedactTokens(t *testing.T) {
	opts := normalize.RedactOptions{Tokens: true}
	in := "token ghp_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA found"
	out := normalize.Redact(opts, in)
	if strings.Contains(out, "ghp_") {
		t.Errorf("token not redacted: %q", out)
	}
}

func TestRedactNoop(t *testing.T) {
	opts := normalize.RedactOptions{}
	in := "alice@example.com ghp_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	out := normalize.Redact(opts, in)
	if out != in {
		t.Errorf("Redact with all-false opts should be a no-op, got %q", out)
	}
}
