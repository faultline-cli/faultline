# pip installation functional test failure

**Playbook ID:** `pip-install-test-failure`
**Category:** test
**Severity:** high
**Tags:** `pip`, `python`, `packaging`, `install`, `virtualenv`, `functional-test`

## What this failure means

One or more pip functional tests for package installation failed.
The pip test suite covers install from local directories, user-site installs,
symlink handling, and conflict detection. Failures indicate a regression in pip's
install logic or an environment issue in the CI virtualenv.

## Common log signals

```text
test_install_from_local_directory_with_symlinks_to_directories
Tests_UserSite.test_install_user_in_global_virtualenv
test_install_user_in_global_virtualenv_with_conflict_fails
test_freeze_basic
test_freeze_user
```

## Diagnosis

The pip functional test suite exercises end-to-end install scenarios:

- `test_install_from_local_directory_with_symlinks_to_directories` — verifies
  that pip correctly follows symlinks when installing from a local checkout.
  Fails if the OS does not support symlinks (e.g., NTFS without privilege) or
  if pip's path normalisation changed.
- `Tests_UserSite.test_install_user_in_global_virtualenv_with_conflict_fails` —
  confirms that `pip install --user` inside a global virtualenv raises an error
  when a conflicting package version already exists. Fails if the conflict
  detection logic was loosened or if the test's virtualenv setup is broken.
- `test_multiple_search` / `test_search` — PyPI index search tests. Fail if
  the test uses a live PyPI endpoint that is unavailable, or if the search API
  response format changed.

Common root causes:

- **Python version skew** — the test was written for Python 2 but is running
  under Python 3 or vice versa, causing assertion mismatches.
- **Virtualenv not isolated** — a stale system `site-packages` bleeds into
  the test environment, causing false conflict detection or false passes.
- **Symlink privilege** — on Windows CI the unprivileged runner cannot create
  symlinks, so symlink-dependent tests are skipped or fail.
- **Wheel cache pollution** — a cached wheel from a previous run satisfies an
  install without exercising the path being tested.

## Fix steps

1. Reproduce in a clean virtualenv:
   ```bash
   python -m venv /tmp/pip-test-env
   source /tmp/pip-test-env/bin/activate
   pip install -e ".[testing]"
   ```
2. Run only the failing test:
   ```bash
   python -m pytest tests/functional/test_install.py \
     -k "test_install_from_local_directory_with_symlinks" -v
   python -m pytest tests/functional/test_usersite.py \
     -k "test_install_user_in_global_virtualenv_with_conflict" -v
   ```
3. Confirm the test environment has no pre-existing conflicting packages:
   ```bash
   pip list
   ```
4. Clear the wheel cache before re-running:
   ```bash
   pip cache purge
   python -m pytest tests/functional/ -v --no-header
   ```

## Validation

- `python -m pytest tests/functional/test_install.py -v` all pass.
- `python -m pytest tests/functional/test_usersite.py -v` all pass.

## Likely files to inspect

- `src/pip/_internal/commands/install.py`
- `src/pip/_internal/req/req_install.py`
- `tests/functional/test_install.py`
- `tests/functional/test_usersite.py`
- `tests/functional/test_freeze.py`


## Run Faultline

```bash
faultline analyze build.log
faultline explain pip-install-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- pip installation functional test failure
- Test: pip installation functional test failure
- test_install_from_local_directory_with_symlinks_to_directories
- faultline explain pip-install-test-failure
- Python pip installation functional test failure


---

*Generated from [playbooks/bundled/log/test/pip-install-test-failure.yaml](../../playbooks/bundled/log/test/pip-install-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
