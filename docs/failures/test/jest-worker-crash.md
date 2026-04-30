# Jest worker process crashed unexpectedly

**Playbook ID:** `jest-worker-crash`
**Category:** test
**Severity:** high
**Tags:** `jest`, `javascript`, `worker`, `crash`, `node`, `signal`

## What this failure means

A Jest worker process crashed mid-test run. Jest spawns multiple worker processes to run tests in parallel, and one of them exited unexpectedly.

## Common log signals

```text
A worker process has quit unexpectedly
Jest worker exited
Force exiting Jest
jest-worker
Worker failed after
worker crashed
The worker has exited.
jest-circus
```

## Diagnosis

A Jest worker process crashed mid-test run. Jest spawns multiple worker processes to run tests in parallel, and one of them exited unexpectedly.

## Fix steps

1. Run with a single worker to isolate the failing test: `jest --runInBand`.
2. Identify which suite crashes first by running Jest in-band with verbose output.
3. If a native addon is involved, rebuild it for the current Node version.
4. If the crash is memory-related, increase Node memory with `NODE_OPTIONS=--max-old-space-size=4096`.
5. Search the test code for `process.exit()` or `process.abort()` calls.

## Validation

- `jest --runInBand` completes without a worker crash.
- Re-run the CI test job and confirm the worker stays alive through the suite.

## Likely files to inspect

- `jest.config.js`
- `jest.config.ts`
- `package.json`


## Run Faultline

```bash
faultline analyze build.log
faultline explain jest-worker-crash
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Jest worker process crashed unexpectedly
- Test: jest worker process crashed unexpectedly
- A worker process has quit unexpectedly
- faultline explain jest-worker-crash
- Node.js jest worker process crashed unexpectedly


---

*Generated from [playbooks/bundled/log/test/jest-worker-crash.yaml](../../../playbooks/bundled/log/test/jest-worker-crash.yaml). Do not edit directly — run `make docs-generate`.*
