# Timezone differences causing test failures

**Playbook ID:** `timezone-diff`
**Category:** test
**Severity:** medium
**Tags:** `timezone`, `utc`, `time`, `test`, `date`, `dst`

## What this failure means

A test that compares date or time values fails because the CI runner operates
in a different timezone than the development machine. Code that implicitly
calls `datetime.now()`, `new Date()`, or similar without a timezone argument
returns local time, which differs across environments.

## Common log signals

```text
expected.*UTC
got.*local time
timezone.*mismatch
assertion.*datetime
Expected.*+0000
received.*+0
wrong timezone
AssertionError.*datetime
```

## Diagnosis

CI runners commonly default to UTC or a cloud region timezone. Developer
machines use local timezones. Tests that construct expected values with date
helpers or hardcode timestamps without UTC anchoring will produce different
results on each environment.

Common patterns that cause timezone failures:
- `datetime.now()` or `new Date()` without `timezone.utc` / `Z` suffix
- Date formatting that produces locale-dependent output
- Database timestamps stored and retrieved without explicit timezone handling
- Assertions comparing `strftime` output that includes timezone offset

Check the runner timezone:

```bash
date
echo $TZ
timedatectl show --property=Timezone --value    # Linux systemd
```

## Fix steps

1. Force the CI environment to UTC:

   ```yaml
   # GitHub Actions
   env:
     TZ: UTC

   # GitLab CI
   variables:
     TZ: UTC
   ```

2. In application code, always use UTC-aware datetime objects and convert to
   local time only at the presentation layer:

   ```python
   # Python — always UTC
   from datetime import datetime, timezone
   now = datetime.now(tz=timezone.utc)
   ```

   ```javascript
   // JavaScript — use ISO strings or UTC methods
   const now = new Date().toISOString();             // "2025-01-01T12:00:00.000Z"
   const timestamp = Date.UTC(2025, 0, 1, 12, 0, 0); // explicit UTC epoch
   ```

3. In tests, freeze or mock the clock with a fixed UTC time rather than
   calling the real system clock:

   ```python
   # Python with freezegun
   from freezegun import freeze_time
   @freeze_time("2025-01-01 12:00:00")
   def test_something():
       ...

   # pytest-mock
   mocker.patch("mymodule.datetime").now.return_value = fixed_utc_dt
   ```

   ```javascript
   // Jest
   jest.useFakeTimers().setSystemTime(new Date("2025-01-01T12:00:00Z"));
   ```

4. Store and compare all database timestamps in UTC; use `TIMESTAMPTZ` in
   PostgreSQL and UTC-mode drivers in MySQL.

5. If tests must exercise timezone conversions, parameterize the timezone
   explicitly rather than relying on `TZ` environment state.

## Validation

- Set `TZ=UTC` locally and re-run the failing tests to reproduce them.
- After applying fixes, run the suite with several different `TZ` values:
  `TZ=America/New_York`, `TZ=Asia/Kolkata`, `TZ=Australia/Sydney`.
- Confirm all assertions pass regardless of system timezone.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain timezone-diff
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Timezone differences causing test failures
- Test: timezone differences causing test failures
- AssertionError.*datetime
- faultline explain timezone-diff


---

*Generated from [playbooks/bundled/log/test/timezone-diff.yaml](../../../playbooks/bundled/log/test/timezone-diff.yaml). Do not edit directly — run `make docs-generate`.*
