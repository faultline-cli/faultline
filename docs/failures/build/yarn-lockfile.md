# Yarn lockfile out of date

**Playbook ID:** `yarn-lockfile`
**Category:** build
**Severity:** medium
**Tags:** `yarn`, `node`, `lockfile`, `dependencies`

## What this failure means

Yarn was run with `--frozen-lockfile` (the recommended CI flag) but the `yarn.lock` file is either missing or no longer matches `package.json`.

## Common log signals

```text
your lockfile needs to be updated
yarn.lock: No such file or directory
error your lockfile needs to be updated
YN0028
The lockfile would have been modified by this install
lockfile is frozen
```

## Diagnosis

Yarn was run with `--frozen-lockfile` (the recommended CI flag) but the `yarn.lock` file is either missing or no longer matches `package.json`.

## Fix steps

1. Run `yarn install` locally.
2. Commit the updated `yarn.lock`.
3. Ensure `yarn.lock` is not in `.gitignore`.
4. If the repository uses Yarn workspaces, regenerate the lockfile from the workspace root with the same Yarn version used in CI.

## Validation

- Re-run the failing workflow step.
- Confirm the original failure signature for Yarn lockfile out of date is gone.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain yarn-lockfile
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Yarn lockfile out of date
- Build: yarn lockfile out of date
- The lockfile would have been modified by this install
- GitHub Actions yarn lockfile out of date
- faultline explain yarn-lockfile
- Yarn yarn lockfile out of date


---

*Generated from [playbooks/bundled/log/build/yarn-lockfile.yaml](../../../playbooks/bundled/log/build/yarn-lockfile.yaml). Do not edit directly — run `make docs-generate`.*
