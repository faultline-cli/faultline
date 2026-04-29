# Generic coverage gate failure

**Playbook ID:** `coverage-gate-failure`
**Category:** test
**Severity:** medium
**Tags:** `test`, `coverage`, `threshold`, `quality-gate`

## What this failure means

The test step completed, but the job failed because measured coverage fell below an enforced minimum. This is a quality policy failure rather than a crash in the test runner itself.

## Common log signals

```text
coverage threshold
minimum coverage
global coverage threshold
does not meet global threshold
total coverage below
coverage requirement not met
lines below threshold
branches below threshold
```

## Diagnosis

The test step completed, but the job failed because measured coverage fell below an enforced minimum. This is a quality policy failure rather than a crash in the test runner itself.

## Fix steps

1. Find exactly which files or packages dropped below the threshold:

   ```bash
   # Go — sort by coverage percentage ascending:
   go test ./... -coverprofile=coverage.out
   go tool cover -func=coverage.out | sort -k3 -n | head -20

   # JavaScript:
   npm test -- --coverage --coverageReporters=text 2>&1 | grep -v "100 |"
   ```

2. Add focused tests for the newly introduced or changed code paths before
   considering any threshold reduction.
3. If generated or non-critical files are inflating the denominator, exclude
   them explicitly in the coverage configuration rather than hand-tuning the
   target percentage.
4. Re-run the same coverage command locally to confirm the failure is
   reproducible outside CI before writing new tests.

## Validation

- `go test ./... -cover` or `npm test -- --coverage`
- Confirm the reported total percentage meets or exceeds the gate threshold.

## Likely files to inspect

- `coverage.out`
- `package.json`
- `pyproject.toml`
- `jest.config.js`
- `pytest.ini`


## Run Faultline

```bash
faultline analyze build.log
faultline explain coverage-gate-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Generic coverage gate failure
- Test: generic coverage gate failure
- does not meet global threshold
- faultline explain coverage-gate-failure


---

*Generated from [playbooks/bundled/log/test/coverage-gate-failure.yaml](../../playbooks/bundled/log/test/coverage-gate-failure.yaml). Do not edit directly — run `make docs-generate`.*
