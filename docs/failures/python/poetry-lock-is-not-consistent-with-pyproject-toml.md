# poetry.lock is not consistent with pyproject.toml in CI

**Target search query:** `poetry.lock is not consistent with pyproject.toml`

## Error snippet

```text
poetry.lock is not consistent with pyproject.toml. Run `poetry lock [--no-update]` to fix it.
```

## What this error means

The Python dependency manifest changed, but the Poetry lockfile did not. CI is trying to build from a stale dependency graph.

## Fix steps

1. Run `poetry lock --no-update` locally to regenerate `poetry.lock`.
2. If resolution still fails, fix the incompatible constraint in `pyproject.toml` and re-run the lock command.
3. Commit `pyproject.toml` and `poetry.lock` together.
4. Re-run `poetry install` from a clean checkout before pushing.

## How Faultline detects it

Faultline maps this failure to `poetry-lockfile-drift`.

Primary signals:

- `poetry.lock is not consistent with pyproject.toml`
- `Run \`poetry lock [--no-update]\` to fix it.`
- lockfile drift during dependency install

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [ERR_PNPM_OUTDATED_LOCKFILE: frozen-lockfile failed in CI](../pnpm/err-pnpm-outdated-lockfile.md)
- [npm ci can only install packages when package.json and package-lock.json are in sync](../npm/npm-ci-package-lock-json-in-sync.md)