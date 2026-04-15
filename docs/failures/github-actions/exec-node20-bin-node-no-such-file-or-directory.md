# exec /__e/node20/bin/node: no such file or directory in GitHub Actions

**Target search query:** `exec /__e/node20/bin/node no such file or directory`

## Error snippet

```text
exec /__e/node20/bin/node: no such file or directory
```

## What this error means

The workflow expected a Node runtime at a hard-coded GitHub Actions runner path, but that binary was missing from the active runner or container environment.

## Fix steps

1. Pin the affected action version instead of floating to a moving major tag.
2. Verify the runner image and any containerized job environment include the expected Node runtime.
3. If the path is hard-coded in a wrapper or action, update it to use the current installed location.
4. Re-run `command -v node` and inspect the expected path inside the same job environment.

## How Faultline detects it

Faultline maps this failure to `missing-executable`.

Primary signals:

- `exec /`
- `no such file or directory`
- missing runtime path context under a CI runner

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [fatal: repository not found in GitHub Actions or GitLab CI](../git/fatal-repository-not-found.md)
- [npm ci can only install packages when package.json and package-lock.json are in sync](../npm/npm-ci-package-lock-json-in-sync.md)