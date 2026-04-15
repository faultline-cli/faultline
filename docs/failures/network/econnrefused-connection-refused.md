# ECONNREFUSED and connection refused in CI services

**Target search query:** `ECONNREFUSED connection refused CI service`

## Error snippet

```text
connect ECONNREFUSED 127.0.0.1:5432
could not connect to server: Connection refused
```

## What this error means

The target host actively refused the TCP connection. The service is not running, not ready yet, or is listening on a different host or port.

## Fix steps

1. Add a readiness probe before the client step.
2. Check service logs to confirm the dependency actually started.
3. Verify the service is listening on the expected port.
4. For containerized services, use the service name instead of `localhost` when appropriate.

## How Faultline detects it

Faultline maps this failure to `connection-refused`.

Primary signals:

- `connection refused`
- `ECONNREFUSED`
- `could not connect to server`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [no such host and getaddrinfo ENOTFOUND in CI](no-such-host-getaddrinfo-enotfound.md)
- [context deadline exceeded and read timeout in CI](context-deadline-exceeded-read-timeout.md)