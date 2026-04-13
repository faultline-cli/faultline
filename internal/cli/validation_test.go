package cli

import "testing"

func TestValidateOutputFormat(t *testing.T) {
	cases := []struct {
		value   string
		wantErr bool
	}{
		{"raw", false},
		{"markdown", false},
		{"json", true},
		{"", true},
		{"HTML", true},
	}
	for _, tc := range cases {
		err := validateOutputFormat(tc.value)
		if (err != nil) != tc.wantErr {
			t.Errorf("validateOutputFormat(%q): got err=%v, wantErr=%v", tc.value, err, tc.wantErr)
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
