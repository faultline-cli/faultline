# Curate Fixture Corpus

Use this workflow when bringing new public evidence into the repository.

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md)
- [`docs/fixture-corpus.md`](../docs/fixture-corpus.md)
- [`fixtures/staging/README.md`](../fixtures/staging/README.md)

## Goal

Move a candidate case through deterministic ingestion, review, promotion, and regression validation.

## Required Approach

1. Ingest one or more public URLs into staging with `faultline fixtures ingest`.
2. Sanitize anything sensitive before promotion.
3. Run `faultline fixtures review` and classify each item:
   - duplicate
   - near-duplicate with new signal
   - candidate for promotion
4. Promote only after setting explicit expectations with `faultline fixtures promote --expected-playbook ...`.
5. Rebuild and run `./bin/faultline fixtures stats --class real --check-baseline`.
6. If the promoted fixture exposed a weak or confusable diagnosis, follow up with the playbook refinement workflow before considering the task complete.

## Promotion Rules

- Do not promote raw staging files just because ingestion succeeded.
- Use `--strict-top-1` only when the boundary should be exact.
- Use `--disallow` when you already know the dangerous nearby false positive.
- Keep staging fixtures only when you are intentionally iterating on them.

## Deliverable

- promoted real fixture or explicit rejection
- updated expectations that are strong enough to protect the corpus over time
