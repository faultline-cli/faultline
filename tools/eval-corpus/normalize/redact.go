package normalize

import (
	"regexp"
)

// RedactOptions controls which sensitive patterns are scrubbed.
type RedactOptions struct {
	// Emails replaces RFC-5321 email addresses with "<email>".
	Emails bool
	// Tokens replaces bearer tokens and common API key patterns with "<token>".
	Tokens bool
}

var (
	// emailPattern matches common email addresses.
	emailPattern = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

	// tokenPatterns covers common secret formats found in CI logs.
	tokenPatterns = []*regexp.Regexp{
		// Bearer / Authorization header values
		regexp.MustCompile(`(?i)\bbearer\s+[A-Za-z0-9._~+/\-]+=*`),
		// GitHub personal access tokens (classic and fine-grained)
		regexp.MustCompile(`\bghp_[A-Za-z0-9]{36,}\b`),
		regexp.MustCompile(`\bgho_[A-Za-z0-9]{36,}\b`),
		regexp.MustCompile(`\bghr_[A-Za-z0-9]{36,}\b`),
		regexp.MustCompile(`\bghx_[A-Za-z0-9]{36,}\b`),
		// Generic base64-encoded secrets of 30+ characters
		regexp.MustCompile(`\b[A-Za-z0-9+/]{30,}={0,2}\b`),
		// Common API key prefixes
		regexp.MustCompile(`\bsk_(?:live|test)_[A-Za-z0-9]{20,}\b`),
		regexp.MustCompile(`\bpk_(?:live|test)_[A-Za-z0-9]{20,}\b`),
	}
)

// Redact applies the configured scrubbing rules to s and returns the result.
// Replacement markers are stable strings suitable for repeated processing.
func Redact(opts RedactOptions, s string) string {
	if opts.Emails {
		s = emailPattern.ReplaceAllString(s, "<email>")
	}
	if opts.Tokens {
		for _, re := range tokenPatterns {
			s = re.ReplaceAllString(s, "<token>")
		}
	}
	return s
}
