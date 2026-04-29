# node-gyp missing build tools

**Playbook ID:** `node-gyp-missing-build-tools`
**Category:** build
**Severity:** medium
**Tags:** `node`, `npm`, `node-gyp`, `build-tools`, `native`

## What this failure means

`node-gyp` failed because required build tools (Python, C++ compiler, etc.) are missing from the CI environment.

## Common log signals

```text
node-gyp ERR!
gyp ERR!
missing: python
Can't find Python executable
No Xcode or CLT version detected
Command failed: python
make: not found
g++: not found
```

## Diagnosis

`node-gyp` is a tool used to compile native Node.js addons. It requires Python, a C++ compiler (like g++ or clang), and other build tools to be installed. CI environments often lack these by default, especially in minimal container images.

The error typically appears when installing packages with native dependencies (e.g., `bcrypt`, `sharp`, `sqlite3`).

## Fix steps

1. Install the necessary build tools before running `npm install` or `npm ci`:

   **Ubuntu/Debian:**
   ```bash
   apt-get update && apt-get install -y python3 make g++
   ```

   **Alpine:**
   ```bash
   apk add --no-cache python3 make g++
   ```

   **macOS (GitHub Actions):**
   ```bash
   brew install python3
   ```

2. If using a Node.js base image, switch to an image that includes build tools, such as `node:20` (includes g++ and Python) instead of `node:20-alpine` (minimal, lacks build tools).

3. For packages that offer pre‑built binaries, ensure the environment matches the binary target (e.g., `npm_config_platform`, `npm_config_arch`). You can also set `npm_config_build_from_source=true` to force compilation.

4. If the project does not actually need native addons, consider removing the offending dependency or switching to a pure‑JavaScript alternative.

## Validation

- Verify that `python3 --version` and `g++ --version` succeed in the CI environment.
- Re‑run the install step and confirm `node-gyp` no longer reports missing tools.

## Likely files to inspect

- `package.json`
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `Dockerfile`


## Run Faultline

```bash
faultline analyze build.log
faultline explain node-gyp-missing-build-tools
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- node-gyp missing build tools
- Build: node-gyp missing build tools
- error: no acceptable C compiler found
- GitHub Actions node-gyp missing build tools
- faultline explain node-gyp-missing-build-tools
- Node.js node-gyp missing build tools


---

*Generated from [playbooks/bundled/log/build/node-gyp-missing-build-tools.yaml](../../playbooks/bundled/log/build/node-gyp-missing-build-tools.yaml). Do not edit directly — run `make docs-generate`.*
