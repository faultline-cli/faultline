# CI workflow not triggered due to path or branch filter

**Playbook ID:** `workflow-not-triggered`
**Category:** ci
**Severity:** medium
**Tags:** `trigger`, `path-filter`, `branch`, `workflow`, `ci`, `skipped`

## What this failure means

A CI workflow was expected to run after a push or pull request but did not
trigger. The `paths`, `paths-ignore`, `branches`, or similar filter
conditions in the workflow configuration did not match the changed files or
target branch, so the CI system silently skipped the workflow.

## Common log signals

```text
no workflows matched
workflow.*not triggered
path filter
paths.*not matched
changes.*not in watched path
skipping workflow
workflow.*skipped
not running.*no changes
```

## Diagnosis

Trigger filters are evaluated against the exact paths and branch names
of the event. Common mistakes include:

1. A `paths` filter that does not cover a recently added directory
2. An overly specific `branches` filter that excludes a new naming pattern
3. A path glob that uses wrong syntax (e.g., `src/**` only matches
   one directory deep in some systems vs. all descendants)
4. A `push` filter that does not include the branch the team now uses

GitHub Actions example of a filter that silently skips:

```yaml
on:
  push:
    paths:
      - 'src/**'      # Misses changes under 'lib/' or 'packages/'
    branches:
      - main          # Misses 'develop' or feature branches
```

## Fix steps

1. Verify the event type and branch match the `on:` / `trigger:` block.

2. In GitHub Actions, use the `workflow_dispatch` event to manually trigger
   the workflow and confirm the job itself is functional:

   ```yaml
   on:
     push: ...
     workflow_dispatch:   # add this for manual testing
   ```

3. Widen the path filter to include all relevant directories:

   ```yaml
   on:
     push:
       paths:
         - 'src/**'
         - 'lib/**'
         - 'packages/**'
         - '.github/workflows/*.yml'
   ```

4. Check glob syntax: `**` in GitHub Actions matches zero or more path
   segments, but in some CI systems it only matches within a directory.

5. For GitLab CI, review `rules:` and `only:/except:` conditions:

   ```yaml
   test:
     rules:
       - if: '$CI_PIPELINE_SOURCE == "push"'
         changes:
           - src/**/*    # use /**/* not just /**
   ```

6. For CircleCI, check `filters.branches` and `when:` conditions.

7. Remove overly restrictive filters that were added to save CI cost but
   now exclude needed branches or files.

## Validation

- Push a change that matches the filter and confirm the workflow appears
  in CI.
- Use `act` (GitHub Actions local runner) or the CI platform's lint/dry-run
  to preview which workflows would trigger for a given event.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain workflow-not-triggered
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- CI workflow not triggered due to path or branch filter
- Ci: ci workflow not triggered due to path or branch filter
- changes.*not in watched path
- GitHub Actions ci workflow not triggered due to path or branch filter
- faultline explain workflow-not-triggered


---

*Generated from [playbooks/bundled/log/ci/workflow-not-triggered.yaml](../../playbooks/bundled/log/ci/workflow-not-triggered.yaml). Do not edit directly — run `make docs-generate`.*
