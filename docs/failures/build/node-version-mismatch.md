# Node.js version mismatch

**Playbook ID:** `node-version-mismatch`
**Category:** build
**Severity:** medium
**Tags:** `node`, `nodejs`, `nvmrc`, `engines`, `nvm`, `version`

## What this failure means

The Node.js version installed on the CI runner does not satisfy the `engines` constraint declared in `package.json` or the version pinned in `.nvmrc`.

## Common log signals

```text
The engine "node" is incompatible with this module
Required engine node
requires node version
engine: unsatisfied
node: /.*: version
Unsupported engine
error node@
warning: You are using Node.js
```

## Diagnosis

The Node.js version installed on the CI runner does not satisfy the `engines` constraint declared in `package.json` or the version pinned in `.nvmrc`.

## Fix steps

1. Identify the required Node.js version from `.nvmrc`, `.node-version`, or `package.json` `engines.node`.
2. GitHub Actions: set `node-version-file: .nvmrc` on the `actions/setup-node` step.
3. GitLab CI: use a versioned image — `image: node:20-alpine` — matching your `.nvmrc`.
4. CircleCI: use `cimg/node:20.x` or pin the version in the `.circleci/config.yml` executor.
5. Azure Pipelines: use the `NodeTool@0` task with a pinned `versionSpec`.
6. Locally: run `nvm use` or `volta install node` to switch to the pinned version.

## Validation

- Re-run the local reproduction command after the fix.
- cat .nvmrc && node --version
- nvm use

## Likely files to inspect

- `.nvmrc`
- `.node-version`
- `package.json`


## Run Faultline

```bash
faultline analyze build.log
faultline explain node-version-mismatch
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Node.js version mismatch
- Build: node.js version mismatch
- The engine "node" is incompatible with this module
- GitHub Actions node.js version mismatch
- faultline explain node-version-mismatch
- Node.js node.js version mismatch


---

*Generated from [playbooks/bundled/log/build/node-version-mismatch.yaml](../../../playbooks/bundled/log/build/node-version-mismatch.yaml). Do not edit directly — run `make docs-generate`.*
