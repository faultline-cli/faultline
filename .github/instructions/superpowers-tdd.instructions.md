---
name: test-driven-development
description: Use when implementing any feature or bugfix, before writing implementation code
applyTo: '**/*.go'
---

# Test-Driven Development (TDD)

## Overview

Write the test first. Watch it fail. Write minimal code to pass.

**Core principle:** If you didn't watch the test fail, you don't know if it tests the right thing.

**Violating the letter of the rules is violating the spirit of the rules.**

## When to Use

**Always:**
- New features
- Bug fixes
- Refactoring
- Behavior changes

**Exceptions (ask your human partner):**
- Throwaway prototypes
- Generated code
- Configuration files

Thinking "skip TDD just this once"? Stop. That's rationalization.

## The Iron Law

```
NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST
```

Write code before the test? Delete it. Start over.

**No exceptions:**
- Don't keep it as "reference"
- Don't "adapt" it while writing tests
- Don't look at it
- Delete means delete

Implement fresh from tests. Period.

## Red-Green-Refactor

### RED — Write Failing Test

Write one minimal test showing what should happen.

**Requirements:**
- One behavior
- Clear name describing the behavior
- Real code (no mocks unless unavoidable)

### Verify RED — Watch It Fail

**MANDATORY. Never skip.**

```bash
go test ./path/to/package/... -run TestName -v
```

Confirm:
- Test fails (not errors)
- Failure message is expected
- Fails because feature missing (not typos)

**Test passes?** You're testing existing behavior. Fix test.
**Test errors?** Fix error, re-run until it fails correctly.

### GREEN — Minimal Code

Write simplest code to pass the test. Don't add features, refactor other code,
or "improve" beyond the test.

### Verify GREEN — Watch It Pass

**MANDATORY.**

```bash
go test ./path/to/package/... -run TestName -v
go test ./... 
```

Confirm:
- Test passes
- Other tests still pass
- Output pristine (no errors, warnings)

**Test fails?** Fix code, not test.
**Other tests fail?** Fix now.

### REFACTOR — Clean Up

After green only:
- Remove duplication
- Improve names
- Extract helpers

Keep tests green. Don't add behavior.

### Repeat

Next failing test for next feature.

## Good Tests

| Quality | Good | Bad |
|---------|------|-----|
| **Minimal** | One thing. "and" in name? Split it. | `TestValidatesEmailAndDomainAndWhitespace` |
| **Clear** | Name describes behavior | `TestThing1` |
| **Shows intent** | Demonstrates desired API | Obscures what code should do |

## Go-Specific Patterns

```go
// Good: table-driven test with clear cases
func TestTruncateReportCell(t *testing.T) {
    cases := []struct {
        name  string
        input string
        limit int
        want  string
    }{
        {"empty string", "", 10, ""},
        {"within limit", "hello", 10, "hello"},
        {"exactly at limit", "hello", 5, "hello"},
        {"over limit adds ellipsis", "hello world", 8, "hello..."},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got := truncateReportCell(tc.input, tc.limit)
            if got != tc.want {
                t.Errorf("truncateReportCell(%q, %d) = %q, want %q", tc.input, tc.limit, got, tc.want)
            }
        })
    }
}
```

## Red Flags — STOP and Start Over

- Code before test
- Test after implementation
- Test passes immediately
- Can't explain why test failed
- Tests added "later"
- Rationalizing "just this once"

**All of these mean: Delete code. Start over with TDD.**

## Verification Checklist

Before marking work complete:

- [ ] Every new function/method has a test
- [ ] Watched each test fail before implementing
- [ ] Each test failed for expected reason (feature missing, not typo)
- [ ] Wrote minimal code to pass each test
- [ ] All tests pass (`make test`)
- [ ] Output pristine (no errors, warnings)
- [ ] Tests use real code (mocks only if unavoidable)
- [ ] Edge cases and errors covered
