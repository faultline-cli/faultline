# The engine "node" is incompatible with this module

**Target search query:** `The engine "node" is incompatible with this module CI`

## Error snippet

```text
error package@1.0.0: The engine "node" is incompatible with this module. Expected version ">=20". Got "18.20.4"
```

## What this error means

The Node.js version on the runner does not satisfy the version constraint declared by the project or dependency metadata.

## Fix steps

1. Identify the required version from `.nvmrc`, `.node-version`, or `package.json` `engines.node`.
2. Configure CI to install that exact Node version instead of using the runner default.
3. Keep local and CI Node versions aligned with the same pinned source of truth.
4. Re-run `node --version` in CI before the install step.

## How Faultline detects it

Faultline maps this failure to `node-version-mismatch`.

Primary signals:

- `The engine "node" is incompatible with this module`
- `Unsupported engine`
- `Expected node version`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [JavaScript heap out of memory in Node.js builds](javascript-heap-out-of-memory.md)
- [npm ci can only install packages when package.json and package-lock.json are in sync](../npm/npm-ci-package-lock-json-in-sync.md)