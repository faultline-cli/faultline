# Cannot connect to the Docker daemon at unix:///var/run/docker.sock

**Target search query:** `Cannot connect to the Docker daemon at unix:///var/run/docker.sock Is the docker daemon running`

## Error snippet

```text
Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?
```

## What this error means

The job tried to use Docker, but the daemon was not reachable from the CI environment. That usually means the service never started, the socket is not mounted, or the job is running in an environment with no Docker service at all.

## Fix steps

1. Check whether the Docker daemon is actually running on the host or runner.
2. Verify the expected socket is mounted or the Docker service is available to the job.
3. If you use Docker-in-Docker, confirm the sidecar service started before the job step.
4. Re-run `docker info` and `docker ps` from the same environment before the failing command.

## How Faultline detects it

Faultline maps this failure to `docker-daemon-unavailable`.

Primary signals:

- `Cannot connect to the Docker daemon`
- `Is the docker daemon running?`
- `failed to connect to the Docker daemon`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [permission denied while trying to connect to the Docker daemon socket](permission-denied-while-trying-to-connect-to-the-docker-daemon-socket.md)
- [pull access denied for ghcr.io image: fix Docker registry auth failures in CI](pull-access-denied-ghcr-authentication-required.md)