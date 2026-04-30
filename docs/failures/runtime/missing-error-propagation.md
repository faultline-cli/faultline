# Error return value ignored or silently discarded

**Playbook ID:** `missing-error-propagation`
**Category:** runtime
**Severity:** high
**Tags:** `source`, `error-handling`, `go`, `silent-failure`, `observability`

## What this failure means

A function returns an error that is discarded with `_ =` or left unchecked, meaning failures in I/O, database, or side-effecting operations would pass silently.

## Common log signals

*(This playbook uses source-code pattern matching rather than log signals.)*

## Diagnosis

A function returns an error that is discarded with `_ =` or left unchecked, meaning failures in I/O, database, or side-effecting operations would pass silently.

## Fix steps

1. Propagate the error to the caller — wrap it with context: `return fmt.Errorf("context: %w", err)`.
2. If the error genuinely cannot affect correctness, document why with a comment and log it at WARN level.
3. Use `errcheck` in CI to catch ignored error returns statically.
4. If you already run Staticcheck, treat `SA4006` as a companion signal for overwritten values, but do not rely on it as the primary ignored-error check.

## Validation

- Re-run the failing workflow step.
- Confirm the original failure signature for Error return value ignored or silently discarded is gone.

## Likely files to inspect

- `service.go`
- `handler.go`
- `client.go`
- `repository.go`
- `cmd/*.go`


## Run Faultline

```bash
faultline analyze build.log
faultline explain missing-error-propagation
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Error return value ignored or silently discarded
- Runtime: error return value ignored or silently discarded
- faultline explain missing-error-propagation
- Go error return value ignored or silently discarded


---

*Generated from [playbooks/bundled/source/missing-error-propagation.yaml](../../../playbooks/bundled/source/missing-error-propagation.yaml). Do not edit directly — run `make docs-generate`.*
