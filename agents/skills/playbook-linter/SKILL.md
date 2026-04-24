---
name: playbook-linter
description: Use this skill as the quality gate immediately before running make review on any new or modified playbook. Trigger it after drafting or editing playbook YAML to catch determinism failures, weak matchers, missing negative coverage, ontology incompleteness, and fixture quality problems. Always run it before make review. It replaces ad-hoc subjective judgment with a grounded, repeatable checklist. A FAIL verdict blocks make review — all critical issues must be resolved first.
---

# Playbook Linter

This skill is the mandatory quality gate before any playbook is considered ready for `make review`.

Use it when:

- a new playbook has been drafted (required)
- an existing playbook has been modified (required)
- a fixture regression suggests the matcher may be incorrect (diagnostic use)

Do not use it for Go or infrastructure changes that do not touch playbook YAML.

## Read First

- [`prompts/playbook-linter.md`](../../../prompts/playbook-linter.md)
- the target playbook YAML
- the paired positive and near-miss fixtures

## Checklist

Apply every criterion from the prompt. All critical items must PASS before proceeding to `make review`.

| Criterion | Critical? |
|-----------|-----------|
| Determinism — same input → same output; no ambiguous multi-playbook overlap without score separation | Yes |
| Matcher precision — every `match.any` phrase appears verbatim in at least one real log | Yes |
| False positive risk — ≥2 scenarios identified; ≥1 `match.none` guard per overlapping scenario | Yes |
| Evidence quality — log evidence is specific and verifiable | Yes |
| Diagnosis quality — identifies root cause, not just error message | Yes |
| Fix steps — concrete, ordered, tied to the diagnosis | Yes |
| Fixture realism — canonical fixture includes setup noise, context, realistic formatting | Yes |
| Negative test coverage — ≥1 near-miss or adversarial fixture that correctly fails to match | Yes |
| Ontology completeness — for new playbooks: `domain`, `class`, `mode` fields are populated | Yes (new only) |
| Improvements — non-blocking suggestions only | No |

## Running the Linter

Apply the prompt as a structured review step. Record a verdict for each criterion:

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

## Guardrails

- A FAIL verdict blocks `make review`. Do not proceed until all critical issues are resolved.
- Missing negative test coverage is always a critical issue, not a suggestion.
- A minimal one-liner fixture fails the realism check. This is a critical issue.
- Ontology incompleteness is a critical issue for new playbooks. Existing playbooks: not required but record as an improvement.
- After fixing critical issues, re-run the linter before proceeding.

## Deliverable

A structured verdict record per playbook reviewed. If FAIL, enumerate all blocking issues before stopping.
