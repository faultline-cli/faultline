# Python module not found

**Playbook ID:** `python-module-missing`
**Category:** build
**Severity:** medium
**Tags:** `python`, `pip`, `modules`, `dependencies`

## What this failure means

Python could not import a required module because it is missing from the environment used by the failing CI step or the import path is wrong for the installed package.

## Common log signals

```text
ModuleNotFoundError: No module named
ImportError: No module named
cannot import name
No module named
```

## Diagnosis

Python could not import a required module because it is missing from the environment used by the failing CI step or the import path is wrong for the installed package.

This usually happens when dependency installation was skipped, the job uses a different virtual environment than expected, the package name differs from the importable module name, or the selected Python version is incompatible with the dependency set.

## Fix steps

1. Ensure the dependency install step runs before tests or application startup in the same environment that executes the failing command.
2. Verify the package name in `requirements.txt`, `pyproject.toml`, or the lockfile matches the module import used in code.
3. Confirm CI is activating the intended virtual environment and Python interpreter instead of falling back to the system Python.
4. If the import moved between package versions, pin the dependency to the compatible release and update the import path deliberately.

## Validation

- Re-run the failing step after reinstalling dependencies in the same environment.
- Confirm the import succeeds and `pip check` reports no broken dependency state.

## Likely files to inspect

- `requirements.txt`
- `pyproject.toml`
- `setup.cfg`
- `Pipfile`


## Run Faultline

```bash
faultline analyze build.log
faultline explain python-module-missing
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Python module not found
- Build: python module not found
- ModuleNotFoundError: No module named
- GitHub Actions python module not found
- faultline explain python-module-missing
- Python python module not found


---

*Generated from [playbooks/bundled/log/build/python-module-missing.yaml](../../../playbooks/bundled/log/build/python-module-missing.yaml). Do not edit directly — run `make docs-generate`.*
