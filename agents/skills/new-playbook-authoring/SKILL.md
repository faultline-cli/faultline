---
name: new-playbook-authoring
description: Use this skill when the task is to author a new Faultline playbook after a gap has already been explicitly justified. Trigger it only when triage-unmatched-log, collect-coverage-evidence, or a confirmed fixture regression has established that a new playbook is warranted and the nearest existing playbook cannot cover the case. Covers placement decision (bundled vs extra pack), required YAML field authoring, pattern discipline for match.any and match.none, fixture pairing, and the full make review / make test / make fixture-check validation sequence. Do not trigger it speculatively or as a first step — refinement should always be ruled out first.
---

# New Playbook Authoring

This skill is for authoring a new Faultline playbook from an explicitly justified gap.

Use it when:

- `triage-unmatched-log.md` has determined a new playbook is warranted
- `collect-coverage-evidence.md` has classified a sample as "gap — new category" and that classification has survived triage
- a confirmed real-fixture regression reveals a root cause with no existing playbook analog

Do not use it for:

- speculative catalog expansion
- cases where the nearest playbook could be tightened instead (use `playbook-refinement`)
- general YAML or Go coding tasks

## Read First

- [`SYSTEM.md`](../../../SYSTEM.md)
- [`docs/playbooks.md`](../../../docs/playbooks.md)
- [`docs/fixture-corpus.md`](../../../docs/fixture-corpus.md)
- [`prompts/author-new-playbook.md`](../../../prompts/author-new-playbook.md)
- The nearest related playbook under `playbooks/bundled/`

## Workflow

1. Answer the pre-flight questions in the prompt before writing any YAML:
   - What is the root cause (one sentence)?
   - What is the nearest existing playbook by ID?
   - Does the proposed pattern overlap with that neighbor? Can `match.none` resolve it instead?
2. Decide placement: bundled `playbooks/bundled/log/<category>/` for high-frequency cross-stack failures; extra pack for provider-specific or deep operational rules.
3. Author the YAML with all required fields:
   `id`, `title`, `category`, `severity`, `base_score`, `tags`, `stage_hints`,
   `summary`, `diagnosis`, `fix`, `validation`, `match.any`, `match.none`,
   `workflow.likely_files`, `workflow.local_repro`, `workflow.verify`
4. Pair a positive fixture before validating:
   - promote from staging if one exists: `faultline fixtures promote <id> --expected-playbook <new-id> --strict-top-1`
   - otherwise create a minimal fixture under `fixtures/minimal/`
5. Run an adversarial check against the nearest confusable neighbor to confirm exclusions hold.
6. Run the full validation sequence in order:
   ```bash
   make review
   make test
   make build
   make fixture-check
   ```
7. If `make review` reports a conflict, tighten patterns before continuing.

## Guardrails

- Do not author without a justified gap decision from a prior workflow.
- Prefer refinement over addition; new playbooks must earn their place.
- Every `match.any` phrase must be grounded in at least one real log sample.
- Every `match.none` exclusion must be grounded in at least one real confusable case.
- Do not ship with placeholder `workflow` fields.
- Do not stop at a passing compile or a single fixture match; run all four validation commands.

## Deliverable

Report:

- the new playbook YAML path and ID
- the fixture used to defend it
- the nearest confusable neighbor and how the boundary is maintained
- the exact validation commands run and their results
