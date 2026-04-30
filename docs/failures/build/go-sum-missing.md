# Missing go.sum entry

**Playbook ID:** `go-sum-missing`
**Category:** build
**Severity:** medium
**Tags:** `go`, `modules`, `dependencies`

## What this failure means

The build needs a module checksum that is missing from `go.sum`.

## Common log signals

```text
missing go.sum entry for module providing package
go.sum file is not up to date
verifying module
no required module provides
go.sum entry missing
go.sum entry not in vendor
partial content
```

## Diagnosis

Go refused to continue because the dependency graph references a module version that does not have a matching checksum entry in `go.sum`.

This usually happens after dependency changes were made locally without committing the resulting module metadata.

## Fix steps

1. Recalculate module metadata with the same Go version used in CI:

   ```bash
   go mod tidy
   ```

2. If the error says no required module provides a package, add or correct the dependency before rerunning `go mod tidy`.
3. Commit both `go.mod` and `go.sum` if either file changed.
4. For private modules, verify `GOPRIVATE` and `GONOSUMDB` are set correctly.
5. If the module cache looks stale, clear it once with `go clean -modcache` and rerun the build.

## Validation

- Run `git diff --exit-code go.mod go.sum` after `go mod tidy`.
- Re-run `go test ./...`.
- Confirm CI no longer reports a missing checksum entry.

## Likely files to inspect

- `go.mod`
- `go.sum`


## Run Faultline

```bash
faultline analyze build.log
faultline explain go-sum-missing
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Missing go.sum entry
- Build: missing go.sum entry
- missing go.sum entry for module providing package
- GitHub Actions missing go.sum entry
- faultline explain go-sum-missing
- Go missing go.sum entry


---

*Generated from [playbooks/bundled/log/build/go-sum-missing.yaml](../../../playbooks/bundled/log/build/go-sum-missing.yaml). Do not edit directly — run `make docs-generate`.*
