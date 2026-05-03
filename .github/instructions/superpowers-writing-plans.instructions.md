---
name: writing-plans
description: Use when you have a spec or requirements for a multi-step task, before touching code
applyTo: '**'
---

# Writing Plans

## Overview

Write comprehensive implementation plans assuming the engineer has zero context for
the codebase and questionable taste. Document everything they need to know:
which files to touch for each task, code, testing, docs they might need to check,
how to test it. Give them the whole plan as bite-sized tasks. DRY. YAGNI. TDD.
Frequent commits.

**Save plans to:** `docs/superpowers/plans/YYYY-MM-DD-<feature-name>.md`

## Bite-Sized Task Granularity

**Each step is one action (2-5 minutes):**
- "Write the failing test" — step
- "Run it to make sure it fails" — step
- "Implement the minimal code to make the test pass" — step
- "Run the tests and make sure they pass" — step
- "Commit" — step

## Go-Specific Task Structure

````markdown
### Task N: [Component Name]

**Files:**
- Create: `exact/path/to/file_test.go`
- Modify: `exact/path/to/existing.go`

- [ ] **Step 1: Write the failing test**

```go
func TestSpecificBehavior(t *testing.T) {
    got := functionUnderTest(input)
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./path/to/package/... -run TestSpecificBehavior -v
```
Expected: FAIL with "functionUnderTest: undefined" or incorrect value

- [ ] **Step 3: Write minimal implementation**

```go
func functionUnderTest(input Type) ReturnType {
    // minimal implementation
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./path/to/package/... -run TestSpecificBehavior -v
go test ./...  # verify no regressions
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: add specific feature"
```
````

## No Placeholders

Every step must contain the actual content an engineer needs. These are **plan
failures** — never write them:
- "TBD", "TODO", "implement later"
- "Add appropriate error handling"
- "Write tests for the above" (without actual test code)
- Steps that describe what to do without showing how

## Self-Review

After writing the complete plan:
1. **Spec coverage:** Can you point to a task that implements every requirement?
2. **Placeholder scan:** Any of the patterns from "No Placeholders"? Fix them.
3. **Type consistency:** Do method signatures in later tasks match earlier definitions?

## Repository Conventions

- Tests use `make test` to run
- Playbook changes require `make review`
- User-facing output changes require `make cli-smoke`
- Keep implementation deterministic — same input must produce same output
- No ML/LLM in product logic
