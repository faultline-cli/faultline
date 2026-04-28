# LDAP server unreachable in CI

**Target search query:** `ldap_start_tls Can't contact LDAP server CI`

## Error snippet

```text
ldap_start_tls: Can't contact LDAP server (-1)
```

## What this error means

The OpenLDAP client returned `LDAP_SERVER_DOWN` (error code -1 / 81). The TCP
connection to the directory server could not be established before any LDAP
protocol exchange took place.

This is a **connectivity failure**, not an authentication failure. The process
exits immediately with code 1.

## Common causes in CI

- The LDAP host is a private IP address reachable only from on-premise runners.
  Cloud or shared runners cannot access it.
- A firewall or security group blocks port 389 (LDAP) or 636 (LDAPS) on the runner.
- `ldap_start_tls` (STARTTLS on port 389) was requested but the server does not
  support it or the port is wrong.
- The `LDAP_HOST` environment variable is unset, empty, or points to the wrong
  address for the runner environment.

## Fix steps

1. Verify connectivity from the runner before the LDAP operation:

   ```bash
   nc -zv "$LDAP_HOST" 389
   ```

2. Use a runner that has network access to the LDAP server, or expose the server
   via a sidecar container or VPN tunnel for cloud jobs.

3. Add a readiness probe to the CI job:

   ```bash
   wait-for-it.sh "$LDAP_HOST":389 --timeout=10 -- run-ldap-sync.sh
   ```

## How Faultline detects it

Faultline maps this failure to `ldap-connection-failure`.

Primary signals:

- `Can't contact LDAP server`
- `ldap_start_tls`
- `LDAP_SERVER_DOWN`

Run:

```bash
cat build.log | faultline analyze
```

## Reference

- [OpenLDAP result codes](https://www.openldap.org/doc/admin26/appendix-ldap-result-codes.html) —
  code 81 (`LDAP_SERVER_DOWN`) means the client could not connect to the server.

## Related failures

- [ECONNREFUSED and connection refused in CI](econnrefused-connection-refused.md)
