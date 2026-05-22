package cli

import "testing"

func TestFirstNonEmpty(t *testing.T) {
	cases := []struct {
		name  string
		input []string
		want  string
	}{
		{"first non-empty wins", []string{"a", "b"}, "a"},
		{"skips empty strings", []string{"", "b"}, "b"},
		{"skips whitespace-only strings", []string{"  ", "\t", "c"}, "c"},
		{"all empty returns empty", []string{"", "  "}, ""},
		{"no args returns empty", nil, ""},
		{"single non-empty", []string{"hello"}, "hello"},
		{"trims leading/trailing whitespace before comparison", []string{"  ", " value "}, "value"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := firstNonEmpty(tc.input...)
			if got != tc.want {
				t.Errorf("firstNonEmpty(%v) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestFirstInt64(t *testing.T) {
	cases := []struct {
		name  string
		input []interface{}
		want  int64
	}{
		{"returns first non-zero int64", []interface{}{int64(42), int64(7)}, 42},
		{"skips zero int64", []interface{}{int64(0), int64(99)}, 99},
		{"parses string int", []interface{}{"", "123"}, 123},
		{"skips blank string", []interface{}{"  ", int64(5)}, 5},
		{"parses string before int64", []interface{}{"10", int64(20)}, 10},
		{"zero string is skipped", []interface{}{"0", int64(7)}, 7},
		{"all zero returns 0", []interface{}{int64(0), "0"}, 0},
		{"invalid string falls through", []interface{}{"abc", int64(3)}, 3},
		{"no args returns 0", nil, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := firstInt64(tc.input...)
			if got != tc.want {
				t.Errorf("firstInt64(%v) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

// ── resolveAnalyzeStoreSetting ────────────────────────────────────────────────

func TestResolveAnalyzeStoreSetting(t *testing.T) {
	cases := []struct {
		name      string
		noHistory bool
		noStore   bool
		storePath string
		want      string
	}{
		{"noHistory returns off", true, false, "", "off"},
		{"noStore returns off", false, true, "", "off"},
		{"both flags returns off", true, true, "", "off"},
		{"explicit path returned as-is", false, false, "/tmp/fl.db", "/tmp/fl.db"},
		{"no flags and no path returns auto", false, false, "", "auto"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure FAULTLINE_STORE is not set so it doesn't pollute explicit-path tests.
			t.Setenv(storeEnv, "")
			got := resolveAnalyzeStoreSetting(tc.noHistory, tc.noStore, tc.storePath)
			if got != tc.want {
				t.Errorf("resolveAnalyzeStoreSetting(noHistory=%v, noStore=%v, storePath=%q) = %q, want %q",
					tc.noHistory, tc.noStore, tc.storePath, got, tc.want)
			}
		})
	}
}

// ── resolveStoreHistoryOutput ─────────────────────────────────────────────────

func TestResolveStoreHistoryOutput(t *testing.T) {
	cases := []struct {
		name      string
		history   bool
		noHistory bool
		noStore   bool
		storePath string
		want      bool
	}{
		{"noHistory returns false", false, true, false, "", false},
		{"noStore returns false", false, false, true, "", false},
		{"both disable flags returns false", false, true, true, "", false},
		{"history flag returns true", true, false, false, "", true},
		{"explicit path returns true", false, false, false, "/tmp/fl.db", true},
		{"off path returns false", false, false, false, "off", false},
		{"no flags no path returns false", false, false, false, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(storeEnv, "")
			got := resolveStoreHistoryOutput(tc.history, tc.noHistory, tc.noStore, tc.storePath)
			if got != tc.want {
				t.Errorf("resolveStoreHistoryOutput(history=%v, noHistory=%v, noStore=%v, storePath=%q) = %v, want %v",
					tc.history, tc.noHistory, tc.noStore, tc.storePath, got, tc.want)
			}
		})
	}
}

// ── resolveStoreSetting ───────────────────────────────────────────────────────

func TestResolveStoreSetting(t *testing.T) {
	cases := []struct {
		name      string
		history   bool
		noHistory bool
		noStore   bool
		storePath string
		want      string
	}{
		{"noHistory returns off", false, true, false, "", "off"},
		{"noStore returns off", false, false, true, "", "off"},
		{"both disable flags returns off", false, true, true, "", "off"},
		{"explicit path returned as-is", false, false, false, "/tmp/fl.db", "/tmp/fl.db"},
		{"history flag returns auto", true, false, false, "", "auto"},
		{"no flags no path returns off", false, false, false, "", "off"},
		{"noHistory overrides explicit path", false, true, false, "/tmp/fl.db", "off"},
		{"noStore overrides explicit path", false, false, true, "/tmp/fl.db", "off"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(storeEnv, "")
			got := resolveStoreSetting(tc.history, tc.noHistory, tc.noStore, tc.storePath)
			if got != tc.want {
				t.Errorf("resolveStoreSetting(history=%v, noHistory=%v, noStore=%v, storePath=%q) = %q, want %q",
					tc.history, tc.noHistory, tc.noStore, tc.storePath, got, tc.want)
			}
		})
	}
}

func TestResolveStoreSettingEnvVar(t *testing.T) {
	t.Setenv(storeEnv, "/env/path.db")
	got := resolveStoreSetting(false, false, false, "")
	if got != "/env/path.db" {
		t.Errorf("resolveStoreSetting with env var = %q, want %q", got, "/env/path.db")
	}
}

func TestResolveStoreSettingEnvVarIgnoredWhenDisabled(t *testing.T) {
	t.Setenv(storeEnv, "/env/path.db")
	got := resolveStoreSetting(false, true, false, "")
	if got != "off" {
		t.Errorf("resolveStoreSetting noHistory with env var = %q, want %q", got, "off")
	}
}
