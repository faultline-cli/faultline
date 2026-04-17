# Verify Deterministic Change

Use this workflow after any non-trivial repository change.

## When To Use It

- a playbook changed
- fixture expectations changed
- workflow output changed
- matching, ranking, or validation behavior changed
- docs or prompts changed in a way that affects how future agents operate

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md)
- [`AGENTS.md`](../AGENTS.md)
- the changed files

## Goal

Prove the slice is complete, deterministic, and properly wired into the repository workflow.

## Required Approach

1. Restate the exact change boundary before running checks.
2. Run `make test`.
3. Run `make review` when any playbook, match pattern, or playbook-adjacent rule changed.
4. If fixture behavior changed, run `./bin/faultline fixtures stats --class real --check-baseline` after `make build`, or explain why that check is not relevant.
5. Read the affected CLI path or docs path end to end instead of stopping at green tests.
6. Confirm the change did not add generic process, hidden TODOs, or unexplained branching.

## Review Questions

- Does the repository still tell the next agent what deterministic command to run?
- Did this change tighten an existing loop, or did it add surface area without improving the agent workflow?
- If a fixture or playbook was touched, is there a positive and confusable validation path?
- If output changed, is the ordering still stable and reproducible?

## Deliverable

- any final cleanup edits
- exact verification commands run
- any residual risk that still needs checked-in coverage
