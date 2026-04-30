# Missing required environment variable

**Playbook ID:** `missing-env`
**Category:** auth
**Severity:** high
**Tags:** `env`, `config`, `secrets`

## What this failure means

A required environment variable was not set in the CI environment. The job cannot continue without the expected configuration or secret value.

## Common log signals

```text
environment variable not set
required env var
missing required variable
unbound variable
please set the
token is empty
api key is empty
secret is empty
```

## Diagnosis

A required environment variable was not set in the CI environment. The job cannot continue without the expected configuration or secret value.

## Fix steps

1. Identify the missing variable name from the error and add it to your CI
   provider:
   - **GitHub Actions**: Settings › Secrets and variables › Actions.
   - **GitLab CI**: Settings › CI/CD › Variables (check the Protected flag).
   - **CircleCI**: Project Settings › Environment Variables or the context.
   - **Bitbucket Pipelines**: Repository settings › Repository variables.

2. Reference the variable in the workflow file with the correct name — names
   are case-sensitive.

3. For GitHub Actions: secrets are withheld from fork PRs by default. If the
   job runs against a fork, move secret-consuming steps to a post-merge job
   or use `pull_request_target` with caution.

4. Verify the variable is actually exported at runtime (use a non-sensitive
   variable to test wiring):

   ```bash
   echo "VAR set: $(test -n "$VAR_NAME" && echo yes || echo NO)"
   ```

5. Check that repository forks and non-default branches have access to the
   required secrets and are not blocked by protection rules.

## Validation

- Add a diagnostic step confirming availability without revealing the value.
- Re-run the failing step and confirm the empty-variable error is gone.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `.gitlab-ci.yml`
- `.env.example`


## Run Faultline

```bash
faultline analyze build.log
faultline explain missing-env
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Missing required environment variable
- Auth: missing required environment variable
- environment variable not set
- faultline explain missing-env


---

*Generated from [playbooks/bundled/log/auth/missing-env.yaml](../../../playbooks/bundled/log/auth/missing-env.yaml). Do not edit directly — run `make docs-generate`.*
