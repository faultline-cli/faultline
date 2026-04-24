# Coverage Gap Analysis

Use this procedure to identify high-value gaps in CI failure coverage against the current playbook catalog.

Called from: `agents/skills/coverage-evidence` and `agents/skills/playbook-sprint`

Output feeds directly into `prompts/generate-playbook-from-gap.md` for each confirmed gap.

---

## Inputs Required Before Starting

- output of `./bin/faultline list` (current catalog)
- `fixtures/real/baseline.json` (accepted regression coverage by class)
- optionally: a specific domain, CI system, or language ecosystem to audit

---

## Step 1: Map Current Coverage

For each failure domain in `docs/ontology.md`, identify:

- which classes are covered (at least one playbook exists)
- which classes are weakly covered (only one playbook, shallow matcher, or no real-corpus fixture)
- which classes are absent (no playbook exists)

Use `faultline explain <id>` on any playbook whose coverage depth is unclear.

Build a table:

```
Domain          | Class                    | Depth      | Notes
----------------|--------------------------|------------|-------
dependency      | lockfile-drift           | deep       | multiple playbooks, real fixtures
dependency      | cache-poisoning          | shallow    | one playbook, minimal fixture only
runtime         | missing-executable       | deep       | ...
runtime         | interpreter-mismatch     | medium     | ...
...
```

---

## Step 2: Identify Gap Types

Look for three types of gap:

### Missing classes
Failure classes from the ontology with no playbook at all. These are the highest-value additions.

### Weak coverage
Classes with a single playbook that:
- has a broad or generic `match.any` pattern
- has no `match.none` exclusions
- has only a minimal toy fixture and no real-corpus fixture

### Overlap or ambiguity
Cases where:
- two playbooks share `match.any` phrases without a `match.none` separator
- a playbook's `base_score` is low enough that confusable neighbors often rank above it
- `make review` reports a conflict

---

## Step 3: Prioritise

Rank gaps using three factors:

**Impact** — how often does this failure occur in real CI pipelines? How long does it take engineers to diagnose without tooling?

**Determinism** — does the failure emit a stable, specific log signal that can be matched precisely without false positives?

**Boundary clarity** — is the root cause distinct enough from existing playbooks that a new one would not introduce ambiguity?

Reject gaps where any of these are low:
- low-frequency edge cases with no real public examples
- failures whose log output is too variable or noisy to match deterministically
- cases that would overlap an existing playbook without a clean boundary

---

## Step 4: Output Gap Records

For each confirmed high-priority gap, produce a gap record:

```
Gap ID:              <domain>/<class>/<mode-slug>
Title:               <short descriptive label>
Domain:              <from ontology>
Class:               <from ontology>
Mode:                <concrete root cause slug>
Impact:              high | medium | low
Determinism:         high | medium | low
Nearest neighbor ID: <existing playbook most likely to confuse>
Root cause:          <one sentence>
Key log signal:      <verbatim phrase expected in the log>
Near-miss risk:      <scenario that must NOT match>
Expansion batch:     safe | experimental
Next action:         playbook-sprint | playbook-refinement
```

Only produce records for gaps that pass the prioritisation filter in Step 3.

---

## Step 5: Suggest Structural Fixes

If the audit reveals structural problems in existing playbooks, list separately:

- playbooks that should be split (too broad, covering multiple root causes)
- playbooks that should be merged (overlapping patterns, same root cause family)
- playbooks needing `match.none` additions to resolve ranking instability

These are not new playbooks — they are inputs to `playbook-refinement`.

---

## Output Format

Return two sections:

### Coverage Summary

A filled-in version of the coverage table from Step 1, with depth ratings for each domain/class.

### Gap Records

One record per confirmed gap, ordered by priority (impact × determinism). Include the batch assignment (safe / experimental) and next action.

---

## Rules

- Do not suggest gaps that cannot be matched deterministically.
- Do not suggest gaps that duplicate existing playbooks with only slight wording differences.
- Do not list more than 10 gap records per run; prioritise ruthlessly.
- Every gap record must have a named nearest neighbor — if you cannot identify one, the gap may be insufficiently scoped.
