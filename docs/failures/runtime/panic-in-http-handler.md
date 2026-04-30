# Unrecovered panic in HTTP or RPC handler

**Playbook ID:** `panic-in-http-handler`
**Category:** runtime
**Severity:** high
**Tags:** `source`, `go`, `panic`, `http`, `rpc`, `recovery`, `middleware`

## What this failure means

A `panic` call appears inside an HTTP or RPC handler function without a corresponding `recover()` in a deferred function. An unrecovered panic in a handler crashes the goroutine serving that request and may take down the entire server process (or cause a 500 with no explanation).

## Common log signals

*(This playbook uses source-code pattern matching rather than log signals.)*

## Diagnosis

A `panic` call appears inside an HTTP or RPC handler function without a corresponding `recover()` in a deferred function. An unrecovered panic in a handler crashes the goroutine serving that request and may take down the entire server process (or cause a 500 with no explanation).

## Fix steps

1. Add a recovery middleware at the router level so all handlers are protected: for Gin use `r.Use(gin.Recovery())`, for Echo use `e.Use(middleware.Recover())`, for stdlib `net/http` write a middleware that defers `recover()`.
2. Replace the explicit `panic(...)` with a returned error and an appropriate HTTP status code response.
3. If intentional panics are used for "impossible" invariants, confirm they cannot be triggered by external input (user-supplied data should never reach a `panic` call).
4. Use `if err != nil { return ..., err }` rather than `panic(err)` in all layers called by handlers.

## Validation

- go test ./...
- go vet ./...

## Likely files to inspect

- `handler.go`
- `router.go`
- `middleware.go`
- `http.go`
- `server.go`


## Run Faultline

```bash
faultline analyze build.log
faultline explain panic-in-http-handler
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Unrecovered panic in HTTP or RPC handler
- Runtime: unrecovered panic in http or rpc handler
- faultline explain panic-in-http-handler
- Go unrecovered panic in http or rpc handler


---

*Generated from [playbooks/bundled/source/panic-in-http-handler.yaml](../../../playbooks/bundled/source/panic-in-http-handler.yaml). Do not edit directly — run `make docs-generate`.*
