# Investigate Baseline Regression

Use this workflow when `make fixture-check` or `./bin/faultline fixtures stats --class real --check-baseline` fails after a repository change.

This is a reactive debugging workflow. Its job is to isolate the regression and produce one of three explicit outcomes:

- **Fix forward**: improve the playbook or scorer so the corpus passes again
- **Fix expectations**: update a fixture's expected playbook or guard when the change was a deliberate and correct improvement
- **Revert**: the change is wrong and must be undone — do not paper over it by weakening the baseline

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md)
- [`docs/fixture-corpus.md`](../docs/fixture-corpus.md)
- `fixtures/real/baseline.json`
- the diff of the change that triggered the failure

## Goal

Identify the exact fixture or set of fixtures that regressed, determine the cause, and restore a passing baseline without weakening the trust boundary.

## Step 1: Read The Baseline

`fixtures/real/baseline.json` records the last accepted snapshot. Key fields:

```
top_1_rate       — fraction of fixtures where the correct playbook ranks first
top_3_rate       — fraction where the correct playbook appears in top 3
unmatched_rate   — fraction with no match above threshold
false_positive_rate — fraction where an incorrect playbook ranks first
weak_match_rate  — fraction where the top-1 match score is below the strong threshold
thresholds       — the gate values this snapshot was checked against
fingerprint      — content hash of the accepted snapshot
```

Run the report in full JSON to see per-fixture detail:

```bash
make build
./bin/faultline fixtures stats --class real --json
```

Compare to `fixtures/real/baseline.json`. The gate fails when any rate crosses its threshold value.

## Step 2: Isolate The Regressed Fixture

Run each affected fixture individually to narrow the failure:

```bash
./bin/faultline analyze fixtures/real/<fixture-id>.yaml --json
```

Record for each:
- `top_match` playbook ID (or absent/null if unmatched)
- `top_match_score`
- whether the expected playbook from the fixture's promotion record is still ranking first

Cross-check the fixture's promotion expectations in its YAML header:

```bash
grep -r "expected_playbook\|strict_top_1\|disallow" fixtures/real/<fixture-id>.yaml
```

## Step 3: Classify The Regression

| Symptom | Likely cause |
|---------|-------------|
| Previously matched fixture is now unmatched | Pattern removed or narrowed too far |
| Previously top-1 fixture now ranks 2nd or 3rd | New playbook added patterns that outCompete the correct one; or scoring changed |
| Score dropped below weak-match threshold | Pattern breadth reduced; base_score lowered; or Bayesian weight changed |
| New false positive on a fixture | New pattern is too broad; missing `match.none` exclusion |
| Multiple fixtures regressed simultaneously | A shared pattern or scoring rule was changed |

## Step 4: Determine The Right Outcome

For each regressed fixture, choose exactly one:

**Fix forward — improve the playbook**
- The change was correct in principle but the playbook or patterns need adjustment to remain accurate.
- Tighten `match.none`, restore a removed `match.any` phrase, or adjust `base_score`.
- Re-run `make review` to confirm the fix does not introduce new overlaps.
- Re-run `make test` and `make fixture-check`.

**Fix expectations — update the fixture guard**
- The change is a deliberate correct improvement: a better playbook now ranks first, or the old expected playbook was wrong.
- Update the fixture's `expected_playbook` field to reflect the new correct answer.
- Do not update expectations just to silence the gate. Justify the change in a comment or commit message.
- Re-run `make fixture-check` to confirm the updated expectation passes cleanly.

**Revert**
- The regression is unexplained, the fix is not obvious, or the correct outcome is ambiguous.
- Revert the triggering change before deciding whether to re-introduce it with the correct guard.
- Do not update the baseline or expectations as a workaround for a genuinely broken change.

## Step 5: Regenerate The Baseline

After all fixtures pass:

```bash
make build
./bin/faultline fixtures stats --class real --check-baseline
```

If the gate passes, the baseline is implicitly confirmed. Do not manually edit `baseline.json`; let the `stats` command regenerate it when a new accepted snapshot is needed.

## Validation Sequence

```bash
make build
./bin/faultline fixtures stats --class real --json          # see per-fixture detail
./bin/faultline analyze fixtures/real/<fixture-id>.yaml --json   # isolate each regression
make review                                                  # if a playbook was changed
make test                                                    # full suite
make fixture-check                                           # gate must pass before stopping
```

## Guardrails

- Do not lower threshold values in `baseline.json` to make the gate pass.
- Do not update fixture `expected_playbook` fields unless the new correct answer is clearly better.
- Do not stop after the first fixture passes; confirm every affected fixture.
- If fixing one regression introduces another, resolve both before stopping.
- `make fixture-check` must pass clean. A weak match is not a pass.

## Deliverable

- the fixture ID(s) that regressed and why
- the outcome chosen for each (fix forward / fix expectations / revert) with justification
- the exact commands run to confirm the gate passes
- any playbook changes made and whether `make review` was re-run
