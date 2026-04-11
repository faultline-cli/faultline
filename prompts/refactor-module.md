# Refactor Module

Use this prompt when improving an existing package without changing product behavior.

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md) — package boundaries and invariants
- The package under refactor and its tests

## Goal

Improve clarity, maintainability, or separation of concerns while preserving behavior.

## Required Approach

- Preserve all public interfaces and observable output unless explicitly asked otherwise.
- Keep the refactor focused on one package or one narrow set of files.
- Avoid introducing new layers unless they demonstrably remove existing complexity.
- Do not change playbook YAML while refactoring Go code.

## Success Criteria

- `make test` passes with no behavior change.
- Output for the same input is byte-for-byte identical before and after.

## Failure Modes

- Accidental public API change → check all call sites in `cmd/` before merging.
- Ordering change in output → detect with `make test` fixture comparison.

## Deliverable

- Refined code
- Confirmation that `make test` still passes
- Notes on any non-obvious structural decisions
