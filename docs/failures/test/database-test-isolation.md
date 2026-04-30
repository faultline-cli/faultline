# Test database state pollution between tests

**Playbook ID:** `database-test-isolation`
**Category:** test
**Severity:** medium
**Tags:** `test`, `database`, `isolation`, `transaction`, `state`

## What this failure means

Tests are polluting a shared database with committed data that later tests
did not expect to find. Tests pass in isolation but fail when the full suite
runs, or fail only in certain orderings.

## Common log signals

```text
duplicate key value
UniqueConstraintViolation
IntegrityError
duplicate entry
expected count to be
PG::UniqueViolation
SQLSTATE 23505
violates unique constraint
```

## Diagnosis

Each test is leaving committed rows, modified schema state, or other database
changes that subsequent tests encounter as unexpected pre-existing data.

Typical symptoms:

- `already exists` or `duplicate key` errors in tests that insert seed data.
- Counts or queries that return more rows than expected.
- Tests pass individually but fail together or in a different order.

## Fix steps

1. Wrap each test in a database transaction that is rolled back after the
   test:

   - **Rails**: `use_transactional_fixtures = true` (default in RSpec Rails).
   - **Go (sqlx/pgx)**: begin a transaction in `TestMain` or `t.Cleanup`,
     pass it to the test, and roll back after:

     ```go
     tx, _ := db.BeginTx(ctx, nil)
     defer tx.Rollback()
     ```

   - **Django**: `TestCase` wraps each test in a transaction automatically.
     Use `TransactionTestCase` only when testing transaction-specific
     behaviour.

2. For tests that cannot use transactions (e.g., testing COMMIT behaviour),
   use a dedicated test database and truncate all tables in a setup/teardown
   hook:

   ```sql
   TRUNCATE TABLE users, orders, events RESTART IDENTITY CASCADE;
   ```

3. Check for tests that use `TRUNCATE` or `DELETE` on shared tables and
   commit — these affect concurrent parallel tests. Namespace test data
   with a unique run ID or isolate these tests in a separate database.

4. Confirm the database is seeded once before the test suite and not
   accumulating inserts across test runs. Use a per-test factory (e.g.,
   `factory_boy`, `FactoryBot`, Go test helpers) that inserts and cleans
   up its own data.

5. Run tests in random order to surface all implicit ordering dependencies:

   ```bash
   pytest --randomly-seed=<N>
   go test -shuffle=on ./...
   ```

## Validation

- Run the full suite multiple times in random order and confirm zero
  failures caused by unexpected pre-existing data.
- Confirm each test asserts only on data it created, not on total row counts.

## Likely files to inspect

- `docker-compose.yml`
- `.github/workflows/*.yml`
- `migrations/`
- `testdata/`


## Run Faultline

```bash
faultline analyze build.log
faultline explain database-test-isolation
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Test database state pollution between tests
- Test: test database state pollution between tests
- violates unique constraint
- faultline explain database-test-isolation


---

*Generated from [playbooks/bundled/log/test/database-test-isolation.yaml](../../../playbooks/bundled/log/test/database-test-isolation.yaml). Do not edit directly — run `make docs-generate`.*
