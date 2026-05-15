# CI command failure hidden by shell exit-code swallowing

**Playbook ID:** `ignored-shell-exit-in-ci`
**Category:** silent_failure
**Severity:** high
**Tags:** `source`, `ci`, `shell`, `silent-failure`, `exit-code`

## What this failure means

A CI script or workflow swallows a critical shell command failure with `|| true` or `set +e`, allowing the job to pass after important work failed.

## Common log signals

*(This playbook uses source-code pattern matching rather than log signals.)*

## Diagnosis

A critical CI command can fail without propagating a non-zero exit code. The workflow may report success even when tests, builds, deploys, or release commands failed.

## Fix steps

1. Remove broad `|| true` or `set +e` around critical commands.
2. If a cleanup command may fail safely, isolate it and document why it is non-blocking.
3. Use explicit `if` handling that fails the job for unexpected command failures.

## Validation

- Re-run `faultline inspect .` or `faultline guard .`.
- Run the affected script with a forced failing command and confirm the process exits non-zero.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `.gitlab-ci.yml`
- `scripts/*.sh`


## Run Faultline

```bash
faultline analyze build.log
faultline explain ignored-shell-exit-in-ci
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- CI command failure hidden by shell exit-code swallowing
- Silent Failure: ci command failure hidden by shell exit-code swallowing
- GitHub Actions ci command failure hidden by shell exit-code swallowing
- faultline explain ignored-shell-exit-in-ci


---

*Generated from [playbooks/bundled/source/ignored-shell-exit-in-ci.yaml](../../../playbooks/bundled/source/ignored-shell-exit-in-ci.yaml). Do not edit directly — run `make docs-generate`.*
