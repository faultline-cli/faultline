# no such host and getaddrinfo ENOTFOUND in CI

**Target search query:** `no such host getaddrinfo ENOTFOUND CI`

## Error snippet

```text
dial tcp: lookup api.internal.example: no such host
getaddrinfo ENOTFOUND api.internal.example
```

## What this error means

The runner could not resolve the hostname to an IP address. The job failed before it could even start the network connection.

## Fix steps

1. Check the hostname for typos in config, secrets, or environment variables.
2. Resolve the host from inside CI with `nslookup`, `dig`, or `getent hosts`.
3. Verify the hostname is not internal-only without the required private DNS or VPN path.
4. Retry once to rule out a transient resolver outage.

## How Faultline detects it

Faultline maps this failure to `dns-resolution`.

Primary signals:

- `no such host`
- `getaddrinfo ENOTFOUND`
- `could not resolve host`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [ECONNREFUSED and connection refused in CI services](econnrefused-connection-refused.md)
- [context deadline exceeded and read timeout in CI](context-deadline-exceeded-read-timeout.md)