# Go race detector found a data race

**Playbook ID:** `go-data-race`
**Category:** test
**Severity:** high
**Tags:** `go`, `race`, `concurrency`, `data-race`, `test`

## What this failure means

The Go race detector identified a data race: two goroutines accessed the same memory location concurrently, with at least one performing a write, without any synchronization between them.

## Common log signals

```text
WARNING: DATA RACE
DATA RACE
Previous read at
Previous write at
Read at 0x
Write at 0x
Goroutine.*running
PASS (race)
```

## Diagnosis

The Go race detector identified a data race: two goroutines accessed the same memory location concurrently, with at least one performing a write, without any synchronization between them.

## Fix steps

1. Read the full race report — it shows both goroutine stacks and the file/line for each conflicting access.
2. Protect the shared variable with `sync.Mutex` or `sync.RWMutex`, or use an atomic operation (`sync/atomic`) for simple counters.
3. Replace concurrent map access with `sync.Map` or add a `sync.RWMutex` around reads and writes.
4. For loop-variable captures: copy the variable into a local before launching the goroutine: `v := v`.
5. Run locally to reproduce: `go test -race ./...` or `go test -race -count=5 ./...` for intermittent races.

## Validation

- Re-run the local reproduction command after the fix.
- go test -race ./...
- go test -race -count=10 ./path/to/package

## Likely files to inspect

- `**/*.go`


## Run Faultline

```bash
faultline analyze build.log
faultline explain go-data-race
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Go race detector found a data race
- Test: go race detector found a data race
- WARNING: DATA RACE
- faultline explain go-data-race
- Go go race detector found a data race


---

*Generated from [playbooks/bundled/log/test/go-data-race.yaml](../../playbooks/bundled/log/test/go-data-race.yaml). Do not edit directly — run `make docs-generate`.*
