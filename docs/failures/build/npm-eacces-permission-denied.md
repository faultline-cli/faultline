# npm EACCES permission denied in node_modules

**Playbook ID:** `npm-eacces-permission-denied`
**Category:** build
**Severity:** medium
**Tags:** `npm`, `node`, `permissions`, `node_modules`, `eacces`

## What this failure means

npm failed with an `EACCES` error while trying to write to `node_modules` or the global npm cache.

## Common log signals

```text
npm ERR! code EACCES
EACCES: permission denied
permission denied.*node_modules
Error: EACCES: permission denied
Please try running this command again as root/Administrator
```

## Diagnosis

npm needs write permission to the `node_modules` directory and the global npm cache. This error occurs when the CI runner or container runs under a user that lacks those permissions.

Common causes:
- The `node_modules` directory is owned by a different user (e.g., from a previous run with `sudo`).
- The CI job runs as `root` but npm is configured to reject root installs.
- A Docker volume mounts `node_modules` with incorrect ownership.
- The npm cache directory (`~/.npm`) is read‑only.

## Fix steps

1. **Clean and reset permissions** (if `node_modules` already exists):
   ```bash
   rm -rf node_modules
   rm -rf package-lock.json
   ```

2. **Ensure the current user owns the project directory**:
   ```bash
   sudo chown -R $(whoami):$(whoami) .
   ```

3. **If running in Docker, avoid root**:
   - Use the `node` official image's non‑root user (e.g., `node:20-alpine`).
   - Set `USER node` in your Dockerfile before `npm install`.

4. **If you must run as root**, configure npm to allow it:
   ```bash
   npm config set unsafe-perm true
   ```

5. **Check npm cache permissions**:
   ```bash
   npm cache verify
   ls -la ~/.npm
   ```

6. **Use `npm ci` instead of `npm install`** when a `package-lock.json` exists; it is stricter and may avoid some permission issues.

7. **In CI, run as a non‑root user** by adding a step:
   ```bash
   useradd -m -u 1001 ci-user
   chown -R ci-user .
   su ci-user -c "npm install"
   ```

## Validation

- `ls -la node_modules` shows the directory is writable by the current user.
- Re‑run the npm command and confirm no `EACCES` error appears.

## Likely files to inspect

- `package.json`
- `Dockerfile`
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain npm-eacces-permission-denied
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- npm EACCES permission denied in node_modules
- Build: npm eacces permission denied in node_modules
- Please try running this command again as root/Administrator
- GitHub Actions npm eacces permission denied in node_modules
- faultline explain npm-eacces-permission-denied
- npm npm eacces permission denied in node_modules


---

*Generated from [playbooks/bundled/log/build/npm-eacces-permission-denied.yaml](../../playbooks/bundled/log/build/npm-eacces-permission-denied.yaml). Do not edit directly — run `make docs-generate`.*
