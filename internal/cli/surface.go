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

// CommandSurfaces lists the known CLI surfaces. Public help text stays centered
// on core commands; Team companion commands remain callable but hidden from the
// default first-run help narrative.
var CommandSurfaces = []CommandSurface{
	{Path: "analyze", Category: SurfaceCore, Maturity: MaturityShipReady, HiddenFlags: []string{"no-history", "no-store", "store"}},
	{Path: "workflow", Category: SurfaceCore, Maturity: MaturityShipReady, HiddenFlags: []string{"no-history", "no-store", "store"}},
	{Path: "list", Category: SurfaceCore, Maturity: MaturityShipReady},
	{Path: "explain", Category: SurfaceCore, Maturity: MaturityShipReady},
	{Path: "fix", Category: SurfaceCore, Maturity: MaturityShipReady, HiddenFlags: []string{"no-history", "no-store", "store"}},
	{Path: "batch", Category: SurfaceCore, Maturity: MaturityShipReady, HiddenFlags: []string{"no-history", "no-store", "store"}},
	{Path: "inspect", Category: SurfaceCore, Maturity: MaturityShipReady, HiddenFlags: []string{"no-history", "no-store", "store"}},
	{Path: "report", Category: SurfaceCompanion, Maturity: MaturitySupported, HiddenFlags: []string{"store"}},
	{Path: "auth", Category: SurfaceCompanion, Maturity: MaturitySupported, Hidden: true},
	{Path: "auth login", Category: SurfaceCompanion, Maturity: MaturitySupported, Hidden: true},
	{Path: "auth logout", Category: SurfaceCompanion, Maturity: MaturitySupported, Hidden: true},
	{Path: "auth status", Category: SurfaceCompanion, Maturity: MaturitySupported, Hidden: true},
	{Path: "auth token", Category: SurfaceCompanion, Maturity: MaturitySupported, Hidden: true},
	{Path: "auth token set", Category: SurfaceCompanion, Maturity: MaturitySupported, Hidden: true},
	{Path: "sync", Category: SurfaceCompanion, Maturity: MaturitySupported, Hidden: true},
	{Path: "fixtures", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures ingest", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures review", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures promote", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures stats", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures sanitize", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures compare-modes", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures patterns", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "fixtures pack-check", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "catalogue", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
	{Path: "catalogue export", Category: SurfaceMaintainer, Maturity: MaturityHidden, Hidden: true},
}
