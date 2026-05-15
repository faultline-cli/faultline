package cli

// SurfaceCategory identifies where a command sits in the release boundary.
type SurfaceCategory string

const (
	SurfaceCore         SurfaceCategory = "core"
	SurfaceCompanion    SurfaceCategory = "companion"
	SurfaceMaintainer   SurfaceCategory = "maintainer"
	SurfaceTransitional SurfaceCategory = "transitional"
)

// SurfaceMaturity identifies how prominently a command should appear.
type SurfaceMaturity string

const (
	MaturityShipReady    SurfaceMaturity = "ship-ready"
	MaturitySupported    SurfaceMaturity = "supported"
	MaturityHidden       SurfaceMaturity = "hidden"
	MaturityExperimental SurfaceMaturity = "experimental"
)

// CommandSurface is the internal release-boundary manifest for Cobra commands.
// It is intentionally small and hand-authored so command visibility changes are
// reviewed as product-surface changes, not accidental wiring drift.
type CommandSurface struct {
	Path        string
	Category    SurfaceCategory
	Maturity    SurfaceMaturity
	Hidden      bool
	HiddenFlags []string
}

// CommandSurfaces lists the known CLI surfaces. Public help text should stay
// centered on core commands even when supported companion commands are visible.
var CommandSurfaces = []CommandSurface{
	{Path: "analyze", Category: SurfaceCore, Maturity: MaturityShipReady, HiddenFlags: []string{"delta-provider", "github-branch", "github-repo", "github-run-id", "gitlab-api-base-url", "gitlab-branch", "gitlab-job-id", "gitlab-pipeline-id", "gitlab-project", "metrics-history", "no-history", "no-store", "store"}},
	{Path: "workflow", Category: SurfaceCore, Maturity: MaturityShipReady, HiddenFlags: []string{"metrics-history", "no-history", "no-store", "store"}},
	{Path: "list", Category: SurfaceCore, Maturity: MaturityShipReady},
	{Path: "explain", Category: SurfaceCore, Maturity: MaturityShipReady},
	{Path: "fix", Category: SurfaceCore, Maturity: MaturityShipReady, HiddenFlags: []string{"no-history", "no-store", "store"}},
	{Path: "report", Category: SurfaceCompanion, Maturity: MaturitySupported, HiddenFlags: []string{"store"}},
	{Path: "trace", Category: SurfaceCompanion, Maturity: MaturitySupported, HiddenFlags: []string{"no-history", "no-store", "store"}},
	{Path: "replay", Category: SurfaceCompanion, Maturity: MaturitySupported},
	{Path: "compare", Category: SurfaceCompanion, Maturity: MaturitySupported},
	{Path: "inspect", Category: SurfaceCompanion, Maturity: MaturitySupported, HiddenFlags: []string{"no-history", "no-store", "store"}},
	{Path: "guard", Category: SurfaceCompanion, Maturity: MaturitySupported},
	{Path: "packs", Category: SurfaceCompanion, Maturity: MaturitySupported},
	{Path: "packs install", Category: SurfaceCompanion, Maturity: MaturitySupported},
	{Path: "packs list", Category: SurfaceCompanion, Maturity: MaturitySupported},
	{Path: "history", Category: SurfaceCompanion, Maturity: MaturitySupported},
	{Path: "signatures", Category: SurfaceCompanion, Maturity: MaturityHidden, Hidden: true},
	{Path: "verify-determinism", Category: SurfaceCompanion, Maturity: MaturityHidden, Hidden: true},
	{Path: "coverage", Category: SurfaceCompanion, Maturity: MaturitySupported},
	{Path: "batch", Category: SurfaceCompanion, Maturity: MaturitySupported, HiddenFlags: []string{"no-history", "no-store", "store"}},
	{Path: "fixtures", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures ingest", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures review", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures promote", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures stats", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures sanitize", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures compare-modes", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures scaffold", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
}
