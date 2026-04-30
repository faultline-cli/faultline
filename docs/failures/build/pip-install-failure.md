# pip package install failure

**Playbook ID:** `pip-install-failure`
**Category:** build
**Severity:** high
**Tags:** `python`, `pip`, `install`, `dependencies`, `build`

## What this failure means

`pip install` could not satisfy one or more package requirements. The package may not exist on PyPI, the version constraint is too restrictive, or a C extension failed to compile.

## Common log signals

```text
ERROR: Could not find a version that satisfies the requirement
No matching distribution found for
pip._internal
ResolutionImpossible
Could not build wheels for
error: command 'gcc' failed
Ignored the following versions that require a different python version
requires a different python version
```

## Diagnosis

`pip install` could not satisfy one or more package requirements. The package may not exist on PyPI, the version constraint is too restrictive, or a C extension failed to compile.

## Fix steps

1. Run `python -m pip install -r requirements.txt -v` in the same
   interpreter or virtual environment that CI uses, so the resolver output
   matches the failing job.
2. For `ResolutionImpossible`: relax or align version constraints in `requirements.txt`.
3. For `Failed building wheel`: install the required system libraries (e.g. `libpq-dev`, `build-essential`).
4. For private indexes: set `PIP_EXTRA_INDEX_URL` or add `--extra-index-url` to the install command.
5. Use `pip-compile` from pip-tools to generate a locked `requirements.txt`.

## Validation

- Re-run the local reproduction command after the fix.
- `python -m pip check` passes in the same environment.
- `python -m pip install -r requirements.txt` completes successfully.

## Likely files to inspect

- `requirements.txt`
- `requirements-dev.txt`
- `setup.cfg`
- `pyproject.toml`
- `Pipfile`


## Run Faultline

```bash
faultline analyze build.log
faultline explain pip-install-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- pip package install failure
- Build: pip package install failure
- Ignored the following versions that require a different python version
- GitHub Actions pip package install failure
- faultline explain pip-install-failure
- Python pip package install failure


---

*Generated from [playbooks/bundled/log/build/pip-install-failure.yaml](../../playbooks/bundled/log/build/pip-install-failure.yaml). Do not edit directly — run `make docs-generate`.*
