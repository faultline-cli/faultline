# npm ERR! ERESOLVE unable to resolve dependency tree in CI

**Target search query:** `npm ERR! ERESOLVE unable to resolve dependency tree`

## Error snippet

```text
npm ERR! code ERESOLVE
npm ERR! ERESOLVE unable to resolve dependency tree
npm ERR! Found: react@18.3.1
npm ERR! Could not resolve dependency:
npm ERR! peer react@"^17" from some-package@2.4.0
```

## What this error means

npm found an explicit peer dependency conflict. The package graph is internally inconsistent, so CI refuses to continue with the install.

## Fix steps

1. Identify the version conflict in the `Found:` and `Could not resolve dependency:` lines.
2. Align the top-level package versions so the peer range is satisfied.
3. Regenerate the lockfile with `npm install`.
4. Avoid `--legacy-peer-deps` in CI unless you are intentionally accepting an unsupported combination.

## How Faultline detects it

Faultline maps this failure to `npm-peer-dependency-conflict`.

Primary signals:

- `npm ERR! code ERESOLVE`
- `ERESOLVE unable to resolve dependency tree`
- peer dependency conflict context

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [npm ci can only install packages when package.json and package-lock.json are in sync](npm-ci-package-lock-json-in-sync.md)
- [ERR_PNPM_OUTDATED_LOCKFILE: frozen-lockfile failed in CI](../pnpm/err-pnpm-outdated-lockfile.md)