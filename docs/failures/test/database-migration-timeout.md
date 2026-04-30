# Database migration timeout or lock contention

**Playbook ID:** `database-migration-timeout`
**Category:** test
**Severity:** high
**Tags:** `database`, `migration`, `timeout`, `flyway`, `liquibase`, `alembic`, `test-setup`

## What this failure means

Database migration failed to complete within the timeout window, usually due to lock contention or expensive DDL operations during test suite initialization.

## Common log signals

```text
Timeout waiting for exclusive lock
Migration.*failed
Flyway validation failed
liquibase: locked
alembic.*cannot upgrade
database.*migration.*timeout
migration.*timed out
```

## Diagnosis

Integration test suites run database migrations (via Flyway, Liquibase, Alembic, or manual scripts) as a setup step. If a migration acquires an exclusive lock on a table (e.g., for schema changes or index creation) and can't complete within the configured timeout, the test run fails before any tests execute.

This is not a connectivity failure — the database is accessible; rather, the migration process itself is blocked waiting for a lock to be released.

Common causes:
- Large index creation on a populated table holds an exclusive lock longer than the timeout
- Concurrent migrations from multiple test processes competing for the same lock
- Transaction from a previous test still open, holding locks
- Database statistics or query planner issues causing slow DDL operations
- Migration fails on first attempt and retries exhaust the timeout

## Fix steps

1. Check which migration is timing out — it will be named in the error (e.g., `V013__Create_large_index.sql`).

2. Increase the migration timeout if the operation is legitimately expensive:

   ```bash
   # For Flyway
   export FLYWAY_CONNECT_RETRIES=3
   export FLYWAY_LOCK_RETRY_COUNT=50
   
   # For Liquibase
   export LIQUIBASE_TIMEOUT=300  # seconds
   
   # For Alembic
   # In alembic.ini: sqlalchemy.pool_timeout = 300
   ```

3. Optimize the migration query to avoid full table locks:

   ```sql
   -- Good: create index without exclusive lock (if supported by DB)
   CREATE INDEX CONCURRENTLY idx_name ON table_name(column);
   
   -- Avoid: operations that block writes for long periods
   ALTER TABLE large_table ADD COLUMN new_col TYPE DEFAULT value;
   ```

4. Clear any stale locks from previous test runs:

   ```sql
   -- PostgreSQL
   SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'idle in transaction';
   ```

5. Run migrations serially instead of in parallel during test setup:

   ```yaml
   # Disable matrix job isolation that causes concurrent migrations
   parallelism: 1
   ```

## Validation

- Re-run the test suite with increased timeout.
- Confirm the migration completes without timeout errors.
- Verify at least one test executes successfully after migration.

## Likely files to inspect

- `db/migrations/*.sql`
- `src/main/resources/db/migration/*.sql`
- `alembic/versions/*.py`
- `flyway.conf`
- `liquibase.properties`


## Run Faultline

```bash
faultline analyze build.log
faultline explain database-migration-timeout
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Database migration timeout or lock contention
- Test: database migration timeout or lock contention
- Timeout waiting for exclusive lock
- faultline explain database-migration-timeout


---

*Generated from [playbooks/bundled/log/test/database-migration-timeout.yaml](../../../playbooks/bundled/log/test/database-migration-timeout.yaml). Do not edit directly — run `make docs-generate`.*
