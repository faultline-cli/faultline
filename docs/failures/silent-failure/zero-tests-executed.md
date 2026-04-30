# Test command ran but zero tests were executed

**Playbook ID:** `zero-tests-executed`
**Category:** silent_failure
**Severity:** high
**Tags:** `silent-failure`, `test`, `coverage`, `zero-tests`, `pytest`, `jest`, `go-test`

## What this failure means

A test runner command was invoked but discovered or executed zero tests.
CI marked the job as passing because the test runner exited zero, but no
tests were actually validated.

## Common log signals

```text
No tests found
No test files found
No test suites found
0 tests
0 passed
0 examples
0 passing
tests: 0
```

## Diagnosis

Most test runners exit with code 0 when they find no tests to execute.
This means a broken test discovery configuration, missing test files, or
an incorrect test filter produces a false green signal.

Common causes:

- A glob pattern or path filter no longer matches any test files (e.g.,
  after a directory reorganization).
- Test files were moved but the runner configuration was not updated.
- A new workspace or working directory is missing test files that exist
  elsewhere in the repository.
- A `--testPathPattern` or similar filter accidentally excludes all tests.
- `pytest` is invoked from a directory with no `test_*.py` files.
- `go test ./...` runs against a module with no `_test.go` files.

Affected runners and their zero-test signals:
- Jest / Vitest: "No tests found", "No test suites found"
- pytest: "collected 0 items", "no tests ran"
- Go test: "0 tests" (implied by empty output)
- Mocha: "0 passing"
- RSpec: "0 examples"

## Fix steps

1. Run the test command locally and inspect the discovery output.
2. Verify the test paths or glob patterns in the runner configuration match
   the actual file layout:
   ```bash
   find . -name "*.test.*" -o -name "test_*.py" -o -name "*_test.go"
   ```
3. Check whether the working directory or `--rootDir` option points at the
   correct source tree.
4. Remove over-restrictive `--testPathPattern` or `--ignore` filters.
5. Consider adding a CI assertion that at least N tests were executed.

## Validation

Re-run the test command and confirm the output shows at least one test
passing.  Add a minimum-test-count assertion to the CI pipeline to prevent
regression.

## Likely files to inspect

- `jest.config.*`
- `pytest.ini`
- `setup.cfg`
- `.mocharc.*`
- `Makefile`


## Run Faultline

```bash
faultline analyze build.log
faultline explain zero-tests-executed
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Test command ran but zero tests were executed
- Silent Failure: test command ran but zero tests were executed
- No test suites found
- faultline explain zero-tests-executed


---

*Generated from [playbooks/bundled/log/silent/zero-tests-executed.yaml](../../../playbooks/bundled/log/silent/zero-tests-executed.yaml). Do not edit directly — run `make docs-generate`.*
