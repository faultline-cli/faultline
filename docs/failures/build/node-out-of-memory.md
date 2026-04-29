# Node.js JavaScript heap out of memory

**Playbook ID:** `node-out-of-memory`
**Category:** build
**Severity:** high
**Tags:** `node`, `javascript`, `webpack`, `nextjs`, `oom`, `heap`, `memory`

## What this failure means

The Node.js process ran out of heap memory during the build. This is common in large Webpack, Next.js, or Jest builds where the default V8 heap limit is too low for the amount of modules being processed.

## Common log signals

```text
FATAL ERROR: Reached heap limit Allocation failed
JavaScript heap out of memory
FATAL ERROR: CALL_AND_RETRY_LAST Allocation failed
Ineffective mark-compacts near heap limit
<--- Last few GCs --->
<--- JS stacktrace --->
Allocation failed - JavaScript heap out of memory
ERR_WORKER_OUT_OF_MEMORY
```

## Diagnosis

The Node.js process ran out of heap memory during the build. This is common in large Webpack, Next.js, or Jest builds where the default V8 heap limit is too low for the amount of modules being processed.

## Fix steps

1. Increase the Node.js heap limit: set `NODE_OPTIONS=--max-old-space-size=4096` (or higher) before the build command.
2. For npm scripts, add it inline in package.json scripts: `"build": "NODE_OPTIONS=--max-old-space-size=4096 next build"`.
3. For Jest OOM: pass `--maxWorkers=2` or `--runInBand` to serialize test execution and lower peak memory.
4. Upgrade the CI runner to a higher-memory tier — heap expansion is not free; more RAM is the most reliable fix.
5. Identify the heap consumer: add `--inspect` or run `node --max-old-space-size=2048 --trace-gc` and profile with Chrome DevTools.
6. Break the build into smaller increments (e.g., build packages individually in a monorepo).

## Validation

- Re-run the local reproduction command after the fix.
- NODE_OPTIONS=--max-old-space-size=4096 npm run build

## Likely files to inspect

- `package.json`
- `webpack.config.js`
- `next.config.js`
- `jest.config.js`


## Run Faultline

```bash
faultline analyze build.log
faultline explain node-out-of-memory
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Node.js JavaScript heap out of memory
- Build: node.js javascript heap out of memory
- FATAL ERROR: CALL_AND_RETRY_LAST Allocation failed
- GitHub Actions node.js javascript heap out of memory
- faultline explain node-out-of-memory
- Node.js node.js javascript heap out of memory


---

*Generated from [playbooks/bundled/log/build/node-out-of-memory.yaml](../../playbooks/bundled/log/build/node-out-of-memory.yaml). Do not edit directly — run `make docs-generate`.*
