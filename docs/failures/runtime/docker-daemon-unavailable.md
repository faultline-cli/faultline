# Docker daemon unavailable

**Playbook ID:** `docker-daemon-unavailable`
**Category:** runtime
**Severity:** high
**Tags:** `docker`, `daemon`, `socket`, `runtime`, `ci`

## What this failure means

Docker could not reach the daemon socket or service, so the step failed before it could inspect containers or run a build.

## Common log signals

```text
Cannot connect to the Docker daemon
Cannot connect to the docker daemon
Is the docker daemon running?
Docker daemon is not running
failed to connect to the Docker daemon
failed to connect to docker daemon
unable to connect to the Docker daemon
dockerd failed to start
```

## Diagnosis

Docker could not reach the daemon socket or service, so the step failed before it could inspect containers or run a build.

This usually means the daemon is stopped, the socket path is wrong, or the environment does not expose Docker at all.

## Fix steps

1. Check whether the Docker daemon is actually running on the host or runner.
2. Verify the expected socket is mounted or the Docker service is available to the job.
3. If you are using Docker-in-Docker, confirm the sidecar or service container started before the job step.
4. Re-run a minimal probe from the same environment:

   ```bash
   docker info
   docker ps
   ```

## Validation

- `docker info` succeeds from the same job environment.
- Re-run the failing step and confirm the Docker daemon is reachable.

## Likely files to inspect

- `Dockerfile`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `.gitlab-ci.yml`
- `docker-compose*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain docker-daemon-unavailable
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Docker daemon unavailable
- Runtime: docker daemon unavailable
- failed to connect to the Docker daemon
- GitHub Actions docker daemon unavailable
- faultline explain docker-daemon-unavailable
- Docker docker daemon unavailable


---

*Generated from [playbooks/bundled/log/runtime/docker-daemon-unavailable.yaml](../../../playbooks/bundled/log/runtime/docker-daemon-unavailable.yaml). Do not edit directly — run `make docs-generate`.*
