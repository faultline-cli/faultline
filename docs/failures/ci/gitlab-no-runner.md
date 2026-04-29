# GitLab CI job stuck — no matching runner

**Playbook ID:** `gitlab-no-runner`
**Category:** ci
**Severity:** high
**Tags:** `gitlab`, `runner`, `stuck`, `ci`

## What this failure means

GitLab CI could not find an available runner that matches the job's tag requirements or the project has no registered runners at all.

## Common log signals

```text
This job is stuck because the project doesn't have any runners
This job is stuck because you don't have any active runners
no runners available
job is stuck
runner not found
waiting for a runner
```

## Diagnosis

GitLab CI could not find an available runner that matches the job's tag requirements or the project has no registered runners at all.

## Fix steps

1. Check runner availability: Settings > CI/CD > Runners.
2. Ensure at least one runner is online and not paused.
3. If the job requires specific tags, confirm at least one runner has those tags.
4. For group or instance runners, verify they are enabled for this project.
5. Start a new runner with `gitlab-runner register` if none are available.

## Validation

- Re-run the failing workflow step.
- Confirm the original failure signature for GitLab CI job stuck — no matching runner is gone.

## Likely files to inspect

- `.gitlab-ci.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain gitlab-no-runner
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- GitLab CI job stuck — no matching runner
- Ci: gitlab ci job stuck — no matching runner
- This job is stuck because the project doesn't have any runners
- GitHub Actions gitlab ci job stuck — no matching runner
- faultline explain gitlab-no-runner


---

*Generated from [playbooks/bundled/log/ci/gitlab-no-runner.yaml](../../playbooks/bundled/log/ci/gitlab-no-runner.yaml). Do not edit directly — run `make docs-generate`.*
