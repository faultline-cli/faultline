# x509: certificate signed by unknown authority in Docker or CI builds

**Target search query:** `x509 certificate signed by unknown authority CI`

## Error snippet

```text
Get "https://registry.example.com/v2/": x509: certificate signed by unknown authority
```

## What this error means

The runner connected to the service, but the TLS certificate chain was not trusted by the system CA bundle in the CI environment.

## Fix steps

1. Verify the certificate chain and expiry with `openssl s_client -connect <host>:443`.
2. Install the internal or proxy CA certificate into the runner trust store.
3. Check the runner clock with `date -u` in case certificate validity is being misread.
4. Re-run the failing command after the trust store update.

## How Faultline detects it

Faultline maps this failure to `ssl-cert-error`.

Primary signals:

- `x509:`
- `certificate signed by unknown authority`
- certificate verification failure during HTTPS access

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [pull access denied for ghcr.io image: fix Docker registry auth failures in CI](../docker/pull-access-denied-ghcr-authentication-required.md)
- [permission denied while trying to connect to the Docker daemon socket](../docker/permission-denied-while-trying-to-connect-to-the-docker-daemon-socket.md)