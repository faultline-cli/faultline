# Flaky test failure

**Playbook ID:** `flaky-test`
**Category:** test
**Severity:** low
**Tags:** `test`, `flaky`, `intermittent`, `timing`

## What this failure means

A test failed but the failure is not consistently reproducible. Flaky tests indicate timing dependencies, global state, or resource contention that produce different outcomes on each run.

## Common log signals

```text
failed on attempt
intermittent failure
test is flaky
flaky test
flakey test
timeout waiting for
race condition detected
retry attempt
```

## Diagnosis

A test failed but the failure is not consistently reproducible. Flaky tests indicate timing dependencies, global state, or resource contention that produce different outcomes on each run.

## Fix steps

1. Retry the failing test job once to confirm the failure is intermittent.
2. Reproduce reliably by running the test in a tight loop:

   ```bash
   # Go — rerun with gotestsum for better output:
   gotestsum --rerun-fails=3 -- -run TestName -count=20 ./...

   # Python:
   pytest -k test_name --count=20  # requires pytest-repeat

   # Jest:
   jest --testNamePattern="failing test" --forceExit
   ```

3. Replace fixed sleeps with deterministic waits on observable conditions,
   event completion, or polling with a bounded timeout.
4. Remove shared global state, reused temp directories, fixed ports, and
   cross-test data dependencies by using explicit setup and teardown.
5. If the test depends on time, randomness, or external services, freeze the
   clock, record and log the random seed, or stub the dependency so the
   test produces the same result on every run.

## Validation

- Run the isolated test in a loop (at least 20 runs) and confirm zero failures.
- Re-run the full suite in CI and confirm no retry or flaky-flag annotations.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain flaky-test
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Flaky test failure
- Test: flaky test failure
- race condition detected
- faultline explain flaky-test


---

*Generated from [playbooks/bundled/log/test/flaky-test.yaml](../../../playbooks/bundled/log/test/flaky-test.yaml). Do not edit directly — run `make docs-generate`.*
