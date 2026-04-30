# GitHub Actions environment variable not persisted across steps

**Playbook ID:** `github-actions-env-not-persisted`
**Category:** ci
**Severity:** high
**Tags:** `github-actions`, `env`, `environment-variable`, `set-env`, `GITHUB_ENV`, `deprecated`

## What this failure means

A GitHub Actions step exported an environment variable using the deprecated
`::set-env` workflow command, which is now disabled. The variable was not
visible to subsequent steps. The fix is to write to `$GITHUB_ENV` instead.

## Common log signals

```text
The 'set-env' command is deprecated
set-env' command is disabled
Unable to process command '::set-env'
::set-env name=
Unprocessed element 'set-env'
```

## Diagnosis

GitHub Actions runs each `run:` step in a new shell process. Variables set
with the shell's `export` built-in do **not** persist across steps. To share
a value with subsequent steps, write to the `$GITHUB_ENV` file:

```bash
echo "MY_VAR=value" >> "$GITHUB_ENV"
```

Before this mechanism existed, GitHub Actions supported a workflow command
syntax (`::set-env name=VAR::value`). This command was deprecated in
October 2020 and permanently **disabled** due to a security vulnerability
(it allowed injection from log output into the runner's environment). Any
workflow still using it receives an explicit error:

- `The 'set-env' command is deprecated and will be disabled soon.`
- `Unable to process command '::set-env' successfully.`
- `Error: Unprocessed element 'set-env'`

The downstream symptom is that the variable appears empty in later steps,
causing the following step to fail with a command-not-found, missing-path,
or empty-value error.

## Fix steps

1. Search the workflow file for any use of the `::set-env` syntax:

   ```bash
   grep -r "::set-env" .github/workflows/
   ```

2. Replace each occurrence with the `$GITHUB_ENV` file append:

   ```bash
   # Before (deprecated / disabled):
   echo "::set-env name=MY_VAR::$VALUE"

   # After (correct):
   echo "MY_VAR=$VALUE" >> "$GITHUB_ENV"
   ```

3. If `export VAR=value` is used to set a variable that needs to be visible
   in a later step, replace it:

   ```bash
   # Before (does not persist across steps):
   export MY_VAR=$(git describe --tags)

   # After (persists to all subsequent steps in the same job):
   echo "MY_VAR=$(git describe --tags)" >> "$GITHUB_ENV"
   ```

4. For step outputs (not environment variables), use `$GITHUB_OUTPUT`:

   ```bash
   echo "result=success" >> "$GITHUB_OUTPUT"
   ```

   Then reference the output in a later step with
   `${{ steps.<step-id>.outputs.result }}`.

5. Re-run the workflow and confirm the variable is available in subsequent
   steps.

## Validation

- Add a diagnostic step immediately after the setting step:

  ```bash
  echo "MY_VAR is: $MY_VAR"
  ```

- Confirm the value is printed correctly, not empty.
- Confirm the `::set-env` error no longer appears in the workflow log.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain github-actions-env-not-persisted
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- GitHub Actions environment variable not persisted across steps
- Ci: github actions environment variable not persisted across steps
- Unable to process command '::set-env'
- GitHub Actions github actions environment variable not persisted across steps
- faultline explain github-actions-env-not-persisted


---

*Generated from [playbooks/bundled/log/ci/github-actions-env-not-persisted.yaml](../../../playbooks/bundled/log/ci/github-actions-env-not-persisted.yaml). Do not edit directly — run `make docs-generate`.*
