# Refine Playbook From Fixture

Use this workflow when a real or minimal fixture exposes a weak match, false positive, confusable ranking, or missing workflow guidance.

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md)
- [`docs/playbooks.md`](../docs/playbooks.md)
- [`docs/fixture-corpus.md`](../docs/fixture-corpus.md)
- the target playbook under [`playbooks/bundled/`](../playbooks/bundled/)
- the relevant fixture under [`fixtures/minimal/`](../fixtures/minimal/) or [`fixtures/real/`](../fixtures/real/)

## Goal

Improve an existing deterministic diagnosis before adding any new catalog surface.

## Required Approach

1. Start from the fixture, not the prose.
2. Identify the smallest credible change:
   - tighten `match.any`
   - add or narrow `match.none`
   - improve authored markdown guidance
   - improve `workflow.likely_files`, `workflow.local_repro`, or `workflow.verify`
3. Prefer strengthening the nearest playbook over creating a sibling playbook with overlapping patterns.
4. If a new playbook is truly required, add the positive fixture and at least one nearby negative or adversarial guard.
5. Run `make review` after any playbook change.
6. Run `make test`.
7. Re-run `./bin/faultline fixtures stats --class real --check-baseline` if the change affects the accepted corpus.

## Decision Rules

- Add a new playbook only when the root cause boundary is distinct.
- Improve workflow fields when the diagnosis is correct but the follow-through is generic.
- Improve exclusions when the top result is right but confusable neighbors are too close.
- Reject broad wording that cannot be defended against a nearby fixture.

## Deliverable

- the minimal playbook and fixture changes needed
- the exact conflict or regression checks used to justify the change
