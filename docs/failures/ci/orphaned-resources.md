# Orphaned CI resources blocking subsequent runs

**Playbook ID:** `orphaned-resources`
**Category:** ci
**Severity:** medium
**Tags:** `docker`, `container`, `volume`, `resource`, `orphan`, `cleanup`, `ci`

## What this failure means

A CI job failed because a previous run left behind a container, network,
volume, or bound port that was not cleaned up. The current run cannot
create or start the same resource because it already exists.

## Common log signals

```text
container name.*already in use
container.*already exists
port.*already in use
network.*already exists
volume.*already exists
address already in use
EADDRINUSE
resource.*already allocated
```

## Diagnosis

Orphaned resources appear when:
1. A previous CI run was cancelled or crashed before its teardown step
2. Docker Compose services were not stopped after a test run
3. A service container keeps its name between retried jobs on the same runner
4. The runner reuses workspace state without a clean slate

Inspect the runner for leftover resources:

```bash
# Containers
docker ps -a | grep <project-prefix>
docker ps -a --filter "name=<service-name>"

# Networks
docker network ls

# Volumes
docker volume ls

# Ports
ss -tlnp | grep <port>
lsof -i :<port>
```

## Fix steps

1. Add a cleanup step at the start of the job to remove any stale resources:

   ```bash
   # At job start — remove stale containers
   docker rm -f myservice 2>/dev/null || true
   docker network rm mynetwork 2>/dev/null || true
   ```

2. For Docker Compose, use `--remove-orphans` and clean up volumes:

   ```bash
   # Teardown
   docker compose down --remove-orphans --volumes

   # Pre-run cleanup
   docker compose rm -f
   ```

3. Add a cleanup step as an `always()` / post-job hook so it runs even on
   failure:

   ```yaml
   # GitHub Actions
   - name: Cleanup
     if: always()
     run: docker compose down --remove-orphans --volumes
   ```

4. Use unique resource names per run to avoid collisions on shared runners:

   ```yaml
   env:
     SERVICE_NAME: myservice-${{ github.run_id }}
   ```

5. For port conflicts, run services on dynamic ports (port 0) and discover
   the assigned port rather than using a fixed port number.

6. For persistent runners (self-hosted), schedule a periodic cleanup job:

   ```bash
   # prune stopped containers, unused networks, dangling volumes
   docker system prune -f
   ```

## Validation

- Run `docker ps -a` on the runner and confirm no stale containers exist.
- Re-run the failing job and confirm it starts cleanly.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain orphaned-resources
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Orphaned CI resources blocking subsequent runs
- Ci: orphaned ci resources blocking subsequent runs
- container name.*already in use
- GitHub Actions orphaned ci resources blocking subsequent runs
- faultline explain orphaned-resources
- Docker orphaned ci resources blocking subsequent runs


---

*Generated from [playbooks/bundled/log/ci/orphaned-resources.yaml](../../../playbooks/bundled/log/ci/orphaned-resources.yaml). Do not edit directly — run `make docs-generate`.*
