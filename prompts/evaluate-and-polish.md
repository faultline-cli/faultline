# Evaluate and Polish

Use this prompt when reviewing a finished slice for correctness and cleanup.

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md) — key invariants and non-goals
- The changed files and their tests
- `make test` output
- `make review` output (if any playbook was added or changed)

## Goal

Find missing wiring, edge cases, noise issues, and unnecessary complexity.

## Required Approach

- Run `make test` and `make review` before reviewing code manually.
- Check for completeness, not just compilation: trace each new feature from CLI entry to output.
- Look for early-stopping gaps and hidden TODOs.
- Confirm deterministic ordering: map iteration → sorted slices, file loading → stable sort.
- Remove accidental complexity and unused structure.
- Verify JSON output stability: same input must produce identical JSON across runs.

## Success Criteria

- `make test` passes cleanly.
- No hidden TODOs or incomplete wiring remains.
- Output is deterministic and low-noise.

## Deliverable

- Any final code or doc fixes
- A short list of residual risks or deferred items, if any
