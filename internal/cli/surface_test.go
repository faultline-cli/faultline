package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandSurfaceManifestMatchesCobra(t *testing.T) {
	root := NewRootCommand("test")
	actual := map[string]*cobra.Command{}
	walkCommands(root, nil, actual)

	for _, surface := range CommandSurfaces {
		cmd := actual[surface.Path]
		if cmd == nil {
			t.Fatalf("surface manifest references missing command %q", surface.Path)
		}
		if cmd.Hidden != surface.Hidden {
			t.Fatalf("%s hidden mismatch: manifest=%v cobra=%v", surface.Path, surface.Hidden, cmd.Hidden)
		}
		for _, name := range surface.HiddenFlags {
			flag := cmd.Flags().Lookup(name)
			if flag == nil {
				t.Fatalf("%s manifest references missing flag --%s", surface.Path, name)
			}
			if !flag.Hidden {
				t.Fatalf("%s --%s must stay hidden per surface manifest", surface.Path, name)
			}
		}
	}

	for path, cmd := range actual {
		if path == "" {
			continue
		}
		if !slices.ContainsFunc(CommandSurfaces, func(surface CommandSurface) bool { return surface.Path == path }) {
			t.Fatalf("command %q is not classified in CommandSurfaces (hidden=%v)", path, cmd.Hidden)
		}
	}
}

func TestVisibleCommandSurfaceStaysWithinReleaseBoundary(t *testing.T) {
	for _, surface := range CommandSurfaces {
		if surface.Hidden {
			continue
		}
		switch surface.Category {
		case SurfaceCore, SurfaceCompanion:
		default:
			t.Fatalf("visible command %q has non-public release category %q", surface.Path, surface.Category)
		}
		if surface.Maturity == MaturityHidden || surface.Maturity == MaturityExperimental {
			t.Fatalf("visible command %q has non-visible maturity %q", surface.Path, surface.Maturity)
		}
	}
}

func TestHiddenCommandSurfaceLimitedToFixtures(t *testing.T) {
	for _, surface := range CommandSurfaces {
		if !surface.Hidden {
			continue
		}
		isMaintainerFixture := surface.Category == SurfaceMaintainer && (strings.HasPrefix(surface.Path, "fixtures") || strings.HasPrefix(surface.Path, "catalogue"))
		isTeamCompanion := surface.Category == SurfaceCompanion && (strings.HasPrefix(surface.Path, "auth") || surface.Path == "sync")
		if !isMaintainerFixture && !isTeamCompanion {
			t.Fatalf("hidden command %q must be a fixtures maintainer workflow or Team companion surface", surface.Path)
		}
	}
}

func TestShipReadyCoreCommandSet(t *testing.T) {
	want := []string{"analyze", "workflow", "list", "explain", "fix", "batch", "inspect"}
	got := []string{}
	for _, surface := range CommandSurfaces {
		if surface.Category == SurfaceCore && surface.Maturity == MaturityShipReady {
			got = append(got, surface.Path)
		}
	}
	if !slices.Equal(got, want) {
		t.Fatalf("ship-ready core commands = %v, want %v", got, want)
	}
}

func walkCommands(cmd *cobra.Command, prefix []string, out map[string]*cobra.Command) {
	for _, child := range cmd.Commands() {
		if child.IsAdditionalHelpTopicCommand() {
			continue
		}
		parts := append(append([]string(nil), prefix...), child.Name())
		path := strings.Join(parts, " ")
		out[path] = child
		walkCommands(child, parts, out)
	}
}
