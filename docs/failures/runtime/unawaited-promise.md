# Unawaited promise in async JavaScript or TypeScript code

**Playbook ID:** `unawaited-promise`
**Category:** runtime
**Severity:** high
**Tags:** `source`, `javascript`, `typescript`, `async`, `promise`, `unhandled-rejection`

## What this failure means

An async JavaScript or TypeScript function starts a promise-returning operation but never awaits or catches it. The rejection can escape the call site, turning a real failure into an unhandled rejection or a race that is hard to reproduce.

## Common log signals

*(This playbook uses source-code pattern matching rather than log signals.)*

## Diagnosis

An async JavaScript or TypeScript function starts a promise-returning operation but never awaits or catches it. The rejection can escape the call site, turning a real failure into an unhandled rejection or a race that is hard to reproduce.

## Fix steps

1. `await` the promise at the call site so the failure propagates through the current function.
2. If the work is intentionally detached, attach an explicit `.catch(...)` that logs or forwards the error.
3. Prefer returning the promise instead of dropping it when the caller should own the lifecycle.
4. Add a regression test for the async path so the missing await does not come back.

## Validation

- Re-run `faultline inspect .` or `faultline guard .` against the repository.
- Confirm the unawaited promise finding is resolved and the top source playbook still points at the intended file.

## Likely files to inspect

- `src/**/*.ts`
- `src/**/*.js`
- `handlers/**/*.ts`
- `handlers/**/*.js`
- `services/**/*.ts`
- `services/**/*.js`


## Run Faultline

```bash
faultline analyze build.log
faultline explain unawaited-promise
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Unawaited promise in async JavaScript or TypeScript code
- Runtime: unawaited promise in async javascript or typescript code
- faultline explain unawaited-promise


---

*Generated from [playbooks/bundled/source/unawaited-promise.yaml](../../playbooks/bundled/source/unawaited-promise.yaml). Do not edit directly — run `make docs-generate`.*
