# Missing test fixture or seed data

**Playbook ID:** `missing-test-fixture`
**Category:** test
**Severity:** high
**Tags:** `test`, `fixture`, `seed`, `database`, `missing-data`

## What this failure means

A test failed because it could not load a required fixture file, seed record, or test helper resource that was expected to exist before the test ran.

## Common log signals

```text
fixture not found
missing fixture
fixture file missing
no such file or directory
FileNotFoundError
FactoryBot
factory_bot
factory not registered
```

## Diagnosis

A test failed because it could not load a required fixture file, seed record, or test helper resource that was expected to exist before the test ran.

## Fix steps

1. Run the database seed step before the test suite:
   `rails db:seed`, `npm run db:seed`, `python manage.py loaddata`, or the project equivalent.
2. Check if the fixture file was recently renamed or deleted:

   ```bash
   git log --diff-filter=D --name-only --pretty=format: -- testdata/ fixtures/ | grep -v "^$" | head -20
   ```

3. Verify the fixture path is committed and not gitignored:

   ```bash
   git ls-files testdata/
   git ls-files fixtures/
   ```

4. For factory errors: confirm the factory definition matches the current model schema. Run pending migrations before seeding:
   `rails db:migrate`, `python manage.py migrate`.
5. Update the fixture file or factory definition to match the current data model, then re-run migrations and seeds before the test suite.

## Validation

- Run the seed step locally and confirm it exits 0.
- Re-run the full test suite and confirm the fixture-not-found errors are gone.

## Likely files to inspect

- `testdata/`
- `fixtures/`
- `.github/workflows/*.yml`
- `migrations/`


## Run Faultline

```bash
faultline analyze build.log
faultline explain missing-test-fixture
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Missing test fixture or seed data
- Test: missing test fixture or seed data
- ActiveRecord::RecordNotFound
- faultline explain missing-test-fixture


---

*Generated from [playbooks/bundled/log/test/missing-test-fixture.yaml](../../../playbooks/bundled/log/test/missing-test-fixture.yaml). Do not edit directly — run `make docs-generate`.*
