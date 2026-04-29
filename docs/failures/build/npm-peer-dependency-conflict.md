# npm peer dependency conflict

**Playbook ID:** `npm-peer-dependency-conflict`
**Category:** build
**Severity:** medium
**Tags:** `npm`, `node`, `dependencies`, `peer-deps`, `resolution`

## What this failure means

npm could not build a valid dependency tree because one package requires a peer version that conflicts with what the project currently installs.

## Common log signals

```text
npm error code ERESOLVE
ERESOLVE unable to resolve dependency tree
Fix the upstream dependency conflict, or retry
--legacy-peer-deps
```

## Diagnosis

npm reached dependency resolution, but a package declared a peer dependency range that does not overlap with the version selected by the root project or another dependency.

This is more specific than generic dependency drift: the resolver is explicitly telling you that a peer contract is incompatible.

## Fix steps

1. Identify the package pair shown in the `Found:` and `Could not resolve dependency:` lines.
2. Align the top-level package versions so the peer dependency range is satisfied.
3. Regenerate the lockfile after updating `package.json`:

   ```bash
   npm install
   ```

4. Avoid relying on `--legacy-peer-deps` or `--force` in CI unless you are intentionally accepting an unsupported combination.
5. If the conflict appeared after a transitive upgrade, pin the affected packages until the upstream packages publish compatible peer ranges.

## Validation

- Re-run `npm install` or `npm ci` from a clean environment.
- Confirm the resolver no longer reports `ERESOLVE` or a peer dependency conflict.

## Likely files to inspect

- `package.json`
- `package-lock.json`
- `.npmrc`


## Run Faultline

```bash
faultline analyze build.log
faultline explain npm-peer-dependency-conflict
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- npm peer dependency conflict
- Build: npm peer dependency conflict
- Fix the upstream dependency conflict, or retry
- GitHub Actions npm peer dependency conflict
- faultline explain npm-peer-dependency-conflict
- npm npm peer dependency conflict


---

*Generated from [playbooks/bundled/log/build/npm-peer-dependency-conflict.yaml](../../playbooks/bundled/log/build/npm-peer-dependency-conflict.yaml). Do not edit directly — run `make docs-generate`.*
