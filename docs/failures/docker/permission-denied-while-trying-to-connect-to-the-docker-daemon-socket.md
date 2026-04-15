# permission denied while trying to connect to the Docker daemon socket

**Target search query:** `permission denied while trying to connect to the Docker daemon socket`

## Error snippet

```text
Got permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock:
Post "http://%2Fvar%2Frun%2Fdocker.sock/v1.41/containers/create": dial unix /var/run/docker.sock: connect: permission denied
```

## What this error means

The job can see the Docker socket path, but the current user or container does not have permission to use it.

## Fix steps

1. Verify the Docker socket is actually mounted into the job environment.
2. Run the step as a user that can access `/var/run/docker.sock`, or adjust the group mapping.
3. If using Docker-in-Docker, make sure the job is configured for DinD instead of assuming a host socket.
4. Re-run `docker info` from the same CI environment before the failing build step.

## How Faultline detects it

Faultline detects this today through the `permission-denied` family, while adjacent Docker runtime playbooks screen out broader daemon and BuildKit failures.

Primary signals:

- `permission denied`
- `EACCES`
- `Docker daemon socket`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [pull access denied for ghcr.io image: fix Docker registry auth failures in CI](pull-access-denied-ghcr-authentication-required.md)
- [x509: certificate signed by unknown authority in Docker or CI builds](../tls/x509-certificate-signed-by-unknown-authority.md)