package normalize_test

import (
	"strings"
	"testing"

	"faultline/tools/eval-corpus/normalize"
)

func TestFixtureIDStable(t *testing.T) {
	id1 := normalize.FixtureID("hello world")
	id2 := normalize.FixtureID("hello world")
	if id1 != id2 {
		t.Errorf("FixtureID not stable: %q != %q", id1, id2)
	}
}

func TestFixtureIDDifferentContent(t *testing.T) {
	id1 := normalize.FixtureID("log A")
	id2 := normalize.FixtureID("log B")
	if id1 == id2 {
		t.Error("different content produced the same ID")
	}
}

func TestFixtureIDTrimsWhitespace(t *testing.T) {
	id1 := normalize.FixtureID("log line")
	id2 := normalize.FixtureID("  log line  ")
	if id1 != id2 {
		t.Errorf("leading/trailing whitespace should not affect ID: %q != %q", id1, id2)
	}
}

func TestFixtureIDIsHex(t *testing.T) {
	id := normalize.FixtureID("test")
	if len(id) != 64 {
		t.Errorf("ID length = %d, want 64", len(id))
	}
	for _, ch := range id {
		if !strings.ContainsRune("0123456789abcdef", ch) {
			t.Errorf("ID contains non-hex character: %q", string(ch))
		}
	}
}

func TestParseSizeValid(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"", 0},
		{"0", 0},
		{"100b", 100},
		{"1kb", 1024},
		{"2KB", 2048},
		{"1mb", 1 << 20},
		{"1MB", 1 << 20},
		{"1gb", 1 << 30},
		{"200kb", 200 * 1024},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := normalize.ParseSize(tt.input)
			if err != nil {
				t.Fatalf("ParseSize(%q): %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSizeInvalid(t *testing.T) {
	cases := []string{"-1kb", "abc", "1.5mb"}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if _, err := normalize.ParseSize(c); err == nil {
				t.Errorf("ParseSize(%q) should return error", c)
			}
		})
	}
}

func TestApplyMaxSizeTruncates(t *testing.T) {
	s := strings.Repeat("x", 1000)
	out := normalize.ApplyMaxSize(s, 100)
	if len(out) <= 100 {
		// truncated plus sentinel, should be > 100 chars total but content capped
		t.Log("truncation applied")
	}
	if strings.HasPrefix(out, strings.Repeat("x", 100)) {
		// OK: first 100 chars preserved
	} else {
		t.Error("truncated output should start with first maxBytes characters")
	}
	if !strings.Contains(out, "[truncated]") {
		t.Error("truncated output should contain sentinel text")
	}
}

func TestApplyMaxSizeNoLimit(t *testing.T) {
	s := strings.Repeat("x", 1000)
	out := normalize.ApplyMaxSize(s, 0)
	if out != s {
		t.Error("maxBytes=0 should return unchanged string")
	}
}
