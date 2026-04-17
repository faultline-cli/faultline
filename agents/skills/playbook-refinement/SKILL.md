---
name: playbook-refinement
description: Use this skill when improving Faultline playbooks from fixture evidence, especially for weak matches, false positives, confusable neighbors, missing workflow guidance, or deciding whether to refine an existing playbook instead of adding a new one. Trigger it for requests about fixture-driven playbook updates, overlap reduction, validation after playbook edits, or improving workflow fields such as likely_files, local_repro, and verify.
---

# Playbook Refinement

This skill is for fixture-driven improvement of Faultline playbooks.

Use it when the task involves:

- a weak or incorrect fixture match
- a false positive or confusable neighbor
- improving workflow handoff quality in an existing playbook
- deciding whether a new playbook is actually justified

Do not use it for generic feature implementation or unrelated Go refactors.

## Read First

- [`SYSTEM.md`](../../../SYSTEM.md)
- [`docs/playbooks.md`](../../../docs/playbooks.md)
- [`docs/fixture-corpus.md`](../../../docs/fixture-corpus.md)
- [`docs/agent-workflows.md`](../../../docs/agent-workflows.md)
- [`prompts/refine-playbook-from-fixture.md`](../../../prompts/refine-playbook-from-fixture.md)

Then read only the specific playbook and fixtures involved in the task.

## Workflow

1. Start from the failing, weak, or confusable fixture.
2. Inspect the nearest existing playbook before considering a new one.
3. Prefer the smallest credible improvement:
   - tighten `match.any`
   - add or narrow `match.none`
   - improve authored guidance
   - improve `workflow.likely_files`
   - improve `workflow.local_repro`
   - improve `workflow.verify`
4. Add a new playbook only when the root-cause boundary is clearly distinct.
5. When adding a new playbook, also add the positive and nearby negative or adversarial regression coverage needed to defend it.

## Required Validation

- `make review`
- `make test`

If accepted real fixtures were affected, also run:

- `make build`
- `./bin/faultline fixtures stats --class real --check-baseline`

## Guardrails

- Refine before expanding the catalog.
- Avoid broad phrases that cannot be defended against nearby fixtures.
- Keep matching logic in structured fields, not prose.
- Do not stop at a passing compile or a single passing test.

## Deliverable

Report:

- the fixture or fixtures that motivated the change
- whether the playbook was refined or a new one was justified
- the exact validation commands run
