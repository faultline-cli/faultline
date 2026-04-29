# pnpm lockfile mismatch or frozen install failed

**Playbook ID:** `pnpm-lockfile`
**Category:** build
**Severity:** medium
**Tags:** `pnpm`, `lockfile`, `node`, `javascript`, `ci`

## What this failure means

The pnpm lockfile (`pnpm-lock.yaml`) is out of sync with `package.json`. CI uses `--frozen-lockfile` to ensure reproducible installs, so any discrepancy causes the install step to fail immediately.

## Common log signals

```text
ERR_PNPM_OUTDATED_LOCKFILE
ERR_PNPM_FROZEN_LOCKFILE
ERR_PNPM_LOCKFILE_VERSION
Cannot install with `frozen-lockfile`
pnpm-lock.yaml is not up to date
Lockfile is not up to date with
specifiers in the lockfile
don't match specs in package.json
```

## Diagnosis

The pnpm lockfile (`pnpm-lock.yaml`) is out of sync with `package.json`. CI uses `--frozen-lockfile` to ensure reproducible installs, so any discrepancy causes the install step to fail immediately.

## Fix steps

1. Run `pnpm install` locally to regenerate `pnpm-lock.yaml`.
2. Commit the updated `pnpm-lock.yaml` to the repository.
3. If using workspaces, run `pnpm install` from the workspace root.

## Validation

- Re-run the local reproduction command after the fix.
- pnpm install
- git diff pnpm-lock.yaml

## Likely files to inspect

- `pnpm-lock.yaml`
- `package.json`


## Run Faultline

```bash
faultline analyze build.log
faultline explain pnpm-lockfile
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- pnpm lockfile mismatch or frozen install failed
- Build: pnpm lockfile mismatch or frozen install failed
- Run `pnpm install` to update the lockfile
- GitHub Actions pnpm lockfile mismatch or frozen install failed
- faultline explain pnpm-lockfile
- pnpm pnpm lockfile mismatch or frozen install failed


---

*Generated from [playbooks/bundled/log/build/pnpm-lockfile.yaml](../../playbooks/bundled/log/build/pnpm-lockfile.yaml). Do not edit directly — run `make docs-generate`.*
