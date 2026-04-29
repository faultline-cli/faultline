# Disk space exhausted

**Playbook ID:** `disk-full`
**Category:** runtime
**Severity:** critical
**Tags:** `disk`, `storage`, `no space`

## What this failure means

The host or container filesystem ran out of disk space. Writes fail
immediately when the disk is full, typically aborting builds, unpack steps,
or runtime file creation.

## Common log signals

```text
disk full
file system is full
no space remaining
not enough space
no space left on device
ENOSPC
errno 28
Write failed: No space left
```

## Diagnosis

The host or container filesystem ran out of disk space. This starter rule is
intentionally kept separate from CI-runner-specific disk exhaustion so the
remediation can stay focused on the machine or filesystem that actually ran
out of space.

## Fix steps

1. Check how much space remains on the filesystem that contains the failing
   path, and identify which paths are largest:

   ```bash
   df -h
   du -sh /* 2>/dev/null | sort -rh | head -20
   ```

2. If Docker is part of the job, prune images, containers, and build cache in
   one pass:

   ```bash
   docker system prune -af --volumes
   docker system df  # confirm how much was recovered
   ```

3. Remove large intermediate build artifacts: `make clean`, `gradle clean`,
   `cargo clean`, or the project equivalent.
4. Clear accumulated language-specific caches:

   ```bash
   rm -rf ~/.npm ~/.gradle/caches ~/.m2/repository ~/.cache/pip
   ```

5. Add a disk space guard at the start of large jobs so failures are explicit
   rather than silent mid-step:

   ```bash
   AVAIL_KB=$(df --output=avail / | tail -1)
   [ "$AVAIL_KB" -lt 2097152 ] && echo "ERROR: less than 2 GB free" && exit 1
   ```

6. If cleanup is not enough, increase the runner disk size or split build and
   test stages into separate smaller jobs with lighter footprints.

## Validation

- `df -h` — confirm at least 1–2 GB is available before re-running.
- Re-run the failing step and confirm it completes without an ENOSPC error.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `Dockerfile`
- `docker-compose*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain disk-full
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Disk space exhausted
- Runtime: disk space exhausted
- Write failed: No space left
- faultline explain disk-full


---

*Generated from [playbooks/bundled/log/runtime/disk-full.yaml](../../playbooks/bundled/log/runtime/disk-full.yaml). Do not edit directly — run `make docs-generate`.*
