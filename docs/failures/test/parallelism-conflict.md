# Test parallelism conflict

**Playbook ID:** `parallelism-conflict`
**Category:** test
**Severity:** medium
**Tags:** `test`, `parallel`, `concurrency`, `port`, `database`

## What this failure means

Parallel test execution caused a resource conflict: two tests tried to bind the same port, access the same fixture, or write to the same temporary file simultaneously.

## Common log signals

```text
resource is busy
test is not parallelizable
race detected
concurrently
too many connections
database locked
concurrent =
re:concurrent access
```

## Diagnosis

Parallel test execution caused a resource conflict: two tests tried to bind the same port, access the same fixture, or write to the same temporary file simultaneously.

## Fix steps

1. Reduce test parallelism with `-parallel 1` to confirm the conflict is parallelism-related.
2. Allocate a unique port, database, or temp directory per test rather than sharing a fixed resource.
3. Use test-specific prefixes or random ports in each test's setup.
4. If the suite must share a resource, serialize only the affected tests instead of disabling parallelism globally.

## Validation

- Re-run the failing workflow step.
- Confirm the original failure signature for Test parallelism conflict is gone.

## Likely files to inspect

- `docker-compose.yml`
- `.github/workflows/*.yml`
- `testdata/`
- `internal/`


## Run Faultline

```bash
faultline analyze build.log
faultline explain parallelism-conflict
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Test parallelism conflict
- Test: test parallelism conflict
- re:running.*tests concurrently
- faultline explain parallelism-conflict


---

*Generated from [playbooks/bundled/log/test/parallelism-conflict.yaml](../../../playbooks/bundled/log/test/parallelism-conflict.yaml). Do not edit directly — run `make docs-generate`.*
