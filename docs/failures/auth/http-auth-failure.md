# HTTP request rejected with 401 or 403 status

**Playbook ID:** `http-auth-failure`
**Category:** auth
**Severity:** high
**Tags:** `http`, `auth`, `401`, `403`, `unauthorized`, `forbidden`, `token`, `credentials`

## What this failure means

An HTTP request was rejected with a 401 Unauthorized or 403 Forbidden status. The
request was missing valid credentials, the supplied token has expired, or the
authenticated principal lacks permission to access the resource.

## Common log signals

```text
403 Forbidden
401 Unauthorized
HTTP 403
HTTP 401
http error 403
http error 401
status 403
status 401
```

## Diagnosis

HTTP authentication failures fall into two classes:

- **401 Unauthorized** — no credentials were supplied, or the credentials were
  invalid (expired token, wrong password, malformed Authorization header).
- **403 Forbidden** — credentials were accepted but the authenticated user or
  role lacks permission for the requested resource (IAM policy, RBAC, network
  policy, or IP allowlist denial).

Common causes:

- **Expired or rotated secret** — an API key, service account credential, or
  personal access token expired or was regenerated; the CI variable was not updated.
- **Missing or wrong scope** — an OAuth token was issued without the scope
  needed for the operation.
- **Wrong environment** — a staging credential is used against a production
  endpoint or vice versa.
- **IP allowlist** — the CI runner's egress IP is not in the service's allowlist,
  causing the request to be denied as 403.
- **Resource permission** — the token is valid but the principal's role does not
  grant access to the specific resource (read-only token attempting a write, etc.).

## Fix steps

1. Identify which request failed and which service returned the error — the log
   usually shows the full URL and status code.
2. Check that the relevant CI secret or environment variable is set and up to date:

   ```bash
   # GitHub Actions: check the secret is defined and not expired
   printenv TOKEN_NAME | cut -c1-10  # show first 10 chars only
   ```

3. For 401 errors: rotate the credential and update the CI secret.
4. For 403 errors: review the IAM role, RBAC policy, or resource permission that
   governs the authenticated principal. Confirm the runner's IP is in any allowlist.
5. Confirm the correct endpoint and environment: compare the service URL in the
   failing step to what the credential was issued for.
6. For OAuth/OIDC tokens: verify the requested scope includes the operation being
   performed (read vs. write, specific resource type, etc.).

## Validation

- Re-run the failing step and confirm the request succeeds with a 2xx status.
- Confirm no authentication or authorization error appears in the log.

## Likely files to inspect

- `.github/workflows/`
- `.gitlab-ci.yml`
- `.env`


## Run Faultline

```bash
faultline analyze build.log
faultline explain http-auth-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- HTTP request rejected with 401 or 403 status
- Auth: http request rejected with 401 or 403 status
- authorization failed
- faultline explain http-auth-failure


---

*Generated from [playbooks/bundled/log/auth/http-auth-failure.yaml](../../../playbooks/bundled/log/auth/http-auth-failure.yaml). Do not edit directly — run `make docs-generate`.*
