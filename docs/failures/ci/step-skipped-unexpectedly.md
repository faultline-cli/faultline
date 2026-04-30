# CI step or job skipped unexpectedly

**Playbook ID:** `step-skipped-unexpectedly`
**Category:** ci
**Severity:** medium
**Tags:** `step`, `skipped`, `condition`, `if`, `when`, `ci`, `pipeline`

## What this failure means

A CI step, job, or stage was silently skipped due to an `if`/`when`/`only`
condition evaluating to false. The rest of the pipeline may appear green, but
a required step such as deployment, release, or security scan did not execute.

## Common log signals

```text
step was skipped
skipping step
job.*skipped
step.*skipped
condition.*evaluated to false
when.*condition.*false
needs.*skipped
skipped because
```

## Diagnosis

Steps are skipped when their conditional expression evaluates to false. This
can happen unexpectedly when:
1. A referenced output variable is empty or not set
2. A required upstream job was itself skipped (skip propagation)
3. A branch name or event type check does not match the current context
4. A step's `needs` job failed or was skipped, propagating the skip
5. An output from a prior step was not set correctly

GitHub Actions context for skipping:

```yaml
- name: Deploy
  if: github.ref == 'refs/heads/main' && success()
  # This step skips silently on any non-main branch AND when prior steps fail
```

Check whether the expected job ran:

```bash
# GitHub Actions CLI
gh run view <run-id> --log | grep -i "skipped\|Warning"
```

## Fix steps

1. Review the `if:` / `when:` / `rules:` condition on the skipped step and
   evaluate it against the actual values in the failing run:

   ```yaml
   # GitHub Actions — print context for debugging
   - name: Debug context
     run: |
       echo "ref: ${{ github.ref }}"
       echo "event: ${{ github.event_name }}"
       echo "success: ${{ job.status }}"
   ```

2. For GitHub Actions, check that upstream job status is correctly handled:

   ```yaml
   deploy:
     needs: [build, test]
     if: success()     # skips if build or test was skipped/failed
   ```

   Use `always()` or explicit status checks if partial failures should
   still allow the step to run:

   ```yaml
   notify:
     needs: [build, test]
     if: always()   # runs regardless of upstream status
   ```

3. For GitLab CI, check `rules:` and `when:` conditions:

   ```yaml
   # Add 'when: always' to ensure certain jobs always run
   notify:
     rules:
       - when: always
   ```

4. Check whether a step's output was properly set before a consuming step
   that gates on it:

   ```yaml
   - id: check_changes
     run: echo "changed=true" >> $GITHUB_OUTPUT

   - if: steps.check_changes.outputs.changed == 'true'
     run: ...   # skips if the output was never set or set to 'false'
   ```

5. If the condition is legitimately preventing the step, add
   `workflow_dispatch` or a manual trigger to force-run it for debugging.

## Validation

- Re-trigger the pipeline with the corrected condition.
- Confirm the previously-skipped step now appears as "completed" in the
  pipeline view.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain step-skipped-unexpectedly
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- CI step or job skipped unexpectedly
- Ci: ci step or job skipped unexpectedly
- condition.*evaluated to false
- GitHub Actions ci step or job skipped unexpectedly
- faultline explain step-skipped-unexpectedly


---

*Generated from [playbooks/bundled/log/ci/step-skipped-unexpectedly.yaml](../../../playbooks/bundled/log/ci/step-skipped-unexpectedly.yaml). Do not edit directly — run `make docs-generate`.*
