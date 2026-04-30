# Failures clustered within a single component boundary

**Playbook ID:** `topology-failure-clustered`
**Category:** runtime
**Severity:** medium
**Tags:** `topology`, `monorepo`, `cluster`, `component`, `localised`

## What this failure means

Multiple failures are concentrated within a single directory or package
boundary, suggesting the root cause is localised to one component rather than
spread across the codebase.

## Common log signals

```text
--- FAIL:
FAILURES:
multiple failures in
suite failed
tests failed in
```

## Diagnosis

Topology analysis shows that all observed failures share a common directory
prefix, indicating the regression is confined to a single component or
subsystem boundary. This is in contrast to a cross-boundary failure where
multiple ownership zones are affected.

Common root causes:
- **Cascading intra-component failures** — a single bad change to a core
  utility inside the component caused multiple dependents within it to fail.
- **Missing test fixture or test data** — the component relies on a resource
  that no longer exists or was renamed.
- **Initialisation or bootstrapping error** — the component fails at startup,
  causing every function that depends on it to also fail.
- **Race condition or shared state** — concurrent test execution is hitting
  shared mutable state inside the component.

## Fix steps

1. Identify the common cluster root:
   ```bash
   # Find the shared directory prefix from the failing test paths
   go test ./... 2>&1 | grep FAIL
   ```
2. Check the most-recent change inside that directory:
   ```bash
   git log --oneline -10 -- <cluster-dir>
   ```
3. Run only the failing component's tests in isolation to confirm the
   root cause is localised:
   ```bash
   go test -v ./<cluster-dir>/...
   ```
4. Look for changes to initialisation code, global state, or shared fixtures
   within the cluster that could cause fan-out failures.
5. Fix the root cause and ensure the full component test suite passes.

## Validation

- All tests within the clustered directory pass.
- No new failures have appeared in adjacent directories.
- CI passes end-to-end.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain topology-failure-clustered
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Failures clustered within a single component boundary
- Runtime: failures clustered within a single component boundary
- multiple failures in
- faultline explain topology-failure-clustered


---

*Generated from [playbooks/bundled/log/runtime/topology-failure-clustered.yaml](../../../playbooks/bundled/log/runtime/topology-failure-clustered.yaml). Do not edit directly — run `make docs-generate`.*
