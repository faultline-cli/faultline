# Testcontainers container failed to start

**Playbook ID:** `testcontainer-startup`
**Category:** test
**Severity:** medium
**Tags:** `testcontainers`, `docker`, `test`, `integration`, `container`

## What this failure means

A test using Testcontainers could not start the required Docker container. This usually means Docker is unavailable on the CI runner, or the runner cannot access the Docker socket.

## Common log signals

```text
Could not find a valid Docker environment
DockerClientProviderStrategy
org.testcontainers
container failed to start
Mapped port can only be obtained after the container is started
testcontainers
ProviderError
Ryuk started
```

## Diagnosis

A test using Testcontainers could not start the required Docker container. This usually means Docker is unavailable on the CI runner, or the runner cannot access the Docker socket.

## Fix steps

1. In GitHub Actions, ensure the runner actually has Docker available or add the appropriate Docker service.
2. In GitLab CI, add `docker:dind` and configure `DOCKER_HOST` if the test environment expects it.
3. If the job runs in a container, confirm `/var/run/docker.sock` is mounted or the remote Docker endpoint is reachable.
4. Use `TESTCONTAINERS_HOST_OVERRIDE` when Docker runs remotely.
5. Pre-pull large images in CI to avoid cold-start timeouts.

## Validation

- `docker info` succeeds in the same environment that runs the tests.
- Re-run the failing test command and confirm the container starts successfully.

## Likely files to inspect

- `src/test/`
- `docker-compose.yml`
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain testcontainer-startup
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Testcontainers container failed to start
- Test: testcontainers container failed to start
- Mapped port can only be obtained after the container is started
- faultline explain testcontainer-startup
- Docker testcontainers container failed to start


---

*Generated from [playbooks/bundled/log/test/testcontainer-startup.yaml](../../../playbooks/bundled/log/test/testcontainer-startup.yaml). Do not edit directly — run `make docs-generate`.*
