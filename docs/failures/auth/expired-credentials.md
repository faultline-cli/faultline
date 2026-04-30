# Expired credentials or rotated secrets

**Playbook ID:** `expired-credentials`
**Category:** auth
**Severity:** high
**Tags:** `auth`, `credentials`, `expired`, `token`, `secret`, `rotation`

## What this failure means

A CI job is using credentials that have expired, been revoked, or rotated
outside of CI. The job can no longer authenticate to the target service
(AWS, Docker registry, GitHub, npm, etc.) and the step fails with an
authentication or authorization error.

## Common log signals

```text
credentials have expired
credential has expired
token has expired
token expired
ExpiredToken
ExpiredTokenException
WebIdentityErr
InvalidClientTokenId
```

## Diagnosis

Credentials expire for several reasons:
- Personal access tokens have a configured expiry (30, 60, 90 days)
- AWS IAM temporary credentials from STS/OIDC have a short TTL
- Service account keys are rotated on schedule
- Passwords are changed and CI was not updated
- API keys are revoked during a security rotation event

Check whether the failure is consistent (every run) or recent (started failing
after a date). A sudden start of failures without code changes is a strong
indicator of expire/rotation.

Identify the affected credential:

```bash
# GitHub Actions — see which secret is referenced in the failing step
# Look for: uses: <action> that calls the service, or explicit env vars
```

## Fix steps

1. Identify the expired credential from the error message (service name,
   token type, or IAM role).

2. Generate a new credential in the upstream service:
   - GitHub PAT: Settings > Developer settings > Personal access tokens
   - AWS IAM: IAM console > Users/Roles > Security credentials
   - Docker Hub: Account settings > Security > New access token
   - npm: `npm token create`

3. Update the secret in CI:
   - GitHub Actions: Settings > Secrets and variables > Actions > update secret
   - GitLab CI: Settings > CI/CD > Variables > update variable
   - CircleCI: Project settings > Environment variables

4. Verify the new credential has the required scopes and permissions.

5. For AWS OIDC, check that the trust relationship and audience claim
   are correctly configured for the new token provider.

6. If using short-lived OIDC tokens, verify the CI platform's OIDC provider
   URL is registered in the identity provider trust policy.

## Validation

- Re-run the failing job and confirm the authentication step passes.
- Check that subsequent dependent steps also succeed.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain expired-credentials
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Expired credentials or rotated secrets
- Auth: expired credentials or rotated secrets
- The security token included in the request is expired
- faultline explain expired-credentials


---

*Generated from [playbooks/bundled/log/auth/expired-credentials.yaml](../../playbooks/bundled/log/auth/expired-credentials.yaml). Do not edit directly — run `make docs-generate`.*
