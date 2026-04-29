# pytest collected zero tests

**Playbook ID:** `pytest-no-tests`
**Category:** test
**Severity:** medium
**Tags:** `pytest`, `python`, `test`, `collection`, `discovery`

## What this failure means

pytest exited with code 5 because it collected zero tests. No tests matched the given path, filter expression, or marker constraint.

## Common log signals

```text
collected 0 items
no tests ran
exit code: 5
exit code 5
ExitCode.NO_TESTS_COLLECTED
no tests were run
ERROR: not found:
collected 0 items / 1 error
```

## Diagnosis

pytest reserves exit code 5 for the specific condition where it starts successfully, finishes collection, and finds nothing to run. This is distinct from a collection error (exit code 2) or a test failure (exit code 1).

Common causes:

- The test path passed to pytest does not match any test files or functions.
- A `-k` expression or `-m` marker filter excludes every collected item.
- Test files exist but do not match the configured `python_files` or `python_functions` glob patterns.
- The working directory is wrong so pytest cannot discover the test root.
- `testpaths` in `pytest.ini` or `pyproject.toml` points at a non-existent directory, or the test directory was not checked out by a shallow clone.

## Fix steps

1. Run collection in dry-run mode to see what pytest discovers:

   ```bash
   pytest --collect-only
   ```

2. If using a path argument, confirm the path exists and contains test files prefixed `test_` or suffixed `_test.py`.

3. Check for an overly restrictive `-k` or `-m` filter in the `pytest` invocation or in `addopts` inside `pytest.ini` / `pyproject.toml`:

   ```ini
   # pyproject.toml — watch for addopts entries that exclude by default
   [tool.pytest.ini_options]
   addopts = "-m not slow"
   ```

4. Verify `testpaths` points at the correct directory:

   ```bash
   python -m pytest --co -q 2>&1 | head -20
   ```

5. If the test directory requires generation or exists only on certain branches, ensure the CI job checks it out before running pytest.

## Validation

- `pytest --collect-only` returns at least one test item with exit code 0.
- `pytest` exits with code 0 (tests ran) rather than code 5 (nothing collected).

## Likely files to inspect

- `pytest.ini`
- `pyproject.toml`
- `setup.cfg`
- `conftest.py`


## Run Faultline

```bash
faultline analyze build.log
faultline explain pytest-no-tests
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- pytest collected zero tests
- Test: pytest collected zero tests
- ExitCode.NO_TESTS_COLLECTED
- faultline explain pytest-no-tests
- Python pytest collected zero tests


---

*Generated from [playbooks/bundled/log/test/pytest-no-tests.yaml](../../playbooks/bundled/log/test/pytest-no-tests.yaml). Do not edit directly — run `make docs-generate`.*
