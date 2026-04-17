# Triage Unmatched Log

Use this workflow when `faultline analyze` returns no deterministic match or when a staging fixture has no credible predicted playbook.

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md)
- [`docs/playbooks.md`](../docs/playbooks.md)
- [`docs/fixture-corpus.md`](../docs/fixture-corpus.md)
- the failing log or staging fixture

## Goal

Turn an unmatched failure into one of three explicit outcomes:

- an existing playbook should be refined
- a new playbook is justified
- the case is too noisy, duplicate, or low-value to add

## Required Approach

1. Reduce the log to stable evidence lines.
2. Check whether the failure is already close to an existing playbook with `faultline list` and `faultline explain <id>`.
3. For public-source evidence, ingest into staging with `faultline fixtures ingest`.
4. Run `faultline fixtures review` to see predicted matches and duplicate hints.
5. Reject duplicates, near-duplicates without new signal, and cases whose fix path is still unclear.
6. Only after that decision, choose between refining an existing playbook or authoring a new one.

## Acceptance Bar

- the root cause is distinct, not just different wording
- the evidence is stable enough for deterministic matching
- the fix path is concrete enough to encode in playbook guidance
- there is a clear place in the fixture corpus for the case

## Deliverable

- a keep or reject decision with justification
- if kept, the next concrete action: refine existing playbook or add a new one
