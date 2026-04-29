# SSL/TLS certificate error

**Playbook ID:** `ssl-cert-error`
**Category:** network
**Severity:** high
**Tags:** `ssl`, `tls`, `certificate`, `https`

## What this failure means

A TLS/SSL certificate error prevented a secure connection. The certificate was expired, self-signed, or signed by an untrusted certificate authority.

## Common log signals

```text
certificate verify failed
ssl certificate problem
certificate has expired
tls handshake failed
unable to verify the first certificate
x509:
ssl_error
certificate signed by unknown authority
```

## Diagnosis

A TLS/SSL certificate error prevented a secure connection. The certificate was expired, self-signed, or signed by an untrusted certificate authority.

## Fix steps

1. Check whether the certificate has expired and inspect its subject and
   issuer:

   ```bash
   openssl s_client -servername <host> -connect <host>:443 </dev/null 2>/dev/null \
     | openssl x509 -noout -dates -subject -issuer
   ```

2. Verify the system clock on the CI runner is accurate — even a few minutes
   of drift can cause certificate validation to fail:

   ```bash
   date -u
   ```

3. If connecting to an internal service, add the CA certificate to the
   system trust store (`update-ca-certificates` on Debian/Ubuntu,
   `update-ca-trust` on RHEL/CentOS).
4. If the failure only appears in corporate CI, a TLS-intercepting proxy may
   be presenting its own certificate. Install the proxy's root CA on the
   runner image.
5. Contact the service owner if their certificate has genuinely expired.

## Validation

- `openssl s_client -servername <host> -connect <host>:443 </dev/null 2>/dev/null | openssl x509 -noout -dates`
  shows a `notAfter` date in the future.
- `date -u` confirms the system clock is accurate.
- Re-run the failing step and confirm the TLS handshake succeeds.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `Dockerfile`
- `docker-compose*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain ssl-cert-error
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- SSL/TLS certificate error
- Network: ssl/tls certificate error
- certificate signed by unknown authority
- faultline explain ssl-cert-error


---

*Generated from [playbooks/bundled/log/network/ssl-cert-error.yaml](../../playbooks/bundled/log/network/ssl-cert-error.yaml). Do not edit directly — run `make docs-generate`.*
