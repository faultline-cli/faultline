# jest command not found or not installed

**Playbook ID:** `jest-command-not-found`
**Category:** test
**Severity:** high
**Tags:** `jest`, `node`, `test`, `command`, `missing`

## What this failure means

The `jest` command could not be found. Jest is either not installed as a dependency, not in the PATH, or the wrong shell was used to run it.

## Common log signals

```text
jest: command not found
path.*\.bin/jest
```

## Diagnosis

Jest is typically installed locally as a dev dependency and accessed through `npm run test` or `npx jest`. Common causes:

- Jest is not listed in `package.json` as a `devDependency`.
- `npm install` or `npm ci` was not run before the test step, or it failed silently.
- The script is running in a different shell that does not have the PATH updated.
- Jest is installed locally but the script tries to use the global `jest` command.
- The test step runs before the install step in CI.

The error typically appears as `jest: command not found` or `jest is not installed`.

## Fix steps

1. Verify Jest is in `package.json`:

   ```bash
   grep jest package.json
   ```

2. If missing, add it as a dev dependency:

   ```bash
   npm install --save-dev jest
   ```

3. Ensure `npm install` completes before running tests:

   ```bash
   npm install
   npm test
   ```

4. Use `npm run test` or `npx jest` instead of calling `jest` directly:

   ```bash
   npm run test    # Uses local jest from node_modules
   npx jest        # Also uses local jest
   jest            # May not work if jest is not in PATH
   ```

5. Check that the CI job runs the test step after the install step:

   ```yaml
   steps:
     - uses: actions/checkout@v4
     - uses: actions/setup-node@v3
       with:
         node-version: '18'
     - run: npm install     # Must run BEFORE test
     - run: npm test        # Now jest is available
   ```

## Validation

- `npm install` completes without errors.
- `npm run test` or `npx jest` executes tests successfully.
- `npm ls jest` shows jest is installed in `node_modules`.

## Likely files to inspect

- `package.json`
- `jest.config.js`
- `.github/workflows/*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain jest-command-not-found
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- jest command not found or not installed
- Test: jest command not found or not installed
- jest: command not found
- faultline explain jest-command-not-found
- Node.js jest command not found or not installed


---

*Generated from [playbooks/bundled/log/test/jest-command-not-found.yaml](../../playbooks/bundled/log/test/jest-command-not-found.yaml). Do not edit directly — run `make docs-generate`.*
