---
name: systematic-debugging
description: Use when encountering any bug, test failure, or unexpected behavior, before proposing fixes
applyTo: '**'
---

# Systematic Debugging

## Overview

Random fixes waste time and create new bugs. Quick patches mask underlying issues.

**Core principle:** ALWAYS find root cause before attempting fixes. Symptom fixes are failure.

## The Iron Law

```
NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST
```

If you haven't completed Phase 1, you cannot propose fixes.

## When to Use

Use for ANY technical issue:
- Test failures
- Bugs in production
- Unexpected behavior
- Performance problems
- Build failures
- Integration issues

**Use ESPECIALLY when:**
- Under time pressure (emergencies make guessing tempting)
- "Just one quick fix" seems obvious
- You've already tried multiple fixes

## The Four Phases

You MUST complete each phase before proceeding to the next.

### Phase 1: Root Cause Investigation

**BEFORE attempting ANY fix:**

1. **Read Error Messages Carefully**
   - Don't skip past errors or warnings
   - They often contain the exact solution
   - Read stack traces completely
   - Note line numbers, file paths, error codes

2. **Reproduce Consistently**
   - Can you trigger it reliably?
   - What are the exact steps?
   - Does it happen every time?
   - If not reproducible → gather more data, don't guess

3. **Check Recent Changes**
   - What changed that could cause this?
   - `git diff`, recent commits
   - New dependencies, config changes

4. **Gather Evidence in Multi-Component Systems**
   - Add diagnostic instrumentation at component boundaries
   - Log what data enters and exits each component
   - Run once to gather evidence showing WHERE it breaks
   - THEN analyze evidence to identify failing component

5. **Trace Data Flow**
   - Where does bad value originate?
   - What called this with bad value?
   - Keep tracing up until you find the source
   - Fix at source, not at symptom

### Phase 2: Pattern Analysis

**Find the pattern before fixing:**

1. **Find Working Examples** — locate similar working code in same codebase
2. **Compare Against References** — read reference implementation COMPLETELY
3. **Identify Differences** — what's different between working and broken?
4. **Understand Dependencies** — what other components does this need?

### Phase 3: Hypothesis and Testing

**Scientific method:**

1. **Form Single Hypothesis** — state clearly: "I think X is the root cause because Y"
2. **Test Minimally** — make the SMALLEST possible change to test hypothesis
3. **Verify Before Continuing** — did it work? Yes → Phase 4. No → form NEW hypothesis
4. **When You Don't Know** — say "I don't understand X", don't pretend to know

### Phase 4: Implementation

**Fix the root cause, not the symptom:**

1. **Create Failing Test Case** — simplest possible reproduction + use TDD skill
2. **Implement Single Fix** — address the root cause, ONE change at a time
3. **Verify Fix** — test passes? No other tests broken? Issue actually resolved?
4. **If Fix Doesn't Work** — STOP. Count fixes tried. If ≥ 3: question the architecture.

## Red Flags — STOP and Follow Process

If you catch yourself thinking:
- "Quick fix for now, investigate later"
- "Just try changing X and see if it works"
- "Add multiple changes, run tests"
- "It's probably X, let me fix that"
- "I don't fully understand but this might work"
- "One more fix attempt" (when already tried 2+)

**ALL of these mean: STOP. Return to Phase 1.**

## Quick Reference

| Phase | Key Activities | Success Criteria |
|-------|---------------|------------------|
| **1. Root Cause** | Read errors, reproduce, check changes, gather evidence | Understand WHAT and WHY |
| **2. Pattern** | Find working examples, compare | Identify differences |
| **3. Hypothesis** | Form theory, test minimally | Confirmed or new hypothesis |
| **4. Implementation** | Create test, fix, verify | Bug resolved, tests pass |
