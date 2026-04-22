# ADR 0005: Command Maturity And Release Boundary

- Status: Accepted
- Date: 2026-04-18

## Context

Faultline now exposes more command surfaces than the default onboarding path, and the repository has an explicit release-boundary contract for what should be considered first-run, companion, and experimental behavior.

The current state is documented in [docs/release-boundary.md](../release-boundary.md), [docs/releases/v0.3.0.md](../releases/v0.3.0.md), and [README.md](../../README.md).

## Decision

Faultline keeps a maturity-tier model for command surfaces:

- stable default path: `analyze`, `workflow`, `list`, `explain`, `fix`
- complete companion surfaces: `trace`, `replay`, `compare`, `inspect`, `guard`, `packs`
- experimental opt-in: provider-backed delta behind `FAULTLINE_EXPERIMENTAL_PROVIDER_DELTA=1` (legacy `FAULTLINE_EXPERIMENTAL_GITHUB_DELTA=1` also accepted)

The default narrative and docs must stay centered on the stable path even when companion commands are fully supported and validated.

## Consequences

- New user-facing capabilities should start hidden, flagged, or non-default until deterministic validation and docs are in place.
- Companion commands can evolve with parity-grade validation without forcing onboarding complexity into first-run docs.
- Product messaging remains deterministic and narrow, while still preserving depth for advanced users.
- Promotion between maturity tiers should be deliberate and documented, not implicit.

## References

- [docs/release-boundary.md](../release-boundary.md)
- [docs/releases/v0.3.0.md](../releases/v0.3.0.md)
- [README.md](../../README.md)
- [SYSTEM.md](../../SYSTEM.md)
