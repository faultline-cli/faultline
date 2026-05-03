---
name: verification-before-completion
description: Use when about to claim work is complete, fixed, or passing, before committing or creating PRs
applyTo: '**'
---

# Verification Before Completion

## Overview

Claiming work is complete without verification is dishonesty, not efficiency.

**Core principle:** Evidence before claims, always.

## The Iron Law

```
NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE
```

If you haven't run the verification command in this message, you cannot claim it passes.

## The Gate Function

```
BEFORE claiming any status or expressing satisfaction:

1. IDENTIFY: What command proves this claim?
2. RUN: Execute the FULL command (fresh, complete)
3. READ: Full output, check exit code, count failures
4. VERIFY: Does output confirm the claim?
   - If NO: State actual status with evidence
   - If YES: State claim WITH evidence
5. ONLY THEN: Make the claim

Skip any step = lying, not verifying
```

## Verification Commands for This Repository

```bash
# Core: all tests must pass
make test

# After playbook additions or pattern changes
make review

# After user-facing output, examples, or packaging changes
make cli-smoke

# Full release gate
make release-check VERSION=<tag>
```

## Common Failures

| Claim | Requires | Not Sufficient |
|-------|----------|----------------|
| Tests pass | `make test` output: 0 failures | Previous run, "should pass" |
| Playbooks valid | `make review` passed | Manual inspection |
| Build succeeds | Build command: exit 0 | Linter passing, logs look good |
| Bug fixed | Test original symptom: passes | Code changed, assumed fixed |
| Regression test works | Red-green cycle verified | Test passes once |

## Red Flags — STOP

- Using "should", "probably", "seems to"
- Expressing satisfaction before verification ("Great!", "Perfect!", "Done!")
- About to commit without verification
- Trusting agent success reports
- Relying on partial verification
- Thinking "just this once"

## Key Patterns

**Tests:**
```
✅ [Run make test]
[See: ok faultline/... (all packages)] "All tests pass"
❌ "Should pass now" / "Looks correct"
```

**Bug fix:**
```
✅ Write failing test → watch it fail → fix code → watch it pass → "bug fixed with evidence"
❌ "I changed the code, it should work now"
```

## The Bottom Line

**No shortcuts for verification.**

Run the command. Read the output. THEN claim the result.

This is non-negotiable.
