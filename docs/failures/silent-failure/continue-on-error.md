# CI step configured to continue on error

**Playbook ID:** `continue-on-error`
**Category:** silent_failure
**Severity:** high
**Tags:** `silent-failure`, `continue-on-error`, `allow_failure`, `github-actions`, `gitlab-ci`

## What this failure means

A CI step uses `continue-on-error: true` or `allow_failure: true`, which
causes the pipeline to proceed even when the step exits non-zero.  This can
mask real failures and produce a misleading success status.

## Common log signals

```text
continue-on-error: true
continueOnError: true
allow_failure: true
```

## Diagnosis

CI platforms provide a "continue on error" flag that suppresses a step's
failure so that downstream steps and the overall job status are not affected.
While useful for truly optional steps, this flag is frequently misused to
paper over real failures.

Patterns detected:
- GitHub Actions: `continue-on-error: true`
- Azure DevOps: `continueOnError: true`
- GitLab CI: `allow_failure: true`

When this flag is present on a step that is actually failing, the CI
dashboard shows green but the log contains failure output.  Downstream
consumers (coverage tools, deployment gates, review bots) may act on a
false "passing" signal.

## Fix steps

1. Identify the step that uses `continue-on-error: true` (or equivalent).
2. Determine whether the step is genuinely optional or whether it should be
   required to pass.
3. For genuinely optional steps: document why the flag is needed and add a
   comment explaining what failure modes are acceptable.
4. For required steps: remove the flag, fix the underlying failure, and
   verify the job passes cleanly.
5. If the step is informational only (e.g., a lint advisory), prefer capturing
   output to a report rather than masking the failure outright.

## Validation

Remove the `continue-on-error` flag and re-run the job.  The job should fail
if the step fails, surfacing the underlying error.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `azure-pipelines.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain continue-on-error
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- CI step configured to continue on error
- Silent Failure: ci step configured to continue on error
- continue-on-error: true
- GitHub Actions ci step configured to continue on error
- faultline explain continue-on-error
- GitLab CI ci step configured to continue on error


---

*Generated from [playbooks/bundled/log/silent/continue-on-error.yaml](../../playbooks/bundled/log/silent/continue-on-error.yaml). Do not edit directly — run `make docs-generate`.*
