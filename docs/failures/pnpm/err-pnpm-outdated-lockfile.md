# ERR_PNPM_OUTDATED_LOCKFILE: frozen-lockfile failed in CI

**Target search query:** `ERR_PNPM_OUTDATED_LOCKFILE frozen-lockfile`

## Error snippet

```text
ERR_PNPM_OUTDATED_LOCKFILE Cannot install with "frozen-lockfile" because pnpm-lock.yaml is not up to date with package.json
```

## What this error means

CI is enforcing reproducible installs with `--frozen-lockfile`, but `pnpm-lock.yaml` no longer matches the current dependency manifest.

## Fix steps

1. Run `pnpm install` locally to regenerate `pnpm-lock.yaml`.
2. Commit the updated lockfile.
3. If you use workspaces, regenerate the lockfile from the workspace root.
4. Re-run the frozen install from a clean checkout.

## How Faultline detects it

Faultline maps this failure to `pnpm-lockfile`.

Primary signals:

- `ERR_PNPM_OUTDATED_LOCKFILE`
- `Cannot install with \`frozen-lockfile\``
- `pnpm-lock.yaml is not up to date`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [npm ci can only install packages when package.json and package-lock.json are in sync](../npm/npm-ci-package-lock-json-in-sync.md)
- [poetry.lock is not consistent with pyproject.toml in CI](../python/poetry-lock-is-not-consistent-with-pyproject-toml.md)