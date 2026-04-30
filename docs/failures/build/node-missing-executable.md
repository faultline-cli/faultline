# Node.js runtime or tool missing in CI

**Playbook ID:** `node-missing-executable`
**Category:** build
**Severity:** high
**Tags:** `node`, `npm`, `npx`, `executable`, `runtime`, `toolchain`

## What this failure means

The CI job tried to run a Node.js binary (`node`, `npm`, or `npx`) but the runtime was not installed or was not on `$PATH` in the runner or container image.

## Common log signals

```text
node: command not found
npm: command not found
npx: command not found
/usr/bin/env: 'node': No such file or directory
env: node: No such file or directory
sh: node: not found
sh: npm: not found
sh: npx: not found
```

## Diagnosis

The failing step relied on the Node.js runtime being present in the environment, but the runner could not find `node`, `npm`, or `npx` in `$PATH`.

Common causes:

- the container or runner image is not a Node.js image and does not include Node.js
- a setup step (e.g. `actions/setup-node`) was skipped or ran after the failing step
- the Node.js version manager (nvm, fnm, volta) changed the PATH but the shell did not inherit it
- the job runs in a separate Docker container where the host-installed Node.js is not available

## Fix steps

1. **In GitHub Actions**, use `actions/setup-node` before any node-dependent step:

   ```yaml
   - uses: actions/setup-node@v4
     with:
       node-version: '20'
   ```

2. **In a Docker-based CI job**, use a Node.js base image:

   ```dockerfile
   FROM node:20-slim
   ```

3. **If using nvm or fnm**, make sure the activation is in the same shell session as the failing command:

   ```bash
   export NVM_DIR="$HOME/.nvm"
   . "$NVM_DIR/nvm.sh"
   node --version
   ```

4. Confirm Node.js is available before running dependent steps:

   ```bash
   node --version
   npm --version
   ```

## Validation

- `node --version` returns the expected version inside the CI environment.
- `npm --version` succeeds in the same shell session.
- Re-run the failing step and confirm it completes past the binary lookup phase.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `.gitlab-ci.yml`
- `.node-version`
- `.nvmrc`
- `package.json`


## Run Faultline

```bash
faultline analyze build.log
faultline explain node-missing-executable
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Node.js runtime or tool missing in CI
- Build: node.js runtime or tool missing in ci
- /usr/bin/env: 'node': No such file or directory
- GitHub Actions node.js runtime or tool missing in ci
- faultline explain node-missing-executable
- Node.js node.js runtime or tool missing in ci


---

*Generated from [playbooks/bundled/log/build/node-missing-executable.yaml](../../../playbooks/bundled/log/build/node-missing-executable.yaml). Do not edit directly — run `make docs-generate`.*
