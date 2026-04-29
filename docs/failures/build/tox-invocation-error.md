# tox environment command failed (InvocationError)

**Playbook ID:** `tox-invocation-error`
**Category:** build
**Severity:** high
**Tags:** `python`, `tox`, `test-runner`, `virtualenv`, `ci`

## What this failure means

tox ran a command inside a virtual environment and the command exited with a non-zero status. The error line begins with `InvocationError:` and shows the full path to the failing executable.

## Common log signals

```text
InvocationError:
```

## Diagnosis

tox wraps test and tool commands so that they run inside isolated virtual environments. When any command in a tox environment exits non-zero, tox reports:

```
InvocationError: '/path/to/.tox/<env>/bin/<command> [args]'
<env>: commands failed
```

The `InvocationError` line itself only tells you *which* command failed, not *why*. The actual failure is always in the output printed before this line. Common root causes:

- Test assertions failed (pytest, nosetests, behave, unittest).
- A dependency is missing or version-incompatible inside the tox virtualenv.
- The test command was killed due to timeout or OOM inside the CI runner.
- A configuration issue in `tox.ini` or `setup.cfg` caused the wrong command to run.
- An environment variable required by the test suite was not set in the tox context.

## Fix steps

1. Scroll up past the `InvocationError:` line to read the actual failure output — that is where the real error is.

2. Reproduce locally using the exact tox environment that failed:

   ```bash
   tox -e <env-name> -v
   ```

   Replace `<env-name>` with the environment shown before `: commands failed` (e.g. `py3.3-django1.6-postgresql-locmem`).

3. If the environment does not exist on your machine, install the required Python version and recreate it:

   ```bash
   tox -e <env-name> --recreate
   ```

4. For dependency failures inside tox, check `tox.ini`:

   ```bash
   tox -e <env-name> --listenvs-all
   cat tox.ini | grep -A20 '\[testenv'
   ```

5. Check if the tox virtualenv itself is stale or corrupt:

   ```bash
   tox -e <env-name> --recreate
   ```

6. For missing environment variables, ensure CI sets them in the job definition and that `tox.ini` passes them through with `passenv`.

## Validation

- `tox -e <env-name>` exits 0.
- All tests in the affected environment pass.
- No `InvocationError` line appears in the tox output.

## Likely files to inspect

- `tox.ini`
- `setup.cfg`
- `pyproject.toml`
- `.travis.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain tox-invocation-error
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- tox environment command failed (InvocationError)
- Build: tox environment command failed (invocationerror)
- InvocationError:
- GitHub Actions tox environment command failed (invocationerror)
- faultline explain tox-invocation-error
- Python tox environment command failed (invocationerror)


---

*Generated from [playbooks/bundled/log/build/tox-invocation-error.yaml](../../playbooks/bundled/log/build/tox-invocation-error.yaml). Do not edit directly — run `make docs-generate`.*
