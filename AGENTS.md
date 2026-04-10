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
- The repo still builds or the build gap is explicitly documented.
- The behavior has been checked for completeness.
- Any new architecture or convention is reflected in docs.
- Any new package boundaries are represented in the repo structure.
