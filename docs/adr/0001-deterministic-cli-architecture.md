# ADR 0001: Deterministic CLI Architecture

- Status: Accepted
- Date: 2026-04-13

## Context

Faultline is positioned as a deterministic CI failure analysis tool, not a hosted diagnosis service or probabilistic assistant. The repository already reflects a CLI-only architecture with explicit layers for command handling, orchestration, detectors, playbook loading, matching, output, workflow generation, and repository context.

This architectural stance appears in [SYSTEM.md](../../SYSTEM.md), [docs/architecture.md](../architecture.md), [IMPLEMENTATION_PLAN.md](../../IMPLEMENTATION_PLAN.md), and the repository history from the initial CLI-focused baseline onward.

## Decision

Faultline remains a single-purpose CLI with explicit deterministic layers and no runtime ML or LLM dependence in product logic.

The architecture keeps these boundaries stable:

- CLI parsing and command wiring stay in `cmd/` and `internal/cli`
- use-case orchestration stays in `internal/app`
- analysis orchestration stays in `internal/engine`
- detector implementations stay explicit in `internal/detectors`
- playbook loading and validation stay in `internal/playbooks`
- ranking and evidence extraction stay in `internal/matcher`
- rendering and serialization stay in `internal/output` and `internal/renderer`
- repository enrichment stays in `internal/repo`
- follow-up planning stays in `internal/workflow`

## Consequences

- Output stays reproducible for the same input.
- The CLI works the same locally and in CI.
- Architectural drift back toward service, frontend, or probabilistic designs should be treated as a deliberate new decision, not an incremental change.
- Documentation should continue to describe the shipped CLI shape, not speculative product directions.

## References

- [SYSTEM.md](../../SYSTEM.md)
- [docs/architecture.md](../architecture.md)
- [IMPLEMENTATION_PLAN.md](../../IMPLEMENTATION_PLAN.md)
- Git history: `cf15d08` initial repository baseline, followed by the CLI-focused cleanup reflected in current docs