# Test order dependency

**Playbook ID:** `order-dependency`
**Category:** test
**Severity:** medium
**Tags:** `test`, `order`, `state`, `isolation`

## What this failure means

A test failed because it depends on side effects or state created by a different test that ran earlier. When test order changes, the assumed preconditions are absent.

## Common log signals

```text
depends on previous test
test order
test assumes
leaked state
global state
left over from previous
```

## Diagnosis

A test failed because it depends on side effects or state created by a different test that ran earlier. When test order changes, the assumed preconditions are absent.

## Fix steps

1. Add proper setup and teardown to the failing test so it creates its own
   required state independently.
2. Remove shared mutable fixtures between tests; every test should own and
   clean up its own data.
3. Run tests in random order to surface all hidden ordering assumptions:

   ```bash
   go test -shuffle=on ./...                  # Go 1.20+
   pytest --randomly-seed=<N>                 # Python, requires pytest-randomly
   jest --randomize                           # Jest 29+
   ```

   Record and re-use the seed from a failing run to reproduce the exact
   order:

   ```bash
   go test -shuffle=on -v ./... 2>&1 | grep -E "^-test.shuffle"
   pytest --randomly-seed=<seed from the failing run>
   ```

4. Add cleanup in `defer` (Go), `afterEach`/`afterAll` (Jest), or `teardown`
   (Python) to guarantee state is reset regardless of test outcome.

## Validation

- Run the full suite with shuffle enabled several times and confirm zero
  failures across different orderings.
- Re-run the originally failing test in isolation and confirm it passes.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `testdata/`
- `internal/`


## Run Faultline

```bash
faultline analyze build.log
faultline explain order-dependency
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Test order dependency
- Test: test order dependency
- depends on previous test
- faultline explain order-dependency


---

*Generated from [playbooks/bundled/log/test/order-dependency.yaml](../../../playbooks/bundled/log/test/order-dependency.yaml). Do not edit directly — run `make docs-generate`.*
