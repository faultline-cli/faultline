# npm ci can only install packages when package.json and package-lock.json are in sync

**Target search query:** `npm ci can only install packages when package.json and package-lock.json are in sync`

## Error snippet

```text
npm ERR! code EUSAGE
npm ERR! `npm ci` can only install packages when your package.json and package-lock.json are in sync.
npm ERR! Missing: eslint@9.0.0 from lock file
```

## What this error means

`npm ci` installs strictly from `package-lock.json`. If the manifest changed without a matching lockfile update, CI stops instead of mutating dependencies on the fly.

## Fix steps

1. Run `npm install` locally to regenerate `package-lock.json`.
2. Commit the updated lockfile.
3. Make sure `package-lock.json` is not ignored and is generated from the workspace root if you use workspaces.
4. Re-run `npm ci` locally before pushing the fix.

## How Faultline detects it

Faultline maps this failure to `npm-ci-lockfile`.

Primary signals:

- `npm ci can only install packages when your package.json and package-lock.json`
- `package.json and package-lock.json are in sync`
- `from lock file`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [npm ERR! ERESOLVE unable to resolve dependency tree in CI](npm-err-eresolve-unable-to-resolve-dependency-tree.md)
- [ERR_PNPM_OUTDATED_LOCKFILE: frozen-lockfile failed in CI](../pnpm/err-pnpm-outdated-lockfile.md)