# Non-deterministic failure due to unfixed random seed

**Playbook ID:** `random-seed-not-fixed`
**Category:** test
**Severity:** low
**Tags:** `random`, `seed`, `nondeterminism`, `flaky`, `shuffle`, `order`, `hash`

## What this failure means

Tests pass locally but fail intermittently in CI because they depend on
random ordering or random data without a fixed seed, producing different
results on each run.

## Common log signals

```text
random seed
seed:
rand.Seed
PYTHONHASHSEED
random_state
shuffle
randomized
non-deterministic
```

## Diagnosis

Random seeds affect:
- Test execution order (pytest-randomly, go test -shuffle, Jest --randomize)
- Random data in fuzz or property-based tests
- Hash map iteration order (Python dict, Go map, HashMap in Java)
- Shuffled training data in ML pipelines

When a seed is not fixed, a test suite passes most runs but fails on some
orderings that expose hidden test coupling (test A mutates shared state, and
test B depends on the original state — only fails when B runs before A).

Check for implicit ordering dependencies:

```bash
# Go: run with a random shuffle to expose order dependencies
go test -shuffle=on ./...

# Pytest: run with a specific failing seed
pytest --randomly-seed=<N>

# Jest: show seed in output
jest --randomize
```

## Fix steps

1. **Reproduce with the failing seed**:

   ```bash
   # Pytest (output shows "Using --randomly-seed=<N>")
   pytest --randomly-seed=12345

   # Go
   go test -shuffle=on -count=1 ./...   # outputs seed on failure

   # Jest
   jest --randomize --randomizeSeed=12345
   ```

2. **Fix test isolation** — the real fix is removing hidden state coupling:

   ```python
   # BAD: test modifies global state
   def test_a():
       app.config['DEBUG'] = True

   # GOOD: use setup/teardown or fixtures
   @pytest.fixture
   def app_config():
       old = app.config.copy()
       yield app.config
       app.config.update(old)
   ```

3. **If randomization is intentional**, pin the seed in CI for
   deterministic runs:

   ```yaml
   # GitHub Actions
   env:
     PYTHONHASHSEED: "0"
     RANDOM_SEED: "42"
   ```

   ```go
   // Go test helper
   rand.Seed(42)
   ```

4. **Use explicit sort** instead of relying on map/set iteration order:

   ```python
   # BAD
   for key in my_dict:     # order not guaranteed across Python versions
   # GOOD
   for key in sorted(my_dict):
   ```

## Validation

- Re-run with the specific failing seed and confirm the failure is reproduced.
- Fix the state isolation, then re-run with multiple random seeds.
- Enable shuffle in CI by default so coverage is broad over time.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain random-seed-not-fixed
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Non-deterministic failure due to unfixed random seed
- Test: non-deterministic failure due to unfixed random seed
- order of test cases
- faultline explain random-seed-not-fixed


---

*Generated from [playbooks/bundled/log/test/random-seed-not-fixed.yaml](../../../playbooks/bundled/log/test/random-seed-not-fixed.yaml). Do not edit directly — run `make docs-generate`.*
