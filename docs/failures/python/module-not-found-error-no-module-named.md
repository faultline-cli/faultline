# ModuleNotFoundError: No module named in Python CI jobs

**Target search query:** `ModuleNotFoundError No module named CI`

## Error snippet

```text
ModuleNotFoundError: No module named 'requests'
```

## What this error means

Python could not import a required module because the dependency is missing from the environment running the CI step, or the import path does not match the installed package.

## Fix steps

1. Ensure the dependency install step runs before tests or startup in the same environment.
2. Verify the declared package name matches the import used in code.
3. Confirm CI activates the intended virtual environment and Python interpreter.
4. Reinstall dependencies and run `pip check` before rerunning the failing command.

## How Faultline detects it

Faultline maps this failure to `python-module-missing`.

Primary signals:

- `ModuleNotFoundError: No module named`
- `ImportError: No module named`
- `cannot import name`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [poetry.lock is not consistent with pyproject.toml in CI](poetry-lock-is-not-consistent-with-pyproject-toml.md)
- [missing go.sum entry for module providing package](../go/missing-go-sum-entry-for-module-providing-package.md)