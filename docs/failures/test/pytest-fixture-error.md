# pytest fixture setup or teardown failure

**Playbook ID:** `pytest-fixture-error`
**Category:** test
**Severity:** medium
**Tags:** `pytest`, `python`, `fixture`, `setup`, `teardown`, `conftest`

## What this failure means

A pytest fixture failed during setup or teardown, causing the associated test to be marked as `ERROR` rather than `FAILED`. The test body never ran, or cleanup after the test raised an exception.

## Common log signals

```text
ERROR at setup of
ERROR at teardown of
fixture not found
fixture error
ScopeMismatch
could not import conftest
ERROR collecting
E   fixture
```

## Diagnosis

A pytest fixture failed during setup or teardown, causing the associated test to be marked as `ERROR` rather than `FAILED`. The test body never ran, or cleanup after the test raised an exception.

## Fix steps

1. Read the full `ERROR at setup of <test_name>` traceback. The relevant line is usually in the fixture body, not the test.
2. For `fixture not found`, check the fixture name spelling and confirm it is defined in `conftest.py` or imported in the same module.
3. For `ScopeMismatch`, widen the dependent fixture scope or pass the needed value directly.
4. Reproduce fixture setup locally with `pytest -x --setup-show tests/path/to_test.py`.
5. For teardown failures, make sure `yield` fixtures wrap cleanup in `try/finally`.

## Validation

- `pytest -x --setup-show tests/` completes without the original fixture error.
- `pytest --co -q tests/` confirms pytest can still collect the expected suite.

## Likely files to inspect

- `conftest.py`
- `tests/conftest.py`
- `pytest.ini`
- `pyproject.toml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain pytest-fixture-error
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- pytest fixture setup or teardown failure
- Test: pytest fixture setup or teardown failure
- PytestWarning: PytestUnraisableExceptionWarning
- faultline explain pytest-fixture-error
- Python pytest fixture setup or teardown failure


---

*Generated from [playbooks/bundled/log/test/pytest-fixture-error.yaml](../../../playbooks/bundled/log/test/pytest-fixture-error.yaml). Do not edit directly — run `make docs-generate`.*
