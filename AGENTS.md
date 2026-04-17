# Agent Operating Rules

This repository is built for deterministic, iterative agentic development. Treat this file as the default workflow contract for any coding agent working here.

## Expected Behavior

- Read [`SYSTEM.md`](./SYSTEM.md) before making changes.
- Prefer the smallest change that fully solves the task.
- Keep implementation deterministic, explicit, and easy to reason about.
- Favor simple data flow over abstractions, frameworks, or clever indirection.
- Preserve existing intent unless the task explicitly asks for an architectural change.
- Update docs when architecture, boundaries, or execution conventions change.

## How To Work

- Start by locating the relevant packages, docs, and tests.
- Restate the problem in concrete terms before editing.
- Make a short plan when the task spans multiple steps.
- Implement the full requested scope, not just the visible slice.
- Verify completeness with builds, tests, or targeted checks before finishing.
- If a change touches architecture, add or update the relevant docs in the same pass.

## Avoid Early Stopping

- Do not stop after the first compile or the first passing test.
- Check for adjacent breakage, missing files, and incomplete wiring.
- Confirm the feature is connected end to end.
- Confirm the output is deterministic and stable under repeated runs.
- Confirm the repository structure still makes sense to the next agent.

## Scope Safety

- Preserve the default release boundary documented in [`docs/release-boundary.md`](./docs/release-boundary.md).
- Keep the default user narrative centered on `analyze`, `workflow`, `list`, `explain`, and `fix`.
- Treat `inspect`, `guard`, and `packs` as companion surfaces: supported, but not the first-run story.
- Keep maintainer workflows such as `faultline fixtures ...` hidden and out of general-user docs unless the task explicitly targets them.
- Any new or weakly integrated user-facing capability must start hidden, experimental, or non-default until it has deterministic tests, release-grade verification, and docs that match the shipped reality.

## Engineering Rules

- Determinism first: same input must produce the same output.
- Keep the product logic free of ML/LLM dependence.
- Prefer explicit ordering and stable selection rules.
- Avoid speculative interfaces, generics, or abstractions until they are needed.
- Keep the runtime fast and the implementation direct.
- Make noise the exception, not the default.

## Completion Standard

A task is not done until:

- The code is implemented.
- `make test` passes, or any build gap is explicitly documented.
- `make cli-smoke` passes when user-facing output, examples, packaging, or release paths changed.
- `make review` is run after any playbook addition or pattern change.
- The behavior has been checked for completeness.
- Any new architecture or convention is reflected in docs.
- Any new package boundaries are represented in the repo structure.
