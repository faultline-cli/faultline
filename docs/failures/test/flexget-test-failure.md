# FlexGet test suite failure

**Playbook ID:** `flexget-test-failure`
**Category:** test
**Severity:** high
**Tags:** `flexget`, `python`, `unittest`, `api`, `torrent`

## What this failure means

One or more test cases in the FlexGet plugin/media-manager test suite failed.
FlexGet uses Python unittest and covers its REST API, plugin system, and feed
processing pipeline. Failures typically indicate a broken API contract, a
missing dependency, or a database migration issue.

## Common log signals

```text
tests.test_api.TestServerAPI
tests.test_api.TestTaskAPI
tests.test_abort.TestAbort
flexget.plugins.cli.perf_tests
tests.test_assume_quality.TestAssumeQuality
```

## Diagnosis

The FlexGet test suite organises tests by module under `tests/`:

- `tests.test_api.TestServerAPI` — REST API lifecycle (config, pid, reload,
  shutdown, version). Failures here often mean the embedded Flask server did
  not start or required auth tokens are missing.
- `tests.test_api.TestTaskAPI` — CRUD operations on FlexGet tasks via the API.
- `tests.test_abort.TestAbort` — abort signal handling for stuck tasks.
- `flexget.plugins.cli.perf_tests` — performance benchmarks run as unit tests;
  failures can indicate an import error or missing optional dependency.

Common root causes across the suite:

- **Database schema out of date** — `alembic upgrade head` was not run after
  a migration was added.
- **Missing optional dependency** — a plugin imports a library not present in
  the CI virtualenv (e.g., `deluge`, `rtorrent`).
- **Port conflict** — the test server tries to bind a fixed port already in use.
- **Config fixture stale** — a test config YAML expects a plugin that was
  renamed or removed.

## Fix steps

1. Upgrade the database schema before running tests:
   ```bash
   flexget db upgrade
   ```
2. Install all development dependencies:
   ```bash
   pip install -e ".[dev]"
   ```
3. Run only the failing module to isolate the failure:
   ```bash
   python -m pytest tests/test_api.py -v
   python -m pytest tests/test_abort.py -v
   ```
4. Check for import errors in the plugins:
   ```bash
   python -c "from flexget.plugins.cli.perf_tests import cli_perf_test"
   ```
5. Ensure no other process holds the test server port:
   ```bash
   lsof -i :5050
   ```

## Validation

- `python -m pytest tests/test_api.py tests/test_abort.py -v` all pass.
- `flexget check` reports no schema errors.

## Likely files to inspect

- `flexget/api/app.py`
- `flexget/api/tasks.py`
- `tests/test_api.py`
- `tests/test_abort.py`
- `setup.py`


## Run Faultline

```bash
faultline analyze build.log
faultline explain flexget-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- FlexGet test suite failure
- Test: flexget test suite failure
- tests.test_assume_quality.TestAssumeQuality
- faultline explain flexget-test-failure
- Python flexget test suite failure


---

*Generated from [playbooks/bundled/log/test/flexget-test-failure.yaml](../../playbooks/bundled/log/test/flexget-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
