package cli

import "testing"

func TestResolveOutputSelection(t *testing.T) {
	t.Run("json flag promotes terminal format", func(t *testing.T) {
		format, jsonOut, err := resolveOutputSelection("terminal", true)
		if err != nil {
			t.Fatalf("resolveOutputSelection: %v", err)
		}
		if format != "json" || !jsonOut {
			t.Fatalf("got format=%q json=%v, want json/true", format, jsonOut)
		}
	})

	t.Run("json flag rejects markdown", func(t *testing.T) {
		_, _, err := resolveOutputSelection("markdown", true)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestNewRootCommandRegistersExpectedCommands(t *testing.T) {
	cmd := NewRootCommand("test")

	want := map[string]bool{
		"analyze":  false,
		"inspect":  false,
		"guard":    false,
		"fix":      false,
		"list":     false,
		"explain":  false,
		"workflow": false,
		"packs":    false,
		"fixtures": false,
	}
	for _, child := range cmd.Commands() {
		if _, ok := want[child.Name()]; ok {
			want[child.Name()] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("missing command %q", name)
		}
	}
}
