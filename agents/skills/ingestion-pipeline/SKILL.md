---
name: ingestion-pipeline
description: Use this skill when working on Faultline fixture intake from public sources, including staging review, duplicate rejection, promotion into fixtures/real, and post-promotion baseline validation. Trigger it for requests about running the ingestion pipeline, ingesting fixtures, reviewing staging candidates, promoting accepted cases, or validating new real-world corpus additions.
---

# Ingestion Pipeline

This skill is for running Faultline's deterministic fixture intake loop with deliberate source diversity.

Use it when the task involves:

- `faultline fixtures ingest`
- `faultline fixtures review`
- `faultline fixtures promote`
- validating a promoted fixture against the real-corpus baseline
- building a varied intake batch across multiple public-source adapters

Do not use it for general coding work, broad repository review, or playbook authoring that is not tied to a fixture intake task.

## Read First

- [`SYSTEM.md`](../../../SYSTEM.md)
- [`docs/fixture-corpus.md`](../../../docs/fixture-corpus.md)
- [`docs/agent-workflows.md`](../../../docs/agent-workflows.md)
- [`prompts/run-ingestion-pipeline.md`](../../../prompts/run-ingestion-pipeline.md)
- [`fixtures/staging/README.md`](../../../fixtures/staging/README.md)
- [`docs/modules/source-adapters.md`](../../../docs/modules/source-adapters.md) - Source adapter reference
- [`docs/modules/ingestion-workflow-steps.md`](../../../docs/modules/ingestion-workflow-steps.md) - Core workflow steps
- [`docs/modules/ingestion-guardrails.md`](../../../docs/modules/ingestion-guardrails.md) - Guardrails and rules
- [`docs/modules/ingestion-deliverables.md`](../../../docs/modules/ingestion-deliverables.md) - Deliverables format

## Workflow

Refer to [`docs/modules/ingestion-workflow-steps.md`](../../../docs/modules/ingestion-workflow-steps.md) for the complete step-by-step workflow.

Key execution notes for agents:
1. Follow the modular workflow steps precisely
2. Use source adapters (`github-issue`, `gitlab-issue`, `stackexchange-question`, `discourse-topic`, `reddit-post`) from the reference module
3. Apply all guardrails during execution
4. Document all deliverables as specified

Essential commands include:
- `./bin/faultline fixtures stats --class real --json` (check current corpus mix)
- `./bin/faultline fixtures stats --class real --check-baseline` (validate baseline integrity)
- Bias toward underrepresented adapters when current corpus statistics show gaps
- Handle repository-local risk findings in `internal/engine/testdata/source/` as source-playbook fixtures

## Guardrails

Refer to [`docs/modules/ingestion-guardrails.md`](../../../docs/modules/ingestion-guardrails.md) for complete guardrails, quality standards, and acceptance criteria.

## Deliverable

Refer to [`docs/modules/ingestion-deliverables.md`](../../../docs/modules/ingestion-deliverables.md) for the complete deliverables specification and report structure.
