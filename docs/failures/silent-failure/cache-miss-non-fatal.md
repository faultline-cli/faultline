# Cache restore or save failed but job continued

**Playbook ID:** `cache-miss-non-fatal`
**Category:** silent_failure
**Severity:** medium
**Tags:** `silent-failure`, `cache`, `github-actions`, `ci`, `performance`

## What this failure means

A cache restore or save step failed (cache not found, restore error, or save
error), but the CI job continued without flagging this as a failure.  Repeated
cache misses degrade CI performance and may indicate a configuration problem
or dependency issue.

## Common log signals

```text
Cache not found
Failed to restore cache
Failed to save cache
cache miss
No cache found
```

## Diagnosis

Cache steps (`actions/cache`, GitLab CI `cache:`, CircleCI `restore_cache`)
are designed to be non-fatal: a cache miss simply means the job builds from
scratch.  However, persistent cache failures can indicate:

- The cache key or path is misconfigured and will never produce a hit.
- The cache backend is experiencing availability issues.
- The cache was invalidated by a dependency change that should itself be
  investigated.
- Save failures may indicate insufficient storage or permissions on the
  runner.

While a single cache miss is expected behaviour, repeated non-fatal cache
failures mask a real configuration or infrastructure problem.

## Fix steps

1. Check the cache key expression:
   ```yaml
   key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
   ```
   Ensure the key file exists and the hash produces a stable value.
2. Verify the cache path matches the actual dependency directory
   (e.g., `~/.npm` vs `node_modules`).
3. For restore failures: confirm the cache was saved on a previous run
   and that the key matches.
4. For save failures: check runner disk space and storage quota.
5. Review CI logs for additional error messages after the cache failure.

## Validation

On the next CI run, verify the cache step reports a cache hit or a
successful save.  Monitor the cache hit rate across recent runs.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.gitlab-ci.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain cache-miss-non-fatal
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Cache restore or save failed but job continued
- Silent Failure: cache restore or save failed but job continued
- Failed to restore cache
- GitHub Actions cache restore or save failed but job continued
- faultline explain cache-miss-non-fatal


---

*Generated from [playbooks/bundled/log/silent/cache-miss-non-fatal.yaml](../../playbooks/bundled/log/silent/cache-miss-non-fatal.yaml). Do not edit directly — run `make docs-generate`.*
