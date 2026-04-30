# Test assertion failure with explicit reason message

**Playbook ID:** `test-assertion-with-reason`
**Category:** test
**Severity:** high
**Tags:** `test`, `assertion`, `erlang`, `eunit`, `reason`

## What this failure means

A test case assertion failed with an explicit reason message. The test ran but
the expected condition did not hold, producing a deterministic failure reason
that names the broken expectation.

## Common log signals

```text
SUITE:.*failed
SUITE_.*failed
failed.*Reason:
Reason: expecting
Reason: expected_
init_per_group failed Reason
assertion failed: expected
AssertionError: expected
```

## Diagnosis

Test assertion failures provide explicit reason messages that name what was
expected versus what occurred. Common patterns:

- **Erlang EUnit/Common Test**: `SUITE_NAME:test_function failed ... Reason: expected_value_got_other`
- **Queue/messaging tests**: `expecting_nack_got_ack`, `expected_empty_got_n_messages`
- **State machine tests**: `expected_state_X_got_state_Y`

The `Reason:` message tells you exactly what the test was checking.

Common causes:

- **Logic error** — the implementation changed and now produces different output than the test expects.
- **Test data setup** — the test fixture or seed data does not match the test assumption.
- **Timing or ordering** — concurrent code or queue processing order changed.
- **API change** — a dependency changed its behavior and the code did not adapt.

## Fix steps

1. Read the `Reason:` message carefully — it names both the expected and actual state.
2. Locate the failing test in the source:
   ```bash
   grep -r "expected_.*_got_" src/test/ eunit_SUITE.erl
   ```
3. Run the test in isolation to reproduce:
   ```bash
   erl -name test@localhost -s mymodule test_function
   rebar3 eunit -m mymodule
   ```
4. Add debug output or inspect the intermediate state:
   - Print the actual state before the assertion
   - Check if external dependencies behave as expected
5. Fix either the implementation to match the test, or update the test if the
   new behavior is correct and documented.

## Validation

- Run the failing test again — it should pass.
- Run the full test suite to ensure no regressions.

## Likely files to inspect

- `**/*_SUITE.erl`
- `**/test/**/*.erl`
- `src/test/**/*.erl`
- `rebar.config`


## Run Faultline

```bash
faultline analyze build.log
faultline explain test-assertion-with-reason
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Test assertion failure with explicit reason message
- Test: test assertion failure with explicit reason message
- init_per_group failed Reason
- faultline explain test-assertion-with-reason


---

*Generated from [playbooks/bundled/log/test/test-assertion-with-reason.yaml](../../../playbooks/bundled/log/test/test-assertion-with-reason.yaml). Do not edit directly — run `make docs-generate`.*
