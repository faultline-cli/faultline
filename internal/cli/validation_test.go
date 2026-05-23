package cli

import (
	"strings"
	"testing"
)

func TestValidateOutputFormat(t *testing.T) {
	cases := []struct {
		value   string
		want    string
		wantErr bool
	}{
		{"terminal", "terminal", false},
		{"markdown", "markdown", false},
		{"json", "json", false},
		{"raw", "", true},
		{"md", "", true},
		{"", "", true},
	}
	for _, tc := range cases {
		got, err := validateOutputFormat(tc.value)
		if (err != nil) != tc.wantErr {
			t.Errorf("validateOutputFormat(%q): got err=%v, wantErr=%v", tc.value, err, tc.wantErr)
			continue
		}
		if string(got) != tc.want {
			t.Errorf("validateOutputFormat(%q): got=%q want=%q", tc.value, got, tc.want)
		}
	}
}

func TestValidateOutputMode(t *testing.T) {
	cases := []struct {
		value   string
		wantErr bool
	}{
		{"quick", false},
		{"detailed", false},
		{"verbose", true},
		{"", true},
	}
	for _, tc := range cases {
		err := validateOutputMode(tc.value)
		if (err != nil) != tc.wantErr {
			t.Errorf("validateOutputMode(%q): got err=%v, wantErr=%v", tc.value, err, tc.wantErr)
		}
	}
}

func TestValidateSelect(t *testing.T) {
	cases := []struct {
		value   int
		wantErr bool
	}{
		{1, false},
		{0, false},
		{10, false},
		{-1, true},
		{-100, true},
	}
	for _, tc := range cases {
		err := validateSelect(tc.value)
		if (err != nil) != tc.wantErr {
			t.Errorf("validateSelect(%d): got err=%v, wantErr=%v", tc.value, err, tc.wantErr)
		}
	}
}

func TestValidateAnalyzeView(t *testing.T) {
	cases := []struct {
		value   string
		want    string
		wantErr bool
	}{
		// Valid non-trace views pass through
		{"", "", false},
		{"summary", "summary", false},
		{"evidence", "evidence", false},
		{"fix", "fix", false},
		{"raw", "raw", false},
		// trace view was removed from analyze
		{"trace", "", true},
		// Invalid values return error
		{"invalid", "", true},
		{"json", "", true},
		// Case-insensitive matching
		{"SUMMARY", "summary", false},
		{"Fix", "fix", false},
	}
	for _, tc := range cases {
		got, err := validateAnalyzeView(tc.value)
		if (err != nil) != tc.wantErr {
			t.Errorf("validateAnalyzeView(%q): got err=%v, wantErr=%v", tc.value, err, tc.wantErr)
			continue
		}
		if !tc.wantErr && string(got) != tc.want {
			t.Errorf("validateAnalyzeView(%q): got=%q want=%q", tc.value, got, tc.want)
		}
		if tc.value == "trace" && err != nil && !strings.Contains(err.Error(), "trace") {
			t.Errorf("validateAnalyzeView(trace): expected error mentioning 'trace', got: %v", err)
		}
	}
}
