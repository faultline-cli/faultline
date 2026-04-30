# GitHub Actions job cancelled by concurrency policy

**Playbook ID:** `github-actions-concurrency-cancel`
**Category:** ci
**Severity:** medium
**Tags:** `github-actions`, `concurrency`, `cancel`, `cancelled`, `workflow`

## What this failure means

A GitHub Actions workflow run was automatically cancelled because a newer
run for the same concurrency group started. All pending and in-progress
steps were cut short. No action or fix is needed — this is expected
behaviour from a `concurrency: cancel-in-progress: true` policy.

## Common log signals

```text
The operation was canceled
the operation was cancelled
This run was cancelled
Cancelling a currently running workflow
Job was canceled
##[error]A task was canceled
```

## Diagnosis

GitHub Actions supports a `concurrency` block at the workflow or job level.
When `cancel-in-progress: true` is set, any in-progress run for the same
group is cancelled as soon as a newer run begins for that group (e.g., a
new push to the same branch):

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

When cancellation fires:
- The runner receives a cancel signal.
- Any currently running step receives `SIGTERM`; pending steps are skipped.
- The log shows: `The operation was canceled.` or
  `Cancelling a currently running workflow` followed by all remaining steps
  marked as cancelled or skipped.

This is not a failure — it is the concurrency policy working as designed.
If the cancellation is unexpected, check whether the branch or event
grouping expression is too broad.

## Fix steps

**If the cancellation is expected (new push after old push):**
No action needed. The newer run will complete.

**If the cancellation is too aggressive (unrelated runs cancelling each
other):**

1. Narrow the concurrency group expression to avoid false matches:

   ```yaml
   # Too broad — all pulls cancel each other regardless of branch:
   concurrency:
     group: ${{ github.workflow }}
     cancel-in-progress: true

   # Better — only runs on the same branch cancel each other:
   concurrency:
     group: ${{ github.workflow }}-${{ github.ref }}
     cancel-in-progress: true
   ```

2. Protect critical post-merge jobs from cancellation by setting
   `cancel-in-progress: false` on the deploy or release workflow:

   ```yaml
   concurrency:
     group: deploy-${{ github.ref }}
     cancel-in-progress: false  # Never cancel an in-progress deploy
   ```

3. Use separate concurrency groups for PR checks (cancel-in-progress: true)
   and post-merge deployment (cancel-in-progress: false).

**If the concurrency block is not intentional:**

Remove or adjust the `concurrency:` block from the workflow YAML.

## Validation

- Push a new commit and confirm only the newer run executes.
- Confirm no unrelated runs are cancelled under the new configuration.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain github-actions-concurrency-cancel
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- GitHub Actions job cancelled by concurrency policy
- Ci: github actions job cancelled by concurrency policy
- Cancelling a currently running workflow
- GitHub Actions github actions job cancelled by concurrency policy
- faultline explain github-actions-concurrency-cancel


---

*Generated from [playbooks/bundled/log/ci/github-actions-concurrency-cancel.yaml](../../../playbooks/bundled/log/ci/github-actions-concurrency-cancel.yaml). Do not edit directly — run `make docs-generate`.*
