# ADR 0005: Command Maturity And Release Boundary

- Status: Accepted
- Date: 2026-04-18

## Context

Faultline now exposes more command surfaces than the default onboarding path, and the repository has an explicit release-boundary contract for what should be considered first-run, companion, and experimental behavior.

The current state is documented in [docs/release-boundary.md](../release-boundary.md), [docs/releases/v0.3.0.md](../releases/v0.3.0.md), and [README.md](../../README.md).

## Decision

Faultline keeps a maturity-tier model for command surfaces:

- stable default path: `analyze`, `batch`, `inspect`, `workflow`, `list`, `explain`, `fix`
- bounded companion surface: `report`
- hidden maintainer workflows: `fixtures`
- provider-backed delta is removed from the shipped `analyze` surface and reserved for a future explicit Team enrichment path

The default narrative and docs must stay centered on the stable path, with
hidden maintainer scope limited to corpus curation and release gates.

As of v0.4.1, **silent failure detection runs automatically** on all analysis via the stable `analyze` and `workflow` commands (see [ADR 0007](0007-silent-failures-as-first-class-detection.md)). Silent findings appear in text, markdown, and JSON output by default; users can opt into failure-exit behavior with `--fail-on-silent`.

## Consequences

- New user-facing capabilities should start hidden, flagged, or non-default until deterministic validation and docs are in place.
- Maintainer workflows can evolve with validation without forcing onboarding complexity into first-run docs.
- Product messaging remains deterministic and narrow, while still preserving depth for advanced users.
- Promotion between maturity tiers should be deliberate and documented, not implicit.

## References

- [docs/release-boundary.md](../release-boundary.md)
- [docs/releases/v0.3.0.md](../releases/v0.3.0.md)
- [README.md](../../README.md)
- [SYSTEM.md](../../SYSTEM.md)
