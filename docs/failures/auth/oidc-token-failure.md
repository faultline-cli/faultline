# OIDC token request failed or token invalid

**Playbook ID:** `oidc-token-failure`
**Category:** auth
**Severity:** high
**Tags:** `oidc`, `jwt`, `token`, `federated`, `aws`, `gcp`, `azure`, `github-actions`

## What this failure means

An OIDC token request failed or the issued token was rejected by the cloud provider. Federated authentication to AWS, GCP, or Azure cannot proceed.

## Common log signals

```text
OIDC token
id_token
Unable to get OIDC token
JWT validation
Invalid JWT
JWT verification failed
JWT expired
JWT signature
```

## Diagnosis

An OIDC token request failed or the issued token was rejected by the cloud provider. Federated authentication to AWS, GCP, or Azure cannot proceed.

## Fix steps

1. In GitHub Actions, ensure the job has `permissions: id-token: write`.
2. Decode the OIDC token and inspect `iss`, `sub`, `aud`, and repository claims to confirm they match the cloud-side trust policy.
3. For AWS, verify the IAM role trust policy `StringLike` condition matches the actual `sub` claim.
4. For GCP, check the Workload Identity Pool provider attribute mapping and condition logic.
5. For Azure, confirm the federated credential issuer, subject, and audience are correct.
6. For token audience mismatch errors, set the expected audience explicitly in the workflow step or provider configuration.

## Validation

- Re-run the failing workflow step with debug logging enabled.
- Confirm `ACTIONS_ID_TOKEN_REQUEST_URL` is present when the workflow expects GitHub-issued OIDC.

## Likely files to inspect

- `.github/workflows/`
- `.gitlab-ci.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain oidc-token-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- OIDC token request failed or token invalid
- Auth: oidc token request failed or token invalid
- Error getting token from GitHub's OIDC provider
- GitHub Actions oidc token request failed or token invalid
- faultline explain oidc-token-failure


---

*Generated from [playbooks/bundled/log/auth/oidc-token-failure.yaml](../../playbooks/bundled/log/auth/oidc-token-failure.yaml). Do not edit directly — run `make docs-generate`.*
