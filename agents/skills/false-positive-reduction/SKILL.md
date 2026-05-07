---
name: false-positive-reduction
description: Use this skill to run a full false-positive reduction cycle on the bundled playbook set. It extracts patterns shared across two or more playbooks, classifies which conflicts are genuine risks, fixes them by adding match.none discriminators, adds adversarial fixture logs and test cases to defend the boundary, and verifies with make review-update and make test. Trigger it when the baseline is green but you want to proactively narrow ambiguous pattern overlaps. Do not use it when a specific fixture is already failing — use playbook-refinement or baseline-regression instead.
---

# False-Positive Reduction

This skill runs a systematic cycle to reduce pattern overlap between bundled playbooks.

Use it when:

- `make test` is green and you want a proactive quality pass
- `make review-verbose` shows new or growing conflict counts
- a recent playbook addition introduced shared patterns with neighbors
- the goal is to narrow `match.none` sets so the scorer can discriminate reliably

Do not use it when:

- a specific fixture is already failing — use `playbook-refinement`
- a baseline gate is failing — use `baseline-regression`
- the goal is to add new playbooks — use `playbook-sprint`

## Read First

- [`SYSTEM.md`](../../../SYSTEM.md)
- [`docs/playbooks.md`](../../../docs/playbooks.md)
- [`internal/engine/playbooks_adversarial_test.go`](../../../internal/engine/playbooks_adversarial_test.go)

## Phase 1 — Extract Conflict Map

Run the review tool in verbose mode and extract all patterns shared across two or more playbooks:

```bash
make review-verbose 2>&1 | python3 - <<'EOF'
import sys, re, collections
current = None
pattern_playbooks = collections.defaultdict(list)
for line in sys.stdin:
    line = line.rstrip()
    m = re.match(r"^\s+pattern: (.+)$", line)
    if m:
        current = m.group(1)
    m = re.match(r"^\s+playbooks: (.+)$", line)
    if m and current:
        pbs = [p.strip() for p in m.group(1).split(",")]
        for pb in pbs:
            if pb not in pattern_playbooks[current]:
                pattern_playbooks[current].append(pb)
        current = None

multi = {p: pbs for p, pbs in pattern_playbooks.items() if len(pbs) >= 2}
for p in sorted(multi):
    print(f"  {p!r}: {multi[p]}")
EOF
```

## Phase 2 — Classify Each Conflict

For every entry in the conflict map, classify it as one of:

| Class | Description | Action |
|-------|-------------|--------|
| **KNOWN-OK** | Specialization pairs where both are correct (e.g., `missing-executable` / `node-missing-executable`), or adversarial tests already defend the boundary | Skip |
| **ALREADY FIXED** | `match.none` guards exist in both playbooks and point in opposite directions | Skip |
| **INVESTIGATE** | No `match.none` discriminator exists; a mis-rank is plausible on a realistic log | Fix |

Check the `match.none` section of each playbook pair before classifying. Look for existing guards that already route the shared signal.

Also check: is there an existing adversarial test case in `internal/engine/playbooks_adversarial_test.go` that already defends this boundary? If yes, classify as KNOWN-OK or ALREADY FIXED.

## Phase 3 — Assess Risk Per INVESTIGATE Item

For each INVESTIGATE item, read both playbooks' full `match:` sections and answer:

1. Which playbook should win on a realistic log that contains only this shared signal?
2. Do the other signals in `match.any` already make discrimination reliable in practice?
3. Would adding a signal to `match.none` of the broader playbook break any existing canonical fixture?

Check existing fixture logs in `internal/engine/testdata/fixtures/` and real corpus fixtures in `fixtures/real/` that exercise either playbook.

**Priority order** (highest risk first):
- 3-way or higher cardinality conflicts
- Pairs with no `match.none` at all in either playbook
- Pairs where the shared pattern is the only realistic triggering signal
- Pairs involving broad categories (quality gates, config errors, OOM)

## Phase 4 — Fix

For each item that warrants a fix:

1. **Prefer adding to `match.none`** over removing from `match.any`. Removing from `match.any` risks breaking canonical fixtures.
2. Add the most specific discriminating signal possible to the `match.none` of the broader/weaker playbook.
3. Never add the same pattern to both `match.any` and `match.none` of the same playbook — the review tool rejects this.
4. After each edit, verify the fixture log for the playbook being narrowed does not contain the signal you just added to `match.none`.
5. Check real corpus fixtures (`fixtures/real/*.yaml`) — if a real fixture's `normalized_log` contains the signal you're adding to `match.none`, and that fixture expects the playbook you're narrowing, do not add that signal.

**Safe discriminators** to look for:
- Platform/vendor identifiers (`.circleci`, `.github/workflows`, `CircleCI resource class`)
- Structural markers unique to one tool (`ERROR at setup of`, `conftest.py`, `listen tcp`)
- Framework-specific terms (`FactoryBot`, `ActiveRecord`, `eslint`)
- Kubernetes/Docker resource nouns (`configmap not found`, `secret not found`)

## Phase 5 — Adversarial Coverage

For each fixed pair, add:

1. A fixture log file at `internal/engine/testdata/fixtures/<specific-name>.log` that:
   - Contains the shared signal
   - Contains context signals that should route to playbook A
   - Does **not** contain signals in playbook A's `match.none`
2. A test case in `internal/engine/playbooks_adversarial_test.go` under the existing slice:
   ```go
   {
       name:      "<playbook-A> beats <playbook-B> on <context>",
       file:      "<specific-name>.log",
       wantTopID: "<playbook-A-id>",
       absentIDs: []string{"<playbook-B-id>"},
   },
   ```

Name fixture files descriptively: `<winner>-no-<loser>.log` or `<winner>-<context>.log`.

## Phase 6 — Validate

```bash
make review-update   # update pattern-conflicts baseline
make test            # must exit 0, all packages green
```

If `TestRealFixtureCorpusBaseline` fails:

1. Find which real fixture regressed: `go test ./internal/fixtures/... -v -run TestRealFixtureCorpusBaseline`
2. Read the regressed fixture's `normalized_log` and `expectation.expected_playbook`
3. Identify which `match.none` addition caused it
4. Revert that specific addition and remove the corresponding adversarial test/fixture
5. Re-run `make test`

If `TestMinimalFixtureCorpusIsStrict` fails, the signal removed from `match.any` was load-bearing — add it back and find a different discriminator.

## Guardrails

- Do not lower thresholds or update `baseline.json` to pass the gate.
- Do not add signals to `match.none` that appear in any existing canonical fixture for that playbook.
- Do not add the same signal to both `match.any` and `match.none`.
- Specialization pairs (e.g. `missing-executable` / `node-missing-executable`) are intentional — do not merge them.
- Each fix cycle must leave `make test` fully green before stopping.

## Deliverable

Report:

- how many conflict patterns were extracted and how many required investigation
- for each INVESTIGATE item: the classification decision (fixed / skipped with reason)
- which `match.none` signals were added and to which playbooks
- adversarial fixtures and test cases added
- the final `make test` result
