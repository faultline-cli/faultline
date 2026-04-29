# Docker registry authentication failure

**Playbook ID:** `docker-auth`
**Category:** auth
**Severity:** high
**Tags:** `docker`, `registry`, `auth`

## What this failure means

CI could not authenticate to the container registry before an image pull or push.


## Common log signals

```text
pull access denied
requested access to the resource is denied
unauthorized: authentication required
no basic auth credentials
denied: requested access to the resource is denied
may require 'docker login'
Error response from daemon: Get https://mcr.microsoft.com/v2/: Forbidden
```

## Diagnosis

Docker reached the registry, but the credentials used by CI were missing, expired, or scoped incorrectly for the target image.

Typical causes:

- the login step never ran
- the secret value rotated without the workflow updating
- the token can read one repo but not the image path being used


## Fix steps

1. Verify the registry username, token, or password configured in CI secrets.
2. Ensure the registry login step runs before any `docker pull` or `docker push` command.
3. Confirm the token has the correct repository scope for the image being accessed.
4. Validate the same credential locally:

   ```bash
   docker login <registry>
   ```


## Validation

- Re-run the registry login step.
- Confirm the pull or push completes without `authentication required`.
- If you changed secret wiring, run one full CI job after rotating the credential.


## Likely files to inspect

- `Dockerfile`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `docker-compose*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain docker-auth
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Docker registry authentication failure
- Auth: docker registry authentication failure
- Error response from daemon: Get https://mcr.microsoft.com/v2/: Forbidden
- faultline explain docker-auth
- Docker docker registry authentication failure


---

*Generated from [playbooks/bundled/log/auth/docker-auth.yaml](../../playbooks/bundled/log/auth/docker-auth.yaml). Do not edit directly — run `make docs-generate`.*
