# Required environment variable missing or empty

**Playbook ID:** `env-var-missing`
**Category:** runtime
**Severity:** high
**Tags:** `env`, `environment-variable`, `config`, `missing`

## What this failure means

A required environment variable was not set or was empty when the process tried to read it. The application, build script, or test suite could not continue without it.

## Common log signals

```text
environment variable
env var
is not set
is not defined
must be set
required env
missing required
unset variable
```

## Diagnosis

A required environment variable was not set or was empty when the process tried to read it. The application, build script, or test suite could not continue without it.

## Fix steps

1. Identify the exact variable name from the error message.
2. Verify the variable is available in the current shell context (use a
   non-secret var to test wiring without exposing values):

   ```bash
   printenv | grep -i VAR_NAME
   ```

3. Add the variable in your CI provider:

   - **GitHub Actions**: Settings › Secrets and variables › Actions, then
     reference with `${{ secrets.VAR_NAME }}` or `${{ vars.VAR_NAME }}`.
     Secrets are withheld from fork PRs — move secret-consuming steps to
     post-merge jobs if this pipeline runs on a fork.
   - **GitLab CI**: Settings › CI/CD › Variables. Ensure the variable is not
     marked protected if the pipeline runs on an unprotected branch.
   - **CircleCI**: Project Settings › Environment Variables, or add it to the
     context that the job references.
   - **Bitbucket Pipelines**: Repository settings › Repository variables.

4. Check if the variable is documented in `.env.example` and provision the
   appropriate value in all required environments.
5. Add a start-of-job validation step that fails early with a clear message:

   ```bash
   : "${VAR_NAME:?VAR_NAME must be set before this job runs}"
   ```

## Validation

- Add a diagnostic step that confirms presence without echoing the value:

  ```bash
  echo "VAR_NAME set: $(test -n "$VAR_NAME" && echo yes || echo NO)"
  ```

- Re-run the failing step and confirm the missing-variable error is gone.

## Likely files to inspect

- `.env.example`
- `.env.template`
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain env-var-missing
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Required environment variable missing or empty
- Runtime: required environment variable missing or empty
- environment variable
- faultline explain env-var-missing


---

*Generated from [playbooks/bundled/log/runtime/env-var-missing.yaml](../../../playbooks/bundled/log/runtime/env-var-missing.yaml). Do not edit directly — run `make docs-generate`.*
