# Go test suite reported one or more test failures

**Playbook ID:** `go-test-failure`
**Category:** test
**Severity:** high
**Tags:** `go`, `golang`, `test`, `fail`, `assertion`

## What this failure means

One or more Go test functions failed or panicked. The `go test` runner
reported the failure with `--- FAIL:` for individual tests and `FAIL` for
the package summary.

## Common log signals

```text
--- FAIL:
FAIL: Test
FAIL github.com/
Test (
TestIntegration
TestBootstrap
#coordinator
#datastore
```

## Diagnosis

The Go test runner (`go test`) terminated with a non-zero exit code. Individual
failing test functions are reported as:

```
--- FAIL: TestSomething (0.01s)
    foo_test.go:42: expected 5 but got 7
FAIL    github.com/example/repo/pkg    0.042s
```

Common causes:

- **Assertion failure** — a test assertion (`t.Fatal`, `t.Error`, `require`,
  `assert`) did not hold; the application behaviour changed.
- **Panic in test** — the code under test or the test itself panicked; Go
  reports this as a test failure with a goroutine stack trace.
- **Test timeout** — the test exceeded the context deadline or the suite
  timeout (`go test -timeout`); the runner forcibly terminates and reports
  `panic: test timed out after Xs`.
- **Race condition** — the race detector (`-race`) flagged unsynchronised
  memory access.
- **Data-dependent failure** — the test depends on external state (database,
  network, filesystem) that is absent or different in CI.
- **Build failure inside test** — the test package failed to compile; the
  output includes a `build failed` summary.

## Fix steps

1. Reproduce the failing tests locally:

   ```bash
   go test ./...
   go test -v ./path/to/failing/package
   ```

2. For a specific failing test:

   ```bash
   go test -run TestFunctionName ./path/to/package
   ```

3. Read the full `--- FAIL:` block — it shows the file, line number, and the
   assertion that failed.
4. For panics, read the goroutine stack trace to find the call site.
5. For timeouts, increase the timeout or investigate why the test hangs:

   ```bash
   go test -timeout 120s ./...
   go test -v -timeout 60s ./path/to/package 2>&1 | grep -A5 "timed out"
   ```

6. For a build failure, run `go build ./...` to see the compiler errors in
   isolation, then fix the compilation errors before re-running tests.

## Validation

- `go test ./...` exits zero.
- `go test -race ./...` exits zero for race-sensitive packages.
- Individual test: `go test -run TestFunctionName ./pkg` exits zero.

## Likely files to inspect

- `**/*_test.go`
- `**/*.go`
- `go.mod`
- `Makefile`


## Run Faultline

```bash
faultline analyze build.log
faultline explain go-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Go test suite reported one or more test failures
- Test: go test suite reported one or more test failures
- FAIL github.com/
- faultline explain go-test-failure
- Go go test suite reported one or more test failures


---

*Generated from [playbooks/bundled/log/test/go-test-failure.yaml](../../../playbooks/bundled/log/test/go-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
