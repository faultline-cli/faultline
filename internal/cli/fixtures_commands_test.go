package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewFixturesCommandRegistersHiddenSubcommands(t *testing.T) {
	cmd := newFixturesCommand()
	if !cmd.Hidden {
		t.Fatal("fixtures command should be hidden")
	}

	want := map[string]bool{
		"ingest":        false,
		"review":        false,
		"promote":       false,
		"stats":         false,
		"sanitize":      false,
		"compare-modes": false,
	}
	for _, child := range cmd.Commands() {
		if _, ok := want[child.Name()]; ok {
			want[child.Name()] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("missing subcommand %q", name)
		}
	}
}

func TestFixturesIngestCommandValidatesRequiredFlags(t *testing.T) {
	t.Run("adapter is required", func(t *testing.T) {
		cmd := newFixturesIngestCommand()
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))

		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "--adapter is required") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("url is required", func(t *testing.T) {
		cmd := newFixturesIngestCommand()
		cmd.SetArgs([]string{"--adapter", "github-issue"})
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))

		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "at least one --url is required") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestFixturesPromoteCommandRequiresExpectedPlaybook(t *testing.T) {
	cmd := newFixturesPromoteCommand()
	cmd.SetArgs([]string{"fixture-123"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--expected-playbook is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFixturesStatsCommandRejectsInvalidClass(t *testing.T) {
	cmd := newFixturesStatsCommand()
	cmd.SetArgs([]string{"--class", "bogus"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid fixture class") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFixturesSanitizeCommandRequiresArgs(t *testing.T) {
	cmd := newFixturesSanitizeCommand()
	cmd.SetArgs([]string{})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no staging IDs are provided")
	}
}
