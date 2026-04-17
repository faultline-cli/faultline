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

func TestValidateWorkflowMode(t *testing.T) {
	cases := []struct {
		value   string
		wantErr bool
	}{
		{"local", false},
		{"agent", false},
		{"ci", true},
		{"", true},
	}
	for _, tc := range cases {
		err := validateWorkflowMode(tc.value)
		if (err != nil) != tc.wantErr {
			t.Errorf("validateWorkflowMode(%q): got err=%v, wantErr=%v", tc.value, err, tc.wantErr)
		}
	}
}

func TestValidateExperimentalDeltaProvider(t *testing.T) {
	t.Run("disabled by default", func(t *testing.T) {
		t.Setenv(experimentalGitHubDeltaEnv, "")
		err := validateExperimentalDeltaProvider("github-actions")
		if err == nil || !strings.Contains(err.Error(), experimentalGitHubDeltaEnv) {
			t.Fatalf("expected experimental env error, got %v", err)
		}
	})

	t.Run("enabled explicitly", func(t *testing.T) {
		t.Setenv(experimentalGitHubDeltaEnv, "1")
		if err := validateExperimentalDeltaProvider("github-actions"); err != nil {
			t.Fatalf("expected provider to be allowed, got %v", err)
		}
	})

	t.Run("empty provider passes", func(t *testing.T) {
		t.Setenv(experimentalGitHubDeltaEnv, "")
		if err := validateExperimentalDeltaProvider(""); err != nil {
			t.Fatalf("expected empty provider to pass, got %v", err)
		}
	})
}
