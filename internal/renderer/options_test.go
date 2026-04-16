package renderer

import (
	"bytes"
	"os"
	"testing"
)

func TestDetectOptionsReturnsDefaultsForPlainWriters(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	opts := DetectOptions(&bytes.Buffer{})
	if !opts.Plain || opts.Width != defaultWidth || opts.DarkBackground {
		t.Fatalf("unexpected options for forced plain output: %+v", opts)
	}

	t.Setenv("NO_COLOR", "")
	opts = DetectOptions(&bytes.Buffer{})
	if !opts.Plain || opts.Width != defaultWidth || opts.DarkBackground {
		t.Fatalf("unexpected options for non-file writer: %+v", opts)
	}
}

func TestDetectOptionsReturnsDefaultsForNonTerminalFile(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "renderer-options-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer file.Close()

	opts := DetectOptions(file)
	if !opts.Plain || opts.Width != defaultWidth || opts.DarkBackground {
		t.Fatalf("unexpected options for non-terminal file: %+v", opts)
	}
}

func TestForcePlainRecognizesSupportedEnvironmentVariables(t *testing.T) {
	for _, key := range []string{"NO_COLOR", "CI", "GITHUB_ACTIONS", "FAULTLINE_PLAIN"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv(key, "1")
			if !forcePlain() {
				t.Fatalf("forcePlain() = false, want true for %s", key)
			}
		})
	}
}

func TestClampWidthAppliesBounds(t *testing.T) {
	cases := []struct {
		width int
		want  int
	}{
		{20, minWidth},
		{defaultWidth, defaultWidth},
		{200, maxWidth},
	}
	for _, tc := range cases {
		if got := clampWidth(tc.width); got != tc.want {
			t.Fatalf("clampWidth(%d) = %d, want %d", tc.width, got, tc.want)
		}
	}
}
