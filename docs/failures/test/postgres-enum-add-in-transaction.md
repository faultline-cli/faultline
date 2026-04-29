# PostgreSQL ALTER TYPE ADD VALUE cannot run inside a transaction block

**Playbook ID:** `postgres-enum-add-in-transaction`
**Category:** test
**Severity:** medium
**Tags:** `postgres`, `migration`, `enum`, `alter-type`, `transaction`, `alembic`, `django`, `test-setup`

## What this failure means

A database migration that adds a value to a PostgreSQL ENUM type fails because
PostgreSQL does not permit ALTER TYPE ... ADD VALUE inside a transaction block.
Migration frameworks that wrap each migration in a transaction (Alembic, Django,
Flyway) hit this restriction by default.

## Common log signals

```text
ALTER TYPE ... ADD cannot run inside a transaction block
ALTER TYPE ... ADD VALUE cannot run inside a transaction block
```

## Diagnosis

PostgreSQL prohibits adding a new value to an existing ENUM type inside a
transaction. Migration frameworks wrap every migration in a `BEGIN`/`COMMIT`
block for atomicity. When a migration calls:

```sql
ALTER TYPE my_enum ADD VALUE 'new_value';
```

PostgreSQL raises:

```
ALTER TYPE ... ADD cannot run inside a transaction block
```

The migration is immediately aborted. Any test run that depends on migrations
being applied will fail with "current transaction is aborted" errors on all
subsequent statements within the same migration.

Note: some migration files also contain raw SQL syntax errors (e.g.,
`syntax error at or near "REMOVE"`) from typos or DB-dialect mismatches,
which produce a separate but related abort cascade.

## Fix steps

### Option A — Split the migration (recommended)

Move the `ALTER TYPE ... ADD VALUE` statement into its own migration file so
the framework can run it outside the normal transaction:

- **Alembic**: use `connection.execute()` with `COMMIT` before the ALTER and
  restart a new transaction after, or declare the migration as
  `transaction_per_migration = False` and manage the transaction manually:

  ```python
  # alembic migration
  def upgrade():
      op.execute("COMMIT")
      op.execute("ALTER TYPE my_enum ADD VALUE IF NOT EXISTS 'new_value'")
      op.execute("BEGIN")
  ```

- **Django**: use `RunSQL` with `atomic=False`:

  ```python
  migrations.RunSQL(
      "ALTER TYPE my_enum ADD VALUE 'new_value'",
      atomic=False,
  )
  ```

### Option B — Replace ENUM with VARCHAR (simpler for frequent changes)

Convert the column to VARCHAR and validate values in application code or with
a CHECK constraint. This avoids the ALTER TYPE restriction entirely.

### Option C — Use `IF NOT EXISTS` to make repeated runs safe

```sql
ALTER TYPE my_enum ADD VALUE IF NOT EXISTS 'new_value';
```

This requires PostgreSQL 9.3+. It does not remove the transaction restriction
but prevents the failure when the value already exists.

## Validation

- Re-run the migration: `alembic upgrade head` or `python manage.py migrate`.
- Confirm no `ALTER TYPE ... ADD cannot run inside a transaction block` error.
- Re-run the test suite. No `current transaction is aborted` cascade errors.

## Likely files to inspect

- `migrations/`
- `alembic/versions/`
- `*/migrations/00*.py`
- `.github/workflows/`


## Run Faultline

```bash
faultline analyze build.log
faultline explain postgres-enum-add-in-transaction
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- PostgreSQL ALTER TYPE ADD VALUE cannot run inside a transaction block
- Test: postgresql alter type add value cannot run inside a transaction block
- ALTER TYPE ... ADD VALUE cannot run inside a transaction block
- faultline explain postgres-enum-add-in-transaction


---

*Generated from [playbooks/bundled/log/test/postgres-enum-add-in-transaction.yaml](../../playbooks/bundled/log/test/postgres-enum-add-in-transaction.yaml). Do not edit directly — run `make docs-generate`.*
