# Your lockfile needs to be updated, but yarn was run with --frozen-lockfile

**Target search query:** `your lockfile needs to be updated but yarn was run with --frozen-lockfile`

## Error snippet

```text
error Your lockfile needs to be updated, but yarn was run with `--frozen-lockfile`.
```

## What this error means

Yarn is enforcing reproducible installs in CI, but `yarn.lock` is missing or out of sync with `package.json`.

## Fix steps

1. Run `yarn install` locally.
2. Commit the updated `yarn.lock`.
3. If using workspaces, regenerate the lockfile from the workspace root with the same Yarn version used in CI.
4. Re-run the frozen install step from a clean checkout.

## How Faultline detects it

Faultline maps this failure to `yarn-lockfile`.

Primary signals:

- `your lockfile needs to be updated`
- `frozen-lockfile`
- `yarn install --frozen-lockfile`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [npm ci can only install packages when package.json and package-lock.json are in sync](../npm/npm-ci-package-lock-json-in-sync.md)
- [ERR_PNPM_OUTDATED_LOCKFILE: frozen-lockfile failed in CI](../pnpm/err-pnpm-outdated-lockfile.md)