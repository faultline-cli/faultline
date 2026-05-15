# pnpm frozen lockfile missing or mismatch

**Playbook ID:** `pnpm-lockfile-missing`
**Category:** build
**Severity:** high
**Tags:** `pnpm`, `node`, `lockfile`, `dependency`

## What this failure means

pnpm failed because the lockfile is missing, out of sync, or was not committed. pnpm requires an exact lockfile reference for reproducible installs.

## Common log signals

```text
ERR_PNPM_FROZEN_LOCKFILE_CHANGED
pnpm-lock.yaml.*not in sync
lockfile.*missing
pnpm ERR!
```

## Diagnosis

pnpm is stricter than npm and yarn about lockfile consistency. Common causes:

- `pnpm-lock.yaml` was not committed to the repository.
- The lockfile is out of sync with `package.json` (versions changed but lockfile not updated).
- `pnpm install` was run locally but the updated lockfile was not pushed before CI.
- CI is configured with `--frozen-lockfile` but the lockfile is stale.

The error appears as `ERR_PNPM_FROZEN_LOCKFILE_CHANGED` or similar pnpm validation errors.

## Fix steps

1. Ensure `pnpm-lock.yaml` is committed to the repository:

   ```bash
   git status pnpm-lock.yaml
   ```

2. If `pnpm-lock.yaml` is missing, regenerate it locally:

   ```bash
   pnpm install
   git add pnpm-lock.yaml
   git commit -m "Update pnpm lockfile"
   git push
   ```

3. If the lockfile is out of sync with `package.json`, update it:

   ```bash
   pnpm install
   git add pnpm-lock.yaml
   git commit -m "Sync pnpm lockfile"
   git push
   ```

4. Verify the lockfile is in sync before re-running CI:

   ```bash
   pnpm install --frozen-lockfile
   ```

## Validation

- `pnpm install --frozen-lockfile` completes without errors.
- `git status` shows no changes to `pnpm-lock.yaml`.
- Re-run the CI job.

## Likely files to inspect

- `pnpm-lock.yaml`
- `package.json`
- `.gitlab-ci.yml`
- `.github/workflows/*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain pnpm-lockfile-missing
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- pnpm frozen lockfile missing or mismatch
- Build: pnpm frozen lockfile missing or mismatch
- ERR_PNPM_FROZEN_LOCKFILE_CHANGED
- GitHub Actions pnpm frozen lockfile missing or mismatch
- faultline explain pnpm-lockfile-missing
- pnpm pnpm frozen lockfile missing or mismatch


---

*Generated from [playbooks/bundled/log/build/pnpm-lockfile-missing.yaml](../../../playbooks/bundled/log/build/pnpm-lockfile-missing.yaml). Do not edit directly — run `make docs-generate`.*
