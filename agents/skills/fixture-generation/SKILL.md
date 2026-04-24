---
name: fixture-generation
description: Use this skill when a playbook needs realistic paired fixtures before validation. Trigger it during new-playbook-authoring when no staged fixture exists, or during playbook-refinement when existing fixtures need new variants (noisy, near-miss, adversarial). The playbook YAML must be drafted before calling this skill — fixtures defend specific patterns, not hypotheses. Do not trigger it speculatively or as a standalone corpus-enrichment activity.
---

# Fixture Generation

This skill is for generating high-fidelity CI log fixtures that defend a specific playbook.

Use it when:

- authoring a new playbook that has no staged or real-corpus fixture
- adding a noisy variant to increase confidence under real-world log conditions
- generating a near-miss to prove a `match.none` exclusion holds

Do not use it for:

- generating fixtures before the playbook YAML is drafted
- replacing ingested real-world evidence at the `fixtures/real/` level

## Read First

- [`prompts/fixture-generation.md`](../../../prompts/fixture-generation.md)
- the target playbook YAML (`match.any`, `match.none`, `base_score`)
- `faultline explain <nearest-neighbor-id>` output (required for near-miss design)

## Workflow

1. Collect the inputs before generating anything:
   - exact `match.any` phrases from the target playbook
   - exact `match.none` exclusions (if any)
   - nearest confusable neighbor ID and its `match.any` phrases
2. Generate the three required variants per the prompt:
   - **canonical** — clear failure with the full error and command context visible
   - **noisy** — same failure buried in unrelated multi-step output
   - **near-miss** — similar-looking failure that must NOT trigger this playbook
3. Place fixtures under `fixtures/minimal/<playbook-id>.yaml`, `fixtures/minimal/<playbook-id>-noisy.yaml`, and `fixtures/minimal/<playbook-id>-near-miss.yaml`.
4. Verify each fixture produces the expected result:
   ```bash
   ./bin/faultline analyze fixtures/minimal/<playbook-id>.yaml --json
   ./bin/faultline analyze fixtures/minimal/<playbook-id>-noisy.yaml --json
   ./bin/faultline analyze fixtures/minimal/<playbook-id>-near-miss.yaml --json
   ```
5. Confirm canonical and noisy rank the target playbook top-1 with score ≥ `base_score`.
6. Confirm near-miss does NOT rank the target playbook top-1.

## Realism Rules

- Every log must start with setup noise: checkout step, environment setup, tool version output.
- At least one step before the failure must complete successfully.
- The error text must appear verbatim as it would from the real tool (npm, gradle, pytest, docker, etc.).
- Canonical and noisy fixtures must include post-failure output (cleanup steps, CI summary lines).
- Near-miss must use the same tool and command as the positive fixture but with a different root cause that the `match.none` exclusion catches.
- Do not write one-liner logs for canonical or noisy variants; they fail the quality bar.

## Guardrails

- Do not generate before the playbook YAML is drafted.
- Do not place generated fixtures in `fixtures/real/`; that corpus is reserved for ingested evidence.
- Do not invent error formats. Every phrase must be verifiable against real tool output.
- Noisy and near-miss fixtures count only when analysis results are stable and deterministic.

## Deliverable

Report:

- the three fixture paths created
- the `faultline analyze` command and top-1 score for each
- confirmation that the near-miss correctly withholds the top-1 match
