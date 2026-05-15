# Goroutine spawned without context or stop channel

**Playbook ID:** `goroutine-leak`
**Category:** runtime
**Severity:** high
**Tags:** `source`, `go`, `goroutine`, `leak`, `context`, `concurrency`

## What this failure means

A goroutine is spawned with an infinite loop and no context cancellation, stop channel, or WaitGroup — it will leak when the calling function returns.

## Common log signals

*(This playbook uses source-code pattern matching rather than log signals.)*

## Diagnosis

A goroutine is started with `go func()` and runs an infinite or long-running
loop without any mechanism to stop it. When the parent context or server shuts
down, this goroutine keeps running, leaking memory and CPU over time.

Common patterns:
- A background poller that loops with `time.Sleep` and no shutdown signal
- A cache-refresh goroutine that runs forever without a done channel
- A background worker spawned in `init()` or `main()` without lifecycle control

## Fix steps

1. Accept a `context.Context` parameter and check `ctx.Done()` in the loop:

   ```go
   func StartWorker(ctx context.Context) {
       go func() {
           for {
               select {
               case <-ctx.Done():
                   return
               default:
                   doWork()
                   time.Sleep(time.Second)
               }
           }
       }()
   }
   ```

2. Alternatively, use a stop channel: `stop <- struct{}{}` from the caller and
   `case <-stop: return` in the goroutine select.
3. Track goroutines with a `sync.WaitGroup` so callers can wait for clean shutdown.
4. For periodic background work, prefer `time.NewTicker` with a select over
   `time.Sleep` in a loop — the ticker can be stopped.

## Validation

- Re-run `faultline inspect .` or `faultline guard .`.
- Run `go test -race ./...` to catch races introduced during cleanup.
- Confirm that the background goroutine exits cleanly when the context is cancelled.

## Likely files to inspect

- `worker.go`
- `background.go`
- `scheduler.go`
- `poller.go`
- `server.go`


## Run Faultline

```bash
faultline analyze build.log
faultline explain goroutine-leak
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Goroutine spawned without context or stop channel
- Runtime: goroutine spawned without context or stop channel
- faultline explain goroutine-leak
- Go goroutine spawned without context or stop channel


---

*Generated from [playbooks/bundled/source/goroutine-leak.yaml](../../../playbooks/bundled/source/goroutine-leak.yaml). Do not edit directly — run `make docs-generate`.*
