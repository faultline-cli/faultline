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

## Workflow

1. Confirm the source material is public and appropriate for ingestion.
2. Prefer a mixed batch across multiple adapters when possible:
   - `github-issue`
   - `gitlab-issue`
   - `stackexchange-question`
   - `discourse-topic`
   - `reddit-post`
3. Prefer breadth over depth:
   - take one or two strong URLs from a source before returning to the same thread family
   - do not over-harvest a single issue or discussion unless it is producing clearly distinct failure signatures
4. Run `faultline fixtures ingest --adapter ... --url ...` for each chosen URL.
5. Run `faultline fixtures review`.
6. Classify staged results:
   - reject duplicates
   - reject near-duplicates without meaningful new signal
   - reject setup-only or workaround-only snippets
   - reject anything that still needs sanitization
   - keep only candidates with a credible expected playbook and useful regression value
7. Promote accepted cases with `faultline fixtures promote <staging-id> --expected-playbook <id>`.
8. Add `--strict-top-1`, `--expected-stage`, or `--disallow` only when the boundary needs that extra protection.
9. Run:
   - `make build`
   - `./bin/faultline fixtures stats --class real --check-baseline`
10. If the new fixture exposes weak matching or a confusable neighbor, switch to the `playbook-refinement` skill before stopping.

## Guardrails

- Do not promote raw staging files just because ingestion succeeded.
- Do not skip sanitization.
- Do not add new product logic when the task is really corpus curation.
- Keep the output deterministic and grounded in checked-in expectations.
- Do not confuse "more snippets" with "more coverage"; repeated-source snippets must earn their place.

## Deliverable

Report:

- the commands run
- the source mix used
- promoted or rejected fixture IDs
- any follow-on playbook refinement work required
