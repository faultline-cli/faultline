# PostgreSQL migration failed due to unsupported type or constraint syntax

**Playbook ID:** `postgres-version-incompatible-migration`
**Category:** test
**Severity:** medium
**Tags:** `postgres`, `migration`, `json`, `not-valid`, `version`, `test-setup`, `database`

## What this failure means

A database migration failed because the PostgreSQL version in CI does not
support a type or constraint syntax used in the migration (e.g. the built-in
JSON type requires PostgreSQL 9.2+, and NOT VALID constraints require 9.1+).
Subsequent migration commands are skipped because the transaction is aborted.

## Common log signals

```text
type "json" does not exist
CHECK constraints cannot be marked NOT VALID
```

## Diagnosis

The test suite runs migrations against a PostgreSQL instance that is older
than the migration requires. The first failing statement aborts the
transaction and causes all remaining migration commands to emit "current
transaction is aborted" errors.

Common triggers:

- `type "json" does not exist` — the migration uses the built-in JSON type, which
  was added in PostgreSQL 9.2. Any version below 9.2 will fail on this.
- `CHECK constraints cannot be marked NOT VALID` — the NOT VALID modifier for
  CHECK constraints was added in PostgreSQL 9.1.
- The CI PostgreSQL service container is pinned to an old image tag (e.g.
  `postgres:9.0` or unversioned `postgres:latest` that resolved to an old image
  when first written).

## Fix steps

1. Note the failing migration statement (e.g. `type "json" does not exist`).
2. Check the PostgreSQL version running in CI:
   ```bash
   psql --version
   # or in CI logs: look for "PostgreSQL X.Y" in the service startup output
   ```
3. Update the CI service to PostgreSQL 9.2 or later (13+ is recommended):
   ```yaml
   # GitHub Actions
   services:
     postgres:
       image: postgres:14
   # Travis CI
   addons:
     postgresql: "14"
   ```
4. If the project must support older PostgreSQL, replace the JSON type with
   TEXT and validate format in application code, or use the `hstore` extension
   for key-value data.
5. Re-run migrations after the version bump:
   ```bash
   alembic upgrade head
   # or
   python manage.py migrate
   ```

## Validation

- Confirm the PostgreSQL version in CI is 9.2 or later.
- Re-run the test suite. No `type "json" does not exist` or
  `current transaction is aborted` errors should appear.
- `make test` passes cleanly.

## Likely files to inspect

- `migrations/`
- `alembic/versions/`
- `.travis.yml`
- `.github/workflows/`
- `docker-compose.yml`
- `docker-compose.test.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain postgres-version-incompatible-migration
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- PostgreSQL migration failed due to unsupported type or constraint syntax
- Test: postgresql migration failed due to unsupported type or constraint syntax
- CHECK constraints cannot be marked NOT VALID
- faultline explain postgres-version-incompatible-migration


---

*Generated from [playbooks/bundled/log/test/postgres-version-incompatible-migration.yaml](../../../playbooks/bundled/log/test/postgres-version-incompatible-migration.yaml). Do not edit directly — run `make docs-generate`.*
