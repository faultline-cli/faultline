# Corrupted or stale dependency cache

**Playbook ID:** `cache-corruption`
**Category:** build
**Severity:** medium
**Tags:** `cache`, `dependencies`, `build`

## What this failure means

CI restored a dependency cache, but the cached contents were stale,
corrupted, or built for an incompatible environment.

## Common log signals

```text
cache is corrupted
corrupt cache
invalid cache
cache hit but install failed
checksum mismatch
hash mismatch
unexpected end of archive
corrupted tar archive
```

## Diagnosis

The cache layer existed and restored successfully, but one or more cached
artifacts could not be trusted for the current build.

This usually happens after toolchain upgrades, runner OS changes, or partial
writes from an interrupted previous job.

## Fix steps

1. Delete the CI cache entry that backs the failing step and allow it to
   rebuild from scratch.
2. Invalidate the cache by changing the cache key, for example by appending a
   version suffix or toolchain hash.
3. For GitHub Actions, clear the cache via Settings > Actions > Caches.
4. If the cache stores platform-specific artifacts, verify the key includes
   OS, architecture, and toolchain version before restoring it again.
5. If the failure is local to one package manager, clear only that tool's
   cache rather than wiping the entire home cache directory.

## Validation

- Re-run the build from a cold cache.
- Confirm dependency install or linking succeeds without archive or checksum
  corruption.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `Dockerfile`
- `package-lock.json`


## Run Faultline

```bash
faultline analyze build.log
faultline explain cache-corruption
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Corrupted or stale dependency cache
- Build: corrupted or stale dependency cache
- cache hit but install failed
- GitHub Actions corrupted or stale dependency cache
- faultline explain cache-corruption


---

*Generated from [playbooks/bundled/log/build/cache-corruption.yaml](../../playbooks/bundled/log/build/cache-corruption.yaml). Do not edit directly — run `make docs-generate`.*
