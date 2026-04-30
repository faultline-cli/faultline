# CI secret or environment variable not available

**Playbook ID:** `secrets-not-available`
**Category:** ci
**Severity:** medium
**Tags:** `ci`, `secrets`, `environment-variable`, `fork`, `pr`, `security`

## What this failure means

A required secret or environment variable is not available in this CI job. This commonly occurs when a secret was not provisioned for the target environment, when running from a fork, or when a protected variable is accessed by an unprotected branch or pipeline.

## Common log signals

```text
secret.*not set
Context access might be invalid: secrets
Secret.*is not defined
No secret named
secrets.* is not available
Environment variable.*not found
Error: Input required and not supplied
secret value was not supplied
```

## Diagnosis

A required secret or environment variable is not available in this CI job. This commonly occurs when a secret was not provisioned for the target environment, when running from a fork, or when a protected variable is accessed by an unprotected branch or pipeline.

## Fix steps

1. Verify the secret is defined at the correct scope (repository, environment, project, group, or context).
2. GitHub Actions: secrets are unavailable in fork PRs — move secret-using steps to post-merge jobs or use `pull_request_target` carefully.
3. GitHub Actions: check the job's `environment:` setting and confirm the secret is defined in that environment, not just at the repo level.
4. GitLab CI: ensure the variable is not marked 'protected' if the pipeline runs on an unprotected branch.
5. CircleCI: confirm the context has been shared with the project and the triggering user has access to it.
6. Jenkins: verify the credentials ID in the `withCredentials` block exactly matches the credential defined in Credentials Manager.
7. Azure Pipelines: check the variable group is linked to the pipeline and the variable is not set as secret without proper permissions.

## Validation

- Re-run the failing workflow step.
- Confirm the original failure signature for CI secret or environment variable not available is gone.

## Likely files to inspect

- `.github/workflows/`
- `.gitlab-ci.yml`
- `.circleci/config.yml`
- `Jenkinsfile`
- `azure-pipelines.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain secrets-not-available
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- CI secret or environment variable not available
- Ci: ci secret or environment variable not available
- Context access might be invalid: secrets
- GitHub Actions ci secret or environment variable not available
- faultline explain secrets-not-available


---

*Generated from [playbooks/bundled/log/ci/secrets-not-available.yaml](../../../playbooks/bundled/log/ci/secrets-not-available.yaml). Do not edit directly — run `make docs-generate`.*
