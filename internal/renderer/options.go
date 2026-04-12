package renderer

import (
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	defaultWidth = 88
	minWidth     = 60
	maxWidth     = 120
)

type Options struct {
	Plain          bool
	Width          int
	DarkBackground bool
}

func DetectOptions(w io.Writer) Options {
	opts := Options{
		Plain: true,
		Width: defaultWidth,
	}

	if forcePlain() {
		return opts
	}

	file, ok := w.(*os.File)
	if !ok {
		return opts
	}
	if !term.IsTerminal(int(file.Fd())) {
		return opts
	}

	opts.Plain = false
	if width, _, err := term.GetSize(int(file.Fd())); err == nil {
		opts.Width = clampWidth(width)
	}
	if value := strings.TrimSpace(os.Getenv("FAULTLINE_DARK_BACKGROUND")); value != "" {
		opts.DarkBackground = value == "1" || strings.EqualFold(value, "true")
	} else {
		opts.DarkBackground = true
	}
	return opts
}

func forcePlain() bool {
	for _, key := range []string{"NO_COLOR", "CI", "GITHUB_ACTIONS", "FAULTLINE_PLAIN"} {
		if strings.TrimSpace(os.Getenv(key)) != "" {
			return true
		}
	}
	return false
}

func clampWidth(width int) int {
	if width < minWidth {
		return minWidth
	}
	if width > maxWidth {
		return maxWidth
	}
	return width
}
