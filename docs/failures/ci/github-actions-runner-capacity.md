# GitHub Actions runner capacity or queue timeout

**Playbook ID:** `github-actions-runner-capacity`
**Category:** ci
**Severity:** medium
**Tags:** `github-actions`, `runner`, `capacity`, `queue`, `timeout`, `hosted`

## What this failure means

The GitHub Actions job waited too long to be picked up by a runner, or the runner became unresponsive mid-job. The workflow was canceled due to the capacity or timeout condition.

## Common log signals

```text
All runners that can run this job are busy
Waiting for a runner to pick up this job
No runner matching the labels
runner was lost
The runner has received a shutdown signal
runner.*disconnected
Job was canceled because no runners matched
No available runner
```

## Diagnosis

The GitHub Actions job waited too long to be picked up by a runner, or the runner became unresponsive mid-job. The workflow was canceled due to the capacity or timeout condition.

## Fix steps

1. Re-run the failed jobs: transient capacity issues resolve as runners free up.
2. For self-hosted runners: verify the runner is registered and healthy in Settings → Actions → Runners.
3. If jobs regularly queue: increase the runner count by adding more self-hosted runners or upgrading the GitHub plan for additional hosted runner seats.
4. If the workflow timeout was reached: increase `timeout-minutes` at the job or step level, or split long jobs into shorter parallel steps.
5. Check runner labels: ensure `runs-on` labels match a valid registered runner label exactly.

## Validation

- Re-run the local reproduction command after the fix.
- Check GitHub Settings → Actions → Runners for registered runner health

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain github-actions-runner-capacity
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- GitHub Actions runner capacity or queue timeout
- Ci: github actions runner capacity or queue timeout
- Job was canceled because no runners matched
- GitHub Actions github actions runner capacity or queue timeout
- faultline explain github-actions-runner-capacity


---

*Generated from [playbooks/bundled/log/ci/github-actions-runner-capacity.yaml](../../playbooks/bundled/log/ci/github-actions-runner-capacity.yaml). Do not edit directly — run `make docs-generate`.*
