# Jest test suite reported one or more test failures

**Playbook ID:** `jest-test-failure`
**Category:** test
**Severity:** high
**Tags:** `jest`, `node`, `javascript`, `typescript`, `react`, `test`, `fail`, `assertion`

## What this failure means

One or more Jest tests failed. The runner exited non-zero. Individual failures
are listed with `●` (bullet) markers and the suite summary shows
`Test Suites: N failed`.

## Common log signals

```text
● 
re:Test Suites:.*failed
Warning: Failed prop type
Invalid prop 
Test Suites: 
```

## Diagnosis

Jest ran and collected results, but one or more tests did not pass. The output
follows this structure:

```
FAIL src/utils/formatDate.test.js
  ● formatDate › formats a timestamp correctly

    expect(received).toEqual(expected)

    Expected: "2024-01-15"
    Received: "01/15/2024"

Test Suites: 1 failed, 3 total
Tests:       1 failed, 4 passed, 5 total
```

Common causes:

- **Assertion failure** — `expect(received).toEqual(expected)` does not pass;
  an implementation change or a stale snapshot broke the contract.
- **Snapshot mismatch** — a snapshot test sees output that differs from the
  stored `.snap` file; update with `jest --updateSnapshot` after verifying
  the change is intentional.
- **Missing DOM element** — RTL / Enzyme query (`getByRole`, `findByText`)
  cannot find an element, often because of a render error, wrong query, or
  incorrect test setup.
- **Thrown error in test body** — the test itself threw synchronously or
  rejected a promise, causing Jest to mark the test as failed.
- **Module mock mismatch** — a `jest.mock()` factory returns unexpected data
  or a spy was not properly reset between tests.
- **`● Test suite failed to run`** — the test file itself could not be
  compiled or executed—a syntax error, missing import, or top-level `throw`.

## Fix steps

1. Reproduce locally:

   ```bash
   npx jest
   npx jest --verbose
   ```

2. Run only the failing suite to isolate the failure:

   ```bash
   npx jest src/utils/formatDate.test.js
   npx jest --testNamePattern "formats a timestamp correctly"
   ```

3. Read the full `●` block in the output — it shows the file path, failing
   *describe/it* name, expected vs. received values, and the exact line.

4. For a snapshot mismatch, review the diff then update if the change is
   intentional:

   ```bash
   npx jest --updateSnapshot
   ```

5. For a `● Test suite failed to run` error, run the file through your
   bundler or TypeScript compiler directly to surface the build error:

   ```bash
   npx tsc --noEmit
   npx jest --showConfig 2>&1 | grep transform
   ```

6. For DOM/RTL failures, add `screen.debug()` before the failing assertion to
   inspect the rendered output.

7. Re-run with `--detectOpenHandles` if Jest appears to hang after failures:

   ```bash
   npx jest --detectOpenHandles
   ```

## Validation

- `npx jest` exits zero.
- `npx jest --coverage` exits zero with no threshold violations.
- The specific failing test: `npx jest --testNamePattern "<test name>"` exits zero.

## Likely files to inspect

- `**/*.test.js`
- `**/*.test.ts`
- `**/*.test.jsx`
- `**/*.test.tsx`
- `**/*.spec.js`
- `**/*.spec.ts`
- `jest.config.js`
- `jest.config.ts`
- `jest.setup.js`
- `package.json`


## Run Faultline

```bash
faultline analyze build.log
faultline explain jest-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Jest test suite reported one or more test failures
- Test: jest test suite reported one or more test failures
- Warning: Failed prop type
- faultline explain jest-test-failure
- Node.js jest test suite reported one or more test failures


---

*Generated from [playbooks/bundled/log/test/jest-test-failure.yaml](../../../playbooks/bundled/log/test/jest-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
