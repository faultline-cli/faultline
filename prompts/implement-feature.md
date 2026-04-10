# Implement Feature

Use this prompt when adding a new slice of product behavior.

## Goal

Implement the requested feature completely, with deterministic behavior and minimal architectural change.

## Required Approach

- Read the relevant docs and package boundaries first.
- Identify the smallest complete implementation path.
- Prefer direct code changes over new abstractions.
- Keep outputs stable, ordered, and reproducible.
- Verify the feature end to end before stopping.

## Deliverable

- Code changes
- Tests when practical
- Doc updates if architecture or behavior changed
