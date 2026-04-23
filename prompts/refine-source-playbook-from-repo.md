# Refine Source-Playbook From Repo

Use this workflow when `faultline inspect` or `faultline guard` surfaces a repository-local risk that belongs in a bundled `source` playbook.

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md)
- [`docs/detectors.md`](../docs/detectors.md)
- [`docs/fixture-corpus.md`](../docs/fixture-corpus.md)
- [`docs/agent-workflows.md`](../docs/agent-workflows.md)
- [`prompts/run-ingestion-pipeline.md`](./run-ingestion-pipeline.md)
- the nearest bundled source playbook under [`playbooks/bundled/source/`](../playbooks/bundled/source/)
- the relevant repository fixture under [`internal/engine/testdata/source/`](../internal/engine/testdata/source/)

## Goal

Turn a repository-local source finding into one of two explicit outcomes:

- refine the nearest bundled source-playbook
- author a new bundled source-playbook when the root cause boundary is genuinely distinct

## Required Approach

1. Start from the repository finding, not the prose.
2. Run the current source surfaces first:
   - `faultline inspect .`
   - `faultline guard .`
3. Compare the result against the nearest source playbook with `faultline explain <id>`.
4. Decide whether the nearest playbook can be tightened or whether a new source playbook is justified.
5. Keep repository-local evidence in `internal/engine/testdata/source/` instead of promoting it into the real fixture corpus.
6. Add or update the paired positive and nearby negative repository fixtures in the same pass.
7. Populate deterministic `workflow.likely_files`, `workflow.local_repro`, and `workflow.verify` guidance in the playbook.
8. Run, in order:
   - `make review`
   - `make test`
   - `make build`
   - `make cli-smoke`
9. If the change touches accepted real fixtures, also run:
   - `./bin/faultline fixtures stats --class real --check-baseline`

## Guardrails

- Do not force repository-inspection findings into `fixtures/real/`.
- Do not add a new source playbook until the nearest existing one has been considered first.
- Keep matching logic in source triggers, mitigations, and compound signals, not in prose.
- Do not skip the confusable-neighbor check.
- Do not stop after a passing compile; verify the inspect and guard workflow too.

## Deliverable

Report:

- the source playbook ID that changed or the new source playbook ID
- the positive and nearby negative repository fixtures used
- the nearest confusable neighbor and how the boundary is maintained
- the exact validation commands run
