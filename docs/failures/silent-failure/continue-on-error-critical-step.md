# Critical CI step allowed to fail

**Playbook ID:** `continue-on-error-critical-step`
**Category:** silent_failure
**Severity:** high
**Tags:** `source`, `ci`, `github-actions`, `silent-failure`, `workflow`

## What this failure means

A CI workflow marks a critical build, test, deploy, artifact, or security step with `continue-on-error: true`, so the workflow can pass while the important work failed.

## Common log signals

*(This playbook uses source-code pattern matching rather than log signals.)*

## Diagnosis

A critical CI step is allowed to fail. This can hide broken tests, failed builds, missing artifacts, or skipped security checks behind a green workflow result.

## Fix steps

1. Remove `continue-on-error: true` from critical build, test, deploy, artifact, and security steps.
2. If the step is intentionally experimental, rename it clearly and keep it away from release or required-check jobs.
3. Split optional probes into a separate non-blocking job so required jobs still fail on real quality gates.

## Validation

- Re-run `faultline inspect .` or `faultline guard .`.
- Re-run the workflow and confirm the critical step fails the job when its command fails.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain continue-on-error-critical-step
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Critical CI step allowed to fail
- Silent Failure: critical ci step allowed to fail
- GitHub Actions critical ci step allowed to fail
- faultline explain continue-on-error-critical-step


---

*Generated from [playbooks/bundled/source/continue-on-error-critical-step.yaml](../../../playbooks/bundled/source/continue-on-error-critical-step.yaml). Do not edit directly — run `make docs-generate`.*
