# Author New Playbook

Use this workflow only after a new playbook has been explicitly justified — by [`triage-unmatched-log.md`](./triage-unmatched-log.md), [`collect-coverage-evidence.md`](./collect-coverage-evidence.md), or a confirmed gap in the accepted real-fixture corpus.

Do not use this workflow speculatively. If a nearby playbook can be refined to cover the case, do that first.

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md)
- [`docs/playbooks.md`](../docs/playbooks.md)
- [`docs/fixture-corpus.md`](../docs/fixture-corpus.md)
- the existing YAML in the nearest related playbook under `playbooks/bundled/`
- the staged fixture or raw log that justified this addition

## Goal

Add a single defensible playbook with:
- distinct, minimal match signals
- a paired positive fixture
- at least one confusable-neighbor guard
- passing `make review` and `make test` after the change

## Pre-Flight

Before writing YAML, answer these explicitly:

1. What is the exact root cause this playbook detects — not the symptom, not the error message, but the underlying mechanism?
2. What is the nearest existing playbook by ID? Run `faultline explain <id>` on it.
3. Does the proposed pattern overlap with that neighbor's `match.any`? If yes, is the root cause boundary genuinely distinct?
4. Can the neighbor's `match.none` be extended to cover this case instead?
5. What are the `domain`, `class`, and `mode` values from `docs/ontology.md`?

If you cannot answer question 1 with one sentence, do not author a playbook yet. Return to triage.
If you cannot answer question 5, read `docs/ontology.md` before continuing.

## Placement Decision

| Criterion | Location |
|-----------|----------|
| High-frequency failure across common language stacks or CI systems | `playbooks/bundled/log/<category>/` |
| Provider-specific, platform-specific, or deep operational tooling | extra pack |
| Uncertain — default | extra pack first; promote to bundled when usage justifies it |

Use the directory structure already in `playbooks/bundled/log/` as the canonical category reference. Do not create new top-level categories unless none of the existing ones fit.

## Required YAML Fields

Every shipped playbook must have all of these:

```yaml
id: <category>-<hyphenated-noun-phrase>
title: <Sentence case, one line>
category: <auth|build|ci|deploy|network|runtime|test>
severity: <critical|high|medium|low>
base_score: <float, 0.0–1.0>
tags: [<comma, separated, lowercase>]
stage_hints: [<build|test|deploy|...>]
domain: <from docs/ontology.md — e.g. dependency, runtime, database, network, auth>
class: <from docs/ontology.md — e.g. missing-executable, migration-failure, tls-validation>
mode: <concrete root cause slug — e.g. binary-not-found, postgres-enum-in-transaction>

summary: |-
  One or two sentences. This appears in ranked output.

diagnosis: |-
  ## Diagnosis
  Explain the root cause in plain language. No fix steps here.

fix: |-
  ## Fix steps
  1. Numbered, ordered, minimal.
  2. Use short code fences only when content is exact and unambiguous.

validation: |-
  ## Validation
  - What command confirms the failure is resolved.
  - Keep to one or two verifiable steps.

match:
  any:
    - <literal string or substring>
  none:
    - <exclusion to block the nearest confusable neighbor>

workflow:
  likely_files:
    - <glob or path most likely to contain the fix>
  local_repro:
    - <command sequence that reproduces the failure>
  verify:
    - <command that confirms the fix>
```

Optional fields: `why_it_matters`, `match.all`.

> **Ontology note:** `domain`, `class`, and `mode` are required for every new playbook.
> Consult `docs/ontology.md` for the full hierarchy and valid values. If the exact
> class or mode is not yet in the ontology, choose the nearest valid parent and add a
> comment in the PR describing the proposed addition.

## Pattern Authoring Rules

**match.any**
- Use the most specific stable substring visible in real logs. Prefer exact error strings over generic phrases.
- Every phrase must appear verbatim in at least one real-world example you have seen or staged.
- Do not use phrases that also appear in the nearest confusable neighbor without a `match.none` exclusion to separate them.
- Fewer, tighter phrases are better than many broad ones.

**match.none**
- Add at least one exclusion when the nearest existing playbook shares any phrase in your `match.any`.
- Exclusions must be grounded in real signals, not hypothetical ones.

**match.all**
- Use only when co-occurrence of two phrases is required to avoid false positives. Use sparingly.

## Fixture Pairing

Before running any validation, create the minimal fixture to defend this playbook:

1. If a staged fixture from ingestion exists, promote it:
   ```bash
   faultline fixtures promote <staging-id> --expected-playbook <new-id> --strict-top-1
   ```

2. If no staged fixture exists, create a minimal one under `fixtures/minimal/`:
   ```yaml
   # fixtures/minimal/<new-id>.yaml
   id: <new-id>
   log: |
     <paste the minimal log lines that trigger match.any and not match.none>
   ```
   Then confirm it matches:
   ```bash
   make build
   ./bin/faultline analyze fixtures/minimal/<new-id>.yaml --json
   ```

3. Add at least one adversarial or near-miss log line to verify the `match.none` exclusions hold:
   ```bash
   ./bin/faultline analyze <confusable-neighbor-fixture> --json
   ```
   Confirm the confusable neighbor still resolves to the correct original playbook, not the new one.

## Validation Sequence

Run in this order:

```bash
make review         # inspect pattern conflicts against the full catalog
make test           # confirm unit and fixture regressions are clean
make build
make fixture-check  # confirm accepted real-fixture corpus is still stable
```

If `make review` reports a new overlap between the new playbook and a neighbor:
- investigate whether the overlap is acceptable (stable, narrow, different root cause) or a real conflict
- if a conflict, tighten `match.any` or add `match.none` before continuing

Do not consider the playbook shipped until all four commands pass.

## Acceptance Bar

- the root cause is distinct from every existing neighbor
- `match.any` phrases are grounded in at least one real log sample
- `match.none` blocks the nearest confusable neighbor
- `domain`, `class`, and `mode` ontology fields are populated (see `docs/ontology.md`)
- a positive fixture exists and is part of the checked-in corpus or a minimal regression test
- `make review`, `make test`, and `make fixture-check` all pass
- `workflow.likely_files`, `workflow.local_repro`, and `workflow.verify` are populated with actionable content, not placeholders

## Deliverable

- the new playbook YAML path and ID
- the fixture used to defend it
- any confusable-neighbor adjustments required
- the exact validation commands run and their results
