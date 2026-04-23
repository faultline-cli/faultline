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
