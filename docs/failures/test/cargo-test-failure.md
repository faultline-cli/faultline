# Cargo test suite reported one or more test failures

**Playbook ID:** `cargo-test-failure`
**Category:** test
**Severity:** high
**Tags:** `rust`, `cargo`, `test`, `fail`, `assertion`

## What this failure means

One or more Rust tests failed. The `cargo test` runner reported individual
failures with `test X ... FAILED` and the suite summary shows
`failures: N tests failed`.

## Common log signals

```text
test result: FAILED
test.*\.\.\.\s*FAILED
failures:
FAILED tests
error\[E
```

## Diagnosis

`cargo test` terminated with a non-zero exit code. Individual failing tests
are reported as:

```
test path::to::test_name ... FAILED

failures:

---- path::to::test_name stdout ----
thread 'main' panicked at 'assertion `left == right` failed', src/lib.rs:42

failures:
    path::to::test_name

test result: FAILED. 0 passed; 1 failed; 0 ignored; 0 measured; 0 filtered out
```

Common causes:

- **Assertion failure** — `assert_eq!`, `assert!`, or `panic!` did not hold;
  the implementation changed or a test expectation is stale.
- **Panic in test** — the code under test panicked (index out of bounds,
  unwrap on None, etc.).
- **Compile error in test** — the test file has a type error or missing
  dependency; `cargo test` reports `error[E...]` before the test run.
- **Thread panic** — a spawned thread panicked and the test did not catch it.

## Fix steps

1. Reproduce locally:

   ```bash
   cargo test
   cargo test -- --nocapture  # show stdout from failing tests
   ```

2. Run only the failing test:

   ```bash
   cargo test path::to::test_name
   cargo test test_name -- --nocapture
   ```

3. Read the panic message and backtrace:

   ```bash
   RUST_BACKTRACE=1 cargo test path::to::test_name -- --nocapture
   ```

4. For `assert_eq!` failures, the output shows both left and right values —
   identify which side changed.

5. For compile errors in tests, run `cargo check --tests` to see all type
   errors without executing tests.

## Validation

- `cargo test` exits zero.
- Individual test: `cargo test path::to::test_name` passes.

## Likely files to inspect

- `src/**/*.rs`
- `tests/**/*.rs`
- `Cargo.toml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain cargo-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Cargo test suite reported one or more test failures
- Test: cargo test suite reported one or more test failures
- test.*\.\.\.\s*FAILED
- faultline explain cargo-test-failure
- Rust cargo test suite reported one or more test failures


---

*Generated from [playbooks/bundled/log/test/cargo-test-failure.yaml](../../playbooks/bundled/log/test/cargo-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
