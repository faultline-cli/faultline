# Port already in use during deployment

**Playbook ID:** `port-conflict`
**Category:** deploy
**Severity:** medium
**Tags:** `port`, `conflict`, `binding`, `deploy`, `network`

## What this failure means

A container or service failed to start because the port it needs to bind to
is already occupied by another process or a previous deployment that has not
been fully removed.

## Common log signals

```text
port is already allocated
failed to bind to port
port conflict
cannot start service
```

## Diagnosis

A container or service failed to start because the port it needs to bind to
is already occupied by another process or a previous deployment that has not
been fully removed.

In deploy contexts this usually means a host-port mapping is colliding during
rollout, or a prior container still owns the port on the target node.

## Fix steps

1. Find the process occupying the port: `lsof -i :<port>` or `ss -tlnp | grep <port>`.
2. In Kubernetes with host ports: ensure the old pod is fully terminated before the new one starts, or adjust the rollout strategy so the hand-off is serial.
3. Check for stopped containers: `docker ps -a | grep <image>` and remove obsolete containers with `docker rm`.
4. If using NodePort or HostPort in Kubernetes, verify no two pods on the same node request the same port.

## Validation

- Re-run the failing workflow step.
- Confirm the original failure signature for Port already in use during deployment is gone.

## Likely files to inspect

- `docker-compose.yml`
- `docker-compose*.yml`
- `k8s/**/*.yaml`
- `deploy/**/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain port-conflict
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Port already in use during deployment
- Deploy: port already in use during deployment
- port is already allocated
- faultline explain port-conflict


---

*Generated from [playbooks/bundled/log/deploy/port-conflict.yaml](../../../playbooks/bundled/log/deploy/port-conflict.yaml). Do not edit directly — run `make docs-generate`.*
