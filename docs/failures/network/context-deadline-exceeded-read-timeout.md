# context deadline exceeded and read timeout in CI

**Target search query:** `context deadline exceeded read timeout CI`

## Error snippet

```text
Get "https://registry.example.com/v2/": context deadline exceeded
read timeout
```

## What this error means

The request reached the network layer but did not complete within the configured timeout. The remote service may be slow, unreachable through CI egress, or temporarily degraded.

## Fix steps

1. Identify the exact host and command that timed out.
2. Test connectivity from the runner with a short timeout using `curl` or `nc`.
3. Retry once to separate a transient blip from a persistent issue.
4. Only increase timeouts after confirming the endpoint is healthy and legitimately slow.

## How Faultline detects it

Faultline maps this failure to `network-timeout`.

Primary signals:

- `context deadline exceeded`
- `read timeout`
- `ETIMEDOUT`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [no such host and getaddrinfo ENOTFOUND in CI](no-such-host-getaddrinfo-enotfound.md)
- [ECONNREFUSED and connection refused in CI services](econnrefused-connection-refused.md)