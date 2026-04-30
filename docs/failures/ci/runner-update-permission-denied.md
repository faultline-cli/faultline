# Self-hosted runner update permission failure

**Playbook ID:** `runner-update-permission-denied`
**Category:** ci
**Severity:** high
**Tags:** `github-actions`, `runner`, `update`, `permissions`, `ci`

## What this failure means

A self-hosted runner tried to update itself or adjust worker process settings, but the environment denied access to the runner filesystem or proc entries.

## Common log signals

```text
Downloading runner update fails
An error occurred: Access to the path is denied
Access to the path '/proc/
Failed to update oom_score_adj
Runner update in progress, do not shutdown runner
UnauthorizedAccessException
```

## Diagnosis

A self-hosted runner tried to update itself or adjust worker process settings, but the environment denied access to the runner filesystem or proc entries.

This usually happens when the runner is containerized with overly strict ownership or when the job environment blocks access to `_work` or `/proc/*/oom_score_adj`.

## Fix steps

1. Check the ownership and permissions on the runner root, work directory, and proc mount that the runner is trying to touch.
2. If the runner is containerized, ensure the container runs with the user and group expected by the runner image.
3. Confirm the runner can write to its working directory and update its own binaries or configuration.
4. If the failure only appears after an autoscaling or container-mode change, compare the pod or container security context with a known-good runner deployment.

## Validation

- The runner starts without `Access to the path is denied`.
- The update step completes and the job proceeds past runner startup.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `.gitlab-ci.yml`
- `docker-compose*.yml`
- `k8s/**/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain runner-update-permission-denied
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Self-hosted runner update permission failure
- Ci: self-hosted runner update permission failure
- Runner update in progress, do not shutdown runner
- GitHub Actions self-hosted runner update permission failure
- faultline explain runner-update-permission-denied


---

*Generated from [playbooks/bundled/log/ci/runner-update-permission-denied.yaml](../../../playbooks/bundled/log/ci/runner-update-permission-denied.yaml). Do not edit directly — run `make docs-generate`.*
