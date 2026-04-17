---
name: baseline-regression
description: Use this skill when make fixture-check or faultline fixtures stats --check-baseline fails after a repository change. Trigger it for investigating which fixture regressed, why, and how to restore a passing baseline without weakening the trust boundary. Covers per-fixture isolation, regression classification (pattern narrowed, outcompeted, false positive, shared rule change), and the three explicit outcomes (fix forward, fix expectations, revert). Do not trigger it for general playbook authoring, new fixture promotion, or proactive corpus improvement — those use other skills.
---

# Baseline Regression Investigation

This skill is for diagnosing and resolving a failing `make fixture-check` gate after a repository change.

Use it when:

- `make fixture-check` exits non-zero
- `./bin/faultline fixtures stats --class real --check-baseline` reports a rate violation
- a CI run fails at the baseline gate step

Do not use it for:

- proactive corpus improvement (use `coverage-evidence`)
- authoring new playbooks (use `new-playbook-authoring`)
- general playbook refinement not triggered by a regression (use `playbook-refinement`)

## Read First

- [`SYSTEM.md`](../../../SYSTEM.md)
- [`docs/fixture-corpus.md`](../../../docs/fixture-corpus.md)
- [`prompts/investigate-baseline-regression.md`](../../../prompts/investigate-baseline-regression.md)
- `fixtures/real/baseline.json`
- the diff of the change that caused the failure

## Workflow

1. Run `make build && ./bin/faultline fixtures stats --class real --json` to get per-fixture detail.
2. Compare against `fixtures/real/baseline.json` to identify which rate crossed its threshold.
3. Isolate the regressed fixture(s):
   ```bash
   ./bin/faultline analyze fixtures/real/<fixture-id>.yaml --json
   ```
4. Classify the regression using the symptom table in the prompt (pattern removed, outcompeted, false positive, shared rule change).
5. Choose exactly one outcome per regressed fixture:
   - **Fix forward** — playbook or pattern needs adjustment to remain correct
   - **Fix expectations** — the change was a deliberate improvement; update `expected_playbook`
   - **Revert** — the regression is unexplained or ambiguous; undo the change before deciding
6. Apply the fix. If a playbook was changed, run `make review`.
7. Confirm with `make test` and `make fixture-check`.

## Guardrails

- Do not lower threshold values in `baseline.json` to pass the gate.
- Do not update fixture expectations unless the new answer is clearly and demonstrably correct.
- Do not stop after the first fixture passes; verify all affected fixtures.
- A weak match is not a pass. The gate must pass fully clean.
- Revert before papering over an unexplained regression.

## Deliverable

Report:

- the regressed fixture ID(s) and root cause classification
- the outcome chosen for each, with justification
- any playbook changes and whether `make review` was re-run
- the exact commands run confirming `make fixture-check` passes
