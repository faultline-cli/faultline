# npm ERESOLVE dependency tree conflict

**Playbook ID:** `npm-eresolve-conflict`
**Category:** build
**Severity:** high
**Tags:** `npm`, `node`, `dependency`, `resolution`, `peer`

## What this failure means

npm encountered a dependency tree conflict that it could not automatically resolve. Usually a peer dependency or incompatible version constraint.

## Common log signals

```text
npm ERR! code ERESOLVE
ERESOLVE could not resolve
Could not resolve dependency:
```

## Diagnosis

npm's resolver found conflicting requirements in the dependency tree. This typically occurs when:

- A peer dependency version constraint conflicts with what npm wants to install
- Two packages require incompatible versions of the same transitive dependency
- A package specifies `peerDependencies` that are not satisfied

The error appears as `npm ERR! code ERESOLVE` followed by a resolution error message.

In npm 7+, this is an error by default (in npm 6 it was a warning).

## Fix steps

1. Read the full ERESOLVE error message to identify the conflicting packages and versions.

2. Check if the error suggests using `--legacy-peer-deps` as a workaround (common when upgrading packages).

3. Update the relevant `package.json` to resolve the conflict:

   - Loosen the version constraint if safe for your project.
   - Upgrade one of the conflicting packages to a compatible version.
   - Remove an unnecessary dependency if it is not actually used.

4. If you cannot modify the dependency to fix the conflict, run `npm install --legacy-peer-deps` as a temporary workaround (not recommended for production):

   ```bash
   npm install --legacy-peer-deps
   ```

5. Verify the installed versions satisfy your requirements:

   ```bash
   npm ls <package-name>
   ```

## Validation

- `npm install` completes without `ERESOLVE` errors.
- `npm ls` shows all dependencies at compatible versions.
- Run the build step to confirm downstream tasks succeed.

## Likely files to inspect

- `package.json`
- `package-lock.json`
- `.npmrc`


## Run Faultline

```bash
faultline analyze build.log
faultline explain npm-eresolve-conflict
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- npm ERESOLVE dependency tree conflict
- Build: npm eresolve dependency tree conflict
- Could not resolve dependency:
- GitHub Actions npm eresolve dependency tree conflict
- faultline explain npm-eresolve-conflict
- npm npm eresolve dependency tree conflict


---

*Generated from [playbooks/bundled/log/build/npm-eresolve-conflict.yaml](../../../playbooks/bundled/log/build/npm-eresolve-conflict.yaml). Do not edit directly — run `make docs-generate`.*
