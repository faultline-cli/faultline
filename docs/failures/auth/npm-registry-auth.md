# npm or yarn registry authentication failure

**Playbook ID:** `npm-registry-auth`
**Category:** auth
**Severity:** high
**Tags:** `npm`, `yarn`, `pnpm`, `registry`, `auth`, `packages`

## What this failure means

The Node.js package manager could not authenticate to a private or
organization npm registry. The request was rejected with a 401 or 403
status, blocking dependency installation.

## Common log signals

```text
npm ERR! code E401
npm ERR! 401 Unauthorized
npm error code E401
npm error 401 Unauthorized
ERR_PNPM_FETCH_401
ERR_PNPM_FETCH_403
npm ERR! code E403
npm ERR! 403 Forbidden
```

## Diagnosis

The package manager reached the npm registry but authentication failed
(HTTP 401 Unauthorized or 403 Forbidden). This blocks dependency
installation for any private or scoped package.

Root cause: the auth token for the target registry scope is missing,
expired, or does not have read access to the requested package.

Common causes:
- `NODE_AUTH_TOKEN` or `NPM_TOKEN` CI secret is unset or expired
- `.npmrc` scope-to-registry mapping is missing or misconfigured
- Token lacks `read:packages` scope (GitHub Packages)
- Wrong registry URL in `.npmrc` for the package scope

## Fix steps

1. Confirm the CI secret holding the registry token is set and current:
   - GitHub Actions: `NODE_AUTH_TOKEN` mapped to the registry token secret
   - GitLab CI: `NPM_TOKEN` variable scoped to the project or group

2. Add or verify the `.npmrc` entry that maps the package scope to the
   registry and injects the auth token:

   ```
   @myorg:registry=https://npm.pkg.github.com/
   //npm.pkg.github.com/:_authToken=${NODE_AUTH_TOKEN}
   ```

3. For GitHub Packages, ensure the token has `read:packages` scope.

4. For npmjs.org private packages, use an automation token with
   read-only or publish access as required.

5. For pnpm, set the same `.npmrc` content — pnpm reads npm's registry
   auth configuration.

6. Test the credential against the registry:

   ```bash
   npm whoami --registry https://npm.pkg.github.com/
   ```

## Validation

- Re-run `npm ci` or `yarn install` and confirm no E401 or E403 appears.
- Confirm `npm whoami --registry <url>` returns the expected identity.

## Likely files to inspect

- `.npmrc`
- `package.json`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `.yarnrc.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain npm-registry-auth
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- npm or yarn registry authentication failure
- Auth: npm or yarn registry authentication failure
- unauthenticated: User cannot be authenticated with the token provided
- faultline explain npm-registry-auth
- npm npm or yarn registry authentication failure


---

*Generated from [playbooks/bundled/log/auth/npm-registry-auth.yaml](../../../playbooks/bundled/log/auth/npm-registry-auth.yaml). Do not edit directly — run `make docs-generate`.*
