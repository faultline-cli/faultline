---
name: playbook-sprint
description: Use this skill to fill coverage gaps with validated playbooks. Can run without any input — when called with no target it runs coverage-gaps.md to identify the highest-priority gap, then executes the full sprint for that gap. Can also be called with a specific failure type, coverage gap name, or public failure URL to skip directly to Phase 2. Composes coverage-gaps, coverage-evidence, fixture-generation, new-playbook-authoring, and playbook-linter into a single end-to-end chain. Do not trigger it when triage has already justified a gap and you just need the YAML — use new-playbook-authoring directly in that case.
---

# Playbook Sprint

This skill composes the full gap-to-validated-playbook loop in a single session.

**Run without input:** the skill will run `coverage-gaps.md` first to find the highest-priority uncovered gap, then execute the full sprint for it automatically. No target needed.

**Run with input:** provide a failure type, gap name, or public failure URL to skip Phase 0 and start at Phase 1.

Use it when:

- you want to fill coverage gaps without specifying what to fill ("just find and fix the most important gap")
- the starting point is a described failure type, a coverage gap name, or a public failure URL
- the goal is a committed, linted, validated playbook with real and synthetic fixtures
- no prior triage decision or staged fixture exists yet

Do not use it when:

- triage has already produced a justified gap decision — use `new-playbook-authoring` directly
- the goal is only to collect and classify evidence without authoring — use `coverage-evidence`
- the goal is to refine an existing weak playbook — use `playbook-refinement`

## Read First

- [`SYSTEM.md`](../../../SYSTEM.md)
- [`docs/playbooks.md`](../../../docs/playbooks.md)
- [`docs/fixture-corpus.md`](../../../docs/fixture-corpus.md)
- [`docs/ontology.md`](../../../docs/ontology.md)
- [`prompts/coverage-gaps.md`](../../../prompts/coverage-gaps.md)
- [`prompts/generate-playbook-from-gap.md`](../../../prompts/generate-playbook-from-gap.md)

## Phases

### Phase 0 — Find The Gap (no-input path only)

**Skip this phase if a specific failure type or URL was provided.**

Run the `coverage-gaps.md` procedure:

```bash
./bin/faultline list
```

Apply Steps 1–4 of the procedure to produce a ranked list of gap records. Select the top **5** gap records (highest impact × determinism, `safe` batch preferred). Each gap record is an input to Phase 1; execute the full Phase 1–7 sprint for each gap in priority order.

If fewer than 5 safe-batch gaps remain uncovered, fill the remainder with top experimental-batch gaps and note them in the deliverable.

### Phase 1 — Orient

```bash
./bin/faultline list
faultline explain <nearest-neighbor-id>
```

- Identify the failure type in one sentence (root cause, not symptom).
- Map it to an ontology domain and class using `docs/ontology.md`.
- Identify the nearest existing playbook by ID.
- Check `fixtures/real/baseline.json` — is there already real-corpus coverage?

**Stop here if** the failure maps cleanly to an existing playbook. Go to `playbook-refinement` instead.

### Phase 2 — Collect Real Evidence

Attempt to find at least one real public log for this failure class before generating synthetically.

```bash
faultline fixtures ingest --adapter <adapter> --url <public-url>
./bin/faultline analyze <staged-file> --json
faultline fixtures review
```

- Use the five-tier classification from `collect-coverage-evidence.md` on each sample.
- If a **covered** or **weakly covered** sample exists, that playbook is the refinement target — switch to `playbook-refinement`.
- If a **gap — known category** or **gap — new category** is confirmed, continue.
- If no public source is found or available, skip to Phase 3 with a synthetic-only path noted.

Collect metadata per the record schema in `prompts/collect-coverage-evidence.md`.

### Phase 3 — Triage

Apply the `triage-unmatched-log.md` decision criteria to each staged gap candidate:

- Reduce to stable evidence lines.
- Confirm the root cause boundary is distinct (not just different wording).
- Confirm the fix path is concrete enough to encode.
- Confirm there is a place in the fixture corpus.

**Stop here if** the case is noise, a duplicate, or the fix path is unclear. Record the rejection with justification.

### Phase 4 — Generate Synthetic Fixtures

Use the `fixture-generation` skill to produce:

| Variant | File |
|---------|------|
| canonical | `fixtures/minimal/<id>.yaml` |
| noisy | `fixtures/minimal/<id>-noisy.yaml` |
| near-miss | `fixtures/minimal/<id>-near-miss.yaml` |

If a real fixture was successfully staged in Phase 2, use its log output as the canonical anchor. Generate noisy and near-miss synthetically.

Verify each with:
```bash
./bin/faultline analyze fixtures/minimal/<id>.yaml --json
./bin/faultline analyze fixtures/minimal/<id>-noisy.yaml --json
./bin/faultline analyze fixtures/minimal/<id>-near-miss.yaml --json
```

Do not proceed to Phase 5 until canonical and noisy score top-1 at or above `base_score`, and near-miss withholds the target.

### Phase 5 — Author The Playbook

Use the `new-playbook-authoring` skill to write the YAML.

Required fields: `id`, `title`, `category`, `severity`, `base_score`, `tags`, `stage_hints`,
`summary`, `diagnosis`, `fix`, `validation`, `match.any`, `match.none`,
`workflow.likely_files`, `workflow.local_repro`, `workflow.verify`

Required ontology fields: `domain`, `class`, `mode`

Ground every `match.any` phrase in at least one log line from Phase 2 or Phase 4.
Ground every `match.none` exclusion in the near-miss fixture from Phase 4.

### Phase 6 — Lint

Run the `playbook-linter` skill against the authored YAML and all three fixtures.

All 9 criteria must PASS, including:
- Criterion 3: ≥2 false-positive scenarios identified and covered
- Criterion 8: near-miss correctly withholds the match
- Criterion 9: ontology fields populated (`domain`, `class`, `mode`)

**A FAIL verdict blocks Phase 7.** Fix all critical issues and re-run the linter.

### Phase 7 — Validate

```bash
make review
make test
make build
make fixture-check
```

If `make review` reports a conflict with an existing playbook pattern, return to Phase 5 and tighten.

## Guardrails

- Do not author a playbook without completing the triage decision in Phase 3.
- Do not skip real evidence collection — synthetic-only is acceptable only when no public source is available, and must be noted in the deliverable.
- Do not skip the `playbook-linter` gate before `make review`.
- Do not stop after one passing test. Run all four validation commands.
- Do not promote staging fixtures to `fixtures/real/` during this workflow; that is the ingestion pipeline's job.

## Deliverable

Report:

- Phase 0 (if run): gap records produced, gap selected and why
- Phase 2: evidence table (or "no real source found — synthetic only")
- Phase 3: triage decision (keep / reject with justification)
- Phase 4: three fixture paths and scores
- Phase 5: playbook path and ID with ontology fields
- Phase 6: `playbook-linter` verdict (PASS / FAIL, issues resolved)
- Phase 7: validation commands run and results
- nearest confusable neighbor and boundary maintenance
