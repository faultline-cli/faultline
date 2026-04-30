# Package manager mismatch in lockfile

**Playbook ID:** `package-manager-mismatch`
**Category:** build
**Severity:** medium
**Tags:** `npm`, `yarn`, `pnpm`, `lockfile`, `package-manager`, `dependency`

## What this failure means

CI used one package manager (e.g., npm) but the repository has a lockfile from a different manager (e.g., pnpm or yarn), causing version and integrity mismatches.

## Common log signals

```text
pnpm-lock.yaml
yarn.lock
package-lock.json
lockfile mismatch
lockfile integrity
integrity mismatch
```

## Diagnosis

The CI job is using a package manager that does not match the lockfile in the repository. Common scenarios:

- Repository has `pnpm-lock.yaml` but CI runs `npm install` or `npm ci`.
- Repository has `yarn.lock` but CI runs `npm install` or `npm ci`.
- CI was updated to use a new package manager but the workflow file was not updated.
- The wrong package manager is installed on the CI runner.

Symptoms include hash mismatches, integrity errors, or unexpected version conflicts.

## Fix steps

1. Check which lockfile exists in the repository:

   ```bash
   ls -la package-lock.json pnpm-lock.yaml yarn.lock 2>/dev/null
   ```

2. Identify the intended package manager. Most projects include it in:

   - `package.json` under `engines.npm`, `engines.yarn`, or `engines.pnpm`
   - `.npmrc`, `.yarnrc`, or `.pnpmrc` files
   - CI workflow comments or documentation

3. Update the CI workflow to use the correct package manager and command:

   - For **npm**: `npm ci`
   - For **yarn**: `yarn install --frozen-lockfile` or `yarn ci`
   - For **pnpm**: `pnpm install --frozen-lockfile`

4. Ensure the correct package manager is installed on the runner. Refer to your CI platform's documentation for switching package managers.

## Validation

- `<package-manager> install` completes without integrity or hash errors.
- `ls -la package-lock.json pnpm-lock.yaml yarn.lock` confirms the expected lockfile matches the package manager.
- Re-run the CI job.

## Likely files to inspect

- `package.json`
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `.npmrc`
- `.yarnrc`
- `.pnpmrc`


## Run Faultline

```bash
faultline analyze build.log
faultline explain package-manager-mismatch
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Package manager mismatch in lockfile
- Build: package manager mismatch in lockfile
- lockfile integrity
- GitHub Actions package manager mismatch in lockfile
- faultline explain package-manager-mismatch
- npm package manager mismatch in lockfile


---

*Generated from [playbooks/bundled/log/build/package-manager-mismatch.yaml](../../../playbooks/bundled/log/build/package-manager-mismatch.yaml). Do not edit directly — run `make docs-generate`.*
