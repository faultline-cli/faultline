# Python or pip command not found

**Playbook ID:** `python-command-not-found`
**Category:** build
**Severity:** medium
**Tags:** `python`, `pip`, `executable`, `missing`, `environment`

## What this failure means

The `python` or `pip` command is not available in the CI environment's PATH.

## Common log signals

```text
python: command not found
python3: command not found
pip: command not found
pip3: command not found
bash: python: command not found
bash: pip: command not found
/usr/bin/env: 'python': No such file or directory
/usr/bin/env: 'python3': No such file or directory
```

## Diagnosis

The CI job tried to run `python`, `python3`, `pip`, or `pip3` but the command was not found. This usually happens when:

- The base image does not include Python (e.g., `node:alpine`, `debian:slim` without python3).
- Python is installed but not in the default PATH.
- The wrong Python version is referenced (e.g., `python` vs `python3`).
- A virtual environment is activated but the interpreter is missing.

This is distinct from a generic missing executable because Python is a fundamental runtime for many build and test workflows.

## Fix steps

1. **Install Python** in the CI environment:

   **Ubuntu/Debian:**
   ```bash
   apt-get update && apt-get install -y python3 python3-pip
   ```

   **Alpine:**
   ```bash
   apk add --no-cache python3 py3-pip
   ```

   **CentOS/RHEL:**
   ```bash
   yum install -y python3 python3-pip
   ```

2. **Check the available Python executable** and adjust the command:
   ```bash
   which python3
   python3 --version
   ```

3. **Use `python3` instead of `python`** if the base image only provides `python3`.

4. **In GitHub Actions**, use the `actions/setup-python` action to ensure a specific Python version:
   ```yaml
   - uses: actions/setup-python@v5
     with:
       python-version: '3.12'
   ```

5. **If using a Docker image**, switch to an image that includes Python, such as `python:3.12-slim` or `node:20` (includes Python for node-gyp).

6. **For pip**, ensure it is installed with Python or install it separately:
   ```bash
   python3 -m ensurepip --upgrade
   ```

## Validation

- `python3 --version` prints a version without "command not found".
- `pip3 --version` or `python3 -m pip --version` shows pip is available.
- Re‑run the failing step and confirm Python/pip commands succeed.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `Dockerfile`
- `requirements.txt`
- `pyproject.toml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain python-command-not-found
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Python or pip command not found
- Build: python or pip command not found
- /usr/bin/env: 'python3': No such file or directory
- GitHub Actions python or pip command not found
- faultline explain python-command-not-found
- Python python or pip command not found


---

*Generated from [playbooks/bundled/log/build/python-command-not-found.yaml](../../../playbooks/bundled/log/build/python-command-not-found.yaml). Do not edit directly — run `make docs-generate`.*
