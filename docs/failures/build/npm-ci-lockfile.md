# npm ci lockfile mismatch

**Playbook ID:** `npm-ci-lockfile`
**Category:** build
**Severity:** medium
**Tags:** `npm`, `node`, `lockfile`, `dependencies`

## What this failure means

`npm ci` found a missing or out-of-sync `package-lock.json`.

## Common log signals

```text
npm ci can only install packages when your package.json and package-lock.json
npm error `npm ci` can only install packages when your
package.json and package-lock.json are in sync
missing package-lock.json
npm ERR! cipm can only install packages
package-lock.json does not exist
run `npm install` to generate a lockfile
```

## Diagnosis

`npm ci` installs strictly from the lockfile. If `package.json` and `package-lock.json` disagree, CI fails instead of regenerating dependencies on the fly.

## Fix steps

1. Regenerate the lockfile locally:

   ```bash
   npm install
   ```

2. Commit the updated `package-lock.json`.
3. Make sure `package-lock.json` is not ignored.
4. If the repo uses workspaces, regenerate the lockfile from the workspace root with the same npm major version used in CI.

## Validation

- Run `npm ci` locally.
- Re-run the CI job.
- Check that `package-lock.json` stays unchanged after the install step.

## Likely files to inspect

- `package.json`
- `package-lock.json`
- `.npmrc`


## Run Faultline

```bash
faultline analyze build.log
faultline explain npm-ci-lockfile
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- npm ci lockfile mismatch
- Build: npm ci lockfile mismatch
- npm ci can only install packages when your package.json and package-lock.json
- GitHub Actions npm ci lockfile mismatch
- faultline explain npm-ci-lockfile
- npm npm ci lockfile mismatch


---

*Generated from [playbooks/bundled/log/build/npm-ci-lockfile.yaml](../../../playbooks/bundled/log/build/npm-ci-lockfile.yaml). Do not edit directly — run `make docs-generate`.*
