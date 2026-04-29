# Command failure suppressed with exit-code override

**Playbook ID:** `ignored-exit-code`
**Category:** silent_failure
**Severity:** high
**Tags:** `silent-failure`, `exit-code`, `set+e`, `bash`, `shell`, `ci`

## What this failure means

A failing command's exit code was deliberately suppressed using `|| true`,
`set +e`, or similar shell constructs.  CI continued without surfacing the
error, so the overall job status appears green even though a step failed.

## Common log signals

```text
|| true
set +e
failed but continuing
ignoring error
```

## Diagnosis

Shell scripts that use `|| true` or `set +e` allow subsequent commands to
run even after a previous command exits non-zero.  This pattern is sometimes
intentional (for optional steps) but is frequently used to silence unexpected
failures that should be investigated.

Common causes:

- A developer added `|| true` to unblock a pipeline without investigating the
  root cause.
- `set +e` was used at the top of a shared script and was never reverted.
- A CI step uses `exit 0` to force success after logging an error.

The result is a misleading "green" CI status that hides real failures from
reviewers and downstream automation.

## Fix steps

1. Locate the suppressed command and understand why it is failing:
   ```bash
   grep -n "|| true\|set +e\|exit 0" <script>
   ```
2. Remove the suppression and let the failure surface naturally.
3. If the failure is expected and intentional, document it explicitly and
   consider an `if/else` guard instead of a blanket exit-code override.
4. Add `set -e` (or `set -euo pipefail`) at the top of CI scripts to fail
   fast on unexpected errors.

## Validation

Re-run the CI job without the exit-code suppression and confirm the job
fails cleanly on the unexpected error rather than continuing silently.

## Likely files to inspect

- `.github/workflows/*.yml`
- `Makefile`
- `scripts/*.sh`


## Run Faultline

```bash
faultline analyze build.log
faultline explain ignored-exit-code
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Command failure suppressed with exit-code override
- Silent Failure: command failure suppressed with exit-code override
- failed but continuing
- GitHub Actions command failure suppressed with exit-code override
- faultline explain ignored-exit-code


---

*Generated from [playbooks/bundled/log/silent/ignored-exit-code.yaml](../../playbooks/bundled/log/silent/ignored-exit-code.yaml). Do not edit directly — run `make docs-generate`.*
