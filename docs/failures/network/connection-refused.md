# Service or dependency connection refused

**Playbook ID:** `connection-refused`
**Category:** network
**Severity:** high
**Tags:** `network`, `connection`, `refused`, `service`, `dependency`

## What this failure means

A TCP connection attempt was actively refused by the target host. The service is either not running on that address and port, has not finished starting up, or is bound to a different interface.

## Common log signals

```text
connection refused
Connection refused
could not connect to server
connection was refused
ECONNREFUSED
ECONNRESET
connect: connection refused
no connection could be made
```

## Diagnosis

A TCP connection attempt was actively refused by the target host. The service is either not running on that address and port, has not finished starting up, or is bound to a different interface.

## Fix steps

1. Add a readiness probe before the connecting step:

   ```bash
   wait-for-it.sh <host>:<port> -t 30
   # or with curl retry:
   curl --retry 10 --retry-delay 3 --retry-connrefused http://<host>:<port>/health
   ```

2. Check whether the service container actually started successfully:

   ```bash
   docker compose logs <service>      # Docker Compose
   kubectl logs <pod>                 # Kubernetes
   ```

3. Confirm the port the service is listening on:

   ```bash
   ss -tlnp | grep <port>
   ```

4. For Docker Compose services: use the service name as the hostname (e.g.,
   `postgres`, `redis`), not `localhost` — containers on a shared network
   resolve each other by service name.
5. Check for race conditions: the client may be starting before the server is
   ready. Add an explicit health-check dependency in the CI service definition.

## Validation

- `curl -sf http://<host>:<port>/health` or `nc -zv <host> <port>` — confirm
  the service responds before re-running.
- Re-run the failing step and confirm no `ECONNREFUSED` error.

## Likely files to inspect

- `docker-compose.yml`
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain connection-refused
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Service or dependency connection refused
- Network: service or dependency connection refused
- could not connect to server
- faultline explain connection-refused


---

*Generated from [playbooks/bundled/log/network/connection-refused.yaml](../../playbooks/bundled/log/network/connection-refused.yaml). Do not edit directly — run `make docs-generate`.*
