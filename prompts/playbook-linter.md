# Playbook Linter

Use this procedure to validate a new or modified playbook before running `make review`.

Called from: `agents/skills/playbook-linter`

Apply each criterion below. Record a PASS or FAIL verdict per criterion. Any critical FAIL blocks `make review` — fix all critical issues before proceeding.

---

## Inputs Required Before Starting

- the playbook YAML under review
- the paired positive fixture (canonical or noisy)
- at least one near-miss or adversarial fixture
- output of `faultline explain <nearest-neighbor-id>` for the confusable neighbor

---

## Criterion 1: Determinism

Does the playbook produce the same result for the same input every time?

- Are there ambiguous or overlapping match conditions?
- Could two playbooks match the same log without clear score separation?

**Critical FAIL if:**
- multiple playbooks match the same log without a `match.none` guard or a meaningful score gap
- matcher relies on a vague or ordering-sensitive pattern

---

## Criterion 2: Matcher Precision

Does every phrase in `match.any` target a specific failure mode?

- Does each phrase appear verbatim in at least one real log sample?
- Is each phrase tight enough to reject unrelated errors?

**Critical FAIL if:**
- any phrase is generic (e.g. matches "error", "failed", "exception" alone)
- a phrase only appears in the near-miss and not in a positive fixture
- the pattern would match the nearest confusable neighbor without a `match.none` exclusion

---

## Criterion 3: False Positive Risk

Identify at least two plausible false-positive scenarios — logs that contain similar phrases but represent a different root cause.

**Critical FAIL if:**
- fewer than two false-positive scenarios are identified
- any identified scenario has no corresponding `match.none` exclusion or negative fixture

---

## Criterion 4: Evidence Quality

Does the playbook extract specific, verifiable log evidence?

**Critical FAIL if:**
- `match.any` phrases are vague or generic
- no concrete log line can be cited as the primary signal

---

## Criterion 5: Diagnosis Quality

Is the diagnosis specific and correct — not a restatement of the error message?

**Critical FAIL if:**
- diagnosis only paraphrases the log
- diagnosis does not name the root cause class (e.g. lockfile drift, interpreter mismatch)

---

## Criterion 6: Fix Steps

Are fix steps actionable, ordered, and tied to the diagnosis?

**Critical FAIL if:**
- steps are vague ("check config", "verify setup")
- no next step is specified after the immediate fix
- fix steps do not follow from the diagnosis

---

## Criterion 7: Fixture Realism

Does the canonical fixture resemble real CI output?

**Critical FAIL if:**
- fixture is fewer than ~10 meaningful lines
- no setup phase (checkout, install) precedes the failure
- error text does not match what the real tool would emit

---

## Criterion 8: Negative Test Coverage

Does at least one near-miss or adversarial fixture exist that correctly fails to match?

**Critical FAIL if:**
- only positive fixtures exist
- the near-miss is trivially different (e.g. different capitalization of the same phrase)

---

## Criterion 9: Ontology Completeness (new playbooks only)

For any newly authored playbook, are these three fields populated?

```yaml
domain: <domain>
class: <class>
mode: <mode>
```

See `docs/ontology.md` for valid values.

**Critical FAIL if (new playbook):** any of the three fields is absent or left as a placeholder.

**Not required for existing playbooks.** Record as an improvement suggestion if absent.

---

## Output Format

Return a structured verdict:

```
Playbook: <id>
Verdict: PASS | FAIL

Critical issues (must fix before make review):
- <issue>

Improvements (non-blocking):
- <suggestion>

False positive scenarios identified:
- <scenario 1>
- <scenario 2>

Ontology fields: complete | incomplete | not applicable (existing playbook)

Confidence score: 0–100
```

---

## Rules

- If unsure → FAIL
- Broad matcher → FAIL
- No negative test → FAIL
- Weak fixture → FAIL
- Ontology absent on new playbook → FAIL

A playbook should match exactly one real failure mode. If it feels like it might match several, it FAILS.
