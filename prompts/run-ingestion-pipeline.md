# Run Ingestion Pipeline

Use this workflow when the task is to bring new public failure evidence into Faultline through the repository's fixture intake path.

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md)
- [`docs/fixture-corpus.md`](../docs/fixture-corpus.md)
- [`docs/agent-workflows.md`](../docs/agent-workflows.md)
- [`fixtures/staging/README.md`](../fixtures/staging/README.md)
- [`docs/modules/source-adapters.md`](../docs/modules/source-adapters.md) - Source adapter reference
- [`docs/modules/ingestion-workflow-steps.md`](../docs/modules/ingestion-workflow-steps.md) - Core workflow steps
- [`docs/modules/ingestion-guardrails.md`](../docs/modules/ingestion-guardrails.md) - Guardrails and rules
- [`docs/modules/ingestion-deliverables.md`](../docs/modules/ingestion-deliverables.md) - Deliverables format

## Goal

Run the full deterministic ingestion pipeline from public URLs to one of two outcomes:

- rejected from staging
- promoted into `fixtures/real/` with explicit expectations

Prefer a varied batch over a single-source sweep.

## Required Approach

Refer to [`docs/modules/ingestion-workflow-steps.md`](../docs/modules/ingestion-workflow-steps.md) for the complete step-by-step workflow.

Key principles:
1. Follow the modular workflow steps for consistency
2. Use source adapters from the reference module
3. Apply guardrails to maintain quality
4. Document deliverables as specified

## Command Skeleton

Refer to [`docs/modules/ingestion-deliverables.md`](../docs/modules/ingestion-deliverables.md) for the command skeleton reference.

## Acceptance Bar

Refer to [`docs/modules/ingestion-guardrails.md`](../docs/modules/ingestion-guardrails.md) for acceptance criteria and guardrails.

## Source Selection Rules

Refer to [`docs/modules/ingestion-guardrails.md`](../docs/modules/ingestion-guardrails.md) for source selection rules and quality standards.

Available source adapters include: `github-issue`, `gitlab-issue`, `stackexchange-question`, `discourse-topic`, and `reddit-post`. Bias selection toward underrepresented adapters when current corpus statistics show gaps.

Essential validation commands:
- `./bin/faultline fixtures stats --class real --json` (check current corpus mix)
- `./bin/faultline fixtures stats --class real --check-baseline` (validate baseline integrity)

Handle repository-local risk findings in `internal/engine/testdata/source/` as source-playbook fixtures.

## Deliverable

Refer to [`docs/modules/ingestion-deliverables.md`](../docs/modules/ingestion-deliverables.md) for the complete deliverables specification.
