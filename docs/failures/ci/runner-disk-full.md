# CI runner disk full

**Playbook ID:** `runner-disk-full`
**Category:** ci
**Severity:** critical
**Tags:** `disk`, `storage`, `runner`, `ci`, `space`

## What this failure means

The CI runner has run out of available disk space. The build, test, or artifact step cannot write files and fails with ENOSPC.

## Common log signals

```text
No space left on device
disk quota exceeded
Write failed: No space left
not enough disk space
insufficient disk space
out of disk space
Disk is full
Error processing tar file
```

## Diagnosis

The CI runner has run out of available disk space. The build, test, or artifact step cannot write files and fails with ENOSPC.

## Fix steps

1. Check remaining space before diagnosing: `df -h` and `du -sh /* | sort -rh | head -20`.
2. Remove Docker build cache and dangling images: `docker system prune -af --volumes`.
3. Clear dependency caches: `rm -rf ~/.npm ~/.gradle/caches ~/.m2/repository ~/.cache/pip`.
4. In GitHub Actions, add the `jlumbroso/free-disk-space` action before large build steps.
5. In GitLab CI, prune images in `before_script` and reduce artifact retention.
6. In CircleCI, choose a larger `resource_class` or clean up large workspace directories.
7. Upgrade to a runner or agent tier with more disk, or split the job across multiple stages.
8. Reduce `.dockerignore` exclusions to keep the build context small.
9. Add a cheap guard near the top of the job so storage problems fail fast with a clear message instead of surfacing mid-build.

## Validation

- Re-run the failing workflow step after confirming free space with `df -h`.
- Confirm the original failure signature for CI runner disk full is gone.

## Likely files to inspect

- `.dockerignore`
- `Dockerfile`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain runner-disk-full
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- CI runner disk full
- Ci: ci runner disk full
- Write failed: No space left
- GitHub Actions ci runner disk full
- faultline explain runner-disk-full


---

*Generated from [playbooks/bundled/log/ci/runner-disk-full.yaml](../../playbooks/bundled/log/ci/runner-disk-full.yaml). Do not edit directly — run `make docs-generate`.*
