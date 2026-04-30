# Critical CI step skipped due to condition

**Playbook ID:** `skipped-critical-step`
**Category:** silent_failure
**Severity:** high
**Tags:** `silent-failure`, `skipped`, `condition`, `if`, `step`, `ci`, `build`, `test`, `deploy`, `security`

## What this failure means

A CI step was skipped because its `if:` condition evaluated to false.  The
context suggests the skipped step may have been a critical one (build, test,
deploy, security scan, coverage report, or publication), so the overall
success status may not reflect a complete pipeline run.

## Common log signals

```text
Skipping step due to condition
Condition evaluated to false
step was skipped
##[warning]step skipped
```

## Diagnosis

CI platforms evaluate `if:` (GitHub Actions), `when:` (GitLab CI), and
similar conditional expressions before each step.  When the condition is
false, the step is skipped and the job continues.  This is intentional for
steps that are only needed in certain contexts (e.g., deploy only on main).

However, when a critical step is unexpectedly skipped, the pipeline appears
to succeed while the associated work was never performed.  Common causes:

- A branch filter, environment variable check, or event filter in the `if:`
  expression is misconfigured.
- A dependent job that sets an output variable failed or was skipped, causing
  downstream conditions to evaluate unexpectedly.
- A PR from a fork does not have access to required secrets, causing
  secret-presence guards to fail and skip protected steps.
- A refactoring changed the step name or ID referenced by a dependent `needs`
  expression.

## Fix steps

1. Identify the skipped step in the CI log.
2. Review its condition expression and evaluate it manually against the
   current branch, event, and environment:
   ```yaml
   if: github.ref == 'refs/heads/main' && secrets.DEPLOY_TOKEN != ''
   ```
3. Check whether any upstream step that produces a required output or sets
   a required environment variable was itself skipped or failed.
4. For fork PRs: consider whether the step should be conditioned on
   `github.event_name != 'pull_request'` or use environment protection rules.
5. Add an explicit assertion or notification when a critical step is skipped
   in an unexpected context.

## Validation

Re-run the pipeline in the expected context (correct branch, event, or
environment) and confirm the previously skipped step now executes.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.gitlab-ci.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain skipped-critical-step
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Critical CI step skipped due to condition
- Silent Failure: critical ci step skipped due to condition
- Skipping step due to condition
- GitHub Actions critical ci step skipped due to condition
- faultline explain skipped-critical-step


---

*Generated from [playbooks/bundled/log/silent/skipped-critical-step.yaml](../../../playbooks/bundled/log/silent/skipped-critical-step.yaml). Do not edit directly — run `make docs-generate`.*
