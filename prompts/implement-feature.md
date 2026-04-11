# Implement Feature

Use this prompt when adding a new CLI command, flag, output format, or product behavior that is not a playbook/engine rule change.

> For playbook additions or engine rule changes, use [`extend-faultline-engine.md`](./extend-faultline-engine.md) instead.

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md) — architectural boundaries and invariants
- [`cmd/root.go`](../cmd/root.go) — command wiring
- The package(s) most likely affected by the feature

## Goal

Implement the requested feature completely, with deterministic behavior and minimal architectural change.

## Required Approach

- Identify the smallest complete implementation path across all affected packages.
- Prefer direct code changes over new abstractions.
- Keep outputs stable, ordered, and reproducible.
- Wire the feature end to end — do not stop at compilation.

## Success Criteria

- `make test` passes.
- The feature produces identical output across repeated runs with the same input.
- Any new CLI flag or command is covered by at least one test.

## Failure Modes

- Incomplete wiring (compiles but not reachable via CLI) → trace from `cmd/root.go` to confirm.
- Output ordering that depends on map iteration → replace with sorted slices.

## Deliverable

- Code changes
- Tests when practical
- Doc updates if architecture or public behavior changed
