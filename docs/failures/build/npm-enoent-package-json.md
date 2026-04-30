# npm ENOENT package.json missing

**Playbook ID:** `npm-enoent-package-json`
**Category:** build
**Severity:** medium
**Tags:** `npm`, `node`, `package.json`, `missing`

## What this failure means

npm could not find a `package.json` file in the current directory.

## Common log signals

```text
ENOENT: no such file or directory, open 'package.json'
ENOENT.*package.json
npm ERR! code ENOENT
npm ERR!.*package.json
could not read package.json
package.json not found
```

## Diagnosis

npm expects a `package.json` file in the current working directory when running commands like `npm install`, `npm ci`, or `npm run`. The file may be missing, the CI job may be running in the wrong directory, or the repository checkout may have failed.

This error typically appears as `ENOENT: no such file or directory, open 'package.json'`.

## Fix steps

1. Verify the repository was checked out successfully and contains a `package.json` file.
2. Check the CI job's `working-directory` configuration. It may be pointing to a subdirectory that doesn't contain `package.json`.
3. If using a monorepo, ensure you're in the correct package directory.
4. Add a step to list files and confirm `package.json` exists before running npm commands:

   ```bash
   ls -la
   pwd
   ```

5. If `package.json` is missing from the repository, create it with the necessary fields or restore it from version control.

## Validation

- Run `ls -la package.json` to confirm the file exists.
- Re-run the npm command that previously failed.
- Ensure the command runs in the correct directory.

## Likely files to inspect

- `package.json`
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain npm-enoent-package-json
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- npm ENOENT package.json missing
- Build: npm enoent package.json missing
- ENOENT: no such file or directory, open 'package.json'
- GitHub Actions npm enoent package.json missing
- faultline explain npm-enoent-package-json
- npm npm enoent package.json missing


---

*Generated from [playbooks/bundled/log/build/npm-enoent-package-json.yaml](../../../playbooks/bundled/log/build/npm-enoent-package-json.yaml). Do not edit directly — run `make docs-generate`.*
