# Python virtualenv not activated or interpreter mismatch

**Playbook ID:** `python-virtualenv-not-activated`
**Category:** build
**Severity:** medium
**Tags:** `python`, `virtualenv`, `venv`, `interpreter`, `environment`

## What this failure means

Python code is using a system interpreter or wrong virtualenv instead of the activated one, causing module import failures or version mismatches.

## Common log signals

```text
activate: not found
source.*venv
/usr/bin/python.*No module named
venv/bin/activate
site-packages from system Python
```

## Diagnosis

When a Python virtualenv is not activated or is deactivated mid-build, pip and modules are installed in the wrong location. Common causes:

- The virtualenv activation step (e.g., `source venv/bin/activate`) was skipped or failed silently
- The build runs multiple shell invocations without persisting the activation across steps
- The script uses `#!/usr/bin/env python` which may resolve to the system interpreter, not the virtualenv one
- GitHub Actions or other CI steps start with a fresh shell that needs re-activation

Symptoms include `ModuleNotFoundError` for packages that were installed, or import failures from the wrong Python version.

## Fix steps

1. Create and activate the virtualenv explicitly in your CI build script:

   ```bash
   python3 -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   pip install -r requirements.txt
   python -m pytest
   ```

2. If using GitHub Actions, use `actions/setup-python` to manage the virtualenv:

   ```yaml
   - uses: actions/setup-python@v4
     with:
       python-version: '3.11'
      cache: 'pip'
   - run: pip install -r requirements.txt
   - run: python -m pytest
   ```

3. If running multiple steps, either source the activation in each step or use a single shell script:

   ```yaml
   - run: |
       source venv/bin/activate
       pip install -r requirements.txt
       python -m pytest
   ```

4. Verify the correct Python interpreter is in use:

   ```bash
   which python3
   python3 --version
   ```

## Validation

- `which python3` shows the path inside the virtualenv, not the system Python.
- `python3 -c "import sys; print(sys.prefix)"` shows the virtualenv directory.
- `pip list` shows the packages from `requirements.txt`.
- Re-run the test or build command successfully.

## Likely files to inspect

- `requirements.txt`
- `setup.py`
- `pyproject.toml`
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain python-virtualenv-not-activated
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Python virtualenv not activated or interpreter mismatch
- Build: python virtualenv not activated or interpreter mismatch
- /usr/bin/python.*No module named
- GitHub Actions python virtualenv not activated or interpreter mismatch
- faultline explain python-virtualenv-not-activated
- Python python virtualenv not activated or interpreter mismatch


---

*Generated from [playbooks/bundled/log/build/python-virtualenv-not-activated.yaml](../../playbooks/bundled/log/build/python-virtualenv-not-activated.yaml). Do not edit directly — run `make docs-generate`.*
