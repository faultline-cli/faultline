---
name: source-playbook-refinement
description: Use this skill when refining or authoring bundled source playbooks from repository-local findings surfaced by inspect or guard. Trigger it for requests about source-detector tuning, adding new bundled source rules, improving workflow guidance on source playbooks, or maintaining paired source fixtures under internal/engine/testdata/source. Use it instead of the ingestion-pipeline skill when the evidence comes from repository inspection rather than public log intake.
---

# Source Playbook Refinement

This skill is for the deterministic loop that keeps repository-local source findings in bundled `source` playbooks and bundled source-playbook regressions.

Use it when the task involves:

- `faultline inspect .`
- `faultline guard .`
- refining an existing bundled `source` playbook
- authoring a new bundled `source` playbook from a repository-local finding
- maintaining paired positive and nearby negative fixtures under `internal/engine/testdata/source/`

Do not use it for:

- public log ingestion or staging review
- accepted real-fixture promotion
- generic playbook authoring that belongs to log playbooks

## Read First

- [`SYSTEM.md`](../../../SYSTEM.md)
- [`docs/detectors.md`](../../../docs/detectors.md)
- [`docs/fixture-corpus.md`](../../../docs/fixture-corpus.md)
- [`docs/agent-workflows.md`](../../../docs/agent-workflows.md)
- [`prompts/refine-source-playbook-from-repo.md`](../../../prompts/refine-source-playbook-from-repo.md)
- the nearest bundled source playbook under [`playbooks/bundled/source/`](../../../playbooks/bundled/source/)

## Workflow

1. Run `faultline inspect .` and `faultline guard .` against the repository or fixture tree that exposed the risk.
2. Compare the top result with `faultline explain <id>` for the nearest bundled source playbook.
3. Prefer the smallest credible change:
   - tighten source triggers
   - narrow mitigations
   - improve compound signals
   - improve `workflow.likely_files`
   - improve `workflow.local_repro`
   - improve `workflow.verify`
4. Add a new bundled source playbook only when the root-cause boundary is clearly distinct.
5. Pair the source playbook with positive and nearby negative repository fixtures under `internal/engine/testdata/source/`.
6. Keep the repository-local evidence out of `fixtures/real/`.
7. Run, in order:
   - `make review`
   - `make test`
   - `make build`
   - `make cli-smoke`
8. If the change affects accepted real fixtures, run `./bin/faultline fixtures stats --class real --check-baseline` as well.

## Guardrails

- Do not broaden a source playbook when a narrower neighbor can be refined instead.
- Do not move repository-inspection findings into the real fixture corpus.
- Keep the source workflow deterministic and fixture-backed.
- Do not skip the inspect/guard check before editing playbooks.

## Deliverable

Report:

- the source playbook or fixtures that motivated the change
- whether the playbook was refined or a new one was justified
- the exact validation commands run
