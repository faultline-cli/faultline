# Poetry lockfile drift

**Playbook ID:** `poetry-lockfile-drift`
**Category:** build
**Severity:** medium
**Tags:** `python`, `poetry`, `lockfile`, `dependencies`

## What this failure means

Poetry is installing from a lockfile that no longer matches `pyproject.toml`, so dependency resolution fails before the environment is usable.

## Common log signals

```text
poetry.lock is not consistent with pyproject.toml
Run `poetry lock [--no-update]` to fix it.
version solving failed
```

## Diagnosis

The project manifest and `poetry.lock` have drifted apart. Poetry warns that the lockfile is stale, then fails to solve or install the dependency set that the project now declares.

Any later import error is usually a downstream effect of the failed install, not the primary root cause.

## Fix steps

1. Regenerate the lockfile from the current manifest:

   ```bash
   poetry lock --no-update
   ```

2. If resolution still fails, update the incompatible dependency constraint in `pyproject.toml` and re-run `poetry lock`.
3. Commit both `pyproject.toml` and `poetry.lock` together.
4. Recreate the virtual environment if Poetry cached a broken environment from an earlier failed install.

## Validation

- Run `poetry install` from a clean checkout.
- Confirm Poetry no longer warns that `poetry.lock` is inconsistent with `pyproject.toml`.
- Re-run the failing Python command after the install completes successfully.

## Likely files to inspect

- `pyproject.toml`
- `poetry.lock`


## Run Faultline

```bash
faultline analyze build.log
faultline explain poetry-lockfile-drift
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Poetry lockfile drift
- Build: poetry lockfile drift
- poetry.lock is not consistent with pyproject.toml
- GitHub Actions poetry lockfile drift
- faultline explain poetry-lockfile-drift
- Python poetry lockfile drift


---

*Generated from [playbooks/bundled/log/build/poetry-lockfile-drift.yaml](../../../playbooks/bundled/log/build/poetry-lockfile-drift.yaml). Do not edit directly — run `make docs-generate`.*
