# Test suite or individual test timed out

**Playbook ID:** `test-timeout`
**Category:** test
**Severity:** high
**Tags:** `test`, `timeout`, `hang`, `slow`

## What this failure means

A test (or the entire test suite) did not complete within the configured timeout period. The test runner killed the pending test and marked it as failed.

## Common log signals

```text
Test timed out
exceeded timeout
Timeout - Async callback was not invoked within
test timeout of
jest timeout
go test -timeout
test exceeded
async test timed out
```

## Diagnosis

A test (or the entire test suite) did not complete within the configured timeout period. The test runner killed the pending test and marked it as failed.

## Fix steps

1. Identify the test name in the timeout message and run it in isolation with
   verbose output to see where it stalled.
2. Check for missing `await`, unresolved promises, or callbacks that are never
   called — these prevent async runners from completing the test.
3. For Go: get a goroutine dump to see which goroutine is blocked:

   ```bash
   GOTRACEBACK=all go test -timeout 60s -run TestName ./...
   # or send SIGQUIT to the running test process: kill -QUIT <pid>
   ```

4. For Jest/Node: run with `--detectOpenHandles` to find resource leaks that
   prevent clean exit:

   ```bash
   jest --testNamePattern="failing test" --detectOpenHandles
   ```

5. Verify that any external services or mocks the test uses are available and
   responding. Add a health-check before the test rather than letting it hang.
6. Increase the timeout as a diagnostic step only — if it then passes, the
   operation is genuinely slow and needs optimisation or mocking.
7. Replace real network or DB calls in unit tests with mocks to ensure the
   test cannot hang on external I/O.

## Validation

- Run the specific test in isolation and confirm it completes within the normal
  time limit.
- Re-run the full suite and confirm no timeout errors remain.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `testdata/`
- `internal/`


## Run Faultline

```bash
faultline analyze build.log
faultline explain test-timeout
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Test suite or individual test timed out
- Test: test suite or individual test timed out
- Timeout - Async callback was not invoked within
- faultline explain test-timeout


---

*Generated from [playbooks/bundled/log/test/test-timeout.yaml](../../../playbooks/bundled/log/test/test-timeout.yaml). Do not edit directly — run `make docs-generate`.*
