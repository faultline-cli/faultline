# pull access denied for ghcr.io image: fix Docker registry auth failures in CI

**Target search query:** `pull access denied for ghcr.io unauthorized authentication required`

## Error snippet

```text
pull access denied for ghcr.io/org/image
unauthorized: authentication required
Error response from daemon: Head "https://ghcr.io/v2/org/image/manifests/latest": denied
```

## What this error means

The CI job reached the registry, but the credential used for the pull or push was missing, expired, or scoped to the wrong image path.

## Fix steps

1. Verify the registry username, token, or password configured in CI secrets.
2. Ensure the login step runs before any `docker pull` or `docker push` command.
3. Confirm the token can access the exact image path being requested.
4. Re-run the login and pull locally with the same credential if possible.

## How Faultline detects it

Faultline maps this failure to `docker-auth`.

Primary signals:

- `pull access denied`
- `unauthorized: authentication required`
- `denied: requested access to the resource is denied`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [permission denied while trying to connect to the Docker daemon socket](permission-denied-while-trying-to-connect-to-the-docker-daemon-socket.md)
- [x509: certificate signed by unknown authority in Docker or CI builds](../tls/x509-certificate-signed-by-unknown-authority.md)