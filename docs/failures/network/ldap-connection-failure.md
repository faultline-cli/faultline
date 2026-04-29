# LDAP server unreachable

**Playbook ID:** `ldap-connection-failure`
**Category:** network
**Severity:** high
**Tags:** `network`, `ldap`, `directory`, `authentication`, `tls`

## What this failure means

The LDAP client could not reach the directory server. The server is offline,
the host or port is misconfigured, or a network/firewall rule is blocking the
connection.

## Common log signals

```text
Can't contact LDAP server
ldap_start_tls
LDAP_SERVER_DOWN
ldap error 81
Can't contact LDAP
re:ldap.*-1\)
```

## Diagnosis

The OpenLDAP client error `ldap_start_tls: Can't contact LDAP server (-1)`
corresponds to the `LDAP_SERVER_DOWN` (-1/81) result code. The TCP connection
to the LDAP server could not be established before any protocol exchange began.

Common causes in CI:
- The LDAP server address (`LDAP_HOST`) is a private IP only reachable from
  on-premise runners; cloud runners cannot connect.
- The LDAP service is down or its port (389/636) is firewalled.
- TLS (`ldap_start_tls`) failed because the server does not support STARTTLS on
  the configured port.

## Fix steps

1. Verify the host and port from the runner:

   ```bash
   nc -zv "$LDAP_HOST" 389   # plain
   nc -zv "$LDAP_HOST" 636   # TLS
   ```

2. If the LDAP server is internal, ensure the CI runner has network access (VPN,
   peering, runner placement in the correct subnet).
3. For GitLab CI, confirm the runner has access to the host specified in
   `LDAP_HOST`. If the job uses a shared runner, it will not reach
   on-premise servers.
4. If using `ldap_start_tls` (port 389 + STARTTLS), confirm the server supports
   STARTTLS. Switch to LDAPS (port 636) if not.
5. Add a connectivity pre-check before the LDAP operation:

   ```bash
   wait-for-it.sh "$LDAP_HOST":389 --timeout=10 -- echo "LDAP reachable"
   ```

## Validation

- `ldapsearch -H ldap://<host>:389 -x -b "" -s base` — confirms the server
  is reachable and responds to anonymous requests.
- Re-run the failing job and confirm the `Can't contact LDAP server` error is gone.

## Likely files to inspect

- `.gitlab-ci.yml`
- `.github/workflows/*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain ldap-connection-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- LDAP server unreachable
- Network: ldap server unreachable
- Can't contact LDAP server
- faultline explain ldap-connection-failure


---

*Generated from [playbooks/bundled/log/network/ldap-connection-failure.yaml](../../playbooks/bundled/log/network/ldap-connection-failure.yaml). Do not edit directly — run `make docs-generate`.*
