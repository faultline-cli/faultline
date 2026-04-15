# missing go.sum entry for module providing package

**Target search query:** `missing go.sum entry for module providing package`

## Error snippet

```text
missing go.sum entry for module providing package github.com/example/dependency; to add:
    go get github.com/example/dependency
```

## What this error means

The Go build needs a module checksum that is missing from `go.sum`, so it refuses to continue with an unverifiable dependency graph.

## Fix steps

1. Run `go mod tidy` with the same Go version used in CI.
2. If the package is missing entirely, add or correct the dependency before rerunning `go mod tidy`.
3. Commit both `go.mod` and `go.sum` if they changed.
4. For private modules, verify `GOPRIVATE` and `GONOSUMDB` are configured correctly.

## How Faultline detects it

Faultline maps this failure to `go-sum-missing`.

Primary signals:

- `missing go.sum entry for module providing package`
- `go.sum file is not up to date`
- `verifying module`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [ModuleNotFoundError: No module named in Python CI jobs](../python/module-not-found-error-no-module-named.md)
- [no such host and getaddrinfo ENOTFOUND in CI](../network/no-such-host-getaddrinfo-enotfound.md)