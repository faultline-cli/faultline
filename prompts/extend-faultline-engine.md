# Extend Faultline Engine

Use this prompt when adding a new failure playbook or modifying engine matching logic.

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md) — architectural boundaries and invariants
- [`internal/playbooks/playbooks.go`](../internal/playbooks/playbooks.go) — playbook loading and validation
- [`internal/matcher/matcher.go`](../internal/matcher/matcher.go) — matching and scoring rules
- An existing playbook in [`playbooks/`](../playbooks/) for the target category

## Goal

Add a new YAML playbook (primary) or engine rule (if pattern logic requires code) that produces stable, explainable output.

## Required Approach

- Prefer a YAML playbook entry over Go code changes unless the match logic cannot be expressed in the pattern DSL.
- Express every rule with explicit `match.any` patterns and `match.none` exclusions where needed.
- Keep patterns narrow: prefer specific strings over broad regexes to avoid false positives.
- Reuse existing playbook structure rather than introducing new fields.
- Ensure the new playbook does not shadow an existing one — run `make review` to check overlap.

## Success Criteria

- `make test` passes.
- `make review` shows no unintended new conflicts.
- A positive log fixture triggers the playbook; a negative fixture does not.

## Failure Modes

- Broad patterns that match unrelated failures → tighten `match.any`, add `match.none`.
- Missing `match.none` causing overlap with an existing playbook → detected by `make review`.

## Deliverable

- Playbook YAML (and engine code only if required)
- Test fixture for positive and negative paths
- Any doc update if a new category is introduced
