# ADR 0006: Snapshot-Backed Companion Surface Validation

- Status: Accepted
- Date: 2026-04-18

## Context

v0.3.0 expanded CLI smoke coverage beyond the default flow to include deterministic validation for companion command surfaces.

The repository now checks replay, compare, and trace output against checked-in snapshots and exercises inspect/guard behavior through a deterministic temporary repository setup in `scripts/cli-smoke.sh`.

## Decision

Companion command surfaces must be validated through deterministic, checked-in smoke evidence as part of the standard release gate.

This includes:

- checked-in snapshots for companion outputs under `examples/`
- CLI smoke assertions for `trace`, `replay`, `compare`, `inspect`, and `guard`
- release-gate integration through `make release-check`

## Consequences

- Output regressions on companion surfaces are detected before release cut.
- Validation expectations for non-default commands are explicit and repeatable.
- Docs, snapshots, and smoke scripts need to stay synchronized when renderer or command behavior changes.
- Companion coverage quality no longer depends on ad hoc manual checks.

## References

- [scripts/cli-smoke.sh](../../scripts/cli-smoke.sh)
- [examples/README.md](../../examples/README.md)
- [docs/releases/v0.3.0.md](../releases/v0.3.0.md)
- [Makefile](../../Makefile)
