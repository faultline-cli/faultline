# AWS credentials missing or invalid

**Playbook ID:** `aws-credentials`
**Category:** auth
**Severity:** high
**Tags:** `aws`, `credentials`, `iam`, `auth`

## What this failure means

The job could not authenticate with AWS. Either no credentials are present in the environment, the credentials are expired, or the IAM principal does not have permission to perform the requested action.

## Common log signals

```text
NoCredentialProviders
no EC2 instance role found
InvalidClientTokenId
AuthFailure
AccessDenied
Unable to locate credentials
An error occurred (AuthFailure)
could not load credentials
```

## Diagnosis

The job could not authenticate with AWS. Either no credentials are present in the environment, the credentials are expired, or the IAM principal does not have permission to perform the requested action.

## Fix steps

1. In the same CI job environment or container, verify credentials are visible to the AWS SDK: `aws configure list` — shows which source (env, profile, instance profile) is active.
2. Confirm the resolved identity with the same runtime context: `aws sts get-caller-identity` — if this fails, the credential chain has no valid entry.
3. For `ExpiredTokenException`: rotate the IAM access key in the console and update the `AWS_SECRET_ACCESS_KEY` CI secret; for assumed-role sessions, regenerate via `aws sts assume-role`.
4. For `AccessDenied`: decode the error to get the full policy evaluation context: `aws sts decode-authorization-message --encoded-message <encoded>` — this reveals which statement denied the action.
5. For OIDC (`aws-actions/configure-aws-credentials`): verify the IAM role trust policy `Condition` block — the `sub` claim format is `repo:<org>/<repo>:ref:refs/heads/<branch>`.
6. For SCPs (Service Control Policies) in AWS Organizations: even with IAM allow, an SCP deny at the org level blocks the action — check with your AWS administrator.
7. Prefer OIDC-based role assumption over long-lived access keys to eliminate credential rotation overhead.

## Validation

- aws configure list
- aws sts get-caller-identity

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `.gitlab-ci.yml`
- `buildspec.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain aws-credentials
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- AWS credentials missing or invalid
- Auth: aws credentials missing or invalid
- AWS was not able to validate the provided access credentials
- faultline explain aws-credentials


---

*Generated from [playbooks/bundled/log/auth/aws-credentials.yaml](../../../playbooks/bundled/log/auth/aws-credentials.yaml). Do not edit directly — run `make docs-generate`.*
