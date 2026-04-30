# Remote build cache miss or misconfiguration

**Playbook ID:** `remote-cache-miss`
**Category:** build
**Severity:** low
**Tags:** `cache`, `remote-cache`, `bazel`, `turbo`, `gradle`, `nx`, `build-cache`

## What this failure means

A remote build cache (Bazel remote cache, Gradle build cache, Turborepo
remote cache, Nx Cloud, etc.) is not populated, misconfigured, or
unavailable. The build falls back to full local builds on every run,
dramatically increasing build and CI execution time.

## Common log signals

```text
remote cache.*miss
cache miss
cache hit rate.*0%
remote cache.*unavailable
cache write.*failed
cache read.*failed
could not contact cache
cache.*connection refused
```

## Diagnosis

Remote caches fail to populate or read for several reasons:
1. Cache credentials are absent or expired
2. The cache server URL is wrong or the server is down
3. The cache key inputs changed, invalidating all existing entries
4. CI runs on a different OS or architecture than the cached builds
5. The cache write permission is revoked for fork PRs

Check cache statistics in the build log:

```
# Bazel
INFO: 42 processes: 42 linux-sandbox, 0 remote cache hit

# Gradle
17 tasks executed, 0 from cache

# Turborepo
Tasks:    10 successful, 0 cached, 0 skipped
```

A 0% cache hit rate on repeated identical builds indicates the cache is
either not being written to or the key is always invalidating.

## Fix steps

1. Verify the remote cache endpoint is reachable from CI:

   ```bash
   curl -I https://cache.example.com/healthcheck
   ```

2. Check that cache credentials are present in the CI environment:

   ```bash
   # Print non-secret credential presence (not value)
   echo "credential set: $([ -n "$CACHE_TOKEN" ] && echo yes || echo NO)"
   ```

3. Review cache key configuration — overly volatile inputs (timestamps,
   random values) produce unique cache keys on every run:

   ```yaml
   # GitHub Actions cache
   - uses: actions/cache@v4
     with:
       key: ${{ runner.os }}-build-${{ hashFiles('**/package-lock.json') }}
       # NOT: key: ${{ github.sha }}  -- always misses
   ```

4. For Turborepo, configure the remote cache:

   ```bash
   npx turbo run build --team=myteam --token=$TURBO_TOKEN
   ```

   Or in `turbo.json`:

   ```json
   {
     "remoteCache": { "enabled": true }
   }
   ```

5. For fork PRs, remote caches are typically read-only. Ensure the fallback
   to a local or read-from-upstream-cache strategy is configured.

6. For Bazel, check `.bazelrc` for remote cache settings:

   ```
   build --remote_cache=grpc://cache.example.com:9090
   build --google_credentials=/path/to/credentials.json
   ```

## Validation

- Trigger two identical CI runs back-to-back.
- The second run should show a high cache hit rate.
- Confirm build times on the second run are significantly shorter.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain remote-cache-miss
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Remote build cache miss or misconfiguration
- Build: remote build cache miss or misconfiguration
- remote execution.*unavailable
- GitHub Actions remote build cache miss or misconfiguration
- faultline explain remote-cache-miss
- Gradle remote build cache miss or misconfiguration


---

*Generated from [playbooks/bundled/log/build/remote-cache-miss.yaml](../../../playbooks/bundled/log/build/remote-cache-miss.yaml). Do not edit directly — run `make docs-generate`.*
