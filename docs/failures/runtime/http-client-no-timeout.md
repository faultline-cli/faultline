# HTTP client used without a timeout

**Playbook ID:** `http-client-no-timeout`
**Category:** runtime
**Severity:** high
**Tags:** `source`, `go`, `http`, `timeout`, `reliability`, `production`

## What this failure means

An HTTP request is made using the package-level `http.Get` / `http.Post`
functions or an empty `&http.Client{}` literal — neither sets a timeout, so
the call can hang indefinitely if the server is slow or unresponsive.

## Common log signals

*(This playbook uses source-code pattern matching rather than log signals.)*

## Diagnosis

The package-level `http.Get`, `http.Post`, or `http.DefaultClient` all rely
on a default `http.Client` with `Timeout: 0` (no timeout). A call to a slow
or unreachable server will block the goroutine until the OS TCP timeout fires
— typically tens of minutes.

Common patterns:
- `resp, err := http.Get(url)` — uses DefaultClient with no timeout
- `resp, err := http.Post(url, ...)` — uses DefaultClient with no timeout
- `client := &http.Client{}` followed by `client.Do(req)` — Timeout is zero

In production services, this causes goroutine exhaustion, request queue
buildup, and cascading timeouts across all downstream callers.

## Fix steps

1. Create a client with a timeout appropriate for the operation:
   ```go
   client := &http.Client{Timeout: 10 * time.Second}
   resp, err := client.Get(url)
   ```
2. For per-request control, use a context with a deadline:
   ```go
   ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
   defer cancel()
   req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
   if err != nil { return nil, err }
   resp, err := client.Do(req)
   ```
3. Define timeouts as named constants or environment-driven configuration so
   they can be tuned without a code change.
4. Avoid using `http.DefaultClient` in any production code path.

## Validation

- Re-run `faultline inspect .` or `faultline guard .`.
- Add a test that confirms the client respects the configured timeout by
  pointing it at a handler that delays its response.

## Likely files to inspect

- `**/*.go`


## Run Faultline

```bash
faultline analyze build.log
faultline explain http-client-no-timeout
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- HTTP client used without a timeout
- Runtime: http client used without a timeout
- faultline explain http-client-no-timeout
- Go http client used without a timeout


---

*Generated from [playbooks/bundled/source/http-client-no-timeout.yaml](../../../playbooks/bundled/source/http-client-no-timeout.yaml). Do not edit directly — run `make docs-generate`.*
