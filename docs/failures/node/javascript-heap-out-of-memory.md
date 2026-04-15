# JavaScript heap out of memory in Node.js builds

**Target search query:** `JavaScript heap out of memory CI build`

## Error snippet

```text
FATAL ERROR: Reached heap limit Allocation failed - JavaScript heap out of memory
<--- Last few GCs --->
```

## What this error means

The Node.js process exhausted its V8 heap during the build or test run. This is common in large Webpack, Next.js, and Jest workloads on memory-constrained runners.

## Fix steps

1. Increase the heap limit with `NODE_OPTIONS=--max-old-space-size=4096`.
2. Reduce parallelism for Jest or worker-heavy tasks.
3. Upgrade the CI runner to a higher-memory tier if the build is legitimately large.
4. Break the build into smaller units if a monorepo or bundle step is too heavy.

## How Faultline detects it

Faultline maps this failure to `node-out-of-memory`.

Primary signals:

- `JavaScript heap out of memory`
- `FATAL ERROR: Reached heap limit Allocation failed`
- `ERR_WORKER_OUT_OF_MEMORY`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [The engine "node" is incompatible with this module](the-engine-node-is-incompatible-with-this-module.md)
- [Your lockfile needs to be updated, but yarn was run with --frozen-lockfile](your-lockfile-needs-to-be-updated-yarn-frozen-lockfile.md)